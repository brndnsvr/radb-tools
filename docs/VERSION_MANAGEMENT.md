# Version Management Guide

This document describes how version management works in the RADb client project.

## Overview

The project uses a centralized version management system with multiple components working together:

1. **VERSION file** - Single source of truth for the version number
2. **internal/version package** - Go package providing version information
3. **Build-time injection** - Git and build metadata injected via ldflags
4. **Version command** - CLI command to display version information
5. **Tagging script** - Automated version tagging and release preparation

## Version File

Location: `/VERSION` (root of repository)

Contains a single line with the semantic version:
```
0.9.0-pre
```

Format: `MAJOR.MINOR.PATCH[-SUFFIX]`
- MAJOR: Breaking changes
- MINOR: New features (backwards compatible)
- PATCH: Bug fixes
- SUFFIX: Optional (pre, alpha, beta, rc1, etc.)

## Version Package

Location: `internal/version/version.go`

Provides centralized version information throughout the application:

```go
import "github.com/bss/radb-client/internal/version"

// Get version string
v := version.Short()           // "0.9.0-pre"
v := version.String()          // "radb-client version 0.9.0-pre (commit: abc1234, built: 2025-10-29)"
v := version.Full()            // Multi-line detailed version

// Get structured data
info := version.Get()          // Returns Info struct
isPreRelease := version.IsPreRelease()  // true/false
```

### Version Information Available

The version package provides:
- **Version**: Semantic version (from VERSION file or ldflags)
- **GitCommit**: Short git commit hash (injected at build time)
- **GitBranch**: Git branch name (injected at build time)
- **BuildDate**: ISO 8601 timestamp (injected at build time)
- **GoVersion**: Go version used to compile
- **Platform**: OS/Architecture (linux/amd64, etc.)

## Version Commands

### Version Subcommand

Display version information:

```bash
# Full version information (default)
radb-client version

# Short version only
radb-client version --short
radb-client version -s

# JSON output
radb-client version --output json
radb-client version -o json

# YAML output
radb-client version -o yaml
```

Example output:

**Text format** (default):
```
radb-client version 0.9.0-pre

Build Information:
  Git Commit:   cad1ab8
  Git Branch:   main
  Build Date:   2025-10-29T21:36:13Z
  Go Version:   go1.23.2
  Platform:     linux/amd64

🧪 Pre-release build - pending final manual testing

See TESTING_RUNBOOK.md for complete testing procedures
```

**Short format** (`-s`):
```
0.9.0-pre
```

**JSON format** (`-o json`):
```json
{
  "version": "0.9.0-pre",
  "git_commit": "cad1ab8",
  "git_branch": "main",
  "build_date": "2025-10-29T21:36:13Z",
  "go_version": "go1.23.2",
  "platform": "linux/amd64"
}
```

### Global --version Flag

Quick version check:

```bash
radb-client --version
# Output: 0.9.0-pre
```

This is automatically provided by Cobra and shows the short version.

## Building with Version Information

### Development Build

Simple build (uses hardcoded defaults):
```bash
go build -o radb-client ./cmd/radb-client
```

This will use:
- Version from `internal/version/version.go` default
- GitCommit: "dev"
- BuildDate: "unknown"

### Production Build

Build with full version injection:

```bash
# Using build script (recommended)
./scripts/build.sh

# Manual build with version injection
VERSION=$(cat VERSION)
GIT_COMMIT=$(git rev-parse --short HEAD)
GIT_BRANCH=$(git branch --show-current)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

go build -o radb-client \
  -ldflags "-s -w \
    -X 'github.com/bss/radb-client/internal/version.Version=$VERSION' \
    -X 'github.com/bss/radb-client/internal/version.GitCommit=$GIT_COMMIT' \
    -X 'github.com/bss/radb-client/internal/version.GitBranch=$GIT_BRANCH' \
    -X 'github.com/bss/radb-client/internal/version.BuildDate=$BUILD_DATE'" \
  ./cmd/radb-client
```

### Build Script

The `scripts/build.sh` script handles everything automatically:

```bash
# Build all platforms with version from VERSION file
./scripts/build.sh

# Override version
VERSION=1.0.0 ./scripts/build.sh

# Custom output directory
OUTPUT_DIR=./release ./scripts/build.sh
```

Features:
- Reads version from VERSION file
- Automatically gets git information
- Builds for multiple platforms
- Generates checksums
- Injects all version metadata

## Version Tagging

### Using the Tagging Script

The `scripts/tag-version.sh` script automates version tagging:

```bash
# Tag using version from VERSION file
./scripts/tag-version.sh

# Tag with specific version
./scripts/tag-version.sh 1.0.0

# Tag pre-release
./scripts/tag-version.sh 1.0.0-rc1
```

The script will:
1. Validate version format
2. Update VERSION file
3. Update `internal/version/version.go`
4. Commit version changes
5. Create annotated git tag
6. Show next steps

### Manual Tagging

If you prefer manual tagging:

```bash
# 1. Update VERSION file
echo "1.0.0" > VERSION

# 2. Update internal/version/version.go
sed -i 's/Version = ".*"/Version = "1.0.0"/' internal/version/version.go

# 3. Commit changes
git add VERSION internal/version/version.go
git commit -m "Bump version to 1.0.0"

# 4. Create tag
git tag -a v1.0.0 -m "Release v1.0.0"

# 5. Push
git push origin main --tags
```

## Version Workflow

### Development Cycle

```
0.9.0-pre  →  Manual Testing  →  0.9.0  →  Testing  →  1.0.0
  (dev)         (pre-release)      (rc)     (stable)   (release)
```

### Release Process

1. **Development**
   ```bash
   # Working on features
   VERSION file: 0.9.0-pre
   ```

2. **Pre-release Testing**
   ```bash
   # Ready for testing
   ./scripts/tag-version.sh 0.9.0
   git push origin main --tags
   ```

3. **Release Candidate**
   ```bash
   # After testing, before final release
   ./scripts/tag-version.sh 1.0.0-rc1
   git push origin main --tags
   ```

4. **Final Release**
   ```bash
   # Production ready
   ./scripts/tag-version.sh 1.0.0
   git push origin main --tags
   ```

5. **Post-release**
   ```bash
   # Start next version
   echo "1.1.0-pre" > VERSION
   git add VERSION
   git commit -m "Start v1.1.0 development"
   ```

## Semantic Versioning

We follow [Semantic Versioning 2.0.0](https://semver.org/):

**MAJOR.MINOR.PATCH**

- **MAJOR**: Incompatible API changes
  - Breaking changes to CLI commands
  - Incompatible config file format changes
  - Breaking changes to public APIs

- **MINOR**: Backwards-compatible functionality
  - New features
  - New commands
  - New configuration options
  - Deprecations (with migration path)

- **PATCH**: Backwards-compatible bug fixes
  - Bug fixes
  - Security patches
  - Documentation updates
  - Performance improvements

### Pre-release Suffixes

- **-pre**: Pre-release, under active development
- **-alpha**: Alpha release, unstable
- **-beta**: Beta release, feature complete but unstable
- **-rc1, -rc2**: Release candidate, stable but final testing

Examples:
- `0.9.0-pre` - Development pre-release
- `1.0.0-alpha` - First alpha
- `1.0.0-beta.1` - First beta
- `1.0.0-rc1` - First release candidate
- `1.0.0` - Final stable release

## Integration with CI/CD

### GitHub Actions

The version is automatically available in GitHub Actions:

```yaml
- name: Get version
  id: version
  run: echo "version=$(cat VERSION)" >> $GITHUB_OUTPUT

- name: Build with version
  run: ./scripts/build.sh
  env:
    VERSION: ${{ steps.version.outputs.version }}
```

### Release Tags

When a tag is pushed, GitHub Actions can automatically:
1. Build binaries with version injection
2. Run tests
3. Create GitHub release
4. Upload artifacts
5. Generate release notes

## Best Practices

### Updating Versions

1. **Always update VERSION file first**
   - This is the single source of truth
   - Scripts and build process read from this file

2. **Use the tagging script**
   - Ensures consistency
   - Handles all necessary updates
   - Prevents mistakes

3. **Follow semantic versioning**
   - Breaking changes → MAJOR
   - New features → MINOR
   - Bug fixes → PATCH

4. **Tag important milestones**
   - Development milestones
   - Release candidates
   - Stable releases

### Checking Versions

```bash
# Check version in code
cat VERSION

# Check latest tag
git describe --tags

# Check what will be built
./scripts/build.sh --dry-run  # If supported

# Check built binary
./radb-client version
```

## Troubleshooting

### Version shows "dev"

**Problem**: Built binary shows version as "dev"

**Solution**: Build with version injection:
```bash
./scripts/build.sh
# OR
go build -ldflags "-X 'github.com/bss/radb-client/internal/version.Version=$(cat VERSION)'" ./cmd/radb-client
```

### GitCommit shows "dev"

**Problem**: Not in a git repository or git not available

**Solution**: Ensure you're in a git repository:
```bash
git status  # Should show current branch
```

### Version mismatch

**Problem**: `radb-client version` shows different version than `cat VERSION`

**Solution**: Rebuild the binary:
```bash
./scripts/build.sh
```

### Tag already exists

**Problem**: `./scripts/tag-version.sh` reports tag exists

**Solution**: Either:
1. Delete and recreate (script will prompt)
2. Use a different version number

## Summary

**Quick Reference:**

- **Source of truth**: `VERSION` file
- **View version**: `radb-client version`
- **Tag version**: `./scripts/tag-version.sh`
- **Build with version**: `./scripts/build.sh`
- **Check in code**: `import "github.com/bss/radb-client/internal/version"`

---

**Version**: 0.9.0-pre
**Last Updated**: 2025-10-29
