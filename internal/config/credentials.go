package config

import (
	"fmt"
	"path/filepath"

	"github.com/bss/radb-client/pkg/keyring"
	"github.com/sirupsen/logrus"
)

// Credentials represents user credentials.
type Credentials struct {
	Username string
	Password string
	APIKey   string
}

// CredentialManager handles secure credential storage and retrieval.
type CredentialManager struct {
	store  *keyring.Store
	logger *logrus.Logger
}

// NewCredentialManager creates a new credential manager.
func NewCredentialManager(configDir string, logger *logrus.Logger) (*CredentialManager, error) {
	fallbackPath := filepath.Join(configDir, "credentials.enc")

	store, err := keyring.NewStore(logger, fallbackPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize credential store: %w", err)
	}

	return &CredentialManager{
		store:  store,
		logger: logger,
	}, nil
}

// SetPassword stores the user's password.
func (cm *CredentialManager) SetPassword(username, password string) error {
	if err := cm.store.Set(username, "password", password); err != nil {
		return fmt.Errorf("failed to store password: %w", err)
	}
	cm.logger.Debugf("Stored password for user %s", username)
	return nil
}

// GetPassword retrieves the user's password.
func (cm *CredentialManager) GetPassword(username string) (string, error) {
	password, err := cm.store.Get(username, "password")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve password: %w", err)
	}
	return password, nil
}

// SetAPIKey stores the user's API key.
func (cm *CredentialManager) SetAPIKey(username, apiKey string) error {
	if err := cm.store.Set(username, "api_key", apiKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}
	cm.logger.Debugf("Stored API key for user %s", username)
	return nil
}

// GetAPIKey retrieves the user's API key.
func (cm *CredentialManager) GetAPIKey(username string) (string, error) {
	apiKey, err := cm.store.Get(username, "api_key")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve API key: %w", err)
	}
	return apiKey, nil
}

// SetCryptedPassword stores the crypted password for write operations.
func (cm *CredentialManager) SetCryptedPassword(username, cryptedPassword string) error {
	if err := cm.store.Set(username, "crypted_password", cryptedPassword); err != nil {
		return fmt.Errorf("failed to store crypted password: %w", err)
	}
	cm.logger.Debugf("Stored crypted password for user %s", username)
	return nil
}

// GetCryptedPassword retrieves the crypted password.
func (cm *CredentialManager) GetCryptedPassword(username string) (string, error) {
	cryptedPassword, err := cm.store.Get(username, "crypted_password")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve crypted password: %w", err)
	}
	return cryptedPassword, nil
}

// DeleteAll removes all credentials for a user.
func (cm *CredentialManager) DeleteAll(username string) error {
	if err := cm.store.DeleteAll(username); err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}
	cm.logger.Infof("Deleted all credentials for user %s", username)
	return nil
}

// GetCredentials retrieves stored credentials.
func (cm *CredentialManager) GetCredentials(username string) (*Credentials, error) {
	password, err := cm.GetPassword(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	apiKey, _ := cm.GetAPIKey(username) // API key is optional

	return &Credentials{
		Username: username,
		Password: password,
		APIKey:   apiKey,
	}, nil
}

// SaveCredentials stores credentials securely.
func (cm *CredentialManager) SaveCredentials(creds *Credentials) error {
	if err := cm.SetPassword(creds.Username, creds.Password); err != nil {
		return err
	}

	if creds.APIKey != "" {
		if err := cm.SetAPIKey(creds.Username, creds.APIKey); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the credential manager and releases resources.
func (cm *CredentialManager) Close() error {
	return cm.store.Close()
}
