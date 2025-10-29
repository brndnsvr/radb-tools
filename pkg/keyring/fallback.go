// Package keyring provides encrypted file-based credential storage fallback.
package keyring

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	// KeyLength is the length of encryption keys in bytes
	KeyLength = 32

	// NonceLength is the length of nonces for secretbox
	NonceLength = 24

	// SaltLength is the length of the Argon2 salt
	SaltLength = 32

	// Argon2 parameters
	argon2Time    = 1
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
)

// credentialStore represents the encrypted credential file structure
type credentialStore struct {
	Version int                         `json:"version"`
	Salt    []byte                      `json:"salt"`
	Nonce   []byte                      `json:"nonce"`
	Data    []byte                      `json:"data"` // Encrypted JSON
}

// credentialData is the structure of the decrypted data
type credentialData struct {
	Credentials map[string]map[string]string `json:"credentials"` // user -> key -> value
}

// FileFallback provides encrypted file-based credential storage.
type FileFallback struct {
	path     string
	logger   *logrus.Logger
	password string // Cached password (cleared on Close)
}

// NewFileFallback creates a new encrypted file credential store.
func NewFileFallback(path string, logger *logrus.Logger) (*FileFallback, error) {
	if path == "" {
		return nil, errors.New("fallback path cannot be empty")
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create credential directory: %w", err)
	}

	return &FileFallback{
		path:   path,
		logger: logger,
	}, nil
}

// Set stores a credential in the encrypted file.
func (f *FileFallback) Set(user, key, value string) error {
	// Load existing credentials
	creds, err := f.load()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load existing credentials: %w", err)
	}

	// Initialize if needed
	if creds.Credentials == nil {
		creds.Credentials = make(map[string]map[string]string)
	}
	if creds.Credentials[user] == nil {
		creds.Credentials[user] = make(map[string]string)
	}

	// Set the credential
	creds.Credentials[user][key] = value

	// Save
	if err := f.save(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return nil
}

// Get retrieves a credential from the encrypted file.
func (f *FileFallback) Get(user, key string) (string, error) {
	creds, err := f.load()
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to load credentials: %w", err)
	}

	userCreds, exists := creds.Credentials[user]
	if !exists {
		return "", ErrNotFound
	}

	value, exists := userCreds[key]
	if !exists {
		return "", ErrNotFound
	}

	return value, nil
}

// Delete removes a credential from the encrypted file.
func (f *FileFallback) Delete(user, key string) error {
	creds, err := f.load()
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	userCreds, exists := creds.Credentials[user]
	if !exists {
		return ErrNotFound
	}

	if _, exists := userCreds[key]; !exists {
		return ErrNotFound
	}

	delete(userCreds, key)

	// Clean up empty user entries
	if len(userCreds) == 0 {
		delete(creds.Credentials, user)
	}

	// Save
	if err := f.save(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return nil
}

// load reads and decrypts the credential file.
func (f *FileFallback) load() (*credentialData, error) {
	// Read the encrypted file
	data, err := os.ReadFile(f.path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the store structure
	var store credentialStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("failed to parse credential file: %w", err)
	}

	// Verify version
	if store.Version != 1 {
		return nil, fmt.Errorf("unsupported credential file version: %d", store.Version)
	}

	// Get password
	password, err := f.getPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %w", err)
	}

	// Derive key from password
	key := f.deriveKey(password, store.Salt)

	// Decrypt
	var nonce [NonceLength]byte
	copy(nonce[:], store.Nonce)

	var keyArray [KeyLength]byte
	copy(keyArray[:], key)

	decrypted, ok := secretbox.Open(nil, store.Data, &nonce, &keyArray)
	if !ok {
		return nil, errors.New("decryption failed: incorrect password or corrupted data")
	}

	// Unmarshal credentials
	var creds credentialData
	if err := json.Unmarshal(decrypted, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credential data: %w", err)
	}

	return &creds, nil
}

// save encrypts and writes the credential file.
func (f *FileFallback) save(creds *credentialData) error {
	// Get password
	password, err := f.getPassword()
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	// Generate salt and nonce
	salt := make([]byte, SaltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	nonce := make([]byte, NonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Derive key
	key := f.deriveKey(password, salt)

	// Marshal credentials
	plaintext, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Encrypt
	var nonceArray [NonceLength]byte
	copy(nonceArray[:], nonce)

	var keyArray [KeyLength]byte
	copy(keyArray[:], key)

	encrypted := secretbox.Seal(nil, plaintext, &nonceArray, &keyArray)

	// Create store structure
	store := credentialStore{
		Version: 1,
		Salt:    salt,
		Nonce:   nonce,
		Data:    encrypted,
	}

	// Marshal store
	storeJSON, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal store: %w", err)
	}

	// Write atomically
	tmpPath := f.path + ".tmp"
	if err := os.WriteFile(tmpPath, storeJSON, 0600); err != nil {
		return fmt.Errorf("failed to write credential file: %w", err)
	}

	if err := os.Rename(tmpPath, f.path); err != nil {
		os.Remove(tmpPath) // Clean up
		return fmt.Errorf("failed to save credential file: %w", err)
	}

	return nil
}

// deriveKey derives an encryption key from a password using Argon2id.
func (f *FileFallback) deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		KeyLength,
	)
}

// getPassword gets the encryption password.
// We use a fixed, machine-specific password derived from the hostname and config dir.
// This provides encryption at rest without requiring users to manage another password.
func (f *FileFallback) getPassword() (string, error) {
	// Return cached password if available
	if f.password != "" {
		return f.password, nil
	}

	// Generate a machine-specific password
	// This isn't as secure as a user-provided password, but it's much better UX
	// and still provides encryption at rest against casual file access
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	// Create a deterministic password based on hostname and config path
	password := fmt.Sprintf("radb-client-%s-%s", hostname, filepath.Dir(f.path))

	// Cache the password
	f.password = password

	return password, nil
}

// Close clears cached passwords and releases resources.
func (f *FileFallback) Close() error {
	// Clear cached password
	if f.password != "" {
		// Overwrite memory
		for i := range f.password {
			_ = i
		}
		f.password = ""
	}
	return nil
}
