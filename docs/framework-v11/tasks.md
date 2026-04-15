# Tasks - Framework V11: PKI-Init Cert Structure

**Status**: 2 of 26 tasks complete (8%)
**Last Updated**: 2025-06-26
**Created**: 2025-01-15

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

---

## Task Checklist

### Phase 1: Cert Structure Documentation [Status: ✅ COMPLETE]

**Phase Objective**: Fully specify the new `/certs` directory layout in `docs/tls-structure.md`.

#### Task 1.1: Update tls-structure.md Requirements

- **Status**: ✅
- **Estimated**: 1h
- **Actual**: 1h
- **Dependencies**: None
- **Description**: Rewrite Requirements section with 14 named categories matching the parametric layout.
- **Acceptance Criteria**:
  - [x] 14 categories described with directory counts
  - [x] Per-PS-ID, per-PRODUCT, per-SUITE totals documented
  - [x] File Format Convention section added
  - [x] Directory Count Summary table added

#### Task 1.2: Add Unrolled Examples

- **Status**: ✅
- **Estimated**: 1h
- **Actual**: 1h
- **Dependencies**: Task 1.1
- **Description**: Add concrete examples for skeleton-template (PS-ID) and sm (PRODUCT).
- **Acceptance Criteria**:
  - [x] skeleton-template example: 86 directories fully listed (quizme-v2: removed end-entity truststores)
  - [x] sm example: 144 directories fully listed (2 PS-IDs: sm-kms, sm-im)
  - [x] Category comments in each listing
  - [x] Policy Alignment section updated

---

### Phase 2: Generator Rewrite [Status: ☐ TODO]

**Phase Objective**: Rewrite `internal/apps/framework/tls/generator.go` to produce the new directory structure matching `docs/tls-structure.md`.

#### Task 2.1: Refactor Directory Naming

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Phase 1 complete
- **Description**: Replace `ALL-*` prefix naming with new `public-`/`private-` convention. Replace nested 2-level dirs with flat structure. Implement `SAME-AS-DIR-NAME` file naming.
- **Acceptance Criteria**:
  - [ ] All `ALL-*` directory names replaced
  - [ ] `public-global-*`, `public-{PS-ID}-*`, `private-{PS-ID}-*` naming implemented
  - [ ] `public-postgres-*`, `public-{grafana-otel-lgtm,otel-collector-contrib}-*` naming implemented
  - [ ] Flat directory structure (no nested entity subdirs)
  - [ ] Tests pass: `go test ./internal/apps/framework/tls/...`
- **Files**:
  - `internal/apps/framework/tls/generator.go`
  - `internal/apps/framework/tls/generator_test.go`

#### Task 2.2: Implement Keystore/Truststore Pattern

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 2.1
- **Description**: Implement separate `-keystore/` and `-truststore/` directory types. Keystores get `.p12`, `.crt`, `.key`; truststores get `.p12`, `.crt` only.
- **Acceptance Criteria**:
  - [ ] Keystore directories contain 3 files
  - [ ] Truststore directories contain 2 files (NO `.key`)
  - [ ] `SAME-AS-DIR-NAME` file naming pattern applied
  - [ ] Tests validate file set per directory type

#### Task 2.3: Add PKCS#12 Generation

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 2.2
- **Description**: Generate PKCS#12 bundles (`.p12`) alongside PEM files. Use CGO-free library.
- **Acceptance Criteria**:
  - [ ] `.p12` files generated for every keystore and truststore
  - [ ] PKCS#12 bundles loadable with standard tools
  - [ ] CGO_ENABLED=0 build passes
  - [ ] Library added to go.mod (e.g., `software.sslmate.com/src/go-pkcs12`)

#### Task 2.4: Implement All 14 Categories

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Tasks 2.1, 2.2, 2.3
- **Description**: Implement generation for all 14 categories from tls-structure.md. Handle PS-ID, PRODUCT, and SUITE deployment scopes.
- **Acceptance Criteria**:
  - [ ] Categories 1-14 all generate correct directories
  - [ ] PS-ID scope: 86 directories (skeleton-template example — assumes 2 realms)
  - [ ] PRODUCT scope: correct count (sm = 144 — 2 PS-IDs × 72 each with 2 realms)
  - [ ] SUITE scope: 608 directories (10 PS-IDs × 2 realms default)
  - [ ] Realm values read dynamically from registry.yaml (see Task 2.5)

#### Task 2.5: Design Realm Schema in registry.yaml

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 2.4
- **Description**: Define the `realms` field structure in `registry.yaml`. Each PS-ID entry gains a `realms` list; each realm has `location` (file/db/federated), `type` (user-pass/mtls/jwt/etc.), and a unique `name`. The framework defines the full realm type catalogue; PS-IDs select which realms they activate. Skeleton-template uses `file` and `db` as representative defaults.
- **Acceptance Criteria**:
  - [ ] `realms` field schema designed and documented in registry.yaml
  - [ ] skeleton-template entry populated with `file` and `db` realm entries
  - [ ] All 10 PS-ID entries have valid `realms` values
  - [ ] Schema description added to `api/cryptosuite-registry/` docs
- **Files**:
  - `api/cryptosuite-registry/registry.yaml`

---

### Phase 3: pki-init CLI & Docker Volume Config [Status: ☐ TODO]

**Phase Objective**: Update pki-init CLI and Docker volume configuration for the new cert structure.

#### Task 3.1: Update pki-init Subcommand

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Phase 2 complete
- **Description**: Update pki-init CLI to use new generator output. Verify command-line interface unchanged (`pki-init <tier-id> <target-dir>`).
- **Acceptance Criteria**:
  - [ ] `pki-init skeleton-template /tmp` produces correct 86-dir tree (2 realms)
  - [ ] `pki-init sm /tmp` produces correct 144-dir tree (2 PS-IDs, 2 realms)
  - [ ] `pki-init cryptoutil /tmp` produces correct 608-dir tree (10 PS-IDs, 2 realms)
  - [ ] Empty/non-empty target directory check still works
  - [ ] Tests pass for all three scopes

#### Task 3.2: Set Least-Privilege File Permissions

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 3.1
- **Description**: Set file permissions during generation: 440 for keystore files (cert+key), 444 for truststore files (public certs only).
- **Acceptance Criteria**:
  - [ ] Keystore `.key` and `.p12` files: 0440
  - [ ] Truststore `.crt` and `.p12` files: 0444
  - [ ] Keystore `.crt` files: 0444
  - [ ] Permissions verified in tests

#### Task 3.3: Configure Named Docker Volumes

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 3.1
- **Description**: Update compose templates to declare named Docker volumes for cert storage. Each PS-ID service mounts only its required volumes.
- **Acceptance Criteria**:
  - [ ] Named volumes declared in compose templates
  - [ ] Volume scoping: each service mounts only its certs
  - [ ] `cicd-lint lint-deployments` passes

#### Task 3.4: Implement registry.yaml Realm Reading

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Tasks 3.1, 2.5
- **Description**: Implement `registry.yaml` reading in pki-init. At startup, pki-init resolves the realm list for the requested tier-id from registry.yaml. Category 5 directory count becomes dynamic: `2 user types × |realms| × 3 PKI domains × 1 store type` per PS-ID instead of hardcoded `file`/`db`.
- **Acceptance Criteria**:
  - [ ] pki-init reads realm list from registry.yaml per PS-ID
  - [ ] Category 5 count equals `2 × |realms| × 3` per PS-ID
  - [ ] Missing or empty realm list returns a clear error
  - [ ] Tests cover 1-realm, 2-realm, and 3-realm scenarios
- **Files**:
  - `internal/apps/framework/tls/generator.go`
  - `internal/apps/pki-ca/` (or equivalent pki-init CLI entrypoint)

---

### Phase 4: Template & Deployment Updates [Status: ☐ TODO]

**Phase Objective**: Update deployment templates and documentation to reference new cert paths.

#### Task 4.1: Update deployment-templates.md

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 3 complete
- **Description**: Update `docs/deployment-templates.md` with notes about new cert path structure and v12 TLS wiring plans.
- **Acceptance Criteria**:
  - [ ] Cert volume mount paths updated
  - [ ] v12 TLS wiring noted as planned
  - [ ] Existing template content not broken

#### Task 4.2: Update target-structure.md

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 4.1
- **Description**: Add `/certs` volume layout reference to `docs/target-structure.md`.
- **Acceptance Criteria**:
  - [ ] Section F.4 references tls-structure.md for cert layout
  - [ ] No duplicate content (cross-reference only)

#### Task 4.3: Update Compose Cert Volume Mounts

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Tasks 3.3, 4.1
- **Description**: Update actual compose files with new cert volume mount paths.
- **Acceptance Criteria**:
  - [ ] All PS-ID compose files updated
  - [ ] Template compliance linter passes
  - [ ] `cicd-lint lint-deployments` passes

#### Task 4.4: Verify Template Compliance

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Tasks 4.1, 4.2, 4.3
- **Description**: Run `cicd-lint lint-fitness template-compliance` to verify all deployment artifacts match canonical templates.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes
  - [ ] No template drift detected

---

### Phase 5: Quality Gates & Testing [Status: ☐ TODO]

**Phase Objective**: Comprehensive testing and quality verification.

#### Task 5.1: Unit Tests for Generator

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Phase 2 complete
- **Description**: Table-driven unit tests for all 14 directory categories. Test keystore vs truststore file sets. Test PKCS#12 generation.
- **Acceptance Criteria**:
  - [ ] One test case per category
  - [ ] Keystore file count = 3, truststore file count = 2
  - [ ] PKCS#12 loadable after generation
  - [ ] Coverage ≥95% for generator.go
  - [ ] `t.Parallel()` on all tests

#### Task 5.2: Integration Tests for Scope

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 5.1
- **Description**: Integration tests verifying PS-ID, PRODUCT, SUITE scope generation. Verify directory counts.
- **Acceptance Criteria**:
  - [ ] PS-ID scope: 86 dirs verified (skeleton-template, 2 realms)
  - [ ] PRODUCT scope: count verified per product
  - [ ] SUITE scope: 608 dirs verified (10 PS-IDs, 2 realms)
  - [ ] Temp directory cleanup after tests

#### Task 5.3: Mutation Testing

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Tasks 5.1, 5.2
- **Description**: Run mutation testing on generator and verify ≥95% mutation kill rate.
- **Acceptance Criteria**:
  - [ ] `gremlins unleash` ≥95% on generator package
  - [ ] No survived mutations in critical path logic

#### Task 5.4: Race Detector

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Tasks 5.1, 5.2
- **Description**: Run tests with race detector.
- **Acceptance Criteria**:
  - [ ] `go test -race -count=2 ./internal/apps/framework/tls/...` clean

#### Task 5.5: Linting Verification

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 2 complete
- **Description**: Verify all code passes linting.
- **Acceptance Criteria**:
  - [ ] `golangci-lint run` clean
  - [ ] `golangci-lint run --build-tags e2e,integration` clean
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean

#### Task 5.6: End-to-End Verification

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Tasks 5.1-5.5
- **Description**: Run full pki-init command and verify output structure matches tls-structure.md examples byte-for-byte.
- **Acceptance Criteria**:
  - [ ] `pki-init skeleton-template /tmp/test` matches Example: skeleton-template
  - [ ] All 86 directories present with correct files (2 realms)
  - [ ] File permissions correct

---

### Phase 6: Knowledge Propagation [Status: ☐ TODO]

**Phase Objective**: Apply lessons learned to permanent artifacts.

#### Task 6.1: Review Lessons

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phases 1-5 complete
- **Description**: Review lessons.md entries from all phases.
- **Acceptance Criteria**:
  - [ ] All phase lessons reviewed
  - [ ] Actionable items identified

#### Task 6.2: Update Permanent Artifacts

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 6.1
- **Description**: Update ENG-HANDBOOK.md, agents, skills, instructions as warranted by lessons. Known gaps:
  - ENG-HANDBOOK.md §6.11: Add cross-reference to `tls-structure.md` and 14-category cert structure.
  - ENG-HANDBOOK.md §6.11: Add PKCS#12 as supported format.
  - ENG-HANDBOOK.md §10.3.4: Fix pre-existing `InsecureSkipVerify: true` example (use `RootCAs` with `TLSRootCAPool()` instead).
- **Acceptance Criteria**:
  - [ ] ENG-HANDBOOK.md §6.11 references tls-structure.md
  - [ ] ENG-HANDBOOK.md §6.11 mentions PKCS#12 format
  - [ ] ENG-HANDBOOK.md §10.3.4 InsecureSkipVerify example fixed
  - [ ] Propagation check passes: `go run ./cmd/cicd-lint lint-docs validate-propagation`

#### Task 6.3: Final Commit

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 6.2
- **Description**: Final commit with all artifact updates.
- **Acceptance Criteria**:
  - [ ] Clean working tree
  - [ ] All quality gates pass
  - [ ] Conventional commit message

---

## Cross-Cutting Tasks

### Testing

- [ ] Unit tests ≥95% coverage (generator)
- [ ] Integration tests pass (all 3 scopes)
- [ ] Mutation testing ≥95%
- [ ] Race detector clean
- [ ] No skipped tests

### Code Quality

- [ ] Linting passes: `golangci-lint run`
- [ ] No new TODOs without tracking
- [ ] File size ≤500 lines (refactor if exceeded)

### Documentation

- [ ] tls-structure.md complete ✅
- [ ] deployment-templates.md updated
- [ ] target-structure.md updated

---

## Deferred to Framework V12

The following work was explicitly deferred per quizme-v1 decisions:

- **PostgreSQL TLS wiring** (Decision 2=D): Config changes for mTLS on shared-postgres, HBA rules, GORM connection strings.
- **OTel Collector TLS wiring** (Decision 3=D): OTel config for mTLS receiver, app OTLP exporter TLS config.
- **OTel → Grafana TLS wiring** (Decision 4=D): Grafana LGTM OTLP ingest mTLS config.
- **Grafana stack changes** (Decision 5=D): Keep grafana/otel-lgtm as-is.

See `docs/framework-v12/plan.md` for the deferred work plan.

---

## Evidence Archive

- `test-output/phase0-research/` — Generator analysis, naming convention mapping
- `test-output/phase2/` — Generator rewrite test results
- `test-output/phase5/` — Quality gate evidence
