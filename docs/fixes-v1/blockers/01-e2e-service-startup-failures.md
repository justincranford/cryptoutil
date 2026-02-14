# Blocker: E2E Service Startup Failures

**Created**: 2026-02-14
**Status**: BLOCKING Phase 1 (E2E Test Execution)
**Severity**: P0 - Blocks all E2E testing

## Summary

Docker Compose services fail to start due to multiple issues in service implementation:

1. **KMS services** (cryptoutil-sqlite, cryptoutil-postgres-1, cryptoutil-postgres-2): Migration filesystem configuration error
2. **CA service** (ca-e2e): Cobra command routing panic
3. **JOSE service** (jose-e2e): Subcommand parsing error

## Evidence

See: `test-output/e2e/kms-workflow-retry.log` (test execution with fixed CLI syntax)

### KMS Error

```
2026/02/14 02:46:25 failed to create KMS server: failed to build KMS server: invalid migration config: migration FS is required for this mode
```

**Container**: compose-cryptoutil-sqlite-1  
**Root Cause**: Server builder migration configuration incomplete

### CA Error

```
panic: runtime error: flag redefined: config
```

**Container**: compose-ca-e2e-1
**Root Cause**: Cobra command flag collision in command tree

### JOSE Error  

```
2026/02/14 02:46:25 Unknown subcommand: --config=/app/config/jose-sqlite.yml
```

**Container**: compose-jose-e2e-1
**Root Cause**: Args not passing correctly through product -> service -> subcommand routing layers

## Impact

- **Phase 1**: E2E tests cannot execute (100% blocked)
- **Phase 3-5**: Dependent on Phase 1 completion
- **Overall Plan**: 0% of planned work executable until fixed

## Next Steps

1. **KMS**: Fix migration FS configuration in server builder
2. **CA**: Resolve cobra flag redefinition in command tree
3. **JOSE**: Fix args routing through jose.go -> ja.go chain

## Workaround

None available - services must start for E2E tests to run.
