# RADb API Integration

## Table of Contents

- [Overview](#overview)
- [RADb API Basics](#radb-api-basics)
- [Authentication](#authentication)
- [API Endpoints](#api-endpoints)
- [Data Formats](#data-formats)
- [Request and Response Examples](#request-and-response-examples)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Best Practices](#best-practices)

## Overview

This document describes how the RADb API Client integrates with the RADb (Routing Assets Database) API. It's useful for:

- Understanding how the client works under the hood
- Troubleshooting API-related issues
- Extending the client with custom functionality
- Integrating RADb API into other tools

## RADb API Basics

### Base URL

```
https://api.radb.net
```

### Supported Databases

The client primarily works with the `RADB` source database. The API supports multiple IRR databases, but this client is optimized for RADb operations.

### API Documentation

Official RADb API documentation:
- Web Docs: https://api.radb.net/docs.html
- OpenAPI Spec: https://api.radb.net/RADB_API_OpenAPI.yaml

### API Version

The client uses the current stable API version. Version information is available through the API discovery endpoint.

## Authentication

### Authentication Methods

The RADb API supports multiple authentication methods:

#### 1. HTTP Basic Authentication (Recommended)

Used by default in the client:

```http
Authorization: Basic base64(username:api_key)
```

**How it works:**
1. Your RADb username (email) and API key are combined
2. The combination is base64 encoded
3. Sent in the Authorization header with each request

**Setting up:**
```bash
radb-client auth login
# Enter username: user@example.com
# Enter API key: your-api-key-here
```

#### 2. API Key Authentication

Alternative method using API key in header:

```http
X-API-Key: your-api-key
```

**Note:** The client uses Basic Auth by default as it's more widely supported.

### Getting Your API Key

1. Log into RADb web interface: https://www.radb.net
2. Navigate to: Account → Settings → API Access
3. Generate or view your API key
4. Copy the key (it looks like a long alphanumeric string)

**Important:**
- API keys are different from your web password
- Never share your API key
- Rotate keys periodically for security
- If compromised, regenerate immediately from web interface

### Credential Storage

The client stores credentials securely:

**Primary method:** System keyring
- macOS: Keychain
- Windows: Credential Manager
- Linux: Secret Service (GNOME Keyring, KWallet, etc.)

**Fallback method:** Encrypted file
- Location: `~/.radb-client/credentials.enc`
- Encrypted with strong encryption (NaCl secretbox)
- Used when keyring is unavailable

**Environment variables** (for CI/CD):
```bash
export RADB_USERNAME="user@example.com"
export RADB_API_KEY="your-api-key"
```

See [SECURITY.md](SECURITY.md) for detailed security information.

## API Endpoints

### Search Endpoint

Search the IRR database for objects:

```http
GET /{source}/search
```

**Parameters:**
- `source`: Database source (typically `RADB`)
- `q`: Search query string
- `filter`: Filter by object class (route, aut-num, etc.)

**Example:**
```bash
curl -u user@example.com:api_key \
  "https://api.radb.net/RADB/search?q=192.0.2.0/24"
```

**Client usage:**
```bash
radb-client search "192.0.2.0/24"
```

### Route Object Endpoints

#### List All Routes

```http
GET /{source}/route
```

**Example:**
```bash
curl -u user@example.com:api_key \
  "https://api.radb.net/RADB/route"
```

**Client usage:**
```bash
radb-client route list
```

#### Get Specific Route

```http
GET /{source}/route/{prefix}
```

**Path parameters:**
- `prefix`: IP prefix (URL-encoded if contains /)

**Example:**
```bash
curl -u user@example.com:api_key \
  "https://api.radb.net/RADB/route/192.0.2.0%2F24"
```

**Client usage:**
```bash
radb-client route show 192.0.2.0/24
```

#### Create Route

```http
POST /{source}/route
Content-Type: application/json
```

**Request body:**
```json
{
  "route": "192.0.2.0/24",
  "origin": "AS64500",
  "descr": "Description",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
```

**Required attributes:**
- `route`: IP prefix (IPv4 or IPv6)
- `origin`: AS number (format: ASN)
- `mnt-by`: Maintainer(s)
- `source`: Database source

**Example:**
```bash
curl -u user@example.com:api_key \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"route":"192.0.2.0/24","origin":"AS64500",...}' \
  "https://api.radb.net/RADB/route"
```

**Client usage:**
```bash
radb-client route create route.json
```

#### Update Route

```http
PUT /{source}/route/{prefix}
Content-Type: application/json
```

**Example:**
```bash
curl -u user@example.com:api_key \
  -X PUT \
  -H "Content-Type: application/json" \
  -d '{"route":"192.0.2.0/24","origin":"AS64500",...}' \
  "https://api.radb.net/RADB/route/192.0.2.0%2F24"
```

**Client usage:**
```bash
radb-client route update 192.0.2.0/24 updated.json
```

#### Delete Route

```http
DELETE /{source}/route/{prefix}
```

**Example:**
```bash
curl -u user@example.com:api_key \
  -X DELETE \
  "https://api.radb.net/RADB/route/192.0.2.0%2F24"
```

**Client usage:**
```bash
radb-client route delete 192.0.2.0/24
```

### Route6 Endpoints

IPv6 routes use similar endpoints with `route6` instead of `route`:

```http
GET    /{source}/route6
GET    /{source}/route6/{prefix}
POST   /{source}/route6
PUT    /{source}/route6/{prefix}
DELETE /{source}/route6/{prefix}
```

**Example IPv6 route:**
```json
{
  "route": "2001:db8::/32",
  "origin": "AS64500",
  "descr": "IPv6 route",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
```

### Contact Endpoints

Manage account contacts:

```http
GET    /{source}/contact              # List contacts
GET    /{source}/contact/{id}         # Get contact
POST   /{source}/contact              # Create contact
PUT    /{source}/contact/{id}         # Update contact
DELETE /{source}/contact/{id}         # Delete contact
```

**Contact object example:**
```json
{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "phone": "+1-555-0123",
  "role": "technical",
  "organization": "Example Networks"
}
```

### Validation Endpoints

#### Validate ASN

```http
GET /{source}/validate/asn/{asn}
```

**Example:**
```bash
curl -u user@example.com:api_key \
  "https://api.radb.net/RADB/validate/asn/AS64500"
```

**Client usage:**
```bash
radb-client validate asn AS64500
```

**Response:**
```json
{
  "asn": "AS64500",
  "valid": true,
  "aut_num": "AS64500",
  "as_name": "EXAMPLE-AS"
}
```

## Data Formats

### Supported Formats

The API supports multiple data formats:

#### JSON (Recommended)

```http
Accept: application/json
Content-Type: application/json
```

**Advantages:**
- Structured data
- Easy to parse
- Better for automation
- Used by default in client

#### Text (RPSL Format)

```http
Accept: text/plain
Content-Type: text/plain
```

**RPSL (Routing Policy Specification Language):**
```
route:        192.0.2.0/24
origin:       AS64500
descr:        Example route
mnt-by:       MAINT-EXAMPLE
source:       RADB
```

**When to use:**
- Compatibility with legacy tools
- Human-readable format
- Direct IRR database operations

### JSON Schema

#### Route Object

```json
{
  "route": "string (IP prefix)",
  "origin": "string (ASN format)",
  "descr": "string (optional)",
  "remarks": "string (optional)",
  "mnt-by": ["string (maintainer)"],
  "source": "string (database)",
  "created": "datetime (read-only)",
  "last-modified": "datetime (read-only)"
}
```

#### Contact Object

```json
{
  "id": "string (read-only)",
  "name": "string",
  "email": "string (required)",
  "phone": "string (optional)",
  "role": "string (admin|tech|billing)",
  "organization": "string (optional)",
  "created": "datetime (read-only)",
  "last-modified": "datetime (read-only)"
}
```

### Attribute Requirements

#### Route Objects

**Required:**
- `route`: Valid IP prefix (CIDR notation)
- `origin`: Valid AS number (ASN format)
- `mnt-by`: At least one valid maintainer
- `source`: Must be "RADB"

**Optional but recommended:**
- `descr`: Description of the route
- `remarks`: Additional notes
- `admin-c`: Administrative contact
- `tech-c`: Technical contact

**IPv4 Example:**
```json
{
  "route": "192.0.2.0/24",
  "origin": "AS64500",
  "descr": "Customer route",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
```

**IPv6 Example:**
```json
{
  "route": "2001:db8::/32",
  "origin": "AS64500",
  "descr": "IPv6 allocation",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
```

## Request and Response Examples

### Example 1: List Routes with JSON

**Request:**
```http
GET /RADB/route HTTP/1.1
Host: api.radb.net
Authorization: Basic dXNlckBleGFtcGxlLmNvbTpzZWNyZXQ=
Accept: application/json
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "routes": [
    {
      "route": "192.0.2.0/24",
      "origin": "AS64500",
      "descr": "Example route 1",
      "mnt-by": ["MAINT-EXAMPLE"],
      "source": "RADB",
      "created": "2025-10-15T10:30:00Z",
      "last-modified": "2025-10-28T14:22:00Z"
    },
    {
      "route": "198.51.100.0/24",
      "origin": "AS64501",
      "descr": "Example route 2",
      "mnt-by": ["MAINT-EXAMPLE"],
      "source": "RADB",
      "created": "2025-10-20T09:15:00Z",
      "last-modified": "2025-10-20T09:15:00Z"
    }
  ],
  "total": 2
}
```

### Example 2: Create Route

**Request:**
```http
POST /RADB/route HTTP/1.1
Host: api.radb.net
Authorization: Basic dXNlckBleGFtcGxlLmNvbTpzZWNyZXQ=
Content-Type: application/json

{
  "route": "203.0.113.0/24",
  "origin": "AS64502",
  "descr": "New customer route",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB"
}
```

**Success Response:**
```http
HTTP/1.1 201 Created
Content-Type: application/json
Location: /RADB/route/203.0.113.0%2F24

{
  "route": "203.0.113.0/24",
  "origin": "AS64502",
  "descr": "New customer route",
  "mnt-by": ["MAINT-EXAMPLE"],
  "source": "RADB",
  "created": "2025-10-29T15:30:00Z",
  "last-modified": "2025-10-29T15:30:00Z"
}
```

### Example 3: Search

**Request:**
```http
GET /RADB/search?q=AS64500&filter=route HTTP/1.1
Host: api.radb.net
Authorization: Basic dXNlckBleGFtcGxlLmNvbTpzZWNyZXQ=
Accept: application/json
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "results": [
    {
      "type": "route",
      "route": "192.0.2.0/24",
      "origin": "AS64500"
    },
    {
      "type": "route",
      "route": "2001:db8::/32",
      "origin": "AS64500"
    }
  ],
  "total": 2
}
```

## Error Handling

### HTTP Status Codes

The API uses standard HTTP status codes:

**Success codes:**
- `200 OK`: Request succeeded
- `201 Created`: Resource created successfully
- `204 No Content`: Delete succeeded

**Client error codes:**
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication failed
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource doesn't exist
- `409 Conflict`: Resource already exists
- `422 Unprocessable Entity`: Validation failed
- `429 Too Many Requests`: Rate limit exceeded

**Server error codes:**
- `500 Internal Server Error`: Server error
- `502 Bad Gateway`: Gateway error
- `503 Service Unavailable`: Service temporarily down
- `504 Gateway Timeout`: Request timeout

### Error Response Format

**JSON error response:**
```json
{
  "error": {
    "code": "validation_error",
    "message": "Invalid route prefix format",
    "details": {
      "field": "route",
      "value": "invalid",
      "expected": "Valid CIDR notation (e.g., 192.0.2.0/24)"
    }
  }
}
```

### Common Error Scenarios

#### Authentication Errors

**401 Unauthorized:**
```json
{
  "error": {
    "code": "authentication_failed",
    "message": "Invalid username or API key"
  }
}
```

**How client handles:**
- Checks credential storage
- Prompts for re-authentication
- Provides clear error message

#### Validation Errors

**422 Unprocessable Entity:**
```json
{
  "error": {
    "code": "validation_error",
    "message": "Route validation failed",
    "details": {
      "route": ["Invalid CIDR notation"],
      "origin": ["Must be valid AS number"],
      "mnt-by": ["At least one maintainer required"]
    }
  }
}
```

**How client handles:**
- Displays validation errors clearly
- Shows expected format
- Suggests corrections

#### Rate Limiting

**429 Too Many Requests:**
```json
{
  "error": {
    "code": "rate_limit_exceeded",
    "message": "Too many requests",
    "retry_after": 60
  }
}
```

**Headers:**
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1698594000
Retry-After: 60
```

**How client handles:**
- Respects rate limits
- Implements exponential backoff
- Queues requests if needed
- Shows retry timing to user

## Rate Limiting

### Limits

The API implements rate limiting to ensure fair usage:

**Default limits:**
- 100 requests per minute per user
- 1000 requests per hour per user
- May vary based on operation type

**Rate limit headers:**
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1698594000
```

### Client Behavior

The client automatically handles rate limiting:

1. **Monitoring:** Tracks rate limit headers
2. **Prevention:** Slows down before hitting limits
3. **Recovery:** Waits appropriate time if limit hit
4. **Queuing:** Queues requests if necessary

**User feedback:**
```
Rate limit approaching (5 requests remaining)
Pausing for 30 seconds before continuing...
```

### Best Practices

**Optimize API usage:**

1. **Batch operations:**
   ```bash
   # Instead of multiple calls
   # for route in routes; do radb-client route show $route; done

   # Use single list call
   radb-client route list --format json > all-routes.json
   ```

2. **Use caching:**
   - Client automatically caches list results
   - Use snapshots instead of repeated API calls
   - Leverage local state for diffs

3. **Respect rate limits:**
   - Don't parallel requests excessively
   - Add delays in scripts
   - Use bulk operations where available

## Best Practices

### Efficient API Usage

#### 1. Minimize Requests

**Bad:**
```bash
# Makes 100 API calls
for i in {1..100}; do
  radb-client route show $route_$i
done
```

**Good:**
```bash
# Makes 1 API call
radb-client route list --format json | \
  jq '.[] | select(...)'
```

#### 2. Use Appropriate Timeouts

```yaml
# config.yaml
api:
  timeout: 30  # Reasonable default
```

For bulk operations:
```bash
radb-client --timeout 60 route list
```

#### 3. Handle Errors Gracefully

```bash
# Retry on failure
for i in {1..3}; do
  radb-client route create new.json && break
  sleep 5
done
```

### Security Best Practices

1. **Never log credentials:**
   ```bash
   # Don't do this
   echo "API Key: $RADB_API_KEY"

   # Client never logs credentials
   radb-client --verbose route list  # Safe
   ```

2. **Use HTTPS only:**
   - Client enforces HTTPS
   - No HTTP fallback
   - Certificate validation enabled

3. **Rotate credentials regularly:**
   ```bash
   # Every 90 days
   radb-client auth login  # Enter new credentials
   ```

4. **Separate credentials for automation:**
   - Different API keys for different purposes
   - Easier to revoke if compromised
   - Better audit trail

### Integration Patterns

#### CI/CD Integration

```yaml
# .github/workflows/deploy-routes.yml
name: Deploy Routes

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Install radb-client
        run: |
          curl -L https://github.com/.../radb-client -o radb-client
          chmod +x radb-client

      - name: Deploy routes
        env:
          RADB_USERNAME: ${{ secrets.RADB_USERNAME }}
          RADB_API_KEY: ${{ secrets.RADB_API_KEY }}
        run: |
          ./radb-client route create routes.json
```

#### Monitoring Integration

```bash
#!/bin/bash
# monitor-radb.sh - Check for changes and alert

radb-client route diff --format json > /tmp/changes.json

if [ -s /tmp/changes.json ]; then
  # Send to monitoring system
  curl -X POST https://monitoring.example.com/alert \
    -H "Content-Type: application/json" \
    -d @/tmp/changes.json
fi
```

#### Backup Integration

```bash
#!/bin/bash
# backup to S3
DATE=$(date +%Y%m%d)
radb-client route list --format json | \
  gzip | \
  aws s3 cp - s3://backups/radb/routes-$DATE.json.gz
```

### Performance Optimization

1. **Connection reuse:**
   - Client maintains connection pool
   - Reduces overhead for multiple requests

2. **Compression:**
   - Enabled automatically for large responses
   - Reduces bandwidth usage

3. **Caching:**
   - Local state cache reduces API calls
   - Configurable cache TTL

4. **Concurrent requests:**
   - Safe operations can run in parallel
   - Respects rate limits

### Troubleshooting API Issues

#### Enable Debug Logging

```bash
radb-client --log-level DEBUG route list
```

**Shows:**
- HTTP requests and responses
- Headers (except credentials)
- Response times
- Error details

#### Test API Connectivity

```bash
# Test authentication
radb-client auth status

# Test basic API call
radb-client route list --verbose
```

#### Check API Status

```bash
# Via web
curl https://api.radb.net/health

# Via client
radb-client config show  # Shows base URL
```

For more troubleshooting, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

## Additional Resources

- [RADb API Documentation](https://api.radb.net/docs.html)
- [OpenAPI Specification](https://api.radb.net/RADB_API_OpenAPI.yaml)
- [RPSL Specification](https://tools.ietf.org/html/rfc2622)
- [IRR Routing Registry](https://www.irr.net/)

## Support

For API-specific issues:
- RADb Support: support@radb.net
- API Status: https://status.radb.net
- Documentation Issues: GitHub Issues

For client-specific issues:
- GitHub Issues
- Documentation: docs/ directory
