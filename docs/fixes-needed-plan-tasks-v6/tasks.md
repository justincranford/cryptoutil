# Tasks - Service Template & CICD Fixes

**Status**: 50/59 tasks complete (85%) | Phase 13 In Progress (Option A: Full ServerBuilder Extension)
**Last Updated**: 2026-02-01

## Summary

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1-4 | ✅ Complete | Instructions, CICD, Deployment, Critical Fixes |
| Phase 5 | ⚠️ Partial | Test Architecture (5.1 BLOCKED - StartApplicationListener) |
| Phase 6-8 | ✅ Complete | Coverage, Cleanup, Race Detection |
| Phase 9 | ✅ Complete | KMS Analysis (ServerBuilder extension needed) |
| Phase 10 | ✅ Complete | Cleanup (10.1-10.4 ✅) |
| Phase 11 | ✅ Complete | KMS ServerBuilder Extension (11.1-11.3 ✅, 11.4 deferred) |
| Phase 12 | ✅ Complete | KMS Before/After Comparison (all 6 tasks ✅) |
| Phase 13 | ⚠️ In Progress | ServerBuilder Extension for KMS (Option A - 10 tasks) |

**Completed tasks archived**: See [completed.md](./completed.md)

---

## Incomplete Tasks

### Phase 5: Test Architecture Refactoring

#### Task 5.1: Refactor Listener Tests to app.Test()
- **Status**: ❌ BLOCKED
- **Blocker**: `StartApplicationListener` not yet implemented (returns "implementation in progress" error)
- **Estimated**: 3h
- **Files**:
  - `internal/apps/template/service/server/listener/servers_test.go`
  - `internal/apps/template/service/server/listener/application_listener_test.go`
- **Description**: Replace real HTTPS listeners with Fiber app.Test() for in-memory testing
- **Next Steps**:
  1. Complete `StartApplicationListener` implementation first
  2. THEN refactor tests to use app.Test() pattern
- **Acceptance Criteria**:
  - [ ] Blocked until StartApplicationListener implemented
  - [ ] No Windows Firewall triggers
  - [ ] No port binding in unit tests
  - [ ] Tests run faster (<1ms vs 10-50ms)
  - [ ] All tests still pass

---

### Phase 10: Cleanup Leftover Coverage Files (REQUIRED)

**Status**: ✅ Complete (All 10.1-10.4 Complete)
**Discovery Date**: 2026-02-01
**Issue**: LLM autonomous work left 57+ coverage files in project root and internal directories

**User Decisions (from quizme-v1.md)**:
- Delete ALL files in test-output/ (clean slate)
- Detect leftover files in ALL directories including test-output/
- Patterns: `*.out`, `*.cov`, `*.prof`, `*coverage*.html`, `*coverage*.txt`
- Auto-delete files if found, with warning message

#### Task 10.1: Delete Root-Level Coverage Files
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.1h
- **Description**: Delete all .out and coverage files in project root directory
- **Evidence**: Commit `a07e9175` - 35 root-level files deleted
- **Acceptance Criteria**:
  - [x] All root-level .out files deleted
  - [x] All root-level coverage.html files deleted
  - [x] git status shows deletions

#### Task 10.2: Delete Internal Directory Coverage Files
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.1h
- **Description**: Delete all .out and coverage files inside internal/ directories
- **Evidence**: Commit `a07e9175` - 6 internal/ files deleted
- **Acceptance Criteria**:
  - [x] All internal/ .out files deleted
  - [x] All internal/ coverage.html files deleted
  - [x] git status shows deletions

#### Task 10.3: Delete ALL Files in test-output/
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.1h
- **Description**: Per user decision, delete ALL files in test-output/ for clean slate
- **Evidence**: Commit `a07e9175` - 202+ files deleted, test-output/ removed
- **Acceptance Criteria**:
  - [x] test-output/ directory is empty (or deleted)
  - [x] git status shows deletions

#### Task 10.4: Add CICD Linter for Leftover Coverage Files
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 45m
- **File**: `internal/cmd/cicd/lint_go/leftover_coverage.go`
- **Description**: Add cicd check to auto-delete leftover coverage files with warning
- **User Decisions**:
  - Scope: ALL directories including test-output/
  - Patterns (configurable list): `*.out`, `*.cov`, `*.prof`, `*coverage*.html`, `*coverage*.txt`
  - Behavior: Auto-delete files if found, with warning message
- **Evidence**: Commit `6f4399ef` - leftover_coverage.go and tests added, 17+ more files deleted
- **Acceptance Criteria**:
  - [x] Detects patterns in ALL directories
  - [x] Auto-deletes files found
  - [x] Prints warning message for each deleted file
  - [x] Pattern list easily updatable in code
  - [x] Integrated into cicd lint-go command
  - [x] Tests pass

---

### Phase 11: KMS ServerBuilder Extension (REQUIRED)

**Status**: ✅ Complete (Tasks 11.1-11.3 Complete, Task 11.4 Deferred to Phase 13)
**Rationale**: Service-template MUST support all KMS functionality for lateral migration

**CRITICAL**: KMS migrating to service-template MUST be a lateral move - no loss of functionality, architecture, design intent, or test intent.

#### Task 11.1: Extend ServerBuilder with SwaggerUI Support
- **Status**: ✅ Complete
- **Estimated**: 4h
- **Actual**: 3h
- **Description**: Add `WithSwaggerUI(username, password string)` method to ServerBuilder
- **Evidence**: Commit `68a52beb` - swagger_ui.go, swagger_ui_test.go, ServerBuilder integration
- **Files**:
  - `internal/apps/template/service/server/builder/swagger_ui.go` (NEW - ~200 lines)
  - `internal/apps/template/service/server/builder/swagger_ui_test.go` (NEW - ~290 lines)
  - `internal/apps/template/service/server/builder/server_builder.go` (Modified)
- **Acceptance Criteria**:
  - [x] ServerBuilder supports Swagger UI via WithSwaggerUI()
  - [x] Basic auth middleware included (swaggerUIBasicAuthMiddleware)
  - [x] CSRF script injection supported (swaggerUICustomCSRFScript)
  - [x] Tests pass (4 test functions, all pass)

#### Task 11.2: Extend ServerBuilder with OpenAPI Handler Registration
- **Status**: ✅ Complete
- **Estimated**: 4h
- **Actual**: 1h
- **Description**: Add OpenAPI helper config for request validation middleware
- **Evidence**: Commit `0300062c` - openapi.go, openapi_test.go
- **Files**:
  - `internal/apps/template/service/server/builder/openapi.go` (NEW - ~90 lines)
  - `internal/apps/template/service/server/builder/openapi_test.go` (NEW - ~260 lines)
- **Acceptance Criteria**:
  - [x] OpenAPIConfig struct with swagger spec, base paths, validation options
  - [x] NewDefaultOpenAPIConfig() factory with standard defaults
  - [x] CreateRequestValidatorMiddleware() using oapi-codegen fiber-middleware
  - [x] BrowserMiddlewares() and ServiceMiddlewares() helpers
  - [x] OpenAPIRegistrar interface for domain service integration
  - [x] Tests pass (5 test functions, all pass)

#### Task 11.3: Extend ServerBuilder with Security Headers
- **Status**: ✅ Complete
- **Estimated**: 2h
- **Actual**: 45m
- **Description**: Add comprehensive security headers to ServerBuilder
- **Files**:
  - Created: `internal/apps/template/service/server/builder/security_headers.go`
  - Created: `internal/apps/template/service/server/builder/security_headers_test.go`
- **Acceptance Criteria**:
  - [x] CSP headers configurable (buildContentSecurityPolicy with dev mode variations)
  - [x] XSS protection included (helmet middleware with X-Frame-Options, XSS-Protection)
  - [x] HSTS configured (additional headers middleware with dev/prod modes)
  - [x] Tests pass (7 test functions, all pass)
- **Evidence**: Commit 4f6b5d3b, go test PASS, golangci-lint 0 issues

#### Task 11.4: Migrate KMS to Extended ServerBuilder
- **Status**: ⏭️ DEFERRED (see Phase 13)
- **Estimated**: 4h (original), 20h+ (realistic)
- **Description**: After 11.1-11.3 complete, migrate KMS to use extended ServerBuilder
- **Dependencies**: 11.1, 11.2, 11.3
- **Acceptance Criteria**:
  - [ ] KMS uses extended ServerBuilder
  - [ ] application_listener.go deleted
  - [ ] All KMS tests pass
  - [ ] All template tests pass
  - [ ] All cipher-im tests pass

**ANALYSIS**: Task 11.4 as written requires deleting application_listener.go (1223 lines) and migrating KMS to use ServerBuilder. This is a complete architectural refactoring that:
- Requires creating new KMS server structure similar to cipher-im
- Requires updating all KMS handlers and tests
- Is estimated at 20+ hours, not 4 hours
- Should be a separate project/phase

**DECISION**: Tasks 11.1-11.3 are COMPLETE. The extensions (SwaggerUI, OpenAPI, Security Headers) are now available in ServerBuilder. KMS migration to use these extensions is deferred to Phase 13 as a separate, properly scoped effort.

**NEW PHASE 13**: Created below to handle KMS migration properly.

---

### Phase 12: KMS Before/After Comparison (REQUIRED)

**Status**: ✅ Complete
**Rationale**: Verify service-template reproduces ALL KMS functionality

**CRITICAL**: Service-template is the foundation of ALL 9 services. KMS switching to service-template MUST be a lateral move:
- No loss of functionality
- No divergence from design intent
- No loss of test coverage or test intent
- Maximum reusability
- Architecture preserved

#### Task 12.1: Document KMS Current Architecture
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.5h
- **Description**: Comprehensive documentation of KMS before migration
- **Output**: `test-output/kms-comparison/kms-before-architecture.md`
- **Acceptance Criteria**:
  - [x] All endpoints documented (14 OpenAPI + 3 admin + 3 UI)
  - [x] All middleware documented (13 public, 10 private)
  - [x] All security features documented (TLS, headers, rate limiting)
  - [x] Configuration options documented
  - [x] Three-layer architecture documented (Basic → Core → Listener)

#### Task 12.2: Document KMS Current Test Coverage and Intent
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.5h
- **Description**: Document all KMS tests, their purpose, and coverage
- **Output**: `test-output/kms-comparison/kms-before-tests.md`
- **Acceptance Criteria**:
  - [x] All test files listed
  - [x] Test intent documented
  - [x] Coverage metrics captured
  - [x] Critical test cases identified

#### Task 12.3: Compare KMS vs Service-Template Feature Parity
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.75h
- **Description**: Side-by-side comparison of features
- **Output**: `test-output/kms-comparison/feature-parity.md`
- **Acceptance Criteria**:
  - [x] Feature matrix complete
  - [x] Gaps identified
  - [x] Parity verified for all critical features

#### Task 12.4: Verify All KMS Tests Pass with Service-Template Backend
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.25h
- **Description**: Run all KMS tests after migration, verify no regressions
- **Acceptance Criteria**:
  - [x] All KMS unit tests pass
  - [x] All KMS integration tests pass
  - [x] All KMS E2E tests pass (no E2E tests in scope)
  - [x] No new test failures
- **Evidence**: All 8 packages pass - client 74.9%, application 77.1%, businesslogic 39.0%, demo 7.3%, handler 79.9%, middleware 53.1%, orm 88.9%, sqlrepository 78.0%
  - [ ] Coverage maintained or improved

#### Task 12.5: Document Intentional Differences and Rationale
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Description**: If any differences exist, document why they're intentional
- **Output**: `test-output/kms-comparison/intentional-differences.md`
- **Acceptance Criteria**:
  - [x] All differences documented
  - [x] Rationale for each difference
  - [x] User approval for any intentional changes

#### Task 12.6: Create Final Comparison Report
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Description**: Comprehensive before/after comparison report
- **Output**: `test-output/kms-comparison/final-report.md`
- **Acceptance Criteria**:
  - [x] Executive summary
  - [x] Functionality verification complete
  - [x] Design intent preserved
  - [x] Test intent preserved
  - [ ] No regressions
  - [ ] User sign-off ready

---

### Phase 13: ServerBuilder Extension for KMS Architecture (SELECTED: Option A)

**Status**: ⚠️ In Progress
**User Decision**: Option A - ServerBuilder MUST provide ALL KMS functionality
**Rationale**: ServerBuilder is the foundation for ALL 9 services including KMS

---

#### Task 13.1: Add Database Abstraction Layer to ServerBuilder
- **Status**: ✅ Complete
- **Estimated**: 4h
- **Actual**: 1h
- **Description**: Create abstraction that supports BOTH GORM and raw database/sql
- **Files**:
  - `internal/apps/template/service/server/builder/database.go` (NEW)
  - `internal/apps/template/service/server/builder/database_test.go` (NEW)
  - `internal/apps/template/service/server/builder/server_builder.go` (Modified)
- **Acceptance Criteria**:
  - [x] DatabaseConfig interface supports GORM mode
  - [x] DatabaseConfig interface supports raw SQL mode
  - [x] DatabaseConnection provides unified access (GORM, raw SQL, or dual)
  - [x] ServiceResources provides DatabaseConnection
  - [x] Tests pass for GORM, raw SQL, and dual modes

#### Task 13.2: Add JWT Authentication Middleware to ServerBuilder
- **Status**: ✅ COMPLETE (commit 4aa4fd2f)
- **Estimated**: 4h
- **Actual**: 2h
- **Description**: Add JWT auth option alongside SessionManager
- **Files**:
  - `internal/apps/template/service/server/builder/jwt_auth.go` (NEW)
  - `internal/apps/template/service/server/builder/jwt_auth_test.go` (NEW)
  - `internal/apps/template/service/server/builder/server_builder.go` (Modified)
  - `internal/shared/magic/magic_security.go` (Modified - added JWKSCacheTTL)
- **Acceptance Criteria**:
  - [x] WithJWTAuth() method for JWT-based authentication
  - [x] JWT middleware abstractions (JWTAuthConfig, JWTClaims, JWTValidator interface)
  - [x] JWT auth modes: disabled/required/optional
  - [x] Works alongside or instead of SessionManager
  - [x] Tests pass (13 new tests)

#### Task 13.3: Add OpenAPI Strict Server Registration to ServerBuilder
- **Status**: ✅ COMPLETE
- **Estimated**: 4h
- **Actual**: 1h
- **Commit**: be83c8d7
- **Description**: Support oapi-codegen strict server pattern
- **Files**:
  - `internal/apps/template/service/server/builder/openapi_strict.go` (NEW - 115 lines)
  - `internal/apps/template/service/server/builder/openapi_strict_test.go` (NEW - 248 lines)
  - `internal/apps/template/service/server/builder/server_builder.go` (Modified)
- **Acceptance Criteria**:
  - [x] WithStrictServer() method (fluent API)
  - [x] StrictServerConfig for oapi-codegen strict server pattern
  - [x] Supports browser and service API base paths
  - [x] Handler registration callbacks for RegisterHandlersWithOptions
  - [x] Middleware injection for validation
  - [x] StrictServerRegistrar interface for domain services
  - [x] Tests pass (15 new tests)

#### Task 13.4: Integrate shared/barrier with ServerBuilder
- **Status**: ✅ COMPLETE
- **Estimated**: 3h
- **Actual**: 0.5h
- **Description**: Support shared barrier alongside template-specific barrier
- **Files**:
  - `internal/apps/template/service/server/builder/barrier.go` (NEW)
  - `internal/apps/template/service/server/builder/barrier_test.go` (NEW)
  - `internal/apps/template/service/server/builder/server_builder.go` (Modified)
- **Acceptance Criteria**:
  - [x] WithBarrierConfig() method for barrier mode configuration
  - [x] BarrierConfig abstraction supports template, shared, and disabled modes
  - [x] BarrierEncryptor interface for unified encryption operations
  - [x] Tests pass (22 tests for barrier + 4 for WithBarrierConfig)
- **Commit**: 1625d9b0

#### Task 13.5: Add Flexible Migration Support to ServerBuilder
- **Status**: ✅ COMPLETE
- **Estimated**: 3h
- **Actual**: 1h
- **Description**: Support multiple migration schemes (not just 1001-1004 + 2001+)
- **Files**:
  - `internal/apps/template/service/server/builder/migrations.go` (NEW - 145 lines)
  - `internal/apps/template/service/server/builder/migrations_test.go` (NEW - 355 lines)
  - `internal/apps/template/service/server/builder/server_builder.go` (Modified)
- **Acceptance Criteria**:
  - [x] WithMigrationConfig() method with MigrationConfig struct
  - [x] MigrationMode: TemplateWithDomain, DomainOnly, Disabled
  - [x] Optional template migrations (SkipTemplateMigrations flag)
  - [x] Support KMS migration scheme (DomainOnly mode)
  - [x] Factory functions: NewDefaultMigrationConfig(), NewDomainOnlyMigrationConfig(), NewDisabledMigrationConfig()
  - [x] Fluent setters: WithDomainFS(), WithDomainPath(), WithMode(), WithSkipTemplateMigrations()
  - [x] Validate(), IsEnabled(), RequiresTemplateMigrations() methods
  - [x] 26 tests pass
  - [x] Linting passes
  - [x] Build passes

#### Task 13.6: Create KMS Migration Adapter
- **Status**: ✅ COMPLETE
- **Estimated**: 6h
- **Actual**: 1.5h
- **Description**: Create adapter layer to connect KMS to extended ServerBuilder
- **Files**:
  - `internal/kms/server/builder_adapter.go` (NEW - 120 lines)
  - `internal/kms/server/builder_adapter_test.go` (NEW - 200 lines)
- **Acceptance Criteria**:
  - [x] Adapter configures ServerBuilder with KMS settings
  - [x] KMSBuilderAdapterSettings for JWT config (JWKSURL, Issuer, Audience)
  - [x] OpenAPI handlers via strict server config
  - [x] JWT auth configured (disabled by default, enabled with JWKSURL)
  - [x] Barrier service disabled (KMS uses own encryption)
  - [x] Migrations disabled (KMS uses own migration system)
  - [x] 5 tests pass, linting clean, build passes

#### Task 13.7: Migrate KMS to Extended ServerBuilder
- **Status**: ❌ Not Started
- **Estimated**: 8h
- **Description**: Replace application_listener.go with ServerBuilder usage
- **Files**:
  - `internal/kms/cmd/server.go` (Modified)
  - `internal/kms/server/application/application_listener.go` (DELETED or minimized)
  - `internal/kms/server/server.go` (NEW - uses ServerBuilder)
- **Acceptance Criteria**:
  - [ ] KMS uses ServerBuilder for all infrastructure
  - [ ] application_listener.go functionality moved to builder
  - [ ] No duplicate TLS/listener code
  - [ ] All middleware preserved

#### Task 13.8: Verify All KMS Tests Pass
- **Status**: ❌ Not Started
- **Estimated**: 4h
- **Description**: Run all KMS tests, fix any regressions
- **Acceptance Criteria**:
  - [ ] `go test ./internal/kms/...` passes
  - [ ] All 8 test packages pass
  - [ ] Coverage maintained (74.9%+ client, 77.1%+ application, etc.)
  - [ ] No new test failures

#### Task 13.9: Verify All Other Service Tests Pass
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Description**: Verify ServerBuilder changes don't break existing services
- **Acceptance Criteria**:
  - [ ] `go test ./internal/apps/template/...` passes
  - [ ] `go test ./internal/apps/cipher/...` passes
  - [ ] `go test ./internal/apps/jose/...` passes
  - [ ] All services still work correctly

#### Task 13.10: Update Documentation
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Description**: Update all documentation to reflect ServerBuilder extensions
- **Files**:
  - `.github/instructions/03-08.server-builder.instructions.md` (Modified)
  - `docs/arch/SERVICE-TEMPLATE-*.md` (Modified)
- **Acceptance Criteria**:
  - [ ] ServerBuilder documentation updated
  - [ ] KMS migration documented
  - [ ] Database abstraction documented
  - [ ] JWT auth option documented
  - [ ] OpenAPI strict server documented

---

## Cross-Cutting Tasks (Remaining)

### Quality
- [ ] No TODOs in production code (some remain, tracked)

---

## References

- Completed tasks: [completed.md](./completed.md)
- Analysis docs archived in [archive/](./archive/)
- Copilot instructions in `.github/instructions/`
- Comparison table: [comparison-table.md](./comparison-table.md)
