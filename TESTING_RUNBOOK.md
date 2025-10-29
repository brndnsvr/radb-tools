# RADb Client v0.9 - Manual Testing Runbook

**Version**: 0.9.0-pre
**Status**: Pre-release - Requires manual testing before v1.0
**Date**: 2025-10-29

---

## Overview

This runbook guides you through comprehensive manual testing of the RADb API client before promoting from v0.9 to v1.0 production release.

**Estimated Time**: 2-3 hours for complete testing
**Prerequisites**:
- RADb API account with credentials
- Test route objects and contacts to work with
- Multiple terminal windows helpful

---

## Pre-Testing Setup

### 1. Build the Binary

```bash
cd /home/bss/code/radb

# Clean build
rm -rf dist/
go build -o dist/radb-client ./cmd/radb-client

# Verify build
ls -lh dist/radb-client
./dist/radb-client --version
```

**Expected Output:**
```
radb-client version 0.9.0-pre
Pre-release build - pending final testing
```

**✅ Pass Criteria**: Binary builds without errors, version shows 0.9.0-pre

---

### 2. Clean Testing Environment

```bash
# Backup existing config if you have one
mv ~/.radb-client ~/.radb-client.backup 2>/dev/null || true

# Start fresh
rm -rf ~/.radb-client

# Verify clean state
ls -la ~/.radb-client 2>&1
```

**Expected Output**: Directory should not exist or be empty

**✅ Pass Criteria**: Starting with clean slate

---

## Phase 1: Basic Functionality Testing

### Test 1.1: Help and Version Commands

```bash
# Test help
./dist/radb-client --help

# Test version
./dist/radb-client version

# Test command-specific help
./dist/radb-client config --help
./dist/radb-client auth --help
./dist/radb-client route --help
```

**✅ Pass Criteria**:
- [ ] Help output is clear and well-formatted
- [ ] All commands listed (auth, config, route, contact, snapshot, history, search, wizard, version)
- [ ] Command aliases visible (r, c, snap, hist, find)
- [ ] No errors or panics

**📝 Notes:**
```
[Your observations here]
```

---

### Test 1.2: Configuration Initialization

```bash
# Initialize configuration
./dist/radb-client config init

# Verify config file created
cat ~/.radb-client/config.yaml

# Show current configuration
./dist/radb-client config show
```

**✅ Pass Criteria**:
- [ ] Config file created at `~/.radb-client/config.yaml`
- [ ] Contains expected sections (api, preferences)
- [ ] Default values are sensible
- [ ] `config show` displays readable output

**📝 Notes:**
```
[Check if default values make sense]
```

---

### Test 1.3: Configuration Management

```bash
# Set a configuration value
./dist/radb-client config set preferences.log_level DEBUG

# Verify it was set
./dist/radb-client config show | grep log_level

# Check the file was updated
cat ~/.radb-client/config.yaml | grep log_level

# Try setting invalid value (should fail gracefully)
./dist/radb-client config set invalid.key value
```

**✅ Pass Criteria**:
- [ ] `config set` updates values correctly
- [ ] Changes persist in YAML file
- [ ] Invalid keys handled gracefully with clear error
- [ ] No crashes

**📝 Notes:**
```
[Note any issues with config management]
```

---

### Test 1.4: Interactive Wizard

```bash
# Run the interactive wizard
./dist/radb-client wizard

# Answer the prompts with test values
# Note: This is interactive, so document what you entered
```

**✅ Pass Criteria**:
- [ ] Wizard runs without errors
- [ ] Prompts are clear and helpful
- [ ] Generated config is valid
- [ ] Can proceed or cancel gracefully

**📝 Notes:**
```
[Document your interaction with the wizard]
Prompts shown:
Defaults offered:
Final config generated:
```

---

## Phase 2: Authentication Testing

### Test 2.1: Authentication Status (Before Login)

```bash
# Check auth status before logging in
./dist/radb-client auth status
```

**✅ Pass Criteria**:
- [ ] Indicates not authenticated
- [ ] Error message is clear and helpful
- [ ] Suggests running `auth login`

**📝 Notes:**
```
[Document the error message]
```

---

### Test 2.2: Authentication Login

**⚠️ IMPORTANT**: You'll need your actual RADb credentials for this test.

```bash
# Login interactively
./dist/radb-client auth login

# You'll be prompted for:
# - Username
# - API Key (password will be hidden)
```

**✅ Pass Criteria**:
- [ ] Prompts for username and API key
- [ ] Password input is hidden (shows ******)
- [ ] Success message on valid credentials
- [ ] Clear error on invalid credentials (test with wrong password if safe)

**🔐 Security Check**:
- [ ] Password not visible when typing
- [ ] Credentials not logged to console
- [ ] No credentials in debug output

**📝 Notes:**
```
[Document the login experience]
System keyring used: [Yes/No, which one?]
Fallback to encrypted file: [Yes/No]
Any warnings or errors:
```

---

### Test 2.3: Authentication Status (After Login)

```bash
# Check auth status after login
./dist/radb-client auth status

# Verify credentials stored
# On macOS: Check Keychain Access app
# On Linux: Check with secret-tool if available
# On Windows: Check Credential Manager
```

**✅ Pass Criteria**:
- [ ] Shows authenticated status
- [ ] Displays username
- [ ] Shows credential storage method (keyring or encrypted file)
- [ ] No sensitive data exposed in output

**📝 Notes:**
```
[Document authentication storage]
Storage method: [keyring/encrypted file]
Keyring type: [macOS Keychain/Linux Secret Service/Windows Credential Manager]
```

---

### Test 2.4: Credential Storage Verification

```bash
# Check where credentials are stored
ls -la ~/.radb-client/

# If using encrypted file fallback, verify:
ls -la ~/.radb-client/credentials.enc 2>/dev/null
```

**✅ Pass Criteria**:
- [ ] Credentials stored securely (keyring preferred)
- [ ] If encrypted file: proper permissions (0600)
- [ ] No plaintext credentials anywhere

**📝 Notes:**
```
[Document credential storage location and permissions]
```

---

## Phase 3: Route Operations Testing

**⚠️ CRITICAL**: These tests interact with the live RADb API. Use test data only!

### Test 3.1: List Routes (Initial)

```bash
# List all your routes
./dist/radb-client route list

# Check with JSON output
./dist/radb-client route list --format json

# Check with YAML output
./dist/radb-client route list --format yaml
```

**✅ Pass Criteria**:
- [ ] Lists routes without errors
- [ ] Table output is readable and formatted
- [ ] JSON output is valid JSON
- [ ] YAML output is valid YAML
- [ ] Snapshot automatically created

**Snapshot Verification**:
```bash
# Check that snapshot was created
ls -la ~/.radb-client/cache/
ls -la ~/.radb-client/history/

# Verify snapshot content
cat ~/.radb-client/cache/route_objects.json | head -20
```

**✅ Pass Criteria**:
- [ ] Snapshot exists in cache/
- [ ] Historical snapshot exists in history/ with timestamp
- [ ] Snapshot contains valid JSON data
- [ ] Checksum field present in snapshot

**📝 Notes:**
```
[Document your routes and snapshot]
Number of routes listed:
Snapshot timestamp:
Snapshot size:
Any errors:
```

---

### Test 3.2: Show Specific Route

```bash
# Pick a route from your list and show details
# Replace with actual prefix from your account
./dist/radb-client route show 192.0.2.0/24

# Try with a non-existent route (should error gracefully)
./dist/radb-client route show 198.51.100.99/32
```

**✅ Pass Criteria**:
- [ ] Shows route details correctly
- [ ] Displays all RPSL fields
- [ ] Handles non-existent routes gracefully
- [ ] Error messages are helpful

**📝 Notes:**
```
[Document route details shown]
Fields displayed:
Format quality:
```

---

### Test 3.3: Route Diff (No Changes)

```bash
# Run list again to create second snapshot
./dist/radb-client route list

# Show diff (should show no changes)
./dist/radb-client route diff
```

**✅ Pass Criteria**:
- [ ] Reports no changes (if you haven't modified routes)
- [ ] Output format is clear
- [ ] Timestamps shown for comparison

**📝 Notes:**
```
[Document diff output]
```

---

### Test 3.4: Route Creation (If Safe)

**⚠️ ONLY IF YOU HAVE PERMISSION TO CREATE TEST ROUTES**

```bash
# Create a test route file
cat > /tmp/test-route.json <<EOF
{
  "route": "203.0.113.0/24",
  "origin": "AS64500",
  "descr": "Test route for radb-client v0.9 testing",
  "mnt-by": ["YOUR-MAINTAINER"],
  "source": "RADB"
}
EOF

# Create the route
./dist/radb-client route create /tmp/test-route.json

# Verify it was created
./dist/radb-client route show 203.0.113.0/24

# List routes and check diff
./dist/radb-client route list
./dist/radb-client route diff
```

**✅ Pass Criteria**:
- [ ] Route created successfully
- [ ] Show command displays new route
- [ ] Diff shows the route as "added"
- [ ] Proper validation of input

**🧹 Cleanup**:
```bash
# Delete test route after testing
./dist/radb-client route delete 203.0.113.0/24
```

**📝 Notes:**
```
[Document route creation]
Created successfully: [Yes/No]
Validation worked: [Yes/No]
Errors encountered:
```

---

### Test 3.5: Input Validation

```bash
# Test with invalid AS number
echo '{"route":"192.0.2.0/24","origin":"INVALID"}' > /tmp/bad-asn.json
./dist/radb-client route create /tmp/bad-asn.json

# Test with invalid IP prefix
echo '{"route":"999.999.999.999/99","origin":"AS64500"}' > /tmp/bad-ip.json
./dist/radb-client route create /tmp/bad-ip.json

# Test with missing required fields
echo '{"route":"192.0.2.0/24"}' > /tmp/incomplete.json
./dist/radb-client route create /tmp/incomplete.json
```

**✅ Pass Criteria**:
- [ ] Invalid AS number rejected with clear error
- [ ] Invalid IP prefix rejected with clear error
- [ ] Missing fields rejected with clear error
- [ ] Error messages suggest how to fix
- [ ] No crashes or panics

**📝 Notes:**
```
[Document validation behavior]
Error messages quality: [Clear/Confusing]
Suggestions helpful: [Yes/No]
```

---

## Phase 4: Contact Operations Testing

### Test 4.1: List Contacts

```bash
# List all contacts
./dist/radb-client contact list

# JSON output
./dist/radb-client contact list --format json
```

**✅ Pass Criteria**:
- [ ] Lists contacts without errors
- [ ] Output is formatted and readable
- [ ] Snapshot created automatically

**📝 Notes:**
```
[Document contacts]
Number of contacts:
Format quality:
```

---

### Test 4.2: Show Specific Contact

```bash
# Show a specific contact (replace with real ID)
./dist/radb-client contact show CONTACT-ID
```

**✅ Pass Criteria**:
- [ ] Displays contact details
- [ ] All fields shown correctly

**📝 Notes:**
```
[Document contact display]
```

---

### Test 4.3: Contact Create/Update (If Safe)

**⚠️ ONLY IF YOU HAVE PERMISSION TO MODIFY CONTACTS**

```bash
# Create test contact file
cat > /tmp/test-contact.json <<EOF
{
  "name": "Test Contact",
  "email": "test@example.com",
  "role": "tech"
}
EOF

# Create (or test validation)
./dist/radb-client contact create /tmp/test-contact.json
```

**✅ Pass Criteria**:
- [ ] Creation works or fails gracefully
- [ ] Validation catches invalid emails
- [ ] Error messages are clear

**📝 Notes:**
```
[Document contact operations]
```

---

## Phase 5: Snapshot Management Testing

### Test 5.1: List Snapshots

```bash
# List all snapshots
./dist/radb-client snapshot list

# Should show snapshots from previous route/contact lists
```

**✅ Pass Criteria**:
- [ ] Shows all snapshots with timestamps
- [ ] Types shown (route_objects, contacts)
- [ ] File sizes shown
- [ ] Timestamps formatted correctly

**📝 Notes:**
```
[Document snapshots]
Number of snapshots:
Types present:
```

---

### Test 5.2: Create Manual Snapshot

```bash
# Create a manual snapshot
./dist/radb-client snapshot create

# Verify it appears in list
./dist/radb-client snapshot list
```

**✅ Pass Criteria**:
- [ ] Snapshot created successfully
- [ ] Appears in snapshot list
- [ ] Proper timestamp

**📝 Notes:**
```
[Document snapshot creation]
```

---

### Test 5.3: Show Snapshot Contents

```bash
# Show a specific snapshot (use timestamp from list)
./dist/radb-client snapshot show <timestamp>

# Try with JSON output
./dist/radb-client snapshot show <timestamp> --format json
```

**✅ Pass Criteria**:
- [ ] Displays snapshot contents
- [ ] Data is readable
- [ ] Checksum verification works

**📝 Notes:**
```
[Document snapshot viewing]
```

---

### Test 5.4: Delete Snapshot

```bash
# Delete a snapshot (choose a test one)
./dist/radb-client snapshot delete <timestamp>

# Verify it's gone
./dist/radb-client snapshot list
```

**✅ Pass Criteria**:
- [ ] Snapshot deleted successfully
- [ ] No longer appears in list
- [ ] Confirmation message clear

**📝 Notes:**
```
[Document snapshot deletion]
```

---

## Phase 6: History & Change Tracking Testing

### Test 6.1: View Change History

```bash
# View all changes
./dist/radb-client history show

# View changes in last 24 hours
./dist/radb-client history show --since 24h

# View only route changes
./dist/radb-client history show --type route
```

**✅ Pass Criteria**:
- [ ] Shows change history from your testing
- [ ] Time filtering works
- [ ] Type filtering works
- [ ] Output is readable

**📝 Notes:**
```
[Document history]
Changes shown:
Format quality:
```

---

### Test 6.2: Diff Between Snapshots

```bash
# List snapshots to get two timestamps
./dist/radb-client snapshot list

# Diff between two snapshots
./dist/radb-client history diff <timestamp1> <timestamp2>
```

**✅ Pass Criteria**:
- [ ] Shows differences between snapshots
- [ ] Added/removed/modified clearly indicated
- [ ] Color coding works (if terminal supports it)

**📝 Notes:**
```
[Document diff functionality]
```

---

## Phase 7: Search & Validation Testing

### Test 7.1: Search Routes

```bash
# Search for routes by prefix
./dist/radb-client search "192.0.2"

# Search with type filter
./dist/radb-client search --type route "AS64500"
```

**✅ Pass Criteria**:
- [ ] Search returns relevant results
- [ ] Type filtering works
- [ ] Output is formatted correctly

**📝 Notes:**
```
[Document search results]
```

---

### Test 7.2: ASN Validation

```bash
# Validate a valid ASN
./dist/radb-client validate asn AS64500

# Try an invalid ASN
./dist/radb-client validate asn AS999999999999
```

**✅ Pass Criteria**:
- [ ] Valid ASN reports correctly
- [ ] Invalid ASN reports error
- [ ] Validation is accurate

**📝 Notes:**
```
[Document ASN validation]
```

---

## Phase 8: Error Handling & Edge Cases

### Test 8.1: Network Errors

```bash
# Test with invalid API URL
./dist/radb-client config set api.base_url https://invalid.example.com
./dist/radb-client route list

# Reset to correct URL
./dist/radb-client config set api.base_url https://api.radb.net
```

**✅ Pass Criteria**:
- [ ] Network errors handled gracefully
- [ ] Clear error message about connectivity
- [ ] No crashes or panics

**📝 Notes:**
```
[Document error handling]
```

---

### Test 8.2: Authentication Errors

```bash
# Logout and try an operation
./dist/radb-client auth logout
./dist/radb-client route list
```

**✅ Pass Criteria**:
- [ ] Requires authentication
- [ ] Error message suggests logging in
- [ ] No credential leakage in errors

**📝 Notes:**
```
[Document auth error handling]
```

---

### Test 8.3: Rate Limiting

```bash
# Make rapid requests (if safe with your API quota)
for i in {1..70}; do
  ./dist/radb-client route list --format json > /dev/null 2>&1 &
done
wait
```

**✅ Pass Criteria**:
- [ ] Rate limiter prevents exceeding limits
- [ ] Requests queued appropriately
- [ ] No API errors from rate limit exceeded

**📝 Notes:**
```
[Document rate limiting behavior]
```

---

### Test 8.4: File Permission Errors

```bash
# Make config directory read-only
chmod 000 ~/.radb-client
./dist/radb-client config show

# Restore permissions
chmod 755 ~/.radb-client
```

**✅ Pass Criteria**:
- [ ] Permission errors handled gracefully
- [ ] Clear error message about permissions
- [ ] Suggests fix (check permissions)

**📝 Notes:**
```
[Document permission error handling]
```

---

### Test 8.5: Corrupted Snapshot

```bash
# Corrupt a snapshot file
echo "corrupted data" > ~/.radb-client/cache/route_objects.json

# Try to read it
./dist/radb-client route diff

# Clean up
rm ~/.radb-client/cache/route_objects.json
./dist/radb-client route list  # Recreate valid snapshot
```

**✅ Pass Criteria**:
- [ ] Detects corruption (checksum mismatch)
- [ ] Handles gracefully
- [ ] Suggests recreating snapshot

**📝 Notes:**
```
[Document corruption handling]
```

---

## Phase 9: Performance & Stress Testing

### Test 9.1: Large Result Sets

```bash
# If you have many routes, test performance
time ./dist/radb-client route list

# Test diff performance on large datasets
time ./dist/radb-client route diff
```

**✅ Pass Criteria**:
- [ ] Handles large datasets efficiently
- [ ] Memory usage reasonable
- [ ] No significant slowdown

**📝 Notes:**
```
[Document performance]
Number of routes:
List time:
Diff time:
Memory usage (if measurable):
```

---

### Test 9.2: Concurrent Operations

```bash
# Run multiple operations in parallel
./dist/radb-client route list &
./dist/radb-client contact list &
./dist/radb-client snapshot list &
wait
```

**✅ Pass Criteria**:
- [ ] File locking prevents corruption
- [ ] Operations complete successfully
- [ ] No race conditions

**📝 Notes:**
```
[Document concurrent access]
```

---

## Phase 10: Output & Formatting Testing

### Test 10.1: Output Formats

```bash
# Test all output formats for routes
./dist/radb-client route list --format table
./dist/radb-client route list --format json | jq .
./dist/radb-client route list --format yaml

# Test diff output with colors
./dist/radb-client route diff
```

**✅ Pass Criteria**:
- [ ] Table format is readable and aligned
- [ ] JSON is valid (parseable by jq)
- [ ] YAML is valid
- [ ] Colors work in diff output (if terminal supports)

**📝 Notes:**
```
[Document output quality]
Table formatting:
JSON validity:
YAML validity:
Color support:
```

---

### Test 10.2: Progress Indicators

```bash
# Operations that show progress (if bulk operations available)
# Note: This may not be testable without bulk data
```

**✅ Pass Criteria**:
- [ ] Progress bars appear for long operations
- [ ] Progress updates smoothly
- [ ] Completes at 100%

**📝 Notes:**
```
[Document progress indicators]
```

---

## Phase 11: Documentation Verification

### Test 11.1: Help Text Accuracy

```bash
# Verify help text matches actual behavior
./dist/radb-client route --help
./dist/radb-client config --help
./dist/radb-client auth --help

# Compare with docs/COMMANDS.md
```

**✅ Pass Criteria**:
- [ ] Help text is accurate
- [ ] Examples in help work
- [ ] Flags documented correctly
- [ ] Matches documentation

**📝 Notes:**
```
[Document any discrepancies]
```

---

### Test 11.2: Error Messages vs Documentation

```bash
# Verify error messages match troubleshooting guide
# Check docs/TROUBLESHOOTING.md against actual errors
```

**✅ Pass Criteria**:
- [ ] Error messages match troubleshooting guide
- [ ] Common errors documented
- [ ] Solutions work

**📝 Notes:**
```
[Document any missing error scenarios]
```

---

## Phase 12: Security Testing

### Test 12.1: Credential Security

```bash
# Verify credentials not exposed
./dist/radb-client --debug auth status 2>&1 | grep -i password
./dist/radb-client --debug route list 2>&1 | grep -i password

# Check log files if any
find ~/.radb-client -name "*.log" -exec grep -l password {} \;
```

**✅ Pass Criteria**:
- [ ] No credentials in debug output
- [ ] No credentials in log files
- [ ] No credentials in error messages

**📝 Notes:**
```
[Document any credential exposure]
```

---

### Test 12.2: File Permissions

```bash
# Check permissions on sensitive files
ls -la ~/.radb-client/
ls -la ~/.radb-client/config.yaml
ls -la ~/.radb-client/credentials.enc 2>/dev/null
```

**✅ Pass Criteria**:
- [ ] Config directory: 0700 or 0755
- [ ] Config file: 0600 or 0644
- [ ] Credentials file (if exists): 0600
- [ ] No world-readable sensitive files

**📝 Notes:**
```
[Document file permissions]
```

---

### Test 12.3: HTTPS Enforcement

```bash
# Try to force HTTP (should fail or warn)
./dist/radb-client config set api.base_url http://api.radb.net
./dist/radb-client route list
```

**✅ Pass Criteria**:
- [ ] Warns about or prevents HTTP
- [ ] Enforces HTTPS
- [ ] Clear security message

**📝 Notes:**
```
[Document HTTPS enforcement]
```

---

## Phase 13: Platform-Specific Testing

### Test 13.1: Platform Detection

```bash
# Verify platform-specific features work
# - macOS: Keychain Access
# - Linux: Secret Service / encrypted file
# - Windows: Credential Manager

# Document which credential storage is being used
./dist/radb-client auth status
```

**✅ Pass Criteria**:
- [ ] Detects platform correctly
- [ ] Uses appropriate credential storage
- [ ] Falls back to encrypted file if needed

**📝 Notes:**
```
[Document platform-specific behavior]
Platform: [macOS/Linux/Windows]
Credential storage: [Keychain/Secret Service/Credential Manager/Encrypted File]
```

---

## Phase 14: Cleanup & State Management

### Test 14.1: Snapshot Cleanup

```bash
# Create many snapshots
for i in {1..10}; do
  ./dist/radb-client snapshot create
  sleep 1
done

# List them
./dist/radb-client snapshot list

# Check if automatic cleanup would occur
# (Verify retention policy in config)
cat ~/.radb-client/config.yaml | grep -A5 preferences
```

**✅ Pass Criteria**:
- [ ] Snapshots created successfully
- [ ] Retention policy is configurable
- [ ] Old snapshots cleaned up (if policy triggers)

**📝 Notes:**
```
[Document snapshot management]
```

---

### Test 14.2: Complete Cleanup

```bash
# Test logout cleanup
./dist/radb-client auth logout

# Verify credentials removed
./dist/radb-client auth status
```

**✅ Pass Criteria**:
- [ ] Logout removes credentials
- [ ] Keyring/file cleaned up
- [ ] Config preserved

**📝 Notes:**
```
[Document cleanup behavior]
```

---

## Final Checklist

### Critical Issues (Blockers for v1.0)

- [ ] **Authentication**: Secure credential storage works
- [ ] **API Integration**: Can communicate with RADb API
- [ ] **Route Operations**: Basic CRUD operations work
- [ ] **Data Integrity**: Snapshots and checksums work correctly
- [ ] **Security**: No credential leakage, proper permissions
- [ ] **Error Handling**: Graceful failures, helpful messages

### Major Issues (Should fix before v1.0)

- [ ] **Performance**: Acceptable for typical use cases
- [ ] **Output Formatting**: Readable and professional
- [ ] **Documentation**: Help text accurate and useful
- [ ] **Validation**: Input validation works correctly
- [ ] **State Management**: Snapshots and diffs work reliably

### Minor Issues (Can defer to v1.1)

- [ ] **UX Polish**: Minor formatting improvements
- [ ] **Edge Cases**: Rare scenarios handled
- [ ] **Additional Features**: Nice-to-haves

---

## Testing Results Summary

### Environment
```
Date: __________________
Tester: __________________
Platform: __________________
OS Version: __________________
Go Version: __________________
```

### Overall Results
```
Total Tests Run: _____ / 50+
Tests Passed: _____
Tests Failed: _____
Tests Skipped: _____
```

### Critical Issues Found
```
1. [Issue description]
   Severity: [Critical/Major/Minor]
   Steps to reproduce:
   Expected:
   Actual:

2. [Issue description]
   ...
```

### Recommendations

**Ready for v1.0?**
- [ ] Yes - All critical tests passed
- [ ] No - Critical issues found (list above)
- [ ] Conditional - With minor fixes

**Next Steps:**
```
[Your recommendations]
```

---

## Post-Testing Actions

### If All Tests Pass

```bash
# Update version to 1.0.0
# Update PROJECT_SUMMARY.md status
# Update README.md status
# Tag v1.0.0
git tag -a v1.0.0 -m "Release v1.0.0 - Production Ready"
git push origin v1.0.0
```

### If Issues Found

```bash
# Document issues in GitHub Issues or tracking system
# Prioritize fixes
# Re-test after fixes
# Remain at v0.9.x until ready
```

---

## Additional Notes

```
[Add any additional observations, suggestions, or concerns here]
```

---

**End of Testing Runbook**

Save this file with your testing results for the project records.
