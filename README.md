# RADb API Client

A command-line client for managing RADb (Routing Assets Database) resources programmatically.

## Features

- 🔐 Secure credential management with system keyring support
- 📋 Manage route objects (IPv4 and IPv6) without web UI
- 👥 Account contact management
- 📊 Change tracking and diff generation between runs
- 🔍 Search and validation capabilities
- 📝 Historical snapshots with audit trail
- 🚀 Single binary distribution - no dependencies required

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

🚧 **In Development** - This project is in the design and early implementation phase.

See [DESIGN.md](DESIGN.md) for the complete architecture and design decisions.

## Documentation

- [Design Document](DESIGN.md) - Complete architecture and design
- [Go Implementation Guide](GO_IMPLEMENTATION.md) - Go-specific details
- [API Reference](https://api.radb.net/docs.html) - RADb API documentation

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
