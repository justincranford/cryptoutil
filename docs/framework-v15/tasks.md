# Tasks - Framework V15: Pre-Flight Gap Fixes + OTel/Grafana mTLS + Public App TLS Trust

**Status**: 14 of 46 tasks complete (30%) — Phase 0+1 COMPLETE, Phases 2-12 remaining
**Last Updated**: 2026-04-21
**Created**: 2026-04-16

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers — NO exceptions:**
- ✅ **Fix issues immediately** — All failing tests/builds/lints are BLOCKING
- ✅ **Treat as BLOCKING** — ALL issues block progress to next task
- ✅ **Document root causes** — root cause analysis required; planning blockers resolved during planning
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task or step complete with known issues

---

## Task Status Legend — MANDATORY

| Symbol | Meaning | When to Use |
|--------|---------|-------------|
| ❌ | Not started | Task not yet begun |
| 🔄 | In progress | Currently being worked on |
| ✅ | Complete | Task finished with evidence |
| ⏳ | Blocked | Requires external dependency (MUST have resolution plan) |

---

## Task Checklist

### Phase 0: Pre-Flight Gap Fixes [Status: ✅ COMPLETE]

**Phase Objective**: Fix all CRITICAL and HIGH gaps from `gaps.md` before TLS work begins.

**Priority order**: CI/CD gaps first (0.1–0.3), then code correctness (0.4), then refactoring
(0.5), then medium fixes (0.6–0.8), then documentation (0.9).

**V14 anti-pattern**: Run `go run ./cmd/cicd-lint lint-go ./...` FIRST to establish baseline.

---

#### Task 0.1: Fix ci-quality.yml — Add lint-docs + lint-deployments + Permissions Block

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: None
- **Gaps Fixed**: Gap 1.1 (CRITICAL), Gap 1.7 (MEDIUM)
- **Description**: Add `lint-docs` and `lint-deployments` CI steps; add top-level permissions block.
- **Acceptance Criteria**:
  - [x] New step: `go run ./cmd/cicd-lint lint-docs` in `ci-quality.yml`
  - [x] New step: `go run ./cmd/cicd-lint lint-deployments` in `ci-quality.yml`
  - [x] Top-level `permissions: { contents: read }` block added
  - [x] Both steps run on every push/PR
  - [x] `golangci-lint run` clean on workflow YAML (if applicable)
- **Files**: `.github/workflows/ci-quality.yml`

---

#### Task 0.2: Fix ci-coverage.yml — Remove continue-on-error

- **Status**: ✅
- **Estimated**: 0.25h
- **Dependencies**: None
- **Gaps Fixed**: Gap 1.2 (HIGH)
- **Description**: Remove `continue-on-error: true` from coverage enforcement step so coverage
  threshold violations actually block CI.
- **Acceptance Criteria**:
  - [x] `continue-on-error: true` removed from coverage enforcement step
  - [x] Coverage failure causes workflow to fail with non-zero exit code
  - [x] No other `continue-on-error: true` on quality-gate steps
- **Files**: `.github/workflows/ci-coverage.yml`

---

#### Task 0.3: Fix ci-identity-validation.yml — Permissions + GO_VERSION

- **Status**: ✅
- **Estimated**: 0.75h
- **Dependencies**: None
- **Gaps Fixed**: Gap 1.3 (HIGH)
- **Description**: Scope permissions to minimum required per job; consume GO_VERSION from shared
  workflow output instead of hardcoding.
- **Acceptance Criteria**:
  - [x] Remove workflow-level `pull-requests: write` (too broad)
  - [x] Add per-job `permissions:` blocks scoped to minimum required
  - [x] `GO_VERSION` consumed from shared `workflow-job-begin` outputs (not hardcoded)
  - [x] Workflow still functions correctly after permission scoping
- **Files**: `.github/workflows/ci-identity-validation.yml`

---

#### Task 0.4: Fix sm-kms Shutdown — Add Timeout Context

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: None
- **Gaps Fixed**: Gap 2.2 (HIGH)
- **Description**: `sm-kms` calls `server.Shutdown()` with `context.Background()` (no timeout).
  All other services use a bounded timeout. Add `context.WithTimeout` with the canonical shutdown
  duration constant.
- **Acceptance Criteria**:
  - [x] `context.WithTimeout(ctx, magic.DefaultDataServerShutdownTimeout)` applied in Shutdown path
  - [x] `magic.DefaultDataServerShutdownTimeout` constant exists in `internal/shared/magic/`
  - [x] `go test ./internal/apps/sm-kms/...` passes
  - [x] `golangci-lint run ./internal/apps/sm-kms/...` passes
- **Files**: `internal/apps/sm-kms/server/server.go`, `internal/shared/magic/magic_*.go` (if constant missing)

---

#### Task 0.5: Refactor Duplicate usage.go Files

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: None
- **Gaps Fixed**: Gap 2.1 (HIGH)
- **Description**: 4 pairs of nearly identical `usage.go` files exist at product + service level.
  Extract shared usage generation to `internal/apps/framework/service/usage/`.
  **NOTE**: If this refactor proves larger than estimated (>2h), defer to V16 with a GAP file.
- **Acceptance Criteria**:
  - [x] `internal/apps/framework/service/usage/` package created with shared usage generation logic
  - [x] `internal/apps/{sm,sm-kms,sm-im}/usage.go` updated to use shared util
  - [x] `internal/apps/{jose,jose-ja}/usage.go` updated to use shared util
  - [x] `internal/apps/{pki,pki-ca}/usage.go` updated to use shared util
  - [x] `go build ./...` clean; `golangci-lint run ./...` clean; all tests pass
- **Files**: New `internal/apps/framework/service/usage/usage.go`,
  `internal/apps/{sm,sm-kms,sm-im,jose,jose-ja,pki,pki-ca}/usage.go` (7 files)

---

#### Task 0.6: Batch Small Code Fixes

- **Status**: ✅
- **Estimated**: 0.75h
- **Dependencies**: None
- **Gaps Fixed**: Gap 2.3 (MEDIUM), Gap 2.4 (MEDIUM), Gap 5.1 (CODE QUALITY)
- **Description**: Three small, independent code fixes batched together.
  - **Gap 2.3**: Add explicit `uint16()` cast for port in `identity-authz` server binding (consistent
    with all 9 other services).
  - **Gap 2.4**: Standardize signal handling cleanup across all 10 service entry points to use
    canonical `signal.Stop(sigChan) + close(sigChan)` pattern.
  - **Gap 5.1**: Replace `pki-ca` TestMain manual 300-attempt polling loop with
    `MustStartAndWaitForDualPorts` helper from shared test infrastructure.
- **Acceptance Criteria**:
  - [x] `identity-authz`: `uint16(port)` cast present
  - [x] Signal handling: all 10 entry points use `signal.Stop(sigChan); close(sigChan)` in defer
  - [x] `pki-ca` TestMain uses `MustStartAndWaitForDualPorts` (not manual loop)
  - [x] `go test ./...` clean; `golangci-lint run ./...` clean
- **Files**: `internal/apps/identity-authz/server/server.go`,
  service entry points for `{sm-kms,sm-im,jose-ja,pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa,skeleton-template}` (10 files),
  `internal/apps/pki-ca/server/testmain_test.go`

---

#### Task 0.7: Fix Pre-Commit and Medium CI/CD Gaps

- **Status**: ✅
- **Estimated**: 0.75h
- **Dependencies**: None
- **Gaps Fixed**: Gap 4.1 (MEDIUM), Gap 1.5 (MEDIUM), Gap 1.9 (MEDIUM)
- **Description**: Batch fix for remaining medium-priority pre-commit and CI/CD gaps.
  - **Gap 4.1**: Pin `golangci-lint` to `v2.7.2` in `.pre-commit-config.yaml`
  - **Gap 1.5**: Fix `ci-race.yml` — add `-tags integration` build tag to coverage
  - **Gap 1.9**: Move Maven cache from `ci-sast.yml` to `ci-load.yml` where it's actually needed
- **Acceptance Criteria**:
  - [x] `.pre-commit-config.yaml`: golangci-lint pinned to `v2.7.2`
  - [x] `ci-race.yml`: `-tags integration` build tag added to coverage step
  - [x] `ci-load.yml`: Maven cache added; Maven steps succeed
  - [x] `ci-sast.yml`: redundant Maven cache removed
- **Files**: `.pre-commit-config.yaml`, `.github/workflows/ci-race.yml`,
  `.github/workflows/ci-load.yml`, `.github/workflows/ci-sast.yml`

---

#### Task 0.8: Update tls-structure.md Documentation

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: None
- **Gaps Fixed**: Gaps 6.1–6.5 (MEDIUM documentation)
- **Description**: Update `docs/tls-structure.md` with the 5 remaining documentation gaps.
  - **6.1**: Add admin CA bundle (`issuing-ca.pem`) documentation
  - **6.2**: Document `tls-config.yml` `TLSModeMixed` pattern
  - **6.3**: Explain realm dynamic binding mechanism
  - **6.4**: Clarify `postgres` vs `postgres-1`/`postgres-2` naming (partially done — verify completeness)
  - **6.5**: Add directory count formula derivation (show `30 global + ...` breakdown)
- **Acceptance Criteria**:
  - [x] `tls-structure.md` addresses all 5 documentation gaps
  - [x] Directory count includes derivation formula (e.g., `10 global + 40 per-PS-ID × 10 = ...`)
  - [x] Cat 4 postgres naming explanation references ENG-HANDBOOK.md §6.11.3
  - [x] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**: `docs/tls-structure.md`

---

#### Task 0.9: Phase 0 Post-Mortem

- **Status**: ✅
- **Estimated**: 0.25h
- **Dependencies**: Tasks 0.1–0.8 complete
- **Description**: Review Phase 0 findings and update lessons.md.
- **Acceptance Criteria**:
  - [x] `lessons.md` Phase 0 section filled with: what worked, what didn't, root causes, patterns
  - [x] All quality gates pass: `go test ./...`, `golangci-lint run`, build clean
  - [x] Clean working tree: `git status --porcelain` returns empty

---

### Phase 1: pki-init Comprehensive Tests — All 16 Tier Domain Parameters [Status: ✅ COMPLETE]

**Phase Objective**: Write comprehensive unit + integration tests for ALL 16 valid tier domain
parameters. The generator code is already fully implemented. Phase 1 is PURELY test-writing.

**CRITICAL FINDING**: ALL 14 certificate categories (Cat 1–14) were already implemented before V15.
Phase 1 validates the existing generator via automated tests so regression is caught in CI.

---

#### Task 1.1: All-16-Tiers Unit Test

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: Phase 0 complete
- **Description**: Table-driven test calling `Generate()` with each of the 16 valid tier IDs,
  using a seam-injected (stub crypto) generator. Assert each call succeeds with no error.
- **Valid tier IDs**:
  ```
  Suite: cryptoutil
  Products: sm, jose, pki, identity, skeleton
  PS-IDs: sm-kms, sm-im, jose-ja, pki-ca,
          identity-authz, identity-idp, identity-rp, identity-rs, identity-spa,
          skeleton-template
  ```
- **Acceptance Criteria**:
  - [x] Table-driven test with all 16 entries — each tier ID gets its own subtest
  - [x] Uses real ECDSA (P-256) crypto (D7=E — no stub seam needed; P-256 is fast for 16 tiers)
  - [x] No hardcoded UUIDs; uses `t.TempDir()` for output dir per subtest
  - [x] `t.Parallel()` on parent test and ALL subtests
  - [x] All 16 subtests pass; CI confirms none skipped
  - [x] Test file name is semantic (e.g., `generator_all_tiers_test.go`, not `generator_coverage_test.go`)
- **Files**: `internal/apps/framework/tls/generator_all_tiers_test.go` (new file)

---

#### Task 1.2: 1-to-Many Category Mapping Test

- **Status**: ✅
- **Estimated**: 2h
- **Dependencies**: Task 1.1
- **Description**: For each of the 14 certificate categories, assert that the expected set of
  output directories is created by the generator.
- **Acceptance Criteria**:
  - [x] For each Cat N (1–14), define the expected output directory names (parameterized by tier)
  - [x] Assert: all expected dirs for that Cat N exist → 1-to-many confirmed
  - [x] Add `// Cat N: <name>` comment at each expected-dir definition for cross-reference
  - [x] Missing dirs = test failure (generator regression caught)
  - [x] Extra dirs beyond expected set = test failure in Task 1.3
  - [x] Tests use real ECDSA (P-256) crypto (D7=E — no stub seam needed)
- **Files**: `internal/apps/framework/tls/generator_category_mapping_test.go` (new file)

---

#### Task 1.3: 1-to-1 Uniqueness Validation Test

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: Task 1.2
- **Description**: Assert each generated directory maps to EXACTLY ONE category. No directory should
  appear in two category definitions, and no generated directory should be missing from the category
  spec (no orphans).
- **Acceptance Criteria**:
  - [x] Build complete expected-dir → category mapping (from Task 1.2 definitions)
  - [x] Walk actual generated output; for each dir found, look up its category — assert exactly 1 match
  - [x] Any generated dir NOT in the mapping → test fails (orphan dir detected)
  - [x] Any category mapping entry NOT generated → test fails (missing dir, caught by Task 1.2)
  - [x] 1-to-1 invariant confirmed bidirectionally
- **Files**: Same file as Task 1.2 or `generator_uniqueness_test.go` (new file)

---

#### Task 1.4: Integration Test with Real Crypto

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: Task 1.3
- **Description**: Run the full generator with real RSA/ECDSA key generation for at least one tier
  to catch encoding errors that stub crypto cannot detect.
- **Acceptance Criteria**:
  - [x] Uses `ExportedProductionNewGenerator` (real crypto — NOT stub)
  - [x] Test skipped under `-short` flag (`if testing.Short() { t.Skip("real crypto") }`)
  - [x] Tests `skeleton-template` tier (smallest — 1 PS-ID × 4 variants)
  - [x] Generated certs: valid X.509 (parsed with `x509.ParseCertificate` — no errors)
  - [x] Generated keys: valid PKCS#8 or EC key format (parsed — no errors)
  - [x] File is named `generator_integration_test.go`; uses `//go:build integration` tag
  - [x] Coverage ≥98%; mutation score ≥98% for generator package
- **Files**: `internal/apps/framework/tls/generator_integration_test.go` (new file)

---

### Phase 2: OTel Collector Server TLS [Status: ✅ COMPLETE]

**Phase Objective**: Configure OTel Collector OTLP receivers to serve TLS.

#### Task 2.1: OTLP gRPC Receiver TLS Config

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: Phase 1 complete
- **Description**: Update otel-collector-contrib YAML config for TLS on gRPC receiver.
- **Acceptance Criteria**:
  - [x] `receivers.otlp.protocols.grpc.tls.cert_file` = Cat 2 OTel server cert path
  - [x] `receivers.otlp.protocols.grpc.tls.key_file` = Cat 2 key path
  - [x] `receivers.otlp.protocols.grpc.tls.client_ca_file` = Cat 8 OTel client issuing CA truststore
  - [x] `insecure: false` (or omitted — false is default)
- **Files**: `deployments/shared-telemetry/otel/otel-collector-config.yaml`

#### Task 2.2: OTLP HTTP Receiver TLS Config

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: Task 2.1
- **Description**: Apply identical TLS config to HTTP receiver (:4318).
- **Acceptance Criteria**:
  - [x] `receivers.otlp.protocols.http.tls.cert_file`, `tls.key_file`, `tls.client_ca_file` set
  - [x] Same Cat 2 / Cat 8 cert paths as gRPC receiver
- **Files**: `deployments/shared-telemetry/otel/otel-collector-config.yaml`

#### Task 2.3: OTel Compose Cert Volume Mounts

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: Tasks 2.1, 2.2
- **Description**: Mount cert dirs in OTel Collector service with least privilege (Phase 2 mounts
  only — Cat 9 infra added in Phase 6).
- **Acceptance Criteria**:
  - [x] OTel mounts: Cat 2 `public-https-server-entity-otel-collector-contrib/` + Cat 8
    `otel-collector-contrib-https-client-issuing-ca/truststore/` ONLY (in each PS-ID compose)
  - [x] Canonical template `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml`
    updated atomically in the same commit (D9=A) — template matches actual
  - [x] `go run ./cmd/cicd-lint lint-deployments` exits 0 after this task
- **Files**: All 10 `deployments/{PS-ID}/compose.yml`,
  `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml`

---

### Phase 3: App→OTel Client mTLS [Status: ✅ COMPLETE]

**Phase Objective**: Configure app OTLP exporters to present Cat 9 app client certs to OTel.

#### Task 3.1: Framework OTLP Exporter TLS Config Fields

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: Phase 2 complete
- **Description**: Add optional TLS config fields to framework OTLP exporter settings.
- **Acceptance Criteria**:
  - [x] Framework config struct: `otlp-tls-cert-file` (Cat 9 app client cert path)
  - [x] Framework config struct: `otlp-tls-key-file` (Cat 9 app client key path)
  - [x] Framework config struct: `otlp-tls-ca-file` (Cat 1 server CA truststore)
  - [x] OTLP exporter uses `grpc.WithTransportCredentials(...)` when cert fields set (gRPCS)
  - [x] OTLP exporter uses `WithTLSClientConfig(...)` when cert fields set (HTTPS)
  - [x] `go build ./...` clean; `golangci-lint run` clean; coverage ≥95%
- **Files**: `internal/apps/framework/service/config/`, `internal/shared/telemetry/`

#### Task 3.2: Deployment Config Per Variant

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: Task 3.1
- **Description**: Add OTLP TLS config to all 40 per-variant deployment config files.
- **Acceptance Criteria**:
  - [x] sqlite-1: `otlp-tls-cert-file` = Cat 9 app sqlite-1 cert; `otlp-tls-key-file`; `otlp-tls-ca-file` = Cat 1 truststore
  - [x] sqlite-2: same for sqlite-2
  - [x] postgres-1: same for postgres-1 (combined with V12 PG SSL fields)
  - [x] postgres-2: same for postgres-2
  - [x] `otlp-endpoint` changed from `http://` to `https://` in `cryptoutil-otel.yml`
  - [x] `validate_schema.go` updated; `config_rules` allowed keys updated; linters pass
- **Files**: `deployments/{PS-ID}/config/*-app-framework-{variant}.yml` (40 files = 4×10)

#### Task 3.3: App Compose Cert Volume Mounts

- **Status**: ✅ (Already satisfied by existing `./certs:/certs:ro` mount pattern)
- **Estimated**: 1h
- **Dependencies**: Task 3.2
- **Description**: Mount Cat 9 app cert dir and Cat 1 truststore per variant in app compose.
- **Acceptance Criteria**:
  - [x] sqlite-1: Cat 9 app cert accessible at `/certs/{PS-ID}/otel-collector-contrib-https-client-entity-{PS-ID}-sqlite-1/`
  - [x] sqlite-2: same for sqlite-2
  - [x] postgres-1: same for postgres-1
  - [x] postgres-2: same for postgres-2
  - [x] All certs accessible via existing `./certs:/certs:ro` bind mount (no new mounts needed)
  - [x] `go run ./cmd/cicd-lint lint-deployments` exits 0
- **Files**: No changes needed — `./certs:/certs:ro` already mounts all cert dirs

---

### Phase 4: Verify OTel Standalone [Status: ✅ COMPLETE]

**Phase Objective**: Confirm server TLS and app→OTel mTLS work end-to-end before Grafana.

**E2E TLS verification pattern**: ALL committed tests use Go code to dial TLS endpoints and assert
cert identity + rejection behavior. NEVER use `openssl s_client` in committed code — it is an
interactive diagnostic tool only. See ENG-HANDBOOK.md §10.4.4.

#### Task 4.1: Write Go E2E Test — OTel Server TLS

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: Phase 3 complete
- **Description**: Write Go E2E test that starts OTel Collector via compose and performs TLS
  handshake to verify server cert and CA enforcement.
- **Acceptance Criteria**:
  - [x] `docker compose build` run BEFORE `docker compose up` (via otelComposeManager.start)
  - [x] Go test uses `crypto/tls.Dial` or `net/http` with TLS config (per §10.4.4 pattern)
  - [x] Assert gRPC `:4317` handshake succeeds; server cert is Cat 2 (`public-https-server-entity-otel-collector-contrib`)
  - [x] Assert HTTP `:4318` handshake succeeds; same Cat 2 cert
  - [x] Test file in `internal/apps/framework/tls/e2e/` with `//go:build e2e` (D8=E)
  - [x] Evidence: `golangci-lint --build-tags e2e` clean; `go build -tags e2e ./...` clean
- **Files**: `internal/apps/framework/tls/e2e/otel_tls_e2e_test.go` (new — D8=E)

#### Task 4.2: Write Go E2E Test — App→OTel mTLS

- **Status**: ✅ (combined into Task 4.1 TestMain + TestOtelServerTLS_GRPC/HTTP)
- **Estimated**: 1.5h
- **Dependencies**: Task 4.1
- **Description**: Write Go E2E test verifying app OTLP exporter connects with Cat 9 app cert.
- **Acceptance Criteria**:
  - [x] Tests use Cat 9 sqlite-1 client cert for mTLS handshake
  - [x] All OTLP uses TLS — confirmed by TLS dial tests
  - [x] `magic_otel_e2e.go`: constants for cert paths, ports, compose files

#### Task 4.3: Write Go E2E Test — Non-mTLS Rejection

- **Status**: ✅ (TestOtelMTLS_Rejection in otel_tls_e2e_test.go)
- **Estimated**: 1h
- **Dependencies**: Task 4.2
- **Description**: Write Go E2E test verifying plaintext and no-client-cert connections are rejected.
- **Acceptance Criteria**:
  - [x] Attempt `crypto/tls.Dial` to `:4317` with nil client cert → assert TLS handshake error
  - [x] Tests do NOT use `openssl s_client`; all checks programmatic in Go
  - [x] `deployments/sm-kms/compose-test-otel-expose.yml` exposes ports for test access

---

### Phase 5: Grafana LGTM HTTPS + OTLP Ingest TLS [Status: ✅ COMPLETE]

**Phase Objective**: Enable Grafana HTTPS UI (D1: grafana.ini) and OTLP ingest TLS (D6: mTLS assumed).

#### Task 5.1: Create grafana.ini with HTTPS Config

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: Phase 4 complete
- **Description**: Create custom `grafana.ini` with `[server]` HTTPS configuration.
- **Acceptance Criteria**:
  - [x] `deployments/shared-telemetry/grafana-otel-lgtm/grafana.ini` created
  - [x] `[server]` section: `protocol = https`, `cert_file`, `cert_key` = Cat 2 Grafana cert paths
  - [x] File uses `/etc/pki-init/certs/` cert path prefix (template compliance)
- **Files**: `deployments/shared-telemetry/grafana-otel-lgtm/grafana.ini`

#### Task 5.2a: Empirically Verify Grafana OTLP Ingest mTLS Support (D6 investigation)

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: Task 5.1
- **Description**: Empirically verified `grafana/otel-lgtm` supports OTLP ingest mTLS via
  `OTELCOL_EXTRA_ARGS` environment variable to inject a custom OTel collector config.
- **Acceptance Criteria**:
  - [x] D6=A confirmed: `grafana/otel-lgtm` image supports mTLS via `OTELCOL_EXTRA_ARGS` env var
  - [x] `otelcol-tls-override.yaml` created for Grafana's internal OTel collector receiver TLS
  - [x] Decision recorded: D6=A (apply via OTELCOL_EXTRA_ARGS override)
- **Files**: `deployments/shared-telemetry/grafana-otel-lgtm/otelcol-tls-override.yaml`

#### Task 5.2b: Apply D6 Config Based on 5.2a Findings

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: Task 5.2a
- **Description**: Applied D6=A: Grafana OTLP ingest configured with Cat 8 `client_ca_file` via
  `OTELCOL_EXTRA_ARGS` override. Cat 2 server cert + Cat 8 client CA enforced on gRPC :4317 and HTTP :4318.
- **Acceptance Criteria**:
  - [x] Grafana OTLP ingest (:14317/:14318) configured with Cat 8 `client_ca_file` (D6=A path taken)
  - [x] Clear documented path committed — no ambiguous states
- **Files**: `deployments/shared-telemetry/grafana-otel-lgtm/otelcol-tls-override.yaml`,
  `deployments/shared-telemetry/compose.yml` (OTELCOL_EXTRA_ARGS added)

#### Task 5.3: Grafana Compose Volume Mounts and Healthcheck

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: Task 5.2
- **Description**: Mounted cert dirs and grafana.ini; updated healthcheck to HTTPS.
- **Acceptance Criteria**:
  - [x] `grafana.ini` mounted at `/etc/grafana/grafana.ini:ro`
  - [x] Cat 2 `public-https-server-entity-grafana-otel-lgtm/` keystore mounted (per PS-ID compose)
  - [x] Cat 8 `grafana-otel-lgtm-https-client-issuing-ca/truststore/` mounted (per PS-ID compose)
  - [x] `otelcol-tls-override.yaml` mounted at `/etc/grafana-otel-lgtm/otelcol-tls-override.yaml:ro`
  - [x] Healthcheck: `https://127.0.0.1:3000/api/health` with `--insecure`; `start_period` (underscore)
  - [x] NOT mounted: Cat 9 dirs in Grafana (Grafana does not present client certs)
  - [x] All 10 PS-ID compose.yml files updated with grafana-otel-lgtm cert mounts (D9=A)
  - [x] `go run ./cmd/cicd-lint lint-deployments` exits 0 after this task
- **Files**: `deployments/shared-telemetry/compose.yml`,
  all 10 `deployments/{PS-ID}/compose.yml`

---

### Phase 6: OTel→Grafana Client mTLS [Status: ✅ COMPLETE]

**Phase Objective**: Configure OTel Collector exporter to present Cat 9 infra cert to Grafana.

#### Task 6.1: OTel Exporter Client Cert Config

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: Phase 5 complete
- **Description**: Updated OTel Collector exporter TLS config with Cat 9 infra client cert.
- **Acceptance Criteria**:
  - [x] `exporters.otlp.tls.ca_file` = Cat 1 truststore (verify Grafana server cert)
  - [x] `exporters.otlp.tls.cert_file` = Cat 9 infra `otel-collector-contrib-https-client-entity-infra/*.crt`
  - [x] `exporters.otlp.tls.key_file` = Cat 9 infra key path
  - [x] `exporters.otlp.endpoint` = `https://grafana-otel-lgtm:4317` (gRPC endpoint within compose network)
- **Files**: `deployments/shared-telemetry/otel/otel-collector-config.yaml`

#### Task 6.2: OTel Compose Add Cat 9 Infra + Cat 1 Mounts

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: Task 6.1
- **Description**: Added Cat 9 infra keystore + Cat 1 truststore to all 10 PS-ID OTel compose extensions.
- **Acceptance Criteria**:
  - [x] Cat 9 infra `otel-collector-contrib-https-client-entity-infra/` keystore mounted
  - [x] Cat 1 `public-https-server-issuing-ca/truststore/` mounted (verify Grafana server cert)
  - [x] Phase 2 mounts retained: Cat 2 keystore + Cat 8 truststore
  - [x] Total OTel mounts: **exactly 4 dirs**: Cat 1 + Cat 2 + Cat 8 + Cat 9 infra (per PS-ID compose)
  - [x] All 10 PS-ID compose.yml files updated (D9=A — same commit as Task 6.1)
  - [x] `go run ./cmd/cicd-lint lint-deployments` exits 0 after this task
- **Files**: all 10 `deployments/{PS-ID}/compose.yml`

---

### Phase 7: Verify OTel→Grafana Pipeline [Status: ✅ COMPLETE]

**Phase Objective**: Full pipeline verification: app→OTel→Grafana mTLS chain.

**E2E TLS verification pattern**: ALL committed tests use Go code. See ENG-HANDBOOK.md §10.4.4.

#### Task 7.1: Write Go E2E Test — Full mTLS Pipeline

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: Phase 6 complete
- **Description**: Phase 7 E2E tests are in `grafana_tls_e2e_test.go`; full pipeline test is
  part of Phase 11 `full_pipeline_test.go`. Phase 7 focuses on Grafana server TLS + mTLS enforcement.
- **Acceptance Criteria**:
  - [x] Tests use same compose stack started by TestMain in otel_tls_e2e_test.go
  - [x] `waitForGrafanaHealth` polls https://127.0.0.1:3000/api/health before assertions

#### Task 7.2: Write Go E2E Test — Grafana HTTPS UI

- **Status**: ✅
- **Estimated**: 0.75h
- **Dependencies**: Task 7.1
- **Description**: Tests written in `grafana_tls_e2e_test.go`.
- **Acceptance Criteria**:
  - [x] `TestGrafanaHTTPS_ServerCert`: TLS dial to :3000, asserts Cat 2 CN
  - [x] `TestGrafanaHTTPS_APIHealth`: GET /api/health returns HTTP 200
  - [x] Tests do NOT use `curl`; all verification programmatic in Go

#### Task 7.3: Write Go E2E Test — OTel→Grafana mTLS Rejection

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: Task 7.2
- **Description**: Tests written in `grafana_tls_e2e_test.go`.
- **Acceptance Criteria**:
  - [x] `TestGrafanaOTLP_GRPC_mTLS_Accepted`: dial :14317 with Cat 9 infra cert, asserts Cat 2 server CN
  - [x] `TestGrafanaOTLP_GRPC_mTLS_Rejected`: dial :14317 without client cert, asserts TLS error
  - [x] Tests do NOT use `openssl s_client`; all checks programmatic

---

### Phase 8: Public PS-ID App Server TLS [Status: ☐ TODO]

**Phase Objective**: Configure framework to load public server cert (Cat 3) and public client CA
(Cat 4) for the app's public :8080 listener.

#### Task 8.1: Framework Public Server Cert Config Fields

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Phase 7 complete
- **Description**: Add public server cert and client CA config fields to framework ServerSettings.
- **Acceptance Criteria**:
  - [ ] Framework config: `server.public-tls-cert-file` (path to Cat 3 `.crt`)
  - [ ] Framework config: `server.public-tls-key-file` (path to Cat 3 `.key`)
  - [ ] Framework config: `server.public-tls-client-ca-file` (path to Cat 4 truststore)
  - [ ] Public listener uses provided cert+key; `tls.RequireAndVerifyClientCert` when ca-file set
  - [ ] Fallback to auto-TLS (existing behavior) when fields absent — tested in unit test
  - [ ] `go build ./...` clean; `golangci-lint run` clean; coverage ≥95%
- **Files**: `internal/apps/framework/service/config/`, framework server builder

#### Task 8.2: Deployment Config Templates for Public TLS

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 8.1
- **Description**: Add `server.public-tls-*` fields to all 40 per-variant deployment config files.
- **Acceptance Criteria**:
  - [ ] sqlite-1: `server.public-tls-cert-file` = Cat 3 `public-https-server-entity-{PS-ID}-sqlite-1/SAME-AS-DIR-NAME.crt`; `server.public-tls-key-file`; `server.public-tls-client-ca-file` = Cat 4 sqlite-domain truststore
  - [ ] sqlite-2: same for sqlite-2
  - [ ] postgres-1: Cat 3 postgres-1 cert + Cat 4 postgres-domain truststore (combined with OTLP TLS fields from Phase 3 + PG SSL fields from V12)
  - [ ] postgres-2: same for postgres-2
  - [ ] All 40 files updated (4 variants × 10 PS-IDs)
- **Files**: `deployments/{PS-ID}/config/*-app-framework-{variant}.yml` (40 files)

#### Task 8.3: App Compose Cert Volume Mounts for Public TLS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 8.2
- **Description**: Mount Cat 3 + Cat 4 cert dirs per variant in app compose.
- **Acceptance Criteria**:
  - [ ] sqlite-1: Cat 3 `public-https-server-entity-{PS-ID}-sqlite-1/` + Cat 4 sqlite-domain truststore
  - [ ] sqlite-2: same for sqlite-2
  - [ ] postgres-1: Cat 3 + Cat 4 postgres-domain (combined with V12 Cat 6+7+10+14 + Phase 3 Cat 9 app)
  - [ ] postgres-2: same for postgres-2
  - [ ] Total per-variant app mounts verified with least-privilege table
  - [ ] Canonical template `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml`
    updated atomically in the same commit (D9=A)
  - [ ] `go run ./cmd/cicd-lint lint-deployments` exits 0 after this task
- **Files**: `deployments/{PS-ID}/compose.yml` (10 files),
  `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml`

---

### Phase 9: Deployment Templates [Status: ☐ TODO]

**Phase Objective**: Update canonical deployment templates to encode V15 cert mounts and config.

#### Task 9.1: Update deployment-templates.md

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 8 complete
- **Description**: Document V15 cert mount rules and combined V12+V15 least privilege table.
- **Acceptance Criteria**:
  - [ ] Combined V12+V15 least privilege table: service row × cert-dir column for all 4 variants + shared-telemetry services
  - [ ] OTel Collector cert dirs: Cat 1 truststore + Cat 2 keystore + Cat 8 truststore + Cat 9 infra keystore
  - [ ] Grafana cert dirs: Cat 2 keystore + Cat 8 truststore
  - [ ] App variants: Cat 3 + Cat 4 + Cat 6 + Cat 7 + Cat 9 app + Cat 10 + Cat 14 (all combined)
- **Files**: `docs/deployment-templates.md`

#### Task 9.2: Verify shared-telemetry OTel Compose Template (verification only)

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 9.1
- **Description**: Verify shared-telemetry OTel Collector compose template (updated atomically in
  Tasks 2.3 and 6.2 per D9=A). This task is verification-only — no new template writing.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` template-compliance check passes for OTel template
  - [ ] Template matches actual `deployments/shared-telemetry/compose.yml` with `__PS_ID__` substituted
  - [ ] OTel mounts present: Cat 1 truststore + Cat 2 keystore + Cat 8 truststore + Cat 9 infra
  - [ ] No new writes — template was updated atomically in Tasks 2.3 and 6.2
- **Files**: `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml` (verify only)

#### Task 9.3: Verify shared-telemetry Grafana Compose Template (verification only)

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 9.1
- **Description**: Verify shared-telemetry Grafana compose template (updated atomically in Tasks 5.3
  per D9=A). This task is verification-only — no new template writing.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` template-compliance check passes for Grafana template
  - [ ] Template matches actual `deployments/shared-telemetry/compose.yml` Grafana service section
  - [ ] Grafana mounts present: Cat 2 keystore + Cat 8 truststore + grafana.ini
  - [ ] Healthcheck uses `https://`; `start_period` (underscore in YAML)
  - [ ] No new writes — template was updated atomically in Task 5.3
- **Files**: `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml` (verify only)

#### Task 9.4: Verify PS-ID Compose Template (verification only)

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 9.1
- **Description**: Verify PS-ID compose template (updated atomically in Tasks 3.3 and 8.3 per D9=A).
  This task is verification-only — no new template writing.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-fitness` template-compliance check passes for PS-ID template
  - [ ] Template matches actual `deployments/{PS-ID}/compose.yml` with `__PS_ID__` substituted
  - [ ] All 4 variant cert mounts present: Cat 3 + Cat 4 + Cat 9 app (per variant)
  - [ ] `__PS_ID__` placeholders consistent throughout
  - [ ] No new writes — template was updated atomically in Tasks 3.3 and 8.3
- **Files**: `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml` (verify only)

---

### Phase 10: Deployment Linting [Status: ☐ TODO]

**Phase Objective**: All updated deployment files pass lint-deployments validators.

#### Task 10.1: Run lint-deployments

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 9 complete
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-deployments` exits 0
  - [ ] All 8 validators pass

#### Task 10.2: Lint Deployments Code

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 10.1
- **Acceptance Criteria**:
  - [ ] `golangci-lint run ./internal/apps/tools/cicd_lint/lint_deployments/...` passes
  - [ ] `go test ./internal/apps/tools/cicd_lint/lint_deployments/...` passes

---

### Phase 11: Deployment Verification — Full Telemetry Stack [Status: ☐ TODO]

**Phase Objective**: Start full deployment and verify complete TLS chain.

#### Task 11.1: Docker Compose Up (pki-init + shared-telemetry + one PS-ID)

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 10 complete
- **Description**: Start pki-init, shared-telemetry (OTel + Grafana), and one PS-ID (all 4 variants).
- **Acceptance Criteria**:
  - [ ] `docker compose build` run BEFORE `docker compose up` (V14 lesson — stale images = failures)
  - [ ] pki-init completes; all V15 cert dirs present
  - [ ] OTel Collector healthy (TLS active on :4317 and :4318)
  - [ ] Grafana healthy (HTTPS active on :3000)
  - [ ] All 4 PS-ID app variants healthy

#### Task 11.2: Write Go E2E Test — App→OTel→Grafana mTLS Pipeline

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 11.1
- **Description**: Write Go E2E test verifying end-to-end mTLS telemetry flow.
- **Acceptance Criteria**:
  - [ ] Go test sends OTLP spans with Cat 9 app client cert; asserts OTel accepts connection
  - [ ] Go test queries Grafana Tempo HTTP API (`https://localhost:3000/api/...`) to find trace
  - [ ] Go test attempts OTLP without client cert; asserts TLS handshake error (rejection verified)
  - [ ] Go test asserts Grafana HTTPS UI (`https://localhost:3000/api/health`) returns 200
  - [ ] All verification uses Go `crypto/tls` / `net/http` — NO `curl`, NO `openssl s_client`
  - [ ] Test file has `//go:build e2e` tag; saved to `internal/apps/framework/tls/e2e/full_pipeline_test.go` (D8=E)

#### Task 11.3: Write Go E2E Test — App Public HTTPS

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 11.2
- **Description**: Write Go E2E test verifying each PS-ID app variant serves public HTTPS with Cat 3
  cert and enforces Cat 4 CA trust.
- **Acceptance Criteria**:
  - [ ] Go `http.Client` with Cat 1 CA pool: GET `https://localhost:{port}/service/api/v1/health` returns 200 for all 4 variants
  - [ ] Go `crypto/tls.Dial` asserts server cert CN matches `public-https-server-entity-{PS-ID}-{variant}` (Cat 3)
  - [ ] Go `crypto/tls.Dial` with nil client cert asserts connection rejected (Cat 4 mTLS enforcement)
  - [ ] All verification programmatic in Go — NO `curl`, NO `openssl s_client`
  - [ ] Test file has `//go:build e2e` tag

---

### Phase 12: Knowledge Propagation [Status: ☐ TODO]

**Phase Objective**: Apply lessons learned to permanent artifacts. NEVER skip this phase.

#### Task 12.1: Review Lessons

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phases 1–11 complete
- **Description**: Review lessons.md entries from all prior phases; identify actionable items.
- **Acceptance Criteria**:
  - [ ] All 12 lessons.md phase sections reviewed
  - [ ] Actionable items identified for ENG-HANDBOOK.md, agents, skills, instructions

#### Task 12.2: Update ENG-HANDBOOK.md

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 12.1
- **Description**: Update ENG-HANDBOOK.md with OTel/Grafana mTLS and public app TLS patterns.
- **Acceptance Criteria**:
  - [ ] §9.4: OTel Collector mTLS receiver config (gRPC `tls:` block) documented
  - [ ] §9.4: Grafana HTTPS `grafana.ini` approach (D1) documented
  - [ ] §9.4: OTel→Grafana mTLS forwarding pattern documented
  - [ ] §12/§13: Public PS-ID app server TLS (Cat 3/4) deployment pattern documented
  - [ ] Combined V12+V15 cert mount least privilege table referenced
  - [ ] Commits per section (V14 lesson: atomic section commits for large doc edits)

#### Task 12.3: Update deployment-templates.md

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 12.2
- **Acceptance Criteria**:
  - [ ] Final combined V12+V15 cert mount table verified accurate
  - [ ] grafana.ini template content documented

#### Task 12.4: Verify Propagation and Final Commit

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 12.3
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
  - [ ] Clean working tree: `git status --porcelain` returns empty
  - [ ] All quality gates pass

---

## Cross-Cutting Tasks

### Testing

- [ ] Phase 0: All CI/CD changes validated in push to PR branch
- [ ] Phase 1: pki-init generator tests pass (≥98% coverage + ≥98% mutation)
- [ ] Phase 3: Framework OTLP exporter config unit tests pass (≥95%)
- [ ] Phase 8: Framework public server cert loading unit tests pass (≥95%)
- [ ] Phases 4, 7, 11: Docker Compose verification tasks pass

### Code Quality

- [ ] Linting passes: `golangci-lint run ./...` AND `golangci-lint run --build-tags e2e,integration ./...`
- [ ] No new TODOs without tracking
- [ ] Formatting clean: `gofumpt -w ./`

### Documentation

- [ ] `deployment-templates.md` updated with combined V12+V15 least privilege table
- [ ] `tls-structure.md` cross-referenced for cert category numbers (updated in Phase 0)
- [ ] ENG-HANDBOOK.md updated with OTel/Grafana/App TLS wiring patterns (Phase 12)

### Deployment

- [ ] `lint-deployments` passes after Phase 9 template updates (Phase 10)
- [ ] Docker Compose health checks pass (Phases 4, 7, 11)
- [ ] `./configs/` unchanged — auto-TLS mode only, no changes to configs/

---

## Notes

- **Least Privilege Enforcement**: Every compose template task MUST list exactly which Cat dirs are
  mounted and explicitly note what is NOT mounted.
- **`./configs/` isolation**: No changes to `./configs/` files; they continue using auto-TLS only.
- **V15 depends on V12 Phase 0**: Cat 9 infra cert (`otel-collector-contrib-https-client-entity-infra`)
  generated in V12 Phase 0. V15 Phase 1 is independent and can be started without waiting for V12
  Phases 1–11.
- **D6 contingency**: If grafana/otel-lgtm does not support OTLP ingest TLS, pivot to OTel sidecar
  (D6=C) in Phase 5 Task 5.2. Document the finding; create a fix task.
- **V14 carry-forward (admin port YAML config)**: `internal/apps/framework/service/cli/`
  subcommands (`livez`, `readyz`, `shutdown`) are CLI-args-only. Extending to support YAML config
  files is tracked as a dedicated phase in V16.
- **Phase 0 Task 0.5 carry-forward**: If usage.go refactor proves too large, defer to V16 with a
  GAP file documenting current state, target state, and acceptance criteria.

---

## Evidence Archive

- `test-output/v15-phase0/` — Phase 0 gap fixes verification
- `test-output/v15-phase1/` — pki-init V15 cert generation verification
- `test-output/v15-phase4/` — OTel standalone mTLS verification
- `test-output/v15-phase7/` — OTel→Grafana pipeline verification
- `test-output/v15-phase11/` — Full deployment stack verification
