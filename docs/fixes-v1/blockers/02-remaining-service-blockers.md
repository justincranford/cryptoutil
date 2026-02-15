# E2E Service Blockers - Post Initial Fixes

**Created**: 2026-02-14  
**Status**: ACTIVE  
**Priority**: P0 (blocks all E2E tests)

## Summary

After fixing initial CLI syntax and migration issues, three new blocking errors discovered:

1. **KMS Services**: Session JWK algorithm configuration missing
2. **JOSE Service**: Args routing still incorrect (service name not stripped)  
3. **CA Service**: Unknown flag --config error

## Blocker Details

### 1. KMS Services (cryptoutil-sqlite, cryptoutil-postgres-1/2)

**Error**:
```
failed to initialize service session JWK: unsupported JWS algorithm:
```

**Root Cause**: Session manager trying to initialize with empty algorithm string - configuration missing.

**Status**: NEW - Configuration issue discovered after migration fix succeeded

---

### 2. JOSE Service (jose-e2e)

**Error**:
```
Failed to parse configuration: failed to parse template settings: invalid subcommand: use "start", "stop", "init", "live", or "ready"
```

**Root Cause**: Args indexing fix was incorrect - `ja.Ja()` receives `["ja", "start"]` but treating `args[0]` as subcommand when it's actually the service name.

**Correct Pattern** (from cipher-im):
- Receive args WITH service name: `["im", "server"]`
- Check if args[0] matches service name OR is subcommand
- Pass args[1:] to subcommand handler
- Handler prepends required subcommand for template Parse()

**Fix Required**: Revert args indexing change, align with cipher-im pattern

---

### 3. CA Service (ca-e2e)

**Error**:  
```
Error: unknown flag: --config
```

**Exit Code**: 0 (clean shutdown)

**Root Cause**: Unclear - may be docker-compose entrypoint issue vs actual flag problem

**Status**: Low priority - exited cleanly, may be test environment issue

---

## Evidence

**KMS logs**: `docker logs cb3c4f0c8b5f`  
**JOSE logs**: `docker logs 1a50cb7ce092`
**CA logs**: `docker logs d95d3a8e429c`

## Next Steps

1. Fix JOSE args routing to match cipher-im pattern  
2. Investigate KMS session JWK algorithm configuration
3. Validate CA flag issue is test-only or real bug
