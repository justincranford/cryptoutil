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
- **Success**: Generator produces directory tree matching tls-structure.md examples exactly.
- **Post-Mortem**: After quality gates pass, update lessons.md.

### Phase 3: pki-init CLI & Docker Volume Config (4h) [Status: ☐ TODO]

**Objective**: Update pki-init CLI and Docker volume configuration.
- Update pki-init subcommand to use new generator output.
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
- Verify directory counts match tls-structure.md (120 per PS-ID, 876 per SUITE).
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
  - ENG-HANDBOOK.md §10.3.4 vs §10.3.7: Pre-existing `InsecureSkipVerify` contradiction — §10.3.4 example uses `InsecureSkipVerify: true` while §10.3.7 says NEVER use it. Update §10.3.4 example to use `RootCAs` with `TLSRootCAPool()` instead.
- **Success**: All artifact updates committed; propagation check passes.

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| PKCS#12 library compatibility | Low | Medium | Use well-maintained `go-pkcs12` library; verify CGO-free |
| Directory count explosion at SUITE level (876 dirs) | Medium | Low | Verify with automated count tests; optimize generation speed |
| Realm values not yet finalized | Medium | Medium | Use `file`, `db` as defaults; parameterize for easy change (quizme-v2 pending) |
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
- [ ] Directory counts verified (120 per PS-ID, 876 per SUITE)
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
