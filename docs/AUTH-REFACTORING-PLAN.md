# Auth → Authn/Authz Refactoring Plan

**Status**: PLAN PENDING USER APPROVAL
**Created**: 2025-12-24
**Scope**: Global search and replace of ambiguous `auth` references across entire project

## Executive Summary

**Problem**: The abbreviation `auth` is ambiguous - it can mean authentication (who you are), authorization (what you can do), or both combined.

**Solution**: Replace all `auth` references with specific abbreviations:
- `authn` = Authentication (identity verification)
- `authz` = Authorization (permission checking)
- `authnz` = Combined authentication AND authorization

**Impact**: This is a MASSIVE refactoring affecting:
- **200+ code references** (grep search returned 200 matches with more available)
- **89 files/directories** with `auth` in their names
- **Potential breaking changes** in API endpoints, configuration keys, database schema

## Phase 1 Completed (2025-12-24)

✅ **Terminology Standards**: Added to `01-01.terminology.instructions.md`
✅ **Reference Documentation**: Renamed `auth-factors.md` → `authn-authz-factors.md`, converted to compact table format
✅ **Tactical Documentation**: Renamed `02-10.authentication.instructions.md` → `02-10.authn.instructions.md`, removed duplicates
✅ **References Updated**: constitution.md, spec.md, copilot instructions (7 total replacements)
✅ **Committed & Pushed**: Commit ef9bd47a pushed to main branch

## Phase 2: Global Refactoring (PENDING USER APPROVAL)

### Step 1: File and Directory Renames

**Files to rename** (89 total, showing critical ones):

**API Files**:
```
api/identity/openapi-gen_config_authz.yaml → NO CHANGE (already specific: authz)
api/identity/openapi_spec_authz.yaml → NO CHANGE (already specific: authz)
```

**Internal Code Files** (requires context analysis):
```
internal/kms/server/middleware/service_auth.go → service_authn.go OR service_authz.go (analyze content)
internal/kms/server/middleware/service_auth_test.go → service_authn_test.go OR service_authz_test.go
internal/identity/server/authz_server.go → NO CHANGE (already specific: authz)
internal/identity/security/client_auth_policy.go → client_authn_policy.go OR client_authz_policy.go (analyze)
internal/identity/security/client_auth_policy_test.go → client_authn_policy_test.go OR client_authz_policy_test.go
internal/infra/realm/authenticator.go → NO CHANGE (clear context: authentication)
internal/infra/realm/authenticator_test.go → NO CHANGE (clear context: authentication)
```

**Repository Files** (requires context analysis):
```
internal/identity/repository/device_authorization_repository.go → NO CHANGE (OAuth spec term)
internal/identity/repository/pushed_authorization_request_repository.go → NO CHANGE (OAuth spec term)
internal/identity/repository/orm/auth_profile_repository.go → authn_profile_repository.go (authentication profiles)
internal/identity/repository/orm/auth_profile_repository_test.go → authn_profile_repository_test.go
internal/identity/repository/orm/auth_flow_repository.go → authn_flow_repository.go (authentication flows)
internal/identity/repository/orm/auth_flow_repository_test.go → authn_flow_repository_test.go
```

**Test Files**:
```
internal/test/e2e/oauth_workflow_test.go → NO CHANGE (OAuth is correct term)
internal/identity/test/e2e/oauth_flows_test.go → NO CHANGE (OAuth is correct term)
internal/identity/test/e2e/oauth_flows_database_test.go → NO CHANGE (OAuth is correct term)
```

**Documentation**:
```
docs/runbooks/adaptive-auth-operations.md → adaptive-authn-operations.md (authentication)
testdata/adaptive-sim/sample-auth-logs.json → sample-authn-logs.json (authentication logs)
```

**Decision Rule**: 
- If file contains authentication logic (identity, login, credentials) → `authn`
- If file contains authorization logic (permissions, scopes, policies) → `authz`
- If file contains both → `authnz`
- If OAuth spec term (authorization_code, device_authorization) → NO CHANGE
- If already specific (`authz_server`, `authenticator`) → NO CHANGE

### Step 2: Code Content Replacements

**Categories requiring contextual replacement**:

1. **Variable Names** (200+ occurrences):
   - `auth_level` → `authn_level` (authentication level)
   - `auth_profile` → `authn_profile` (authentication profile)
   - `auth_flow` → `authn_flow` (authentication flow)
   - `current_auth` → `current_authn` (current authentication)
   - `insufficient_auth_level` → `insufficient_authn_level`
   - `client_auth_policy` → `client_authn_policy` (authentication policy)

2. **Function Names**:
   - `ServiceAuth()` → `ServiceAuthn()` OR `ServiceAuthz()` (analyze context)
   - `GetAuthProfile()` → `GetAuthnProfile()`
   - `ValidateAuth()` → `ValidateAuthn()` OR `ValidateAuthz()`

3. **Configuration Keys**:
   - `auth_method` → `authn_method` (authentication method)
   - `auth_backend` → `authn_backend` (authentication backend)
   - `auth_required` → `authn_required` (authentication required)

4. **API Endpoints** (⚠️ BREAKING CHANGES):
   - `/oauth/authorize` → NO CHANGE (OAuth 2.1 spec term)
   - `/oauth/token` → NO CHANGE (OAuth 2.1 spec term)
   - `/admin/auth-status` → `/admin/authn-status` (if authentication status)
   - `/admin/auth-config` → `/admin/authnz-config` (if combined)

5. **Database Columns** (⚠️ MIGRATION REQUIRED):
   - `auth_method` → `authn_method`
   - `auth_level` → `authn_level`
   - `current_auth` → `current_authn`

6. **Comments and Documentation**:
   - "auth check" → "authn check" OR "authz check" OR "authnz check"
   - "validate auth" → "validate authn" OR "validate authz"
   - "auth failed" → "authn failed" OR "authz failed"

### Step 3: Special Cases (NO CHANGE)

**OAuth 2.1 Standard Terms** (DO NOT CHANGE):
- `authorization` (OAuth 2.1 spec uses this spelling)
- `authorization_code` (OAuth 2.1 grant type)
- `device_authorization` (OAuth 2.1 RFC 8628)
- `pushed_authorization_request` (OAuth 2.1 RFC 9126)
- `authorization_endpoint` (OAuth 2.1 metadata)
- `oauth/authorize` (OAuth 2.1 endpoint)

**OIDC Standard Terms** (DO NOT CHANGE):
- `authentication` (OIDC 1.0 spec uses this spelling)
- `id_token` (OIDC 1.0 token)
- `userinfo_endpoint` (OIDC 1.0 endpoint)

**HTTP Headers** (DO NOT CHANGE):
- `Authorization: Bearer` (RFC 6750 standard header)
- `WWW-Authenticate` (RFC 7235 standard header)

**Already Specific Terms** (NO CHANGE NEEDED):
- `authz_server.go` (already uses `authz`)
- `authenticator.go` (clear authentication context)
- `openapi_spec_authz.yaml` (already uses `authz`)

### Step 4: Breaking Changes Assessment

**API Breaking Changes**:
- If any public API endpoints contain `/auth/` paths, renaming requires:
  - Version bump (if following semver)
  - Deprecation notice for old endpoints
  - Dual support period (old + new endpoints)
  - Client migration guide

**Configuration Breaking Changes**:
- If YAML/JSON configs contain `auth_*` keys, renaming requires:
  - Migration script or backward compatibility layer
  - Documentation update
  - User notification

**Database Breaking Changes**:
- If database columns contain `auth_*` names, renaming requires:
  - SQL migration scripts (ALTER TABLE)
  - Data preservation strategy
  - Rollback plan

## Implementation Strategy

### Option A: Big Bang (NOT RECOMMENDED)

**Approach**: Rename everything in one massive commit
**Pros**: Clean history, no transition period
**Cons**: High risk, breaks everything, hard to debug, difficult rollback

### Option B: Incremental (RECOMMENDED)

**Approach**: Refactor in phases with backward compatibility
**Pros**: Safe, testable, gradual migration, easy rollback
**Cons**: Longer timeline, dual code paths during transition

**Recommended Phases**:
1. **Phase 2a**: Internal code only (no API changes) - 10-15 commits
2. **Phase 2b**: Configuration keys with backward compat - 3-5 commits
3. **Phase 2c**: Database migrations (if needed) - 1-2 commits
4. **Phase 2d**: API endpoint deprecation (if needed) - 2-3 commits
5. **Phase 2e**: Remove deprecated code after grace period - 1 commit

### Option C: Selective (ALTERNATIVE)

**Approach**: Only rename most ambiguous cases, leave clear contexts unchanged
**Pros**: Lower risk, faster completion, fewer breaking changes
**Cons**: Inconsistent naming, some ambiguity remains

## Testing Requirements

**Per-Phase Validation**:
- ✅ All unit tests pass (`go test ./...`)
- ✅ All integration tests pass
- ✅ All E2E tests pass
- ✅ No new linting errors (`golangci-lint run`)
- ✅ Coverage maintained (≥95% production, ≥98% infra/utility)
- ✅ Mutation score maintained (≥85% Phase 4, ≥98% Phase 5+)

**Backward Compatibility Tests**:
- Old configuration keys still work (with deprecation warnings)
- Old API endpoints still work (with deprecation headers)
- Migration scripts preserve data integrity

## Risk Assessment

**High Risk Areas**:
- Public API endpoints (breaking changes for clients)
- Database schema (requires careful migration)
- OAuth 2.1 flow (standard terms must not change)

**Medium Risk Areas**:
- Configuration keys (backward compat possible)
- Internal variable names (no external impact)
- File/directory renames (git handles gracefully)

**Low Risk Areas**:
- Comments and documentation (no functional impact)
- Test files (isolated from production)
- Private functions (internal only)

## Timeline Estimate

**Option B (Incremental - RECOMMENDED)**:
- Phase 2a (Internal code): 2-3 days (10-15 commits)
- Phase 2b (Config keys): 1 day (3-5 commits)
- Phase 2c (Database): 0.5 days (1-2 commits, if needed)
- Phase 2d (API deprecation): 1 day (2-3 commits, if needed)
- Phase 2e (Cleanup): 0.5 days (1 commit, after grace period)
- **Total**: 5-6 days active work + grace period

**Option C (Selective - ALTERNATIVE)**:
- Select 20-30 most ambiguous files/variables
- Rename in 3-5 commits with full test validation
- **Total**: 1-2 days

## User Decision Required

**Questions for User**:
1. **Scope**: Option B (Incremental), Option C (Selective), or custom scope?
2. **Breaking Changes**: Are API endpoint renames acceptable? If so, what's the deprecation timeline?
3. **Database**: Are database column renames acceptable? If so, when to run migrations?
4. **Priority**: Which areas to refactor first (e.g., most confusing code, documentation, tests)?
5. **Timeline**: Acceptable timeline for completion (days/weeks)?

## Next Steps (After User Approval)

1. **Analysis**: Deep-dive analysis of all 89 files to classify authn vs authz vs authnz
2. **Plan Refinement**: Create detailed file-by-file renaming plan with context justification
3. **Test Baseline**: Run full test suite, record baseline coverage/mutation scores
4. **Execute Phase 2a**: Start with internal code renames (lowest risk)
5. **Validate**: Run tests after each commit, track regressions
6. **Iterate**: Continue through phases 2b-2e based on user approval

## Related Work (Phase 3+: Extract 22 Instruction Files)

**User's second request**: Move content from 23 instruction files to `.specify/memory/TOPIC.md` with references

**Affected Files**:
- 01-03.speckit.instructions.md → .specify/memory/speckit.md
- 02-01.architecture.instructions.md → .specify/memory/architecture.md
- 02-02.service-template.instructions.md → .specify/memory/service-template.md
- [... 19 more files]

**Timeline**: 22 commits × ~30 minutes each = 11 hours (1.5 days)

**Dependencies**: Should be done AFTER Phase 2 auth refactoring completes (to avoid double work)

---

**Status**: Awaiting user input on scope, approach, and priority before proceeding with execution.
