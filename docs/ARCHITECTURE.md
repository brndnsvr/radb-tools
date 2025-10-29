# Architecture Documentation

Technical architecture and design of the RADb API Client.

## Table of Contents

- [System Overview](#system-overview)
- [Architecture Diagram](#architecture-diagram)
- [Component Architecture](#component-architecture)
- [Data Flow](#data-flow)
- [Design Decisions](#design-decisions)
- [Technology Stack](#technology-stack)
- [Security Architecture](#security-architecture)
- [Performance Considerations](#performance-considerations)
- [Future Enhancements](#future-enhancements)

## System Overview

The RADb API Client is a command-line tool built in Go that provides a user-friendly interface to the RADb (Routing Assets Database) API. It enables users to manage route objects and contacts programmatically while maintaining local state for change tracking.

### Key Features

1. **API Client** - HTTP client for RADb API with authentication
2. **State Management** - Local snapshots for change detection
3. **Change Tracking** - Automatic diff generation between runs
4. **Secure Credentials** - System keyring integration
5. **CLI Interface** - Intuitive command structure

### Architecture Goals

- **Simplicity** - Easy to use and understand
- **Security** - Secure credential storage
- **Reliability** - Robust error handling and retry logic
- **Performance** - Efficient API usage and local caching
- **Maintainability** - Clean code structure and testing

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    RADb API Client                      │
└─────────────────────────────────────────────────────────┘

┌─────────────┐
│     CLI     │  User Interface
│  (Cobra)    │  - Commands and flags
└──────┬──────┘  - Input validation
       │         - Output formatting
       │
┌──────▼──────────────────────────────────────────────────┐
│               Core Application Layer                     │
├──────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Config Mgr   │  │  Auth Mgr    │  │  State Mgr   │  │
│  │              │  │              │  │              │  │
│  │ - Load config│  │ - Credentials│  │ - Snapshots  │  │
│  │ - Validation │  │ - Keyring    │  │ - Diff gen   │  │
│  │ - Persistence│  │ - Session    │  │ - History    │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└──────┬────────────────────┬────────────────────┬────────┘
       │                    │                    │
┌──────▼────────────────────▼────────────────────▼────────┐
│                    API Client Layer                      │
├──────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ HTTP Client  │  │ Route Ops    │  │ Contact Ops  │  │
│  │              │  │              │  │              │  │
│  │ - Auth       │  │ - CRUD       │  │ - CRUD       │  │
│  │ - Retry      │  │ - Validation │  │ - Validation │  │
│  │ - Rate limit │  │ - Search     │  │ - Search     │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└──────────────────────────┬──────────────────────────────┘
                           │
                ┌──────────▼──────────┐
                │     RADb API        │
                │  (HTTPS/JSON/REST)  │
                └─────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                    Data Storage                          │
├─────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Config     │  │  Credentials │  │   Snapshots  │  │
│  │              │  │              │  │              │  │
│  │ config.yaml  │  │ Keyring/     │  │ cache/       │  │
│  │              │  │ Encrypted    │  │ history/     │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. CLI Layer (internal/cli)

**Purpose:** User interface and command handling

**Components:**
- Root command (radb-client)
- Subcommands (config, auth, route, contact, etc.)
- Flag parsing and validation
- Output formatting

**Key Files:**
- `root.go` - Root command and global flags
- `config.go` - Configuration commands
- `auth.go` - Authentication commands
- `route.go` - Route management commands
- `contact.go` - Contact management commands

**Responsibilities:**
- Parse command-line arguments
- Validate user input
- Call appropriate business logic
- Format and display output
- Handle errors gracefully

**Example:**
```go
var routeListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all route objects",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Get configuration
        cfg, err := config.Load()
        
        // 2. Create API client
        client := api.NewClient(cfg)
        
        // 3. Fetch routes
        routes, err := client.ListRoutes(context.Background())
        
        // 4. Save snapshot
        stateMgr.SaveSnapshot("routes", routes)
        
        // 5. Format output
        display.PrintRoutes(routes)
        
        return nil
    },
}
```

### 2. Configuration Manager (internal/config)

**Purpose:** Manage application configuration

**Components:**
- Config loading (YAML, env vars)
- Config validation
- Default values
- Persistence

**Key Files:**
- `config.go` - Configuration structure and loading
- `credentials.go` - Credential storage/retrieval

**Responsibilities:**
- Load configuration from file and environment
- Validate configuration values
- Provide configuration to other components
- Save configuration changes

**Data Structure:**
```go
type Config struct {
    API struct {
        BaseURL string
        Source  string
        Format  string
        Timeout int
    }
    Preferences struct {
        CacheDir     string
        HistoryDir   string
        LogLevel     string
        MaxSnapshots int
        AutoSnapshot bool
        OutputFormat string
        Color        bool
    }
}
```

### 3. Authentication Manager (internal/config)

**Purpose:** Secure credential storage and management

**Components:**
- Keyring integration
- Encrypted file fallback
- Session management

**Responsibilities:**
- Store credentials securely
- Retrieve credentials
- Validate authentication
- Handle keyring errors

**Flow:**
```
User Login
    ↓
Try System Keyring
    ├─ Success → Store in keyring
    └─ Failure → Store in encrypted file
    
Retrieve Credentials
    ↓
Try System Keyring
    ├─ Success → Return credentials
    └─ Failure → Read encrypted file
```

### 4. API Client (internal/api)

**Purpose:** Interface with RADb API

**Components:**
- HTTP client wrapper
- Request/response handling
- Authentication
- Retry logic
- Rate limiting

**Key Files:**
- `client.go` - Base HTTP client
- `routes.go` - Route operations
- `contacts.go` - Contact operations
- `search.go` - Search operations

**Responsibilities:**
- Execute HTTP requests
- Handle authentication
- Retry on failures
- Respect rate limits
- Parse responses
- Handle API errors

**Request Flow:**
```
1. Prepare request (URL, method, body)
2. Add authentication headers
3. Execute request
4. Check rate limits
5. Handle errors (retry if needed)
6. Parse response
7. Return data
```

### 5. State Manager (internal/state)

**Purpose:** Local state and snapshot management

**Components:**
- Snapshot creation
- Snapshot storage
- Diff generation
- History tracking

**Key Files:**
- `manager.go` - State management
- `snapshot.go` - Snapshot operations
- `diff.go` - Diff generation
- `history.go` - Change history

**Responsibilities:**
- Save current state snapshots
- Load previous snapshots
- Generate diffs between snapshots
- Maintain change history
- Cleanup old snapshots

**Snapshot Structure:**
```json
{
  "timestamp": "2025-10-29T12:00:00Z",
  "type": "routes",
  "data": {
    "routes": [ ... ],
    "total": 42
  },
  "checksum": "sha256:..."
}
```

### 6. Domain Models (internal/models)

**Purpose:** Business logic and data structures

**Components:**
- Route model
- Contact model
- Snapshot model
- Changelog model

**Key Files:**
- `route.go` - Route business logic
- `contact.go` - Contact business logic
- `snapshot.go` - Snapshot model

**Responsibilities:**
- Define data structures
- Validation logic
- Business rules
- Data transformations

## Data Flow

### Route Listing Flow

```
User: radb-client route list
    ↓
CLI: Parse command and flags
    ↓
Config: Load configuration
    ↓
Auth: Retrieve credentials
    ↓
API Client: GET /RADB/route
    ↓
RADb API: Return routes
    ↓
State Manager: Save snapshot
    ↓
CLI: Format and display output
    ↓
User: See route table
```

### Change Detection Flow

```
User: radb-client route diff
    ↓
CLI: Parse command
    ↓
State Manager: Load current snapshot
    ↓
State Manager: Load previous snapshot
    ↓
State Manager: Generate diff
    ↓
CLI: Format and display diff
    ↓
User: See changes
```

### Route Creation Flow

```
User: radb-client route create route.json
    ↓
CLI: Parse command and file
    ↓
Validate: Check route format
    ↓
API Client: POST /RADB/route
    ↓
RADb API: Create route
    ↓
State Manager: Update snapshot
    ↓
CLI: Confirm success
    ↓
User: Route created
```

## Design Decisions

### 1. Why Go?

**Chosen:** Go

**Reasons:**
- Single binary distribution (no dependencies)
- Fast execution and low memory usage
- Strong standard library for HTTP and JSON
- Excellent cross-compilation support
- Good testing framework

**Alternatives Considered:**
- Python: Requires interpreter, dependencies
- Rust: Steeper learning curve, longer compile times
- TypeScript/Node.js: Requires runtime

### 2. Why Cobra for CLI?

**Chosen:** Cobra framework

**Reasons:**
- Industry standard for Go CLI tools
- Excellent documentation
- Built-in help generation
- Subcommand support
- Flag parsing

**Alternatives Considered:**
- urfave/cli: Less features
- Custom: Too much work

### 3. Why Local State?

**Chosen:** Local filesystem storage

**Reasons:**
- Fast diff generation without API calls
- Offline access to cached data
- Historical tracking and audit trail
- No dependency on external database

**Tradeoffs:**
- Disk space usage (mitigated by cleanup)
- Not shared between machines (acceptable)

### 4. Why System Keyring?

**Chosen:** System keyring with encrypted file fallback

**Reasons:**
- Most secure option available
- Native OS integration
- No credentials in plain text
- Fallback for headless systems

**Implementation:**
- zalando/go-keyring library
- NaCl secretbox for encrypted file

### 5. Why JSON for API?

**Chosen:** JSON format

**Reasons:**
- Structured data
- Easy to parse
- Better for automation
- API supports it well

**Alternative:** RPSL text format (supported but not default)

## Technology Stack

### Core Dependencies

**CLI Framework:**
- github.com/spf13/cobra - Command structure
- github.com/spf13/viper - Configuration management

**Security:**
- github.com/zalando/go-keyring - System keyring access
- golang.org/x/crypto - Encryption

**HTTP Client:**
- net/http - Standard library HTTP client

**Data Processing:**
- encoding/json - JSON handling
- gopkg.in/yaml.v3 - YAML configuration

**Terminal UI:**
- github.com/fatih/color - Colored output
- github.com/olekukonko/tablewriter - Table formatting

**Testing:**
- testing - Standard testing
- github.com/stretchr/testify - Assertions and mocking

### No External Dependencies for Core Logic

The core logic intentionally uses minimal external dependencies:
- No ORM or database driver
- No complex frameworks
- Standard library when possible

## Security Architecture

### Threat Model

**Threats:**
1. Credential theft
2. Man-in-the-middle attacks
3. Unauthorized access
4. Data tampering

**Mitigations:**
1. Keyring storage + encryption
2. HTTPS only, certificate validation
3. Proper authentication, no caching of credentials
4. Snapshot checksums

### Security Layers

**Layer 1: Transport Security**
- HTTPS only (no HTTP fallback)
- Certificate validation (verify_ssl: true)
- TLS 1.2 minimum

**Layer 2: Authentication**
- HTTP Basic Auth
- Credentials never logged
- Session management

**Layer 3: Credential Storage**
- System keyring (primary)
- Encrypted file (fallback)
- NaCl secretbox encryption

**Layer 4: Data Integrity**
- Snapshot checksums
- Config validation
- Input sanitization

### Security Best Practices

1. **Never log credentials**
2. **Always use HTTPS**
3. **Validate all inputs**
4. **Encrypt sensitive data at rest**
5. **Minimal privilege principle**

## Performance Considerations

### Optimization Strategies

**1. Connection Pooling**
- Reuse HTTP connections
- Configurable pool size
- Reduces overhead

**2. Local Caching**
- Cache API responses
- Reduce redundant calls
- Faster diff generation

**3. Efficient Serialization**
- Use encoding/json (fast)
- Minimize allocations
- Streaming for large datasets

**4. Rate Limiting**
- Respect API limits
- Exponential backoff
- Request queuing

**5. Concurrent Operations**
- Goroutines for parallel requests
- Context for cancellation
- Worker pools for bulk operations

### Performance Metrics

**Target Performance:**
- Startup time: < 100ms
- Route list: < 2s (100 routes)
- Diff generation: < 500ms
- Memory usage: < 50MB

**Benchmarking:**
```bash
# CPU profiling
go test -bench=. -cpuprofile=cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof

# Benchmarks
go test -bench=. ./...
```

## Future Enhancements

### Phase 1 (v2.0)
- Interactive TUI mode (bubbletea)
- Watch mode for continuous monitoring
- Plugin system for extensions

### Phase 2 (v3.0)
- Multi-account support
- Team collaboration features
- Web dashboard

### Phase 3 (v4.0)
- Terraform provider
- Ansible module
- API proxy mode

### Extensibility Points

**1. Output Formatters**
- Custom formatters can be added
- Interface-based design

**2. Storage Backends**
- Pluggable storage (filesystem, S3, etc.)
- Abstract storage interface

**3. Authentication Methods**
- Support additional auth methods
- OAuth2, JWT, etc.

**4. Notification Channels**
- Webhook support
- Email, Slack, etc.
- Event-driven architecture

## See Also

- [Development](DEVELOPMENT.md) - Development guide
- [Security](SECURITY.md) - Security details
- [Design Document](../DESIGN.md) - Original design
- [Implementation Guide](../GO_IMPLEMENTATION.md) - Go specifics
