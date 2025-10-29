# Installation Guide

Complete installation instructions for the RADb API Client on all supported platforms.

## Table of Contents

- [Quick Installation](#quick-installation)
- [Platform-Specific Instructions](#platform-specific-instructions)
  - [Linux](#linux)
  - [macOS](#macos)
  - [Windows](#windows)
- [Installation from Source](#installation-from-source)
- [Container Installation](#container-installation)
- [Post-Installation](#post-installation)
- [Upgrading](#upgrading)
- [Uninstallation](#uninstallation)
- [Troubleshooting](#troubleshooting)

## Quick Installation

### Binary Download (Recommended)

Download the latest release for your platform:

```bash
# Linux (amd64)
curl -L https://github.com/example/radb-client/releases/latest/download/radb-client-linux-amd64 -o radb-client
chmod +x radb-client
sudo mv radb-client /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/example/radb-client/releases/latest/download/radb-client-darwin-amd64 -o radb-client
chmod +x radb-client
sudo mv radb-client /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/example/radb-client/releases/latest/download/radb-client-darwin-arm64 -o radb-client
chmod +x radb-client
sudo mv radb-client /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/example/radb-client/releases/latest/download/radb-client-windows-amd64.exe" -OutFile "radb-client.exe"
# Move to a directory in your PATH
```

### Verify Installation

```bash
radb-client --version
# Output: radb-client version 1.0.0
```

## Platform-Specific Instructions

### Linux

#### Option 1: Binary Installation (Recommended)

**For most users (amd64):**

```bash
# Download
curl -L https://github.com/example/radb-client/releases/latest/download/radb-client-linux-amd64 -o radb-client

# Make executable
chmod +x radb-client

# Install system-wide
sudo mv radb-client /usr/local/bin/

# Verify
radb-client --version
```

**For ARM64 systems:**

```bash
curl -L https://github.com/example/radb-client/releases/latest/download/radb-client-linux-arm64 -o radb-client
chmod +x radb-client
sudo mv radb-client /usr/local/bin/
```

#### Option 2: Package Manager (Future)

**Note:** Package manager support is planned for future releases.

```bash
# Ubuntu/Debian (Future)
# sudo apt install radb-client

# CentOS/RHEL (Future)
# sudo yum install radb-client

# Arch Linux (Future)
# yay -S radb-client
```

#### Option 3: From Source

See [Installation from Source](#installation-from-source) section.

#### Post-Installation on Linux

**Install keyring support (optional but recommended):**

```bash
# GNOME Desktop
sudo apt install gnome-keyring

# KDE Plasma
sudo apt install kwalletmanager

# For headless servers, encrypted file fallback is used automatically
```

**Verify keyring is working:**

```bash
# Should show keyring info or indicate fallback
radb-client auth status
```

---

### macOS

#### Option 1: Binary Installation (Recommended)

**For Intel Macs:**

```bash
# Download
curl -L https://github.com/example/radb-client/releases/latest/download/radb-client-darwin-amd64 -o radb-client

# Make executable
chmod +x radb-client

# Remove quarantine attribute (required on macOS)
xattr -d com.apple.quarantine radb-client

# Install
sudo mv radb-client /usr/local/bin/

# Verify
radb-client --version
```

**For Apple Silicon (M1/M2/M3) Macs:**

```bash
curl -L https://github.com/example/radb-client/releases/latest/download/radb-client-darwin-arm64 -o radb-client
chmod +x radb-client
xattr -d com.apple.quarantine radb-client
sudo mv radb-client /usr/local/bin/
```

#### Option 2: Homebrew (Future)

**Note:** Homebrew formula is planned for future releases.

```bash
# Future
# brew install radb-client
```

#### Option 3: From Source

See [Installation from Source](#installation-from-source) section.

#### macOS Security Note

On first run, macOS may warn about unverified developer:

1. **If binary won't run:**
   ```bash
   xattr -d com.apple.quarantine /usr/local/bin/radb-client
   ```

2. **Or manually approve:**
   - System Preferences → Security & Privacy
   - Click "Allow Anyway" for radb-client

#### Keychain Access

The client uses macOS Keychain automatically. No additional setup required.

---

### Windows

#### Option 1: Binary Installation

**Using PowerShell (Recommended):**

```powershell
# Download
Invoke-WebRequest -Uri "https://github.com/example/radb-client/releases/latest/download/radb-client-windows-amd64.exe" -OutFile "radb-client.exe"

# Add to PATH (User level)
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
$BinPath = "$HOME\bin"
New-Item -ItemType Directory -Force -Path $BinPath
Move-Item radb-client.exe $BinPath\
[Environment]::SetEnvironmentVariable("Path", "$UserPath;$BinPath", "User")

# Verify (restart PowerShell first)
radb-client --version
```

**Manual Installation:**

1. Download `radb-client-windows-amd64.exe` from [releases page](https://github.com/example/radb-client/releases)
2. Rename to `radb-client.exe`
3. Move to a directory in your PATH (e.g., `C:\Windows\System32` or `C:\Program Files\radb-client\`)
4. Open new Command Prompt or PowerShell
5. Run: `radb-client --version`

#### Option 2: Package Manager (Future)

**Chocolatey (Future):**
```powershell
# Future
# choco install radb-client
```

**Scoop (Future):**
```powershell
# Future
# scoop install radb-client
```

#### Windows Credential Manager

The client uses Windows Credential Manager automatically. No additional setup required.

#### Windows Defender SmartScreen

If Windows Defender blocks execution:

1. Right-click on radb-client.exe
2. Select "Properties"
3. Check "Unblock" at the bottom
4. Click "Apply"

---

## Installation from Source

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, but recommended)

### Install Go

**Linux:**
```bash
wget https://go.dev/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

**macOS:**
```bash
brew install go
```

**Windows:**
Download and run installer from https://go.dev/dl/

### Build from Source

```bash
# Clone repository
git clone https://github.com/example/radb-client.git
cd radb-client

# Build
make build
# Or without make:
go build -o radb-client ./cmd/radb-client

# Install
sudo mv radb-client /usr/local/bin/
# Or on Windows:
# Move-Item radb-client.exe C:\Windows\System32\

# Verify
radb-client --version
```

### Development Installation

If you're developing the client:

```bash
# Install directly from source
go install ./cmd/radb-client

# This installs to $GOPATH/bin (usually ~/go/bin)
# Make sure it's in your PATH
export PATH=$PATH:$HOME/go/bin
```

---

## Container Installation

### Docker (Future)

**Note:** Docker image is planned for future releases.

```bash
# Future
# docker pull ghcr.io/example/radb-client:latest
# docker run -it ghcr.io/example/radb-client:latest --version
```

### Podman (Future)

```bash
# Future
# podman pull ghcr.io/example/radb-client:latest
# podman run -it ghcr.io/example/radb-client:latest --version
```

---

## Post-Installation

### Initial Setup

1. **Verify installation:**
   ```bash
   radb-client --version
   ```

2. **Initialize configuration:**
   ```bash
   radb-client config init
   ```

3. **Authenticate:**
   ```bash
   radb-client auth login
   ```
   Enter your RADb username (email) and API key.

4. **Test connection:**
   ```bash
   radb-client auth test
   ```

5. **List routes:**
   ```bash
   radb-client route list
   ```

### Shell Completion

Enable tab completion for your shell:

**Bash:**
```bash
radb-client completion bash > /etc/bash_completion.d/radb-client
# Or for user only:
radb-client completion bash > ~/.bash_completion
source ~/.bash_completion
```

**Zsh:**
```bash
radb-client completion zsh > ~/.zsh/completion/_radb-client
# Add to .zshrc:
fpath=(~/.zsh/completion $fpath)
autoload -Uz compinit && compinit
```

**Fish:**
```bash
radb-client completion fish > ~/.config/fish/completions/radb-client.fish
```

**PowerShell:**
```powershell
radb-client completion powershell | Out-String | Invoke-Expression
# Add to profile for persistence:
radb-client completion powershell >> $PROFILE
```

### System Integration

#### Systemd Service (Linux)

For running as a service:

```ini
# /etc/systemd/system/radb-client.service
[Unit]
Description=RADb Client Service
After=network.target

[Service]
Type=simple
User=radb
Environment="RADB_USERNAME=user@example.com"
Environment="RADB_API_KEY=your-api-key"
ExecStart=/usr/local/bin/radb-client route list
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable radb-client
sudo systemctl start radb-client
```

#### Cron Job

For periodic checks:

```bash
# Edit crontab
crontab -e

# Add line (runs daily at 9 AM):
0 9 * * * /usr/local/bin/radb-client route list >> /var/log/radb-client.log 2>&1
```

#### Launchd (macOS)

```xml
<!-- ~/Library/LaunchAgents/com.example.radb-client.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.example.radb-client</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/radb-client</string>
        <string>route</string>
        <string>list</string>
    </array>
    <key>StartInterval</key>
    <integer>86400</integer>
</dict>
</plist>
```

Load:
```bash
launchctl load ~/Library/LaunchAgents/com.example.radb-client.plist
```

#### Windows Task Scheduler

Create a scheduled task:

```powershell
# PowerShell as Administrator
$action = New-ScheduledTaskAction -Execute 'radb-client' -Argument 'route list'
$trigger = New-ScheduledTaskTrigger -Daily -At 9am
Register-ScheduledTask -TaskName "RADb Client" -Action $action -Trigger $trigger
```

---

## Upgrading

### Binary Upgrade

**Linux/macOS:**

```bash
# Download new version
curl -L https://github.com/example/radb-client/releases/latest/download/radb-client-linux-amd64 -o radb-client
chmod +x radb-client

# Backup old version (optional)
sudo mv /usr/local/bin/radb-client /usr/local/bin/radb-client.backup

# Install new version
sudo mv radb-client /usr/local/bin/

# Verify
radb-client --version
```

**Windows:**

```powershell
# Backup old version
Copy-Item C:\Windows\System32\radb-client.exe C:\Windows\System32\radb-client.exe.backup

# Download and install new version
Invoke-WebRequest -Uri "https://github.com/example/radb-client/releases/latest/download/radb-client-windows-amd64.exe" -OutFile "radb-client.exe"
Move-Item radb-client.exe C:\Windows\System32\ -Force
```

### Configuration Migration

Configuration is usually backward compatible. Check [CHANGELOG.md](CHANGELOG.md) for breaking changes.

```bash
# Backup configuration
cp ~/.radb-client/config.yaml ~/.radb-client/config.yaml.backup

# Validate after upgrade
radb-client config validate

# If issues, reset to defaults:
# radb-client config reset
```

---

## Uninstallation

### Remove Binary

**Linux/macOS:**
```bash
sudo rm /usr/local/bin/radb-client
```

**Windows:**
```powershell
Remove-Item C:\Windows\System32\radb-client.exe
```

### Remove Configuration and Data

**Warning:** This removes all local data including snapshots and history.

**Linux/macOS:**
```bash
rm -rf ~/.radb-client/
```

**Windows:**
```powershell
Remove-Item -Recurse -Force $HOME\.radb-client\
```

### Remove Credentials

**macOS:**
```bash
security delete-generic-password -s radb-client
```

**Linux:**
```bash
# GNOME Keyring
secret-tool clear service radb-client

# Or just remove data directory (includes encrypted file)
rm -rf ~/.radb-client/
```

**Windows:**
```powershell
cmdkey /delete:radb-client
```

---

## Troubleshooting

### Installation Issues

**"Command not found" after installation**

Check PATH:
```bash
echo $PATH
# Should include /usr/local/bin or installation directory
```

Add to PATH if needed:
```bash
export PATH=$PATH:/usr/local/bin
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
```

**"Permission denied" when installing**

Use sudo:
```bash
sudo mv radb-client /usr/local/bin/
```

Or install to user directory:
```bash
mkdir -p ~/bin
mv radb-client ~/bin/
export PATH=$PATH:~/bin
```

**macOS: "Cannot be opened because the developer cannot be verified"**

```bash
xattr -d com.apple.quarantine /usr/local/bin/radb-client
```

**Windows: "Windows protected your PC"**

1. Click "More info"
2. Click "Run anyway"

Or remove mark of the web:
```powershell
Unblock-File radb-client.exe
```

### Runtime Issues

**"Failed to load configuration"**

Initialize configuration:
```bash
radb-client config init
```

**"Failed to store credentials"**

Check keyring availability or use environment variables:
```bash
export RADB_USERNAME="user@example.com"
export RADB_API_KEY="your-api-key"
```

**For more troubleshooting, see [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)**

---

## Platform Requirements

### Minimum Requirements

**Linux:**
- Kernel: 3.10 or later
- libc: glibc 2.17 or later / musl libc

**macOS:**
- macOS 10.13 (High Sierra) or later

**Windows:**
- Windows 7 or later
- Windows Server 2012 or later

### Recommended System Specifications

- RAM: 256 MB available
- Disk: 50 MB for binary + 100 MB for data
- Network: Internet connection for API access

---

## Next Steps

After installation:

1. Read the [User Guide](docs/USER_GUIDE.md)
2. Review [Configuration Guide](docs/CONFIGURATION.md)
3. Explore [Examples](docs/EXAMPLES.md)

## Support

- Documentation: [docs/](docs/)
- Issues: [GitHub Issues](https://github.com/example/radb-client/issues)
- Discussions: [GitHub Discussions](https://github.com/example/radb-client/discussions)
