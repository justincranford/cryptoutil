# Docker Compose File Analysis - 2026-01-27

## Current State

**Total Compose Files**: 13

**By Service**:
- Identity: 4 files (compose.yml, compose.advanced.yml, compose.simple.yml, compose.e2e.yml)
- CA: 3 files (compose.yml, compose.simple.yml, compose/compose.yml)
- KMS: 2 files (compose.yml, compose.demo.yml)
- JOSE: 1 file (compose.yml)
- Telemetry: 1 file (compose.yml)
- Project-level: 2 files (compose/compose.yml, compose.integration.yml)

## Analysis by Service

### Identity Service (4 files)

**deployments/identity/compose.yml**: Full stack
**deployments/identity/compose.advanced.yml**: Extended configuration
**deployments/identity/compose.simple.yml**: Minimal configuration
**deployments/identity/compose.e2e.yml**: E2E testing

**Issue**: 4 different files for different use cases instead of one configurable file

### CA Service (3 files)

**deployments/ca/compose.yml**: Standard configuration
**deployments/ca/compose.simple.yml**: Minimal configuration
**deployments/ca/compose/compose.yml**: Nested subdirectory (duplicate?)

**Issue**: Multiple files for simple vs standard, plus duplicate in subdirectory

### KMS Service (2 files)

**deployments/kms/compose.yml**: Standard configuration
**deployments/kms/compose.demo.yml**: Demo configuration

**Issue**: Separate demo file instead of environment variable configuration

### JOSE Service (1 file) ✅

**deployments/jose/compose.yml**: Single file

**Note**: This is the CORRECT pattern - one file per service

### Telemetry (1 file) ✅

**deployments/telemetry/compose.yml**: Single file

**Note**: Shared infrastructure - one file is correct

## Root Cause Analysis

### Why Multiple Files Exist

1. **Different Environment Configurations**: Simple vs Advanced vs Demo
   - **Problem**: Should use environment variables or profiles
   - **Solution**: Single compose.yml with `.env` files or profiles

2. **E2E Testing Variations**:
   - **Problem**: E2E compose files with test-specific settings
   - **Solution**: Same compose.yml with different `.env` or override files

3. **Historical Duplication**:
   - **Problem**: deployments/ca/compose/compose.yml vs deployments/ca/compose.yml
   - **Solution**: Delete duplicate, keep only one

4. **Feature Parity Gaps**:
   - **Problem**: Not all files support both production and E2E
   - **Solution**: Design compose files for both use cases

## Recommended Solution

### Strategy: One Compose File Per Service

**Pattern**:
```
deployments/<service>/compose.yml          # Single file for ALL use cases
deployments/<service>/.env.production      # Production configuration
deployments/<service>/.env.development     # Development configuration
deployments/<service>/.env.e2e             # E2E test configuration
deployments/<service>/.env.demo            # Demo configuration
```

**Usage**:
```bash
# Production
docker compose --env-file deployments/identity/.env.production up

# E2E testing
docker compose --env-file deployments/identity/.env.e2e up

# Demo
docker compose --env-file deployments/identity/.env.demo up
```

### Compose File Requirements

**ALL compose.yml files MUST**:
1. Use environment variables for configuration (no hardcoded values)
2. Support both production and E2E testing
3. Include health checks
4. Use secrets for sensitive data (not environment variables)
5. Support multiple database backends (PostgreSQL, SQLite)

**Example Structure**:
```yaml
services:
  identity-authz:
    image: cryptoutil-identity-authz:${VERSION:-latest}
    environment:
      - BIND_PUBLIC_PORT=${AUTHZ_PUBLIC_PORT:-8180}
      - BIND_PRIVATE_PORT=${AUTHZ_PRIVATE_PORT:-9090}
      - DATABASE_TYPE=${DATABASE_TYPE:-postgresql}
      - DATABASE_URL=${DATABASE_URL}
    secrets:
      - unseal_key
      - tls_cert
      - tls_key
    healthcheck:
      test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
```

## Migration Plan

### Phase 1: Consolidate Identity

- [ ] Merge compose.yml + compose.advanced.yml + compose.simple.yml + compose.e2e.yml → single compose.yml
- [ ] Create .env.production, .env.development, .env.e2e
- [ ] Test production deployment
- [ ] Test E2E deployment
- [ ] Delete deprecated files
- [ ] Update documentation

### Phase 2: Consolidate CA

- [ ] Merge compose.yml + compose.simple.yml → single compose.yml
- [ ] Delete deployments/ca/compose/compose.yml (duplicate)
- [ ] Create .env files
- [ ] Test and verify
- [ ] Delete deprecated files

### Phase 3: Consolidate KMS

- [ ] Merge compose.yml + compose.demo.yml → single compose.yml
- [ ] Create .env.demo
- [ ] Test and verify
- [ ] Delete deprecated files

### Phase 4: Verify JOSE & Telemetry

- [ ] Verify deployments/jose/compose.yml supports all use cases
- [ ] Verify deployments/telemetry/compose.yml supports all use cases
- [ ] Add .env files if needed

### Phase 5: Project-Level Files

**Investigate**:
- deployments/compose/compose.yml - What is this for?
- deployments/compose.integration.yml - Integration testing?

**Decision Needed**: Keep or consolidate into per-service files?

## Expected Outcome

**Before**: 13 compose files
**After**: 5-7 compose files (one per service + telemetry + possibly integration)

**Benefits**:
1. Single source of truth per service
2. Easier maintenance (change once, applies everywhere)
3. Consistent patterns across all services
4. Environment-based configuration (not file-based)
5. Production and E2E use same infrastructure

## Success Criteria

- [ ] Identity: 1 compose.yml (from 4)
- [ ] CA: 1 compose.yml (from 3)
- [ ] KMS: 1 compose.yml (from 2)
- [ ] JOSE: 1 compose.yml ✅ (already correct)
- [ ] Telemetry: 1 compose.yml ✅ (already correct)
- [ ] All compose files support production + E2E
- [ ] All compose files use environment variables
- [ ] All compose files use secrets (not env vars for sensitive data)
- [ ] Documentation updated (DEV-SETUP.md, README.md)
- [ ] Deprecated files deleted

## Next Steps

1. Create detailed consolidation plan per service
2. Test current deployments before migration
3. Implement Phase 1 (Identity)
4. Implement Phases 2-4 incrementally
5. Update documentation
6. Delete deprecated files
7. Verify E2E workflows still pass
