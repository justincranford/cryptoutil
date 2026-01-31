# V6 Implementation Plan - Rock Solid

**Version:** 6.0
**Date:** 2026-01-31
**Purpose:** Comprehensive fix plan for service-template and copilot instructions

---

## Plan Overview

This plan addresses all issues identified in:
1. [01-copilot-instructions-analysis.md](./01-copilot-instructions-analysis.md)
2. [02-service-template-analysis.md](./02-service-template-analysis.md)

### Success Criteria
- [ ] All unit tests pass
- [ ] Coverage ≥95% for production code, ≥98% for infrastructure
- [ ] All linting errors resolved
- [ ] All TODOs addressed or tracked
- [ ] No standalone tests (table-driven only)
- [ ] No real HTTPS listeners in unit tests
- [ ] All copilot instructions accurate and complete

---

## Phase 1: Critical Fixes (TODOs and Security)

**Estimated LOE:** 4-6 hours
**Priority:** HIGH - Security and functionality gaps

### Task 1.1: Complete Registration Handler TODOs
**File:** `internal/apps/template/service/server/apis/registration_handlers.go`

**Current State:**
```go
// TODO: Validate request fields
// TODO: Hash password
// TODO: Create user in database
// TODO: Call registration service
```

**Required Actions:**
1. Implement request validation using validation tags
2. Implement password hashing using PBKDF2-HMAC-SHA256 (per 02-08.hashes.instructions.md)
3. Create user via user repository
4. Integrate with registration service

**Acceptance Criteria:**
- [ ] All 4 TODOs resolved
- [ ] Password hashing uses FIPS-approved algorithm
- [ ] Unit tests cover happy path and error cases
- [ ] Coverage ≥95%

---

### Task 1.2: Add Admin Middleware to Registration Routes
**File:** `internal/apps/template/service/server/apis/registration_routes.go:48`

**Current State:**
```go
// These endpoints require admin authentication (TODO: add admin middleware).
```

**Required Actions:**
1. Implement or use existing admin authentication middleware
2. Apply to protected registration routes
3. Add tests verifying unauthorized access is blocked

**Acceptance Criteria:**
- [ ] TODO resolved
- [ ] Admin routes protected
- [ ] Tests verify 401/403 for non-admin access

---

### Task 1.3: Implement Realm Lookup for Multi-Tenant
**File:** `internal/apps/template/service/server/realms/handlers.go:270`

**Current State:**
```go
// TODO: Implement proper realm lookup for multi-tenant deployments.
```

**Required Actions:**
1. Implement realm resolution based on tenant context
2. Support configurable realm selection strategy
3. Add multi-tenant realm lookup tests

**Acceptance Criteria:**
- [ ] TODO resolved
- [ ] Multi-tenant realm lookup functional
- [ ] Tests cover single-tenant and multi-tenant scenarios

---

## Phase 2: Test Architecture Refactoring

**Estimated LOE:** 6-8 hours
**Priority:** HIGH - CI/CD reliability and maintainability

### Task 2.1: Refactor Listener Tests to Use app.Test()
**Files:**
- `internal/apps/template/service/server/listener/servers_test.go`
- `internal/apps/template/service/server/listener/application_listener_test.go`

**Current State:**
Tests create real HTTPS listeners via `NewHTTPServers()`.

**Required Actions:**
1. Identify tests that truly need real listeners (E2E) vs handler tests
2. Refactor handler tests to use Fiber `app.Test()`
3. Keep only essential listener tests (lifecycle, port allocation)

**Pattern:**
```go
// BEFORE (real listener)
h, _ := NewHTTPServers(ctx, settings)
resp, _ := http.Get(h.PublicBaseURL() + "/path")

// AFTER (in-memory)
app := fiber.New()
app.Get("/path", handler)
req := httptest.NewRequest("GET", "/path", nil)
resp, _ := app.Test(req, -1)
```

**Acceptance Criteria:**
- [ ] No Windows Firewall triggers
- [ ] No port binding in unit tests
- [ ] Tests run faster (<1ms vs 10-50ms)
- [ ] All tests still pass

---

### Task 2.2: Consolidate Standalone Tests to Table-Driven
**Files:**
- `internal/apps/template/service/config/config_validation_test.go`
- `internal/apps/template/service/server/businesslogic/session_manager_jws_test.go`
- `internal/apps/template/service/server/businesslogic/session_manager_jwe_test.go`

**Current State:**
```go
func TestValidateConfiguration_InvalidProtocol(t *testing.T) { ... }
func TestValidateConfiguration_InvalidLogLevel(t *testing.T) { ... }
```

**Required Pattern:**
```go
func TestValidateConfiguration_Errors(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name    string
        modify  func(*Settings)
        wantErr string
    }{
        {name: "invalid protocol", modify: func(s *Settings) { s.Protocol = "ftp" }, wantErr: "protocol"},
        {name: "invalid log level", modify: func(s *Settings) { s.LogLevel = "invalid" }, wantErr: "log level"},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            // test logic
        })
    }
}
```

**Acceptance Criteria:**
- [ ] config_validation_test.go: Single table-driven function
- [ ] session_manager_jws_test.go: Consolidated table-driven
- [ ] session_manager_jwe_test.go: Consolidated table-driven
- [ ] Line count reduced by ~50%

---

## Phase 3: Coverage Improvements

**Estimated LOE:** 4-6 hours
**Priority:** MEDIUM - Quality gate compliance

### Task 3.1: Improve Repository Package Coverage (84.8% → 95%)
**Package:** `internal/apps/template/service/server/repository/`

**Required Actions:**
1. Generate coverage HTML: `go test -coverprofile=repo.cov && go tool cover -html=repo.cov`
2. Identify RED lines (uncovered code)
3. Add targeted tests for:
   - Migration error paths
   - Repository CRUD edge cases
   - Concurrent access scenarios

**Acceptance Criteria:**
- [ ] Coverage ≥95%
- [ ] All error paths tested

---

### Task 3.2: Improve Application Package Coverage (89.8% → 95%)
**Package:** `internal/apps/template/service/server/application/`

**Required Actions:**
1. Analyze coverage gaps via HTML report
2. Add tests for:
   - Database provisioning failures
   - Container mode fallbacks
   - Service initialization errors

**Acceptance Criteria:**
- [ ] Coverage ≥95%

---

### Task 3.3: Improve Businesslogic Package Coverage (87.4% → 95%)
**Package:** `internal/apps/template/service/server/businesslogic/`

**Required Actions:**
1. Analyze coverage gaps
2. Add tests for:
   - Session manager edge cases
   - Tenant registration errors
   - Concurrent session handling

**Acceptance Criteria:**
- [ ] Coverage ≥95%

---

### Task 3.4: Improve Remaining Below-Target Packages
**Packages:**
- config (86.9% → 95%)
- config/tls_generator (87.1% → 95%)
- server/builder (90.8% → 95%)
- server/listener (88.2% → 95%)
- server/barrier (91.2% → 95%)

**Acceptance Criteria:**
- [ ] All packages ≥95%

---

## Phase 4: Code Cleanup

**Estimated LOE:** 2-3 hours
**Priority:** LOW - Code hygiene

### Task 4.1: Verify and Remove Dead Code
**Potential Dead Code:**
- UnsealKeysServiceFromSettings wrappers (0% coverage per V4)
- EnsureSignatureAlgorithmType (23.1% coverage)
- PublicServer.PublicBaseURL duplicate method

**Required Actions:**
1. Verify each item is actually unused
2. Remove confirmed dead code
3. Document intentional code that appears unused

**Acceptance Criteria:**
- [ ] No dead code or documented exceptions

---

### Task 4.2: Fix Config Bug Acknowledged in Tests
**File:** `internal/apps/template/service/config/config_gaps_test.go:37-39`

**Current State:**
Test acknowledges bug in `NewFromFile` but doesn't fix it.

**Required Actions:**
1. Analyze the actual bug in NewFromFile
2. Fix the bug OR document as known limitation
3. Update test to reflect correct behavior

**Acceptance Criteria:**
- [ ] Bug fixed or documented with issue reference

---

## Phase 5: Deployment and Workflow

**Estimated LOE:** 2-3 hours
**Priority:** LOW - Completeness

### Task 5.1: Create Template Deployment
**Location:** `deployments/template/`

**Required Actions:**
1. Create `deployments/template/compose.yml`
2. Create `deployments/template/secrets/` with placeholder secrets
3. Test E2E compose helper with new deployment

**Acceptance Criteria:**
- [ ] compose.yml exists and validates
- [ ] E2E tests can use template deployment
- [ ] Follows Docker secrets pattern

---

### Task 5.2: Document Template Testing Strategy
**Alternative:** If cipher-im serves as template validation, document this explicitly.

**Required Actions:**
1. Add README to `internal/apps/template/` explaining:
   - Purpose of template
   - How to use as reference
   - Testing strategy (tested via cipher-im)

**Acceptance Criteria:**
- [ ] Documentation exists
- [ ] Strategy is clear

---

## Phase 6: Copilot Instructions Updates

**Estimated LOE:** 2-3 hours
**Priority:** MEDIUM - Documentation accuracy

### Task 6.1: Update 02-02.service-template.instructions.md
**Issues from Analysis:**
- Migration versioning unclear (1001-1004 vs 2001+)
- Missing TestMain registration pattern

**Required Actions:**
1. Add clear migration numbering section
2. Add complete TestMain pattern example
3. Document registration flow requirement

---

### Task 6.2: Update 03-02.testing.instructions.md
**Issues from Analysis:**
- app.Test() pattern could be more prominent
- Missing cross-reference to TestMain pattern

**Required Actions:**
1. Move app.Test() pattern earlier in document
2. Add cross-reference to 07-01.testmain-integration-pattern.md
3. Ensure table-driven test emphasis is clear

---

### Task 6.3: Update 03-08.server-builder.instructions.md
**Issues from Analysis:**
- Merged migrations pattern incomplete
- Missing error handling guidance

**Required Actions:**
1. Complete merged migrations documentation
2. Add error handling patterns
3. Add troubleshooting section

---

## Execution Order

```
Phase 1 (Critical) ──→ Phase 2 (Test Architecture) ──→ Phase 3 (Coverage)
                                                              │
                                                              ▼
Phase 6 (Instructions) ←── Phase 5 (Deployment) ←── Phase 4 (Cleanup)
```

**Rationale:**
1. Phase 1 first: Security TODOs are blocking issues
2. Phase 2 next: Test architecture enables reliable coverage measurement
3. Phase 3 follows: Can accurately measure coverage after test refactoring
4. Phase 4-6: Lower priority cleanup and documentation

---

## Tracking

### Phase Status
| Phase | Status | Start | Complete |
|-------|--------|-------|----------|
| 1: Critical Fixes | ⬜ Not Started | - | - |
| 2: Test Architecture | ⬜ Not Started | - | - |
| 3: Coverage | ⬜ Not Started | - | - |
| 4: Code Cleanup | ⬜ Not Started | - | - |
| 5: Deployment | ⬜ Not Started | - | - |
| 6: Instructions | ⬜ Not Started | - | - |

### Quality Gates
- [ ] `go build ./internal/apps/template/...` passes
- [ ] `golangci-lint run ./internal/apps/template/...` passes
- [ ] `go test ./internal/apps/template/...` passes
- [ ] Coverage ≥95% all production packages
- [ ] Zero TODOs in production code
- [ ] Zero standalone tests

---

## Cross-References

- [01-copilot-instructions-analysis.md](./01-copilot-instructions-analysis.md)
- [02-service-template-analysis.md](./02-service-template-analysis.md)
- [V5 Review](../fixes-needed-plan-tasks-v5/review-tasks-v4.md)
