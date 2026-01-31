# Tasks - Service Template & CICD Fixes

**Status**: 0 of 25 tasks complete (0%)
**Last Updated**: 2026-01-31

## Task Checklist

### Phase 1: Critical Fixes (TODOs and Security)

#### Task 1.1: Complete Registration Handler TODOs
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **File**: `internal/apps/template/service/server/apis/registration_handlers.go`
- **Description**: Resolve 4 TODOs: validate request fields, hash password with PBKDF2-HMAC-SHA256, create user in DB, call registration service
- **Acceptance Criteria**:
  - [ ] All 4 TODOs resolved
  - [ ] Password hashing uses FIPS-approved algorithm
  - [ ] Unit tests cover happy path and error cases
  - [ ] Coverage ≥95%

#### Task 1.2: Add Admin Middleware to Registration Routes
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/server/apis/registration_routes.go:48`
- **Description**: Add admin authentication middleware to protected routes
- **Acceptance Criteria**:
  - [ ] TODO resolved
  - [ ] Admin routes protected
  - [ ] Tests verify 401/403 for non-admin access

#### Task 1.3: Implement Realm Lookup for Multi-Tenant
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **File**: `internal/apps/template/service/server/realms/handlers.go:270`
- **Description**: Implement realm resolution based on tenant context
- **Acceptance Criteria**:
  - [ ] TODO resolved
  - [ ] Multi-tenant realm lookup functional
  - [ ] Tests cover single-tenant and multi-tenant scenarios

---

### Phase 2: Test Architecture Refactoring

#### Task 2.1: Refactor Listener Tests to app.Test()
- **Status**: ❌ Not Started
- **Estimated**: 3h
- **Files**:
  - `internal/apps/template/service/server/listener/servers_test.go`
  - `internal/apps/template/service/server/listener/application_listener_test.go`
- **Description**: Replace real HTTPS listeners with Fiber app.Test() for in-memory testing
- **Acceptance Criteria**:
  - [ ] No Windows Firewall triggers
  - [ ] No port binding in unit tests
  - [ ] Tests run faster (<1ms vs 10-50ms)
  - [ ] All tests still pass

#### Task 2.2: Consolidate config_validation_test.go
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/config/config_validation_test.go`
- **Description**: Convert standalone TestValidateConfiguration_* functions to table-driven
- **Acceptance Criteria**:
  - [ ] Single table-driven function
  - [ ] Line count reduced by ~50%

#### Task 2.3: Consolidate session_manager_jws_test.go
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/server/businesslogic/session_manager_jws_test.go`
- **Description**: Convert standalone tests to table-driven pattern
- **Acceptance Criteria**:
  - [ ] Consolidated table-driven tests
  - [ ] t.Parallel() in all tests

#### Task 2.4: Consolidate session_manager_jwe_test.go
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/server/businesslogic/session_manager_jwe_test.go`
- **Description**: Convert standalone tests to table-driven pattern
- **Acceptance Criteria**:
  - [ ] Consolidated table-driven tests
  - [ ] t.Parallel() in all tests

#### Task 2.5: Consolidate config_coverage_test.go (12 functions)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/config/config_coverage_test.go`
- **Description**: 12 standalone functions → table-driven
- **Acceptance Criteria**:
  - [ ] Single table-driven test
  - [ ] ~80 lines reduced

#### Task 2.6: Consolidate application_middleware_test.go (7 functions)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/kms/server/application/application_middleware_test.go`
- **Description**: 7 TestSwaggerUIBasicAuthMiddleware_* functions → table-driven
- **Acceptance Criteria**:
  - [ ] Single table-driven test
  - [ ] ~100 lines reduced

---

### Phase 3: Coverage Improvements

#### Task 3.1: Repository Package (84.8% → 95%)
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Package**: `internal/apps/template/service/server/repository/`
- **Description**: Add tests for migration errors, CRUD edge cases, concurrent access
- **Acceptance Criteria**:
  - [ ] Coverage ≥95%
  - [ ] All error paths tested

#### Task 3.2: Application Package (89.8% → 95%)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Package**: `internal/apps/template/service/server/application/`
- **Description**: Add tests for DB provisioning failures, container mode fallbacks
- **Acceptance Criteria**:
  - [ ] Coverage ≥95%

#### Task 3.3: Businesslogic Package (87.4% → 95%)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Package**: `internal/apps/template/service/server/businesslogic/`
- **Description**: Add tests for session manager edge cases, tenant registration errors
- **Acceptance Criteria**:
  - [ ] Coverage ≥95%

#### Task 3.4: Config Packages (86.9%/87.1% → 95%)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Packages**: `config/`, `config/tls_generator/`
- **Acceptance Criteria**:
  - [ ] Both packages ≥95%

#### Task 3.5: Remaining Packages
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Packages**: `server/builder/` (90.8%), `server/listener/` (88.2%), `server/barrier/` (91.2%)
- **Acceptance Criteria**:
  - [ ] All packages ≥95%

---

### Phase 4: Code Cleanup

#### Task 4.1: Verify and Remove Dead Code
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Description**: Check UnsealKeysServiceFromSettings (0% cov), EnsureSignatureAlgorithmType (23.1% cov)
- **Acceptance Criteria**:
  - [ ] Dead code removed or documented

#### Task 4.2: Fix Config Bug in config_gaps_test.go
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/config/config_gaps_test.go:37-39`
- **Description**: Fix acknowledged bug in NewFromFile or document as known limitation
- **Acceptance Criteria**:
  - [ ] Bug fixed or documented with issue reference

---

### Phase 5: CICD Enforcement Improvements

#### Task 5.1: Docker Secrets Pattern Linter
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Description**: Add `lint-compose: docker-secrets` to detect inline credentials
- **Acceptance Criteria**:
  - [ ] Detects inline POSTGRES_PASSWORD, API_KEY, etc.
  - [ ] Requires Docker secrets pattern

#### Task 5.2: Testify Require Over Assert Linter
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Description**: Add `lint-go-test: testify-require` to enforce require over assert
- **Acceptance Criteria**:
  - [ ] Detects assert.NoError → require.NoError
  - [ ] Pre-commit hook enforces

#### Task 5.3: t.Parallel() Enforcement Linter
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Description**: Add `lint-go-test: t-parallel` to require t.Parallel() in tests
- **Acceptance Criteria**:
  - [ ] Detects missing t.Parallel() in test functions
  - [ ] Detects missing t.Parallel() in subtests

#### Task 5.4: crypto/rand Enforcement Linter
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Description**: Add `lint-go: crypto-rand` to detect math/rand in crypto code
- **Acceptance Criteria**:
  - [ ] Detects math/rand imports
  - [ ] Suggests crypto/rand

#### Task 5.5: No InsecureSkipVerify Linter
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Description**: Add `lint-go: tls-verify` to detect InsecureSkipVerify: true
- **Acceptance Criteria**:
  - [ ] Detects InsecureSkipVerify: true
  - [ ] Fails CI if found

---

### Phase 6: Deployment and Workflow

#### Task 6.1: Create Template Deployment
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Location**: `deployments/template/`
- **Description**: Create compose.yml, secrets/, test E2E helper
- **Acceptance Criteria**:
  - [ ] compose.yml exists and validates
  - [ ] E2E tests can use template deployment
  - [ ] Follows Docker secrets pattern

#### Task 6.2: Document Template Testing Strategy
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/README.md`
- **Description**: Document how cipher-im validates template
- **Acceptance Criteria**:
  - [ ] README explains testing strategy

---

### Phase 7: Copilot Instructions Updates

#### Task 7.1: Update service-template.instructions.md
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `.github/instructions/02-02.service-template.instructions.md`
- **Description**: Add migration numbering section, TestMain pattern, registration flow
- **Acceptance Criteria**:
  - [ ] Migration versioning clear (1001-1004 vs 2001+)
  - [ ] Complete TestMain pattern example

#### Task 7.2: Update testing.instructions.md
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `.github/instructions/03-02.testing.instructions.md`
- **Description**: Move app.Test() pattern earlier, add cross-references
- **Acceptance Criteria**:
  - [ ] app.Test() pattern prominent
  - [ ] Cross-reference to 07-01.testmain-integration-pattern.md

#### Task 7.3: Update server-builder.instructions.md
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `.github/instructions/03-08.server-builder.instructions.md`
- **Description**: Complete merged migrations docs, add troubleshooting
- **Acceptance Criteria**:
  - [ ] Merged migrations fully documented
  - [ ] Troubleshooting section added

---

## Cross-Cutting Tasks

### Documentation
- [ ] README.md updated
- [ ] Archive docs preserved

### Testing
- [ ] All unit tests ≥95% coverage
- [ ] Table-driven tests only
- [ ] No real HTTPS listeners

### Quality
- [ ] Linting passes
- [ ] No security vulnerabilities
- [ ] No TODOs in production code

---

## References

- Analysis docs: [archive/](./archive/)
- [03-02.testing.instructions.md](../../.github/instructions/03-02.testing.instructions.md)
- [07-01.testmain-integration-pattern.instructions.md](../../.github/instructions/07-01.testmain-integration-pattern.instructions.md)
