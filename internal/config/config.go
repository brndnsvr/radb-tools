// Package config provides configuration management using Viper.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// DefaultConfigDir is the default configuration directory
	DefaultConfigDir = ".radb-client"

	// DefaultConfigFile is the default configuration file name
	DefaultConfigFile = "config.yaml"
)

// Config represents the application configuration.
type Config struct {
	API          APIConfig          `mapstructure:"api"`
	Credentials  CredentialsConfig  `mapstructure:"credentials"`
	Preferences  PreferencesConfig  `mapstructure:"preferences"`
	Performance  PerformanceConfig  `mapstructure:"performance"`
	State        StateConfig        `mapstructure:"state"`

	// Runtime fields (not persisted)
	ConfigDir  string `mapstructure:"-"`
	ConfigFile string `mapstructure:"-"`
}

// APIConfig contains API-related configuration.
type APIConfig struct {
	BaseURL    string       `mapstructure:"base_url"`
	Source     string       `mapstructure:"source"`
	Format     string       `mapstructure:"format"`
	Timeout    int          `mapstructure:"timeout"`
	RateLimit  RateLimit    `mapstructure:"rate_limit"`
	Retry      RetryConfig  `mapstructure:"retry"`
}

// RateLimit contains rate limiting configuration.
type RateLimit struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	BurstSize         int `mapstructure:"burst_size"`
}

// RetryConfig contains retry configuration.
type RetryConfig struct {
	MaxAttempts        int `mapstructure:"max_attempts"`
	BackoffMultiplier  int `mapstructure:"backoff_multiplier"`
	InitialDelayMs     int `mapstructure:"initial_delay_ms"`
}

// CredentialsConfig contains credential storage configuration.
type CredentialsConfig struct {
	Username string `mapstructure:"username"`
	// Password and API key are stored in keyring, not in config file
}

// PreferencesConfig contains user preferences.
type PreferencesConfig struct {
	CacheDir   string `mapstructure:"cache_dir"`
	HistoryDir string `mapstructure:"history_dir"`
	LogLevel   string `mapstructure:"log_level"`
}

// PerformanceConfig contains performance-related settings.
type PerformanceConfig struct {
	StreamThreshold       int  `mapstructure:"stream_threshold"`
	CompressHistory       bool `mapstructure:"compress_history"`
	MaxConcurrentRequests int  `mapstructure:"max_concurrent_requests"`
}

// StateConfig contains state management settings.
type StateConfig struct {
	EnableLocking bool   `mapstructure:"enable_locking"`
	AtomicWrites  bool   `mapstructure:"atomic_writes"`
	FormatVersion string `mapstructure:"format_version"`
}

// Default returns a configuration with sensible defaults.
func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, DefaultConfigDir)

	return &Config{
		API: APIConfig{
			BaseURL: "https://api.radb.net/api",
			Source:  "RADB",
			Format:  "json",
			Timeout: 30,
			RateLimit: RateLimit{
				RequestsPerMinute: 60,
				BurstSize:         10,
			},
			Retry: RetryConfig{
				MaxAttempts:       3,
				BackoffMultiplier: 2,
				InitialDelayMs:    1000,
			},
		},
		Credentials: CredentialsConfig{
			Username: "",
		},
		Preferences: PreferencesConfig{
			CacheDir:   filepath.Join(configDir, "cache"),
			HistoryDir: filepath.Join(configDir, "history"),
			LogLevel:   "INFO",
		},
		Performance: PerformanceConfig{
			StreamThreshold:       1000,
			CompressHistory:       true,
			MaxConcurrentRequests: 5,
		},
		State: StateConfig{
			EnableLocking: true,
			AtomicWrites:  true,
			FormatVersion: "1.0",
		},
		ConfigDir:  configDir,
		ConfigFile: filepath.Join(configDir, DefaultConfigFile),
	}
}

// Load loads configuration from file and environment variables.
func Load() (*Config, error) {
	cfg := Default()

	// Set up Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(cfg.ConfigDir)
	viper.AddConfigPath(".")

	// Environment variable support
	viper.SetEnvPrefix("RADB")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found is okay, we'll use defaults
	}

	// Unmarshal into struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

// Save writes the configuration to file.
func (c *Config) Save() error {
	// Ensure config directory exists
	if err := os.MkdirAll(c.ConfigDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Update viper with current values
	viper.Set("api", c.API)
	viper.Set("credentials", c.Credentials)
	viper.Set("preferences", c.Preferences)
	viper.Set("performance", c.Performance)
	viper.Set("state", c.State)

	// Write config file
	if err := viper.WriteConfigAs(c.ConfigFile); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Initialize creates a new configuration file with defaults.
func Initialize() (*Config, error) {
	cfg := Default()

	// Check if config already exists
	if _, err := os.Stat(cfg.ConfigFile); err == nil {
		return nil, fmt.Errorf("configuration already exists at %s", cfg.ConfigFile)
	}

	// Create directories
	if err := os.MkdirAll(cfg.ConfigDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.MkdirAll(cfg.Preferences.CacheDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := os.MkdirAll(cfg.Preferences.HistoryDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	// Save default configuration
	if err := cfg.Save(); err != nil {
		return nil, fmt.Errorf("failed to save default config: %w", err)
	}

	return cfg, nil
}

// GetLogger returns a configured logger based on config settings.
func (c *Config) GetLogger() *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(c.Preferences.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set formatter
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return logger
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.API.BaseURL == "" {
		return fmt.Errorf("api.base_url is required")
	}

	if c.API.Source == "" {
		return fmt.Errorf("api.source is required")
	}

	if c.API.Timeout <= 0 {
		return fmt.Errorf("api.timeout must be positive")
	}

	if c.Preferences.CacheDir == "" {
		return fmt.Errorf("preferences.cache_dir is required")
	}

	if c.Preferences.HistoryDir == "" {
		return fmt.Errorf("preferences.history_dir is required")
	}

	return nil
}

// StateDir returns the state directory (alias for CacheDir for snapshots).
func (c *Config) StateDir() string {
	return c.Preferences.CacheDir
}

// Save is a package-level function that saves a config.
func Save(cfg *Config) error {
	return cfg.Save()
}

// LoadCredentials loads credentials from the credential manager.
// It tries to load credentials for the configured username.
func LoadCredentials() (*Credentials, error) {
	cfg, err := Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Credentials.Username == "" {
		return nil, fmt.Errorf("no username configured")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	credMgr, err := NewCredentialManager(cfg.ConfigDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize credential manager: %w", err)
	}
	defer credMgr.Close()

	return credMgr.GetCredentials(cfg.Credentials.Username)
}

// SaveCredentials saves credentials using the credential manager.
func SaveCredentials(creds *Credentials) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, DefaultConfigDir)

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	credMgr, err := NewCredentialManager(configDir, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize credential manager: %w", err)
	}
	defer credMgr.Close()

	return credMgr.SaveCredentials(creds)
}
