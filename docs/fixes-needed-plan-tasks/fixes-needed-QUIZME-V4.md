# JOSE-JA V4 Implementation QUIZME

**Purpose**: Identify issues, gaps, conflicts, concerns, inefficiencies, and risks in V4 planning and implementation.

**Last Updated**: 2026-01-18

---

## 1. Default Tenant Pattern Violation

**Issue**: JOSE-JA server.go still uses `ensureDefaultTenant()` helper after ARCHITECTURE.md mandates NO default tenant pattern.

**Context**: 
- ARCHITECTURE.md: "Services start cold without any default tenant/realm. All tenants MUST be created through explicit registration flow."
- PLAN.md Phase 0: "Remove WithDefaultTenant from ServerBuilder"
- TASKS.md Phase 0: Detailed 10-subtask removal process
- **BUT** `internal/apps/jose/ja/server/server.go` lines 114-118 calls `ensureDefaultTenant()` with hardcoded JoseJADefaultTenantID/JoseJADefaultRealmID

**Question**: How should JOSE-JA handle tenant setup for educational demos WITHOUT violating NO default tenant pattern?

**A.** Keep `ensureDefaultTenant()` but mark as DEMO_ONLY with warning comments
**B.** Remove `ensureDefaultTenant()`, require manual tenant registration via API before any JOSE operations
**C.** Create `docs/jose-ja/DEMO-QUICKSTART.md` with explicit registration API call examples instead of automatic default
**D.** Add `--demo-mode` CLI flag that enables default tenant ONLY when explicitly opted in
**E.** (write in):

**Answer**: C or D (C preferred for clarity, D acceptable for convenience). Pattern violation MUST be fixed - JOSE-JA calling `ensureDefaultTenant()` directly contradicts Phase 0 removal plan and ARCHITECTURE.md mandate.

---

## 2. Default Tenant Constants Still Present

**Issue**: JOSE-JA has hardcoded default tenant/realm UUIDs with TODO comments but no tracking issue.

**Context**:
- `internal/apps/jose/ja/server/server.go` lines 28-33 define `JoseJADefaultTenantID` and `JoseJADefaultRealmID`
- Comments say "TODO: Move to magic constants once feature is stable"
- Phase 0 removes default tenant pattern entirely, making these constants obsolete
- NO tracking issue for removal

**Question**: What should happen to JoseJADefaultTenantID and JoseJADefaultRealmID constants?

**A.** Delete immediately as part of Phase 0.1 (WithDefaultTenant removal)
**B.** Move to `internal/shared/magic/magic_jose.go` for test usage only
**C.** Keep in server.go but rename to `JoseJADemoTenantID` with DEMO_ONLY warning
**D.** Delete constants, update all JOSE tests to use registration flow instead
**E.** (write in):

**Answer**: D (aligns with Phase 0 objectives). Hardcoded tenant IDs violate multi-tenancy isolation principle. Tests MUST use registration flow.

---

## 3. Cipher-IM Registration Flow Missing

**Issue**: Cipher-IM server.go does NOT call registration service during startup or show registration flow examples.

**Context**:
- ARCHITECTURE.md Section "Registration Flow Pattern - REQUIRED": All services MUST support registration flow
- PLAN.md Phase 1: "Adapt Cipher-IM to use ServerBuilder WITHOUT default tenant"
- TASKS.md Phase 1: NO subtasks for registration flow implementation in Cipher-IM
- `internal/apps/cipher/im/server/server.go` uses ServerBuilder but NO registration examples
- Cipher-IM tests in `testmain_test.go` likely use OLD default tenant pattern

**Question**: Does Cipher-IM require registration flow updates as part of Phase 1?

**A.** No, Cipher-IM inherits registration from ServerBuilder automatically (no changes needed)
**B.** Yes, add `testmain_test.go` pattern with `registerTestUser()` helper in Phase 1
**C.** Yes, create `docs/cipher-im/DEMO-QUICKSTART.md` showing registration API call examples
**D.** Yes, update ALL Cipher-IM tests to use registration flow instead of assumed default tenant
**E.** (write in):

**Answer**: B and D (both required). Phase 1 tasks are INCOMPLETE - missing explicit registration flow update subtasks for Cipher-IM tests.

---

## 4. ServerBuilder TestMain Pattern Inconsistency

**Issue**: ARCHITECTURE.md and server-builder.instructions.md mandate TestMain pattern but PLAN.md/TASKS.md do NOT include explicit subtasks for it.

**Context**:
- ARCHITECTURE.md: "Tests MUST use TestMain to set up server with proper registration"
- server-builder.instructions.md: "TestMain Pattern - REQUIRED" with detailed example
- PLAN.md Phase 3: "Integrate JOSE-JA with ServerBuilder" - NO TestMain subtasks
- PLAN.md Phase 8: "E2E Testing" mentions "TestMain pattern" but Phase 3 should establish it
- TASKS.md Phase 3: NO subtasks for creating TestMain with registration flow

**Question**: Which phase should implement ServerBuilder TestMain pattern for JOSE-JA?

**A.** Phase 2 (JOSE DB Schema) - establish test patterns early
**B.** Phase 3 (JOSE ServerBuilder) - align with builder integration
**C.** Phase 8 (JOSE E2E Testing) - defer until E2E test creation
**D.** Not needed - unit tests don't require full server startup
**E.** (write in):

**Answer**: B (Phase 3 ServerBuilder integration). TestMain pattern is fundamental to builder-based testing, should be established when integrating ServerBuilder.

---

## 5. Migration Numbering Conflict Risk

**Issue**: PLAN.md shows migration numbers 2001+ for domain but does NOT verify no conflicts with existing migrations.

**Context**:
- ServerBuilder uses merged migrations: template 1001-1004, domain 2001+
- PLAN.md Phase 2.1: "migrations/2001_init_schema.up.sql"
- NO verification that existing JOSE-JA migrations (if any) don't conflict
- NO check for migration number collisions between services
- server-builder.instructions.md: "Migration Numbering: Template 1001-1004, domain 2001+"

**Question**: How should migration numbering be verified before implementation?

**A.** Run `grep -r "CREATE TABLE" internal/apps/jose/ja/repository/migrations/` to check existing migrations
**B.** Add subtask 2.0.1: "Verify no existing migrations in JOSE-JA repository" before 2.1
**C.** Document migration number ranges per service in ARCHITECTURE.md
**D.** Create migration registry file tracking all service migration number ranges
**E.** (write in):

**Answer**: B and C (both needed). Subtask 2.0.1 prevents conflicts, ARCHITECTURE.md provides reference.

---

## 6. Session Config Separation Not Reflected in ServerBuilder

**Issue**: PLAN.md mandates "Separate browser vs service session configs" but ServerBuilder signature does NOT show this separation.

**Context**:
- PLAN.md Critical Fixes: "✅ Separate browser vs service session configs"
- ServerBuilder constructor: `NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)`
- ServiceTemplateServerSettings likely contains SessionSettings field (not visible in builder code)
- ARCHITECTURE.md: "Session token format configurable: Opaque, JWE, JWS"
- NO explicit guidance on HOW session configs are separated (file-based? YAML sections?)

**Question**: How are browser vs service session configs separated in ServiceTemplateServerSettings?

**A.** Two separate YAML sections: `browser_session:` and `service_session:`
**B.** Single `session:` section with nested `browser:` and `service:` subsections
**C.** Separate config files: `browser-session.yaml` and `service-session.yaml`
**D.** Single config, session type auto-detected from request path (/browser/* vs /service/*)
**E.** (write in):

**Answer**: B (nested subsections pattern). Provides clear separation while keeping single config file.

---

## 7. Realm Filtering Removal Not Applied to JOSE-JA

**Issue**: PLAN.md mandates "Realms for authn only, removed realm_id from repository WHERE clauses" but NO verification that JOSE-JA repositories comply.

**Context**:
- PLAN.md Critical Fixes: "✅ Realms are authn only (removed from repository WHERE clauses)"
- TASKS.md Phase 0: NO subtasks for verifying realm_id removal
- JOSE-JA repositories in `internal/apps/jose/ja/repository/*.go` likely have tenant_id + realm_id WHERE clauses
- ARCHITECTURE.md: "Realms provide authentication context ONLY (not authorization filtering)"

**Question**: Which phase should verify and fix realm_id in JOSE-JA repository WHERE clauses?

**A.** Phase 0 (Service-Template) - establish pattern globally
**B.** Phase 2 (JOSE DB Schema) - align with schema design
**C.** Phase 4 (JOSE Elastic JWK) - fix during repository implementation
**D.** Phase 8 (JOSE E2E Testing) - catch during integration testing
**E.** (write in):

**Answer**: C (Phase 4 implementation). Logical to fix WHERE clauses when implementing repository methods.

---

## 8. API Simplification Not Detailed in TASKS.md

**Issue**: PLAN.md shows simplified API (key_type/key_size removed) but TASKS.md does NOT include subtasks for updating request/response structs.

**Context**:
- PLAN.md Phase 3.3: Shows GenerateJWKRequest with ONLY algorithm, use, max_materials, kid
- PLAN.md: "✅ key_type REMOVED - implied by algorithm"
- PLAN.md: "✅ key_size REMOVED - implied by algorithm"
- TASKS.md Phase 3.3: "Create JOSE HTTP Handlers" - NO subtasks for API simplification
- OpenAPI spec files likely need updates to reflect simplified parameters

**Question**: Where should API simplification be implemented?

**A.** Phase 2.2 (OpenAPI Spec) - update request/response schemas before handlers
**B.** Phase 3.3 (HTTP Handlers) - update structs when implementing handlers
**C.** Phase 4 (Elastic JWK) - align with business logic implementation
**D.** Create new Phase 2.5: "Simplify API Parameters" with explicit subtasks
**E.** (write in):

**Answer**: A (Phase 2.2 OpenAPI Spec). Schema should be simplified BEFORE handler implementation to prevent rework.

---

## 9. E2E Test Location Pattern Conflict

**Issue**: PLAN.md shows E2E tests in `internal/apps/jose/ja/e2e/` but conventional Go pattern uses `cmd/*/e2e_test.go` or `test/e2e/`.

**Context**:
- PLAN.md Phase 8: "Per product-service e2e/ subdirectory (`internal/apps/jose/ja/e2e/` pattern)"
- Go convention: Integration tests in `*_test.go` files, E2E in `cmd/` or `test/` directory
- internal/ directory typically for private library code, not tests
- Existing project has `test/load/` for Gatling load tests

**Question**: What is the correct E2E test directory pattern for cryptoutil?

**A.** Keep `internal/apps/jose/ja/e2e/` for proximity to implementation
**B.** Move to `cmd/jose-server/e2e_test.go` for proximity to entry point
**C.** Centralize in `test/e2e/jose-ja/` for consistency with `test/load/`
**D.** Use `internal/apps/jose/ja/server/e2e_test.go` (alongside server implementation)
**E.** (write in):

**Answer**: C (centralized `test/e2e/jose-ja/`). Aligns with existing `test/load/` pattern, separates E2E from library code.

---

## 10. Docker Compose E2E Pattern Not Specified

**Issue**: PLAN.md mandates Docker Compose for E2E but does NOT specify compose file location or naming convention.

**Context**:
- PLAN.md Phase 8 Q9.1: "Docker Compose for E2E tests (realistic customer experience)"
- PLAN.md Phase 8 Q9.2: "Docker Compose starts PostgreSQL container (NOT test-containers)"
- Existing compose files in `deployments/compose/`, `deployments/jose-ja/`
- NO guidance on whether E2E uses production compose files or test-specific compose files

**Question**: Where should E2E Docker Compose files reside?

**A.** `deployments/jose-ja/compose.e2e.yml` (test variant of production compose)
**B.** `test/e2e/jose-ja/compose.yml` (dedicated E2E compose file)
**C.** `internal/apps/jose/ja/e2e/compose.yml` (co-located with E2E tests)
**D.** Reuse existing `deployments/jose-ja/compose.yml` (production compose file)
**E.** (write in):

**Answer**: B (`test/e2e/jose-ja/compose.yml`). Dedicated E2E compose file enables test-specific configs without affecting production deployments.

---

## 11. Mutation Testing Scope Not Specified

**Issue**: TASKS.md mandates "≥85% production, ≥98% infrastructure" mutation score but does NOT clarify which packages are infrastructure vs production.

**Context**:
- TASKS.md Quality Gates: "Mutation: gremlins unleash ./internal/[package] ≥85% production, ≥98% infrastructure"
- internal/apps/jose/ja/domain/ - production or infrastructure?
- internal/apps/jose/ja/repository/ - production or infrastructure?
- internal/apps/jose/ja/service/ - production or infrastructure?
- internal/apps/jose/ja/server/ - production or infrastructure?
- NO definition of infrastructure vs production in PLAN.md or ARCHITECTURE.md

**Question**: How should JOSE-JA packages be classified for mutation testing?

**A.** domain/ and service/ = production (≥85%), repository/ and server/ = infrastructure (≥98%)
**B.** All internal/apps/* = production (≥85%), only internal/shared/* = infrastructure (≥98%)
**C.** server/ and repository/ = infrastructure (≥98%), domain/ and service/ = production (≥85%)
**D.** Add classification table to ARCHITECTURE.md before starting Phase 2
**E.** (write in):

**Answer**: D (add classification table). Domain models and business logic services are production (≥85%). Repository (data access) and server (HTTP handling) are infrastructure (≥98%). Should be documented BEFORE implementation.

---

## 12. Path Migration Timing Ambiguity

**Issue**: PLAN.md Phase 7 is "Path Migration" but paths should already be correct from Phase 2 (OpenAPI spec) onward.

**Context**:
- PLAN.md Phase 7: "Migrate from `/api/jose/*` to `/service/api/v1/*` and `/browser/api/v1/*`"
- PLAN.md Phase 2.2: "Create OpenAPI Spec" - should define correct paths from start
- PLAN.md Phase 3.3: "Create HTTP Handlers" - should use correct paths
- Phase 7 implies paths are wrong in Phases 2-6, then fixed in Phase 7

**Question**: When should correct API paths be implemented?

**A.** Phase 2.2 (OpenAPI Spec) - define correct paths from beginning
**B.** Phase 7 (Path Migration) - implement wrong paths first, migrate later
**C.** Delete Phase 7 entirely, ensure correct paths in Phases 2-6
**D.** Keep Phase 7 as validation phase (verify paths, no migration needed)
**E.** (write in):

**Answer**: C (delete Phase 7, correct paths from Phase 2). No reason to implement wrong paths then migrate - wastes effort.

---

## 13. Hardcoded Password Removal Strategy Missing

**Issue**: PLAN.md mandates "No hardcoded passwords in tests" but provides NO guidance on migration from existing hardcoded passwords.

**Context**:
- PLAN.md Critical Fixes: "✅ No hardcoded passwords in tests (use magic constants or UUIDv7)"
- TASKS.md Quality Gates: NO validation step for checking hardcoded passwords
- Existing tests likely have "password123", "test", "admin" hardcoded
- NO subtasks for identifying and replacing hardcoded passwords

**Question**: How should hardcoded password removal be enforced?

**A.** Add Phase 0.X: "Audit and Replace Hardcoded Passwords Across All Tests"
**B.** Add Quality Gate #8: `grep -r "password.*:" --include="*_test.go"` zero matches
**C.** Add linting rule to detect string literals matching password patterns
**D.** Document in ARCHITECTURE.md: ALL tests MUST use cryptoutilMagic.TestPassword or UUIDv7
**E.** (write in):

**Answer**: A and B (both needed). Phase 0.X provides structured removal, Quality Gate #8 prevents regression.

---

## 14. Cross-Tenant JWKS Access Not Designed

**Issue**: PLAN.md Phase 5 mentions "Cross-tenant JWKS access via tenant management API" but NO API design or implementation details.

**Context**:
- PLAN.md Phase 5: "Cross-tenant JWKS access via tenant management API (not DB config)"
- NO tenant management API design in ARCHITECTURE.md
- NO explanation of authorization model (how does Tenant A access Tenant B's JWKS?)
- NO API endpoints defined (GET /api/v1/tenants/{tenant_id}/jwks?)
- Phase 5 duration 2-3 days seems SHORT for new API design + implementation + tests

**Question**: How should cross-tenant JWKS access be designed?

**A.** Tenant A registers "trusted_tenants" list, grants access to JWKS for listed tenants
**B.** Tenant B (JWKS owner) configures "allowed_requestors" list via admin API
**C.** System admin configures cross-tenant relationships via global tenant management
**D.** Add Phase 4.5: "Design Cross-Tenant JWKS Authorization Model" with detailed subtasks
**E.** (write in):

**Answer**: D (add Phase 4.5 design phase). Cross-tenant authorization is complex, needs design phase BEFORE implementation.

---

## 15. Audit Logging Event Taxonomy Missing

**Issue**: PLAN.md Phase 6 "Audit Logging" does NOT define which events MUST be logged.

**Context**:
- PLAN.md Phase 6: "See V3 for detailed tasks - NO substantive changes"
- NO list of mandatory audit events (JWK generation? Key rotation? JWKS access? Failed auth?)
- ARCHITECTURE.md: "Audit log security events (90-day retention minimum)" - too vague
- NO guidance on audit log schema (timestamp, tenant_id, event_type, user_id, details?)

**Question**: What audit events MUST JOSE-JA log?

**A.** Create docs/jose-ja/AUDIT-EVENTS.md listing all auditable events before Phase 6
**B.** Add Phase 6.1: "Define Audit Event Taxonomy" with event catalog
**C.** Reference OWASP Logging Cheat Sheet for standard audit event patterns
**D.** All of the above (A, B, C)
**E.** (write in):

**Answer**: D (all three). Audit event taxonomy is security-critical, needs thorough documentation.

---

## 16. JOSE-JA Test Password Pattern Inconsistency

**Issue**: Cipher-IM tests likely use UUIDv7 for passwords but PLAN.md allows "cryptoutilMagic.TestPassword or UUIDv7".

**Context**:
- PLAN.md: "use cryptoutilMagic.TestPassword or UUIDv7"
- UUIDv7 passwords are random, hard to reproduce test failures
- cryptoutilMagic.TestPassword is deterministic, easier debugging
- NO guidance on WHEN to use TestPassword vs UUIDv7

**Question**: Which password pattern should tests use?

**A.** ALWAYS cryptoutilMagic.TestPassword (deterministic, reproducible)
**B.** ALWAYS UUIDv7 (realistic, high entropy)
**C.** TestPassword for positive tests, UUIDv7 for invalid password tests
**D.** Add to ARCHITECTURE.md: "Tests MUST use cryptoutilMagic.TestPassword for valid credentials"
**E.** (write in):

**Answer**: D (document pattern in ARCHITECTURE.md). TestPassword for valid credentials, UUIDv7 for invalid/random data.

---

## 17. PostgreSQL 18 Requirement Not Reflected in Dockerfiles

**Issue**: PLAN.md requires "PostgreSQL 18+" but Dockerfiles likely use older versions.

**Context**:
- PLAN.md Critical Fixes: "✅ PostgreSQL 18+ requirement (was incorrectly 16+)"
- Existing Dockerfiles may reference `postgres:16-alpine`
- Docker Compose files may specify older versions
- NO subtask for updating Docker image references

**Question**: Which phase should update PostgreSQL Docker image versions?

**A.** Phase 0 (Service-Template) - update ALL Docker files globally
**B.** Phase 2 (JOSE DB Schema) - update JOSE-specific Docker files only
**C.** Phase 8 (JOSE E2E Testing) - update when E2E fails on old PostgreSQL
**D.** Add Phase 0.0: "Update PostgreSQL Docker Images to 18+" before ALL other work
**E.** (write in):

**Answer**: D (Phase 0.0 prerequisite). PostgreSQL 18+ is project-wide requirement, should be fixed globally before service-specific work.

---

## 18. OTLP Configuration Not Shown in JOSE-JA Config

**Issue**: PLAN.md mandates "OTLP only" but NO example JOSE-JA config showing OTLP endpoints.

**Context**:
- PLAN.md Critical Fixes: "✅ OTLP only (removed Prometheus scraping endpoint)"
- `configs/jose-ja/jose-ja-server.yaml` likely empty or missing OTLP section
- ARCHITECTURE.md mentions "OTLP → otel-collector-contrib → Grafana LGTM" but NO config example
- NO subtask for creating OTLP config in Phase 2 or 3

**Question**: When should JOSE-JA OTLP config be created?

**A.** Phase 0 (Service-Template) - create template OTLP config for ALL services
**B.** Phase 2 (JOSE DB Schema) - create JOSE-specific OTLP config
**C.** Phase 3 (JOSE ServerBuilder) - configure OTLP during builder integration
**D.** Add subtask 3.0.1: "Create configs/jose-ja/jose-ja-server.yaml with OTLP settings"
**E.** (write in):

**Answer**: D (subtask 3.0.1 in Phase 3). ServerBuilder integration is logical time to establish complete config including OTLP.

---

## 19. Cipher-IM as Template Validation Insufficient

**Issue**: PLAN.md Phase 1 "Cipher-IM" is meant to validate Service-Template pattern but duration (3-4 days) is too short for thorough validation.

**Context**:
- PLAN.md Phase 1: "Adapt Cipher-IM to use ServerBuilder WITHOUT default tenant"
- Phase 1 duration: 3-4 days
- NO subtasks for comprehensive template validation (all ServerBuilder features exercised?)
- NO checklist for template pattern coverage (migrations? sessions? barrier? realms? registration?)

**Question**: What validation coverage should Phase 1 provide?

**A.** Phase 1 should only validate ServerBuilder integration (minimal scope)
**B.** Add Phase 1.X: "Template Pattern Validation" with comprehensive checklist
**C.** Extend Phase 1 duration to 5-7 days for thorough validation
**D.** Create docs/service-template/VALIDATION-CHECKLIST.md before Phase 1
**E.** (write in):

**Answer**: B and D (both needed). Phase 1 should validate ALL template features, needs explicit checklist and adequate time.

---

## 20. Key Rotation Schedule Not Specified

**Issue**: ARCHITECTURE.md mentions "quarterly" rotation for intermediate keys but JOSE-JA Elastic JWK rotation schedule is undefined.

**Context**:
- ARCHITECTURE.md: "Intermediate Keys: Encrypted with root key, rotated quarterly"
- PLAN.md Phase 4: "Elastic JWK Implementation" - NO rotation schedule
- JOSE JWKs are Content Keys in hierarchy, but rotation frequency undefined
- Per-message rotation vs periodic rotation not specified

**Question**: What rotation schedule should JOSE-JA Elastic JWKs use?

**A.** Per-message rotation (new Material Key per JWS/JWE operation)
**B.** Hourly rotation (new Material Key every hour)
**C.** Daily rotation (new Material Key daily at 00:00 UTC)
**D.** Add to ARCHITECTURE.md: "Content Keys: Rotated per-operation or hourly (configurable)"
**E.** (write in):

**Answer**: D (document in ARCHITECTURE.md). Per-message rotation is most secure but may be performance issue - should be configurable.

---

## Summary Statistics

**Total Issues**: 20
**Category Breakdown**:
- Design Conflicts: 8 (Default Tenant, TestMain, Session Configs, Realms, API Simplification, Cross-Tenant Access, Audit Events, Key Rotation)
- Missing Implementation Details: 7 (Registration Flow, Migration Numbering, OTLP Config, PostgreSQL 18, Docker Compose E2E, Mutation Classification, Password Pattern)
- Process Gaps: 5 (Path Migration Timing, E2E Test Location, Hardcoded Password Removal, Template Validation, Constant Cleanup)

**Severity Breakdown**:
- Critical (Blocks Implementation): 6 (Default Tenant Violation, Constants, Registration Flow, Path Migration, Cross-Tenant Design, Audit Taxonomy)
- High (Significant Rework): 8 (TestMain, Session Configs, Realm Filtering, API Simplification, Migration Numbering, PostgreSQL 18, Template Validation, Key Rotation)
- Medium (Process Improvement): 6 (E2E Location, Docker Compose, Mutation Classification, Password Pattern, OTLP Config, Constant Cleanup)

**Recommended Priority**:
1. Fix Default Tenant Violation (JOSE-JA + Cipher-IM) - CRITICAL
2. Add PostgreSQL 18 update to Phase 0.0
3. Add Phase 4.5: Cross-Tenant JWKS Authorization Design
4. Add Phase 6.1: Audit Event Taxonomy
5. Document Mutation Classification in ARCHITECTURE.md
6. Document Key Rotation Schedule in ARCHITECTURE.md
7. Expand Phase 1 with Template Validation Checklist
8. Add Hardcoded Password Removal to Phase 0
9. Delete Phase 7 (Path Migration) - fix paths from Phase 2
10. Create OTLP config subtask in Phase 3

---

## Next Steps

1. Review QUIZME findings with project stakeholders
2. Update PLAN.md with missing phases/subtasks
3. Update TASKS.md with detailed subtask breakdowns
4. Update ARCHITECTURE.md with missing design decisions
5. Create validation checklists before starting implementation
6. Re-estimate phase durations based on expanded scope
