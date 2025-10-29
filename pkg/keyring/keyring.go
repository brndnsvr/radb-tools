// Package keyring provides secure credential storage using system keyring
// with automatic fallback to encrypted file storage.
package keyring

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/zalando/go-keyring"
)

const (
	// ServiceName is the identifier for this application in the keyring
	ServiceName = "radb-client"
)

var (
	// ErrNotFound indicates the credential was not found
	ErrNotFound = errors.New("credential not found")

	// ErrKeyringUnavailable indicates the system keyring is not available
	ErrKeyringUnavailable = errors.New("system keyring unavailable")
)

// Store provides a unified interface for credential storage.
// It attempts to use the system keyring first, falling back to encrypted
// file storage if the keyring is unavailable.
type Store struct {
	fallback *FileFallback
	logger   *logrus.Logger
}

// NewStore creates a new credential store.
// It attempts to detect keyring availability and initializes the fallback if needed.
func NewStore(logger *logrus.Logger, fallbackPath string) (*Store, error) {
	if logger == nil {
		logger = logrus.New()
	}

	// Initialize fallback
	fallback, err := NewFileFallback(fallbackPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize fallback storage: %w", err)
	}

	return &Store{
		fallback: fallback,
		logger:   logger,
	}, nil
}

// Set stores a credential with the given key.
// It attempts to use the system keyring first, falling back to encrypted file storage.
func (s *Store) Set(user, key, value string) error {
	// Try system keyring first
	err := keyring.Set(ServiceName, fmt.Sprintf("%s:%s", user, key), value)
	if err == nil {
		s.logger.Debugf("Stored credential %s for user %s in system keyring", key, user)
		return nil
	}

	// Log keyring failure and fall back
	s.logger.Debugf("System keyring unavailable (%v), using encrypted file fallback", err)

	// Use encrypted file fallback
	if err := s.fallback.Set(user, key, value); err != nil {
		return fmt.Errorf("failed to store credential in fallback: %w", err)
	}

	s.logger.Debugf("Stored credential %s for user %s in encrypted file", key, user)
	return nil
}

// Get retrieves a credential with the given key.
// It checks the system keyring first, then falls back to encrypted file storage.
func (s *Store) Get(user, key string) (string, error) {
	// Try system keyring first
	value, err := keyring.Get(ServiceName, fmt.Sprintf("%s:%s", user, key))
	if err == nil {
		s.logger.Debugf("Retrieved credential %s for user %s from system keyring", key, user)
		return value, nil
	}

	// If not found in keyring, try fallback
	value, fallbackErr := s.fallback.Get(user, key)
	if fallbackErr == nil {
		s.logger.Debugf("Retrieved credential %s for user %s from encrypted file", key, user)
		return value, nil
	}

	// Neither storage method has the credential
	if errors.Is(fallbackErr, ErrNotFound) {
		return "", ErrNotFound
	}

	return "", fmt.Errorf("failed to retrieve credential: keyring: %v, fallback: %w", err, fallbackErr)
}

// Delete removes a credential with the given key.
// It removes from both system keyring and fallback storage.
func (s *Store) Delete(user, key string) error {
	var errs []error

	// Delete from keyring (ignore errors if not present)
	err := keyring.Delete(ServiceName, fmt.Sprintf("%s:%s", user, key))
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		errs = append(errs, fmt.Errorf("keyring delete failed: %w", err))
	}

	// Delete from fallback
	if err := s.fallback.Delete(user, key); err != nil && !errors.Is(err, ErrNotFound) {
		errs = append(errs, fmt.Errorf("fallback delete failed: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during deletion: %v", errs)
	}

	s.logger.Debugf("Deleted credential %s for user %s", key, user)
	return nil
}

// DeleteAll removes all credentials for a user.
func (s *Store) DeleteAll(user string) error {
	// Common credential keys
	keys := []string{"password", "api_key", "crypted_password"}

	var errs []error
	for _, key := range keys {
		if err := s.Delete(user, key); err != nil && !errors.Is(err, ErrNotFound) {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during delete all: %v", errs)
	}

	return nil
}

// IsAvailable checks if any credential storage is available.
func (s *Store) IsAvailable() bool {
	// Fallback is always available if Store was successfully created
	return s.fallback != nil
}

// Close closes the credential store and releases any resources.
func (s *Store) Close() error {
	if s.fallback != nil {
		return s.fallback.Close()
	}
	return nil
}
