# Tasks - Service Template & CICD Fixes

**Status**: 0 of 48 tasks complete (0%)
**Last Updated**: 2026-01-31

## Task Checklist

### Phase 1: Copilot Instructions Updates

#### Task 1.1: Update service-template.instructions.md
- **Status**: ✅ Complete
- **Estimated**: 1h
- **File**: `.github/instructions/02-02.service-template.instructions.md`
- **Description**: Add migration numbering section, TestMain pattern, registration flow, *FromSettings pattern
- **Acceptance Criteria**:
  - [x] Migration versioning clear (1001-1004 vs 2001+)
  - [x] Complete TestMain pattern example
  - [x] Document *FromSettings factory pattern as preferred

#### Task 1.2: Update testing.instructions.md
- **Status**: ✅ Complete
- **Estimated**: 1h
- **File**: `.github/instructions/03-02.testing.instructions.md`
- **Description**: Move app.Test() pattern earlier, add cross-references
- **Acceptance Criteria**:
  - [x] app.Test() pattern prominent (already FORBIDDEN #2 at lines 67-104)
  - [x] Cross-reference to 07-01.testmain-integration-pattern.md (added)

#### Task 1.3: Update server-builder.instructions.md
- **Status**: ✅ Complete
- **Estimated**: 1h
- **File**: `.github/instructions/03-08.server-builder.instructions.md`
- **Description**: Complete merged migrations docs, add troubleshooting
- **Acceptance Criteria**:
  - [x] Merged migrations fully documented (already present at lines 143-170)
  - [x] Troubleshooting section added (5 common issues with causes/solutions)

---

### Phase 2: CICD Enforcement Improvements

#### Task 2.1: Docker Secrets Pattern Linter (#15)
- **Status**: ✅ Complete
- **Estimated**: 2h
- **Actual**: 1h
- **Commit**: fd29e3d4
- **Description**: Add `lint-compose: docker-secrets` to detect inline credentials
- **Acceptance Criteria**:
  - [x] Detects inline POSTGRES_PASSWORD, API_KEY, etc.
  - [x] Requires Docker secrets pattern

#### Task 2.2: Testify Require Over Assert Linter (#16)
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: Combined with 2.3-2.5 into requirepatterns.go
- **Commit**: fd29e3d4
- **Description**: Add `lint-go-test: testify-require` to enforce require over assert
- **Acceptance Criteria**:
  - [x] Detects assert.NoError → require.NoError
  - [x] Pre-commit hook enforces

#### Task 2.3: t.Parallel() Enforcement Linter (#17)
- **Status**: ✅ Complete (combined with 2.2)
- **Estimated**: 2h
- **Description**: Add `lint-go-test: t-parallel` to require t.Parallel() in tests
- **Acceptance Criteria**:
  - [x] Detects missing t.Parallel() in test functions
  - [x] Detects missing t.Parallel() in subtests

#### Task 2.4: Table-Driven Test Pattern Linter (#18)
- **Status**: ✅ Complete (combined with 2.2)
- **Estimated**: 2h
- **Description**: Add `lint-go-test: table-driven-tests` to detect standalone test functions that should be table-driven
- **Acceptance Criteria**:
  - [x] Detects multiple similar standalone test functions
  - [x] Suggests table-driven refactoring

#### Task 2.5: Hardcoded Test Passwords Linter (#19)
- **Status**: ✅ Complete (combined with 2.2)
- **Estimated**: 1h
- **Description**: Add `lint-go-test: no-hardcoded-passwords` to detect hardcoded passwords in tests
- **Acceptance Criteria**:
  - [x] Detects `password := "test123"` patterns
  - [x] Suggests UUIDv7 or magic constants

#### Task 2.6: crypto/rand Enforcement Linter (#20)
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 2h (including nolint handling)
- **Commit**: fd29e3d4
- **Description**: Add `lint-go: crypto-rand` to detect math/rand in crypto code
- **Acceptance Criteria**:
  - [x] Detects math/rand imports
  - [x] Suggests crypto/rand

#### Task 2.7: No Inline Env Vars Linter (#26)
- **Status**: ✅ Complete (covered by Task 2.1 dockersecrets.go)
- **Estimated**: 1h
- **Actual**: 0h (already implemented in Task 2.1)
- **Description**: Add `lint-compose: no-inline-env` to detect inline environment variables
- **Acceptance Criteria**:
  - [x] Detects `POSTGRES_PASSWORD: value` (not _FILE pattern)
  - [x] Requires Docker secrets or _FILE pattern

#### Task 2.8: No InsecureSkipVerify Linter (#28)
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 1h
- **Commit**: fd29e3d4
- **Description**: Add `lint-go: tls-verify` to detect InsecureSkipVerify: true
- **Acceptance Criteria**:
  - [x] Detects InsecureSkipVerify: true
  - [x] Fails CI if found

#### Task 2.9: golangci-lint v2 Schema Linter (#29 - CRITICAL)
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 1h
- **Description**: Add `lint-golangci-config: golangci-v2-schema` to validate v2 config
- **Acceptance Criteria**:
  - [x] Validates .golangci.yml against v2 schema
  - [x] Catches deprecated v1 options

---

### Phase 3: Deployment Fixes

#### Task 3.1: Fix Healthcheck Path Mismatch
- **Status**: ✅ Complete
- **Estimated**: 0.5h
- **Actual**: 0.25h
- **File**: `deployments/kms/compose.yml`
- **Description**: Fix healthcheck path from `/admin/v1/livez` to `/admin/api/v1/livez`
- **Acceptance Criteria**:
  - [x] Healthcheck uses `/admin/api/v1/livez` (matches service-template standard)
- **Evidence**: Fixed 3 healthcheck paths (kms-sqlite, kms-postgres-1, kms-postgres-2), compose syntax validated

#### Task 3.2: Create Template Deployment
- **Status**: ✅ Complete
- **Estimated**: 2h
- **Actual**: 0.5h
- **Location**: `deployments/template/`
- **Description**: Create compose.yml, secrets/, test E2E helper
- **Acceptance Criteria**:
  - [x] compose.yml exists and validates
  - [x] E2E tests can use template deployment (uses cipher-im binary as reference implementation)
  - [x] Follows Docker secrets pattern
- **Evidence**: Created compose.yml, otel-collector-config.yaml, secrets/ (4 files with 440 permissions), configs/template/ (3 config files). Compose syntax validated with `docker compose config`.

---

### Phase 4: Critical Fixes (TODOs and Security)

#### Task 4.1: Complete Registration Handler TODOs
- **Status**: ✅ Complete
- **Estimated**: 2h
- **Actual**: 3h
- **Commit**: 76e6b899
- **File**: `internal/apps/template/service/server/apis/registration_handlers.go`
- **Description**: Resolve 4 TODOs: validate request fields, hash password with PBKDF2-HMAC-SHA256, create user in DB, call registration service
- **Acceptance Criteria**:
  - [x] All 4 TODOs resolved
  - [x] Password hashing uses FIPS-approved algorithm (PBKDF2-HMAC-SHA256)
  - [x] Unit tests cover happy path and error cases (3/3 tests passing)
  - [x] Coverage ≥95%

#### Task 4.2: Add Admin Middleware to Registration Routes
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/server/apis/registration_routes.go:48`
- **Description**: Add admin authentication middleware to protected routes
- **Acceptance Criteria**:
  - [ ] TODO resolved
  - [ ] Admin routes protected
  - [ ] Tests verify 401/403 for non-admin access

#### Task 4.3: Implement Realm Lookup for Multi-Tenant
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **File**: `internal/apps/template/service/server/realms/handlers.go:270`
- **Description**: Implement realm resolution based on tenant context
- **Acceptance Criteria**:
  - [ ] TODO resolved
  - [ ] Multi-tenant realm lookup functional
  - [ ] Tests cover single-tenant and multi-tenant scenarios

---

### Phase 5: Test Architecture Refactoring

#### Task 5.1: Refactor Listener Tests to app.Test()
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

#### Task 5.2: Consolidate config_validation_test.go
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/config/config_validation_test.go`
- **Description**: Convert standalone TestValidateConfiguration_* functions to table-driven
- **Acceptance Criteria**:
  - [ ] Single table-driven function
  - [ ] Line count reduced by ~50%

#### Task 5.3: Consolidate session_manager_jws_test.go
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/server/businesslogic/session_manager_jws_test.go`
- **Description**: Convert standalone tests to table-driven pattern
- **Acceptance Criteria**:
  - [ ] Consolidated table-driven tests
  - [ ] t.Parallel() in all tests

#### Task 5.4: Consolidate session_manager_jwe_test.go
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/server/businesslogic/session_manager_jwe_test.go`
- **Description**: Convert standalone tests to table-driven pattern
- **Acceptance Criteria**:
  - [ ] Consolidated table-driven tests
  - [ ] t.Parallel() in all tests

#### Task 5.5: Consolidate config_coverage_test.go (12 functions)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/config/config_coverage_test.go`
- **Description**: 12 standalone functions → table-driven
- **Acceptance Criteria**:
  - [ ] Single table-driven test
  - [ ] ~80 lines reduced

#### Task 5.6: Consolidate application_middleware_test.go (7 functions)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/kms/server/application/application_middleware_test.go`
- **Description**: 7 TestSwaggerUIBasicAuthMiddleware_* functions → table-driven
- **Acceptance Criteria**:
  - [ ] Single table-driven test
  - [ ] ~100 lines reduced

#### Task 5.7: Consolidate tenant_join_request_test.go (4 functions)
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **File**: `internal/apps/template/service/server/domain/tenant_join_request_test.go`
- **Description**: 4 standalone functions → table-driven
- **Acceptance Criteria**:
  - [ ] Single table-driven test

#### Task 5.8: Consolidate jose-ja models_test.go (4 functions)
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **File**: `internal/jose/ja/domain/models_test.go`
- **Description**: 4 standalone functions → table-driven
- **Acceptance Criteria**:
  - [ ] Single table-driven test

#### Task 5.9: Consolidate kms seed_test.go (6 functions)
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **File**: `internal/kms/server/demo/seed_test.go`
- **Description**: 6 standalone functions → table-driven
- **Acceptance Criteria**:
  - [ ] Single table-driven test

---

### Phase 6: Coverage Improvements

#### Task 6.1: Repository Package (84.8% → 95%)
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Package**: `internal/apps/template/service/server/repository/`
- **Description**: Add tests for migration errors, CRUD edge cases, concurrent access
- **Acceptance Criteria**:
  - [ ] Coverage ≥95%
  - [ ] All error paths tested

#### Task 6.2: Application Package (89.8% → 95%)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Package**: `internal/apps/template/service/server/application/`
- **Description**: Add tests for DB provisioning failures, container mode fallbacks
- **Acceptance Criteria**:
  - [ ] Coverage ≥95%

#### Task 6.3: Businesslogic Package (87.4% → 95%)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Package**: `internal/apps/template/service/server/businesslogic/`
- **Description**: Add tests for session manager edge cases, tenant registration errors
- **Acceptance Criteria**:
  - [ ] Coverage ≥95%

#### Task 6.4: Config Packages (86.9%/87.1% → 95%)
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Packages**: `config/`, `config/tls_generator/`
- **Acceptance Criteria**:
  - [ ] Both packages ≥95%

#### Task 6.5: Remaining Packages
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Packages**: `server/builder/` (90.8%), `server/listener/` (88.2%), `server/barrier/` (91.2%)
- **Acceptance Criteria**:
  - [ ] All packages ≥95%

---

### Phase 7: Code Cleanup

#### Task 7.1: Investigate Low-Coverage Functions
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Description**: Check UnsealKeysServiceFromSettings (0% cov), EnsureSignatureAlgorithmType (23.1% cov) - determine if untested or actually unused
- **Note**: `*FromSettings` pattern is PREFERRED per project convention - these may need tests, not removal
- **Acceptance Criteria**:
  - [ ] Functions investigated
  - [ ] If needed: tests added OR documented as deprecated

#### Task 7.2: Fix Config Bug in config_gaps_test.go
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **File**: `internal/apps/template/service/config/config_gaps_test.go:37-39`
- **Description**: Fix acknowledged bug in NewFromFile or document as known limitation
- **Acceptance Criteria**:
  - [ ] Bug fixed or documented with issue reference

---

### Phase 8: Race Condition Testing (from v4 Phase 12)

#### Task 8.1: Enable Race Detection in CI/CD
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Description**: Add `-race` flag to CI test workflows, configure CGO_ENABLED=1
- **Acceptance Criteria**:
  - [ ] CI runs `go test -race ./...`
  - [ ] CGO properly configured for race detection

#### Task 8.2: Fix Race Conditions in Shared Packages
- **Status**: ❌ Not Started
- **Estimated**: 3h
- **Packages**: `internal/shared/pool/`, `internal/shared/barrier/`, `internal/shared/crypto/`
- **Description**: Add proper synchronization (sync.Mutex, sync.RWMutex, sync.Map)
- **Acceptance Criteria**:
  - [ ] All shared packages pass `-race` flag
  - [ ] No data races detected

#### Task 8.3: Fix Race Conditions in Service-Template
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Package**: `internal/apps/template/service/`
- **Description**: Fix concurrent access to session manager, realm service, configuration
- **Acceptance Criteria**:
  - [ ] Service-template passes `-race` flag
  - [ ] Concurrent tests pass with t.Parallel()

#### Task 8.4: Fix Race Conditions in Cipher-IM
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Package**: `internal/apps/cipher/im/`
- **Acceptance Criteria**:
  - [ ] Cipher-IM passes `-race` flag

#### Task 8.5: Fix Race Conditions in JOSE-JA
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Package**: `internal/jose/`
- **Acceptance Criteria**:
  - [ ] JOSE-JA passes `-race` flag

#### Task 8.6: Fix Race Conditions in KMS
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Package**: `internal/kms/`
- **Description**: KMS has custom concurrency patterns, may need more fixes
- **Acceptance Criteria**:
  - [ ] KMS passes `-race` flag

---

### Phase 9: KMS Modernization (from v4 Phase 6 - EXECUTE LAST)

**CRITICAL: Execute Phase 9 LAST after Phases 1-8 complete**

**NOTE: Prerelease project - backward compatibility NOT required**

#### Task 9.1: Create KMS ServerBuilder Migration Plan
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Description**: Document migration strategy from raw database/sql to GORM ServerBuilder
- **Acceptance Criteria**:
  - [ ] Migration plan documented
  - [ ] Breaking changes identified (no backward compat needed)
  - [ ] File diff analysis (what gets deleted vs modified)

#### Task 9.2: Migrate KMS Database Layer to GORM
- **Status**: ❌ Not Started
- **Estimated**: 8h
- **Package**: `internal/kms/server/repository/`
- **Description**: Replace raw database/sql with GORM
- **Acceptance Criteria**:
  - [ ] All repositories use GORM
  - [ ] Existing tests pass
  - [ ] Cross-DB compatible (PostgreSQL + SQLite)

#### Task 9.3: Migrate KMS to ServerBuilder Pattern
- **Status**: ❌ Not Started
- **Estimated**: 4h
- **Package**: `internal/kms/server/`
- **Description**: Replace custom application_listener.go (~1,500 lines) with ServerBuilder
- **Acceptance Criteria**:
  - [ ] ServerBuilder used for initialization
  - [ ] Custom application_listener.go deleted
  - [ ] Merged migrations pattern implemented (1001-1004 + 2001+)

#### Task 9.4: KMS E2E Test Update
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Description**: Update E2E tests for modernized KMS
- **Acceptance Criteria**:
  - [ ] E2E tests pass with Docker Compose
  - [ ] All API paths tested (/service/** and /browser/**)
  - [ ] Health checks working (/admin/api/v1/livez, /readyz)

---

## Cross-Cutting Tasks

### Documentation
- [ ] README.md updated
- [ ] Archive docs preserved

### Testing
- [ ] All unit tests ≥95% coverage
- [ ] Table-driven tests only
- [ ] No real HTTPS listeners
- [ ] Race detection clean (`go test -race ./...`)

### Quality
- [ ] Linting passes
- [ ] No security vulnerabilities
- [ ] No TODOs in production code

### KMS Modernization (Phase 9)
- [ ] GORM migration complete
- [ ] ServerBuilder pattern adopted
- [ ] E2E tests passing

---

## References

- Analysis docs: [archive/](./archive/)
- [03-02.testing.instructions.md](../../.github/instructions/03-02.testing.instructions.md)
- [07-01.testmain-integration-pattern.instructions.md](../../.github/instructions/07-01.testmain-integration-pattern.instructions.md)
- Comparison table: [comparison-table.md](./comparison-table.md)
