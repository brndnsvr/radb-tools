# RADb Client - Daemon Deployment Guide

**Version**: 0.9.0-pre
**Platform**: Ubuntu 22.04 LTS and later
**Deployment Type**: Systemd service

---

## Overview

This guide explains how to deploy the RADb client as a system daemon that runs continuously on Linux servers, monitoring RADb for changes and maintaining historical snapshots.

### Daemon Features

- **Automatic Monitoring**: Periodically fetches route objects from RADb
- **Change Detection**: Automatically detects and logs changes
- **Historical Snapshots**: Maintains a complete history of changes
- **Automatic Cleanup**: Enforces retention policies
- **Systemd Integration**: Full integration with systemd for management
- **Graceful Shutdown**: Handles signals properly for clean restarts
- **Log Rotation**: Automatic log rotation via logrotate
- **Security Hardening**: Runs as unprivileged user with restricted permissions

---

## Quick Installation

### Prerequisites

- Ubuntu 22.04 LTS or later
- Root/sudo access
- Compiled `radb-client` binary
- RADb API credentials

### One-Command Installation

```bash
# Build the binary first
go build -o dist/radb-client ./cmd/radb-client

# Run installer with credentials from environment
export RADB_API_USERNAME="your-username"
export RADB_API_KEY="your-api-key"

sudo ./scripts/install-daemon.sh --binary ./dist/radb-client
```

That's it! The daemon is now running and monitoring every hour.

---

## Installation Details

### Installation Script

The `install-daemon.sh` script handles complete daemon setup:

```bash
sudo ./scripts/install-daemon.sh --help
```

**What it does:**
1. Creates system user (`radb`)
2. Creates directories with proper permissions
3. Installs binary to `/opt/radb-client/`
4. Creates configuration in `/etc/radb-client/`
5. Sets up data directory in `/var/lib/radb-client/`
6. Creates systemd service and timer
7. Configures log rotation
8. Starts and enables the service

### Installation Options

```bash
# Custom user
sudo ./scripts/install-daemon.sh \
  --binary ./dist/radb-client \
  --user myuser

# Custom check interval (every 30 minutes)
sudo ./scripts/install-daemon.sh \
  --binary ./dist/radb-client \
  --interval 1800

# Install without starting immediately
sudo ./scripts/install-daemon.sh \
  --binary ./dist/radb-client \
  --no-start

# Custom directories
sudo ./scripts/install-daemon.sh \
  --binary ./dist/radb-client \
  --config-dir /etc/radb \
  --data-dir /var/lib/radb \
  --log-dir /var/log/radb
```

### Directory Structure

After installation:

```
/opt/radb-client/              # Installation directory
├── radb-client                # Binary

/etc/radb-client/              # Configuration
├── config.yaml                # Main configuration

/var/lib/radb-client/          # Data directory
├── cache/                     # Current snapshots
│   ├── route_objects.json
│   └── contacts.json
└── history/                   # Historical snapshots
    ├── 2025-10-29T12-00-00_route_objects.json
    ├── 2025-10-29T13-00-00_route_objects.json
    └── changelog.jsonl

/var/log/radb-client/          # Logs
├── radb-client.log            # Standard output
└── radb-client-error.log      # Error output

/usr/local/bin/
├── radb-client                # Symlink for CLI access
└── radb-daemon                # Management helper script
```

---

## Configuration

### Main Configuration File

Located at `/etc/radb-client/config.yaml`:

```yaml
api:
  base_url: https://api.radb.net
  source: RADB
  format: json
  timeout: 30

  rate_limit:
    requests_per_minute: 60
    burst_size: 10

  retry:
    max_attempts: 3
    backoff_multiplier: 2
    initial_delay_ms: 1000

preferences:
  cache_dir: /var/lib/radb-client/cache
  history_dir: /var/lib/radb-client/history
  log_level: INFO
  max_snapshots: 100
  auto_snapshot: true
  output_format: json
  color: false  # Disabled for daemon mode

daemon:
  check_interval: 3600  # 1 hour in seconds
  notify_on_changes: false
  auto_cleanup: true

  retention:
    max_age_days: 90
    min_snapshots: 10
```

### Editing Configuration

```bash
# Edit configuration
sudo nano /etc/radb-client/config.yaml

# Reload configuration (sends SIGHUP to daemon)
sudo systemctl reload radb-client

# Or restart service
sudo systemctl restart radb-client
```

### Configure Credentials

If not set during installation:

```bash
# Login as radb user
sudo -u radb radb-client auth login

# Or use environment variables
export RADB_API_USERNAME="your-username"
export RADB_API_KEY="your-api-key"
sudo -u radb -E radb-client auth login <<< "$RADB_API_KEY"
```

---

## Service Management

### Systemd Service

The daemon runs as `radb-client.service`:

```bash
# Start service
sudo systemctl start radb-client

# Stop service
sudo systemctl stop radb-client

# Restart service
sudo systemctl restart radb-client

# Reload configuration (graceful)
sudo systemctl reload radb-client

# Check status
sudo systemctl status radb-client

# Enable on boot
sudo systemctl enable radb-client

# Disable on boot
sudo systemctl disable radb-client
```

### Service Status

```bash
# Detailed status
sudo systemctl status radb-client

# Check if running
sudo systemctl is-active radb-client

# Check if enabled
sudo systemctl is-enabled radb-client
```

### View Logs

```bash
# Real-time logs (journalctl)
sudo journalctl -u radb-client -f

# Last 100 lines
sudo journalctl -u radb-client -n 100

# Since yesterday
sudo journalctl -u radb-client --since yesterday

# Specific time range
sudo journalctl -u radb-client --since "2025-10-29 12:00" --until "2025-10-29 13:00"

# Log files directly
tail -f /var/log/radb-client/radb-client.log
tail -f /var/log/radb-client/radb-client-error.log
```

---

## Helper Commands

The `radb-daemon` helper provides quick access to common operations:

```bash
# Service management
radb-daemon status      # Show service status
radb-daemon logs        # Tail logs in real-time
radb-daemon start       # Start service
radb-daemon stop        # Stop service
radb-daemon restart     # Restart service
radb-daemon enable      # Enable on boot
radb-daemon disable     # Disable on boot

# Data access
radb-daemon diff        # Show recent route changes
radb-daemon snapshots   # List all snapshots
radb-daemon history     # Show change history
```

---

## Alternative: Systemd Timer

If you prefer periodic execution instead of a long-running daemon:

```bash
# Disable daemon service
sudo systemctl disable --now radb-client.service

# Enable timer-based execution
sudo systemctl enable --now radb-client-sync.timer

# Check timer status
sudo systemctl status radb-client-sync.timer

# List timers
sudo systemctl list-timers radb-client-sync.timer

# View timer logs
sudo journalctl -u radb-client-sync.service
```

The timer runs the same check operation but doesn't keep a process running continuously.

---

## Daemon Operation

### What the Daemon Does

Every check interval (default: 1 hour), the daemon:

1. **Fetches Routes**: Retrieves all route objects from RADb API
2. **Creates Snapshot**: Saves current state with timestamp and checksum
3. **Generates Diff**: Compares with previous snapshot
4. **Logs Changes**: Records any added, removed, or modified routes
5. **Cleanup**: Removes old snapshots beyond retention policy

### Daemon Lifecycle

```
Start → Initial Check → Wait (interval) → Periodic Check → Wait → ...
                         ↓
                    Signal Handler
                         ↓
                    SIGHUP → Reload Config
                    SIGTERM → Graceful Shutdown
```

### Change Detection

When changes are detected, they're logged:

```json
{
  "timestamp": "2025-10-29T14:30:00Z",
  "level": "info",
  "message": "Changes detected: 3 added, 1 removed, 2 modified"
}
```

Future versions will support:
- Email notifications
- Webhook integration
- Slack/Teams alerts
- Custom notification scripts

---

## Monitoring & Maintenance

### Health Checks

```bash
# Is service running?
systemctl is-active radb-client

# Last check time (from logs)
sudo journalctl -u radb-client -n 1 --output=cat | grep "Check completed"

# Error count
sudo journalctl -u radb-client --since today | grep -c ERROR

# Systemd health check
systemctl status radb-client --no-pager
```

### Monitoring Integration

For monitoring systems (Nagios, Zabbix, Prometheus):

```bash
# Exit code check
systemctl is-active radb-client --quiet && echo "OK" || echo "CRITICAL"

# Log-based monitoring
grep -c "ERROR" /var/log/radb-client/radb-client-error.log

# Snapshot age check
find /var/lib/radb-client/cache -name "route_objects.json" -mmin +120 && echo "STALE"
```

### Disk Space Management

```bash
# Check data directory size
du -sh /var/lib/radb-client

# Check log size
du -sh /var/log/radb-client

# Snapshot count
ls -1 /var/lib/radb-client/history/*.json | wc -l

# Manual cleanup (if needed)
sudo -u radb radb-client snapshot cleanup --older-than 30d
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check service status
sudo systemctl status radb-client

# View full logs
sudo journalctl -u radb-client -n 50

# Check configuration syntax
sudo -u radb radb-client config show

# Verify credentials
sudo -u radb radb-client auth status

# Test binary manually
sudo -u radb radb-client daemon --once
```

### Authentication Errors

```bash
# Re-configure credentials
sudo -u radb radb-client auth logout
sudo -u radb radb-client auth login

# Check credential storage
sudo -u radb ls -la /var/lib/radb-client/.credentials

# Test API connection
sudo -u radb radb-client route list
```

### High Memory/CPU Usage

```bash
# Check resource usage
systemctl status radb-client

# View process details
ps aux | grep radb-client

# Increase check interval to reduce frequency
sudo nano /etc/radb-client/config.yaml
# Set check_interval: 7200  # 2 hours

sudo systemctl restart radb-client
```

### Logs Not Appearing

```bash
# Check log directory permissions
ls -la /var/log/radb-client

# Check logrotate
sudo logrotate -f /etc/logrotate.d/radb-client

# View journalctl instead
sudo journalctl -u radb-client -f
```

### Stale Data

```bash
# Force immediate check
sudo systemctl restart radb-client

# Check last modification time
stat /var/lib/radb-client/cache/route_objects.json

# Run manual check
sudo -u radb radb-client daemon --once
```

---

## Security Considerations

### User Permissions

The daemon runs as the `radb` system user:
- No login shell
- No home directory
- Limited to data and log directories
- Cannot access other user files

### File Permissions

```bash
# Verify permissions
ls -la /etc/radb-client/config.yaml     # 640 root:radb
ls -la /var/lib/radb-client/            # 750 radb:radb
ls -la /var/log/radb-client/            # 750 radb:radb
```

### Systemd Hardening

The service includes security features:
- `NoNewPrivileges=true` - Prevents privilege escalation
- `PrivateTmp=true` - Isolated /tmp directory
- `ProtectSystem=strict` - Read-only system directories
- `ProtectHome=true` - No access to user homes
- `ReadWritePaths` - Explicit whitelist of writable paths

### Network Security

- Only HTTPS connections (no HTTP fallback)
- TLS certificate validation enforced
- Rate limiting to prevent abuse
- Timeout configurations prevent hanging

### Credential Security

- Stored in system keyring if available
- Encrypted file fallback with Argon2id + NaCl
- Never logged or exposed
- File permissions restrict access

---

## Upgrading

### Upgrade Process

```bash
# 1. Build new version
go build -o dist/radb-client ./cmd/radb-client

# 2. Stop service
sudo systemctl stop radb-client

# 3. Backup current binary
sudo cp /opt/radb-client/radb-client /opt/radb-client/radb-client.backup

# 4. Install new binary
sudo cp dist/radb-client /opt/radb-client/radb-client
sudo chmod 755 /opt/radb-client/radb-client

# 5. Test new version
sudo -u radb /opt/radb-client/radb-client version

# 6. Start service
sudo systemctl start radb-client

# 7. Verify operation
sudo systemctl status radb-client
radb-daemon logs
```

### Configuration Migration

```bash
# Backup configuration
sudo cp /etc/radb-client/config.yaml /etc/radb-client/config.yaml.backup

# After upgrade, check for new config options
radb-client config show

# Merge any new settings manually
```

### Rollback

```bash
# Stop service
sudo systemctl stop radb-client

# Restore backup
sudo cp /opt/radb-client/radb-client.backup /opt/radb-client/radb-client

# Start service
sudo systemctl start radb-client
```

---

## Uninstallation

### Complete Removal

```bash
# Stop and disable service
sudo systemctl stop radb-client
sudo systemctl disable radb-client

# Remove systemd files
sudo rm /etc/systemd/system/radb-client.service
sudo rm /etc/systemd/system/radb-client-sync.*
sudo systemctl daemon-reload

# Remove binary and helpers
sudo rm /opt/radb-client/radb-client
sudo rm /usr/local/bin/radb-client
sudo rm /usr/local/bin/radb-daemon

# Remove configuration (optional - may want to keep)
sudo rm -rf /etc/radb-client

# Remove data (optional - may want to keep)
sudo rm -rf /var/lib/radb-client

# Remove logs
sudo rm -rf /var/log/radb-client

# Remove logrotate config
sudo rm /etc/logrotate.d/radb-client

# Remove user (optional)
sudo userdel radb
```

### Keep Data for Reinstall

```bash
# Stop service but keep data
sudo systemctl stop radb-client
sudo systemctl disable radb-client

# Remove only binary and service files
sudo rm /etc/systemd/system/radb-client.service
sudo rm /opt/radb-client/radb-client

# Keep:
# - /etc/radb-client/ (configuration)
# - /var/lib/radb-client/ (data)
# - /var/log/radb-client/ (logs)
```

---

## Best Practices

### Monitoring

- Set up alerts for service failures
- Monitor disk space in data directories
- Track error rates in logs
- Monitor API rate limit usage

### Maintenance

- Regularly review logs for errors
- Verify snapshots are being created
- Check disk space monthly
- Test backup restoration periodically

### Configuration

- Use environment-specific config files
- Document any custom settings
- Keep check interval reasonable (not too frequent)
- Set appropriate retention policies

### Security

- Regularly update the binary
- Monitor for security advisories
- Audit file permissions periodically
- Rotate credentials as per policy

---

## Appendix: Systemd Service File

The complete service file (for reference):

```ini
[Unit]
Description=RADb Client Daemon - Route Object Monitoring
Documentation=https://github.com/bss/radb-client
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=radb
Group=radb

Environment="HOME=/var/lib/radb-client"
Environment="XDG_CONFIG_HOME=/etc/radb-client"

ExecStart=/opt/radb-client/radb-client daemon --config /etc/radb-client/config.yaml
ExecReload=/bin/kill -HUP $MAINPID

WorkingDirectory=/var/lib/radb-client

Restart=on-failure
RestartSec=30s

LimitNOFILE=65536

StandardOutput=append:/var/log/radb-client/radb-client.log
StandardError=append:/var/log/radb-client/radb-client-error.log
SyslogIdentifier=radb-client

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/radb-client /var/log/radb-client
ReadOnlyPaths=/etc/radb-client
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictRealtime=true
RestrictNamespaces=true

[Install]
WantedBy=multi-user.target
```

---

## Support

For issues or questions:
- Check logs: `radb-daemon logs`
- View documentation: `/opt/radb-client/docs/`
- Test manually: `sudo -u radb radb-client daemon --once`
- Review troubleshooting guide above

---

**Version**: 0.9.0-pre
**Last Updated**: 2025-10-29
