# Research Option: PKI-CA-MIGRATE

**Option**: Migrate pki-ca to service-template (same pattern as cipher-im/jose-ja)
**Status**: Research Only (not yet selected)
**Created**: 2026-02-23
**Related**: docs/fixes-v7/research/tasks-PKI-CA-MIGRATE.md

---

## Overview

Migrate `internal/apps/pki/ca/` to fully use the service-template builder pattern,
matching how `cipher-im` and `jose-ja` are implemented, AFTER first completing all
consistency fixes in the 3 already-migrated services (cipher-im, jose-ja, sm-kms).

This option preserves pki-ca as a standalone service with its own deployment and
distinct identity in the product suite.

---

## Current State (Deep Research Findings)

### pki-ca Structure (~11,418 non-test LOC, 111 Go files)

**Already using template**:
- `server/server.go`: Uses `cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder`
- `ca.go`: Uses `cryptoutilTemplateCli.RouteService` (CLI routing consistent)
- `cmd/pki-ca/main.go`: Thin entrypoint (consistent pattern)
- Builder: `WithDomainMigrations`, `WithPublicRouteRegistration` — correct pattern

**NOT yet using template / gaps**:
- `storage/`: In-memory only (`MemoryStore`). No GORM repository. No SQLite/PostgreSQL persistence.
- `server/server.go:285-290`: `SetReady(true)` called from `ca.go` caServerStart before `Start()` — UNIQUE pattern vs cipher-im/jose-ja which call it from TestMain.
- `magic/`: Local magic package NOT consolidated to `internal/shared/magic/` (tracked in fixes-v7 Phase 3)
- `server/middleware/`: 3 files (vs cipher-im: 0 custom middleware) — may duplicate template middleware
- `server/cmd/`: Has its own `cmd.go` (purpose unclear vs template routing)
- Integration test: `server/server_integration_test.go` only (no `integration/` package like cipher-im)
- E2E: No `e2e/` directory
- TestMain: Raw polling loop (50 × 100ms), not using template helpers
- No `testing/` helper package (no `StartCAServer()` equivalent)

**Unique pki-ca components (must preserve)**:
- `compliance/`: CA/Browser Forum baseline requirement checks
- `crypto/`: CA-specific crypto primitives and provider interface
- `intermediate/`: Intermediate CA certificate management
- `observability/`: Certificate lifecycle observability hooks
- `profile/certificate/` + `profile/subject/`: Certificate profile and subject configuration
- `security/`: CA security policy enforcement
- `service/issuer/`: Certificate issuance (EST protocol support)
- `service/ra/`: Registration Authority workflow
- `service/revocation/`: CRL + OCSP services
- `service/timestamp/`: Time-Stamping Authority (RFC 3161)
- `cli/cli.go`: 492 LOC certificate generation CLI (key gen, self-signed CA, intermediate CA, end-entity) — SEPARATE from the server CLI routing
- `api/handler/`: EST protocol handler + OCSP handler (25 Go files total)
- `domain/`: CA domain models
- `bootstrap/`: CA bootstrapping

**Tasks 11-20 in pki-ca README are unimplemented**:
- Time-Stamping Authority (TSA) full support
- RA Workflows (approval queues, multi-approver)
- Profile Library (pre-defined profiles)
- Storage Layer (persistent — currently in-memory only)

---

## Service-Template Consistency Gaps Found in cipher-im, jose-ja, sm-kms

### cipher-im (reference implementation — closest to correct)
| Gap | Impact | Effort |
|-----|--------|--------|
| `testing/testmain_helper.go` startup helper is service-specific, not in template | Medium | 1h |
| `StartCipherIMService()` uses raw port polling (not template poll utility) | Low | 30min |
| No `e2e/` test for jose-ja or sm-kms (only cipher-im has it) | High | 2h each |

### jose-ja (migrated, but gaps)
| Gap | Impact | Effort |
|-----|--------|--------|
| No `testing/` helper package (no `StartJoseJAService()`) | Medium | 1h |
| TestMain uses raw 50×100ms polling loop | Low | 30min |
| No `e2e/` directory | High | 2h |
| `jwk_handler.go:358,368` + `jwk_handler_material.go:234,244,254,264`: JWK generation, signing, verification, encryption, decryption are UNIMPLEMENTED stubs | CRITICAL | 8-16h |
| `server/testmain_test.go` uses `InsecureSkipVerify:true` without using template TLS test client | Low | 30min |

### sm-kms (migrated, but has OLD-PATTERN debt)
| Gap | Impact | Effort |
|-----|--------|--------|
| `server/application/application_core.go`: Old-pattern wrapper around template core | High | 4h |
| `server/application/application_basic.go`: Old-pattern — likely duplicates template basic | High | 2h |
| `server/application/fiber_middleware_otel_request_logger.go`: May be in template | Medium | 1h |
| `server/middleware/` (15 non-test files): Custom JWT, claims, introspection, realm_context, scopes, service_auth, tenant, session — most likely duplicate template functionality | HIGH | 8-16h |
| `server/repository/orm/`: ORM repository exists but server.go has TODO "Migrate SQLRepository to template's ORM pattern" | High | 4h |
| `server.go:49`: TODO "Replace with template's GORM database and barrier" | High | 4h |
| No `e2e/` directory | High | 2h |
| No integration tests at all | High | 3h |

---

## Prerequisites (Before Migrating pki-ca)

Per ARCHITECTURE.md line 958 and quiz answer Q1=E, the following MUST be complete
before pki-ca migrates:

1. **fixes-v7 Phase 6**: cipher-im E2E reliable; jose-ja/sm-kms E2E unblocked; template has generic startup helper
2. **jose-ja critical TODOs**: JWK generation stubs must be implemented (otherwise jose-ja is not actually migrated)
3. **sm-kms migration debt**: Old application_core wrappers removed; middleware duplication resolved; ORM unified with template
4. **Template testing extraction**: `StartServiceFromConfig()` generic helper must be in template for pki-ca to inherit

---

## Migration Plan

### Phase A: Complete sm-kms Migration Debt (~12-20h)
Remove old-pattern wrappers and custom middleware that should come from template:
1. Audit sm-kms middleware vs template middleware — identify exact overlaps
2. Remove/replace `application_core.go` and `application_basic.go` old wrappers
3. Unify sm-kms ORM repository with template's GORM pattern
4. Remove or move custom JWT/session middleware to template (if generic enough)
5. Add sm-kms integration tests using template TestMain pattern
6. Add sm-kms E2E test suite

### Phase B: Complete jose-ja Critical TODOs (~8-16h)
1. Implement actual JWK generation in `jwk_handler.go:358,368`
2. Implement sign/verify/encrypt/decrypt in `jwk_handler_material.go:234,244,254,264`
3. Add integration and E2E tests for jose-ja
4. Verify jose-ja is functionally complete (not just structurally migrated)

### Phase C: Migrate pki-ca (~16-24h)
1. Add GORM repository for certificate storage (SQLite + PostgreSQL backends)
   - Convert `MemoryStore` to GORM model + repository implementing `Store` interface
   - Domain migrations 2001+ in `server/repository/migrations/`
2. Add pki-ca `testing/` helper package (`StartCAServer()`)
3. Fix `SetReady(true)` startup sequence (align with cipher-im/jose-ja pattern)
4. Consolidate `magic/` package into `internal/shared/magic/`
5. Add pki-ca integration test suite (port pattern from cipher-im)
6. Add pki-ca E2E test suite (uses template `e2e_infra` ComposeManager)
7. Update CI E2E workflow to include pki-ca E2E tests
8. Validate all pki-ca unique components still function correctly

---

## OSS Tooling for Service-Template Consistency Enforcement

| Tool | Use Case | Notes |
|------|----------|-------|
| `go ast` / `go/analysis` | Custom linter to enforce template usage patterns | Could detect services not using `RouteService`, `NewServerBuilder`, etc. |
| `gomodguard` | Block forbidden imports (e.g. old application_core) | Already configured in `.golangci.yml` |
| `gocritic` | Code pattern checks | Generic, not template-specific |
| `revive` | Custom rules via plugin | Could enforce `WithDomainMigrations` usage |
| `staticcheck` | Detect deprecated usage patterns | Limited custom rule support |
| `custom cicd validator` | Best option: add new `go run ./cmd/cicd validate-service-template <path>` subcommand | Checks: RouteService usage, builder pattern, migration format, TestMain template |

**Recommended**: Custom `validate-service-template` subcommand in `cmd/cicd/` that checks:
- Service uses `cryptoutilTemplateCli.RouteService`
- Server uses `cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder`
- Migration files follow 2001+ numbering convention
- TestMain uses template test helpers (via import check)

---

## Effort Estimate

| Work Item | Estimated Effort |
|-----------|-----------------|
| Phase A: sm-kms migration debt | 12-20h |
| Phase B: jose-ja critical TODOs | 8-16h |
| Phase C: pki-ca template migration | 16-24h |
| Template testing extraction | 2h |
| Custom service-template validator | 4h |
| **Total** | **42-66h** |

---

## Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| sm-kms middleware removal breaks KMS authn/authz | High | Critical | Comprehensive test coverage before removal; audit each middleware against template |
| jose-ja JWK implementation takes longer than estimated | High | High | Time-box to 16h; escalate if still incomplete |
| GORM storage for pki-ca breaks certificate storage semantics | Medium | High | Keep `Store` interface; swap implementation only |
| pki-ca E2E requires living CA for certificate tests | Medium | Medium | Use self-signed test CA (already exists in cli/cli.go) |
| Migration breaks EST protocol support | Low | High | EST tests exist in api/handler/ — run after migration |

---

## Advantages of This Option

- ✅ Cleanest architecture: one pattern across all 4 services
- ✅ Follows ARCHITECTURE.md migration priority exactly
- ✅ pki-ca inherits all template reliabilities (barrier, session, TLS, migrations)
- ✅ Enables proper E2E testing with Docker Compose
- ✅ Persistent certificate storage (currently in-memory only)
- ✅ Reduces future maintenance burden

## Disadvantages

- ❌ Largest scope: requires completing sm-kms and jose-ja debt first
- ❌ Total effort 42-66h is significant
- ❌ jose-ja critical TODOs (unimplemented crypto) add risk
- ❌ pki-ca is most complex service (11,418 LOC) with many unique components

---

## Recommendation Score: ⭐⭐⭐⭐ (Strong Recommend)

This is the architecturally correct path. The migration debt in sm-kms and jose-ja
should be fixed regardless of what happens with pki-ca — they are pre-existing gaps
that represent false positives in the current "migration complete" claim.

The key insight is that completing sm-kms migration debt and jose-ja critical TODOs
is NOT optional — those gaps exist right now and must be addressed. With those done,
pki-ca migration becomes straightforward because it leverages well-tested patterns.
