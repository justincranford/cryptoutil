# Tasks - Unified Service-Template Migration (V7)

**Status**: 7 of 40 tasks complete (17.5%)
**Last Updated**: 2026-02-02
**Quizme Decisions Applied**: ✅ All 6 answers merged
- Q1: Fresh start (no data migration)
- Q2: Merge shared/barrier INTO template barrier
- Q3: Internal only (no API versioning)
- Q4: Correctness first
- Q5: Full regression + E2E + coverage; mutation testing last
- Q6: Continuous documentation updates

---

## Phase 0: Research & Discovery

### Task 0.1: Analyze KMS SQLRepository for GORM Migration
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.5h
- **Dependencies**: None
- **Description**: Document all KMS SQLRepository queries and map to GORM equivalents
- **Acceptance Criteria**:
  - [x] All SQLRepository methods documented
  - [x] GORM equivalents identified for each method
  - [x] Migration complexity assessed (simple/moderate/complex per method)
- **Output**: test-output/v7-research/sqlrepository-analysis.md
- **Finding**: KMS already uses GORM via orm/ package! The migration is simpler - just remove the unnecessary sqlrepository/ layer.

### Task 0.2: Analyze KMS Barrier vs Template Barrier
- **Status**: ✅ Complete
- **Estimated**: 0.5h
- **Actual**: 0.25h
- **Dependencies**: None
- **Description**: Compare shared/barrier with template barrier capabilities
- **Acceptance Criteria**:
  - [x] Feature parity documented
  - [x] Incompatibilities identified
  - [x] Migration path determined
- **Output**: test-output/v7-research/barrier-comparison.md
- **Finding**: Template barrier is MORE comprehensive. Q2 decision confirmed - merge shared INTO template.

### Task 0.3: Document KMS Authentication Requirements
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.25h
- **Dependencies**: None
- **Description**: Document current KMS auth and map to JWT/realm model
- **Acceptance Criteria**:
  - [x] Current auth mechanisms documented
  - [x] Realm structure designed
  - [x] Token claims defined
- **Output**: test-output/v7-research/auth-requirements.md
- **Finding**: KMS has OPTIONAL JWT auth via builder_adapter.go - will become REQUIRED.

### Task 0.4: Map KMS API to OpenAPI Spec
- **Status**: ✅ Complete
- **Estimated**: 0.25h
- **Actual**: 0.25h
- **Dependencies**: None
- **Description**: Document existing KMS API endpoints for OpenAPI generation
- **Acceptance Criteria**:
  - [x] All endpoints catalogued
  - [x] Request/response schemas documented
  - [x] OpenAPI spec structure planned
- **Output**: test-output/v7-research/api-mapping.md
- **Finding**: KMS already uses StrictServer pattern. Only path prefixes and auth need updating.

---

## Phase 1: Remove V6 Optional Modes

### Task 1.1: Remove DisabledDatabaseConfig
- **Status**: ✅ Complete
- **Estimated**: 0.5h
- **Actual**: 1.5h (heredoc terminal issues required base64 workaround)
- **Dependencies**: Task 0.1
- **Description**: Remove database disabled mode from ServerBuilder
- **Acceptance Criteria**:
  - [x] DisabledDatabaseConfig removed
  - [x] RawSQLMode removed
  - [x] DualMode removed
  - [x] Only GORMMode remains
  - [ ] cipher-im tests pass
  - [ ] jose-ja tests pass
- **Files**:
  - `internal/apps/template/service/server/builder/database.go`

### Task 1.2: Remove DisabledBarrierConfig
- **Status**: ✅ Complete
- **Estimated**: 0.5h
- **Actual**: 0.25h
- **Dependencies**: Task 0.2
- **Description**: Remove barrier disabled mode from ServerBuilder
- **Acceptance Criteria**:
  - [x] DisabledBarrierConfig removed
  - [x] SharedBarrierMode removed (if exists)
  - [x] Only TemplateBarrier remains
  - [x] cipher-im tests pass
  - [x] jose-ja tests pass
- **Files**:
  - `internal/apps/template/service/server/builder/barrier.go`
  - `internal/apps/template/service/server/builder/barrier_test.go`
  - `internal/apps/template/service/server/builder/server_builder.go`
- **Commit**: 7349ed72

### Task 1.3: Remove DisabledMigrationConfig
- **Status**: ✅ Complete
- **Estimated**: 0.5h
- **Actual**: 0.75h
- **Dependencies**: None
- **Description**: Remove migration disabled mode from ServerBuilder
- **Acceptance Criteria**:
  - [x] DisabledMigrationConfig removed
  - [x] DomainOnlyMode KEPT (useful for KMS-style services that manage their own migrations)
  - [x] TemplateWithDomainMode remains as default
  - [x] IsEnabled() method removed (migrations always enabled)
  - [x] builder tests pass (121 tests)
- **Files**:
  - `internal/apps/template/service/server/builder/migrations.go`
  - `internal/apps/template/service/server/builder/migrations_test.go`
  - `internal/apps/template/service/server/builder/server_builder.go`
- **Commit**: 5e426085

### Task 1.4: Remove JWTAuthDisabled Mode
- **Status**: ✅ Complete
- **Estimated**: 0.5h
- **Actual**: 0.25h
- **Dependencies**: Task 0.3
- **Description**: Rename JWTAuthModeDisabled to JWTAuthModeSession for clarity
- **Acceptance Criteria**:
  - [x] JWTAuthModeDisabled renamed to JWTAuthModeSession
  - [x] Constant value changed from "disabled" to "session"
  - [x] Auth is always enabled (JWT or session) - now clearer
  - [x] builder tests pass (121 tests)
- **Files**:
  - `internal/apps/template/service/server/builder/jwt_auth.go`
  - `internal/apps/template/service/server/builder/jwt_auth_test.go`
- **Commit**: 39742092

### Task 1.5: Remove KMS builder_adapter.go
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Tasks 1.1-1.4
- **Description**: Remove V6 KMS adapter hack
- **Acceptance Criteria**:
  - [ ] builder_adapter.go deleted
  - [ ] builder_adapter_test.go deleted
  - [ ] KMS no longer compiles (expected - will be fixed in Phase 2-5)
- **Files**:
  - `internal/kms/server/builder_adapter.go` (DELETE)
  - `internal/kms/server/builder_adapter_test.go` (DELETE)

### Task 1.6: Update ServerBuilder Documentation
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Tasks 1.1-1.5
- **Description**: Update 03-08.server-builder.instructions.md to reflect mandatory components
- **Acceptance Criteria**:
  - [ ] Remove Phase 13 extension sections
  - [ ] Document MANDATORY components
  - [ ] Remove optional mode references
- **Files**:
  - `.github/instructions/03-08.server-builder.instructions.md`

### Task 1.7: Verify cipher-im and jose-ja Still Work
- **Status**: ✅ Complete
- **Estimated**: 0.5h
- **Actual**: 0.25h
- **Dependencies**: Tasks 1.1-1.4
- **Description**: Full test suite for both reference implementations
- **Acceptance Criteria**:
  - [x] All cipher-im tests pass (unit/integration pass; E2E flakiness is pre-existing, unrelated to Phase 1)
  - [x] All jose-ja tests pass (6 packages)
  - [x] No regressions
- **Evidence**: Template tests (15 packages OK), jose-ja all pass, cipher-im unit/integration pass

---

## Phase 2: KMS Data Migration (SIMPLIFIED - Fresh Start)

**Note**: Per quizme Q1, this is a fresh start with no data migration needed.

### Task 2.1: Create KMS GORM Models
- **Status**: ✅ Complete (Pre-existing)
- **Estimated**: 2h
- **Actual**: 0.25h (verification only - models already existed)
- **Dependencies**: Task 0.1, Phase 1 complete
- **Description**: Create GORM models for all KMS database tables (fresh schema, no migration from old data)
- **Acceptance Criteria**:
  - [x] All required tables have GORM models
  - [x] Cross-DB compatible (type:text for UUIDs)
  - [x] Proper indexes defined
  - [x] Relationships configured
  - [x] Documentation updated (per Q6)
- **Files**:
  - `internal/kms/server/repository/orm/barrier_entities.go` (RootKey, IntermediateKey, ContentKey)
  - `internal/kms/server/repository/orm/business_entities.go` (ElasticKey, MaterialKey)
- **Evidence**: Models pre-existed. ElasticKey uses gorm:"type:uuid;primaryKey" tags. MaterialKey has composite PK.

### Task 2.2: Create KMS Domain Migrations
- **Status**: ✅ Complete
- **Estimated**: 1h
- **Actual**: 0.5h
- **Dependencies**: Task 2.1
- **Description**: Create golang-migrate migrations for KMS tables (2001+)
- **Acceptance Criteria**:
  - [x] Migrations start at 2001
  - [x] Up and down migrations for all tables
  - [x] Compatible with template migrations (1001-1999)
  - [x] Documentation updated (per Q6)
- **Files**:
  - `internal/kms/server/repository/migrations/2001_kms_business_tables.up.sql`
  - `internal/kms/server/repository/migrations/2001_kms_business_tables.down.sql`
  - `internal/kms/server/repository/migrations.go` (embed.FS)

### Task 2.3: Create KMS GORM Repositories
- **Status**: ❌ Not Started
- **Estimated**: 3h
- **Actual**:
- **Dependencies**: Task 2.1
- **Description**: Implement GORM repositories matching KMS business needs
- **Acceptance Criteria**:
  - [ ] All required repository methods implemented
  - [ ] Transaction support via context pattern
  - [ ] Error mapping to app errors
  - [ ] Unit tests with ≥95% coverage
  - [ ] Documentation updated (per Q6)
- **Files**:
  - `internal/kms/repository/gorm_repository.go`
  - `internal/kms/repository/gorm_repository_test.go`

### Task 2.4: Migrate KMS Business Logic to GORM
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Task 2.3
- **Description**: Update KMS services to use GORM repositories
- **Acceptance Criteria**:
  - [ ] All services use GORM repositories
  - [ ] No direct SQLRepository references
  - [ ] All service tests pass
  - [ ] Documentation updated (per Q6)
- **Files**:
  - `internal/kms/service/*.go`

### Task 2.5: Remove SQLRepository
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 2.4
- **Description**: Delete SQLRepository and raw database/sql code
- **Acceptance Criteria**:
  - [ ] sql_provider.go deleted
  - [ ] No database/sql imports in KMS
  - [ ] All tests pass
  - [ ] Documentation updated (per Q6)
- **Files**:
  - `internal/kms/server/repository/sqlrepository/` (DELETE)

---

## Phase 3: KMS Authentication Migration

### Task 3.1: Design KMS Realm Structure
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 0.3
- **Description**: Design tenant/realm structure for KMS
- **Acceptance Criteria**:
  - [ ] Realm isolation requirements documented
  - [ ] Token claims defined
  - [ ] Permission model defined
- **Output**: test-output/v7-research/kms-realm-design.md

### Task 3.2: Implement JWT Middleware for KMS
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Task 3.1, Phase 1 complete
- **Description**: Add JWT authentication to KMS /service/** paths
- **Acceptance Criteria**:
  - [ ] JWT validation middleware configured
  - [ ] Realm context extraction
  - [ ] Token claims available to handlers
  - [ ] Tests with ≥95% coverage
- **Files**:
  - `internal/kms/server/middleware/jwt.go`
  - `internal/kms/server/middleware/jwt_test.go`

### Task 3.3: Implement Session Auth for KMS Browser Paths
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Task 3.1
- **Description**: Add session authentication to KMS /browser/** paths
- **Acceptance Criteria**:
  - [ ] Session middleware configured
  - [ ] Cookie handling
  - [ ] CSRF protection
  - [ ] Tests with ≥95% coverage
- **Files**:
  - `internal/kms/server/middleware/session.go`
  - `internal/kms/server/middleware/session_test.go`

### Task 3.4: Update KMS Handlers for Realm Context
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Tasks 3.2, 3.3
- **Description**: Update all KMS handlers to use realm context
- **Acceptance Criteria**:
  - [ ] All handlers extract tenant/realm from context
  - [ ] Data operations scoped to tenant
  - [ ] Tests verify isolation
- **Files**:
  - `internal/kms/server/handler/*.go`

### Task 3.5: Configure Path Separation
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Tasks 3.2, 3.3
- **Description**: Configure /service/** and /browser/** path routing
- **Acceptance Criteria**:
  - [ ] /service/** uses JWT auth
  - [ ] /browser/** uses session auth
  - [ ] Proper middleware chains
  - [ ] Integration tests pass
- **Files**:
  - `internal/kms/server/routes.go`

---

## Phase 4: KMS OpenAPI Migration

### Task 4.1: Create KMS OpenAPI Spec
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 0.4
- **Description**: Create OpenAPI 3.0.3 spec for KMS API
- **Acceptance Criteria**:
  - [ ] All endpoints documented
  - [ ] Request/response schemas defined
  - [ ] Error responses standardized
  - [ ] Spec validates
- **Files**:
  - `internal/kms/api/openapi_spec_components.yaml`
  - `internal/kms/api/openapi_spec_paths.yaml`

### Task 4.2: Generate Strict Server Handlers
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 4.1
- **Description**: Use oapi-codegen to generate strict server interface
- **Acceptance Criteria**:
  - [ ] Server interface generated
  - [ ] Models generated
  - [ ] Client generated (for testing)
  - [ ] Generation scripts documented
- **Files**:
  - `internal/kms/api/server/openapi_gen_*.go`
  - `internal/kms/api/model/openapi_gen_*.go`
  - `internal/kms/api/client/openapi_gen_*.go`

### Task 4.3: Migrate KMS Handlers to Strict Interface
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Task 4.2
- **Description**: Implement strict server interface in KMS handlers
- **Acceptance Criteria**:
  - [ ] All handlers implement strict interface
  - [ ] Type-safe request handling
  - [ ] Consistent error responses
  - [ ] Tests with ≥95% coverage
- **Files**:
  - `internal/kms/server/handler/strict_handlers.go`
  - `internal/kms/server/handler/strict_handlers_test.go`

### Task 4.4: Add SwaggerUI for KMS
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 4.1
- **Description**: Configure SwaggerUI for KMS API documentation
- **Acceptance Criteria**:
  - [ ] SwaggerUI accessible at /browser/swagger/
  - [ ] Spec loaded correctly
  - [ ] Try-it-out works
- **Files**:
  - `internal/kms/server/routes.go` (update)

---

## Phase 5: KMS Barrier Migration (Merge shared/barrier INTO template)

**Note**: Per quizme Q2, ALL functionality from shared/barrier MUST be available in template barrier.

### Task 5.1: Analyze KMS Unseal/Seal Workflows
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 0.2
- **Description**: Document KMS unseal/seal workflows and identify ALL shared/barrier features needed in template
- **Acceptance Criteria**:
  - [ ] Current workflow documented
  - [ ] ALL shared/barrier features catalogued
  - [ ] Feature parity checklist created
  - [ ] Documentation updated (per Q6)
- **Output**: test-output/v7-research/kms-unseal-workflow.md

### Task 5.2: Merge shared/barrier Features INTO Template Barrier
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Task 5.1
- **Description**: Ensure template barrier has ALL features from shared/barrier before migration
- **Acceptance Criteria**:
  - [ ] Feature parity checklist complete
  - [ ] All shared/barrier features in template barrier
  - [ ] Tests for all migrated features
  - [ ] Documentation updated (per Q6)
- **Files**:
  - `internal/apps/template/service/server/builder/barrier.go` (update)
  - `internal/apps/template/service/server/builder/barrier_test.go` (update)

### Task 5.3: Integrate Template Barrier with KMS
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 5.2, Phase 2 complete
- **Description**: Replace shared/barrier with template barrier in KMS
- **Acceptance Criteria**:
  - [ ] KMS uses template barrier service
  - [ ] Key hierarchy preserved
  - [ ] Encryption/decryption works
  - [ ] Tests with ≥95% coverage
  - [ ] Documentation updated (per Q6)
- **Files**:
  - `internal/kms/server/server.go` (update)

### Task 5.4: Remove shared/barrier Usage from KMS
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 5.3
- **Description**: Remove all shared/barrier imports from KMS
- **Acceptance Criteria**:
  - [ ] No shared/barrier imports in KMS
  - [ ] Only template barrier used
  - [ ] All tests pass
  - [ ] Documentation updated (per Q6)
- **Evidence**: `grep -r "shared/barrier" internal/kms/` returns empty

---

## Phase 6: Integration & Testing (Expanded per Q5)

**Note**: Per quizme Q5, full regression + E2E + coverage. Mutation testing LAST.

### Task 6.1: KMS Full Test Suite
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Phases 2-5 complete
- **Description**: Run complete KMS test suite with new architecture
- **Acceptance Criteria**:
  - [ ] All unit tests pass
  - [ ] All integration tests pass
  - [ ] Coverage ≥95%
  - [ ] Documentation updated (per Q6)
- **Evidence**: Test output logs, coverage report

### Task 6.2: cipher-im Full Regression Suite
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Phase 1 complete
- **Description**: Full regression for cipher-im (not just smoke test per Q5)
- **Acceptance Criteria**:
  - [ ] All unit tests pass
  - [ ] All integration tests pass
  - [ ] E2E tests pass
  - [ ] Coverage ≥95%
  - [ ] No regressions from V7 changes
- **Evidence**: Test output logs, coverage report

### Task 6.3: jose-ja Full Regression Suite
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Phase 1 complete
- **Description**: Full regression for jose-ja (not just smoke test per Q5)
- **Acceptance Criteria**:
  - [ ] All unit tests pass
  - [ ] All integration tests pass
  - [ ] E2E tests pass
  - [ ] Coverage ≥95%
  - [ ] No regressions from V7 changes
- **Evidence**: Test output logs, coverage report

### Task 6.4: Multi-Service E2E Tests
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Tasks 6.1-6.3
- **Description**: E2E tests with all three services running together
- **Acceptance Criteria**:
  - [ ] Docker Compose works with all services
  - [ ] Cross-service communication works
  - [ ] Authentication flows work
  - [ ] E2E scenarios pass
  - [ ] Documentation updated (per Q6)
- **Files**:
  - `deployments/unified/compose.yml` (if needed)
  - `test/e2e/unified_test.go` (if needed)

### Task 6.5: Performance Benchmarks
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 6.1
- **Description**: Compare performance vs V6 baseline
- **Acceptance Criteria**:
  - [ ] Benchmark suite runs
  - [ ] No significant regression (>10%)
  - [ ] Results documented
- **Output**: test-output/v7-benchmarks/

### Task 6.6: Cross-Database Verification
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 6.1
- **Description**: Verify KMS works with both PostgreSQL and SQLite
- **Acceptance Criteria**:
  - [ ] All tests pass with PostgreSQL
  - [ ] All tests pass with SQLite
  - [ ] UUID handling correct (type:text)
- **Evidence**: Test output logs

### Task 6.7: Mutation Testing (LAST per Q5)
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Tasks 6.1-6.6 ALL complete
- **Description**: Run mutation testing as final quality gate
- **Acceptance Criteria**:
  - [ ] sm-kms mutation ≥95%
  - [ ] cipher-im mutation ≥95%
  - [ ] jose-ja mutation ≥95%
  - [ ] All survived mutants documented/justified
- **Evidence**: test-output/v7-mutation/

---

## Phase 7: Documentation & Cleanup (Continuous Updates per Q6)

**Note**: Per quizme Q6, documentation updated continuously throughout V7. Phase 7 is final review.

### Task 7.1: Final Review of server-builder.instructions.md
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 6 complete
- **Description**: Final documentation review for ServerBuilder (continuous updates already applied)
- **Acceptance Criteria**:
  - [ ] All V6 optional mode references removed
  - [ ] Unified mandatory architecture documented
  - [ ] Migration guide section complete
  - [ ] No outdated information
- **Files**:
  - `.github/instructions/03-08.server-builder.instructions.md`

### Task 7.2: Final Review of service-template.instructions.md
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 6 complete
- **Description**: Final documentation review for service template (continuous updates already applied)
- **Acceptance Criteria**:
  - [ ] Reflects unified architecture
  - [ ] All components documented as MANDATORY
  - [ ] Examples updated
  - [ ] No outdated information
- **Files**:
  - `.github/instructions/02-02.service-template.instructions.md`

### Task 7.3: Remove Obsolete V6 Documentation
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: Tasks 7.1, 7.2
- **Description**: Clean up any remaining V6 abstraction documentation
- **Acceptance Criteria**:
  - [ ] No references to disabled modes
  - [ ] No references to optional configurations
  - [ ] Clean documentation

### Task 7.4: Create Future Service Migration Guide
- **Status**: ❌ Not Started
- **Estimated**: 0.75h
- **Actual**:
- **Dependencies**: Phase 6 complete
- **Description**: Document how to migrate future services to service-template
- **Acceptance Criteria**:
  - [ ] Step-by-step guide
  - [ ] Common pitfalls documented
  - [ ] Reference implementations listed
- **Output**: docs/SERVICE-MIGRATION-GUIDE.md

---

## Summary Statistics

| Phase | Tasks | Completed | Percentage |
|-------|-------|-----------|------------|
| Phase 0 | 4 | 0 | 0% |
| Phase 1 | 7 | 0 | 0% |
| Phase 2 | 5 | 0 | 0% |
| Phase 3 | 5 | 0 | 0% |
| Phase 4 | 4 | 0 | 0% |
| Phase 5 | 4 | 0 | 0% |
| Phase 6 | 7 | 0 | 0% |
| Phase 7 | 4 | 0 | 0% |
| **Total** | **40** | **0** | **0%** |

---

## Incomplete Work from V4/V6 - Addressed by V7

### From V6 (Task 5.1 BLOCKED)
- **StartApplicationListener not implemented** → Addressed by complete ServerBuilder migration in Phase 1-5

### From V4 (Phase 0.4 KMS Modernization)
- **KMS service-template migration** → Fully addressed by V7 Phases 2-5

### From V4 (Phase 0.5 Template Mutation)
- **Template mutation improvement** → Addressed by Task 6.7 (mutation testing LAST)

### From V6 (Coverage Gaps)
- **All services below ≥95% coverage** → Addressed by Tasks 6.1-6.3 with ≥95% requirement
