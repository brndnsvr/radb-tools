# Security Guide

Security practices and considerations for the RADb API Client.

## Table of Contents

- [Security Overview](#security-overview)
- [Credential Management](#credential-management)
- [Network Security](#network-security)
- [Data Security](#data-security)
- [Operational Security](#operational-security)
- [Audit and Compliance](#audit-and-compliance)
- [Security Best Practices](#security-best-practices)
- [Reporting Security Issues](#reporting-security-issues)

## Security Overview

The RADb API Client handles sensitive credentials and routing data. This guide covers security features and best practices.

### Security Features

1. **Secure Credential Storage** - System keyring integration
2. **Encrypted Fallback** - NaCl secretbox encryption
3. **HTTPS Only** - No insecure HTTP connections
4. **Certificate Validation** - TLS certificate verification
5. **No Logging of Secrets** - Credentials never logged
6. **Audit Trail** - All operations logged

### Threat Model

**Assets:**
- RADb API credentials
- Route configuration data
- Historical snapshots
- Configuration files

**Threats:**
- Credential theft
- Unauthorized access
- Man-in-the-middle attacks
- Data tampering
- Information disclosure

## Credential Management

### Primary Method: System Keyring

The client uses your operating system's secure credential storage:

**macOS:**
- Keychain Access
- Location: Login Keychain
- Encrypted by system

**Linux:**
- GNOME Keyring (GNOME)
- KWallet (KDE)
- Secret Service API
- Encrypted by system

**Windows:**
- Credential Manager
- Windows Data Protection API (DPAPI)
- Encrypted by system

### How It Works

```
User enters credentials
    ↓
Client attempts system keyring
    ├─ Success → Stored in keyring
    │           Encrypted by OS
    │           Accessible only to user
    └─ Failure → Fallback to encrypted file
```

**Advantages:**
- OS-level security
- Integration with system security policies
- No credentials in plain text
- Protected by user's login

### Fallback Method: Encrypted File

If keyring unavailable (headless systems, permissions issues):

**Location:** `~/.radb-client/credentials.enc`

**Encryption:**
- Algorithm: NaCl secretbox (XSalsa20-Poly1305)
- Key derivation: System-specific key
- 256-bit encryption

**Security Properties:**
- Authenticated encryption
- Protection against tampering
- Secure against chosen-ciphertext attacks

**File Permissions:**
```bash
# Automatically set to:
chmod 600 ~/.radb-client/credentials.enc
# Owner: read/write only
```

### Environment Variables (CI/CD)

For automated environments:

```bash
export RADB_USERNAME="user@example.com"
export RADB_API_KEY="your-api-key"
```

**Security Considerations:**
- Use secrets management (GitHub Secrets, AWS Secrets Manager)
- Never commit to version control
- Rotate regularly
- Limit access

### Credential Rotation

**Recommended rotation schedule:**
- Production: Every 90 days
- Development: Every 180 days
- Compromised: Immediately

**Rotation process:**
```bash
# 1. Generate new API key on RADb website
# 2. Update credentials
radb-client auth logout
radb-client auth login
# 3. Verify
radb-client auth test
# 4. Revoke old key on RADb website
```

### What's Never Stored

- Passwords in plain text
- Session tokens (if implemented)
- Temporary credentials
- API responses containing sensitive data

## Network Security

### HTTPS Only

**Configuration:**
```yaml
api:
  base_url: https://api.radb.net  # Always HTTPS
```

**Enforcement:**
- HTTP requests rejected
- No automatic downgrade
- Must use TLS 1.2 or higher

### Certificate Validation

**Default behavior:**
```yaml
advanced:
  verify_ssl: true  # Always true in production
```

**What's verified:**
- Certificate chain
- Certificate expiration
- Hostname matching
- Revocation status

**Never disable in production:**
```bash
# DON'T DO THIS in production
radb-client config set advanced.verify_ssl false
```

**When to disable:**
- Local development only
- Self-signed certificates
- Explicit testing scenarios

### TLS Configuration

**Supported versions:**
- TLS 1.2 (minimum)
- TLS 1.3 (preferred)

**Cipher suites:**
- Modern, secure ciphers only
- Forward secrecy enabled
- No weak or deprecated ciphers

### Network Proxies

**Proxy support:**
```bash
# HTTP proxy
export HTTP_PROXY=http://proxy.example.com:8080

# HTTPS proxy
export HTTPS_PROXY=https://proxy.example.com:8080

# No proxy for certain domains
export NO_PROXY=localhost,127.0.0.1
```

**Corporate proxies:**
- May intercept TLS
- Install corporate CA certificate
- Verify proxy security policies

### Rate Limiting Protection

**Client-side rate limiting:**
- Prevents accidental abuse
- Respects API limits
- Exponential backoff on errors

**Automatic handling:**
```
Request → Check rate limit
    ├─ Within limit → Proceed
    └─ Exceeded → Wait and retry
```

## Data Security

### Configuration File Security

**Location:** `~/.radb-client/config.yaml`

**Permissions:**
```bash
chmod 600 ~/.radb-client/config.yaml
# Owner: read/write only
```

**What's stored:**
- API endpoint
- Preferences
- Non-sensitive configuration

**What's NOT stored:**
- Credentials
- API keys
- Passwords

### Snapshot Security

**Location:** `~/.radb-client/cache/` and `~/.radb-client/history/`

**Permissions:**
```bash
chmod 700 ~/.radb-client/
# Owner: full access only
```

**Contents:**
- Route objects (not secret)
- Contact information
- Historical snapshots

**Security considerations:**
- May contain organizational data
- Protected by file permissions
- Regular cleanup recommended

### Data At Rest

**Encrypted:**
- Credentials (keyring or encrypted file)

**Not encrypted:**
- Configuration files
- Snapshots
- Logs

**Reason:** Route objects are public IRR data, not secret

**If additional encryption needed:**
```bash
# Encrypt entire directory
# Example with LUKS (Linux):
# Mount encrypted filesystem for ~/.radb-client/
```

### Data In Transit

**Always encrypted:**
- API requests (HTTPS)
- API responses (HTTPS)
- Certificate validation enabled

**Never sent over network:**
- Local snapshots
- Configuration files
- Credentials (only in Auth header, encrypted)

### Secure Deletion

**Remove credentials:**
```bash
radb-client auth logout
# Removes from keyring and encrypted file
```

**Remove all data:**
```bash
# Removes configuration, snapshots, history
rm -rf ~/.radb-client/
```

**Secure file deletion:**
```bash
# Use secure deletion tools if needed
shred -u ~/.radb-client/credentials.enc
```

## Operational Security

### Access Control

**File system permissions:**
```bash
# Automatically set by client
~/.radb-client/                 # drwx------ (700)
~/.radb-client/config.yaml      # -rw------- (600)
~/.radb-client/credentials.enc  # -rw------- (600)
```

**Multi-user systems:**
- Each user has own credentials
- No shared credentials
- Separate config per user

### Logging Security

**What's logged:**
- Commands executed
- API request metadata (URL, method)
- Response status codes
- Errors (sanitized)

**What's NEVER logged:**
- API keys
- Passwords
- Authentication headers
- Credential content

**Log security:**
```bash
# Logs are safe to share for debugging
radb-client --log-level DEBUG route list > debug.log
# No credentials in debug.log
```

### Process Security

**Memory protection:**
- Credentials cleared from memory after use
- No core dumps with credentials
- Secure zeroing of sensitive data

**Process isolation:**
- Runs with user privileges
- No elevated permissions required
- No background daemons

### Temporary Files

**Usage:**
- Minimal use of temp files
- Cleaned up automatically
- Secure permissions (600)

**Location:**
```bash
# System temp directory
/tmp/radb-client-*  # On Unix-like systems
```

## Audit and Compliance

### Audit Trail

**Change history:**
```bash
# All changes tracked
radb-client history show
```

**What's tracked:**
- Route creations
- Route modifications
- Route deletions
- Timestamps
- User actions

**Audit log format:**
```json
{
  "timestamp": "2025-10-29T12:00:00Z",
  "type": "route",
  "action": "created",
  "object": "192.0.2.0/24",
  "user": "user@example.com"
}
```

### Compliance Features

**Audit requirements:**
- All operations logged
- Timestamped actions
- User attribution
- Change history retention

**Data retention:**
```yaml
preferences:
  max_snapshots: 365  # Keep 1 year
```

**Export for compliance:**
```bash
# Export audit trail
radb-client history export audit-trail.json

# Export current state
radb-client route list --format json > current-state.json
```

### Security Monitoring

**Regular checks:**
```bash
# Check authentication
radb-client auth status

# Verify configuration
radb-client config validate

# Review recent changes
radb-client history show --since 7d
```

**Monitoring script:**
```bash
#!/bin/bash
# security-check.sh

# Check auth status
if ! radb-client auth status > /dev/null 2>&1; then
  echo "WARNING: Not authenticated"
fi

# Check for unexpected changes
CHANGES=$(radb-client route diff --format json)
if [ -n "$CHANGES" ]; then
  echo "WARNING: Unexpected route changes detected"
  echo "$CHANGES"
fi
```

## Security Best Practices

### For End Users

1. **Protect your credentials**
   ```bash
   # Never share API keys
   # Never commit credentials to git
   # Use separate keys for automation
   ```

2. **Rotate credentials regularly**
   ```bash
   # Every 90 days minimum
   radb-client auth login
   ```

3. **Monitor for unauthorized changes**
   ```bash
   # Daily checks
   radb-client route diff
   ```

4. **Keep software updated**
   ```bash
   # Check for updates
   radb-client --version
   # Update when new version available
   ```

5. **Use strong API keys**
   - Generated by RADb (strong random)
   - Never reuse passwords
   - Use unique keys per purpose

### For Administrators

1. **Implement least privilege**
   - Separate accounts for different purposes
   - Production vs development keys
   - Limited scope where possible

2. **Use secrets management**
   ```bash
   # CI/CD: Use GitHub Secrets, AWS Secrets Manager
   export RADB_USERNAME="${{ secrets.RADB_USERNAME }}"
   export RADB_API_KEY="${{ secrets.RADB_API_KEY }}"
   ```

3. **Monitor and alert**
   ```bash
   # Set up monitoring
   # Alert on unexpected changes
   # Review audit logs regularly
   ```

4. **Backup securely**
   ```bash
   # Encrypted backups
   radb-client route export backup.json
   gpg --encrypt backup.json
   ```

5. **Security reviews**
   - Regular security audits
   - Review access logs
   - Check for anomalies

### For Developers

1. **Never log credentials**
   ```go
   // DON'T
   log.Printf("API Key: %s", apiKey)

   // DO
   log.Printf("Authentication successful")
   ```

2. **Secure coding practices**
   - Input validation
   - Output encoding
   - Error handling

3. **Security testing**
   ```bash
   # Run security tests
   make test-security

   # Check for vulnerabilities
   go list -json -m all | nancy sleuth
   ```

4. **Dependency management**
   ```bash
   # Keep dependencies updated
   go get -u ./...

   # Check for known vulnerabilities
   go list -json -m all | nancy sleuth
   ```

### For CI/CD

1. **Use secrets management**
   ```yaml
   # GitHub Actions
   env:
     RADB_USERNAME: ${{ secrets.RADB_USERNAME }}
     RADB_API_KEY: ${{ secrets.RADB_API_KEY }}
   ```

2. **Separate credentials**
   - Production vs staging
   - Different keys per environment
   - Rotate regularly

3. **Audit CI/CD changes**
   - Review pipeline changes
   - Monitor secret access
   - Log all operations

4. **Secure runners**
   - Use trusted runners
   - Clean up after jobs
   - No persistent credentials

## Reporting Security Issues

### Responsible Disclosure

If you discover a security vulnerability:

1. **DO NOT** open a public GitHub issue
2. **DO** email security@example.com
3. **DO** provide details:
   - Description of vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Report

**Report these:**
- Authentication bypasses
- Credential leaks
- Injection vulnerabilities
- Unauthorized access
- Cryptographic weaknesses

**Not security issues:**
- Feature requests
- General bugs (use GitHub Issues)
- Questions (use Discussions)

### Response Timeline

- **24 hours:** Initial response
- **72 hours:** Preliminary assessment
- **7 days:** Patch development
- **14 days:** Patch release
- **30 days:** Public disclosure (coordinated)

### Hall of Fame

Contributors who responsibly disclose security issues will be acknowledged (with permission) in:
- SECURITY.md
- Release notes
- Hall of Fame

## Security Checklist

### Initial Setup
- [ ] Run `radb-client config init`
- [ ] Set proper permissions (automatic)
- [ ] Store credentials with `radb-client auth login`
- [ ] Verify with `radb-client auth test`

### Regular Maintenance
- [ ] Rotate credentials every 90 days
- [ ] Review audit logs monthly
- [ ] Check for software updates
- [ ] Clean up old snapshots
- [ ] Monitor for unauthorized changes

### Before Automation
- [ ] Use separate API keys
- [ ] Store in secrets manager
- [ ] Test in development first
- [ ] Set up monitoring
- [ ] Document security controls

### Security Incident Response
- [ ] Revoke compromised credentials immediately
- [ ] Review audit logs for unauthorized access
- [ ] Change all credentials
- [ ] Assess impact
- [ ] Document incident
- [ ] Implement preventive measures

## See Also

- [User Guide](USER_GUIDE.md) - General usage
- [Configuration](CONFIGURATION.md) - Configuration security
- [Architecture](ARCHITECTURE.md) - Security architecture
- [Troubleshooting](TROUBLESHOOTING.md) - Security issues
