# Go Implementation Guide

## Architecture Overview

This document details the Go-specific implementation of the RADb API client.

## Project Structure

```
radb-client/
├── cmd/
│   └── radb-client/
│       └── main.go                 # Entry point
├── internal/
│   ├── api/
│   │   ├── client.go              # HTTP client wrapper
│   │   ├── auth.go                # Authentication handling
│   │   ├── routes.go              # Route object operations
│   │   ├── contacts.go            # Contact operations
│   │   ├── search.go              # Search operations
│   │   └── models.go              # API request/response models
│   ├── config/
│   │   ├── config.go              # Configuration management
│   │   └── credentials.go         # Secure credential storage
│   ├── state/
│   │   ├── manager.go             # State management
│   │   ├── snapshot.go            # Snapshot creation/loading
│   │   ├── diff.go                # Diff generation
│   │   └── history.go             # Historical data management
│   ├── cli/
│   │   ├── root.go                # Root command
│   │   ├── config.go              # Config commands
│   │   ├── auth.go                # Auth commands
│   │   ├── route.go               # Route commands
│   │   ├── contact.go             # Contact commands
│   │   ├── search.go              # Search commands
│   │   ├── history.go             # History commands
│   │   └── snapshot.go            # Snapshot commands
│   └── models/
│       ├── route.go               # Route object domain model
│       ├── contact.go             # Contact domain model
│       ├── snapshot.go            # Snapshot domain model
│       └── changelog.go           # Changelog domain model
├── pkg/
│   ├── keyring/
│   │   └── keyring.go             # Keyring wrapper with fallback
│   └── httpclient/
│       └── retry.go               # HTTP retry logic
├── testdata/
│   ├── fixtures/                  # Test fixtures
│   └── mocks/                     # Mock data
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── DESIGN.md
└── GO_IMPLEMENTATION.md
```

## Key Go Libraries

### Core Dependencies

```go
// CLI Framework
"github.com/spf13/cobra"           // Command structure and parsing
"github.com/spf13/viper"           // Configuration management

// Security
"github.com/zalando/go-keyring"    // System keyring access
"golang.org/x/crypto/bcrypt"       // Password hashing (if needed)

// HTTP Client
"net/http"                          // Standard library HTTP
// Consider: "github.com/hashicorp/go-retryablehttp" for retry logic

// Data Processing
"encoding/json"                     // JSON handling (stdlib)
"gopkg.in/yaml.v3"                 // YAML config files

// Terminal UI
"github.com/fatih/color"           // Colored output
"github.com/olekukonko/tablewriter" // Table formatting
// Consider: "github.com/charmbracelet/bubbles" for interactive UI

// Utilities
"github.com/google/go-cmp/cmp"     // Deep comparison for diffs
"github.com/sirupsen/logrus"       // Structured logging
```

### Testing

```go
"testing"                           // Standard testing
"github.com/stretchr/testify/assert" // Assertions
"github.com/stretchr/testify/mock"   // Mocking
"net/http/httptest"                 // HTTP testing
```

## Implementation Details

### 1. Configuration Management

```go
// internal/config/config.go
package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    API struct {
        BaseURL string `mapstructure:"base_url"`
        Source  string `mapstructure:"source"`
        Format  string `mapstructure:"format"`
        Timeout int    `mapstructure:"timeout"`
    } `mapstructure:"api"`

    Preferences struct {
        CacheDir   string `mapstructure:"cache_dir"`
        HistoryDir string `mapstructure:"history_dir"`
        LogLevel   string `mapstructure:"log_level"`
    } `mapstructure:"preferences"`
}

// Load configuration from file and environment
func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("$HOME/.radb-client")
    viper.AddConfigPath(".")

    // Environment variable overrides
    viper.SetEnvPrefix("RADB")
    viper.AutomaticEnv()

    // Defaults
    viper.SetDefault("api.base_url", "https://api.radb.net")
    viper.SetDefault("api.source", "RADB")
    viper.SetDefault("api.format", "json")
    viper.SetDefault("api.timeout", 30)

    var cfg Config
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }

    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

### 2. Secure Credential Storage

```go
// internal/config/credentials.go
package config

import (
    "encoding/json"
    "errors"
    "github.com/zalando/go-keyring"
    "golang.org/x/crypto/nacl/secretbox"
    "crypto/rand"
    "os"
    "path/filepath"
)

const (
    serviceName = "radb-client"
)

type Credentials struct {
    Username string `json:"username"`
    APIKey   string `json:"api_key"`
}

// Store credentials in system keyring (primary method)
func StoreCredentials(creds *Credentials) error {
    data, err := json.Marshal(creds)
    if err != nil {
        return err
    }

    // Try keyring first
    err = keyring.Set(serviceName, creds.Username, string(data))
    if err == nil {
        return nil
    }

    // Fallback to encrypted file
    return storeEncryptedFile(creds)
}

// Load credentials from keyring or encrypted file
func LoadCredentials(username string) (*Credentials, error) {
    // Try keyring first
    data, err := keyring.Get(serviceName, username)
    if err == nil {
        var creds Credentials
        if err := json.Unmarshal([]byte(data), &creds); err != nil {
            return nil, err
        }
        return &creds, nil
    }

    // Fallback to encrypted file
    return loadEncryptedFile(username)
}

// Encrypted file fallback implementation
func storeEncryptedFile(creds *Credentials) error {
    // Implementation with secretbox
    // Store in ~/.radb-client/credentials.enc
    return errors.New("encrypted file storage not yet implemented")
}

func loadEncryptedFile(username string) (*Credentials, error) {
    return nil, errors.New("encrypted file storage not yet implemented")
}
```

### 3. API Client

```go
// internal/api/client.go
package api

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type Client struct {
    httpClient *http.Client
    baseURL    string
    source     string
    username   string
    apiKey     string
}

func NewClient(baseURL, source, username, apiKey string) *Client {
    return &Client{
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        baseURL:  baseURL,
        source:   source,
        username: username,
        apiKey:   apiKey,
    }
}

// Execute HTTP request with authentication
func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
    var reqBody io.Reader
    if body != nil {
        data, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("marshal request: %w", err)
        }
        reqBody = bytes.NewReader(data)
    }

    url := c.baseURL + path
    req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    // Set authentication
    req.SetBasicAuth(c.username, c.apiKey)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    // Execute with retry logic (simplified - use retry library in production)
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("execute request: %w", err)
    }

    return resp, nil
}
```

### 4. State Management

```go
// internal/state/manager.go
package state

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
)

type Manager struct {
    cacheDir   string
    historyDir string
}

func NewManager(cacheDir, historyDir string) (*Manager, error) {
    // Ensure directories exist
    dirs := []string{cacheDir, historyDir}
    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0700); err != nil {
            return nil, fmt.Errorf("create directory %s: %w", dir, err)
        }
    }

    return &Manager{
        cacheDir:   cacheDir,
        historyDir: historyDir,
    }, nil
}

// Save current state snapshot
func (m *Manager) SaveSnapshot(snapshotType string, data interface{}) error {
    timestamp := time.Now().Format("2006-01-02T15-04-05")

    // Save to cache (current state)
    cachePath := filepath.Join(m.cacheDir, fmt.Sprintf("%s.json", snapshotType))
    if err := m.writeJSON(cachePath, data); err != nil {
        return fmt.Errorf("write cache: %w", err)
    }

    // Save to history (timestamped)
    historyPath := filepath.Join(m.historyDir, fmt.Sprintf("%s_%s.json", timestamp, snapshotType))
    if err := m.writeJSON(historyPath, data); err != nil {
        return fmt.Errorf("write history: %w", err)
    }

    return nil
}

// Load the most recent snapshot
func (m *Manager) LoadSnapshot(snapshotType string, target interface{}) error {
    cachePath := filepath.Join(m.cacheDir, fmt.Sprintf("%s.json", snapshotType))
    return m.readJSON(cachePath, target)
}

func (m *Manager) writeJSON(path string, data interface{}) error {
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    return encoder.Encode(data)
}

func (m *Manager) readJSON(path string, target interface{}) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    return json.NewDecoder(file).Decode(target)
}
```

### 5. CLI Structure

```go
// internal/cli/root.go
package cli

import (
    "github.com/spf13/cobra"
    "os"
)

var rootCmd = &cobra.Command{
    Use:   "radb-client",
    Short: "RADb API client for managing route objects and contacts",
    Long: `A command-line interface for the RADb API that allows you to
manage route objects, contacts, and track changes over time without
using the web interface.`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    // Global flags
    rootCmd.PersistentFlags().String("config", "", "config file (default: $HOME/.radb-client/config.yaml)")
    rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
    rootCmd.PersistentFlags().String("format", "table", "output format (table, json, yaml)")

    // Add subcommands
    rootCmd.AddCommand(configCmd)
    rootCmd.AddCommand(authCmd)
    rootCmd.AddCommand(routeCmd)
    rootCmd.AddCommand(contactCmd)
    rootCmd.AddCommand(searchCmd)
    rootCmd.AddCommand(historyCmd)
    rootCmd.AddCommand(snapshotCmd)
}
```

### 6. Domain Models

```go
// internal/models/route.go
package models

import "time"

type Route struct {
    Route        string            `json:"route"`         // IPv4/IPv6 prefix
    Origin       string            `json:"origin"`        // AS number
    Description  string            `json:"descr"`
    Maintainers  []string          `json:"mnt-by"`
    Source       string            `json:"source"`
    Created      time.Time         `json:"created,omitempty"`
    LastModified time.Time         `json:"last-modified,omitempty"`
    Attributes   map[string]string `json:"attributes,omitempty"`
}

// Key returns a unique identifier for the route
func (r *Route) Key() string {
    return r.Route + r.Origin
}

type RouteList struct {
    Routes []Route   `json:"routes"`
    Total  int       `json:"total"`
    FetchedAt time.Time `json:"fetched_at"`
}
```

## Build and Distribution

### Makefile

```makefile
.PHONY: build test clean install

# Binary name
BINARY=radb-client

# Build directory
BUILD_DIR=dist

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w"

all: test build

build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/radb-client

build-all:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/radb-client
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/radb-client
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/radb-client
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/radb-client

test:
	$(GOTEST) -v -cover ./...

test-coverage:
	$(GOTEST) -v -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.txt coverage.html

install:
	$(GOCMD) install ./cmd/radb-client

deps:
	$(GOMOD) download
	$(GOMOD) tidy

lint:
	golangci-lint run ./...

.DEFAULT_GOAL := build
```

## Testing Strategy

### Unit Test Example

```go
// internal/state/manager_test.go
package state

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestManager_SaveAndLoadSnapshot(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    cacheDir := filepath.Join(tmpDir, "cache")
    historyDir := filepath.Join(tmpDir, "history")

    manager, err := NewManager(cacheDir, historyDir)
    require.NoError(t, err)

    // Test data
    testData := map[string]string{
        "route": "192.0.2.0/24",
        "origin": "AS64500",
    }

    // Save snapshot
    err = manager.SaveSnapshot("test", testData)
    require.NoError(t, err)

    // Load snapshot
    var loaded map[string]string
    err = manager.LoadSnapshot("test", &loaded)
    require.NoError(t, err)

    // Assert
    assert.Equal(t, testData, loaded)
}
```

## Error Handling Patterns

```go
// Custom error types
type APIError struct {
    StatusCode int
    Message    string
    Details    map[string]interface{}
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// Error wrapping for context
func (c *Client) GetRoute(ctx context.Context, prefix string) (*Route, error) {
    resp, err := c.do(ctx, "GET", fmt.Sprintf("/RADB/route/%s", prefix), nil)
    if err != nil {
        return nil, fmt.Errorf("get route %s: %w", prefix, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, &APIError{
            StatusCode: resp.StatusCode,
            Message:    "failed to get route",
        }
    }

    var route Route
    if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
        return nil, fmt.Errorf("decode route response: %w", err)
    }

    return &route, nil
}
```

## Performance Considerations

1. **Connection Pooling**: Use http.Client with appropriate transport settings
2. **Concurrent Requests**: Use goroutines with context for timeouts and cancellation
3. **Caching**: Implement TTL-based caching for frequently accessed data
4. **Batch Operations**: Support bulk operations where the API allows
5. **Rate Limiting**: Implement token bucket or similar algorithm

## Next Steps for Implementation

1. Set up the basic project structure with directories
2. Initialize go.mod with proper dependencies
3. Implement configuration loading (Viper)
4. Create credential storage (keyring with fallback)
5. Build basic API client with authentication
6. Implement route listing as first command
7. Add snapshot/state management
8. Implement diff generation
9. Add remaining commands progressively
10. Write comprehensive tests

## Additional Resources

- [Cobra Documentation](https://cobra.dev/)
- [Viper Configuration](https://github.com/spf13/viper)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Effective Go](https://go.dev/doc/effective_go)
