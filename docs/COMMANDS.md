# Command Reference

Complete reference for all RADb API Client commands.

## Table of Contents

- [Global Flags](#global-flags)
- [Config Commands](#config-commands)
- [Auth Commands](#auth-commands)
- [Route Commands](#route-commands)
- [Contact Commands](#contact-commands)
- [Search Commands](#search-commands)
- [History Commands](#history-commands)
- [Snapshot Commands](#snapshot-commands)
- [Validation Commands](#validation-commands)

## Global Flags

These flags are available for all commands:

```bash
radb-client [global-flags] <command> [command-flags] [arguments]
```

### `--config <path>`

Specify configuration file location.

**Default:** `~/.radb-client/config.yaml`

**Example:**
```bash
radb-client --config /etc/radb/config.yaml route list
```

---

### `--verbose, -v`

Enable verbose output.

**Example:**
```bash
radb-client --verbose route list
```

---

### `--format <format>`

Output format for command results.

**Values:** `table`, `json`, `yaml`

**Default:** `table` (from config)

**Example:**
```bash
radb-client --format json route list
```

---

### `--log-level <level>`

Set logging level.

**Values:** `DEBUG`, `INFO`, `WARN`, `ERROR`

**Default:** `INFO` (from config)

**Example:**
```bash
radb-client --log-level DEBUG route list
```

---

### `--timeout <seconds>`

Request timeout in seconds.

**Default:** `30` (from config)

**Example:**
```bash
radb-client --timeout 60 route list
```

---

### `--no-color`

Disable colored output.

**Example:**
```bash
radb-client --no-color route list
```

---

### `--help, -h`

Show help for command.

**Example:**
```bash
radb-client --help
radb-client route --help
radb-client route list --help
```

---

### `--version`

Show version information.

**Example:**
```bash
radb-client --version
# Output: radb-client version 1.0.0
```

---

## Config Commands

Manage configuration settings.

### `radb-client config init`

Initialize configuration file with defaults.

**Usage:**
```bash
radb-client config init [flags]
```

**Flags:**
- `--force` - Overwrite existing configuration

**Examples:**
```bash
# Create default config
radb-client config init

# Force overwrite
radb-client config init --force

# Custom location
radb-client --config /path/to/config.yaml config init
```

**Creates:**
- Configuration file
- Cache directory
- History directory

---

### `radb-client config show`

Display current configuration.

**Usage:**
```bash
radb-client config show [section] [flags]
```

**Arguments:**
- `section` - Optional: `api`, `preferences`, `advanced`

**Flags:**
- `--format <format>` - Output format (`yaml`, `json`, `table`)
- `--sources` - Show value sources (file, env, default)

**Examples:**
```bash
# Show all configuration
radb-client config show

# Show specific section
radb-client config show api
radb-client config show preferences

# JSON output
radb-client config show --format json

# Show with sources
radb-client config show --sources
```

---

### `radb-client config set`

Set configuration value.

**Usage:**
```bash
radb-client config set <key> <value>
```

**Examples:**
```bash
# Set API timeout
radb-client config set api.timeout 60

# Set log level
radb-client config set preferences.log_level DEBUG

# Disable auto-snapshot
radb-client config set preferences.auto_snapshot false

# Set cache directory
radb-client config set preferences.cache_dir /var/cache/radb
```

---

### `radb-client config get`

Get configuration value.

**Usage:**
```bash
radb-client config get <key>
```

**Examples:**
```bash
radb-client config get api.timeout
# Output: 30

radb-client config get preferences.log_level
# Output: INFO
```

---

### `radb-client config validate`

Validate configuration file.

**Usage:**
```bash
radb-client config validate
```

**Checks:**
- YAML syntax
- Required fields
- Valid value types
- Value ranges
- Directory permissions

**Example:**
```bash
radb-client config validate
# ✓ Configuration is valid
```

---

### `radb-client config reset`

Reset configuration to defaults.

**Usage:**
```bash
radb-client config reset [section|key]
```

**Examples:**
```bash
# Reset all configuration
radb-client config reset

# Reset specific section
radb-client config reset api
radb-client config reset preferences

# Reset specific value
radb-client config reset api.timeout
```

---

## Auth Commands

Manage authentication credentials.

### `radb-client auth login`

Store authentication credentials.

**Usage:**
```bash
radb-client auth login [flags]
```

**Flags:**
- `--username <email>` - RADb username (email)
- `--api-key <key>` - RADb API key

**Interactive:**
```bash
radb-client auth login
# Username: user@example.com
# API Key: ********
# ✓ Successfully authenticated
```

**Non-interactive:**
```bash
radb-client auth login \
  --username user@example.com \
  --api-key your-api-key
```

**From environment:**
```bash
export RADB_USERNAME=user@example.com
export RADB_API_KEY=your-api-key
radb-client auth login
```

---

### `radb-client auth status`

Check authentication status.

**Usage:**
```bash
radb-client auth status
```

**Example output:**
```
✓ Authenticated as: user@example.com
  Storage: system keyring (macOS Keychain)
  Last verified: 2025-10-29 12:00:00
```

**Exit codes:**
- `0` - Authenticated successfully
- `1` - Not authenticated or credentials invalid

---

### `radb-client auth logout`

Remove stored credentials.

**Usage:**
```bash
radb-client auth logout
```

**Example:**
```bash
radb-client auth logout
# ✓ Credentials removed
```

**Notes:**
- Removes from keyring or encrypted file
- Config file unchanged
- Requires re-authentication for API calls

---

### `radb-client auth test`

Test authentication with API.

**Usage:**
```bash
radb-client auth test
```

**Example:**
```bash
radb-client auth test
# Testing authentication...
# ✓ Authentication successful
# ✓ API accessible
```

---

## Route Commands

Manage route objects (IPv4 and IPv6).

### `radb-client route list`

List all route objects.

**Usage:**
```bash
radb-client route list [flags]
```

**Flags:**
- `--format <format>` - Output format (`table`, `json`, `yaml`)
- `--filter <query>` - Filter results
- `--sort <field>` - Sort by field
- `--reverse` - Reverse sort order
- `--limit <n>` - Limit results
- `--no-snapshot` - Don't create snapshot

**Examples:**
```bash
# List all routes (table format)
radb-client route list

# JSON output
radb-client route list --format json

# Filter by AS number
radb-client route list --filter "AS64500"

# Sort by route
radb-client route list --sort route

# Limit results
radb-client route list --limit 10

# Without creating snapshot
radb-client route list --no-snapshot
```

**Table output:**
```
ROUTE              ORIGIN     DESCRIPTION                    MAINTAINER
192.0.2.0/24       AS64500    Example route 1                MAINT-EXAMPLE
198.51.100.0/24    AS64501    Example route 2                MAINT-EXAMPLE
2001:db8::/32      AS64500    IPv6 example route             MAINT-EXAMPLE
```

---

### `radb-client route show`

Show details of a specific route.

**Usage:**
```bash
radb-client route show <prefix> [flags]
```

**Arguments:**
- `prefix` - IP prefix (e.g., `192.0.2.0/24`)

**Flags:**
- `--format <format>` - Output format

**Examples:**
```bash
# Show IPv4 route
radb-client route show 192.0.2.0/24

# Show IPv6 route
radb-client route show 2001:db8::/32

# JSON output
radb-client route show 192.0.2.0/24 --format json
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
  admin-c: ADMIN-CONTACT
  tech-c: TECH-CONTACT
  mnt-by: MAINT-EXAMPLE
  source: RADB
```

---

### `radb-client route create`

Create a new route object.

**Usage:**
```bash
radb-client route create <file> [flags]
```

**Arguments:**
- `file` - JSON file with route data (use `-` for stdin)

**Flags:**
- `--validate` - Validate without creating
- `--force` - Skip confirmation

**Route JSON format:**
```json
{
  "route": "192.0.2.0/24",
  "origin": "AS64500",
  "descr": "My new route",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
```

**Examples:**
```bash
# Create from file
radb-client route create route.json

# Create from stdin
echo '{"route":"192.0.2.0/24",...}' | radb-client route create -

# Validate only
radb-client route create route.json --validate

# Skip confirmation
radb-client route create route.json --force
```

**IPv6 example:**
```json
{
  "route": "2001:db8::/32",
  "origin": "AS64500",
  "descr": "My IPv6 route",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
```

---

### `radb-client route update`

Update an existing route object.

**Usage:**
```bash
radb-client route update <prefix> <file> [flags]
```

**Arguments:**
- `prefix` - Route prefix to update
- `file` - JSON file with updated route data

**Flags:**
- `--validate` - Validate without updating
- `--force` - Skip confirmation

**Examples:**
```bash
# Update route
radb-client route update 192.0.2.0/24 updated-route.json

# Validate only
radb-client route update 192.0.2.0/24 updated-route.json --validate

# Skip confirmation
radb-client route update 192.0.2.0/24 updated-route.json --force
```

---

### `radb-client route delete`

Delete a route object.

**Usage:**
```bash
radb-client route delete <prefix> [flags]
```

**Arguments:**
- `prefix` - Route prefix to delete

**Flags:**
- `--force` - Skip confirmation

**Examples:**
```bash
# Delete with confirmation
radb-client route delete 192.0.2.0/24

# Skip confirmation
radb-client route delete 192.0.2.0/24 --force

# Delete IPv6 route
radb-client route delete 2001:db8::/32
```

**Warning:** Deletion is immediate and cannot be undone.

---

### `radb-client route diff`

Show changes since last snapshot.

**Usage:**
```bash
radb-client route diff [flags]
```

**Flags:**
- `--since <timestamp>` - Compare with specific snapshot
- `--format <format>` - Output format
- `--summary` - Show summary only

**Examples:**
```bash
# Show changes since last snapshot
radb-client route diff

# Compare with specific snapshot
radb-client route diff --since 2025-10-29T12:00:00

# Summary only
radb-client route diff --summary

# JSON output
radb-client route diff --format json
```

**Example output:**
```
Changes since 2025-10-29 12:00:00:

Added Routes (2):
  + 203.0.113.0/24 AS64502
    Description: New customer route

  + 2001:db8:100::/48 AS64503
    Description: New IPv6 allocation

Removed Routes (1):
  - 198.51.100.0/24 AS64501
    Description: Decommissioned route

Modified Routes (1):
  ~ 192.0.2.0/24 AS64500
    - descr: "Old description"
    + descr: "Updated description"

Summary: 2 added, 1 removed, 1 modified
```

---

### `radb-client route export`

Export routes to file.

**Usage:**
```bash
radb-client route export <file> [flags]
```

**Arguments:**
- `file` - Output file (use `-` for stdout)

**Flags:**
- `--format <format>` - Export format (`json`, `yaml`, `rpsl`)
- `--filter <query>` - Filter routes to export

**Examples:**
```bash
# Export all routes to JSON
radb-client route export routes.json

# Export to stdout
radb-client route export - --format json

# Export specific AS
radb-client route export as64500-routes.json --filter "AS64500"

# Export as RPSL
radb-client route export routes.txt --format rpsl
```

---

### `radb-client route import`

Import routes from file.

**Usage:**
```bash
radb-client route import <file> [flags]
```

**Arguments:**
- `file` - Input file with routes

**Flags:**
- `--validate` - Validate without importing
- `--skip-existing` - Skip routes that already exist
- `--update-existing` - Update routes that already exist
- `--dry-run` - Show what would be imported

**Examples:**
```bash
# Import routes
radb-client route import routes.json

# Validate only
radb-client route import routes.json --validate

# Skip existing
radb-client route import routes.json --skip-existing

# Update existing
radb-client route import routes.json --update-existing

# Dry run
radb-client route import routes.json --dry-run
```

---

## Contact Commands

Manage account contacts.

### `radb-client contact list`

List all contacts.

**Usage:**
```bash
radb-client contact list [flags]
```

**Flags:**
- `--format <format>` - Output format
- `--role <role>` - Filter by role (`admin`, `tech`, `billing`)

**Examples:**
```bash
# List all contacts
radb-client contact list

# JSON output
radb-client contact list --format json

# Technical contacts only
radb-client contact list --role tech
```

**Example output:**
```
ID              NAME            EMAIL                   ROLE         ORGANIZATION
CONTACT-1       John Doe        john@example.com        admin        Example Networks
CONTACT-2       Jane Smith      jane@example.com        tech         Example Networks
```

---

### `radb-client contact show`

Show contact details.

**Usage:**
```bash
radb-client contact show <id> [flags]
```

**Arguments:**
- `id` - Contact ID

**Flags:**
- `--format <format>` - Output format

**Examples:**
```bash
radb-client contact show CONTACT-1
radb-client contact show CONTACT-1 --format json
```

---

### `radb-client contact create`

Create a new contact.

**Usage:**
```bash
radb-client contact create <file> [flags]
```

**Arguments:**
- `file` - JSON file with contact data

**Contact JSON format:**
```json
{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "phone": "+1-555-0123",
  "role": "technical",
  "organization": "Example Networks"
}
```

**Examples:**
```bash
radb-client contact create contact.json
```

---

### `radb-client contact update`

Update an existing contact.

**Usage:**
```bash
radb-client contact update <id> <file> [flags]
```

**Arguments:**
- `id` - Contact ID
- `file` - JSON file with updated contact data

**Examples:**
```bash
radb-client contact update CONTACT-1 updated-contact.json
```

---

### `radb-client contact delete`

Delete a contact.

**Usage:**
```bash
radb-client contact delete <id> [flags]
```

**Arguments:**
- `id` - Contact ID

**Flags:**
- `--force` - Skip confirmation

**Examples:**
```bash
radb-client contact delete CONTACT-1
radb-client contact delete CONTACT-1 --force
```

---

## Search Commands

Search the IRR database.

### `radb-client search`

Search for objects in the database.

**Usage:**
```bash
radb-client search <query> [flags]
```

**Arguments:**
- `query` - Search query string

**Flags:**
- `--type <type>` - Filter by object type (`route`, `aut-num`, `as-set`, etc.)
- `--format <format>` - Output format
- `--limit <n>` - Limit results

**Examples:**
```bash
# Search for IP prefix
radb-client search "192.0.2.0/24"

# Search for AS number
radb-client search "AS64500"

# Search with type filter
radb-client search "64500" --type route

# Limit results
radb-client search "example" --limit 10

# JSON output
radb-client search "AS64500" --format json
```

**Example output:**
```
TYPE      OBJECT                DESCRIPTION
route     192.0.2.0/24          AS64500 - Example route 1
route     2001:db8::/32         AS64500 - IPv6 example route
aut-num   AS64500               Example Networks
```

---

## History Commands

View change history and compare snapshots.

### `radb-client history show`

Show change history.

**Usage:**
```bash
radb-client history show [flags]
```

**Flags:**
- `--since <date>` - Show changes since date
- `--until <date>` - Show changes until date
- `--type <type>` - Filter by type (`route`, `contact`)
- `--action <action>` - Filter by action (`added`, `removed`, `modified`)
- `--limit <n>` - Limit results
- `--format <format>` - Output format

**Examples:**
```bash
# Show all history
radb-client history show

# Since specific date
radb-client history show --since 2025-10-01

# Date range
radb-client history show --since 2025-10-01 --until 2025-10-31

# Route changes only
radb-client history show --type route

# Added routes only
radb-client history show --action added

# Last 10 changes
radb-client history show --limit 10

# JSON output
radb-client history show --format json
```

**Example output:**
```
TIMESTAMP            TYPE    ACTION     OBJECT                      DETAILS
2025-10-29 14:30:00  route   added      203.0.113.0/24 AS64502     New customer
2025-10-29 12:15:00  route   modified   192.0.2.0/24 AS64500       Description updated
2025-10-28 16:45:00  route   removed    198.51.100.0/24 AS64501    Decommissioned
2025-10-28 10:00:00  route   added      2001:db8:100::/48 AS64503  IPv6 allocation
```

---

### `radb-client history diff`

Compare two snapshots.

**Usage:**
```bash
radb-client history diff <timestamp1> <timestamp2> [flags]
```

**Arguments:**
- `timestamp1` - First snapshot timestamp
- `timestamp2` - Second snapshot timestamp

**Flags:**
- `--format <format>` - Output format

**Examples:**
```bash
# Compare two snapshots
radb-client history diff 2025-10-29T12:00:00 2025-10-29T18:00:00

# JSON output
radb-client history diff 2025-10-29T12:00:00 2025-10-29T18:00:00 --format json
```

---

### `radb-client history export`

Export change history.

**Usage:**
```bash
radb-client history export <file> [flags]
```

**Arguments:**
- `file` - Output file

**Flags:**
- `--since <date>` - Export since date
- `--until <date>` - Export until date
- `--format <format>` - Export format

**Examples:**
```bash
# Export all history
radb-client history export history.json

# Export date range
radb-client history export oct-history.json \
  --since 2025-10-01 \
  --until 2025-10-31
```

---

## Snapshot Commands

Manage snapshots.

### `radb-client snapshot create`

Create a manual snapshot.

**Usage:**
```bash
radb-client snapshot create [flags]
```

**Flags:**
- `--note <note>` - Add note to snapshot
- `--type <type>` - Snapshot type (`route`, `contact`, `full`)

**Examples:**
```bash
# Create snapshot
radb-client snapshot create

# With note
radb-client snapshot create --note "Before major routing change"

# Routes only
radb-client snapshot create --type route
```

---

### `radb-client snapshot list`

List all snapshots.

**Usage:**
```bash
radb-client snapshot list [flags]
```

**Flags:**
- `--type <type>` - Filter by type
- `--format <format>` - Output format
- `--limit <n>` - Limit results

**Examples:**
```bash
# List all snapshots
radb-client snapshot list

# Route snapshots only
radb-client snapshot list --type route

# JSON output
radb-client snapshot list --format json
```

**Example output:**
```
TIMESTAMP                TYPE     SIZE      NOTE
2025-10-29T18:00:00     route    1.2 MB    Before major routing change
2025-10-29T12:00:00     route    1.2 MB    Automatic snapshot
2025-10-28T18:30:00     route    1.1 MB    Pre-deployment check
```

---

### `radb-client snapshot show`

Show snapshot contents.

**Usage:**
```bash
radb-client snapshot show <timestamp> [flags]
```

**Arguments:**
- `timestamp` - Snapshot timestamp

**Flags:**
- `--format <format>` - Output format

**Examples:**
```bash
radb-client snapshot show 2025-10-29T12:00:00
radb-client snapshot show 2025-10-29T12:00:00 --format json
```

---

### `radb-client snapshot delete`

Delete a snapshot.

**Usage:**
```bash
radb-client snapshot delete <timestamp> [flags]
```

**Arguments:**
- `timestamp` - Snapshot timestamp

**Flags:**
- `--force` - Skip confirmation

**Examples:**
```bash
radb-client snapshot delete 2025-10-29T12:00:00
radb-client snapshot delete 2025-10-29T12:00:00 --force
```

---

### `radb-client snapshot cleanup`

Clean up old snapshots.

**Usage:**
```bash
radb-client snapshot cleanup [flags]
```

**Flags:**
- `--keep <n>` - Keep N most recent snapshots
- `--older-than <days>` - Delete snapshots older than N days
- `--dry-run` - Show what would be deleted

**Examples:**
```bash
# Keep last 50 snapshots
radb-client snapshot cleanup --keep 50

# Delete snapshots older than 90 days
radb-client snapshot cleanup --older-than 90

# Dry run
radb-client snapshot cleanup --keep 50 --dry-run
```

---

## Validation Commands

Validate objects and data.

### `radb-client validate asn`

Validate an AS number.

**Usage:**
```bash
radb-client validate asn <asn> [flags]
```

**Arguments:**
- `asn` - AS number to validate

**Flags:**
- `--format <format>` - Output format

**Examples:**
```bash
radb-client validate asn AS64500
radb-client validate asn AS64500 --format json
```

**Example output:**
```
AS Number: AS64500
Valid: Yes
AS Name: EXAMPLE-AS
Organization: Example Networks
```

---

### `radb-client validate route`

Validate a route object.

**Usage:**
```bash
radb-client validate route <file> [flags]
```

**Arguments:**
- `file` - JSON file with route data

**Flags:**
- `--format <format>` - Output format

**Examples:**
```bash
radb-client validate route route.json
```

---

## Exit Codes

All commands use standard exit codes:

- `0` - Success
- `1` - General error
- `2` - Command usage error
- `3` - Authentication error
- `4` - Network error
- `5` - Validation error
- `6` - Not found error
- `7` - Configuration error

**Example usage in scripts:**
```bash
#!/bin/bash
radb-client route list
if [ $? -eq 0 ]; then
  echo "Success"
else
  echo "Failed with exit code $?"
fi
```

## Shell Completion

Enable command completion for your shell:

**Bash:**
```bash
radb-client completion bash > /etc/bash_completion.d/radb-client
```

**Zsh:**
```bash
radb-client completion zsh > ~/.zsh/completion/_radb-client
```

**Fish:**
```bash
radb-client completion fish > ~/.config/fish/completions/radb-client.fish
```

## See Also

- [User Guide](USER_GUIDE.md) - Getting started and workflows
- [Configuration](CONFIGURATION.md) - Configuration options
- [Examples](EXAMPLES.md) - Usage examples
- [API Integration](API_INTEGRATION.md) - API details
