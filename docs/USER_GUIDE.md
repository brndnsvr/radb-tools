# RADb API Client - User Guide

## Table of Contents

- [Introduction](#introduction)
- [Getting Started](#getting-started)
- [First-Time Setup](#first-time-setup)
- [Basic Operations](#basic-operations)
- [Understanding Change Tracking](#understanding-change-tracking)
- [Common Workflows](#common-workflows)
- [Advanced Usage](#advanced-usage)
- [Tips and Best Practices](#tips-and-best-practices)

## Introduction

The RADb API Client is a command-line tool for managing your RADb (Routing Assets Database) resources without using the web interface. It provides:

- Route object management (IPv4 and IPv6)
- Account contact management
- Automatic change detection between runs
- Historical snapshots with audit trail
- Secure credential management

### Who Should Use This Tool?

- Network administrators managing routing policies
- DevOps teams automating RADb operations
- Organizations requiring audit trails of routing changes
- Anyone preferring CLI over web interfaces

### Prerequisites

- RADb account with API access
- Your RADb username and API key
- Linux, macOS, or Windows system

## Getting Started

### Installation

See [INSTALL.md](../INSTALL.md) for detailed installation instructions for your platform.

Quick check if installed correctly:

```bash
radb-client --version
```

### Your First Commands

```bash
# Initialize configuration
radb-client config init

# Authenticate with RADb
radb-client auth login

# List your route objects
radb-client route list
```

## First-Time Setup

### Step 1: Initialize Configuration

Create the configuration file and directory structure:

```bash
radb-client config init
```

This creates:
- `~/.radb-client/config.yaml` - Configuration file
- `~/.radb-client/cache/` - Current state cache
- `~/.radb-client/history/` - Historical snapshots

### Step 2: Configure API Settings (Optional)

The default settings work for most users, but you can customize:

```bash
# View current configuration
radb-client config show

# Change specific settings
radb-client config set api.timeout 60
radb-client config set preferences.log_level DEBUG
```

See [CONFIGURATION.md](CONFIGURATION.md) for all available options.

### Step 3: Authenticate

Store your RADb credentials securely:

```bash
radb-client auth login
```

You'll be prompted for:
- **Username**: Your RADb email address
- **API Key**: Your RADb API key (not your web password)

**Where to find your API key:**
1. Log into RADb web interface at https://www.radb.net
2. Navigate to Account Settings
3. Find or generate your API key

**Security note:** Your credentials are stored in your system's secure keyring (Keychain on macOS, Credential Manager on Windows, Secret Service on Linux). They are never logged or exposed in plain text.

### Step 4: Verify Authentication

Check that authentication is working:

```bash
radb-client auth status
```

Expected output:
```
Authenticated as: user@example.com
Keyring: system (macOS Keychain)
Last verified: 2025-10-29 12:00:00
```

### Step 5: Test Basic Operation

Fetch your route objects:

```bash
radb-client route list
```

If this works, you're ready to use the tool!

## Basic Operations

### Listing Routes

View all your route objects:

```bash
# Table format (default)
radb-client route list

# JSON format
radb-client route list --format json

# YAML format
radb-client route list --format yaml
```

**Example output (table):**
```
ROUTE              ORIGIN     DESCRIPTION                    MAINTAINER
192.0.2.0/24       AS64500    Example route 1                MAINT-EXAMPLE
198.51.100.0/24    AS64501    Example route 2                MAINT-EXAMPLE
2001:db8::/32      AS64500    IPv6 example route             MAINT-EXAMPLE
```

### Viewing a Specific Route

Get details about a single route:

```bash
radb-client route show 192.0.2.0/24
```

**Example output:**
```
Route: 192.0.2.0/24
Origin: AS64500
Description: Example route for documentation
Maintainer: MAINT-EXAMPLE
Source: RADB
Created: 2025-10-15 10:30:00
Last Modified: 2025-10-28 14:22:00

Attributes:
  remarks: Example route
  mnt-by: MAINT-EXAMPLE
  source: RADB
```

### Creating Routes

Create a new route object from a JSON file:

```bash
radb-client route create route.json
```

**Example route.json:**
```json
{
  "route": "192.0.2.0/24",
  "origin": "AS64500",
  "descr": "My new route",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
```

**For IPv6 routes:**
```json
{
  "route": "2001:db8::/32",
  "origin": "AS64500",
  "descr": "My IPv6 route",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
```

### Updating Routes

Modify an existing route:

```bash
radb-client route update 192.0.2.0/24 updated-route.json
```

**Example updated-route.json:**
```json
{
  "route": "192.0.2.0/24",
  "origin": "AS64500",
  "descr": "Updated description",
  "remarks": "Added remarks field",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
```

### Deleting Routes

Remove a route object:

```bash
radb-client route delete 192.0.2.0/24
```

**Warning:** This action cannot be undone. The route will be removed from RADb immediately.

**Safety tip:** Before deleting, create a snapshot:
```bash
radb-client snapshot create --note "Before deleting 192.0.2.0/24"
radb-client route delete 192.0.2.0/24
```

## Understanding Change Tracking

One of the most powerful features is automatic change detection.

### How It Works

1. **First run:** When you list routes, the client saves a snapshot
2. **Subsequent runs:** The client compares current state with the previous snapshot
3. **Diff generation:** Changes are automatically detected and logged

### Viewing Changes

See what changed since your last check:

```bash
radb-client route diff
```

**Example output:**
```
Changes detected since 2025-10-29 12:00:00:

Added Routes (2):
  + 203.0.113.0/24 AS64502 - New customer route
  + 2001:db8:100::/48 AS64503 - New IPv6 allocation

Removed Routes (1):
  - 198.51.100.0/24 AS64501 - Decommissioned route

Modified Routes (1):
  ~ 192.0.2.0/24 AS64500
    - descr: "Old description"
    + descr: "Updated description"
    - remarks: ""
    + remarks: "Added important note"

Summary: 2 added, 1 removed, 1 modified
```

### Snapshots

Snapshots are automatic by default, but you can also create them manually:

```bash
# Create a snapshot with a note
radb-client snapshot create --note "Before major routing change"

# List all snapshots
radb-client snapshot list

# View a specific snapshot
radb-client snapshot show 2025-10-29T12-00-00
```

**Example snapshot list:**
```
TIMESTAMP                TYPE          NOTE
2025-10-29T12:00:00     route         Before major routing change
2025-10-29T09:00:00     route         Automatic snapshot
2025-10-28T18:30:00     route         Pre-deployment check
```

### Change History

View all detected changes over time:

```bash
# All changes
radb-client history show

# Changes since a specific date
radb-client history show --since 2025-10-01

# Only route changes (exclude contacts)
radb-client history show --type route

# Last 10 changes
radb-client history show --limit 10
```

**Example output:**
```
TIMESTAMP            TYPE    ACTION     OBJECT                      DETAILS
2025-10-29 14:30:00  route   added      203.0.113.0/24 AS64502     New customer
2025-10-29 12:15:00  route   modified   192.0.2.0/24 AS64500       Description updated
2025-10-28 16:45:00  route   removed    198.51.100.0/24 AS64501    Decommissioned
2025-10-28 10:00:00  route   added      2001:db8:100::/48 AS64503  IPv6 allocation
```

## Common Workflows

### Daily Operations Workflow

Monitor your routes daily:

```bash
#!/bin/bash
# daily-check.sh - Run this daily via cron

# List current routes (creates snapshot)
radb-client route list > /tmp/routes.txt

# Check for changes
radb-client route diff

# If changes detected, investigate
if [ $? -eq 1 ]; then
  echo "Changes detected! Review the diff above."
  # Optional: Send notification
fi
```

Schedule with cron:
```
0 9 * * * /path/to/daily-check.sh
```

### Bulk Route Management

Managing multiple routes at once:

```bash
# Export current routes
radb-client route list --format json > all-routes.json

# Modify the file (add/remove/update routes)
vim all-routes.json

# Create new routes from file
cat new-routes.json | jq -c '.[]' | while read route; do
  echo "$route" > /tmp/route.json
  radb-client route create /tmp/route.json
done
```

### Pre-Deployment Checklist

Before making routing changes:

```bash
# 1. Create a snapshot
radb-client snapshot create --note "Pre-deployment $(date +%Y-%m-%d)"

# 2. Export current state
radb-client route list --format json > backup-$(date +%Y%m%d).json

# 3. Make changes
radb-client route create new-route.json

# 4. Verify changes
radb-client route diff

# 5. If something goes wrong, you have the backup
```

### Detecting Unauthorized Changes

Monitor for unexpected changes:

```bash
# Check if routes were modified
radb-client route diff > /tmp/changes.txt

# Alert if any changes found
if [ -s /tmp/changes.txt ]; then
  echo "WARNING: Unauthorized route changes detected!"
  cat /tmp/changes.txt
  # Send alert email, Slack notification, etc.
fi
```

### Managing Contacts

Contacts are similar to routes:

```bash
# List all contacts
radb-client contact list

# View specific contact
radb-client contact show CONTACT-ID

# Create new contact
radb-client contact create contact.json

# Update contact
radb-client contact update CONTACT-ID updated.json

# Delete contact
radb-client contact delete CONTACT-ID
```

**Example contact.json:**
```json
{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "phone": "+1-555-0123",
  "role": "technical",
  "organization": "Example Networks"
}
```

### Historical Analysis

Analyze changes over time:

```bash
# All changes in October 2025
radb-client history show --since 2025-10-01 --until 2025-10-31

# Compare two specific points in time
radb-client history diff 2025-10-01T00:00:00 2025-10-31T23:59:59

# Export history to JSON
radb-client history show --format json > history.json
```

## Advanced Usage

### Filtering and Searching

Search for specific routes:

```bash
# Search for routes containing "customer"
radb-client search "customer"

# Search for specific ASN
radb-client search "AS64500"

# Validate an ASN
radb-client validate asn AS64500
```

### Custom Output Formatting

Customize table output:

```bash
# Show only specific columns
radb-client route list --columns route,origin,descr

# Sort by origin ASN
radb-client route list --sort origin

# Reverse sort
radb-client route list --sort origin --reverse
```

### Configuration Profiles

Manage multiple environments:

```bash
# Development environment
radb-client --config ~/.radb-client/config-dev.yaml route list

# Production environment
radb-client --config ~/.radb-client/config-prod.yaml route list

# Set via environment variable
export RADB_CONFIG=/path/to/config.yaml
radb-client route list
```

### Verbose and Debug Output

Troubleshooting:

```bash
# Verbose output
radb-client --verbose route list

# Debug mode (shows API calls)
radb-client config set preferences.log_level DEBUG
radb-client route list

# Or use flag
radb-client --log-level DEBUG route list
```

### Batch Operations with Scripts

Process multiple routes efficiently:

```bash
#!/bin/bash
# update-descriptions.sh

# Update descriptions for all routes in AS64500
radb-client route list --format json | \
  jq -c '.[] | select(.origin == "AS64500")' | \
  while read route; do
    prefix=$(echo "$route" | jq -r '.route')
    echo "$route" | \
      jq '.descr = "Updated description"' > /tmp/update.json

    radb-client route update "$prefix" /tmp/update.json
    echo "Updated $prefix"
  done
```

### Automated Backups

Create automated backups:

```bash
#!/bin/bash
# backup-radb.sh

BACKUP_DIR="/backups/radb"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p "$BACKUP_DIR"

# Create snapshot
radb-client snapshot create --note "Automated backup $DATE"

# Export routes
radb-client route list --format json > "$BACKUP_DIR/routes-$DATE.json"

# Export contacts
radb-client contact list --format json > "$BACKUP_DIR/contacts-$DATE.json"

# Export history
radb-client history show --format json > "$BACKUP_DIR/history-$DATE.json"

# Keep only last 30 days
find "$BACKUP_DIR" -name "*.json" -mtime +30 -delete

echo "Backup completed: $BACKUP_DIR"
```

### Integration with CI/CD

Use in automated pipelines:

```bash
# In your CI/CD pipeline
export RADB_USERNAME="${RADB_USER}"
export RADB_API_KEY="${RADB_KEY}"

# Deploy route changes
radb-client route create deployment/routes.json

# Verify deployment
radb-client route diff

# Check for expected state
radb-client route list --format json | \
  jq -e '.[] | select(.route == "192.0.2.0/24")' || exit 1
```

## Tips and Best Practices

### Security Best Practices

1. **Never hardcode credentials**
   - Use `radb-client auth login` to store securely
   - Use environment variables in CI/CD

2. **Protect your config file**
   ```bash
   chmod 600 ~/.radb-client/config.yaml
   ```

3. **Regular credential rotation**
   - Rotate API keys periodically
   - Update with: `radb-client auth login`

4. **Audit trail**
   - Review history regularly: `radb-client history show`
   - Monitor for unauthorized changes

### Performance Tips

1. **Use JSON format for scripting**
   ```bash
   radb-client route list --format json | jq '.[]'
   ```

2. **Disable auto-snapshots for bulk operations**
   ```bash
   radb-client config set preferences.auto_snapshot false
   # ... perform bulk operations ...
   radb-client config set preferences.auto_snapshot true
   ```

3. **Limit snapshot retention**
   ```bash
   radb-client config set preferences.max_snapshots 50
   ```

### Workflow Recommendations

1. **Always create snapshots before major changes**
   ```bash
   radb-client snapshot create --note "Before migration"
   ```

2. **Use descriptive commit notes**
   - Helps with historical analysis
   - Makes audits easier

3. **Test changes in development first**
   - Use separate config profiles
   - Verify with `route diff`

4. **Set up monitoring**
   - Daily diff checks
   - Alert on unexpected changes
   - Regular backup schedule

### Troubleshooting Tips

1. **Check authentication first**
   ```bash
   radb-client auth status
   ```

2. **Enable verbose output**
   ```bash
   radb-client --verbose route list
   ```

3. **Verify configuration**
   ```bash
   radb-client config show
   ```

4. **Check snapshot integrity**
   ```bash
   radb-client snapshot list
   ```

For detailed troubleshooting, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

### Common Mistakes to Avoid

1. **Don't modify snapshots manually**
   - Snapshots are managed by the tool
   - Manual changes will break diff functionality

2. **Don't share credentials**
   - Each user should have their own API key
   - Use separate accounts for automation

3. **Don't ignore diff output**
   - Unexpected changes may indicate issues
   - Always investigate unknown modifications

4. **Don't delete snapshots unless necessary**
   - They're used for change detection
   - Disk space is managed automatically

## Getting Help

### Command Help

```bash
# General help
radb-client --help

# Command-specific help
radb-client route --help
radb-client route list --help
```

### Documentation

- [Command Reference](COMMANDS.md) - Complete command documentation
- [Configuration Guide](CONFIGURATION.md) - All configuration options
- [Examples](EXAMPLES.md) - More usage examples
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues and solutions
- [API Integration](API_INTEGRATION.md) - RADb API details

### Support

- GitHub Issues: Report bugs or request features
- Documentation: Check the docs/ directory
- RADb API Docs: https://api.radb.net/docs.html

## What's Next?

- Explore the [Command Reference](COMMANDS.md) for all available commands
- Read [Examples](EXAMPLES.md) for more real-world scenarios
- Learn about [Configuration](CONFIGURATION.md) options
- Understand the [Architecture](ARCHITECTURE.md) for advanced usage

Happy routing!
