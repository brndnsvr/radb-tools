# RADb API Client - Design Document

## Project Overview

A well-crafted client application for interacting with the RADb (Routing Assets Database) API to manage Account Contacts and route objects without requiring web UI access.

## Core Requirements

### Primary Use Cases
1. **Account Contact Management** - Manage account contacts programmatically
2. **Route Object Management** - Create, read, update, delete route/route6 objects
3. **Change Tracking** - Store historical state to detect changes between runs
4. **Authentication** - Secure credential management and session handling

## API Understanding

### Authentication Methods
- **HTTP Basic Auth** - Username/password for API access
- **API Key Auth** - Alternative password-based authentication
- **Crypted Password** - Required for most write operations

### Key Endpoints
- `/{source}/search` - Search IRR database (filtering by attributes)
- Route object CRUD operations (IPv4 and IPv6)
- `/{source}/validate/asn/{asn}` - ASN validation
- Currently supports `RADB` as source

### Data Formats
- Input: Text or JSON
- Output: Text or JSON
- Object classes: route, route6, as-set, aut-num, etc.
- Required attributes: `mnt-by`, `origin`, `source`

## Architecture Design

### 1. Configuration Management

**Local Storage Needs:**
- Credentials (encrypted at rest)
- API endpoint configuration
- Default source (RADB)
- Preferred data format (JSON)

**Configuration File Format:**
```yaml
api:
  base_url: https://api.radb.net
  source: RADB
  format: json
  timeout: 30

credentials:
  # Stored encrypted or via system keyring
  username: <stored-securely>
  api_key: <stored-securely>

preferences:
  cache_dir: ~/.radb-client/cache
  history_dir: ~/.radb-client/history
  log_level: INFO
```

### 2. State & History Management

**Purpose:** Track changes between runs for route objects and contacts

**Storage Strategy:**
```
~/.radb-client/
├── cache/
│   ├── route_objects.json       # Latest state
│   ├── contacts.json             # Latest contacts
│   └── metadata.json             # Last fetch timestamps
├── history/
│   ├── 2025-10-29T12-00-00_route_objects.json
│   ├── 2025-10-29T12-00-00_contacts.json
│   └── changelog.jsonl          # Structured change log
└── config.yaml
```

**Change Detection:**
- Store snapshot of each query result with timestamp
- Compare current state with previous snapshot
- Generate diff report showing:
  - Added route objects
  - Removed route objects
  - Modified attributes
  - Timestamp of change detection

**Changelog Format (JSONL):**
```json
{"timestamp": "2025-10-29T12:00:00Z", "type": "route", "action": "added", "object_id": "192.0.2.0/24AS64500", "details": {...}}
{"timestamp": "2025-10-29T13:00:00Z", "type": "route", "action": "modified", "object_id": "198.51.100.0/24AS64501", "changes": {...}}
```

### 3. Core Components

#### A. Authentication Manager
- Secure credential storage (keyring/encrypted file)
- Session management
- Token/credential refresh if needed
- Support multiple auth methods

#### B. API Client
- HTTP client wrapper with retry logic
- Rate limiting awareness
- Error handling and detailed logging
- Request/response validation
- Support for both text and JSON formats

#### C. Route Object Manager
- CRUD operations for route and route6 objects
- Validation before submission
- Bulk operations support
- Conflict detection

#### D. Contact Manager
- List, create, update, delete contacts
- Contact validation
- Role-based contact management

#### E. State Manager
- Snapshot creation and storage
- Diff generation
- History querying
- Cleanup of old snapshots

#### F. CLI Interface
- Intuitive command structure
- Interactive mode for complex operations
- Batch operation support
- Output formatting (table, JSON, YAML)

### 4. Command Structure

```bash
# Configuration
radb-client config init
radb-client config set <key> <value>
radb-client config show

# Authentication
radb-client auth login
radb-client auth status
radb-client auth logout

# Route Management
radb-client route list [--filter <criteria>]
radb-client route show <prefix>/<asn>
radb-client route create <file>
radb-client route update <prefix>/<asn> <file>
radb-client route delete <prefix>/<asn>
radb-client route diff [--since <timestamp>]

# Contact Management
radb-client contact list
radb-client contact show <contact-id>
radb-client contact create <file>
radb-client contact update <contact-id> <file>
radb-client contact delete <contact-id>

# Search & Discovery
radb-client search <query> [--type route|contact|asn]
radb-client validate asn <asn>

# History & Tracking
radb-client history show [--type route|contact] [--since <date>]
radb-client history diff <timestamp1> <timestamp2>
radb-client snapshot create [--note "description"]
radb-client snapshot list
```

### 5. Data Models

#### Route Object
```python
class RouteObject:
    route: str                    # IPv4/IPv6 prefix
    origin: str                   # AS number
    descr: str                    # Description
    mnt_by: List[str]            # Maintainer(s)
    source: str                   # Database source
    created: datetime
    last_modified: datetime
    attributes: Dict[str, Any]   # Additional attributes
```

#### Contact
```python
class Contact:
    id: str
    name: str
    email: str
    phone: Optional[str]
    role: str                     # admin, tech, billing
    organization: Optional[str]
    created: datetime
    last_modified: datetime
```

#### Snapshot
```python
class Snapshot:
    timestamp: datetime
    snapshot_type: str           # 'route', 'contact', 'full'
    data: Dict[str, Any]
    checksum: str               # For integrity verification
    note: Optional[str]
```

### 6. Error Handling Strategy

**API Errors:**
- Connection failures → retry with exponential backoff
- Authentication errors → prompt for re-authentication
- Rate limiting → respect limits, queue requests
- Validation errors → detailed error messages with suggestions

**Local Storage Errors:**
- Disk full → warning and cleanup suggestion
- Corrupted data → fallback to previous snapshot
- Permission errors → clear instructions

### 7. Security Considerations

**Credential Storage:**
- Use system keyring (keyring library) when available
- Fallback to encrypted file (Fernet encryption)
- Never log credentials
- Support environment variables for CI/CD

**API Communication:**
- Always use HTTPS
- Validate SSL certificates
- Timeout configuration
- Rate limiting to avoid abuse detection

**Data Privacy:**
- Option to exclude sensitive data from logs
- Secure deletion of old snapshots
- Clear audit trail of operations

### 8. Extensibility

**Plugin System:**
- Custom validators
- Output formatters
- Import/export handlers
- Notification handlers (webhook, email, etc.)

**API Version Support:**
- Abstract API client interface
- Version-specific implementations
- Automatic API version detection

### 9. Testing Strategy

**Unit Tests:**
- API client mocking
- State manager logic
- Diff algorithm validation
- Authentication flows

**Integration Tests:**
- Real API calls (with test credentials)
- End-to-end workflows
- Error recovery scenarios

**Test Data:**
- Mock API responses
- Sample route objects
- Test snapshots

## Technology Stack Considerations

### Language Options

**Python** (Recommended)
- Excellent HTTP libraries (requests, httpx)
- Rich CLI frameworks (Click, Typer)
- Great data handling (Pydantic, dataclasses)
- Keyring support built-in
- Easy deployment

**Go**
- Single binary distribution
- Excellent performance
- Strong typing
- Good standard library

**TypeScript/Node.js**
- npm ecosystem
- Good async support
- JSON-native

### Key Libraries (Python Example)

```
# Core
requests or httpx          # HTTP client
pydantic                   # Data validation
click or typer            # CLI framework
python-dotenv             # Environment config
pyyaml                    # YAML config

# Security
keyring                   # Secure credential storage
cryptography              # Encryption

# Storage
tinydb or sqlite3         # Local state storage

# Utilities
rich                      # Beautiful terminal output
loguru                    # Logging
deepdiff                  # Intelligent diffing
tabulate                  # Table formatting
```

## Implementation Phases

### Phase 1: Foundation (MVP)
- [ ] Project structure and configuration
- [ ] Authentication manager
- [ ] Basic API client
- [ ] Simple route object listing
- [ ] Basic snapshot storage

### Phase 2: Core Functionality
- [ ] Full route object CRUD
- [ ] Contact management
- [ ] Diff generation
- [ ] Complete CLI commands

### Phase 3: Advanced Features
- [ ] Change tracking with history
- [ ] Bulk operations
- [ ] Search and filtering
- [ ] Rich output formatting

### Phase 4: Polish
- [ ] Interactive mode
- [ ] Plugin system
- [ ] Comprehensive testing
- [ ] Documentation and examples

## Open Questions

1. **Credential Management:** Prefer keyring or encrypted file as default?
2. **Language Choice:** Python, Go, or TypeScript?
3. **Distribution:** PyPI package, binary, or Docker container?
4. **Snapshot Retention:** How many historical snapshots to keep by default?
5. **Notification System:** Should change detection trigger notifications?
6. **Multi-account Support:** Support managing multiple RADb accounts?

## Next Steps

1. Choose implementation language
2. Set up project structure
3. Implement authentication flow
4. Build basic API client
5. Create initial CLI commands
6. Add state management
7. Implement change tracking
