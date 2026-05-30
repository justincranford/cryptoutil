# Implementation Plan â€” Framework v25: 8-to-7 PS-ID Consolidation

**Status**: Planning
**Created**: 2026-05-25
**Last Updated**: 2026-05-25
**Purpose**: Consolidate from 8 PS-IDs (4 products) to 7 PS-IDs (3 products) by merging
pki-ca APIs into sm-kms under a `/pki/` path prefix, then deleting pki-ca, the pki product,
and all associated deployment artifacts.

---

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- âœ… **Correctness**: ALL code must be functionally correct with comprehensive tests
- âœ… **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- âœ… **Thoroughness**: Evidence-based validation at every step
- âœ… **Reliability**: Quality gates enforced (â‰¥95%/98% coverage/mutation)
- âœ… **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- âœ… **Accuracy**: Changes must address root cause, not just symptoms
- âŒ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- âŒ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions:**
- âœ… **Fix issues immediately** â€” blockers must be resolved before advancing to the next phase
- âœ… **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- âœ… **NEVER skip**: Cannot mark phase or task or step complete with known issues

---

## Overview

Consolidate the Cryptoutil suite from:
- **Before** (post framework-v24): 4 products (sm, pki, identity, skeleton), 8 PS-IDs
- **After**: 3 products (sm, identity, skeleton), 7 PS-IDs

**Removed PS-IDs**: `pki-ca` (APIs merged into `sm-kms` under `/pki/` path prefix)
**Removed Product**: `pki` (no remaining PS-IDs after removing pki-ca)

The consolidation preserves 100% of existing pki-ca API surface. All 18 PKI endpoints
that existed in pki-ca become a new endpoint group in sm-kms under the `/pki/` path prefix.
No API functionality is deleted â€” only the deployment unit changes.

---

## Background

### Assumption: framework-v24 Already Complete

This plan assumes framework-v24 has been fully executed:
- `sm-kms` APIs merged into `sm-kms` (JOSE operations)
- `sm-kms` APIs merged into `sm-kms` (encrypted messaging)
- `sm-kms`, `sm-kms` PS-IDs deleted
- retired product deleted
- sm-kms domain migrations 2001â€“2008 in use

### Starting State (post framework-v24)

| PS-ID | Product | Port | Migration Range | Kept/Merged |
|-------|---------|------|-----------------|-------------|
| sm-kms | sm | 8000 | 2001â€“2999 | âœ… Kept (receives pki-ca merge) |
| pki-ca | pki | 8300 | 5001â€“5999 | âŒ Merged â†’ sm-kms |
| identity-authz | identity | 8400 | 6001â€“6999 | âœ… Kept |
| identity-idp | identity | 8500 | 7001â€“7999 | âœ… Kept |
| identity-rs | identity | 8600 | 8001â€“8999 | âœ… Kept |
| identity-rp | identity | 8700 | 9001â€“9999 | âœ… Kept |
| identity-spa | identity | 8800 | 10001â€“10999 | âœ… Kept |
| skeleton-template | skeleton | 8900 | 11001â€“11999 | âœ… Kept |

### What pki-ca Provides (All Ports to sm-kms)

pki-ca operates as a Certificate Authority: CA hierarchy management, certificate enrollment
(PKCS#10 CSR), X.509 certificate issuance, CRL generation, OCSP, EST protocol (RFC 7030),
RFC 3161 Timestamp Authority, and CA/Browser Forum compliance enforcement.

**pki-ca internal storage architecture:**
pki-ca uses **in-memory storage** (`storage/storage.go`, `sync.RWMutex` + Go maps) for
certificates. The database only has the `ca_items` demo table (migration 5001). This means
the migration work is lightweight â€” only 1 SQL migration file needs porting to sm-kms.

**pki-ca API endpoints (18 total, all new for sm-kms under `/pki/` prefix):**
- `GET  /pki/cas`                              â€” listCAs
- `GET  /pki/cas/{caID}`                       â€” getCA
- `GET  /pki/cas/{caID}/crl`                   â€” getCRL (DER-encoded CRL)
- `POST /pki/enrollments`                      â€” submitEnrollment (PKCS#10 CSR)
- `GET  /pki/enrollments/{requestID}`          â€” getEnrollmentStatus
- `GET  /pki/certificates`                     â€” listCertificates (paginated)
- `GET  /pki/certificates/{serialNumber}`      â€” getCertificate
- `GET  /pki/certificates/{serialNumber}/chain` â€” getCertificateChain
- `POST /pki/certificates/{serialNumber}/revoke` â€” revokeCertificate
- `GET  /pki/profiles`                         â€” listProfiles
- `GET  /pki/profiles/{profileID}`             â€” getProfile
- `POST /pki/ocsp`                             â€” handleOCSP (RFC 6960, binary)
- `GET  /pki/est/cacerts`                      â€” estCACerts (RFC 7030, PKCS#7)
- `POST /pki/est/simpleenroll`                 â€” estSimpleEnroll (PKCS#10 â†’ PKCS#7)
- `POST /pki/est/simplereenroll`               â€” estSimpleReenroll
- `POST /pki/est/serverkeygen`                 â€” estServerKeyGen
- `GET  /pki/est/csrattrs`                     â€” estCSRAttrs
- `POST /pki/tsa/timestamp`                    â€” tsaTimestamp (RFC 3161, binary)

**pki-ca domain objects (1 DB table, ported as new sm-kms migration 2009):**
- `ca_items` â€” minimal CA demo table (`id TEXT PK`, `tenant_id TEXT`, `created_at DATETIME`)

**pki-ca services (all ported to `internal/apps/sm-kms/server/pki/`):**
- `compliance/` â€” CA/Browser Forum certificate compliance checker
- `crypto/` â€” crypto provider (RSA, ECDSA, Ed25519 key generation)
- `domain-v2/` â€” GORM models (CAItem)
- `storage/` â€” in-memory certificate storage (sync.RWMutex + maps)
- `profile/certificate/` â€” X.509 certificate profile enforcement
- `profile/subject/` â€” subject DN profile enforcement
- `security/` â€” CSR validation and threat modeling
- `service/issuer/` â€” end-entity certificate issuance
- `service/ra/` â€” Registration Authority (CSR intake, validation, routing)
- `service/revocation/` â€” CRL generation and OCSP response signing
- `service/timestamp/` â€” RFC 3161 Timestamp Authority
- `observability/` â€” OpenTelemetry metrics for PKI operations
- `bootstrap/` â€” offline root CA bootstrapping utility
- `intermediate/` â€” intermediate CA operations
- `api/handler/` â€” HTTP handlers (handler.go, handler_certs.go, handler_est.go, handler_ocsp.go)

---

## Technical Context

- **Language**: Go 1.26.1
- **Framework**: `internal/apps-framework/service/` service builder
- **Database**: PostgreSQL (E2E) + SQLite in-memory (unit/integration)
- **PKI storage note**: pki-ca uses in-memory storage for certificates; only `ca_items` is DB-backed
- **Binary content types**: OCSP (application/ocsp-request), TSA (application/timestamp-query),
  EST (application/pkcs10, application/pkcs7-mime) â€” oapi-codegen handles these as `[]byte`
- **Test references**: [ENG-HANDBOOK Â§10](../../docs/ENG-HANDBOOK.md#10-testing-architecture), [Â§10.2](../../docs/ENG-HANDBOOK.md#102-unit-testing-strategy), [Â§10.3](../../docs/ENG-HANDBOOK.md#103-integration-testing-strategy), [Â§10.4](../../docs/ENG-HANDBOOK.md#104-e2e-testing-strategy), [Â§10.5](../../docs/ENG-HANDBOOK.md#105-mutation-testing-strategy)
- **Quality reference**: [ENG-HANDBOOK Â§11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates), [Â§11.3](../../docs/ENG-HANDBOOK.md#113-code-quality-standards)
- **Coding standards**: [ENG-HANDBOOK Â§14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards)

### Affected Files â€” Complete Enumeration

#### Phase 1: Port pki-ca packages to sm-kms

**New migration files (sm-kms domain 2009):**
```
internal/apps/sm-kms/server/repository/migrations/
  2009_pki_ca_items.up.sql
  2009_pki_ca_items.down.sql
```

**New pki sub-packages in sm-kms (ported from pki-ca):**
```
internal/apps/sm-kms/server/pki/
  compliance/
    checker.go
    checker_test.go
  crypto/
    provider.go
    provider_test.go
  storage/
    storage.go
    storage_test.go
  domain/
    model.go
    model_test.go
  profile/
    certificate/
      profile.go
      profile_test.go
    subject/
      profile.go
      profile_test.go
  security/
    validator.go
    validator_test.go
    threat_model.go
    threat_model_test.go
  service/
    issuer/
      issuer.go
      issuer_test.go
    ra/
      ra.go
      ra_test.go
    revocation/
      revocation.go
      revocation_test.go
    timestamp/
      timestamp.go
      timestamp_test.go
  observability/
    metrics.go
    metrics_test.go
  bootstrap/
    bootstrap.go
    bootstrap_test.go
  intermediate/
    intermediate.go
    intermediate_test.go
  handler/
    handler.go
    handler_certs.go
    handler_est.go
    handler_ocsp.go
    handler_test.go
    handler_certs_test.go
    handler_est_test.go
    handler_ocsp_test.go
    testmain_test.go
```
Subtotal: ~40 new files in sm-kms pki sub-packages.

Note: pki-ca handler package is under `api/handler/` in pki-ca; ported to `server/pki/handler/`
in sm-kms for consistency with the sm-kms handler package conventions.

**Updated files for migration embedding:**
```
internal/apps/sm-kms/server/repository/migrations.go  (+embed 2009_pki_ca_items)
```

**Moved config files:**
```
configs/sm-kms/profiles/                               (moved from configs/pki-ca/profiles/)
```

#### Phase 2: Extend sm-kms API with PKI endpoints

**Updated OpenAPI spec files:**
```
api/sm-kms/openapi_spec_paths.yaml                    (+18 PKI paths under /pki/)
api/sm-kms/openapi_spec_components.yaml               (+PKI schemas from pki-ca spec)
api/sm-kms/openapi-gen_config_server.yaml             (+CSR, CA, CRL, OCSP, URI, SAN, DN, CN, OU)
api/sm-kms/openapi-gen_config_model.yaml              (+same initialisms)
api/sm-kms/openapi-gen_config_client.yaml             (+same initialisms)
```

**Regenerated API code:**
```
api/sm-kms/server/server.gen.go                       (regenerated)
api/sm-kms/model/models.gen.go                        (regenerated)
api/sm-kms/client/client.gen.go                       (regenerated)
```

**Updated server wiring:**
```
internal/apps/sm-kms/server/server.go                 (+PKI route registration)
```

#### Phase 3: Delete pki-ca service artifacts

**Deleted directories (entire trees):**
```
api/pki-ca/                                           (openapi spec + oapi-codegen configs + gen code)
internal/apps/pki-ca/                                 (~100 files: all pki-ca packages)
cmd/pki-ca/                                           (main.go)
configs/pki-ca/                                       (pki-ca-domain.yml, pki-ca-framework.yml, profiles/)
deployments/pki-ca/                                   (compose.yml, Dockerfile, configs/, secrets/)
```
Subtotal: ~110 files deleted.

**Updated magic file (remove pki-ca-specific TLS cert CNs):**
```
internal/shared/magic/magic_pki_tls.go                (remove AppPKICA* server cert CN constants)
```

#### Phase 4: Delete pki product and update registry

**Deleted directories/files:**
```
internal/apps/pki/                                    (pki.go, pki_test.go â€” 2 files)
cmd/pki/                                              (main.go â€” 1 file)
deployments/pki/                                      (compose.yml, secrets/ â€” 2 files)
internal/shared/magic/magic_pki.go                    (OTLPServicePKICA, PKIProductName, etc.)
```
Note: `magic_pki_ca.go` is **kept** â€” constants like `BackdateBuffer`, `HexBase`, `SerialNumberLength`
remain used by the ported pki packages in `internal/apps/sm-kms/server/pki/`.

**Updated registry and magic:**
```
api/cryptosuite-registry/registry.yaml                (remove pki product + pki-ca PS-ID)
internal/shared/magic/magic_tier.go                   (remove PKIProductName from ProductToPSIDs,
                                                        remove OTLPServicePKICA from AllPSIDs)
internal/apps-tools/cicd_lint/lint_fitness/registry/registry.go  (remove pki-ca entry)
```

**Updated cryptoutil suite router (remove pki routing):**
```
internal/apps/cryptoutil/                             (remove pki product delegation)
```

**Updated lint_ports (remove pki-ca port references):**
```
internal/apps-tools/cicd_lint/lint_ports/             (remove PKICAServicePort = 8300 references)
```

---

## Phases

**Phase Status Legend**: `â˜ TODO` | `ðŸ”„ IN PROGRESS` | `âœ… COMPLETE` | `â³ BLOCKED`

### Phase 1: Port pki-ca Packages to sm-kms (3d) [Status: â˜ TODO]

**Objective**: Copy all pki-ca internal packages to `internal/apps/sm-kms/server/pki/` with
updated package names and import paths. Add DB migration 2009 for ca_items. Move profile
configs to sm-kms. Verify the ported packages compile cleanly in their new location.

- Create migration 2009 (`ca_items`) in sm-kms domain
- Copy compliance, crypto, storage, domain, profile, security packages to `server/pki/`
- Copy service sub-packages (issuer, ra, revocation, timestamp) to `server/pki/service/`
- Copy observability, bootstrap, intermediate packages to `server/pki/`
- Copy api/handler files to `server/pki/handler/`
- Update all package declarations and cross-package imports to new paths
- Move `configs/pki-ca/profiles/` â†’ `configs/sm-kms/profiles/`
- Ensure all tests in new locations pass

**Success**: `go build ./internal/apps/sm-kms/...` clean; all ported package unit tests pass.
**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned â€” what worked,
what didn't, root causes, patterns. Evaluate artifacts for contradictions/omissions; create fix
tasks immediately.

### Phase 2: Extend sm-kms API with PKI Endpoints (2d) [Status: â˜ TODO]

**Objective**: Add all 18 pki-ca endpoints to the sm-kms OpenAPI spec under `/pki/` path prefix.
Update oapi-codegen configs with PKI initialisms. Regenerate API code. Wire PKI handlers into
the sm-kms server at `/service/api/v1/pki/...` and `/browser/api/v1/pki/...`.

- Port PKI schemas from `api/pki-ca/openapi_spec_enrollment.yaml` to sm-kms components spec
- Add PKI paths to sm-kms paths spec under `/pki/` prefix (following dual `/service/`+`/browser/` pattern)
- Add domain-specific initialisms to all three oapi-codegen configs: `CSR`, `CA`, `CRL`, `OCSP`, `URI`, `SAN`, `DN`, `CN`, `OU`
- Run `oapi-codegen` to regenerate server.gen.go, models.gen.go, client.gen.go
- Register PKI StrictServer handlers in `internal/apps/sm-kms/server/server.go`
- Add integration tests for PKI endpoint registration

**Success**: sm-kms builds clean with PKI endpoints registered; integration test hits
`GET /service/api/v1/pki/profiles` and gets 200 response.
**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned â€” what worked,
what didn't, root causes, patterns. Evaluate artifacts for contradictions/omissions; create fix
tasks immediately.

### Phase 3: Delete pki-ca Service Artifacts (1d) [Status: â˜ TODO]

**Objective**: Remove the entire pki-ca standalone service: API specs, all internal packages,
cmd entry point, config files, deployment artifacts. Remove pki-ca TLS cert CN constants from
magic. The ported packages in sm-kms replace all deleted functionality.

- Delete `api/pki-ca/` (openapi spec, oapi-codegen configs, generated code)
- Delete `internal/apps/pki-ca/` (entire ~100-file package tree)
- Delete `cmd/pki-ca/main.go`
- Delete `configs/pki-ca/` (domain config, framework config, profiles/)
- Delete `deployments/pki-ca/` (compose.yml, Dockerfile, configs/, secrets/, ca-crls/, certs/)
- Remove `AppPKICA*ServerCertCN` constants from `internal/shared/magic/magic_pki_tls.go`

**Success**: `go build ./...` clean (no pki-ca imports); `go build -tags e2e,integration ./...` clean.
**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned â€” what worked,
what didn't, root causes, patterns. Evaluate artifacts for contradictions/omissions; create fix
tasks immediately.

### Phase 4: Delete pki Product and Update Registry (0.5d) [Status: â˜ TODO]

**Objective**: Remove the pki product: CLI router, deployment artifacts, magic constants, entity
registry entry. Update magic_tier.go to reflect the new 7-PS-ID / 3-product topology. Update
lint_fitness entity registry and lint_ports to remove pki-ca references.

- Delete `internal/apps/pki/` (pki.go, pki_test.go)
- Delete `cmd/pki/main.go`
- Delete `deployments/pki/` (compose.yml, secrets/)
- Delete `internal/shared/magic/magic_pki.go` (OTLPServicePKICA, PKIProductName, PKICAServicePort, etc.)
- Update `internal/shared/magic/magic_tier.go` (remove PKIProductName + OTLPServicePKICA)
- Update `api/cryptosuite-registry/registry.yaml` (remove pki product + pki-ca PS-ID entry)
- Update `internal/apps-tools/cicd_lint/lint_fitness/registry/registry.go` (remove pki-ca)
- Update `internal/apps/cryptoutil/cryptoutil.go` (remove pki product delegation)
- Update `internal/apps-tools/cicd_lint/lint_ports/` (remove PKICAServicePort = 8300 references)
- Verify `go run ./cmd/cicd-lint lint-fitness` passes with updated registry

**Success**: `go run ./cmd/cicd-lint lint-fitness` passes; registry shows 3 products / 7 PS-IDs.
**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned â€” what worked,
what didn't, root causes, patterns. Evaluate artifacts for contradictions/omissions; create fix
tasks immediately.

### Phase 5: Quality Gates (0.5d) [Status: â˜ TODO]

**Objective**: Verify all quality gates pass end-to-end. Full build (with and without E2E tags),
golangci-lint, unit/integration tests, coverage, cicd-lint suite.

- `go build ./...` clean
- `go build -tags e2e,integration ./...` clean
- `golangci-lint run ./...` clean
- `golangci-lint run --build-tags e2e,integration ./...` clean
- `go test ./...` 100% passing, zero skips
- Coverage â‰¥95% for all sm-kms production packages
- `go run ./cmd/cicd-lint lint-text lint-go lint-go-test lint-go-mod lint-fitness lint-deployments lint-openapi` passes
- `go test -race -count=2 ./internal/apps/sm-kms/...` passes

**Success**: All CI/CD gates green; zero new linting violations; coverage maintained or improved.
**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned â€” what worked,
what didn't, root causes, patterns. Evaluate artifacts for contradictions/omissions; create fix
tasks immediately.

### Phase 6: Knowledge Propagation (0.5d) [Status: â˜ TODO]

**Objective**: Apply lessons learned across all permanent artifacts. Review lessons.md from all
prior phases. Update ENG-HANDBOOK.md, agents, skills, instructions, code, tests, workflows,
and docs where warranted.

- Review all lessons.md phase entries
- Update ENG-HANDBOOK.md (service consolidation patterns, in-memory storage migration notes)
- Update agents and skills where pki-ca references remain
- Update docs/ENG-HANDBOOK.md service catalog (remove pki-ca, update sm-kms API list)
- Verify propagation: `go run ./cmd/cicd-lint lint-docs`
- Commit all artifact updates with separate semantic commits per artifact type

**Success**: `go run ./cmd/cicd-lint lint-docs` passes; all handbook references to pki-ca removed.
**Post-Mortem**: Final lessons captured in lessons.md Executive Summary and Actions sections.

---

## Executive Decisions

### Decision 1: PKI Path Prefix in sm-kms

**Options**:
- A: No prefix â€” mount PKI endpoints directly at `/service/api/v1/cas`, `/service/api/v1/certificates`, etc.
- B: `/pki/` prefix â€” mount at `/service/api/v1/pki/cas`, etc. âœ“ **SELECTED**
- C: `/ca/` prefix â€” shorter but less descriptive
- D: Keep `/api/v1/ca/` prefix from pki-ca original spec

**Decision**: Option B â€” `/pki/` prefix.
**Rationale**: Option A risks naming conflicts with future KMS resources. Option B clearly namespaces
all PKI resources and aligns with the principle of no service name in paths while still
differentiating PKI from KMS operations. The `/pki/` prefix matches the product name and is
easily discoverable in documentation.

### Decision 2: pki-ca Storage Migration Strategy

**Options**:
- A: Keep in-memory storage â€” port `storage/storage.go` as-is to sm-kms
- B: Migrate to database-backed storage â€” convert to GORM models stored in PostgreSQL âœ“ **DEFERRED**
- C: Hybrid â€” keep in-memory but add persistence for critical state

**Decision**: Option A for framework-v25. Database-backed storage is a separate feature.
**Rationale**: The goal is service consolidation, not architecture redesign. Changing storage
architecture during a migration multiplies risk and scope. In-memory storage works for the current
use case. A separate plan can address DB-backed cert storage as a future enhancement.

### Decision 3: Package Organization in sm-kms

**Options**:
- A: Flat â€” all pki files directly in `internal/apps/sm-kms/server/pki*.go`
- B: Sub-package tree â€” `internal/apps/sm-kms/server/pki/` mirroring pki-ca structure âœ“ **SELECTED**
- C: New abstraction â€” redesign package boundaries during porting

**Decision**: Option B â€” mirror pki-ca sub-package structure under `server/pki/`.
**Rationale**: Minimizes cognitive load and merge conflicts. The existing pki-ca package boundaries
were designed for the same functionality. Introducing new abstractions during a migration risks
regressions and delays. Option C is deferred to a future refactor after consolidation stabilizes.

### Decision 4: magic_pki_ca.go Constants

**Options**:
- A: Delete `magic_pki_ca.go` â€” move constants (`BackdateBuffer`, `HexBase`, etc.) inline into pki packages
- B: Keep `magic_pki_ca.go` as-is â€” constants remain in magic package, still used by ported pki packages âœ“ **SELECTED**
- C: Rename constants with PKI prefix â€” `PKIBackdateBuffer`, `PKIHexBase`, etc.

**Decision**: Option B â€” keep `magic_pki_ca.go` in the magic package.
**Rationale**: The ported packages under `internal/apps/sm-kms/server/pki/` will still import
`cryptoutilSharedMagic`. Keeping these constants in magic avoids unnecessary churn. Renaming (Option C)
is valid but out of scope for a migration plan.

### Decision 5: pki-init After pki Product Deletion

**Options**:
- A: Delete pki-init entirely (it was only called via `pki init pki-init`)
- B: Keep framework TLS init, remove pki product CLI routing only âœ“ **SELECTED**
- C: Re-route pki-init through sm product CLI

**Decision**: Option B â€” keep `internal/apps-framework/tls` and `magic_pkiinit.go` intact.
**Rationale**: `internal/apps-framework/tls.InitForProduct()` is a shared framework utility used
by ALL products (sm, identity, skeleton). Deleting it would break TLS cert generation for all
remaining services. Only the pki product CLI entry point (`internal/apps/pki/pki.go`) is deleted.
sm-kms already has its own pki-init via `sm init pki-init`. No new routing is needed.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Import path cascade (100 files to update) | High | Medium | Use IDE refactor-rename; grep verify post-change |
| Binary content-type handlers (OCSP, TSA, EST) | Medium | High | Test with real binary fixtures; verify oapi-codegen output |
| sm-kms startup grows with CA initialization | Medium | Medium | Profile startup time before/after; add benchmark test |
| cicd-lint lint-fitness fails on removed PS-ID | Low | Medium | Update registry.go in same phase as deletion |
| Circular import after package reorganization | Low | High | Verify with `go build` after each package port |
| magic constant conflicts (same name in different magic files) | Low | Low | Search magic package before adding any new constant |

---

## Quality Gates - MANDATORY

**Per-Action Quality Gates**:
- âœ… All tests pass (`go test ./...`) â€” 100% passing, zero skips
- âœ… Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`) â€” zero errors
- âœ… Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`) â€” zero warnings
- âœ… No new TODOs without tracking in tasks.md

**Coverage Targets (from copilot instructions)**:
- âœ… Production code: â‰¥95% line coverage
- âœ… Infrastructure/utility code: â‰¥98% line coverage
- âœ… `internal/shared/magic/` excluded from coverage (constants only)
- âœ… Generated code excluded (`api/sm-kms/{server,model,client}/*.gen.go`)

**Per-Phase Quality Gates**:
- âœ… Unit + integration tests complete before moving to next phase
- âœ… Deployment validators pass (`go run ./cmd/cicd-lint lint-deployments`) after Phase 3+4
- âœ… Race detector clean (`go test -race -count=2 ./...`)

---

## Success Criteria

- [ ] All phases complete with evidence
- [ ] sm-kms serves all 18 PKI endpoints under `/service/api/v1/pki/` and `/browser/api/v1/pki/`
- [ ] All pki-ca artifacts deleted (no orphaned files)
- [ ] Registry reflects 3 products / 7 PS-IDs
- [ ] All quality gates passing
- [ ] `go run ./cmd/cicd-lint lint-fitness` reports 3 products / 7 PS-IDs
- [ ] No references to `pki-ca` or `PKIProductName` remain in non-historical code
- [ ] Documentation updated

---

## ENG-HANDBOOK.md Cross-References - MANDATORY

| Topic | ENG-HANDBOOK.md Section | When to Reference |
|-------|------------------------|-------------------|
| Testing Strategy | [Section 10](../../docs/ENG-HANDBOOK.md#10-testing-architecture) | ALL phases with implementation |
| Unit Testing | [Section 10.2](../../docs/ENG-HANDBOOK.md#102-unit-testing-strategy) | Phases 1-2 (porting + API) |
| Integration Testing | [Section 10.3](../../docs/ENG-HANDBOOK.md#103-integration-testing-strategy) | Phase 2 (route wiring tests) |
| Quality Gates | [Section 11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) | ALL phases (mandatory) |
| Code Quality | [Section 11.3](../../docs/ENG-HANDBOOK.md#113-code-quality-standards) | Phases 1-2 (new code) |
| Coding Standards | [Section 14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards) | Phases 1-2 |
| Service Template | [Section 5.1](../../docs/ENG-HANDBOOK.md#51-service-framework-pattern) | Phase 2 (route registration) |
| OpenAPI Architecture | [Section 8](../../docs/ENG-HANDBOOK.md#8-api-architecture) | Phase 2 (OpenAPI spec update) |
| Security Architecture | [Section 6](../../docs/ENG-HANDBOOK.md#6-security-architecture) | Phases 1-2 (PKI/cert handling) |
| Deployment Architecture | [Section 12](../../docs/ENG-HANDBOOK.md#12-deployment-architecture) | Phases 3-4 (deletion) |
| Version Control | [Section 14.2](../../docs/ENG-HANDBOOK.md#142-version-control) | ALL phases (commit strategy) |
| Plan Lifecycle | [Section 14.6](../../docs/ENG-HANDBOOK.md#146-plan-lifecycle-management) | ALL phases |
| Post-Mortem & Knowledge Propagation | [Section 14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) | Every phase + Phase 6 |
