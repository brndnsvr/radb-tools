# Changelog

All notable changes to the RADb API Client will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned Features
- Interactive TUI mode
- Watch mode for continuous monitoring
- Webhook notifications
- Multi-account support

## [1.0.0] - 2025-10-29

### Initial Release

#### Added
- Complete RADb API client implementation
- Route object management (IPv4 and IPv6)
  - Create, read, update, delete operations
  - List and filter routes
  - Route validation
  - Bulk import/export
- Account contact management
  - CRUD operations for contacts
  - Contact search and filtering
- Change tracking system
  - Automatic snapshot creation
  - Diff generation between snapshots
  - Change history with timestamps
  - Audit trail in JSONL format
- Secure credential management
  - System keyring integration (macOS Keychain, Windows Credential Manager, Linux Secret Service)
  - Encrypted file fallback for headless systems
  - Environment variable support for CI/CD
  - Never logs or exposes credentials
- Configuration management
  - YAML configuration file support
  - Environment variable overrides
  - Command-line flag overrides
  - Configuration validation
- CLI interface with Cobra
  - Intuitive command structure
  - Global flags (verbose, format, timeout, etc.)
  - Built-in help and documentation
  - Shell completion support (bash, zsh, fish)
- State management
  - Local snapshot storage
  - Historical snapshots with retention policy
  - Automatic cleanup of old snapshots
  - Snapshot metadata and checksums
- Search functionality
  - Search IRR database
  - Filter by object type
  - ASN validation
- Output formatting
  - Table format (human-readable)
  - JSON format (machine-readable)
  - YAML format
  - Colored output support
- Error handling
  - Detailed error messages
  - Retry logic with exponential backoff
  - Rate limit handling
  - Network error recovery
- Cross-platform support
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64)
- Comprehensive documentation
  - User guide
  - Command reference
  - API integration guide
  - Configuration guide
  - Examples and tutorials
  - Troubleshooting guide
  - Development guide
  - Security guide
  - Architecture documentation

#### Security
- HTTPS-only communication with certificate validation
- Secure credential storage using OS keyring or encrypted files
- No logging of sensitive information
- File permissions automatically set to secure values (600/700)
- Audit trail for all operations

#### Performance
- HTTP connection pooling
- Local caching for faster operations
- Efficient JSON serialization
- Minimal memory footprint
- Fast startup time

### Developer Experience
- Clean Go codebase following best practices
- Comprehensive test suite
- Integration test support
- Code coverage reporting
- Linting with golangci-lint
- Makefile for common tasks
- GitHub Actions for CI/CD

---

## Release Guidelines

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backward compatible manner
- **PATCH** version for backward compatible bug fixes

### Release Process

1. Update version in `internal/version/version.go`
2. Update this CHANGELOG.md
3. Commit changes: `git commit -am "Release vX.Y.Z"`
4. Tag release: `git tag vX.Y.Z`
5. Push: `git push origin main --tags`
6. Create GitHub release with binaries

### Changelog Categories

**Added** - New features
**Changed** - Changes in existing functionality
**Deprecated** - Soon-to-be removed features
**Removed** - Removed features
**Fixed** - Bug fixes
**Security** - Security improvements

---

## Future Roadmap

### v1.1.0 (Q1 2026)
- Enhanced filtering and search capabilities
- Batch operations optimization
- Performance improvements
- Additional output formats

### v1.2.0 (Q2 2026)
- Interactive mode (TUI)
- Watch mode for real-time monitoring
- Enhanced diff visualization
- Snapshot comparison improvements

### v2.0.0 (Q3 2026)
- Multi-account support
- Plugin system
- Webhook notifications
- Team collaboration features
- Breaking changes may be introduced

---

## Maintenance

This changelog is maintained by the project maintainers. For a complete list of changes, see the [commit history](https://github.com/example/radb-client/commits/main).

## Links

- [Repository](https://github.com/example/radb-client)
- [Issue Tracker](https://github.com/example/radb-client/issues)
- [Releases](https://github.com/example/radb-client/releases)
- [Documentation](./docs/)

---

## Comparison Links

[Unreleased]: https://github.com/example/radb-client/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/example/radb-client/releases/tag/v1.0.0
