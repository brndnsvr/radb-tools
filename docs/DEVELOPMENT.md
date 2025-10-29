# Development Guide

Guide for developing and contributing to the RADb API Client.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Building](#building)
- [Testing](#testing)
- [Code Style](#code-style)
- [Contributing Workflow](#contributing-workflow)
- [Debugging](#debugging)
- [Release Process](#release-process)

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make
- golangci-lint (for linting)

### Quick Start

```bash
# Clone the repository
git clone https://github.com/example/radb-client.git
cd radb-client

# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Run locally
./dist/radb-client --help
```

## Development Setup

### Install Go

**Linux:**
```bash
wget https://go.dev/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**macOS:**
```bash
brew install go
```

**Verify installation:**
```bash
go version
# go version go1.21.0 linux/amd64
```

### Clone Repository

```bash
git clone https://github.com/example/radb-client.git
cd radb-client
```

### Install Development Tools

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install other tools
make deps
```

### IDE Setup

**VS Code:**
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true
}
```

**GoLand/IntelliJ:**
- Go plugin should be installed automatically
- Enable golangci-lint in Settings → Go → Linter

## Project Structure

```
radb-client/
├── cmd/
│   └── radb-client/
│       └── main.go              # Entry point
├── internal/                    # Private application code
│   ├── api/                     # RADb API client
│   │   ├── client.go
│   │   ├── routes.go
│   │   └── contacts.go
│   ├── cli/                     # CLI commands
│   │   ├── root.go
│   │   ├── config.go
│   │   ├── auth.go
│   │   └── route.go
│   ├── config/                  # Configuration management
│   │   ├── config.go
│   │   └── credentials.go
│   ├── models/                  # Domain models
│   │   ├── route.go
│   │   └── contact.go
│   └── state/                   # State management
│       ├── manager.go
│       ├── snapshot.go
│       └── diff.go
├── pkg/                         # Public reusable packages
│   ├── keyring/
│   │   └── keyring.go
│   └── httpclient/
│       └── retry.go
├── testdata/                    # Test fixtures
│   ├── fixtures/
│   └── mocks/
├── docs/                        # Documentation
├── scripts/                     # Utility scripts
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── Makefile                     # Build automation
├── README.md
└── LICENSE
```

### Key Directories

- **cmd/** - Application entry points
- **internal/** - Private packages (cannot be imported by other projects)
- **pkg/** - Public packages (can be imported by other projects)
- **testdata/** - Test data and fixtures

## Building

### Build Commands

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Install locally
make install

# Clean build artifacts
make clean
```

### Build Targets

The Makefile supports multiple build targets:

```bash
# Development build (with debug info)
make build

# Production build (optimized)
make build LDFLAGS="-s -w"

# Specific platform
GOOS=linux GOARCH=amd64 make build
```

### Cross-Compilation

Build for multiple platforms:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o dist/radb-client-linux-amd64 ./cmd/radb-client

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o dist/radb-client-darwin-amd64 ./cmd/radb-client

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o dist/radb-client-darwin-arm64 ./cmd/radb-client

# Windows
GOOS=windows GOARCH=amd64 go build -o dist/radb-client-windows-amd64.exe ./cmd/radb-client
```

## Testing

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test ./internal/api/...

# Verbose
go test -v ./...

# Run specific test
go test -run TestClient_GetRoute ./internal/api/
```

### Writing Tests

**Example unit test:**

```go
// internal/state/manager_test.go
package state

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestManager_SaveSnapshot(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    manager, err := NewManager(tmpDir+"/cache", tmpDir+"/history")
    require.NoError(t, err)

    // Test data
    testData := map[string]string{
        "route": "192.0.2.0/24",
        "origin": "AS64500",
    }

    // Execute
    err = manager.SaveSnapshot("test", testData)

    // Assert
    require.NoError(t, err)

    // Verify snapshot exists
    var loaded map[string]string
    err = manager.LoadSnapshot("test", &loaded)
    require.NoError(t, err)
    assert.Equal(t, testData, loaded)
}
```

**Example integration test:**

```go
// internal/api/client_test.go
package api

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestClient_GetRoute(t *testing.T) {
    // Mock API server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"route":"192.0.2.0/24","origin":"AS64500"}`))
    }))
    defer server.Close()

    // Create client
    client := NewClient(server.URL, "RADB", "user", "key")

    // Test
    route, err := client.GetRoute(context.Background(), "192.0.2.0/24")
    require.NoError(t, err)
    assert.Equal(t, "192.0.2.0/24", route.Route)
    assert.Equal(t, "AS64500", route.Origin)
}
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
go tool cover -html=coverage.txt
```

**Coverage goals:**
- Overall: > 80%
- Critical paths (auth, API client): > 90%
- CLI commands: > 70%

## Code Style

### Go Conventions

Follow standard Go conventions:

- Use `gofmt` for formatting
- Use `golint` for style checking
- Follow [Effective Go](https://go.dev/doc/effective_go)

### Code Formatting

```bash
# Format all files
gofmt -w .

# Or use goimports (preferred)
goimports -w .

# Check formatting
test -z $(gofmt -l .)
```

### Linting

```bash
# Run golangci-lint
make lint

# Or directly
golangci-lint run ./...

# Auto-fix issues
golangci-lint run --fix ./...
```

### Naming Conventions

**Packages:**
- Short, lowercase names
- No underscores

**Functions:**
- CamelCase
- Start with verb: `GetRoute`, `CreateSnapshot`

**Variables:**
- CamelCase
- Descriptive names
- Avoid single letters except in loops

**Constants:**
- CamelCase or UPPER_SNAKE_CASE

### Error Handling

Always wrap errors with context:

```go
if err != nil {
    return fmt.Errorf("failed to create route: %w", err)
}
```

### Documentation

**Package documentation:**
```go
// Package api provides a client for the RADb API.
//
// It handles authentication, request/response serialization,
// and error handling for all RADb API operations.
package api
```

**Function documentation:**
```go
// GetRoute retrieves a route object by prefix.
//
// The prefix must be in CIDR notation (e.g., "192.0.2.0/24").
// Returns ErrNotFound if the route doesn't exist.
func (c *Client) GetRoute(ctx context.Context, prefix string) (*Route, error) {
    // Implementation
}
```

## Contributing Workflow

### 1. Fork and Clone

```bash
# Fork on GitHub, then:
git clone https://github.com/YOUR_USERNAME/radb-client.git
cd radb-client

# Add upstream remote
git remote add upstream https://github.com/example/radb-client.git
```

### 2. Create Branch

```bash
# Update main
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
```

Branch naming:
- Features: `feature/feature-name`
- Bugs: `fix/bug-description`
- Docs: `docs/improvement-description`

### 3. Make Changes

```bash
# Make your changes
vim internal/api/client.go

# Add tests
vim internal/api/client_test.go

# Run tests
make test

# Format code
make fmt

# Lint
make lint
```

### 4. Commit Changes

```bash
# Stage changes
git add .

# Commit with clear message
git commit -m "Add support for IPv6 route validation"
```

**Commit message guidelines:**
- Use present tense: "Add feature" not "Added feature"
- First line: concise summary (50 chars or less)
- Blank line, then detailed description if needed
- Reference issues: "Fixes #123"

**Example:**
```
Add IPv6 route validation

Implements validation for IPv6 routes with proper CIDR
notation checking. Also adds unit tests for edge cases.

Fixes #123
```

### 5. Push and Create PR

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create pull request on GitHub
```

**PR checklist:**
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] All tests passing
- [ ] Code linted
- [ ] No conflicts with main

### 6. Address Review Comments

```bash
# Make requested changes
vim internal/api/client.go

# Commit
git commit -m "Address review comments"

# Push
git push origin feature/your-feature-name
```

## Debugging

### Enable Debug Logging

```bash
# Run with debug logging
go run ./cmd/radb-client --log-level DEBUG route list

# Or built binary
./radb-client --log-level DEBUG route list
```

### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug ./cmd/radb-client -- route list

# Set breakpoints
(dlv) break main.main
(dlv) break internal/api.(*Client).GetRoute

# Continue
(dlv) continue

# Inspect variables
(dlv) print variableName
```

### VS Code Debugging

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug radb-client",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/radb-client",
      "args": ["route", "list", "--log-level", "DEBUG"]
    }
  ]
}
```

### Profiling

**CPU profiling:**
```go
import _ "net/http/pprof"

// In main.go
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

**Run and profile:**
```bash
# Run with profiling
./radb-client route list

# Capture profile
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

## Release Process

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):
- MAJOR.MINOR.PATCH
- Example: 1.2.3

**Increment:**
- MAJOR: Breaking changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes

### Creating a Release

1. **Update version:**
   ```bash
   # In internal/version/version.go
   const Version = "1.2.3"
   ```

2. **Update CHANGELOG.md:**
   ```markdown
   ## [1.2.3] - 2025-10-29

   ### Added
   - New feature X

   ### Fixed
   - Bug Y
   ```

3. **Commit and tag:**
   ```bash
   git commit -am "Release v1.2.3"
   git tag v1.2.3
   git push origin main --tags
   ```

4. **Build release binaries:**
   ```bash
   make build-all
   ```

5. **Create GitHub release:**
   - Go to GitHub Releases
   - Create new release from tag
   - Upload binaries
   - Copy changelog

### Automated Releases

Use GitHub Actions:

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: make build-all
      - uses: softprops/action-gh-release@v1
        with:
          files: dist/*
```

## Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Project Layout](https://github.com/golang-standards/project-layout)

## See Also

- [Architecture](ARCHITECTURE.md) - System architecture
- [Contributing](../CONTRIBUTING.md) - Contribution guidelines
- [Security](SECURITY.md) - Security practices
