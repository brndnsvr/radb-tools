# RADb API Client

A command-line client for managing RADb (Routing Assets Database) resources programmatically.

## Features

### Core Capabilities
- 🔐 **Secure Authentication** - System keyring integration with encrypted file fallback
- 📋 **Route Management** - Full CRUD for IPv4/IPv6 route objects with RPSL support
- 👥 **Contact Management** - Manage account contacts programmatically
- 📊 **Change Tracking** - Automatic diff generation between runs
- 🔍 **Search & Discovery** - Multi-criteria search and ASN validation
- 📝 **Historical Analysis** - Time-based snapshots with JSONL audit trail
- 🚀 **Single Binary** - No dependencies, 14MB binary for all platforms

### Advanced Features
- ⚡ **Performance** - O(n) diff algorithm, streaming for large datasets, rate limiting
- 🔄 **Bulk Operations** - Worker pool for batch create/update/delete
- 📈 **Progress Tracking** - Progress bars and real-time feedback
- 🎨 **Multiple Formats** - Table, JSON, YAML output with color support
- 🔒 **Security** - Argon2id + NaCl encryption, SHA-256 integrity checks
- 🧹 **Smart Cleanup** - Configurable snapshot retention policies
- 🔧 **Interactive Setup** - Configuration wizard for easy onboarding

## Installation

### From Source

```bash
go build -o radb-client ./cmd/radb-client
```

### Binary Release

Download the latest release for your platform from the releases page.

## Quick Start

```bash
# Initialize configuration
radb-client config init

# Authenticate
radb-client auth login

# List route objects
radb-client route list

# Create a snapshot
radb-client snapshot create

# Check for changes since last run
radb-client route diff
```

## Project Status

🧪 **v0.9 Pre-Release** - Complete implementation pending final manual testing.

See [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) for complete project details and [DESIGN.md](DESIGN.md) for architecture.

## Documentation

### For Users
- [User Guide](docs/USER_GUIDE.md) - Getting started and workflows
- [Command Reference](docs/COMMANDS.md) - Complete command documentation
- [Configuration Guide](docs/CONFIGURATION.md) - Configuration options
- [Examples](docs/EXAMPLES.md) - Real-world usage examples
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues and solutions
- [Installation Guide](INSTALL.md) - Platform-specific installation

### For Developers
- [Project Summary](PROJECT_SUMMARY.md) - Complete project overview
- [Design Document](DESIGN.md) - Architecture and design decisions
- [Development Guide](docs/DEVELOPMENT.md) - Contributing and development
- [Architecture](docs/ARCHITECTURE.md) - Technical architecture details
- [Security](docs/SECURITY.md) - Security implementation and best practices
- [Go Implementation](GO_IMPLEMENTATION.md) - Go-specific patterns

### API Documentation
- [RADb API Integration](docs/API_INTEGRATION.md) - RADb API details
- [RADb API Reference](https://api.radb.net/docs.html) - Official API docs

## Development

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Build
go build ./cmd/radb-client

# Install locally
go install ./cmd/radb-client
```

## License

MIT License - See LICENSE file for details

## Contributing

Contributions welcome! Please read the design document first to understand the architecture.
