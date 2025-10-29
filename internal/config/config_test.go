package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.API.BaseURL == "" {
		t.Error("Expected default base URL")
	}

	if cfg.API.Source != "RADB" {
		t.Errorf("Expected source RADB, got %s", cfg.API.Source)
	}

	if cfg.API.Timeout <= 0 {
		t.Error("Expected positive timeout")
	}

	if cfg.API.RateLimit.RequestsPerMinute <= 0 {
		t.Error("Expected positive rate limit")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{
			name:    "valid config",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name: "missing base URL",
			modify: func(c *Config) {
				c.API.BaseURL = ""
			},
			wantErr: true,
		},
		{
			name: "missing source",
			modify: func(c *Config) {
				c.API.Source = ""
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			modify: func(c *Config) {
				c.API.Timeout = -1
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.modify(cfg)

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInitialize(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "radb-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Check that config file was created
	if _, err := os.Stat(cfg.ConfigFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Check that directories were created
	if _, err := os.Stat(cfg.Preferences.CacheDir); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}

	if _, err := os.Stat(cfg.Preferences.HistoryDir); os.IsNotExist(err) {
		t.Error("History directory was not created")
	}
}
