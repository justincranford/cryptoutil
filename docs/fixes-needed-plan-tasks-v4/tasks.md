# Tasks - Remaining Work (V4)

**Status**: 0 of 111 tasks complete (0%)
**Last Updated**: 2026-01-26
**Priority Order**: Template → Cipher-IM → JOSE-JA → Shared → Infra → KMS → Compose → Mutation CI/CD → Race Testing

**Previous Version**: docs/fixes-needed-plan-tasks-v3/ (47/115 tasks complete, 40.9%)

**User Feedback**: Phase ordering updated to prioritize template quality first, then services in architectural conformance order (cipher-im before JOSE-JA), KMS last to leverage validated patterns.

## Phase 1: Service-Template Coverage (HIGHEST PRIORITY)

**Objective**: Bring service-template to ≥95% coverage (reference implementation)
**Status**: ⏳ NOT STARTED
**Current**: 82.5% coverage (-12.5% below minimum)

### Task 1.1: Add Tests for Template Server/Application Lifecycle

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: None
**Priority**: CRITICAL

**Description**: Add tests for template server/application lifecycle methods (StartBasic, Shutdown, InitializeServicesOnCore).

**Acceptance Criteria**:
- [ ] 1.1.1: Add unit tests for StartBasic()
- [ ] 1.1.2: Add unit tests for Shutdown()
- [ ] 1.1.3: Add unit tests for InitializeServicesOnCore()
- [ ] 1.1.4: Add error path tests
- [ ] 1.1.5: Verify coverage ≥95% for application package
- [ ] 1.1.6: All tests pass
- [ ] 1.1.7: Commit: "test(template): add application lifecycle tests"

**Files**:
- internal/apps/template/service/server/application/application_test.go (new)

---

### Task 1.2: Add Tests for Template Server Builder

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 1.1 complete
**Priority**: CRITICAL

**Description**: Add tests for service-template server builder pattern.

**Acceptance Criteria**:
- [ ] 1.2.1: Add unit tests for NewServerBuilder()
- [ ] 1.2.2: Add tests for WithDomainMigrations()
- [ ] 1.2.3: Add tests for WithPublicRouteRegistration()
- [ ] 1.2.4: Add tests for Build()
- [ ] 1.2.5: Add integration tests for full builder flow
- [ ] 1.2.6: Verify coverage ≥95% for builder package
- [ ] 1.2.7: Commit: "test(template): add server builder tests"

**Files**:
- internal/apps/template/service/server/builder/server_builder_test.go (new)

---

### Task 1.3: Add Tests for Template Application Listeners

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 1.2 complete
**Priority**: HIGH

**Description**: Add tests for template application listeners (dual HTTPS servers).

**Acceptance Criteria**:
- [ ] 1.3.1: Add tests for listener initialization
- [ ] 1.3.2: Add tests for listener start/stop
- [ ] 1.3.3: Add tests for listener error handling
- [ ] 1.3.4: Verify coverage ≥95% for listener package
- [ ] 1.3.5: Commit: "test(template): add application listener tests"

**Files**:
- internal/apps/template/service/server/listener/*_test.go (new)

---

### Task 1.4: Add Tests for Template Service Client

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 1.3 complete
**Priority**: HIGH

**Description**: Add tests for template service/client (authentication client).

**Acceptance Criteria**:
- [ ] 1.4.1: Add tests for client initialization
- [ ] 1.4.2: Add tests for authentication methods
- [ ] 1.4.3: Add tests for error handling
- [ ] 1.4.4: Verify coverage ≥95% for client package
- [ ] 1.4.5: Commit: "test(template): add service client tests"

**Files**:
- internal/apps/template/service/client/*_test.go (new)

---

### Task 1.5: Add Tests for Template Config Parsing

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 1.4 complete
**Priority**: HIGH

**Description**: Add tests for template config parsing and validation.

**Acceptance Criteria**:
- [ ] 1.5.1: Add tests for config loading
- [ ] 1.5.2: Add tests for validation rules
- [ ] 1.5.3: Add tests for error cases
- [ ] 1.5.4: Verify coverage ≥95% for config package
- [ ] 1.5.5: Commit: "test(template): add config parsing tests"

**Files**:
- internal/apps/template/service/config/*_test.go (add)

---

### Task 1.6: Add Integration Tests for Dual HTTPS Servers

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 1.5 complete
**Priority**: HIGH

**Description**: Add integration tests for template dual HTTPS servers (public + admin).

**Acceptance Criteria**:
- [ ] 1.6.1: Add tests for public server endpoints
- [ ] 1.6.2: Add tests for admin server endpoints
- [ ] 1.6.3: Add tests for health checks
- [ ] 1.6.4: Add tests for graceful shutdown
- [ ] 1.6.5: Verify both servers accessible
- [ ] 1.6.6: Commit: "test(template): add dual HTTPS server integration tests"

**Files**:
- internal/apps/template/service/server/integration_test.go (new)

---

### Task 1.7: Add Tests for Template Middleware Stack

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 1.6 complete
**Priority**: HIGH

**Description**: Add tests for template middleware stack (/service vs /browser paths).

**Acceptance Criteria**:
- [ ] 1.7.1: Add tests for /service/** middleware (IP allowlist, rate limiting)
- [ ] 1.7.2: Add tests for /browser/** middleware (CSRF, CORS, CSP)
- [ ] 1.7.3: Add tests for mutual exclusivity enforcement
- [ ] 1.7.4: Verify coverage ≥95% for middleware packages
- [ ] 1.7.5: Commit: "test(template): add middleware stack tests"

**Files**:
- internal/apps/template/service/server/middleware/*_test.go (add)

---

### Task 1.8: Verify Template ≥95% Coverage Achieved

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 1.7 complete
**Priority**: CRITICAL

**Description**: Run coverage analysis and verify service-template achieves ≥95% coverage minimum (≥98% ideal).

**Acceptance Criteria**:
- [ ] 1.8.1: Run coverage: `go test -cover ./internal/apps/template/...`
- [ ] 1.8.2: Verify ≥95% coverage (≥98% ideal)
- [ ] 1.8.3: Generate HTML report for gap analysis
- [ ] 1.8.4: Document actual coverage achieved
- [ ] 1.8.5: Update plan.md with Phase 1 completion
- [ ] 1.8.6: Commit: "docs(v4): Phase 1 complete - template ≥95% coverage"

**Files**:
- docs/fixes-needed-plan-tasks-v4/plan.md (update)
- docs/fixes-needed-plan-tasks-v4/tasks.md (update)

---

## Phase 2: Cipher-IM Coverage + Mutation (BEFORE JOSE-JA)

**Objective**: Complete cipher-im coverage AND unblock mutation testing
**Status**: ⏳ NOT STARTED
**Current**: 78.9% coverage (-16.1%), mutation BLOCKED

**User Decision**: "cipher-im is closer to architecture conformance. it has less issues... should be worked on before jose-ja"

### Task 2.1: Add Tests for Cipher-IM Message Repository

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Phase 1 complete
**Priority**: HIGH

**Description**: Add tests for cipher-im message repository edge cases.

**Acceptance Criteria**:
- [ ] 2.1.1: Add tests for Create edge cases
- [ ] 2.1.2: Add tests for GetByID error paths
- [ ] 2.1.3: Add tests for List pagination
- [ ] 2.1.4: Add tests for database errors
- [ ] 2.1.5: Verify coverage improvement
- [ ] 2.1.6: Commit: "test(cipher-im): add message repository tests"

**Files**:
- internal/apps/cipher/im/repository/message_repository_test.go (add)

---

### Task 2.2: Add Tests for Cipher-IM Message Service

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 2.1 complete
**Priority**: HIGH

**Description**: Add tests for cipher-im message service business logic.

**Acceptance Criteria**:
- [ ] 2.2.1: Add tests for SendMessage
- [ ] 2.2.2: Add tests for encryption workflows
- [ ] 2.2.3: Add tests for validation rules
- [ ] 2.2.4: Add tests for error handling
- [ ] 2.2.5: Verify coverage improvement
- [ ] 2.2.6: Commit: "test(cipher-im): add message service tests"

**Files**:
- internal/apps/cipher/im/service/message_service_test.go (add)

---

### Task 2.3: Add Tests for Cipher-IM Server Configuration

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 2.2 complete
**Priority**: MEDIUM

**Description**: Add tests for cipher-im server configuration.

**Acceptance Criteria**:
- [ ] 2.3.1: Add tests for config loading
- [ ] 2.3.2: Add tests for validation
- [ ] 2.3.3: Add tests for defaults
- [ ] 2.3.4: Verify coverage improvement
- [ ] 2.3.5: Commit: "test(cipher-im): add server config tests"

**Files**:
- internal/apps/cipher/im/config/*_test.go (add)

---

### Task 2.4: Add Integration Tests for Cipher-IM Dual HTTPS

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 2.3 complete
**Priority**: HIGH

**Description**: Add integration tests for cipher-im dual HTTPS servers.

**Acceptance Criteria**:
- [ ] 2.4.1: Add E2E tests for message sending
- [ ] 2.4.2: Add tests for dual path verification (/service vs /browser)
- [ ] 2.4.3: Add tests for health checks
- [ ] 2.4.4: Verify all endpoints functional
- [ ] 2.4.5: Commit: "test(cipher-im): add dual HTTPS integration tests"

**Files**:
- internal/apps/cipher/im/integration_test.go (new)

---

### Task 2.5: Verify Cipher-IM ≥95% Coverage

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 2.4 complete
**Priority**: CRITICAL

**Description**: Verify cipher-im achieves ≥95% coverage.

**Acceptance Criteria**:
- [ ] 2.5.1: Run coverage: `go test -cover ./internal/apps/cipher/im/...`
- [ ] 2.5.2: Verify ≥95% coverage
- [ ] 2.5.3: Generate HTML report
- [ ] 2.5.4: Document actual coverage
- [ ] 2.5.5: Commit: "docs(cipher-im): ≥95% coverage achieved"

**Files**:
- docs/fixes-needed-plan-tasks-v4/tasks.md (update)

---

### Task 2.6: Fix Cipher-IM Docker Infrastructure

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 2.5 complete
**Priority**: CRITICAL

**Description**: Fix Docker compose issues blocking cipher-im mutation testing.

**Acceptance Criteria**:
- [ ] 2.6.1: Resolve OTEL HTTP/gRPC mismatch
- [ ] 2.6.2: Fix E2E tag bypass issue
- [ ] 2.6.3: Verify health checks pass
- [ ] 2.6.4: Run Docker Compose
- [ ] 2.6.5: All services healthy
- [ ] 2.6.6: Commit: "fix(cipher-im): unblock Docker for mutation testing"

**Files**:
- deployments/cipher/compose.yml (fix)
- configs/cipher/ (update)

---

### Task 2.7: Run Gremlins on Cipher-IM for ≥98% Efficacy

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 2.6 complete
**Priority**: CRITICAL

**Description**: Run gremlins, analyze mutations, kill for ≥98% efficacy.

**Acceptance Criteria**:
- [ ] 2.7.1: Run: `gremlins unleash ./internal/apps/cipher/im/`
- [ ] 2.7.2: Analyze lived mutations
- [ ] 2.7.3: Write targeted tests
- [ ] 2.7.4: Re-run gremlins
- [ ] 2.7.5: Verify ≥98% efficacy
- [ ] 2.7.6: Commit: "test(cipher-im): 98% mutation efficacy achieved"

**Files**:
- internal/apps/cipher/im/*_test.go (add)

---

## Phase 3: JOSE-JA Migration + Coverage (AFTER Cipher-IM)

**Objective**: Complete JOSE-JA template migration AND improve coverage to ≥95%
**Status**: ⏳ NOT STARTED
**Current**: 92.5% coverage (-2.5%), 97.20% mutation, partial template migration

**User Concern**: "extremely concerned with all of the architectural conformance... issues you found for jose-ja"

**Critical Issues**: Multi-tenancy, SQLite, ServerBuilder, merged migrations, registration, Docker config, browser APIs (7 pending)

### Task 3.1: Add createMaterialJWK Error Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: None
**Priority**: HIGH

**Description**: Create comprehensive comparison table analyzing kms, service-template, cipher-im, and jose-ja implementations to identify code duplication, inconsistencies, and opportunities for service-template extraction.

**Acceptance Criteria**:
- [ ] 0.1.1: Read all four service implementations
  - internal/kms/server/ (reference KMS implementation)
  - internal/apps/template/service/ (extracted template)
  - internal/apps/cipher/im/service/ (cipher-im service)
  - internal/apps/jose/ja/service/ (jose-ja service)
- [ ] 0.1.2: Create comparison table with columns:
  - Component (Server struct, Config, Handlers, Middleware, TLS setup, etc.)
  - KMS implementation (file location, pattern used)
  - Service-template implementation (file location, pattern used)
  - Cipher-IM implementation (file location, pattern used)
  - JOSE-JA implementation (file location, pattern used)
  - Duplication analysis (identical, similar, different)
  - Reusability recommendation (extract to template, keep service-specific, etc.)
- [ ] 0.1.3: Document findings in research.md
- [ ] 0.1.4: Identify top 10 duplication candidates for extraction
- [ ] 0.1.5: Estimate effort to extract each candidate

**Files**:
- docs/fixes-needed-plan-tasks-v4/research.md (new)

---

### Task 0.2: Mutation Efficacy Standards Clarification

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent

**Dependencies**: None
**Priority**: MEDIUM

**Description**: Clarify and document the distinction between 98% IDEAL target and 85% MINIMUM acceptable mutation efficacy standards in plan.md quality gates section.

**Acceptance Criteria**:
- [ ] 0.2.1: Document 98% as IDEAL target (Template ✅ 98.91%, JOSE-JA ✅ 97.20%)
- [ ] 0.2.2: Document 85% as MINIMUM acceptable (with documented blockers only)
- [ ] 0.2.3: Update plan.md quality gates section with clear distinction
- [ ] 0.2.4: Add examples of acceptable blockers (test unreachable code, etc.)
- [ ] 0.2.5: Commit: "docs(plan): clarify mutation efficacy 98% ideal vs 85% minimum"

**Files**:
- docs/fixes-needed-plan-tasks-v4/plan.md (update)

---

### Task 0.3: CI/CD Mutation Workflow Research

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent

**Dependencies**: None
**Priority**: MEDIUM

**Description**: Research and document Linux-based CI/CD mutation testing execution requirements, timeout configurations, and artifact collection patterns.

**Acceptance Criteria**:
- [ ] 0.3.1: Review existing .github/workflows/ci-mutation.yml
- [ ] 0.3.2: Document Linux execution requirements
- [ ] 0.3.3: Document timeout configuration (per package recommended)
- [ ] 0.3.4: Document artifact collection patterns
- [ ] 0.3.5: Create CI/CD execution checklist in research.md
- [ ] 0.3.6: Commit: "docs(research): CI/CD mutation testing patterns"

**Files**:
- docs/fixes-needed-plan-tasks-v4/research.md (update)
- .github/workflows/ci-mutation.yml (reference)

---

## Phase 5: Infrastructure Code Coverage (Crypto + Barrier Services)

**Objective**: Bring infrastructure packages to ≥98% coverage
**Status**: ⏳ NOT STARTED
**Current**: Multiple packages below minimum

### Task 5.1: Add Barrier Intermediate Key Service Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Phase 4 complete
**Priority**: HIGH

**Description**: Add tests for barrier intermediate key service.

**Acceptance Criteria**:
- [ ] 5.1.1: Add tests for key generation
- [ ] 5.1.2: Add tests for key rotation
- [ ] 5.1.3: Add tests for error paths
- [ ] 5.1.4: Verify ≥98% coverage
- [ ] 5.1.5: Commit: "test(barrier): add intermediate key service tests"

**Files**:
- internal/shared/barrier/intermediate_key_service_test.go (add)

---

[Remaining tasks 5.2-12.35 to be added in next commit - file too complex for single operation]

---

## Cross-Cutting Tasks

### Documentation
- [ ] Update README.md with mutation testing instructions
- [ ] Update DEV-SETUP.md with workflow setup
- [ ] Create research.md with comparison table
- [ ] Update completed.md as phases finish

### Testing
- [ ] All tests pass (`runTests`)
- [ ] Coverage ≥95% production, ≥98% infrastructure
- [ ] Mutation efficacy ≥98% ideal (ALL services)
- [ ] Race detector clean on Linux

### Quality
- [ ] Linting passes (`golangci-lint run`)
- [ ] No new TODOs without tracking
- [ ] Conventional commits enforced


**Objective**: Achieve 95% coverage for jose/service (currently 87.3%, gap: 7.7%)
**Status**: ⏳ NOT STARTED

### Task 1.1: Add createMaterialJWK Error Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Phase 2 complete
**Priority**: HIGH

**Description**: Add error path tests for createMaterialJWK function.

**Acceptance Criteria**:
- [ ] 3.1.1: Analyze createMaterialJWK error paths
- [ ] 3.1.2: Write tests for invalid parameters
- [ ] 3.1.3: Write tests for JWKGen errors
- [ ] 3.1.4: Write tests for database errors
- [ ] 3.1.5: Verify coverage improvement
- [ ] 3.1.6: Commit: "test(jose): add createMaterialJWK error tests"

**Files**:
- internal/apps/jose/ja/service/service_test.go (add)

---

### Task 3.2: Add Encrypt Error Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.1 complete
**Priority**: HIGH

**Description**: Add error path tests for Encrypt function.

**Acceptance Criteria**:
- [ ] 3.2.1: Analyze Encrypt error paths
- [ ] 3.2.2: Write tests for invalid plaintext
- [ ] 3.2.3: Write tests for encryption failures
- [ ] 3.2.4: Write tests for repository errors
- [ ] 3.2.5: Verify coverage improvement
- [ ] 3.2.6: Commit: "test(jose): add Encrypt error tests"

**Files**:
- internal/apps/jose/ja/service/service_test.go (add)

---

### Task 3.3: Add RotateMaterial Error Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.2 complete
**Priority**: HIGH

**Description**: Add error path tests for RotateMaterial function.

**Acceptance Criteria**:
- [ ] 3.3.1: Analyze RotateMaterial error paths
- [ ] 3.3.2: Write tests for invalid key IDs
- [ ] 3.3.3: Write tests for rotation failures
- [ ] 3.3.4: Write tests for database errors
- [ ] 3.3.5: Verify coverage improvement
- [ ] 3.3.6: Commit: "test(jose): add RotateMaterial error tests"

**Files**:
- internal/apps/jose/ja/service/service_test.go (add)

---

### Task 3.4: Add CreateEncryptedJWT Error Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.3 complete
**Priority**: HIGH

**Description**: Add error path tests for CreateEncryptedJWT function.

**Acceptance Criteria**:
- [ ] 3.4.1: Analyze CreateEncryptedJWT error paths
- [ ] 3.4.2: Write tests for invalid claims
- [ ] 3.4.3: Write tests for JWE creation failures
- [ ] 3.4.4: Write tests for signing errors
- [ ] 3.4.5: Verify coverage improvement
- [ ] 3.4.6: Commit: "test(jose): add CreateEncryptedJWT error tests"

**Files**:
- internal/apps/jose/ja/service/service_test.go (add)

---

### Task 3.5: Add EncryptWithKID Error Tests

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.4 complete
**Priority**: HIGH

**Description**: Add error path tests for EncryptWithKID function.

**Acceptance Criteria**:
- [ ] 3.5.1: Analyze EncryptWithKID error paths
- [ ] 3.5.2: Write tests for invalid KID
- [ ] 3.5.3: Write tests for key not found
- [ ] 3.5.4: Write tests for encryption failures
- [ ] 3.5.5: Verify coverage improvement
- [ ] 3.5.6: Commit: "test(jose): add EncryptWithKID error tests"

**Files**:
- internal/apps/jose/ja/service/service_test.go (add)

---

### Task 3.6: Verify JOSE-JA ≥95% Coverage

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.5 complete
**Priority**: HIGH

**Description**: Verify jose/service achieves ≥95% coverage.

**Acceptance Criteria**:
- [ ] 3.6.1: Run coverage: `go test -cover ./internal/apps/jose/ja/service/`
- [ ] 3.6.2: Verify ≥95% coverage
- [ ] 3.6.3: Generate HTML report
- [ ] 3.6.4: Document actual coverage
- [ ] 3.6.5: Commit: "docs(jose): ≥95% coverage achieved"

**Files**:
- docs/fixes-needed-plan-tasks-v4/tasks.md (update)

---

### Task 3.7: Migrate JOSE-JA to ServerBuilder Pattern

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.6 complete
**Priority**: CRITICAL

**Description**: Migrate JOSE-JA from custom server infrastructure to ServerBuilder pattern.

**Acceptance Criteria**:
- [ ] 3.7.1: Replace custom server setup with ServerBuilder
- [ ] 3.7.2: Implement domain route registration callback
- [ ] 3.7.3: Verify dual HTTPS servers functional
- [ ] 3.7.4: Remove obsolete custom server code
- [ ] 3.7.5: All tests pass
- [ ] 3.7.6: Commit: "refactor(jose): migrate to ServerBuilder pattern"

**Files**:
- internal/apps/jose/ja/server.go (refactor)

---

### Task 3.8: Implement JOSE-JA Merged Migrations

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.7 complete
**Priority**: CRITICAL

**Description**: Implement merged migrations pattern (template 1001-1004 + domain 2001+).

**Acceptance Criteria**:
- [ ] 3.8.1: Create domain migrations (2001+)
- [ ] 3.8.2: Configure merged migrations in ServerBuilder
- [ ] 3.8.3: Test migrations on PostgreSQL
- [ ] 3.8.4: Verify schema correct
- [ ] 3.8.5: Commit: "feat(jose): implement merged migrations"

**Files**:
- internal/apps/jose/ja/repository/migrations/ (create)

---

### Task 3.9: Add SQLite Support to JOSE-JA

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.8 complete
**Priority**: HIGH

**Description**: Add cross-database compatibility (PostgreSQL + SQLite).

**Acceptance Criteria**:
- [ ] 3.9.1: Update UUID fields to TEXT type
- [ ] 3.9.2: Update JSON fields to serializer:json
- [ ] 3.9.3: Add NullableUUID for foreign keys
- [ ] 3.9.4: Configure SQLite WAL mode + busy timeout
- [ ] 3.9.5: Test on both databases
- [ ] 3.9.6: Commit: "feat(jose): add SQLite cross-DB support"

**Files**:
- internal/apps/jose/ja/repository/models.go (update)
- internal/apps/jose/ja/config/ (update)

---

### Task 3.10: Implement Multi-Tenancy

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.9 complete
**Priority**: CRITICAL

**Description**: Implement schema-level multi-tenancy isolation.

**Acceptance Criteria**:
- [ ] 3.10.1: Add tenant_id columns to all tables
- [ ] 3.10.2: Add tenant_id indexes
- [ ] 3.10.3: Update all queries with tenant filtering
- [ ] 3.10.4: Add tests for tenant isolation
- [ ] 3.10.5: Commit: "feat(jose): implement multi-tenancy"

**Files**:
- internal/apps/jose/ja/repository/models.go (update)
- internal/apps/jose/ja/repository/*_repository.go (update)

---

### Task 3.11: Add Registration Flow Endpoint

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.10 complete
**Priority**: HIGH

**Description**: Add /auth/register endpoint for tenant/user creation.

**Acceptance Criteria**:
- [ ] 3.11.1: Implement registration handler
- [ ] 3.11.2: Add validation rules
- [ ] 3.11.3: Test create_tenant=true flow
- [ ] 3.11.4: Test join existing tenant flow
- [ ] 3.11.5: Commit: "feat(jose): add registration flow endpoint"

**Files**:
- internal/apps/jose/ja/apis/handler/auth_register.go (create)

---

### Task 3.12: Add Session Management

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.11 complete
**Priority**: HIGH

**Description**: Add SessionManagerService from template.

**Acceptance Criteria**:
- [ ] 3.12.1: Integrate SessionManagerService
- [ ] 3.12.2: Add session creation on auth
- [ ] 3.12.3: Add session validation middleware
- [ ] 3.12.4: Test session lifecycle
- [ ] 3.12.5: Commit: "feat(jose): add session management"

**Files**:
- internal/apps/jose/ja/service/ (integrate)

---

### Task 3.13: Add Realm Service

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.12 complete
**Priority**: MEDIUM

**Description**: Add RealmService for authentication context.

**Acceptance Criteria**:
- [ ] 3.13.1: Integrate RealmService
- [ ] 3.13.2: Configure realm policies
- [ ] 3.13.3: Test realm isolation
- [ ] 3.13.4: Commit: "feat(jose): add realm service"

**Files**:
- internal/apps/jose/ja/service/ (integrate)

---

### Task 3.14: Add Browser API Patterns

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.13 complete
**Priority**: HIGH

**Description**: Add /browser/** paths with CSRF/CORS/CSP middleware.

**Acceptance Criteria**:
- [ ] 3.14.1: Add /browser/** route registration
- [ ] 3.14.2: Configure CSRF middleware
- [ ] 3.14.3: Configure CORS middleware
- [ ] 3.14.4: Configure CSP headers
- [ ] 3.14.5: Test browser vs service path isolation
- [ ] 3.14.6: Commit: "feat(jose): add browser API patterns"

**Files**:
- internal/apps/jose/ja/apis/ (add browser handlers)

---

### Task 3.15: Migrate Docker Compose to YAML + Docker Secrets

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent
**Dependencies**: Task 3.14 complete
**Priority**: HIGH

**Description**: Update Docker Compose to use YAML configs + Docker secrets (NOT .env).

**Acceptance Criteria**:
- [ ] 3.15.1: Create YAML config files (dev, prod, test)
- [ ] 3.15.2: Move sensitive values to Docker secrets
- [ ] 3.15.3: Update compose.yml to use YAML + secrets
- [ ] 3.15.4: Document .env as LAST RESORT
- [ ] 3.15.5: Test all environments
- [ ] 3.15.6: Commit: "refactor(jose): Docker compose YAML + secrets"

**Files**:
- deployments/jose/compose.yml (update)
- configs/jose/ (create YAML configs)

---

## Phase 4: Shared Packages Coverage (Foundation Quality)

**Objective**: Bring shared packages to ≥98% coverage
**Status**: ⏳ NOT STARTED
**Current**: pool 61.5%, telemetry 67.5%

### Task 4.1: Add Pool Worker Thread Tests

**Objective**: Unblock cipher-im mutation testing (currently 0% - UNACCEPTABLE)

**Status**: ⏳ NOT STARTED

### Task 2.1: Fix Cipher-IM Docker Infrastructure

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent

**Dependencies**: Phase 1 complete
**Priority**: CRITICAL

**Description**: Fix Docker compose issues blocking cipher-im mutation testing (OTEL mismatch, E2E tag bypass, health checks).

**Acceptance Criteria**:
- [ ] 2.1.1: Resolve OTEL HTTP/gRPC mismatch
- [ ] 2.1.2: Fix E2E tag bypass issue
- [ ] 2.1.3: Verify health checks pass
- [ ] 2.1.4: Run `docker compose -f cmd/cipher-im/docker-compose.yml up -d`
- [ ] 2.1.5: All services healthy (0 unhealthy)
- [ ] 2.1.6: Commit: "fix(cipher-im): unblock Docker compose for mutation testing"

**Files**:
- cmd/cipher-im/docker-compose.yml (fix)
- configs/cipher/ (update)

---

### Task 2.2: Run Gremlins Baseline on Cipher-IM

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent

**Dependencies**: Task 2.1 complete
**Priority**: HIGH

**Description**: Run initial gremlins mutation testing campaign on cipher-im to establish baseline efficacy.

**Acceptance Criteria**:
- [ ] 2.2.1: Run: `gremlins unleash ./internal/apps/cipher/im/`
- [ ] 2.2.2: Collect output to /tmp/gremlins_cipher_baseline.log
- [ ] 2.2.3: Extract efficacy percentage
- [ ] 2.2.4: Document baseline in research.md
- [ ] 2.2.5: Commit: "docs(cipher-im): mutation baseline - XX.XX% efficacy"

**Files**:
- docs/fixes-needed-plan-tasks-v4/research.md (update)

---

### Task 2.3: Analyze Cipher-IM Lived Mutations

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent

**Dependencies**: Task 2.2 complete
**Priority**: HIGH

**Description**: Analyze survived mutations from gremlins run, categorize by type and priority.

**Acceptance Criteria**:
- [ ] 2.3.1: Parse gremlins output for lived mutations
- [ ] 2.3.2: Categorize by mutation type (arithmetic, conditionals, etc.)
- [ ] 2.3.3: Prioritize by ROI (test complexity vs efficacy gain)
- [ ] 2.3.4: Document in research.md
- [ ] 2.3.5: Create kill plan (target 98% efficacy)
- [ ] 2.3.6: Commit: "docs(cipher-im): mutation analysis with kill plan"

**Files**:
- docs/fixes-needed-plan-tasks-v4/research.md (update)

---

### Task 2.4: Kill Cipher-IM Mutations for 98% Efficacy

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent

**Dependencies**: Task 2.3 complete
**Priority**: CRITICAL

**Description**: Write targeted tests to kill survived mutations and achieve ≥98% efficacy ideal target.

**Acceptance Criteria**:
- [ ] 2.4.1: Implement tests for HIGH priority mutations
- [ ] 2.4.2: Implement tests for MEDIUM priority mutations
- [ ] 2.4.3: Re-run gremlins: `gremlins unleash ./internal/apps/cipher/im/`
- [ ] 2.4.4: Verify efficacy ≥98%
- [ ] 2.4.5: All tests pass
- [ ] 2.4.6: Coverage maintained or improved
- [ ] 2.4.7: Commit: "test(cipher-im): achieve 98% mutation efficacy - XX.XX%"

**Files**:
- internal/apps/cipher/im/repository/*_test.go (add tests)
- internal/apps/cipher/im/service/*_test.go (add tests)

---

### Task 2.5: Verify Cipher-IM Mutation Testing Complete

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent

**Dependencies**: Task 2.4 complete
**Priority**: HIGH

**Description**: Final verification that cipher-im achieves ≥98% mutation efficacy.

**Acceptance Criteria**:
- [ ] 2.5.1: Run final gremlins: `gremlins unleash ./internal/apps/cipher/im/`
- [ ] 2.5.2: Verify efficacy ≥98%
- [ ] 2.5.3: Update tasks.md with actual efficacy
- [ ] 2.5.4: Update plan.md with Phase 2 completion
- [ ] 2.5.5: Document in completed.md
- [ ] 2.5.6: Commit: "docs(v4): mark Phase 2 complete - cipher-im 98% efficacy"

**Files**:
- docs/fixes-needed-plan-tasks-v4/tasks.md (update)
- docs/fixes-needed-plan-tasks-v4/plan.md (update)
- docs/fixes-needed-plan-tasks-v4/completed.md (new)

---

## Phase 3: Template Mutation Cleanup (OPTIONAL - LOW PRIORITY)

**Objective**: Address remaining template mutation (currently 98.91% efficacy)

**Status**: ⏳ DEFERRED (template already exceeds 98% target)

### Task 3.1: Analyze Remaining tls_generator.go Mutation

**Status**: ⏳ DEFERRED
**Owner**: LLM Agent

**Dependencies**: Phase 2 complete
**Priority**: LOW (optional cleanup)

**Description**: Analyze the 1 remaining lived mutation in tls_generator.go to determine if killable.

**Acceptance Criteria**:
- [ ] 3.1.1: Review gremlins output for tls_generator.go mutation
- [ ] 3.1.2: Analyze mutation type and location
- [ ] 3.1.3: Determine if killable with tests
- [ ] 3.1.4: Document findings in research.md

**Files**:
- docs/fixes-needed-plan-tasks-v4/research.md (update)

---

### Task 3.2: Determine Killability or Inherent Limitation

**Status**: ⏳ DEFERRED
**Owner**: LLM Agent

**Dependencies**: Task 3.1 complete
**Priority**: LOW

**Description**: Make decision on whether mutation is killable or represents inherent testing limitation.

**Acceptance Criteria**:
- [ ] 3.2.1: Assess test implementation complexity
- [ ] 3.2.2: Assess efficacy gain (0.09% to reach 99%)
- [ ] 3.2.3: Document decision (killable vs inherent limitation)
- [ ] 3.2.4: Update mutation-analysis.md

**Files**:
- docs/gremlins/mutation-analysis.md (update)

---

### Task 3.3: Implement Test if Feasible

**Status**: ⏳ DEFERRED
**Owner**: LLM Agent

**Dependencies**: Task 3.2 complete
**Priority**: LOW

**Description**: If mutation determined killable with reasonable effort, implement test.

**Acceptance Criteria**:
- [ ] 3.3.1: Implement test (if feasible)
- [ ] 3.3.2: Run gremlins verification
- [ ] 3.3.3: Verify efficacy improvement (98.91% → 99%+)
- [ ] 3.3.4: Update tasks.md and plan.md
- [ ] 3.3.5: Commit: "test(template): kill final mutation - 99%+ efficacy"

**Files**:
- internal/apps/template/service/config/*_test.go (add test)

---

## Phase 4: Continuous Mutation Testing

**Objective**: Enable automated mutation testing in CI/CD

**Status**: ⏳ NOT STARTED

### Task 4.1: Verify ci-mutation.yml Workflow

**Status**: ⏳ NOT STARTED
**Owner**: LLM Agent

**Dependencies**: Phase 2 complete (cipher-im unblocked)
**Priority**: HIGH

**Description**: Verify existing CI/CD mutation testing workflow is correctly configured.

**Acceptance Criteria**:
- [ ] 4.1.1: Review .github/workflows/ci-mutation.yml
- [ ] 4.1.2: Verify workflow triggers correctly
- [ ] 4.1.3: Verify artifact upload configured
- [ ] 4.1.4: Document any required changes
- [ ] 4.1.5: Commit if changes needed: "ci(mutation): verify workflow configuration"

**Files**:
- .github/workflows/ci-mutation.yml (verify)

---

[Additional tasks 4.2-7.35 follow similar pattern - truncated for brevity]

---

## Cross-Cutting Tasks

### Documentation
- [ ] Update README.md with mutation testing instructions
- [ ] Update DEV-SETUP.md with workflow setup
- [ ] Create research.md with comparison table
- [ ] Update completed.md as phases finish

### Testing
- [ ] All tests pass (`runTests`)
- [ ] Coverage ≥95% production, ≥98% infrastructure
- [ ] Mutation efficacy ≥98% ideal (ALL services)
- [ ] Race detector clean on Linux

### Quality
- [ ] Linting passes (`golangci-lint run`)
- [ ] No new TODOs without tracking
- [ ] Conventional commits enforced
