# Implementation Roadmap

## Overview

This document outlines the phased implementation plan for the RADb API client.

## Phase 1: Foundation (MVP) - Week 1-2

### Goals
- Project structure established
- Basic configuration management working
- Authentication functional
- First API call successful

### Tasks

#### 1.1 Project Setup
- [x] Initialize Go module
- [x] Create directory structure
- [x] Setup .gitignore
- [x] Create Makefile
- [ ] Initialize git repository
- [ ] First commit

#### 1.2 Configuration Management
- [ ] Implement config loading with Viper
- [ ] Support YAML config files
- [ ] Environment variable overrides
- [ ] Config initialization command
- [ ] Config validation

**Files to create:**
- `internal/config/config.go`
- `internal/config/config_test.go`
- `config.example.yaml`

**Acceptance Criteria:**
```bash
radb-client config init  # Creates ~/.radb-client/config.yaml
radb-client config show  # Displays current configuration
radb-client config set api.timeout 60  # Updates config value
```

#### 1.3 Credential Management
- [ ] Implement keyring storage wrapper
- [ ] Encrypted file fallback
- [ ] Credential storage functions
- [ ] Credential retrieval functions
- [ ] Tests with mock keyring

**Files to create:**
- `internal/config/credentials.go`
- `internal/config/credentials_test.go`
- `pkg/keyring/keyring.go`

**Acceptance Criteria:**
- Credentials stored securely in system keyring
- Graceful fallback to encrypted file if keyring unavailable
- Never log or expose credentials

#### 1.4 API Client Foundation
- [ ] HTTP client wrapper
- [ ] Authentication (Basic Auth)
- [ ] Error handling
- [ ] Request/response logging (debug mode)
- [ ] Unit tests with httptest

**Files to create:**
- `internal/api/client.go`
- `internal/api/client_test.go`
- `internal/api/models.go`

**Acceptance Criteria:**
- Successfully authenticate with RADb API
- Handle common HTTP errors gracefully
- Retry logic for transient failures

#### 1.5 CLI Framework
- [ ] Root command setup with Cobra
- [ ] Global flags (verbose, config, format)
- [ ] Version command
- [ ] Help documentation
- [ ] Command structure

**Files to create:**
- `cmd/radb-client/main.go`
- `internal/cli/root.go`
- `internal/cli/version.go`

#### 1.6 Authentication Commands
- [ ] `auth login` - Interactive login
- [ ] `auth status` - Check authentication
- [ ] `auth logout` - Clear credentials
- [ ] Integration tests

**Files to create:**
- `internal/cli/auth.go`
- `internal/cli/auth_test.go`

**Acceptance Criteria:**
```bash
radb-client auth login
# Username: user@example.com
# API Key: ********
# Successfully authenticated

radb-client auth status
# Authenticated as: user@example.com
# Keyring: system (macOS Keychain)
```

#### 1.7 First API Operation: Route Listing
- [ ] Route list API implementation
- [ ] Route domain model
- [ ] Basic table output
- [ ] JSON output option
- [ ] Error handling

**Files to create:**
- `internal/api/routes.go`
- `internal/models/route.go`
- `internal/cli/route.go`

**Acceptance Criteria:**
```bash
radb-client route list
# Displays table of routes

radb-client route list --format json
# Outputs JSON array of routes
```

---

## Phase 2: Core Functionality - Week 3-4

### Goals
- Full CRUD for route objects
- Contact management
- State management and snapshots
- Basic diff capability

### Tasks

#### 2.1 Complete Route Operations
- [ ] `route show <prefix>` - Get single route
- [ ] `route create <file>` - Create new route
- [ ] `route update <prefix> <file>` - Update route
- [ ] `route delete <prefix>` - Delete route
- [ ] Input validation
- [ ] Batch operations support

**Acceptance Criteria:**
- All CRUD operations work correctly
- Clear error messages for validation failures
- Support for both IPv4 and IPv6 routes

#### 2.2 Contact Management
- [ ] Contact API implementation
- [ ] Contact domain model
- [ ] `contact list` command
- [ ] `contact show <id>` command
- [ ] `contact create <file>` command
- [ ] `contact update <id> <file>` command
- [ ] `contact delete <id>` command

**Files to create:**
- `internal/api/contacts.go`
- `internal/models/contact.go`
- `internal/cli/contact.go`

#### 2.3 State Management
- [ ] State manager implementation
- [ ] Snapshot creation
- [ ] Snapshot loading
- [ ] Directory structure management
- [ ] File I/O with proper error handling
- [ ] Tests

**Files to create:**
- `internal/state/manager.go`
- `internal/state/manager_test.go`
- `internal/state/snapshot.go`

**Acceptance Criteria:**
- Snapshots saved with timestamps
- Current state cached for quick access
- Historical snapshots retained

#### 2.4 Snapshot Commands
- [ ] `snapshot create` - Manual snapshot
- [ ] `snapshot list` - List all snapshots
- [ ] `snapshot show <timestamp>` - View snapshot
- [ ] `snapshot delete <timestamp>` - Delete snapshot
- [ ] Automatic snapshots on list operations

**Files to create:**
- `internal/cli/snapshot.go`

#### 2.5 Diff Generation (Basic)
- [ ] Diff algorithm implementation
- [ ] Compare two snapshots
- [ ] Detect added/removed/modified items
- [ ] `route diff` command
- [ ] Pretty output formatting

**Files to create:**
- `internal/state/diff.go`
- `internal/state/diff_test.go`

**Acceptance Criteria:**
```bash
radb-client route diff
# Shows changes since last snapshot
# + Added: 192.0.2.0/24 AS64500
# - Removed: 198.51.100.0/24 AS64501
# ~ Modified: 203.0.113.0/24 AS64502
```

---

## Phase 3: Advanced Features - Week 5-6

### Goals
- Change tracking with history
- Search capabilities
- Enhanced output formatting
- Bulk operations

### Tasks

#### 3.1 Change History
- [ ] Changelog implementation
- [ ] JSONL append-only log
- [ ] `history show` command
- [ ] Time-range filtering
- [ ] Type filtering (route/contact)

**Files to create:**
- `internal/state/history.go`
- `internal/models/changelog.go`
- `internal/cli/history.go`

**Acceptance Criteria:**
```bash
radb-client history show --since 2025-10-01
radb-client history show --type route
radb-client history show --limit 10
```

#### 3.2 Search & Validation
- [ ] Search API implementation
- [ ] ASN validation
- [ ] `search <query>` command
- [ ] `validate asn <asn>` command
- [ ] Filter support

**Files to create:**
- `internal/api/search.go`
- `internal/cli/search.go`

#### 3.3 Enhanced Output Formatting
- [ ] Table formatting with tablewriter
- [ ] Colored output with fatih/color
- [ ] YAML output option
- [ ] Pagination for large results
- [ ] Export to file

#### 3.4 Bulk Operations
- [ ] Read operations from file/stdin
- [ ] Batch processing
- [ ] Progress indicators
- [ ] Error collection and reporting
- [ ] Rollback support

**Acceptance Criteria:**
```bash
radb-client route create --batch routes.json
# Processing: [=========>    ] 65% (65/100)
# Success: 63, Failed: 2
```

#### 3.5 Advanced Diff Features
- [ ] Deep attribute comparison
- [ ] Diff between any two timestamps
- [ ] Export diff to various formats
- [ ] Summary statistics

---

## Phase 4: Polish & Production Ready - Week 7-8

### Goals
- Comprehensive testing
- Documentation
- Error handling refinement
- Performance optimization

### Tasks

#### 4.1 Testing
- [ ] Unit test coverage > 80%
- [ ] Integration tests for all commands
- [ ] Mock API server for tests
- [ ] End-to-end test scenarios
- [ ] Benchmark tests for performance

#### 4.2 Documentation
- [ ] Command reference documentation
- [ ] API usage examples
- [ ] Troubleshooting guide
- [ ] Contributing guidelines
- [ ] Architecture diagrams

**Files to create:**
- `docs/commands.md`
- `docs/examples.md`
- `docs/troubleshooting.md`
- `CONTRIBUTING.md`

#### 4.3 Error Handling & Logging
- [ ] Structured logging with logrus
- [ ] Debug mode with detailed output
- [ ] User-friendly error messages
- [ ] Retry logic with exponential backoff
- [ ] Circuit breaker for API calls

#### 4.4 Configuration Enhancements
- [ ] Config migration support
- [ ] Config validation with schemas
- [ ] Profile support (dev, prod, test)
- [ ] Interactive config wizard

#### 4.5 Performance & Optimization
- [ ] Concurrent API calls where safe
- [ ] Response caching with TTL
- [ ] Connection pooling
- [ ] Memory profiling and optimization
- [ ] Binary size optimization

#### 4.6 Distribution
- [ ] GitHub Actions for CI/CD
- [ ] Multi-platform builds
- [ ] Release automation
- [ ] Checksums and signatures
- [ ] Installation scripts

**Files to create:**
- `.github/workflows/test.yml`
- `.github/workflows/release.yml`
- `scripts/install.sh`

---

## Future Enhancements (Post-Launch)

### Advanced Features
- [ ] Interactive TUI mode (bubbletea)
- [ ] Watch mode for continuous monitoring
- [ ] Webhook notifications on changes
- [ ] Email alerts for critical changes
- [ ] Multi-account support
- [ ] Team collaboration features

### Integrations
- [ ] Export to monitoring systems
- [ ] Terraform provider
- [ ] Ansible module
- [ ] API proxy mode
- [ ] GraphQL interface

### Analytics
- [ ] Change frequency analysis
- [ ] Anomaly detection
- [ ] Reporting and dashboards
- [ ] Audit logging

---

## Success Metrics

### Phase 1 (MVP)
- [ ] Can authenticate successfully
- [ ] Can list route objects
- [ ] Configuration persists
- [ ] Basic error handling works

### Phase 2 (Core)
- [ ] All CRUD operations functional
- [ ] Snapshots saved automatically
- [ ] Diff shows changes accurately
- [ ] Contact management works

### Phase 3 (Advanced)
- [ ] Historical tracking complete
- [ ] Search returns relevant results
- [ ] Bulk operations efficient
- [ ] Output formatting excellent

### Phase 4 (Production)
- [ ] Test coverage > 80%
- [ ] Documentation complete
- [ ] Zero critical bugs
- [ ] Performance acceptable
- [ ] Ready for daily use

---

## Risk Mitigation

### Technical Risks
1. **API Changes**: Monitor API documentation, version checking
2. **Rate Limiting**: Implement proper rate limiting, respect API limits
3. **Data Corruption**: Checksums, validation, backups
4. **Authentication Issues**: Clear error messages, troubleshooting docs

### Operational Risks
1. **Breaking Changes**: Semantic versioning, migration guides
2. **Data Loss**: Regular backups, atomic operations
3. **Security**: Regular dependency updates, security audits

---

## Development Workflow

### Daily Workflow
1. Start with highest priority task from current phase
2. Write tests first (TDD when appropriate)
3. Implement feature
4. Run tests and linting
5. Update documentation
6. Commit with clear message

### Weekly Review
- Progress against roadmap
- Adjust priorities based on learnings
- Update documentation
- Review and address technical debt

### Release Cycle
- Alpha: After Phase 1 (internal testing)
- Beta: After Phase 2 (limited users)
- RC: After Phase 3 (wider testing)
- v1.0: After Phase 4 (production ready)

---

## Getting Started (Next Steps)

To begin implementation:

1. **Setup Development Environment**
   ```bash
   cd /home/bss/code/radb
   make deps
   ```

2. **Start with Phase 1.1 - 1.2**
   - Create directory structure
   - Implement configuration loading

3. **First Working Command**
   - Target: `radb-client config init`
   - This validates the entire setup

4. **Iterate Rapidly**
   - Small commits
   - Frequent testing
   - Continuous documentation

**Ready to start coding?** Begin with Phase 1.1 tasks above!
