# Troubleshooting Guide

Common issues and solutions for the RADb API Client.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Authentication Issues](#authentication-issues)
- [Connection Problems](#connection-problems)
- [Configuration Issues](#configuration-issues)
- [Command Errors](#command-errors)
- [Snapshot and History Issues](#snapshot-and-history-issues)
- [Performance Problems](#performance-problems)
- [Platform-Specific Issues](#platform-specific-issues)
- [Error Messages](#error-messages)
- [Getting Help](#getting-help)

## Quick Diagnostics

### Run These First

When experiencing issues, start with these diagnostic commands:

```bash
# 1. Check version
radb-client --version

# 2. Verify authentication
radb-client auth status

# 3. Test API connectivity
radb-client auth test

# 4. Check configuration
radb-client config validate

# 5. Enable debug logging
radb-client --log-level DEBUG route list
```

### Collect Information

Before reporting issues, gather this information:

```bash
# System information
uname -a
radb-client --version

# Configuration
radb-client config show

# Recent errors
radb-client --log-level DEBUG <failing-command> 2>&1 | tail -50
```

## Authentication Issues

### Issue: "Authentication failed: invalid username or API key"

**Symptoms:**
```
Error: Authentication failed: invalid username or API key
```

**Causes:**
- Wrong username
- Wrong API key
- Expired API key
- Credentials not stored

**Solutions:**

1. **Verify credentials:**
   ```bash
   radb-client auth status
   ```

2. **Re-authenticate:**
   ```bash
   radb-client auth logout
   radb-client auth login
   ```

3. **Check API key on RADb website:**
   - Log into https://www.radb.net
   - Go to Account → Settings → API Access
   - Verify or regenerate API key

4. **Use environment variables:**
   ```bash
   export RADB_USERNAME="user@example.com"
   export RADB_API_KEY="your-api-key"
   radb-client auth test
   ```

---

### Issue: "Failed to store credentials in keyring"

**Symptoms:**
```
Warning: Failed to store credentials in system keyring
Falling back to encrypted file storage
```

**Causes:**
- Keyring service not available
- Permissions issue
- Keyring locked

**Solutions:**

**Linux:**
```bash
# Install keyring service
# For GNOME:
sudo apt install gnome-keyring

# For KDE:
sudo apt install kwalletmanager

# For headless systems, encrypted file fallback is used automatically
```

**macOS:**
- Keychain should work automatically
- If issues persist, unlock Keychain:
  ```bash
  security unlock-keychain ~/Library/Keychains/login.keychain
  ```

**Windows:**
- Credential Manager should work automatically
- Check Windows Services for "Credential Manager"

**Workaround:**
```bash
# Use environment variables instead
export RADB_USERNAME="user@example.com"
export RADB_API_KEY="your-api-key"
```

---

### Issue: "No credentials found"

**Symptoms:**
```
Error: No credentials found. Please run 'radb-client auth login'
```

**Solutions:**

1. **Authenticate:**
   ```bash
   radb-client auth login
   ```

2. **Or use environment variables:**
   ```bash
   export RADB_USERNAME="user@example.com"
   export RADB_API_KEY="your-api-key"
   ```

---

## Connection Problems

### Issue: "Connection timeout"

**Symptoms:**
```
Error: Request timeout after 30 seconds
```

**Causes:**
- Slow network
- Firewall blocking HTTPS
- API service issues

**Solutions:**

1. **Increase timeout:**
   ```bash
   radb-client --timeout 60 route list
   ```

   Or permanently:
   ```bash
   radb-client config set api.timeout 60
   ```

2. **Check network connectivity:**
   ```bash
   curl -I https://api.radb.net
   ping api.radb.net
   ```

3. **Check firewall:**
   ```bash
   # Ensure HTTPS (443) is allowed
   telnet api.radb.net 443
   ```

4. **Check proxy settings:**
   ```bash
   echo $HTTP_PROXY
   echo $HTTPS_PROXY
   ```

---

### Issue: "SSL certificate verification failed"

**Symptoms:**
```
Error: SSL certificate verification failed
```

**Causes:**
- System CA certificates outdated
- Corporate proxy intercepting SSL
- System time incorrect

**Solutions:**

1. **Update CA certificates:**
   ```bash
   # Ubuntu/Debian
   sudo apt update && sudo apt install ca-certificates

   # CentOS/RHEL
   sudo yum update ca-certificates

   # macOS
   # Usually automatic, but can run:
   sudo security update
   ```

2. **Check system time:**
   ```bash
   date
   # Should be accurate to avoid certificate validation issues
   ```

3. **Corporate proxy:**
   - Install corporate CA certificate
   - Or configure proxy settings

4. **Temporary bypass (NOT recommended for production):**
   ```bash
   radb-client config set advanced.verify_ssl false
   ```

---

### Issue: "Rate limit exceeded"

**Symptoms:**
```
Error: Too many requests. Rate limit exceeded.
Retry after: 60 seconds
```

**Causes:**
- Too many API requests in short time
- Multiple concurrent operations
- Bulk operations without delays

**Solutions:**

1. **Wait and retry:**
   ```bash
   # Client automatically retries after the specified delay
   ```

2. **Add delays in scripts:**
   ```bash
   for route in routes/*.json; do
     radb-client route create "$route"
     sleep 2  # Wait 2 seconds between requests
   done
   ```

3. **Reduce concurrent operations:**
   ```bash
   # Instead of parallel operations, run sequentially
   ```

4. **Use bulk operations:**
   ```bash
   # Use import instead of multiple creates
   radb-client route import routes.json
   ```

---

## Configuration Issues

### Issue: "Configuration file not found"

**Symptoms:**
```
Error: Configuration file not found: ~/.radb-client/config.yaml
```

**Solutions:**

1. **Initialize configuration:**
   ```bash
   radb-client config init
   ```

2. **Specify custom location:**
   ```bash
   radb-client --config /path/to/config.yaml config init
   ```

---

### Issue: "Invalid configuration"

**Symptoms:**
```
Error: Failed to parse configuration: yaml: line 5: mapping values are not allowed
```

**Causes:**
- Invalid YAML syntax
- Wrong indentation
- Tabs instead of spaces

**Solutions:**

1. **Validate configuration:**
   ```bash
   radb-client config validate
   ```

2. **Check YAML syntax:**
   - Use online validator: https://www.yamllint.com/
   - Ensure proper indentation (2 spaces, no tabs)

3. **Reset to defaults:**
   ```bash
   # Backup first
   cp ~/.radb-client/config.yaml ~/.radb-client/config.yaml.backup

   # Reset
   radb-client config reset
   ```

---

### Issue: "Permission denied on configuration file"

**Symptoms:**
```
Error: Permission denied: ~/.radb-client/config.yaml
```

**Solutions:**

```bash
# Fix permissions
chmod 600 ~/.radb-client/config.yaml
chown $USER ~/.radb-client/config.yaml
```

---

### Issue: "Failed to create cache directory"

**Symptoms:**
```
Error: Failed to create directory ~/.radb-client/cache: permission denied
```

**Solutions:**

1. **Fix permissions:**
   ```bash
   mkdir -p ~/.radb-client/cache
   chmod 700 ~/.radb-client
   ```

2. **Use alternative location:**
   ```bash
   radb-client config set preferences.cache_dir ~/tmp/radb-cache
   radb-client config set preferences.history_dir ~/tmp/radb-history
   ```

---

## Command Errors

### Issue: "Route not found"

**Symptoms:**
```
Error: Route not found: 192.0.2.0/24
```

**Solutions:**

1. **Verify route exists:**
   ```bash
   radb-client route list | grep "192.0.2.0/24"
   ```

2. **Check prefix format:**
   ```bash
   # Correct format:
   radb-client route show 192.0.2.0/24

   # Not:
   radb-client route show 192.0.2.0  # Missing /24
   ```

3. **Search for the route:**
   ```bash
   radb-client search "192.0.2.0"
   ```

---

### Issue: "Validation failed"

**Symptoms:**
```
Error: Route validation failed:
  - route: Invalid CIDR notation
  - origin: Must be valid AS number (ASN)
```

**Causes:**
- Invalid IP prefix format
- Invalid AS number format
- Missing required fields

**Solutions:**

1. **Check required fields:**
   ```json
   {
     "route": "192.0.2.0/24",    // Required: Valid CIDR
     "origin": "AS64500",        // Required: ASN format
     "mnt-by": ["MAINT-EXAMPLE"], // Required: Array
     "source": "RADB"            // Required: "RADB"
   }
   ```

2. **Validate before creating:**
   ```bash
   radb-client validate route route.json
   ```

3. **Check examples:**
   ```bash
   # IPv4
   radb-client route show 192.0.2.0/24 --format json > example.json

   # IPv6
   radb-client route show 2001:db8::/32 --format json > example-ipv6.json
   ```

---

### Issue: "Route already exists"

**Symptoms:**
```
Error: Route already exists: 192.0.2.0/24
Use 'route update' to modify existing route
```

**Solutions:**

1. **Update instead of create:**
   ```bash
   radb-client route update 192.0.2.0/24 updated-route.json
   ```

2. **Or delete and recreate:**
   ```bash
   radb-client route delete 192.0.2.0/24
   radb-client route create new-route.json
   ```

---

## Snapshot and History Issues

### Issue: "No snapshots found"

**Symptoms:**
```
Error: No snapshots found
Cannot generate diff without previous snapshot
```

**Causes:**
- First run (no previous snapshots)
- Snapshots deleted
- Using different cache directory

**Solutions:**

1. **Create initial snapshot:**
   ```bash
   radb-client route list  # Creates first snapshot
   ```

2. **Check snapshot directory:**
   ```bash
   ls -lh ~/.radb-client/history/
   ```

3. **Verify configuration:**
   ```bash
   radb-client config get preferences.history_dir
   ```

---

### Issue: "Corrupted snapshot"

**Symptoms:**
```
Warning: Failed to load snapshot from 2025-10-29T12:00:00
Snapshot may be corrupted
```

**Solutions:**

1. **Delete corrupted snapshot:**
   ```bash
   radb-client snapshot delete 2025-10-29T12:00:00
   ```

2. **Create fresh snapshot:**
   ```bash
   radb-client route list --no-snapshot  # Don't use broken snapshot
   radb-client snapshot create --note "Fresh snapshot"
   ```

3. **Clean up and rebuild:**
   ```bash
   # Backup first
   mv ~/.radb-client/history ~/.radb-client/history.backup

   # Recreate
   mkdir -p ~/.radb-client/history
   radb-client route list
   ```

---

### Issue: "Disk space full"

**Symptoms:**
```
Error: Failed to create snapshot: no space left on device
```

**Solutions:**

1. **Check disk space:**
   ```bash
   df -h ~/.radb-client/
   ```

2. **Clean up old snapshots:**
   ```bash
   radb-client snapshot cleanup --keep 50
   ```

3. **Reduce snapshot retention:**
   ```bash
   radb-client config set preferences.max_snapshots 30
   ```

4. **Use different location:**
   ```bash
   radb-client config set preferences.history_dir /large/disk/radb-history
   ```

---

## Performance Problems

### Issue: "Slow commands"

**Symptoms:**
- Commands take long to complete
- Timeouts

**Causes:**
- Large number of routes
- Slow network
- API performance issues

**Solutions:**

1. **Increase timeout:**
   ```bash
   radb-client config set api.timeout 120
   ```

2. **Use JSON format (faster parsing):**
   ```bash
   radb-client route list --format json
   ```

3. **Disable auto-snapshot for bulk operations:**
   ```bash
   radb-client config set preferences.auto_snapshot false
   # ... perform operations ...
   radb-client config set preferences.auto_snapshot true
   ```

4. **Check API status:**
   ```bash
   curl -I https://api.radb.net
   ```

---

### Issue: "High memory usage"

**Symptoms:**
- Client consuming excessive memory
- System becomes slow

**Causes:**
- Large datasets
- Many snapshots

**Solutions:**

1. **Limit results:**
   ```bash
   radb-client route list --limit 1000
   ```

2. **Clean up snapshots:**
   ```bash
   radb-client snapshot cleanup --keep 20
   ```

3. **Use streaming for large operations:**
   ```bash
   # Process in smaller chunks
   radb-client route list --format json | jq -c '.[] | select(...)'
   ```

---

## Platform-Specific Issues

### Linux Issues

**Issue: Keyring not available on headless systems**

**Solution:**
- Encrypted file fallback is used automatically
- Or use environment variables

**Issue: Permission denied on /tmp**

**Solution:**
```bash
# Use home directory for temporary files
export TMPDIR=~/tmp
mkdir -p ~/tmp
```

---

### macOS Issues

**Issue: Keychain keeps asking for password**

**Solution:**
```bash
# Allow radb-client to access Keychain
security add-generic-password -a $USER -s radb-client -w
```

**Issue: Binary won't run (security warning)**

**Solution:**
```bash
# Allow running of downloaded binary
xattr -d com.apple.quarantine radb-client
```

---

### Windows Issues

**Issue: Command not found**

**Solution:**
- Ensure binary is in PATH
- Or use full path: `C:\path\to\radb-client.exe`

**Issue: Credential Manager not accessible**

**Solution:**
- Run as Administrator
- Or use environment variables

---

## Error Messages

### Exit Code Reference

- `0` - Success
- `1` - General error
- `2` - Command usage error
- `3` - Authentication error
- `4` - Network error
- `5` - Validation error
- `6` - Not found error
- `7` - Configuration error

**Usage in scripts:**
```bash
radb-client route list
case $? in
  0) echo "Success" ;;
  3) echo "Authentication failed" ;;
  4) echo "Network error" ;;
  *) echo "Other error: $?" ;;
esac
```

---

### Common Error Patterns

**"context deadline exceeded"**
- Timeout occurred
- Solution: Increase timeout with `--timeout 60`

**"connection refused"**
- Cannot connect to API
- Solution: Check network connectivity and firewall

**"invalid character"**
- JSON parsing error
- Solution: Validate JSON syntax

**"permission denied"**
- File/directory permission issue
- Solution: Check and fix permissions

---

## Getting Help

### Enable Debug Logging

```bash
# Debug mode shows detailed information
radb-client --log-level DEBUG <command>

# Save debug output
radb-client --log-level DEBUG route list 2>&1 | tee debug.log
```

### Check Logs

```bash
# If using systemd service
journalctl -u radb-client -f

# Manual runs
tail -f ~/.radb-client/logs/radb-client.log
```

### Collect Diagnostic Information

```bash
#!/bin/bash
# collect-diagnostics.sh

echo "=== RADb Client Diagnostics ==="
echo

echo "System Information:"
uname -a
echo

echo "Client Version:"
radb-client --version
echo

echo "Configuration:"
radb-client config show
echo

echo "Authentication Status:"
radb-client auth status
echo

echo "API Test:"
radb-client auth test
echo

echo "Configuration Validation:"
radb-client config validate
echo

echo "Disk Space:"
df -h ~/.radb-client/
echo

echo "Snapshots:"
radb-client snapshot list
echo
```

### Report Issues

When reporting issues, include:

1. **System information:**
   - OS and version
   - Client version

2. **Command that failed:**
   ```bash
   radb-client --log-level DEBUG <failing-command>
   ```

3. **Error message:**
   - Full error output
   - Stack trace if available

4. **Configuration:**
   ```bash
   radb-client config show
   ```

5. **Steps to reproduce:**
   - What you were trying to do
   - What happened
   - What you expected

### Resources

- **Documentation:** docs/ directory
- **Examples:** [EXAMPLES.md](EXAMPLES.md)
- **GitHub Issues:** Report bugs and request features
- **RADb Support:** support@radb.net for API issues

### Quick Reference

**Diagnostic Commands:**
```bash
radb-client --version              # Check version
radb-client auth status            # Check authentication
radb-client config validate        # Validate config
radb-client --log-level DEBUG ...  # Debug mode
```

**Common Fixes:**
```bash
# Re-authenticate
radb-client auth logout && radb-client auth login

# Reset configuration
radb-client config reset

# Clean up snapshots
radb-client snapshot cleanup --keep 50

# Fix permissions
chmod 600 ~/.radb-client/config.yaml
chmod 700 ~/.radb-client/
```

**Still Stuck?**

1. Check the [User Guide](USER_GUIDE.md)
2. Review [Examples](EXAMPLES.md)
3. Search [GitHub Issues](https://github.com/example/radb-client/issues)
4. Ask for help (provide diagnostic info above)

## See Also

- [User Guide](USER_GUIDE.md) - Complete usage guide
- [Examples](EXAMPLES.md) - Usage examples
- [Configuration](CONFIGURATION.md) - Configuration options
- [Security](SECURITY.md) - Security best practices
