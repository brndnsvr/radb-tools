# RADb API Authentication Guide

Comprehensive guide to authentication methods supported by the RADb API.

---

## Overview

The RADb API supports two primary authentication methods:
1. **HTTP Basic Authentication** (recommended for API access)
2. **IRR Password Authentication** (query parameter-based)

---

## Authentication Methods

### 1. HTTP Basic Authentication

**Type:** Standard HTTP Basic Auth
**Scheme:** `Basic <base64-encoded-credentials>`

#### How It Works

HTTP Basic Authentication encodes your RADb portal username and password in Base64 format and sends it in the `Authorization` header.

**Format:**
```
Authorization: Basic <base64(username:password)>
```

**Example:**
```bash
# If username is: user@example.com
# And password is: MySecurePassword

# Encode: user@example.com:MySecurePassword
# Result: dXNlckBleGFtcGxlLmNvbTpNeVNlY3VyZVBhc3N3b3Jk

# HTTP Header:
Authorization: Basic dXNlckBleGFtcGxlLmNvbTpNeVNlY3VyZVBhc3N3b3Jk
```

#### Credentials Required

- **Username:** Your RADb portal username (usually your email)
- **Password:** Your RADb portal password

#### When to Use

- ✅ All API operations (GET, POST, PUT, DELETE)
- ✅ Route object management
- ✅ Contact management
- ✅ Most secure method for API access
- ✅ Recommended for programmatic access

---

### 2. IRR Password Authentication

**Type:** API Key (query parameter)
**Parameter:** `password`

#### How It Works

Pass the maintainer password as a query parameter in the URL.

**Format:**
```
https://api.radb.net/v1/route?password=YourMaintainerPassword
```

#### Credentials Required

- **Maintainer Password:** The crypted password from your RADb maintainer object

#### When to Use

- ⚠️ Legacy systems
- ⚠️ Simple GET requests
- ❌ Not recommended for production (password in URL/logs)

**Note:** For GET requests, "any non-empty string may be used for password authentication" according to the OpenAPI spec. However, for modifications (POST/PUT/DELETE), the actual maintainer password is required.

---

## Authentication Requirements by Operation

### GET Operations (Read)

**Required:** Either Basic Auth OR IRR Password
**Security Level:** Low to Medium

Examples:
- List routes
- View route details
- List contacts
- ASN validation (no auth required)

### POST/PUT Operations (Create/Update)

**Required:** Basic Auth with valid maintainer credentials
**Security Level:** High

**Important:** The credentials must be associated with the object's maintainer (`mntner` field).

Examples:
- Create new route object
- Update existing route
- Create contact

### DELETE Operations (Remove)

**Required:** Basic Auth with valid maintainer credentials
**Security Level:** High

**Important:** Must have authorization for the object's maintainer.

Examples:
- Delete route object
- Remove contact

---

## Current Implementation in radb-client

### What We Store

The radb-client stores:
1. **Username** - In config file (`~/.radb-client/config.yaml`)
2. **Password** - Encrypted in system keyring or encrypted file

### Encryption Details

**Method:** Argon2id + NaCl Secretbox
**Storage:**
- **Primary:** System keyring (if available)
- **Fallback:** Encrypted file (`~/.radb-client/credentials.enc`)

**Security Features:**
- Argon2id key derivation (industry standard)
- NaCl secretbox for encryption
- SHA-256 checksums for integrity
- Secure memory handling

### How Authentication Works

When you run `radb-client auth login`:

1. Prompt for username and password
2. Derive encryption key using Argon2id
3. Encrypt password with NaCl secretbox
4. Try to store in system keyring
5. If keyring fails, store in encrypted file
6. Save username to config file

When making API requests:

1. Load username from config
2. Retrieve encrypted password from keyring/file
3. Decrypt password
4. Encode `username:password` in Base64
5. Add `Authorization: Basic <base64>` header to request

---

## Recommended Authentication Flow

### For End Users

```bash
# 1. Initialize configuration
radb-client config init

# 2. Login with RADb credentials
radb-client auth login
# Enter username: your.email@example.com
# Enter password: ********

# 3. Verify authentication
radb-client auth status

# 4. Make authenticated requests
radb-client route list
```

### For Automation/CI/CD

**Option 1: Pre-configured credentials**
```bash
# Set up once
radb-client auth login

# Use in scripts
radb-client route list
radb-client snapshot create
```

**Option 2: Environment variables** (future enhancement)
```bash
export RADB_USERNAME="user@example.com"
export RADB_PASSWORD="securepassword"
radb-client route list
```

---

## Security Best Practices

### DO ✅

1. **Use HTTP Basic Auth** for all API access
2. **Store credentials securely** (system keyring preferred)
3. **Use HTTPS only** (default: https://api.radb.net)
4. **Rotate passwords regularly**
5. **Use strong passwords** for RADb portal account
6. **Encrypt credentials at rest** (radb-client does this)
7. **Limit credential access** (file permissions 600)

### DON'T ❌

1. **Don't use IRR password in query params** for production
2. **Don't commit credentials** to version control
3. **Don't share credentials** between users
4. **Don't log passwords** in application logs
5. **Don't use HTTP** (always HTTPS)
6. **Don't store passwords in plain text**
7. **Don't hardcode passwords** in scripts

---

## Credential Types Explained

### RADb Portal Credentials

**Username:** Your email address used to log into radb.net
**Password:** Password for radb.net portal
**Used for:** API authentication via HTTP Basic Auth
**Where to get:** Create account at https://www.radb.net/

**This is what radb-client uses.**

### Maintainer (mntner) Credentials

**Type:** RPSL authentication credential
**Format:** Various (CRYPT-PW, PGPKEY, MAIL-FROM)
**Used for:** Authorizing changes to RPSL objects
**Where to find:** In your RADb account mntner object

**Note:** When you use HTTP Basic Auth with portal credentials, the API validates that you have authority over the maintainer associated with the objects you're modifying.

---

## Authentication Errors

### Common Errors

**401 Unauthorized**
- Invalid username or password
- Credentials not provided
- Base64 encoding issue

**403 Forbidden**
- Valid credentials but no permission for operation
- Not authorized for the object's maintainer
- Trying to modify someone else's objects

**400 Bad Request**
- Missing required password parameter (IRR method)
- Malformed authorization header

---

## Example API Calls

### Using curl with Basic Auth

```bash
# List routes
curl -X GET "https://api.radb.net/v1/route" \
  -H "Authorization: Basic $(echo -n 'user@example.com:password' | base64)"

# Create route
curl -X POST "https://api.radb.net/v1/route" \
  -H "Authorization: Basic $(echo -n 'user@example.com:password' | base64)" \
  -H "Content-Type: application/json" \
  -d '{
    "route": "192.0.2.0/24",
    "origin": "AS65000",
    "descr": "Example route",
    "mnt-by": "MAINT-EXAMPLE"
  }'
```

### Using Python with requests

```python
import requests
from requests.auth import HTTPBasicAuth

# Configure
username = "user@example.com"
password = "your_password"
base_url = "https://api.radb.net/v1"

# List routes
response = requests.get(
    f"{base_url}/route",
    auth=HTTPBasicAuth(username, password)
)

# Create route
route_data = {
    "route": "192.0.2.0/24",
    "origin": "AS65000",
    "descr": "Example route",
    "mnt-by": "MAINT-EXAMPLE"
}

response = requests.post(
    f"{base_url}/route",
    auth=HTTPBasicAuth(username, password),
    json=route_data
)
```

### Using radb-client

```bash
# Authenticate once
radb-client auth login

# All subsequent commands use stored credentials
radb-client route list
radb-client route create --route "192.0.2.0/24" --origin "AS65000"
radb-client contact list
```

---

## Troubleshooting

### Cannot Authenticate

**Check:**
1. Username is correct (case-sensitive)
2. Password is correct
3. RADb portal account is active
4. Network can reach api.radb.net
5. HTTPS is being used (not HTTP)

**Test:**
```bash
# Verify credentials with curl
curl -X GET "https://api.radb.net/v1/route" \
  -u "your.email@example.com:your_password"
```

### Credentials Not Saving

**Check:**
1. Config directory exists (`~/.radb-client/`)
2. Permissions are correct (700 for directory)
3. Keyring daemon is running (Linux)
4. Fallback to encrypted file works

**Test:**
```bash
# Check config directory
ls -la ~/.radb-client/

# Try login again
radb-client auth login

# Check status
radb-client auth status
```

### 403 Forbidden on Modifications

**Check:**
1. You are authenticated (not just 401)
2. Your RADb account has the correct maintainer
3. The object's `mnt-by` field matches your maintainer
4. You have permission for this maintainer

**Resolution:**
- Contact RADb support to verify maintainer assignments
- Ensure your RADb portal account is linked to the correct maintainer

---

## Future Enhancements

Potential authentication improvements for radb-client:

1. **API Key Support** - If RADb adds API key authentication
2. **OAuth 2.0** - If RADb implements OAuth
3. **Environment Variables** - `RADB_USERNAME` and `RADB_PASSWORD`
4. **Token Refresh** - If RADb adds token-based auth
5. **MFA Support** - If RADb enables two-factor authentication
6. **Service Accounts** - Dedicated API credentials

---

## References

- **RADb API Documentation:** https://api.radb.net/docs.html
- **RADb OpenAPI Spec:** https://api.radb.net/RADB_API_OpenAPI.yaml
- **RADb Portal:** https://www.radb.net/
- **RADb Authentication Types:** https://www.radb.net/support/informational/authentication/
- **Example Scripts:**
  - https://www.radb.net/radb_api_demo.py
  - https://www.radb.net/radb_api_demo_telstra_prod.py

---

## Summary

**For radb-client users:**
- Authentication uses HTTP Basic Auth
- Credentials are your RADb portal username and password
- Password is encrypted and stored securely
- One-time setup with `radb-client auth login`
- All subsequent commands use stored credentials automatically

**Current implementation is correct and secure.** ✅

The radb-client properly implements HTTP Basic Authentication as required by the RADb API, with strong encryption for credential storage.
