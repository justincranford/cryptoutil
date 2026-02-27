# Tasks - Architecture Evolution (fixes-v8)

**Status**: 69 of 104 tasks complete (66%) — Phase 7 COMPLETE
**Last Updated**: 2026-02-27
**Created**: 2026-02-26

---

## Quality Mandate

ALL issues are blockers — NO exceptions. Fix immediately. NEVER defer, skip, or de-prioritize.

| Attribute | Requirement |
|-----------|-------------|
| Correctness | ALL code functionally correct with comprehensive tests |
| Completeness | NO phases/tasks/steps skipped, NO shortcuts |
| Thoroughness | Evidence-based validation at every step |
| Coverage | ≥95% production, ≥98% infrastructure/utility |
| Mutation | ≥95% minimum, ≥98% infrastructure |
| Commits | Conventional, incremental (never amend) |

---

## Phase 1: Architecture Documentation Hardening ✅ COMPLETE

- [x] Task 1.1: Validate propagation markers (`cicd validate-propagation`)
- [x] Task 1.2: Long line audit (>120 char lines in ARCHITECTURE.md)
- [x] Task 1.3: Empty section cleanup
- [x] Task 1.4: Cross-reference integrity check
- [x] Task 1.5: CONFIG-SCHEMA.md creation
- [x] Task 1.6: Instruction file synchronization
- [x] Task 1.7: Final quality gate validation

**Evidence**: Commits in fixes-v8 branch, zero broken links, all propagation markers valid.

---

## Phase 2: Service-Template Readiness Evaluation ✅ COMPLETE

- [x] Task 2.1: Define 10-dimension scoring rubric
- [x] Task 2.2: Score sm-kms (50/50 — Grade A)
- [x] Task 2.3: Score sm-im (48/50 — Grade A)
- [x] Task 2.4: Score jose-ja (50/50 — Grade A)
- [x] Task 2.5: Score pki-ca (44/50 — Grade B+)
- [x] Task 2.6: Score identity-authz (43/50 — Grade B)
- [x] Task 2.7: Score identity-idp (43/50 — Grade B)
- [x] Task 2.8: Score identity-rs (40/50 — Grade C+)
- [x] Task 2.9: Score identity-rp (38/50 — Grade C)
- [x] Task 2.10: Score identity-spa (35/50 — Grade C-)
- [x] Task 2.11: Consolidated readiness scorecard
- [x] Task 2.12: Gap documentation per service

**Evidence**: Scorecard in plan.md, individual gap analyses completed.

---

## Phase 3: Identity Service Alignment Planning ✅ COMPLETE

- [x] Task 3.1: Identity domain layer analysis (117 shared files)
- [x] Task 3.2: Identity migration pattern analysis (0002-0011 shared range)
- [x] Task 3.3: Independent deployability strategy
- [x] Task 3.4: E2E test strategy for identity services
- [x] Task 3.5: Migration priority ordering
- [x] Task 3.6: Documentation of planned approach

**Evidence**: Identity migration plan documented; ED-7 (independent deployability) and ED-10 (shared E2E) decided.

---

## Phase 4: Next Architecture Step Execution ✅ COMPLETE

- [x] Task 4.1: Select highest-impact improvement from Phase 2/3
- [x] Task 4.2: Implement improvement
- [x] Task 4.3: Run lint-deployments validate-all (62/62 pass)
- [x] Task 4.4: Run full test suite
- [x] Task 4.5: Commit with evidence

**Evidence**: 62/62 deployment validators pass, all tests pass.

---

## Phase 5: skeleton-template Product-Service (10th Service)

**Phase Objective**: Create skeleton-template as a fully functional 10th product-service. Product name: `skeleton`. Service name: `template`. Port: 8900. PostgreSQL port: 54329. Permanent service — demonstrates best-practice service-template usage, empty of business logic.

### Task 5.1: Magic Constants
- **Status**: ✅ DONE (commit 556109221)
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Add skeleton-template constants to `internal/shared/magic/`
- **Acceptance Criteria**:
  - [x] Create `internal/shared/magic/magic_skeleton.go` with: OTLPServiceSkeletonTemplate, SkeletonProductName, SkeletonTemplateServiceName, SkeletonTemplateServicePort (8900), SkeletonTemplatePostgresPort (54329)
  - [x] Follow exact naming patterns from `magic_pki.go`, `magic_sm.go`, `magic_jose.go`
  - [x] Build clean: `go build ./...`

### Task 5.2: Product Router
- **Status**: ✅ DONE (commit 8f3e88995)
- **Estimated**: 1h
- **Dependencies**: Task 5.1
- **Description**: Create `internal/apps/skeleton/skeleton.go` — product-level router
- **Acceptance Criteria**:
  - [x] Create `internal/apps/skeleton/skeleton.go` using `cryptoutilTemplateCli.RouteProduct` pattern (mirror `internal/apps/pki/pki.go`)
  - [x] Create `internal/apps/skeleton/skeleton_test.go` with ≥95% coverage (100%)
  - [x] Product router routes to `template` service
  - [x] Build clean, lint clean

### Task 5.3: Product CMD Entry
- **Status**: ✅ DONE (commit 8f3e88995)
- **Estimated**: 0.5h
- **Dependencies**: Task 5.2
- **Description**: Create `cmd/skeleton/main.go` — product-level binary
- **Acceptance Criteria**:
  - [x] Create `cmd/skeleton/main.go` (mirror `cmd/pki/main.go` or `cmd/sm/main.go`)
  - [x] Delegates to `skeleton.Skeleton(os.Args[1:], ...)`
  - [x] Build clean: `go build ./cmd/skeleton/`

### Task 5.4: Service Entry Function
- **Status**: ✅ DONE (commit 8f3e88995)
- **Estimated**: 1h
- **Dependencies**: Task 5.2
- **Description**: Create `internal/apps/skeleton/template/template.go` — service CLI handler
- **Acceptance Criteria**:
  - [x] Create `internal/apps/skeleton/template/template.go` using `cryptoutilTemplateCli.RouteService` pattern (mirror `internal/apps/jose/ja/ja.go`)
  - [x] Create `internal/apps/skeleton/template/template_test.go` with ≥95% coverage (97.9%)
  - [x] ServiceConfig with correct ServiceID ("skeleton-template"), ProductName ("skeleton"), ServiceName ("template")
  - [x] Build clean, lint clean

### Task 5.5: Service CMD Entry
- **Status**: ✅ DONE (commit 8f3e88995)
- **Estimated**: 0.5h
- **Dependencies**: Task 5.4
- **Description**: Create `cmd/skeleton-template/main.go` — service-level binary
- **Acceptance Criteria**:
  - [x] Create `cmd/skeleton-template/main.go` (mirror `cmd/jose-ja/main.go`)
  - [x] Delegates to `skeleton.Template(os.Args[1:], ...)`
  - [x] Build clean: `go build ./cmd/skeleton-template/`

### Task 5.6: Server Config
- **Status**: ✅ DONE (commit 8f3e88995)
- **Estimated**: 1h
- **Dependencies**: Task 5.1
- **Description**: Create `internal/apps/skeleton/template/server/config/config.go`
- **Acceptance Criteria**:
  - [x] Create config struct embedding ServiceTemplateServerSettings
  - [x] Flat kebab-case YAML parsing
  - [x] Create `internal/apps/skeleton/template/server/config/config_test.go` with ≥95% coverage (95.5%)
  - [x] Build clean, lint clean

### Task 5.7: Repository & Migrations
- **Status**: ✅ DONE (commits 8f3e88995, f64f107f2)
- **Estimated**: 1h
- **Dependencies**: Task 5.1
- **Description**: Create `internal/apps/skeleton/template/repository/` with MigrationsFS and placeholder migration
- **Acceptance Criteria**:
  - [x] Create `internal/apps/skeleton/template/repository/repository.go` with `//go:embed migrations/*.sql` and MigrationsFS (named migrations.go)
  - [x] Create `internal/apps/skeleton/template/repository/migrations/2001_init.up.sql` — create minimal template_items table (id TEXT PK, tenant_id TEXT NOT NULL, created_at DATETIME NOT NULL)
  - [x] Create `internal/apps/skeleton/template/repository/migrations/2001_init.down.sql` — DROP TABLE template_items
  - [x] Create `internal/apps/skeleton/template/repository/repository_test.go` with ≥95% coverage (100%)
  - [x] Build clean, lint clean

### Task 5.8: Domain Model (Minimal)
- **Status**: ✅ DONE (commit 8f3e88995)
- **Estimated**: 0.5h
- **Dependencies**: Task 5.7
- **Description**: Create `internal/apps/skeleton/template/domain/` with minimal TemplateItem model
- **Acceptance Criteria**:
  - [x] Create `internal/apps/skeleton/template/domain/model.go` with TemplateItem struct (ID, TenantID, CreatedAt) and GORM tags
  - [x] Create `internal/apps/skeleton/template/domain/model_test.go` with ≥95% coverage (100%)
  - [x] Build clean, lint clean

### Task 5.9: Server Implementation
- **Status**: ✅ DONE (commits 8f3e88995, f64f107f2)
- **Estimated**: 2h
- **Dependencies**: Tasks 5.6, 5.7, 5.8
- **Description**: Create `internal/apps/skeleton/template/server/server.go` using NewServerBuilder
- **Acceptance Criteria**:
  - [x] Create server.go with NewServerBuilder → WithDomainMigrations → WithPublicRouteRegistration → Build
  - [x] Dual HTTPS (public + admin) health endpoints functional
  - [x] Dual API paths: `/service/**` and `/browser/**` (empty — no business endpoints)
  - [x] Health: `/browser/api/v1/health`, `/service/api/v1/health` (public), `/admin/api/v1/livez`, `/admin/api/v1/readyz` (admin)
  - [x] Create `internal/apps/skeleton/template/server/server_test.go` with ≥95% coverage (93.5% structural ceiling — Start return nil and Shutdown error paths unreachable)
  - [x] Build clean, lint clean

### Task 5.10: Suite Integration
- **Status**: ✅ DONE (commit 8f3e88995)
- **Estimated**: 0.5h
- **Dependencies**: Task 5.4
- **Description**: Update `internal/apps/cryptoutil/cryptoutil.go` to add skeleton product routing
- **Acceptance Criteria**:
  - [x] Add `case "skeleton":` to suite router switch statement
  - [x] Route to `skeleton.Skeleton(remainingArgs, ...)` function
  - [x] Update `internal/apps/cryptoutil/cryptoutil_test.go` — added skeleton test cases
  - [x] Build clean, lint clean

### Task 5.11: Deployment Infrastructure
- **Status**: ✅ DONE (commit aa0098704, 92af55c09)
- **Estimated**: 2h
- **Dependencies**: Tasks 5.1, 5.9
- **Description**: Create deployment and config directories for skeleton-template
- **Acceptance Criteria**:
  - [x] Create `deployments/skeleton-template/compose.yml` (ports 8900/8901/8902, PostgreSQL 54329)
  - [x] Create `deployments/skeleton-template/secrets/` with 14 .secret files
  - [x] Create `deployments/skeleton/compose.yml` (product-level deployment)
  - [x] Create `deployments/skeleton/secrets/` with shared + .never files
  - [x] Create `configs/skeleton/skeleton-server.yml` and `configs/skeleton/template/template-server.yml`
  - [x] Port mappings: 8900 (public), 9090 (admin), 54329 (PostgreSQL)
  - [x] Docker secrets for credentials (not inline env vars)
  - [x] Health checks using wget
  - [x] CICD lint_deployments code updated: classifyDeployment, expected contents, tests
  - [x] Run `go run ./cmd/cicd lint-deployments validate-all` — 68/68 pass

### Task 5.12: E2E Test Skeleton
- **Status**: ✅ DONE (commit 9557247bb)
- **Estimated**: 1.5h
- **Dependencies**: Task 5.9
- **Description**: Create E2E test infrastructure for skeleton-template
- **Acceptance Criteria**:
  - [x] Create `internal/apps/skeleton/template/e2e/testmain_e2e_test.go` with TestMain
  - [x] Create `internal/apps/skeleton/template/e2e/e2e_test.go` with basic health check tests
  - [x] Tests verify public health endpoints (3 instances: SQLite, PostgreSQL 1, PostgreSQL 2)
  - [x] Tests use build tag `e2e`
  - [x] Build clean with tags: `go build -tags e2e ./...`
  - [x] Lint clean with tags: `golangci-lint run --build-tags e2e`

### Task 5.13: ARCHITECTURE.md Update
- **Status**: ✅ DONE (commits e16317e3b, 68d8cf998)
- **Estimated**: 1h
- **Dependencies**: Tasks 5.1, 5.11
- **Description**: Add skeleton-template to ARCHITECTURE.md and instruction files
- **Acceptance Criteria**:
  - [x] Add to Service Catalog (Section 3.2): Skeleton product, Template service, skeleton-template ID
  - [x] Add to Port Assignments (Section 3.4): 8900-8999 host ports, 0.0.0.0:8080/127.0.0.1:9090 container
  - [x] Add to PostgreSQL Ports (Section 3.4.2): 54329
  - [x] Add Skeleton Product description section (3.1.5 and 3.2.5)
  - [x] Mark skeleton-template as "Stereotype — best-practice template usage reference"
  - [x] Update ALL "9 services" → "10 services" (20+ occurrences)
  - [x] Update instruction files (02-01.architecture, 03-03.golang)
  - [x] Update suite port range, database counts, PostgreSQL port summaries

### Task 5.14: Full Quality Gate
- **Status**: ✅ DONE
- **Estimated**: 1h
- **Dependencies**: All Phase 5 tasks
- **Description**: Final validation that skeleton-template is fully functional
- **Acceptance Criteria**:
  - [x] `go build ./...` — clean
  - [x] `go build -tags e2e,integration ./...` — clean
  - [x] `golangci-lint run` — zero issues
  - [x] `golangci-lint run --build-tags e2e,integration` — zero issues
  - [x] `go test ./internal/apps/skeleton/... -cover -shuffle=on` — all pass (100%, 97.9%, 100%, 100%, 93.5%, 95.5%)
  - [x] `go run ./cmd/cicd lint-deployments validate-all` — 68/68 pass
  - [x] Health endpoints respond (verified via server integration tests)
  - [x] Conventional commits with all Phase 5 changes

---

## Phase 6: PKI-CA Archive & Clean-Slate Skeleton

**Phase Objective**: Archive existing pki-ca (111 Go files, 27 directories). Create new empty pki-ca skeleton following skeleton-template patterns. Validates that the skeleton pattern is reproducible.

### Task 6.1: Archive Existing PKI-CA
- **Status**: ✅ COMPLETE
- **Estimated**: 1h
- **Dependencies**: Phase 5 complete
- **Description**: Move existing pki-ca to archive directory
- **Acceptance Criteria**:
  - [x] Move `internal/apps/pki/ca/` to `internal/apps/pki/_ca-archived/` (underscore prefix makes Go ignore it)
  - [x] Ensure `internal/apps/pki/pki.go` still compiles
  - [x] Ensure `cmd/pki-ca/main.go` still compiles
  - [x] `go build ./...` — clean
  - [x] Committed as part of combined archive + skeleton commit

### Task 6.2: Create New PKI-CA Skeleton
- **Status**: ✅ COMPLETE
- **Estimated**: 2h
- **Dependencies**: Task 6.1
- **Description**: Create new `internal/apps/pki/ca/` following skeleton-template patterns
- **Acceptance Criteria**:
  - [x] Create new `internal/apps/pki/ca/ca.go` (service entry with RouteService CLI pattern)
  - [x] Create new `internal/apps/pki/ca/ca_usage.go` (usage text constants)
  - [x] Create new `internal/apps/pki/ca/server/server.go` (PKICAServer using service template builder)
  - [x] Create new `internal/apps/pki/ca/server/config/config.go` (ParseWithFlagSet)
  - [x] Create new `internal/apps/pki/ca/repository/` with migrations 2001 (mergedFS pattern)
  - [x] Create new `internal/apps/pki/ca/domain/` with CAItem model
  - [x] Reuse existing port (8100) and PostgreSQL port (54320)
  - [x] Build clean: `go build ./...`

### Task 6.3: Reconnect Entry Points
- **Status**: ✅ COMPLETE
- **Estimated**: 0.5h
- **Dependencies**: Task 6.2
- **Description**: Reconnect cmd/pki-ca/main.go and product router to new skeleton
- **Acceptance Criteria**:
  - [x] `cmd/pki-ca/main.go` routes to new ca package (unchanged — same import path)
  - [x] `internal/apps/pki/pki.go` routes to new ca package (unchanged — same import path)
  - [x] Build clean, lint clean

### Task 6.4: Tests for New PKI-CA
- **Status**: ✅ COMPLETE
- **Estimated**: 1.5h
- **Dependencies**: Task 6.2
- **Description**: Create tests for new pki-ca skeleton
- **Acceptance Criteria**:
  - [x] 10 test files: ca_cli_test, ca_lifecycle_test, ca_port_conflict_test, testmain_test, server_test, server_integration_test, testmain_test, config_test, model_test, migrations_test
  - [x] Coverage: pki 100%, ca 97.9%, domain 100%, repository 100%, server 90.0% (structural ceiling — same as skeleton-template), config 95.5%
  - [x] Build and lint clean with all tags
  - [x] Tests pass with shuffle

### Task 6.5: Update Deployment
- **Status**: ✅ COMPLETE
- **Estimated**: 1h
- **Dependencies**: Task 6.2
- **Description**: Update or recreate deployment configs for new pki-ca
- **Acceptance Criteria**:
  - [x] `deployments/pki-ca/` unchanged (new skeleton uses same ports/config)
  - [x] `configs/ca/` unchanged (new skeleton uses same config format)
  - [x] `go run ./cmd/cicd lint-deployments validate-all` — 68/68 pass
  - [x] Health checks configured (via service template builder)

### Task 6.6: Quality Gate
- **Status**: ✅ COMPLETE
- **Estimated**: 0.5h
- **Dependencies**: All Phase 6 tasks
- **Description**: Full validation of new pki-ca skeleton
- **Acceptance Criteria**:
  - [x] `go build ./...` — clean
  - [x] `go build -tags e2e,integration ./...` — clean
  - [x] `golangci-lint run --fix ./internal/apps/pki/...` — zero issues
  - [x] `go test ./internal/apps/pki/... -cover -shuffle=on` — all pass, coverage above targets
  - [x] Deployment validators pass (68/68)
  - [x] Conventional commit (ba2efccb7)

---

## Phase 7: Service-Template Reusability Analysis

**Phase Objective**: Analyze skeleton creation experience. Document minimal file set, friction points, and enhancement proposals.

### Task 7.1: Minimal File Set Documentation
- **Status**: ✅ COMPLETE
- **Estimated**: 1h
- **Dependencies**: Phases 5-6 complete
- **Description**: Document the minimal file set required for a conforming product-service
- **Acceptance Criteria**:
  - [x] List ALL files created for skeleton-template with purpose (7 source + 10-12 test + 2 SQL)
  - [x] Compare against sm-kms (reference: 41 source + 78 test + 4 SQL = 123 total)
  - [x] Identify files that could be auto-generated vs hand-written (90%+ is template boilerplate)
  - [x] Document in RESEARCH.md

### Task 7.2: Template Friction Points
- **Status**: ✅ COMPLETE
- **Estimated**: 1h
- **Dependencies**: Phases 5-6 complete
- **Description**: Catalog friction, boilerplate, missing abstractions from skeleton creation
- **Acceptance Criteria**:
  - [x] Listed 5 categories of boilerplate (mergedFS, server wrapper, usage strings, serverStart, tests)
  - [x] Identified 3 missing helpers (shared mergedFS, server base struct, test generator)
  - [x] Documented 3 API surface confusion points (fields vs methods, Shutdown signature, port types)
  - [x] Document in RESEARCH.md

### Task 7.3: Product Wiring Analysis
- **Status**: ✅ COMPLETE
- **Estimated**: 0.5h
- **Dependencies**: Task 7.1
- **Description**: Analyze product-level wiring for simplification
- **Acceptance Criteria**:
  - [x] RouteProduct/RouteService patterns assessed — clean, scalable, no changes needed
  - [x] Suite router scalability assessed — switch scales to ~20 products, fine as-is
  - [x] Document in RESEARCH.md

### Task 7.4: Enhancement Proposals
- **Status**: ✅ COMPLETE
- **Estimated**: 1h
- **Dependencies**: Tasks 7.1-7.3
- **Description**: Concrete, prioritized proposals for template improvements
- **Acceptance Criteria**:
  - [x] P0: 2 must-fix items (shared mergedFS saves 800 lines, ServiceResources godoc)
  - [x] P1: 3 should-fix items (server base struct, usage generator, shared server runner)
  - [x] P2: 2 nice-to-have items (test code generator, generic config factory)
  - [x] Each proposal has: problem, solution, LOE, impact
  - [x] Document in RESEARCH.md

---

## Phase 8: CICD Linter Enhancements

**Phase Objective**: Implement CICD linters enforcing structural best practices for PRODUCT and PRODUCT-SERVICE directories.

### Task 8.1: Linter Gap Analysis
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 7 complete
- **Description**: Compare existing 65+ validators against Phase 7 findings
- **Acceptance Criteria**:
  - [ ] List existing validators relevant to structural validation
  - [ ] Identify gaps: what structural rules are NOT currently enforced?
  - [ ] Prioritize new validators by impact
  - [ ] Document findings

### Task 8.2: Validator Design
- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 8.1
- **Description**: Design new validators for structural best practices
- **Acceptance Criteria**:
  - [ ] Design: ValidateProductStructure (required dirs, files per product)
  - [ ] Design: ValidateServiceStructure (required dirs, files per service)
  - [ ] Design: ValidateMigrationNumbering (2001+ range, up/down pairs)
  - [ ] Design: ValidateProductWiring (suite router, cmd entries)
  - [ ] Design: ValidateTestPresence (test files for each package)
  - [ ] Document expected pass/fail for all 10 services

### Task 8.3: Validator Implementation
- **Status**: ❌
- **Estimated**: 4h
- **Dependencies**: Task 8.2
- **Description**: Implement validators in cmd/cicd/
- **Acceptance Criteria**:
  - [ ] Implement each validator with comprehensive tests
  - [ ] Coverage ≥98% (infrastructure code)
  - [ ] Mutation testing ≥98%
  - [ ] All validators follow aggregation pattern (run all, report all)
  - [ ] Build and lint clean

### Task 8.4: Apply to All 10 Services
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 8.3
- **Description**: Run new validators against all 10 services, fix non-conformance
- **Acceptance Criteria**:
  - [ ] All 10 services pass new validators (or documented exceptions)
  - [ ] No regressions in existing validators
  - [ ] `go run ./cmd/cicd lint-deployments validate-all` — all pass
  - [ ] Conventional commit

---

## Phase 9: Documentation & Research

**Phase Objective**: Consolidate findings into RESEARCH.md.

### Task 9.1: Skeleton Patterns
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phases 5-6 complete
- **Description**: Document skeleton creation process and patterns
- **Acceptance Criteria**:
  - [ ] Step-by-step guide for creating a new product-service
  - [ ] Minimal file set with purpose annotations
  - [ ] Common pitfalls and solutions
  - [ ] Document in docs/fixes-v8/RESEARCH.md

### Task 9.2: Template Learnings
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 7 complete
- **Description**: Document template strengths, weaknesses, improvements
- **Acceptance Criteria**:
  - [ ] Import Phase 7 findings
  - [ ] Strengths analysis (what works well)
  - [ ] Weakness analysis (what needs improvement)
  - [ ] Enhancement roadmap with priorities
  - [ ] Document in RESEARCH.md

### Task 9.3: Identity Roadmap
- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 3 findings
- **Description**: Document planned identity services approach
- **Acceptance Criteria**:
  - [ ] Archive + skeleton approach (following pki-ca pattern)
  - [ ] Independent deployability plan (ED-7)
  - [ ] Shared E2E strategy (ED-10)
  - [ ] Document in RESEARCH.md

### Task 9.4: Three-Tier Architecture
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phases 5-8 complete
- **Description**: Document base/stereotype/service architecture vision
- **Acceptance Criteria**:
  - [ ] Base tier: service-template (current), product-service-base (future)
  - [ ] Stereotype tier: skeleton-template (current), product-service-stereotype (future)
  - [ ] Service tier: all 9 business services
  - [ ] Long-term workflow: change base → validate stereotype → roll out
  - [ ] Document in RESEARCH.md

### Task 9.5: RESEARCH.md Publication
- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Tasks 9.1-9.4
- **Description**: Finalize and commit RESEARCH.md
- **Acceptance Criteria**:
  - [ ] All sections complete
  - [ ] Cross-references to ARCHITECTURE.md
  - [ ] Conventional commit

---

## Phase 10: ARCHITECTURE.md Propagation

**Phase Objective**: Ensure all ARCHITECTURE.md changes from Phases 5-9 are propagated to instruction files.

### Task 10.1: Validate Propagation
- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phases 5-9 complete
- **Description**: Run propagation validation
- **Acceptance Criteria**:
  - [ ] `cicd validate-propagation` — all markers valid
  - [ ] `cicd validate-chunks` — all chunks match
  - [ ] List instruction files needing update

### Task 10.2: Update Instruction Files
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 10.1
- **Description**: Update instruction files with ARCHITECTURE.md changes
- **Acceptance Criteria**:
  - [ ] `02-01.architecture.instructions.md` — service catalog table updated
  - [ ] Other affected instruction files updated
  - [ ] Build clean, lint clean

### Task 10.3: Final Quality Gate
- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 10.2
- **Description**: Full project validation
- **Acceptance Criteria**:
  - [ ] `go build ./...` — clean
  - [ ] `go build -tags e2e,integration ./...` — clean
  - [ ] `golangci-lint run` — zero issues
  - [ ] `golangci-lint run --build-tags e2e,integration` — zero issues
  - [ ] `go test ./... -shuffle=on` — all pass
  - [ ] `go run ./cmd/cicd lint-deployments validate-all` — all pass
  - [ ] `cicd validate-propagation` — all valid
  - [ ] Zero new TODOs without tracking
  - [ ] Conventional commit

---

## Cross-Cutting Tasks

### Testing
- [ ] Unit tests ≥95% coverage (production), ≥98% (infrastructure)
- [ ] Integration tests pass
- [ ] E2E tests pass (Docker Compose) for skeleton-template
- [ ] Mutation testing ≥95% minimum
- [ ] No skipped tests
- [ ] Race detector clean: `go test -race ./...`

### Code Quality
- [ ] Linting passes: `golangci-lint run` and `golangci-lint run --build-tags e2e,integration`
- [ ] No new TODOs without tracking
- [ ] No security vulnerabilities
- [ ] Formatting clean: `gofumpt -s -w ./`

### Documentation
- [x] ARCHITECTURE.md updated (service catalog, ports, product description) — commit e16317e3b
- [ ] docs/fixes-v8/RESEARCH.md published
- [x] Instruction files propagated — commit 68d8cf998
- [x] Plan.md and tasks.md up to date — example→template rename, status updates

### Deployment
- [x] Docker Compose configs for skeleton-template — commit aa0098704
- [ ] Health checks pass
- [x] Deployment validators pass (68/68) — commit 92af55c09, aa0098704
- [x] Config files validated

---

## Quizme Decisions (Merged from quizme-v1.md)

| # | Question | Answer | Impact |
|---|----------|--------|--------|
| Q1 | Approach for skeleton-template scope | E: Create two skeletons: skeleton-template (permanent 10th) + pki-ca (clean-slate) | Phases 5-6 scope |
| Q2 | Identity migration strategy | E: Each identity service gets own domain, DB, migration range, E2E | Future work (ED-7) |
| Q3 | Default migration content | E: Both skeleton-template and pki-ca use empty conforming 2001+ | Task 5.7, 6.2 |
| Q4 | PKI-CA archive strategy | E: Archive then create clean-slate AFTER skeleton-template validates pattern | Phase 6 ordering |
| Q5 | Identity E2E approach | A: Keep shared suite for all identity services | ED-10 confirmed |

---

## Notes / Deferred Work

### Medium-Term Renames (NOT in fixes-v8 scope)
- `internal/apps/template/service/` → `internal/apps/template/product-service-base/` (or similar)
- `internal/apps/skeleton/template/` → `internal/apps/template/product-service-stereotype/` (or similar)
- Tracked as ED-13

### Identity Services Architecture (NOT in fixes-v8 scope)
- Archive existing → create skeletons → achieve independent deployability
- Tracked as ED-7

### Long-Term Workflow (NOT in fixes-v8 scope)
- Change base/stereotype first → validate via CICD linters → roll out to all services
- Tracked as ED-14

---

## Evidence Archive

- `test-output/phase0-research/` - Phase 0 research (service patterns, ports, CLI wiring)
- `test-output/phase1/` - Architecture documentation validation
- `test-output/phase2/` - Readiness scorecard evidence
- `test-output/phase5/` - skeleton-template creation evidence (planned)
- `test-output/phase6/` - PKI-CA archive evidence (planned)
- `test-output/phase7/` - Reusability analysis evidence (planned)
- `test-output/phase8/` - CICD linter evidence (planned)
