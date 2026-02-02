# Tasks - Unified Service-Template Migration (V7)

**Status**: 0 of 36 tasks complete (0%)
**Last Updated**: 2026-02-02

---

## Phase 0: Research & Discovery

### Task 0.1: Analyze KMS SQLRepository for GORM Migration
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: None
- **Description**: Document all KMS SQLRepository queries and map to GORM equivalents
- **Acceptance Criteria**:
  - [ ] All SQLRepository methods documented
  - [ ] GORM equivalents identified for each method
  - [ ] Migration complexity assessed (simple/moderate/complex per method)
- **Output**: test-output/v7-research/sqlrepository-analysis.md

### Task 0.2: Analyze KMS Barrier vs Template Barrier
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: None
- **Description**: Compare shared/barrier with template barrier capabilities
- **Acceptance Criteria**:
  - [ ] Feature parity documented
  - [ ] Incompatibilities identified
  - [ ] Migration path determined
- **Output**: test-output/v7-research/barrier-comparison.md

### Task 0.3: Document KMS Authentication Requirements
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None
- **Description**: Document current KMS auth and map to JWT/realm model
- **Acceptance Criteria**:
  - [ ] Current auth mechanisms documented
  - [ ] Realm structure designed
  - [ ] Token claims defined
- **Output**: test-output/v7-research/auth-requirements.md

### Task 0.4: Map KMS API to OpenAPI Spec
- **Status**: ❌ Not Started
- **Estimated**: 0.25h
- **Actual**:
- **Dependencies**: None
- **Description**: Document existing KMS API endpoints for OpenAPI generation
- **Acceptance Criteria**:
  - [ ] All endpoints catalogued
  - [ ] Request/response schemas documented
  - [ ] OpenAPI spec structure planned
- **Output**: test-output/v7-research/api-mapping.md

---

## Phase 1: Remove V6 Optional Modes

### Task 1.1: Remove DisabledDatabaseConfig
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 0.1
- **Description**: Remove database disabled mode from ServerBuilder
- **Acceptance Criteria**:
  - [ ] DisabledDatabaseConfig removed
  - [ ] RawSQLMode removed
  - [ ] DualMode removed
  - [ ] Only GORMMode remains
  - [ ] cipher-im tests pass
  - [ ] jose-ja tests pass
- **Files**:
  - `internal/apps/template/service/server/builder/database.go`

### Task 1.2: Remove DisabledBarrierConfig
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 0.2
- **Description**: Remove barrier disabled mode from ServerBuilder
- **Acceptance Criteria**:
  - [ ] DisabledBarrierConfig removed
  - [ ] SharedBarrierMode removed (if exists)
  - [ ] Only TemplateBarrier remains
  - [ ] cipher-im tests pass
  - [ ] jose-ja tests pass
- **Files**:
  - `internal/apps/template/service/server/builder/barrier.go`

### Task 1.3: Remove DisabledMigrationConfig
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: None
- **Description**: Remove migration disabled mode from ServerBuilder
- **Acceptance Criteria**:
  - [ ] DisabledMigrationConfig removed
  - [ ] DomainOnlyMode removed
  - [ ] Only TemplateWithDomainMode remains
  - [ ] cipher-im tests pass
  - [ ] jose-ja tests pass
- **Files**:
  - `internal/apps/template/service/server/builder/migrations.go`

### Task 1.4: Remove JWTAuthDisabled Mode
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Task 0.3
- **Description**: Remove JWT auth disabled mode from ServerBuilder
- **Acceptance Criteria**:
  - [ ] JWTAuthDisabled removed
  - [ ] Auth is always enabled (JWT or session)
  - [ ] cipher-im tests pass
  - [ ] jose-ja tests pass
- **Files**:
  - `internal/apps/template/service/server/builder/jwt_auth.go`

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
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Tasks 1.1-1.4
- **Description**: Full test suite for both reference implementations
- **Acceptance Criteria**:
  - [ ] All cipher-im tests pass (10 packages)
  - [ ] All jose-ja tests pass (6 packages)
  - [ ] No regressions
- **Evidence**: Test output logs

---

## Phase 2: KMS Data Migration

### Task 2.1: Create KMS GORM Models
- **Status**: ❌ Not Started
- **Estimated**: 2h
- **Actual**:
- **Dependencies**: Task 0.1, Phase 1 complete
- **Description**: Create GORM models for all KMS database tables
- **Acceptance Criteria**:
  - [ ] All SQLRepository tables have GORM models
  - [ ] Cross-DB compatible (type:text for UUIDs)
  - [ ] Proper indexes defined
  - [ ] Relationships configured
- **Files**:
  - `internal/kms/domain/models.go`

### Task 2.2: Create KMS Domain Migrations
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 2.1
- **Description**: Create golang-migrate migrations for KMS tables (2001+)
- **Acceptance Criteria**:
  - [ ] Migrations start at 2001
  - [ ] Up and down migrations for all tables
  - [ ] Compatible with template migrations (1001-1999)
- **Files**:
  - `internal/kms/repository/migrations/2001_kms_init.up.sql`
  - `internal/kms/repository/migrations/2001_kms_init.down.sql`

### Task 2.3: Create KMS GORM Repositories
- **Status**: ❌ Not Started
- **Estimated**: 3h
- **Actual**:
- **Dependencies**: Task 2.1
- **Description**: Implement GORM repositories matching SQLRepository interfaces
- **Acceptance Criteria**:
  - [ ] All SQLRepository methods have GORM equivalents
  - [ ] Transaction support via context pattern
  - [ ] Error mapping to app errors
  - [ ] Unit tests with ≥95% coverage
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

## Phase 5: KMS Barrier Migration

### Task 5.1: Analyze KMS Unseal/Seal Workflows
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 0.2
- **Description**: Document KMS unseal/seal workflows for template barrier integration
- **Acceptance Criteria**:
  - [ ] Current workflow documented
  - [ ] Template barrier integration points identified
  - [ ] Migration plan defined
- **Output**: test-output/v7-research/kms-unseal-workflow.md

### Task 5.2: Integrate Template Barrier with KMS
- **Status**: ❌ Not Started
- **Estimated**: 1.5h
- **Actual**:
- **Dependencies**: Task 5.1, Phase 2 complete
- **Description**: Replace shared/barrier with template barrier in KMS
- **Acceptance Criteria**:
  - [ ] KMS uses template barrier service
  - [ ] Key hierarchy preserved
  - [ ] Encryption/decryption works
  - [ ] Tests with ≥95% coverage
- **Files**:
  - `internal/kms/server/server.go` (update)
  - `internal/kms/service/barrier_adapter.go` (if needed)

### Task 5.3: Migrate Encryption Operations
- **Status**: ❌ Not Started
- **Estimated**: 1h
- **Actual**:
- **Dependencies**: Task 5.2
- **Description**: Update all KMS encryption operations to use template barrier
- **Acceptance Criteria**:
  - [ ] All encryption via template barrier
  - [ ] All decryption via template barrier
  - [ ] Existing data readable
  - [ ] New data encrypted correctly
- **Files**:
  - `internal/kms/service/*.go` (update encryption calls)

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
- **Evidence**: `grep -r "shared/barrier" internal/kms/` returns empty

---

## Phase 6: Integration & Testing

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
  - [ ] Mutation testing ≥95%
- **Evidence**: Test output logs, coverage report

### Task 6.2: cipher-im Regression Suite
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 1 complete
- **Description**: Verify cipher-im not affected by changes
- **Acceptance Criteria**:
  - [ ] All 10 test packages pass
  - [ ] No new failures
  - [ ] Coverage maintained
- **Evidence**: Test output logs

### Task 6.3: jose-ja Regression Suite
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 1 complete
- **Description**: Verify jose-ja not affected by changes
- **Acceptance Criteria**:
  - [ ] All 6 test packages pass
  - [ ] No new failures
  - [ ] Coverage maintained
- **Evidence**: Test output logs

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

---

## Phase 7: Documentation & Cleanup

### Task 7.1: Update server-builder.instructions.md
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 6 complete
- **Description**: Final documentation update for ServerBuilder
- **Acceptance Criteria**:
  - [ ] Remove all V6 optional mode documentation
  - [ ] Document unified mandatory architecture
  - [ ] Add migration guide section
- **Files**:
  - `.github/instructions/03-08.server-builder.instructions.md`

### Task 7.2: Update service-template.instructions.md
- **Status**: ❌ Not Started
- **Estimated**: 0.5h
- **Actual**:
- **Dependencies**: Phase 6 complete
- **Description**: Update service template documentation
- **Acceptance Criteria**:
  - [ ] Reflects unified architecture
  - [ ] All components documented as MANDATORY
  - [ ] Examples updated
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
| Phase 6 | 6 | 0 | 0% |
| Phase 7 | 4 | 0 | 0% |
| **Total** | **39** | **0** | **0%** |
