# Implementation Plan - Framework V11: PKI-Init Cert Structure

**Status**: In Progress
**Created**: 2025-01-15
**Last Updated**: 2025-06-26
**Purpose**: Implement the new `/certs` volume directory structure in `pki-init`, matching the design defined in `docs/tls-structure.md`. All TLS wiring (PostgreSQL, OTel, Grafana) is deferred to framework-v12.

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions.**

## Overview

Framework V11 focuses exclusively on the pki-init certificate generation structure. The generator (`internal/apps/framework/tls/generator.go`) currently uses the old `ALL-*` prefix naming convention with raw PEM files and nested 2-level directories. This plan updates it to match the new design in `docs/tls-structure.md`:

- **New naming**: `public-global-*`, `public-{PS-ID}-*`, `private-{PS-ID}-*`, `public-postgres-*`, `public-{grafana-otel-lgtm,otel-collector-contrib}-*`
- **Keystore/truststore pattern**: Separate `-keystore/` and `-truststore/` directories; keystores contain `.p12`, `.crt`, `.key`; truststores contain `.p12`, `.crt` only (never `.key`)
- **PKCS#12 support**: Added alongside PEM (`.p12` + `.crt` + `.key` per keystore)
- **Flat structure**: No nested entity subdirectories; directory name IS the file base name (`SAME-AS-DIR-NAME.{p12,crt,key}`)
- **14 categories**: Covering global CAs, per-PS-ID servers/clients, admin mTLS, Grafana/OTel, PostgreSQL

## Background

Previous framework versions established the service builder, barrier layer, and initial TLS cert generation. The generator currently produces 13 `ALL-*` shared directories and 6 per-PS-ID directories using `*-crt.pem` and `*-key.pem` file naming. V11 redesigns this to the keystore/truststore model documented in `tls-structure.md`.

## Decisions (Resolved from quizme-v1)

### Decision 1: Docker Volume Mount Strategy

**Options**:
- A: Bind-mount the full `/certs` tree with read-only everywhere
- B: Named volume per PS-ID (one volume, one service scope)
- C: Named volume per cert domain (fine-grained isolation)
- D: Named Docker volumes with least-privilege permissions ✓ **SELECTED**

**Decision**: Option D — Named Docker volumes with per-service volume scoping and least-privilege file permissions.

**Rationale**: Named volumes are Docker-native, portable across environments, and composable. Least-privilege permissions (440 for cert+key, 444 for CA certs) prevent lateral credential access between services. Volume-per-service scoping limits blast radius if one container is compromised.

**Impact**: Each PS-ID compose file mounts only the volumes it needs. `pki-init` sets file permissions during generation.

### Decision 2: PostgreSQL TLS Integration

**Options**:
- A: Full mTLS wiring in v11
- B: Server-only TLS in v11, client mTLS in v12
- C: Connection-string sslmode=verify-full in v11, mTLS in v12
- D: Defer all PostgreSQL TLS wiring to v12 ✓ **SELECTED**

**Decision**: Option D — Defer entirely to framework-v12.

**Rationale**: PostgreSQL TLS wiring requires config changes across all 10 PS-ID compose files, shared-postgres init scripts, PostgreSQL HBA rules, and GORM connection strings. This is a separate concern from cert structure generation and would double v11 scope without benefit to the cert generation correctness goal.

### Decision 3: OTel Collector TLS

**Options**:
- A: Full mTLS for app→OTel in v11
- B: Server-only TLS for OTel in v11
- C: TLS for OTel→Grafana only in v11
- D: Defer all OTel TLS wiring to v12 ✓ **SELECTED**

**Decision**: Option D — Defer entirely to framework-v12.

**Rationale**: OTel TLS requires otel-collector-contrib config changes, app-level OTLP exporter TLS config, and may require cert rotation awareness. Separate concern from cert generation.

### Decision 4: OTel → Grafana TLS

**Options**:
- A: Full mTLS for OTel→Grafana in v11
- B: Server-only TLS for Grafana in v11
- C: Gradual (server in v11, mTLS in v12)
- D: Defer all OTel→Grafana TLS wiring to v12 ✓ **SELECTED**

**Decision**: Option D — Defer entirely to framework-v12.

**Rationale**: Grafana LGTM OTLP ingest TLS requires the grafana-otel-lgtm container to accept mTLS connections, which may require custom entrypoint scripts or environment config. Separate concern.

### Decision 5: Grafana Stack

**Options**:
- A: Switch to separate Grafana+Loki+Tempo+Mimir components
- B: Switch to Grafana Alloy
- C: Keep grafana/otel-lgtm but add custom TLS config
- D: Keep grafana/otel-lgtm as-is ✓ **SELECTED**

**Decision**: Option D — Keep `grafana/otel-lgtm` as-is.

**Rationale**: The all-in-one image simplifies the dev/test stack. TLS for this image will be addressed in v12 when actual wiring is needed. No need to split the stack now.

### Decision 6: pki-init in shared-telemetry

**Options**:
- A: Add pki-init service to shared-telemetry compose
- B: No pki-init in shared-telemetry; rely on PS-ID compose pki-init services ✓ **SELECTED**

**Decision**: Option B — No pki-init in shared-telemetry.

**Rationale**: Each PS-ID compose already includes a `pki-init` service that generates all required certs (including Grafana/OTel/PostgreSQL certs). Adding a separate pki-init to shared-telemetry would create cert duplication and ownership confusion. Telemetry services mount volumes populated by the PS-ID's pki-init.

### Decision 7: PostgreSQL Client PKI Domains (Q1=A — quizme-v2)

**Options**:
- A: postgres-1 and postgres-2 share one client identity = 3 PKI domains (`sqlite-1`, `sqlite-2`, `postgres`) ✓ **SELECTED**
- B: postgres-1 and postgres-2 each have separate client identity = 4 PKI domains
- C: All 4 instances share one client identity = 1 PKI domain
- D: sqlite-1/sqlite-2 share; postgres-1/postgres-2 separate = 3 PKI domains (different grouping)

**Decision**: Option A — postgres-1 and postgres-2 share one client identity (3 PKI domains: `sqlite-1`, `sqlite-2`, `postgres`).

**Rationale**: postgres-1 and postgres-2 connect to the same PostgreSQL cluster as logically equivalent instances. Sharing client identity simplifies cert management and accurately reflects that both are essentially the same role within the PostgreSQL deployment.

**Impact**: Category 5 has 12 dirs per PS-ID (2 user types × 2 realms × 3 PKI domains × 1 store type). Confirmed in tls-structure.md.

### Decision 8: Realm Values (Q2=E — quizme-v2)

**Options**:
- A: Fixed defaults `file` and `db` (hardcoded)
- B: Read from config file at `pki-init` invocation
- C: Passed as CLI flags
- D: Derived from deployment tier
- E: Realm values are dynamic — each PS-ID declares its realms in `registry.yaml` ✓ **SELECTED**

**Decision**: Option E — Each PS-ID has a `realms` list in `registry.yaml`. Each realm entry has: `location` (file, db, federated), `type` (e.g., user/pass, mTLS, JWT), and `unique name`. All realm types are implemented by the framework and inherited; PS-IDs select which they use.

**Rationale**: Realm count and names vary per PS-ID. Hardcoding `file`/`db` makes pki-init inflexible. Using `registry.yaml` as the single source of truth for realm config is consistent with the registry's role as the canonical entity catalog.

**Impact** (significant):
- `registry.yaml` must gain a `realms` field per PS-ID entry.
- `pki-init` must read `registry.yaml` at runtime to determine realm names for Category 5 cert generation.
- Category 5 directory count becomes dynamic: `2 user types × |realms| × 3 PKI domains × 1 store type`.
- Phase 2 Task 2.4 and Phase 3 Task 3.1 must account for realm-driven generation.
- New tasks required: schema design for `realms` field, pki-init registry reading implementation.
- Examples in `tls-structure.md` (skeleton-template, sm) assume 2 realms (`file`, `db`) as representative defaults.

### Decision 9: `admin` Identity for Grafana/OTel (Q3=A — quizme-v2)

**Options**:
- A: `admin` = dedicated admin/ops user identity for directly accessing Grafana UI and OTel APIs ✓ **SELECTED**
- B: `admin` = a shared identity for all services
- C: `admin` = a system service account for internal use
- D: No `admin` identity — use per-PS-ID identities only

**Decision**: Option A — `admin` is a dedicated ops user identity for directly accessing Grafana UI (port 3000) and OTel Collector APIs (:4317/:4318). It is NOT a service account and NOT shared with PS-ID services.

**Rationale**: Human operators monitoring the system need a dedicated cert identity to authenticate to Grafana and OTel. This should be separate from PS-ID service identities to maintain auditability and least-privilege.

**Impact**: Category 9 consistently includes one `admin` entity alongside per-PS-ID entities. For skeleton-template at PS-ID scope: 2 services × (1 PS-ID + 1 admin) × 1 store type = 4 dirs.

### Decision 10: Cat 4 PKI Domain Structure (Q4=E)

**Question**: How many PKI domains for Cat 4 (PS-ID HTTPS Client CAs)?

**Options**:
- A: 4 PKI domains: `{sqlite,postgres} × {1,2}` = sqlite-1, sqlite-2, postgres-1, postgres-2
- B: 2 PKI domains: `{sqlite,postgres}`
- C: 3 PKI domains: `{sqlite-1,sqlite-2,postgres}` ✓ **SELECTED**
- D: 1 shared PKI domain for all instances
- E: *(custom)*

**Decision**: Option C — Cat 4 uses 3 PKI domains: `{sqlite-1,sqlite-2,postgres}`. postgres-1 and postgres-2 share the same backend database, unseal secrets, and username/password credentials → they share a client TLS identity too. This corrects the prior assumption of 4 PKI domains.

**Rationale**: Unlike Cat 6 (private admin channel, where each instance needs a unique identity for mTLS), the public HTTPS client CA authenticates to the shared logical postgres database. postgres-1 and postgres-2 access the same logical database; sharing a client TLS identity is correct and avoids unnecessary divergence.

**Impact**:
- Cat 4 drops from 16 dirs per PS-ID to 12 dirs (3 PKI domains × 2 CA tiers × 2 store types).
- Total per PS-ID: 82 dirs (was 86); total per SUITE: 568 dirs (was 608).
- skeleton-template example: 82 dirs total; sm example: 136 dirs total.
- Cat 6 is unchanged: stays at `{sqlite,postgres}-{1,2}` = 4 PKI domains (each app instance has its own private admin channel identity).
- `tls-structure.md` Required Logical Layout, directory count table, and unrolled examples updated accordingly.

### Decision 11: Certificate Algorithm (Q5=A)

**Question**: Which key algorithm for all certificates?

**Options**:
- A: ECDSA P-384 for ALL cert types (root CAs, issuing CAs, leaf certs) ✓ **SELECTED**
- B: RSA 3072 for root/issuing CAs; ECDSA P-384 for leaf certs
- C: ECDSA P-256 for all types
- D: RSA 4096 root, RSA 3072 issuing, ECDSA P-256 leaf
- E: *(custom)*

**Decision**: Option A — ECDSA P-384 for ALL cert types.

**Rationale**: Consistency across all tiers eliminates algorithm-mismatch errors. P-384 is FIPS 140-3 approved, CA/B Forum compliant, and provides strong 192-bit security. Per algorithm agility mandate, all implementations use configurable algorithms with P-384 as the hardcoded default.

**Impact**: `generator.go` defaults to ECDSA P-384 key generation for all cert types; configuration allows override per algorithm agility requirement.

### Decision 12: Certificate Validity Periods (Q6=A)

**Question**: What default validity periods to use?

**Options**:
- A: CA/B Forum strictly: root 20yr, issuing 5yr, leaf 398 days ✓ **SELECTED**
- B: Relaxed: root 25yr, issuing 10yr, leaf 2yr (dev-friendly)
- C: Short-lived: root 10yr, issuing 2yr, leaf 90 days (security-focused)
- D: No hardcoded defaults — must be fully configured
- E: *(custom)*

**Decision**: Option A — CA/B Forum strictly: root 20yr, issuing 5yr, leaf 398 days.

**Rationale**: CA/B Forum compliance minimizes friction if certs are ever tested against strict validators and encourages good hygiene. Per algorithm agility mandate, validity periods are configurable with these CA/B Forum values as defaults.

**Impact**: `generator.go` sets default validity periods per cert type; configuration allows override.

## Technical Context

- **Language**: Go 1.26.1
- **Generator**: `internal/apps/framework/tls/generator.go`
- **TLS Structure Doc**: `docs/tls-structure.md`
- **Registry**: `api/cryptosuite-registry/registry.yaml`
- **Dependencies**: `crypto/x509`, `software.sslmate.com/src/go-pkcs12` (for PKCS#12), standard library TLS

## Phases

### Phase 1: Cert Structure Documentation (2h) [Status: ✅ COMPLETE]

**Objective**: Update `docs/tls-structure.md` to fully specify the new directory layout.
- Updated Requirements section with 14 categories.
- Updated naming conventions (public/private prefix, keystore/truststore suffix).
- Added File Format Convention section.
- Added Directory Count Summary table.
- Added unrolled examples for skeleton-template (PS-ID) and sm (PRODUCT).
- Updated Policy Alignment section.
- **Success**: tls-structure.md is the complete specification for generator.go implementation.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 2: Generator Rewrite (8h) [Status: ☐ TODO]

**Objective**: Rewrite `generator.go` to produce the new directory structure.
- Replace `ALL-*` prefix naming with new `public-`/`private-` convention.
- Add keystore/truststore directory types with correct file sets.
- Add PKCS#12 generation alongside PEM.
- Implement all 14 categories from tls-structure.md.
- Implement `SAME-AS-DIR-NAME` file naming pattern.
- Refactor from nested 2-level dirs to flat structure.
- Handle all three deployment scopes: PS-ID, PRODUCT, SUITE.
- Category 5 directory count is realm-driven (read from `registry.yaml`): `2 user types × |realms| × 3 PKI domains × 1 store type`.
- **Success**: Generator produces directory tree matching tls-structure.md examples exactly (82 dirs per PS-ID with 2 realms).
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 3: pki-init CLI & Docker Volume Config (4h) [Status: ☐ TODO]

**Objective**: Update pki-init CLI and Docker volume configuration.
- Update pki-init subcommand to use new generator output.
- Implement reading of `realms` list from `registry.yaml` to drive Category 5 directory count.
- Configure named Docker volumes in compose templates.
- Set least-privilege file permissions (440 keystore, 444 truststore CA certs).
- Update compose file volume mount declarations.
- **Success**: `pki-init skeleton-template /tmp` produces correct tree; compose volumes mount correctly.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 4: Template & Deployment Updates (3h) [Status: ☐ TODO]

**Objective**: Update deployment templates to reference new cert paths.
- Update `deployment-templates.md` with new cert path references.
- Update `target-structure.md` with /certs volume layout reference.
- Update any compose file cert volume mount paths.
- Ensure `cicd-lint lint-fitness template-compliance` passes.
- **Success**: All deployment templates reference correct new cert paths; lint passes.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 5: Quality Gates & Testing (6h) [Status: ☐ TODO]

**Objective**: Comprehensive testing of the new cert structure.
- Unit tests for all 14 category directory generation.
- Unit tests for keystore vs truststore file sets.
- Unit tests for PKCS#12 generation.
- Integration tests for PS-ID, PRODUCT, SUITE scope generation.
- Verify directory counts match tls-structure.md (82 per PS-ID with 2 realms, 568 per SUITE with 2 realms).
- Coverage ≥95% for generator.go, ≥98% for utility functions.
- Mutation testing ≥95%.
- Race detector clean.
- **Success**: All quality gates pass; directory counts verified.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 6: Knowledge Propagation (2h) [Status: ☐ TODO]

**Objective**: Apply lessons learned to permanent artifacts.
- Review lessons.md from all prior phases.
- Update ENG-HANDBOOK.md with new TLS patterns if warranted.
- Update agents, skills, instructions where warranted.
- Verify propagation integrity (`go run ./cmd/cicd-lint lint-docs validate-propagation`).
- **Known documentation gaps to address** (from deep analysis):
  - ENG-HANDBOOK.md §6.11: Add cross-reference to `tls-structure.md` and the 14-category pki-init cert structure.
  - ENG-HANDBOOK.md §6.11: Add PKCS#12 (`.p12`) as a supported certificate format alongside PEM.
  - ENG-HANDBOOK.md §10.3.4 `InsecureSkipVerify` fix: Deferred to framework-v12 per Q7=E — address when cert wiring is complete in framework-v12 Phase 9.
- **Success**: All artifact updates committed; propagation check passes.

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| PKCS#12 library compatibility | Low | Medium | Use well-maintained `go-pkcs12` library; verify CGO-free |
| Directory count explosion at SUITE level | Medium | Low | Verified: 568 dirs per SUITE with 2 realms; automated count tests ensure correct generation. |
| Realm values not yet finalized | Medium | Medium | Decision 8 (Q2=E): realms defined per PS-ID in registry.yaml; use `file`, `db` as representative defaults in examples |
| registry.yaml realm schema design | Medium | High | New tasks in Phase 2/3: design schema field, implement reading in pki-init |
| Existing tests break from renamed directories | High | Medium | Update all test assertions systematically; use golden files |

## Quality Gates - MANDATORY

**Per-Phase Quality Gates**:
- ✅ All tests pass (`go test ./...`) — 100% passing, zero skips
- ✅ Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`) — zero errors
- ✅ Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`) — zero warnings
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets**:
- ✅ Production code: ≥95% line coverage
- ✅ Infrastructure/utility code: ≥98% line coverage

**Mutation Testing Targets**:
- ✅ Production code: ≥95%
- ✅ Infrastructure/utility code: ≥98%

## Success Criteria

- [ ] All 6 phases complete
- [ ] Generator produces directory tree matching tls-structure.md examples
- [ ] All quality gates passing
- [ ] Directory counts verified (82 per PS-ID with 2 realms, 568 per SUITE with 2 realms)
- [ ] Documentation updated
- [ ] CI/CD workflows green
- [ ] Evidence archived

## ENG-HANDBOOK.md Cross-References - MANDATORY

| Topic | Section | Relevance |
|-------|---------|-----------|
| Testing Strategy | [Section 10](../../docs/ENG-HANDBOOK.md#10-testing-architecture) | Unit + integration test patterns |
| Quality Gates | [Section 11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) | Coverage and mutation targets |
| Code Quality | [Section 11.3](../../docs/ENG-HANDBOOK.md#113-code-quality-standards) | Linter configuration |
| Coding Standards | [Section 14.1](../../docs/ENG-HANDBOOK.md#141-coding-standards) | Go patterns |
| Version Control | [Section 14.2](../../docs/ENG-HANDBOOK.md#142-version-control) | Commit strategy |
| Security Architecture | [Section 6](../../docs/ENG-HANDBOOK.md#6-security-architecture) | PKI and TLS patterns |
| PKI Architecture | [Section 6.5](../../docs/ENG-HANDBOOK.md#65-pki-architecture--strategy) | CA hierarchy and cert lifecycle |
| Plan Lifecycle | [Section 14.6](../../docs/ENG-HANDBOOK.md#146-plan-lifecycle-management) | Plan management |
| Post-Mortem | [Section 14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) | Phase lessons |
