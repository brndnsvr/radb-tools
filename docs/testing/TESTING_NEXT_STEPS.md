# Testing Next Steps - API Query Validation

Created: 2025-10-30
Status: Ready for testing

---

## Current Status

✅ **Authentication System Working**
- Config initialization: Working
- Credential storage: Working (encrypted)
- Credential retrieval: Working
- Auto-load on command run: Working

⚠️ **API Queries Need Testing**
- Code is ready
- Credentials need to be re-entered correctly
- Test queries prepared below

---

## Step 1: Re-authenticate

The password might have been entered incorrectly initially (shell escaping or typo).
Please re-login with correct credentials:

```bash
# Clear old credentials
radb-client auth logout

# Login with fresh credentials
radb-client auth login
```

**When prompted:**
- Username: `brandon.seaver@centersquaredc.com`
- Password: `a3MF!6Guj2!U8cDqcW` **(COPY-PASTE this exactly!)**

**Verify:**
```bash
radb-client auth status
```

**Expected output:**
```
Username: brandon.seaver@centersquaredc.com
Status: Authenticated (credentials stored)
```

---

## Step 2: Test Search for Well-Known Prefix

Search for Google's public DNS prefix:

```bash
radb-client search query "8.8.8.0/24" --type route
```

**What to look for:**
- ✅ Should return route object(s) for 8.8.8.0/24
- ✅ Should show origin AS15169 (Google)
- ❌ If 401: Password wasn't pasted correctly, retry Step 1
- ❌ If 404: Route might not be in RADb (try another)
- ❌ If 400: API parameter issue (log it)

**Alternative test (Cloudflare DNS):**
```bash
radb-client search query "1.1.1.0/24" --type route
```

---

## Step 3: Search for AS32298 Routes

List all routes associated with AS32298:

```bash
radb-client route list --origin AS32298
```

**What to look for:**
- ✅ Should return list of prefixes for AS32298
- ✅ Should show in table format
- ❌ If empty: AS might not have routes in RADb
- ❌ If 401: Authentication failed

**Alternative methods:**
```bash
# Via search
radb-client search query "AS32298" --type route

# Via route show (if you know a specific prefix)
radb-client route show <prefix> <asn>
```

---

## Step 4: Test Different Output Formats

```bash
# JSON output
radb-client route list --origin AS32298 --output json

# YAML output
radb-client route list --origin AS32298 --output yaml

# Table output (default)
radb-client route list --origin AS32298 --output table
```

---

## Step 5: Validate ASN

```bash
radb-client search validate-asn AS32298
```

**Expected:**
- Should confirm if AS32298 exists in RADb
- Returns true/false

---

## Expected Issues and Solutions

### Issue: 401 Unauthorized

**Symptom:**
```
Error: search failed: search failed with status 401: {"errors":[{"message":"Bad Password"...
```

**Solution:**
1. Run `radb-client auth logout`
2. Run `radb-client auth login`
3. **Copy-paste** the password (don't type it)
4. Try query again

**Why:** The password stored might have typos or shell escaping (like `\!` instead of `!`)

### Issue: 400 Bad Request - Source

**Symptom:**
```
Error: ... status 400: {"errors":[{"message":"Not in enum list: radb.",...
```

**Solution:**
- This is an API parameter issue
- Check that config has `source: RADB` (uppercase)
- Run: `radb-client config show | grep -i source`
- If lowercase, edit `~/.radb-client/config.yaml` and set `source: RADB`

### Issue: Empty Results

**Symptom:**
```
No routes found
```

**Possible causes:**
1. AS32298 might not have routes registered in RADb
2. Your RADb account might not have access to those routes
3. The routes might be in a different IRR database (not RADb)

**Solution:**
- Try a well-known AS like AS15169 (Google): `radb-client route list --origin AS15169`
- Try searching instead: `radb-client search query "AS32298"`

### Issue: 404 Not Found

**Symptom:**
```
Error: ... status 404
```

**Possible causes:**
- Route doesn't exist in RADb
- API endpoint incorrect

**Solution:**
- Try a different route/ASN
- Check logs for the actual API URL being called

---

## Debugging Commands

If things aren't working, run these to help diagnose:

```bash
# Check config
radb-client config show

# Check auth status
radb-client auth status

# Enable debug logging
radb-client --debug route list --origin AS32298

# Check version
radb-client version
```

---

## What to Report Back

Please capture and share:

1. **Authentication result:**
   ```bash
   radb-client auth status
   ```

2. **First search result:**
   ```bash
   radb-client search query "8.8.8.0/24" --type route
   ```

3. **AS32298 query result:**
   ```bash
   radb-client route list --origin AS32298
   ```

4. **Any errors:** Full error messages (especially HTTP status codes)

5. **Success case:** Sample output if queries work!

---

## Success Criteria

✅ Authentication works without prompting for password on each command
✅ Search queries return results (or clear "not found" messages)
✅ Route list shows data for known ASNs
✅ Output formats (table/json/yaml) all work
✅ Error messages are clear and actionable

---

## Known Limitations

1. **Daemon mode:** Placeholder only (needs API client completion)
2. **Some commands:** May be stubs (route create/update/delete need testing)
3. **API coverage:** Search and route list are implemented, others may need work

---

## Next Steps After Testing

Once queries work:

1. Test route creation/modification
2. Test contact management
3. Test snapshot functionality
4. Test diff operations
5. Work toward v1.0.0 release

---

## Quick Reference

**Version:** v0.0.42 (The Answer to Life, Auth, and Everything!)

**Key Files:**
- Config: `~/.radb-client/config.yaml`
- Credentials: `~/.radb-client/credentials.enc` (encrypted)
- Cache: `~/.radb-client/cache/`
- History: `~/.radb-client/history/`

**Support:**
- Issues: https://github.com/brndnsvr/radb-tools/issues
- Docs: See `QUICKSTART.md`, `INSTALL.md`, `TESTING_RUNBOOK.md`

---

**Ready to test!** 🚀

Run the commands above and let me know what you find.
