# Usage Examples

Real-world examples and common scenarios for the RADb API Client.

## Table of Contents

- [Getting Started Examples](#getting-started-examples)
- [Daily Operations](#daily-operations)
- [Change Tracking](#change-tracking)
- [Automation Scripts](#automation-scripts)
- [CI/CD Integration](#cicd-integration)
- [Bulk Operations](#bulk-operations)
- [Monitoring and Alerting](#monitoring-and-alerting)
- [Backup and Recovery](#backup-and-recovery)
- [Advanced Workflows](#advanced-workflows)

## Getting Started Examples

### Example 1: First Time Setup

Complete setup from scratch:

```bash
# Step 1: Install (assuming binary is in PATH)
which radb-client
# /usr/local/bin/radb-client

# Step 2: Initialize configuration
radb-client config init
# ✓ Created configuration file: ~/.radb-client/config.yaml
# ✓ Created cache directory: ~/.radb-client/cache
# ✓ Created history directory: ~/.radb-client/history

# Step 3: Authenticate
radb-client auth login
# Username: user@example.com
# API Key: ********
# ✓ Successfully authenticated
# ✓ Credentials stored in system keyring

# Step 4: Verify setup
radb-client auth status
# ✓ Authenticated as: user@example.com

# Step 5: List your routes
radb-client route list
# ROUTE              ORIGIN     DESCRIPTION
# 192.0.2.0/24       AS64500    Example route 1
```

### Example 2: Quick Configuration Check

Verify your configuration:

```bash
# Show current config
radb-client config show

# Test API connectivity
radb-client auth test

# Get a specific setting
radb-client config get api.timeout
```

## Daily Operations

### Example 3: Daily Route Check

Script to run daily to check for changes:

```bash
#!/bin/bash
# daily-route-check.sh

echo "=== RADb Route Check - $(date) ==="

# List routes and create snapshot
echo "Fetching current routes..."
radb-client route list > /tmp/routes.txt

# Check for changes
echo "Checking for changes..."
radb-client route diff

# Save exit code
DIFF_EXIT=$?

if [ $DIFF_EXIT -eq 0 ]; then
  echo "No changes detected"
elif [ $DIFF_EXIT -eq 1 ]; then
  echo "Changes detected! See diff above."
  # Optional: Send notification
  # send-alert.sh "RADb routes changed"
else
  echo "Error checking for changes"
  exit 1
fi

echo "=== Check complete ==="
```

Run daily with cron:
```cron
# Check RADb routes every day at 9 AM
0 9 * * * /home/user/scripts/daily-route-check.sh >> /var/log/radb-check.log 2>&1
```

### Example 4: Adding a New Route

Complete workflow for adding a route:

```bash
# Step 1: Create route definition
cat > new-route.json <<EOF
{
  "route": "203.0.113.0/24",
  "origin": "AS64502",
  "descr": "New customer - Acme Corp",
  "remarks": "Provisioned $(date +%Y-%m-%d)",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
EOF

# Step 2: Validate the route
radb-client validate route new-route.json
# ✓ Route validation passed

# Step 3: Create snapshot before change
radb-client snapshot create --note "Before adding Acme Corp route"

# Step 4: Create the route
radb-client route create new-route.json
# ✓ Route 203.0.113.0/24 created successfully

# Step 5: Verify it was created
radb-client route show 203.0.113.0/24

# Step 6: Check the diff
radb-client route diff
# Added Routes (1):
#   + 203.0.113.0/24 AS64502 - New customer - Acme Corp
```

### Example 5: Updating Route Description

Update an existing route:

```bash
# Step 1: Get current route
radb-client route show 192.0.2.0/24 --format json > route.json

# Step 2: Modify the description
jq '.descr = "Updated description - Migration to new AS"' route.json > updated-route.json

# Step 3: Preview the change
diff <(jq . route.json) <(jq . updated-route.json)

# Step 4: Apply the update
radb-client route update 192.0.2.0/24 updated-route.json
# ✓ Route 192.0.2.0/24 updated successfully

# Step 5: Verify
radb-client route show 192.0.2.0/24
```

## Change Tracking

### Example 6: Weekly Change Report

Generate a weekly change report:

```bash
#!/bin/bash
# weekly-report.sh

# Calculate date 7 days ago
SINCE=$(date -d '7 days ago' +%Y-%m-%dT%H:%M:%S)

echo "=== Weekly RADb Change Report ==="
echo "Period: $SINCE to $(date)"
echo

# Get changes
radb-client history show \
  --since "$SINCE" \
  --format json > changes.json

# Count changes by type
ADDED=$(jq '[.[] | select(.action == "added")] | length' changes.json)
REMOVED=$(jq '[.[] | select(.action == "removed")] | length' changes.json)
MODIFIED=$(jq '[.[] | select(.action == "modified")] | length' changes.json)

echo "Summary:"
echo "  Added: $ADDED"
echo "  Removed: $REMOVED"
  Modified: $MODIFIED"
echo
echo "Total changes: $((ADDED + REMOVED + MODIFIED))"

# Show details
echo
echo "Details:"
radb-client history show --since "$SINCE"
```

### Example 7: Compare Two Specific Points in Time

Compare snapshots from two different times:

```bash
# List available snapshots
radb-client snapshot list

# Compare specific snapshots
radb-client history diff \
  2025-10-01T00:00:00 \
  2025-10-31T23:59:59

# Export comparison to file
radb-client history diff \
  2025-10-01T00:00:00 \
  2025-10-31T23:59:59 \
  --format json > october-changes.json
```

## Automation Scripts

### Example 8: Bulk Route Creation

Create multiple routes from a CSV file:

```bash
#!/bin/bash
# bulk-create-routes.sh

CSV_FILE=$1

if [ ! -f "$CSV_FILE" ]; then
  echo "Usage: $0 <csv-file>"
  exit 1
fi

# CSV format: prefix,origin,description
# Example: 192.0.2.0/24,AS64500,Customer route

# Skip header and process each line
tail -n +2 "$CSV_FILE" | while IFS=',' read -r prefix origin description; do
  echo "Creating route: $prefix"

  # Create JSON
  cat > /tmp/route.json <<EOF
{
  "route": "$prefix",
  "origin": "$origin",
  "descr": "$description",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
EOF

  # Create route
  if radb-client route create /tmp/route.json --force; then
    echo "  ✓ Created successfully"
  else
    echo "  ✗ Failed to create"
  fi

  # Rate limiting - wait 1 second between requests
  sleep 1
done

echo "Bulk creation complete"
```

Usage:
```bash
# Create routes.csv
cat > routes.csv <<EOF
prefix,origin,description
203.0.113.0/24,AS64502,Customer A
203.0.113.64/26,AS64502,Customer B
203.0.113.128/25,AS64502,Customer C
EOF

# Run bulk creation
./bulk-create-routes.sh routes.csv
```

### Example 9: Automated Route Cleanup

Remove routes matching criteria:

```bash
#!/bin/bash
# cleanup-old-routes.sh

# Get all routes
radb-client route list --format json > all-routes.json

# Find routes with "deprecated" in description
jq -r '.[] | select(.descr | contains("deprecated")) | .route' all-routes.json | \
  while read -r route; do
    echo "Removing deprecated route: $route"
    radb-client route delete "$route" --force
  done
```

### Example 10: Sync Routes from File

Keep routes in sync with a source file:

```bash
#!/bin/bash
# sync-routes.sh

SOURCE_FILE=$1

if [ ! -f "$SOURCE_FILE" ]; then
  echo "Usage: $0 <routes.json>"
  exit 1
fi

# Get current routes
radb-client route list --format json > current-routes.json

# Compare and identify changes needed
# This is simplified - production version would be more robust

# Get routes that should exist
jq -r '.[].route' "$SOURCE_FILE" | sort > desired-routes.txt

# Get routes that currently exist
jq -r '.[].route' current-routes.json | sort > current-routes.txt

# Routes to add
comm -23 desired-routes.txt current-routes.txt > routes-to-add.txt

# Routes to remove
comm -13 desired-routes.txt current-routes.txt > routes-to-remove.txt

# Add missing routes
while read -r route; do
  echo "Adding route: $route"
  jq --arg route "$route" '.[] | select(.route == $route)' "$SOURCE_FILE" > /tmp/route.json
  radb-client route create /tmp/route.json --force
done < routes-to-add.txt

# Remove extra routes
while read -r route; do
  echo "Removing route: $route"
  radb-client route delete "$route" --force
done < routes-to-remove.txt

echo "Sync complete"
```

## CI/CD Integration

### Example 11: GitHub Actions Workflow

Deploy routes via GitHub Actions:

```yaml
# .github/workflows/deploy-routes.yml
name: Deploy RADb Routes

on:
  push:
    branches: [main]
    paths:
      - 'routes/**'

jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v2

      - name: Download radb-client
        run: |
          curl -L https://github.com/example/radb-client/releases/latest/download/radb-client-linux-amd64 \
            -o radb-client
          chmod +x radb-client
          sudo mv radb-client /usr/local/bin/

      - name: Verify installation
        run: radb-client --version

      - name: Authenticate
        env:
          RADB_USERNAME: ${{ secrets.RADB_USERNAME }}
          RADB_API_KEY: ${{ secrets.RADB_API_KEY }}
        run: |
          radb-client auth test

      - name: Create snapshot
        run: |
          radb-client snapshot create --note "Pre-deployment $(date)"

      - name: Deploy routes
        env:
          RADB_USERNAME: ${{ secrets.RADB_USERNAME }}
          RADB_API_KEY: ${{ secrets.RADB_API_KEY }}
        run: |
          for file in routes/*.json; do
            echo "Deploying $file"
            radb-client route create "$file" --force || exit 1
          done

      - name: Verify deployment
        run: |
          radb-client route diff
          radb-client route list

      - name: Notify on failure
        if: failure()
        run: |
          echo "Deployment failed!"
          # Send notification
```

### Example 12: GitLab CI Pipeline

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - deploy

variables:
  RADB_CLIENT_VERSION: "1.0.0"

before_script:
  - curl -L https://github.com/example/radb-client/releases/download/v${RADB_CLIENT_VERSION}/radb-client-linux-amd64 -o radb-client
  - chmod +x radb-client
  - export PATH="$PWD:$PATH"

validate:
  stage: validate
  script:
    - radb-client auth test
    - for file in routes/*.json; do radb-client validate route "$file"; done
  only:
    - merge_requests

deploy:
  stage: deploy
  script:
    - radb-client snapshot create --note "CI deployment $CI_COMMIT_SHORT_SHA"
    - for file in routes/*.json; do radb-client route create "$file" --force; done
    - radb-client route diff
  only:
    - main
  environment:
    name: production
```

## Bulk Operations

### Example 13: Export All Routes

Export routes for backup or analysis:

```bash
#!/bin/bash
# export-all-routes.sh

DATE=$(date +%Y%m%d-%H%M%S)
BACKUP_DIR="$HOME/radb-backups"

mkdir -p "$BACKUP_DIR"

# Export in multiple formats
echo "Exporting routes..."

# JSON format
radb-client route list --format json > "$BACKUP_DIR/routes-$DATE.json"
echo "✓ Exported JSON: routes-$DATE.json"

# YAML format
radb-client route list --format yaml > "$BACKUP_DIR/routes-$DATE.yaml"
echo "✓ Exported YAML: routes-$DATE.yaml"

# Table format (human-readable)
radb-client route list > "$BACKUP_DIR/routes-$DATE.txt"
echo "✓ Exported TXT: routes-$DATE.txt"

# Compress
tar -czf "$BACKUP_DIR/routes-$DATE.tar.gz" \
  "$BACKUP_DIR/routes-$DATE.json" \
  "$BACKUP_DIR/routes-$DATE.yaml" \
  "$BACKUP_DIR/routes-$DATE.txt"

# Remove uncompressed files
rm "$BACKUP_DIR/routes-$DATE."{json,yaml,txt}

echo "✓ Backup complete: routes-$DATE.tar.gz"
echo "Size: $(du -h "$BACKUP_DIR/routes-$DATE.tar.gz" | cut -f1)"
```

### Example 14: Update Multiple Routes

Update description for all routes in a specific AS:

```bash
#!/bin/bash
# update-as-description.sh

ASN=$1
NEW_DESC=$2

if [ -z "$ASN" ] || [ -z "$NEW_DESC" ]; then
  echo "Usage: $0 <ASN> <new-description>"
  exit 1
fi

# Get all routes for ASN
radb-client route list --format json | \
  jq --arg asn "$ASN" '.[] | select(.origin == $asn)' | \
  jq -c '.' | \
  while read -r route; do
    PREFIX=$(echo "$route" | jq -r '.route')
    echo "Updating $PREFIX..."

    # Update description
    echo "$route" | \
      jq --arg desc "$NEW_DESC" '.descr = $desc' > /tmp/updated-route.json

    # Apply update
    radb-client route update "$PREFIX" /tmp/updated-route.json --force

    # Small delay to respect rate limits
    sleep 0.5
  done

echo "Update complete"
```

Usage:
```bash
./update-as-description.sh AS64500 "Updated by automation"
```

## Monitoring and Alerting

### Example 15: Slack Alert on Changes

Send Slack notification when routes change:

```bash
#!/bin/bash
# monitor-with-slack.sh

SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

# Check for changes
DIFF_OUTPUT=$(radb-client route diff)
DIFF_EXIT=$?

if [ $DIFF_EXIT -eq 1 ]; then
  # Changes detected
  MESSAGE="RADb Route Changes Detected\n\n\`\`\`\n$DIFF_OUTPUT\n\`\`\`"

  # Send to Slack
  curl -X POST "$SLACK_WEBHOOK" \
    -H 'Content-Type: application/json' \
    -d "{\"text\": \"$MESSAGE\"}"

  echo "Alert sent to Slack"
elif [ $DIFF_EXIT -eq 0 ]; then
  echo "No changes detected"
else
  # Error occurred
  ERROR_MSG="Error checking RADb routes: exit code $DIFF_EXIT"

  curl -X POST "$SLACK_WEBHOOK" \
    -H 'Content-Type: application/json' \
    -d "{\"text\": \":warning: $ERROR_MSG\"}"

  echo "Error alert sent to Slack"
fi
```

### Example 16: Prometheus Metrics Export

Export metrics for Prometheus:

```bash
#!/bin/bash
# export-metrics.sh

METRICS_FILE="/var/lib/node_exporter/textfile_collector/radb.prom"

# Get route count
ROUTE_COUNT=$(radb-client route list --format json | jq 'length')

# Get last change time
LAST_CHANGE=$(radb-client history show --limit 1 --format json | \
  jq -r '.[0].timestamp // "never"')

LAST_CHANGE_TS=0
if [ "$LAST_CHANGE" != "never" ]; then
  LAST_CHANGE_TS=$(date -d "$LAST_CHANGE" +%s)
fi

# Write metrics
cat > "$METRICS_FILE" <<EOF
# HELP radb_routes_total Total number of routes in RADb
# TYPE radb_routes_total gauge
radb_routes_total $ROUTE_COUNT

# HELP radb_last_change_timestamp Unix timestamp of last detected change
# TYPE radb_last_change_timestamp gauge
radb_last_change_timestamp $LAST_CHANGE_TS

# HELP radb_last_check_timestamp Unix timestamp of last successful check
# TYPE radb_last_check_timestamp gauge
radb_last_check_timestamp $(date +%s)
EOF

echo "Metrics exported to $METRICS_FILE"
```

## Backup and Recovery

### Example 17: Automated Backup to S3

Backup RADb data to S3:

```bash
#!/bin/bash
# backup-to-s3.sh

S3_BUCKET="s3://my-backups/radb"
DATE=$(date +%Y%m%d)

# Create snapshot
radb-client snapshot create --note "Automated backup $DATE"

# Export routes
radb-client route list --format json | gzip > routes-$DATE.json.gz

# Export contacts
radb-client contact list --format json | gzip > contacts-$DATE.json.gz

# Export history
radb-client history show --format json | gzip > history-$DATE.json.gz

# Upload to S3
aws s3 cp routes-$DATE.json.gz "$S3_BUCKET/routes-$DATE.json.gz"
aws s3 cp contacts-$DATE.json.gz "$S3_BUCKET/contacts-$DATE.json.gz"
aws s3 cp history-$DATE.json.gz "$S3_BUCKET/history-$DATE.json.gz"

# Cleanup local files
rm routes-$DATE.json.gz contacts-$DATE.json.gz history-$DATE.json.gz

echo "Backup complete: $S3_BUCKET/"

# Remove backups older than 90 days
aws s3 ls "$S3_BUCKET/" | \
  awk '{print $4}' | \
  while read -r file; do
    FILE_DATE=$(echo "$file" | grep -oP '\d{8}' | head -1)
    if [ -n "$FILE_DATE" ]; then
      AGE=$(( ($(date +%s) - $(date -d "$FILE_DATE" +%s)) / 86400 ))
      if [ $AGE -gt 90 ]; then
        echo "Removing old backup: $file (${AGE} days old)"
        aws s3 rm "$S3_BUCKET/$file"
      fi
    fi
  done
```

### Example 18: Restore from Backup

Restore routes from backup:

```bash
#!/bin/bash
# restore-from-backup.sh

BACKUP_FILE=$1

if [ ! -f "$BACKUP_FILE" ]; then
  echo "Usage: $0 <backup-file.json>"
  exit 1
fi

echo "WARNING: This will restore routes from backup"
echo "Current routes may be affected"
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
  echo "Aborted"
  exit 0
fi

# Create pre-restore snapshot
radb-client snapshot create --note "Pre-restore snapshot $(date)"

# Import routes
radb-client route import "$BACKUP_FILE" --update-existing

echo "Restore complete"

# Show changes
radb-client route diff
```

## Advanced Workflows

### Example 19: Route Migration Between ASNs

Migrate routes from one ASN to another:

```bash
#!/bin/bash
# migrate-routes.sh

OLD_ASN=$1
NEW_ASN=$2

if [ -z "$OLD_ASN" ] || [ -z "$NEW_ASN" ]; then
  echo "Usage: $0 <old-asn> <new-asn>"
  exit 1
fi

echo "Migrating routes from $OLD_ASN to $NEW_ASN"

# Create snapshot
radb-client snapshot create --note "Before ASN migration $OLD_ASN to $NEW_ASN"

# Get routes for old ASN
radb-client route list --format json | \
  jq --arg old "$OLD_ASN" '.[] | select(.origin == $old)' | \
  jq -c '.' | \
  while read -r route; do
    PREFIX=$(echo "$route" | jq -r '.route')
    echo "Migrating $PREFIX from $OLD_ASN to $NEW_ASN"

    # Create new route with new ASN
    echo "$route" | \
      jq --arg new "$NEW_ASN" '.origin = $new' | \
      jq '.descr = (.descr + " (migrated from '"$OLD_ASN"')")' > /tmp/new-route.json

    # Create new route
    radb-client route create /tmp/new-route.json --force

    # Delete old route
    radb-client route delete "$PREFIX" --force

    echo "✓ Migrated $PREFIX"
    sleep 1
  done

echo "Migration complete"
radb-client route diff
```

### Example 20: Compliance Audit Report

Generate compliance audit report:

```bash
#!/bin/bash
# compliance-audit.sh

REPORT_FILE="radb-audit-$(date +%Y%m%d).txt"

{
  echo "==================================="
  echo "RADb Compliance Audit Report"
  echo "Generated: $(date)"
  echo "==================================="
  echo

  echo "1. Authentication Status"
  echo "------------------------"
  radb-client auth status
  echo

  echo "2. Current Routes"
  echo "-----------------"
  radb-client route list
  echo

  echo "3. Changes in Last 30 Days"
  echo "--------------------------"
  SINCE=$(date -d '30 days ago' +%Y-%m-%dT00:00:00)
  radb-client history show --since "$SINCE"
  echo

  echo "4. Snapshot History"
  echo "-------------------"
  radb-client snapshot list
  echo

  echo "5. Configuration"
  echo "----------------"
  radb-client config show
  echo

  echo "==================================="
  echo "End of Report"
  echo "==================================="
} > "$REPORT_FILE"

echo "Audit report generated: $REPORT_FILE"
```

## See Also

- [User Guide](USER_GUIDE.md) - Complete user guide
- [Commands](COMMANDS.md) - Command reference
- [Configuration](CONFIGURATION.md) - Configuration options
- [API Integration](API_INTEGRATION.md) - API details
