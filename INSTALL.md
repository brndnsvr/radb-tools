# RADb Client - Installation Guide

Complete installation guide for radb-client, covering all installation methods and configurations.

## Table of Contents

- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Interactive Installation (Recommended)](#interactive-installation-recommended)
- [Manual Installation](#manual-installation)
- [Configuration](#configuration)
- [Authentication](#authentication)
- [Daemon Installation](#daemon-installation)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)
- [Uninstallation](#uninstallation)

---

## Quick Start

For the fastest installation experience:

```bash
# Clone the repository
git clone https://github.com/brndnsvr/radb-tools.git
cd radb-tools

# Run interactive installer
./scripts/install-interactive.sh
```

The interactive installer will guide you through the complete setup process.

---

## Prerequisites

### Required

- **Go 1.23 or higher**
  - Check: `go version`
  - Install: https://golang.org/doc/install

### Optional

- **Git** - For version information in builds
- **Sudo access** - Only for system-wide installation or daemon mode

### Platform Support

- ✅ Linux (amd64, arm64)
- ✅ macOS (Intel, Apple Silicon)
- ✅ Windows (amd64) - Limited testing

---

## Interactive Installation (Recommended)

The interactive installer provides a guided experience:

```bash
./scripts/install-interactive.sh
```

### Installation Options

1. **User Installation** (`$HOME/bin`)
   - No sudo required
   - User-specific installation
   - May need to add to PATH

2. **System Installation** (`/usr/local/bin`)
   - Requires sudo
   - Available to all users
   - Usually in PATH by default

3. **Custom Location**
   - Specify your own directory
   - Flexibility for specific setups

4. **Build Only**
   - Just compile the binary
   - No installation step
   - Binary in `bin/radb-client`

### Example: Complete Setup

```bash
# Run interactive installer
./scripts/install-interactive.sh

# Example responses for full setup:
# 1. Installation type: 1 (User installation)
# 2. Initialize config: y
# 3. Setup credentials: y
#    - Enter username: your-radb-username
#    - Enter password: your-radb-password
# 4. Install daemon: y (Linux only)
#    - Configure daemon credentials: y
```

### What the Installer Does

✅ Checks prerequisites (Go version, etc.)
✅ Downloads Go module dependencies
✅ Builds binary with version information
✅ Installs to your chosen location
✅ Checks if location is in PATH
✅ Offers to initialize configuration
✅ Offers to set up credentials
✅ (Linux) Offers to install systemd daemon
✅ Displays next steps

---

## Manual Installation

If you prefer manual control:

### Step 1: Download Dependencies

```bash
go mod download
```

### Step 2: Build Binary

```bash
# Simple build
go build -o bin/radb-client ./cmd/radb-client

# Build with version information (recommended)
VERSION=$(cat VERSION)
GIT_COMMIT=$(git rev-parse --short HEAD)
GIT_BRANCH=$(git branch --show-current)
BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S_UTC')

LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.Version=$VERSION'"
LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.GitCommit=$GIT_COMMIT'"
LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.GitBranch=$GIT_BRANCH'"
LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.BuildDate=$BUILD_DATE'"

go build -ldflags "$LDFLAGS" -o bin/radb-client ./cmd/radb-client
```

### Step 3: Install Binary

**User Installation:**
```bash
mkdir -p $HOME/bin
cp bin/radb-client $HOME/bin/
chmod +x $HOME/bin/radb-client

# Add to PATH if needed
echo 'export PATH="$PATH:$HOME/bin"' >> ~/.bashrc
source ~/.bashrc
```

**System Installation:**
```bash
sudo cp bin/radb-client /usr/local/bin/
sudo chmod +x /usr/local/bin/radb-client
```

### Step 4: Initialize Configuration

```bash
radb-client config init
```

### Step 5: Authenticate

```bash
radb-client auth login
```

---

## Configuration

### Initialize Configuration

Create default configuration:

```bash
radb-client config init
```

This creates:
- `~/.radb-client/config.yaml` - Configuration file
- `~/.radb-client/cache/` - Cache directory
- `~/.radb-client/history/` - History/snapshot directory

### View Configuration

```bash
# Show full configuration
radb-client config show

# Show specific format
radb-client config show --format yaml
```

### Configuration File Location

Default: `~/.radb-client/config.yaml`

Override with:
```bash
radb-client --config /path/to/config.yaml <command>
```

### Configuration Options

Edit `~/.radb-client/config.yaml`:

```yaml
api:
  baseurl: https://api.radb.net
  source: RADB
  format: json
  timeout: 30
  ratelimit:
    requestsperminute: 60
    burstsize: 10
  retry:
    maxattempts: 3
    backoffmultiplier: 2
    initialdelayms: 1000

credentials:
  username: "your-username"  # Set via 'auth login'

performance:
  streamthreshold: 1000
  compresshistory: true
  maxconcurrentrequests: 5

preferences:
  cachedir: /home/user/.radb-client/cache
  historydir: /home/user/.radb-client/history
  loglevel: INFO  # DEBUG, INFO, WARN, ERROR

state:
  enablelocking: true
  atomicwrites: true
  formatversion: "1.0"
```

---

## Authentication

### Login

Store credentials securely:

```bash
radb-client auth login
```

You'll be prompted for:
- Username: Your RADb account username
- Password: Your RADb account password (hidden input)

**Security:**
- Password encrypted with Argon2id + NaCl secretbox
- Stored in system keyring if available
- Falls back to encrypted file storage

### Logout

Remove stored credentials:

```bash
radb-client auth logout
```

### Check Status

Verify authentication status:

```bash
radb-client auth status
```

---

## Daemon Installation

**(Linux only - Ubuntu 22.04 LTS and newer)**

Install radb-client as a systemd service for continuous monitoring:

### Interactive Daemon Setup

During interactive installation, choose "Yes" when prompted for daemon installation.

### Manual Daemon Installation

```bash
# Run daemon installer
sudo ./scripts/install-daemon.sh

# With credentials
sudo RADB_USERNAME="your-username" RADB_PASSWORD="your-password" \
  ./scripts/install-daemon.sh
```

### Daemon Management

```bash
# Start daemon
sudo systemctl start radb-daemon

# Stop daemon
sudo systemctl stop radb-daemon

# Restart daemon
sudo systemctl restart radb-daemon

# Enable on boot
sudo systemctl enable radb-daemon

# View status
sudo systemctl status radb-daemon

# View logs
sudo journalctl -u radb-daemon -f
```

### Daemon Management Helper

A helper script is installed at `/usr/local/bin/radb-daemon`:

```bash
# Start daemon
radb-daemon start

# Stop daemon
radb-daemon stop

# Restart daemon
radb-daemon restart

# View status
radb-daemon status

# View logs
radb-daemon logs

# Follow logs
radb-daemon tail
```

**Note:** Daemon mode is currently a placeholder. Full implementation requires completion of API client and state manager.

See `docs/DAEMON_DEPLOYMENT.md` for detailed daemon documentation.

---

## Verification

### Test Installation

```bash
# Check version
radb-client version

# Full version info
radb-client version --format json

# Check help
radb-client --help

# Test configuration
radb-client config show

# Test authentication status
radb-client auth status
```

### Test Basic Commands

```bash
# List routes (requires authentication)
radb-client route list

# List contacts
radb-client contact list

# Create snapshot
radb-client snapshot create

# View history
radb-client history list
```

### Run Test Suite

```bash
# Run unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/config/...
```

### Manual Testing

Follow comprehensive testing procedures:

```bash
cat docs/testing/TESTING_RUNBOOK.md
```

---

## Troubleshooting

### Build Issues

**Issue: Go version too old**
```
go: directive requires go version >= 1.23
```

Solution:
```bash
# Update Go
go install golang.org/dl/go1.24@latest
go1.24 download
```

**Issue: Missing dependencies**
```
missing go.sum entry
```

Solution:
```bash
go mod tidy
go mod download
```

### Installation Issues

**Issue: Permission denied**
```
cannot create directory: permission denied
```

Solution:
- Use user installation option (1) in interactive installer
- Or use sudo for system installation

**Issue: Binary not found in PATH**
```
radb-client: command not found
```

Solution:
```bash
# Add to PATH
echo 'export PATH="$PATH:$HOME/bin"' >> ~/.bashrc
source ~/.bashrc

# Or use full path
/home/user/bin/radb-client version
```

### Configuration Issues

**Issue: Config file not found**
```
Error: config file not found
```

Solution:
```bash
# Initialize config
radb-client config init

# Or specify config location
radb-client --config /path/to/config.yaml <command>
```

**Issue: Permission denied on config directory**
```
permission denied: ~/.radb-client
```

Solution:
```bash
# Fix permissions
chmod 700 ~/.radb-client
chmod 600 ~/.radb-client/config.yaml
```

### Authentication Issues

**Issue: Credentials not stored**
```
Error: no credentials found
```

Solution:
```bash
# Login again
radb-client auth login

# Check status
radb-client auth status
```

**Issue: Keyring access denied**
```
Error: could not access keyring
```

Solution:
- Credentials will fall back to encrypted file storage
- Check system keyring is accessible
- Verify keyring daemon is running (Linux)

### Daemon Issues

**Issue: Daemon fails to start**
```
Failed to start radb-daemon.service
```

Solution:
```bash
# Check logs
sudo journalctl -u radb-daemon -n 50

# Verify binary
which radb-client

# Check permissions
ls -l /usr/local/bin/radb-client

# Reinstall daemon
sudo ./scripts/install-daemon.sh
```

---

## Uninstallation

### Remove Binary

**User installation:**
```bash
rm -f $HOME/bin/radb-client
```

**System installation:**
```bash
sudo rm -f /usr/local/bin/radb-client
```

### Remove Configuration

```bash
rm -rf ~/.radb-client
```

### Remove Daemon (Linux)

```bash
# Stop and disable service
sudo systemctl stop radb-daemon
sudo systemctl disable radb-daemon

# Remove files
sudo rm -f /etc/systemd/system/radb-daemon.service
sudo rm -f /usr/local/bin/radb-daemon
sudo rm -rf /etc/radb-client
sudo rm -rf /var/log/radb-client

# Remove user
sudo userdel -r radb

# Reload systemd
sudo systemctl daemon-reload
```

### Complete Removal

```bash
# Remove everything
rm -f $HOME/bin/radb-client
sudo rm -f /usr/local/bin/radb-client
rm -rf ~/.radb-client

# If daemon was installed
sudo systemctl stop radb-daemon 2>/dev/null
sudo systemctl disable radb-daemon 2>/dev/null
sudo rm -f /etc/systemd/system/radb-daemon.service
sudo rm -f /usr/local/bin/radb-daemon
sudo rm -rf /etc/radb-client
sudo rm -rf /var/log/radb-client
sudo userdel -r radb 2>/dev/null
sudo systemctl daemon-reload
```

---

## Next Steps

After successful installation:

1. **Read the Quick Start Guide**
   ```bash
   cat QUICKSTART.md
   ```

2. **Learn the CLI**
   ```bash
   radb-client --help
   radb-client route --help
   ```

3. **Run Manual Tests**
   ```bash
   cat docs/testing/TESTING_RUNBOOK.md
   ```

4. **Configure for Your Environment**
   - Edit `~/.radb-client/config.yaml`
   - Adjust rate limits, cache settings, etc.

5. **Set Up Automation**
   - Install daemon for continuous monitoring
   - Or set up cron jobs for periodic checks

---

## Additional Documentation

- **Quick Start**: `docs/installation/QUICKSTART.md` - Get started quickly
- **Testing Guide**: `docs/testing/TESTING_RUNBOOK.md` - Comprehensive testing procedures
- **Daemon Deployment**: `docs/installation/DAEMON_DEPLOYMENT.md` - Daemon setup and management
- **Version Management**: `docs/design/VERSION_MANAGEMENT.md` - Version and release process
- **Design Documents**: `docs/design/DESIGN.md`, `docs/design/GO_IMPLEMENTATION.md` - Architecture details

---

## Support

- **Issues**: https://github.com/brndnsvr/radb-tools/issues
- **Discussions**: https://github.com/brndnsvr/radb-tools/discussions
- **Documentation**: https://github.com/brndnsvr/radb-tools/tree/main/docs

---

**Installation complete! 🚀**

Run `radb-client --help` to get started.
