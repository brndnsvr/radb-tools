# Quick Start Guide

This guide will help you understand the project structure and start developing.

## What You Have Now

A complete design and architecture for a RADb API client in Go with:

✅ **Design Documents**
- `DESIGN.md` - Overall architecture and requirements
- `GO_IMPLEMENTATION.md` - Go-specific implementation details
- `ROADMAP.md` - Phased implementation plan

✅ **Project Files**
- `go.mod` - Go module definition
- `Makefile` - Build automation
- `config.example.yaml` - Configuration template
- `.gitignore` - Git ignore rules
- `LICENSE` - MIT License

✅ **Documentation**
- `README.md` - Project overview
- This guide!

## Understanding the Architecture

### Key Concepts

1. **State Management**: The client stores snapshots of your RADb data locally
   - Current state in `~/.radb-client/cache/`
   - Historical snapshots in `~/.radb-client/history/`
   - Enables diff detection between runs

2. **Secure Credentials**: Uses system keyring (macOS Keychain, Windows Credential Manager, Linux Secret Service)
   - Falls back to encrypted file if keyring unavailable
   - Never logs or exposes credentials

3. **Change Tracking**: Automatically detects and logs changes
   - Compare current state with previous snapshots
   - JSONL changelog for audit trail
   - `route diff` command to see what changed

4. **CLI-First Design**: Intuitive command structure
   ```bash
   radb-client config init           # Setup
   radb-client auth login            # Authenticate
   radb-client route list            # List routes
   radb-client route diff            # See changes
   ```

### Project Structure (To Be Created)

```
radb-client/
├── cmd/
│   └── radb-client/
│       └── main.go                 # Entry point
├── internal/                       # Private application code
│   ├── api/                        # RADb API client
│   ├── cli/                        # Command implementations
│   ├── config/                     # Configuration & credentials
│   ├── models/                     # Domain models
│   └── state/                      # State & snapshot management
├── pkg/                            # Public reusable packages
│   ├── httpclient/                 # HTTP utilities
│   └── keyring/                    # Keyring wrapper
├── testdata/                       # Test fixtures
├── docs/                           # Additional documentation
└── [project files]                 # go.mod, Makefile, etc.
```

## What the Tool Does

### Primary Use Cases

**1. Manage Route Objects**
```bash
# List all your routes
radb-client route list

# View specific route
radb-client route show 192.0.2.0/24

# Create new route from file
radb-client route create route.json

# Update existing route
radb-client route update 192.0.2.0/24 updated-route.json

# Delete route
radb-client route delete 192.0.2.0/24
```

**2. Manage Contacts**
```bash
radb-client contact list
radb-client contact create contact.json
radb-client contact update contact-id updated.json
```

**3. Track Changes**
```bash
# Run this periodically to detect changes
radb-client route list
radb-client route diff

# Output shows what changed since last run:
# + Added: 192.0.2.0/24 AS64500
# - Removed: 198.51.100.0/24 AS64501
# ~ Modified: 203.0.113.0/24 AS64502
```

**4. Historical Analysis**
```bash
# View change history
radb-client history show --since 2025-10-01

# Compare two specific snapshots
radb-client history diff 2025-10-29T12:00:00 2025-10-29T18:00:00

# List all snapshots
radb-client snapshot list
```

## Data Storage Strategy

### Why Local Storage?

The client maintains local state to:
1. **Detect changes** between runs without complex queries
2. **Provide instant diffs** without hitting the API
3. **Maintain history** for audit and analysis
4. **Work offline** for viewing cached data

### What Gets Stored?

**Current State** (`~/.radb-client/cache/`)
```
route_objects.json       # Latest route list
contacts.json            # Latest contacts
metadata.json           # Last fetch timestamps
```

**Historical Snapshots** (`~/.radb-client/history/`)
```
2025-10-29T12-00-00_route_objects.json
2025-10-29T18-00-00_route_objects.json
changelog.jsonl                          # Append-only change log
```

**Example Changelog Entry**
```json
{"timestamp":"2025-10-29T12:00:00Z","type":"route","action":"added","object_id":"192.0.2.0/24AS64500","details":{...}}
```

## RADb API Integration

### Authentication
- HTTP Basic Auth with username and API key
- Stored securely in system keyring
- No need to re-enter credentials

### Key Endpoints Used
```
GET  /{source}/search              # Search database
GET  /{source}/route/{prefix}      # Get route object
POST /{source}/route               # Create route
PUT  /{source}/route/{prefix}      # Update route
DELETE /{source}/route/{prefix}    # Delete route
GET  /{source}/validate/asn/{asn}  # Validate ASN
```

### Data Format
- Supports both JSON and text formats
- Client prefers JSON for structured data
- Includes all required attributes (mnt-by, origin, source)

## Implementation Approach

### Phase 1: MVP (Start Here)
1. **Config Management** - Load/save YAML configuration
2. **Authentication** - Store credentials in keyring
3. **API Client** - Basic HTTP client with auth
4. **First Command** - `radb-client route list`
5. **State Storage** - Save snapshots locally

**Why this order?** Each builds on the previous, giving you working functionality quickly.

### Phase 2: Core Features
- Complete CRUD for routes and contacts
- Diff generation
- Snapshot management

### Phase 3: Advanced
- Change tracking with history
- Search capabilities
- Bulk operations

### Phase 4: Polish
- Comprehensive testing
- Documentation
- Performance optimization

## Next Steps: Ready to Code?

### Step 1: Initialize Git Repository
```bash
cd /home/bss/code/radb
git init
git add .
git commit -m "Initial project setup and design documents"
```

### Step 2: Create Directory Structure
```bash
mkdir -p cmd/radb-client
mkdir -p internal/{api,cli,config,models,state}
mkdir -p pkg/{httpclient,keyring}
mkdir -p testdata/{fixtures,mocks}
mkdir -p docs
```

### Step 3: Start with Configuration
Create `internal/config/config.go` first - this is the foundation.

**Why config first?** Everything else depends on it:
- API client needs base URL
- State manager needs directory paths
- CLI needs preferences

### Step 4: Test Early and Often
```bash
make test        # Run tests
make build       # Build binary
make help        # See all commands
```

### Development Workflow

**TDD Approach (Recommended)**
1. Write test first (what should it do?)
2. Implement feature (make test pass)
3. Refactor (clean up code)
4. Commit (small, focused commits)

**Example: First Feature**
```bash
# 1. Create test
vim internal/config/config_test.go

# 2. Implement
vim internal/config/config.go

# 3. Test
go test ./internal/config/...

# 4. Commit
git add internal/config/
git commit -m "Add configuration loading with Viper"
```

## Key Design Decisions

### Why Go?
- **Single binary**: Easy distribution, no dependencies
- **Performance**: Fast execution, low memory
- **Strong typing**: Catch errors at compile time
- **Great stdlib**: Excellent HTTP, JSON, and file handling

### Why Cobra + Viper?
- **Cobra**: Industry standard for CLI tools
- **Viper**: Flexible configuration (files, env vars, flags)
- **Well documented**: Lots of examples and support

### Why Local State?
- **Fast diffs**: No need to query API twice
- **Offline capability**: View cached data without network
- **Historical tracking**: Maintain audit trail
- **Change detection**: Immediate visibility into what changed

### Security Priorities
1. Credentials never logged or printed
2. System keyring for secure storage
3. HTTPS only (no HTTP fallback)
4. Encrypted file fallback for keyring
5. Audit trail of all operations

## Common Patterns

### Error Handling
```go
if err != nil {
    return fmt.Errorf("descriptive context: %w", err)
}
```

### API Calls
```go
resp, err := client.do(ctx, "GET", "/RADB/route/192.0.2.0/24", nil)
if err != nil {
    return fmt.Errorf("get route: %w", err)
}
defer resp.Body.Close()
```

### State Management
```go
// Save snapshot automatically
if err := stateManager.SaveSnapshot("route_objects", routes); err != nil {
    log.Warnf("Failed to save snapshot: %v", err)
}
```

### CLI Commands
```go
var routeListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all route objects",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
        return nil
    },
}
```

## Testing Strategy

### Unit Tests
- Test individual functions and methods
- Mock external dependencies (API, filesystem, keyring)
- Fast execution

### Integration Tests
- Test multiple components together
- Use test API or mock server
- Verify end-to-end workflows

### Example Test
```go
func TestManager_SaveSnapshot(t *testing.T) {
    manager, _ := NewManager(t.TempDir(), t.TempDir())

    data := map[string]string{"test": "data"}
    err := manager.SaveSnapshot("test", data)

    assert.NoError(t, err)
    assert.FileExists(t, manager.cacheDir + "/test.json")
}
```

## Resources

### Go Learning
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

### Libraries Documentation
- [Cobra](https://cobra.dev/)
- [Viper](https://github.com/spf13/viper)
- [Go Keyring](https://github.com/zalando/go-keyring)

### RADb API
- [API Documentation](https://api.radb.net/docs.html)
- [OpenAPI Spec](https://api.radb.net/RADB_API_OpenAPI.yaml)

## Questions to Consider

Before starting implementation, think about:

1. **Snapshot Retention**: How many historical snapshots to keep? (Default: 100)
2. **Auto-snapshot**: Create snapshot on every list operation? (Default: yes)
3. **Output Format**: Default to table or JSON? (Default: table)
4. **Notification**: Alert on changes? (Future feature)
5. **Concurrency**: How many parallel API calls? (Start with 1, add later)

## Success Criteria

### Phase 1 Complete When:
- [x] Configuration loads from file
- [x] Credentials stored in keyring
- [x] Can authenticate with API
- [x] `route list` works
- [x] Snapshots saved automatically

### Ready for Daily Use When:
- [ ] All CRUD operations work
- [ ] Diff shows accurate changes
- [ ] Error messages are helpful
- [ ] Tests pass consistently
- [ ] Documentation complete

## Getting Help

If you encounter issues:

1. Check the design docs (DESIGN.md, GO_IMPLEMENTATION.md)
2. Review the roadmap (ROADMAP.md)
3. Look at example implementations (will be added as code grows)
4. RADb API documentation

## Let's Build!

You now have everything needed to start implementation:

- ✅ Complete design and architecture
- ✅ Technology stack chosen (Go)
- ✅ Project structure defined
- ✅ Implementation roadmap
- ✅ Development workflow

**Ready to create your first file?**

Start with `internal/config/config.go` and work through Phase 1 of the roadmap!

Good luck! 🚀
