# Tasks - Framework V15: Pre-Flight Gap Fixes + OTel/Grafana mTLS + Public App TLS Trust

**Status**: 0 of 46 tasks complete (0%)
**Last Updated**: 2026-04-22
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

### Phase 0: Pre-Flight Gap Fixes [Status: ☐ TODO]

**Phase Objective**: Fix all CRITICAL and HIGH gaps from `gaps.md` before TLS work begins.

**Priority order**: CI/CD gaps first (0.1–0.3), then code correctness (0.4), then refactoring
(0.5), then medium fixes (0.6–0.8), then documentation (0.9).

**V14 anti-pattern**: Run `go run ./cmd/cicd-lint lint-go ./...` FIRST to establish baseline.

---

#### Task 0.1: Fix ci-quality.yml — Add lint-docs + lint-deployments + Permissions Block

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: None
- **Gaps Fixed**: Gap 1.1 (CRITICAL), Gap 1.7 (MEDIUM)
- **Description**: Add `lint-docs` and `lint-deployments` CI steps; add top-level permissions block.
- **Acceptance Criteria**:
  - [ ] New step: `go run ./cmd/cicd-lint lint-docs` in `ci-quality.yml`
  - [ ] New step: `go run ./cmd/cicd-lint lint-deployments` in `ci-quality.yml`
  - [ ] Top-level `permissions: { contents: read }` block added
  - [ ] Both steps run on every push/PR
  - [ ] `golangci-lint run` clean on workflow YAML (if applicable)
- **Files**: `.github/workflows/ci-quality.yml`

---

#### Task 0.2: Fix ci-coverage.yml — Remove continue-on-error

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: None
- **Gaps Fixed**: Gap 1.2 (HIGH)
- **Description**: Remove `continue-on-error: true` from coverage enforcement step so coverage
  threshold violations actually block CI.
- **Acceptance Criteria**:
  - [ ] `continue-on-error: true` removed from coverage enforcement step
  - [ ] Coverage failure causes workflow to fail with non-zero exit code
  - [ ] No other `continue-on-error: true` on quality-gate steps
- **Files**: `.github/workflows/ci-coverage.yml`

---

#### Task 0.3: Fix ci-identity-validation.yml — Permissions + GO_VERSION

- **Status**: ❌
- **Estimated**: 0.75h
- **Dependencies**: None
- **Gaps Fixed**: Gap 1.3 (HIGH)
- **Description**: Scope permissions to minimum required per job; consume GO_VERSION from shared
  workflow output instead of hardcoding.
- **Acceptance Criteria**:
  - [ ] Remove workflow-level `pull-requests: write` (too broad)
  - [ ] Add per-job `permissions:` blocks scoped to minimum required
  - [ ] `GO_VERSION` consumed from shared `workflow-job-begin` outputs (not hardcoded)
  - [ ] Workflow still functions correctly after permission scoping
- **Files**: `.github/workflows/ci-identity-validation.yml`

---

#### Task 0.4: Fix sm-kms Shutdown — Add Timeout Context

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: None
- **Gaps Fixed**: Gap 2.2 (HIGH)
- **Description**: `sm-kms` calls `server.Shutdown()` with `context.Background()` (no timeout).
  All other services use a bounded timeout. Add `context.WithTimeout` with the canonical shutdown
  duration constant.
- **Acceptance Criteria**:
  - [ ] `context.WithTimeout(ctx, magic.DefaultDataServerShutdownTimeout)` applied in Shutdown path
  - [ ] `magic.DefaultDataServerShutdownTimeout` constant exists in `internal/shared/magic/`
  - [ ] `go test ./internal/apps/sm-kms/...` passes
  - [ ] `golangci-lint run ./internal/apps/sm-kms/...` passes
- **Files**: `internal/apps/sm-kms/server/server.go`, `internal/shared/magic/magic_*.go` (if constant missing)

---

#### Task 0.5: Refactor Duplicate usage.go Files

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: None
- **Gaps Fixed**: Gap 2.1 (HIGH)
- **Description**: 4 pairs of nearly identical `usage.go` files exist at product + service level.
  Extract shared usage generation to `internal/apps/framework/service/usage/`.
  **NOTE**: If this refactor proves larger than estimated (>2h), defer to V16 with a GAP file.
- **Acceptance Criteria**:
  - [ ] `internal/apps/framework/service/usage/` package created with shared usage generation logic
  - [ ] `internal/apps/{sm,sm-kms,sm-im}/usage.go` updated to use shared util
  - [ ] `internal/apps/{jose,jose-ja}/usage.go` updated to use shared util
  - [ ] `internal/apps/{pki,pki-ca}/usage.go` updated to use shared util
  - [ ] `go build ./...` clean; `golangci-lint run ./...` clean; all tests pass
- **Files**: New `internal/apps/framework/service/usage/usage.go`,
  `internal/apps/{sm,sm-kms,sm-im,jose,jose-ja,pki,pki-ca}/usage.go` (7 files)

---

#### Task 0.6: Batch Small Code Fixes

- **Status**: ❌
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
  - [ ] `identity-authz`: `uint16(port)` cast present
  - [ ] Signal handling: all 10 entry points use `signal.Stop(sigChan); close(sigChan)` in defer
  - [ ] `pki-ca` TestMain uses `MustStartAndWaitForDualPorts` (not manual loop)
  - [ ] `go test ./...` clean; `golangci-lint run ./...` clean
- **Files**: `internal/apps/identity-authz/server/server.go`,
  service entry points for `{sm-kms,sm-im,jose-ja,pki-ca,identity-authz,identity-idp,identity-rp,identity-rs,identity-spa,skeleton-template}` (10 files),
  `internal/apps/pki-ca/server/testmain_test.go`

---

#### Task 0.7: Fix Pre-Commit and Medium CI/CD Gaps

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: None
- **Gaps Fixed**: Gap 4.1 (MEDIUM), Gap 1.4 (MEDIUM), Gap 1.6 (MEDIUM), Gap 1.8 (MEDIUM), Gap 1.9 (MEDIUM)
- **Description**: Batch fix for remaining medium-priority pre-commit and CI/CD gaps.
  - **Gap 4.1**: Pin `golangci-lint` to `v2.7.2` in `.pre-commit-config.yaml`
  - **Gap 1.4**: Pin Docker images (`postgres:latest` → `postgres:17.2`, etc.) in affected workflows
  - **Gap 1.6**: `ci-mutation.yml` retention 7 → 30 days
  - **Gap 1.8**: `ci-load.yml` add `retention-days: 7` to artifact upload
  - **Gap 1.9**: Move Maven cache from `ci-sast.yml` to `ci-load.yml` where it's actually needed
- **Acceptance Criteria**:
  - [ ] `.pre-commit-config.yaml`: golangci-lint pinned to `v2.7.2`
  - [ ] Docker image tags pinned (no `postgres:latest`, `zaproxy:stable`) in all affected workflows
  - [ ] `ci-mutation.yml`: `retention-days: 30`
  - [ ] `ci-load.yml`: `retention-days: 7` on artifact upload steps; Maven cache moved here
  - [ ] `ci-sast.yml`: redundant Maven cache removed
- **Files**: `.pre-commit-config.yaml`, `.github/workflows/ci-mutation.yml`,
  `.github/workflows/ci-load.yml`, `.github/workflows/ci-sast.yml`,
  `.github/workflows/ci-e2e.yml` (Docker image pins), `.github/workflows/ci-dast.yml` (Docker pins)

---

#### Task 0.8: Update tls-structure.md Documentation

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: None
- **Gaps Fixed**: Gaps 6.1–6.6 (MEDIUM documentation)
- **Description**: Update `docs/tls-structure.md` with the 6 documentation gaps identified in the
  deep analysis.
  - **6.1**: Add admin CA bundle (`issuing-ca.pem`) documentation
  - **6.2**: Document `tls-config.yml` `TLSModeMixed` pattern
  - **6.3**: Explain realm dynamic binding mechanism
  - **6.4**: Clarify `postgres` vs `postgres-1`/`postgres-2` naming convention
  - **6.5**: Add directory count formula derivation (show `30 global + ...` breakdown)
  - **6.6**: Fix `pki-init-order.md` contradiction about V13 parallel execution
- **Acceptance Criteria**:
  - [ ] `tls-structure.md` addresses all 6 documentation gaps
  - [ ] Directory count includes derivation formula
  - [ ] `pki-init-order.md` contradiction removed (V13 Phase 0 REQUIRES V12 Phase 0 first)
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
- **Files**: `docs/tls-structure.md`, `docs/framework-v15/pki-init-order.md`

---

#### Task 0.9: Phase 0 Post-Mortem

- **Status**: ❌
- **Estimated**: 0.25h
- **Dependencies**: Tasks 0.1–0.8 complete
- **Description**: Review Phase 0 findings and update lessons.md.
- **Acceptance Criteria**:
  - [ ] `lessons.md` Phase 0 section filled with: what worked, what didn't, root causes, patterns
  - [ ] All quality gates pass: `go test ./...`, `golangci-lint run`, build clean
  - [ ] Clean working tree: `git status --porcelain` returns empty

---

### Phase 1: pki-init Patch — Cat 2, Cat 3, Cat 4, Cat 8, Cat 9 app [Status: ☐ TODO]

**Phase Objective**: Add V15 cert category generation to pki-init generator.

**V14 lessons**: Read `internal/apps/framework/tls/generator.go` fully first. Add `// Cat N: <name>`
comments at each call site. Both ExportedNewTestXxx AND ExportedProductionNewXxx test paths required.

#### Task 1.1: Add Cat 2 Server Entity Generation

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 0 complete
- **Description**: Add `public-https-server-entity-otel-collector-contrib` and
  `public-https-server-entity-grafana-otel-lgtm` cert generation to pki-init.
- **Acceptance Criteria**:
  - [ ] Cat 2 OTel entity generated with Server Auth EKU
  - [ ] Cat 2 Grafana entity generated with Server Auth EKU
  - [ ] SAME-AS-DIR-NAME file naming convention used
  - [ ] `go test ./...` passes with ≥98% coverage on generator code
- **Files**: `internal/apps/framework/tls/generator.go` + generator tests

#### Task 1.2: Add Cat 3 + Cat 4 Per-PS-ID Entity Generation

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 1.1
- **Description**: Add PS-ID public server cert (Cat 3) and PS-ID public client CA (Cat 4)
  generation for all 10 PS-IDs × 4 variants (Cat 3) and all 10 PS-IDs × 2 domains (Cat 4).
- **Acceptance Criteria**:
  - [ ] Cat 3: `public-https-server-entity-{PS-ID}-{sqlite-1,sqlite-2,postgres-1,postgres-2}` per
    PS-ID with Server Auth EKU (40 entities = 4×10)
  - [ ] Cat 4: `public-https-client-issuing-ca-{PS-ID}-{sqlite-domain,postgres-domain}/truststore/`
    CA per PS-ID (20 entities = 2×10)
  - [ ] Tests assert 4 new Cat 3 dirs + 2 new Cat 4 dirs per PS-ID
- **Files**: `internal/apps/framework/tls/generator.go` + generator tests

#### Task 1.3: Add Cat 8 + Cat 9 app Entity Generation

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 1.2
- **Description**: Add OTel/Grafana client CA (Cat 8) and app→OTel client certs (Cat 9 app)
  generation.
- **Acceptance Criteria**:
  - [ ] Cat 8: `otel-collector-contrib-https-client-issuing-ca/truststore/` CA generated
  - [ ] Cat 8: `grafana-otel-lgtm-https-client-issuing-ca/truststore/` CA generated
  - [ ] Cat 9 app: `otel-collector-contrib-https-client-entity-{PS-ID}-{variant}` per PS-ID with
    Client Auth EKU (40 entities = 4×10)
  - [ ] Tests cover all new dirs; `go test ./...` clean; coverage ≥98%; mutation ≥98%
- **Files**: `internal/apps/framework/tls/generator.go` + generator tests

#### Task 1.4: Integration Verify pki-init Output

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 1.3
- **Description**: Run pki-init and verify all expected V15 cert dirs are created.
- **Acceptance Criteria**:
  - [ ] All Cat 2, 3, 4, 8, 9 app dirs present in output volume
  - [ ] No V12 cert dirs missing (regression check — all previously generated dirs still present)
  - [ ] `golangci-lint run ./...` clean on pki-init package

---

### Phase 2: OTel Collector Server TLS [Status: ☐ TODO]

**Phase Objective**: Configure OTel Collector OTLP receivers to serve TLS.

#### Task 2.1: OTLP gRPC Receiver TLS Config

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 1 complete
- **Description**: Update otel-collector-contrib YAML config for TLS on gRPC receiver.
- **Acceptance Criteria**:
  - [ ] `receivers.otlp.protocols.grpc.tls.cert_file` = Cat 2 OTel server cert path
  - [ ] `receivers.otlp.protocols.grpc.tls.key_file` = Cat 2 key path
  - [ ] `receivers.otlp.protocols.grpc.tls.client_ca_file` = Cat 8 OTel client issuing CA truststore
  - [ ] `insecure: false` (or omitted — false is default)
- **Files**: `deployments/shared-telemetry/otel-collector-contrib/config.yml`

#### Task 2.2: OTLP HTTP Receiver TLS Config

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 2.1
- **Description**: Apply identical TLS config to HTTP receiver (:4318).
- **Acceptance Criteria**:
  - [ ] `receivers.otlp.protocols.http.tls.cert_file`, `tls.key_file`, `tls.client_ca_file` set
  - [ ] Same Cat 2 / Cat 8 cert paths as gRPC receiver
- **Files**: `deployments/shared-telemetry/otel-collector-contrib/config.yml`

#### Task 2.3: OTel Compose Cert Volume Mounts

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Tasks 2.1, 2.2
- **Description**: Mount cert dirs in OTel Collector service with least privilege (Phase 2 mounts
  only — Cat 9 infra added in Phase 6).
- **Acceptance Criteria**:
  - [ ] `__PS_ID__-certs` named volume referenced (D5 include-merged, not re-declared)
  - [ ] OTel mounts: Cat 2 `public-https-server-entity-otel-collector-contrib/` + Cat 8
    `otel-collector-contrib-https-client-issuing-ca/truststore/` ONLY
  - [ ] NOT mounted: Cat 9 infra (Phase 6), Cat 1, Cat 3, Cat 4 (OTel server does not need these)
  - [ ] Healthcheck updated: endpoint uses `https://` scheme
  - [ ] `start_period` used (underscore — not `start-period`)
- **Files**: `deployments/shared-telemetry/compose.yml`

---

### Phase 3: App→OTel Client mTLS [Status: ☐ TODO]

**Phase Objective**: Configure app OTLP exporters to present Cat 9 app client certs to OTel.

#### Task 3.1: Framework OTLP Exporter TLS Config Fields

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Phase 2 complete
- **Description**: Add optional TLS config fields to framework OTLP exporter settings.
- **Acceptance Criteria**:
  - [ ] Framework config struct: `otlp.tls.cert-file` (Cat 9 app client cert path)
  - [ ] Framework config struct: `otlp.tls.key-file` (Cat 9 app client key path)
  - [ ] Framework config struct: `otlp.tls.ca-file` (Cat 1 server CA truststore)
  - [ ] OTLP exporter uses `grpc.WithTransportCredentials(...)` when cert fields set
  - [ ] `otlp.insecure` defaults to `false` when TLS fields present
  - [ ] `go build ./...` clean; `golangci-lint run` clean; coverage ≥95%
- **Files**: `internal/apps/framework/service/config/`, `internal/shared/telemetry/`

#### Task 3.2: Deployment Config Per Variant

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 3.1
- **Description**: Add OTLP TLS config to all 40 per-variant deployment config files.
- **Acceptance Criteria**:
  - [ ] sqlite-1: `otlp.tls.cert-file` = Cat 9 app sqlite-1 cert; `otlp.tls.key-file`; `otlp.tls.ca-file` = Cat 1 truststore
  - [ ] sqlite-2: same for sqlite-2
  - [ ] postgres-1: same for postgres-1 (combined with V12 PG SSL fields)
  - [ ] postgres-2: same for postgres-2
  - [ ] `otlp.endpoint` changed from `http://` to `https://` for all 4 variants × 10 PS-IDs = 40 files
- **Files**: `deployments/{PS-ID}/config/*-app-framework-{variant}.yml` (40 files = 4×10)

#### Task 3.3: App Compose Cert Volume Mounts

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 3.2
- **Description**: Mount Cat 9 app cert dir and Cat 1 truststore per variant in app compose.
- **Acceptance Criteria**:
  - [ ] sqlite-1: Cat 9 app `otel-collector-contrib-https-client-entity-{PS-ID}-sqlite-1/` + Cat 1 truststore
  - [ ] sqlite-2: same for sqlite-2
  - [ ] postgres-1: Cat 9 app postgres-1 + Cat 1 (combined with V12 Cat 14 + Cat 10 mounts)
  - [ ] postgres-2: same for postgres-2
  - [ ] NO extra cert dirs beyond minimum required
- **Files**: `deployments/{PS-ID}/compose.yml` (10 files = 1 per PS-ID)

---

### Phase 4: Verify OTel Standalone [Status: ☐ TODO]

**Phase Objective**: Confirm server TLS and app→OTel mTLS work end-to-end before Grafana.

#### Task 4.1: Verify OTel Server TLS

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 3 complete
- **Description**: Verify OTel server TLS alone (before adding client cert requirements).
- **Acceptance Criteria**:
  - [ ] `docker compose build` run BEFORE `docker compose up` (V14 lesson)
  - [ ] `openssl s_client -connect localhost:4317 -servername otel-collector-contrib` shows Cat 2 cert
  - [ ] TLS 1.3 negotiated
  - [ ] HTTP :4318 also shows TLS

#### Task 4.2: Verify App→OTel mTLS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 4.1
- **Description**: Verify telemetry flows from app to OTel Collector via mTLS.
- **Acceptance Criteria**:
  - [ ] App instances start; OTLP exporter connects to OTel successfully (Cat 9 app cert)
  - [ ] Traces visible in OTel pipeline (logs from collector show data received)
  - [ ] OTel healthcheck passes
  - [ ] Evidence saved: `test-output/v15-phase4/app-to-otel-mtls.log`

#### Task 4.3: Verify Non-mTLS Export Rejected

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 4.2
- **Description**: Verify plaintext and missing-client-cert connections are rejected.
- **Acceptance Criteria**:
  - [ ] OTLP export with `insecure: true` rejected (TLS error)
  - [ ] OTLP export with TLS but no client cert rejected (handshake failure)
  - [ ] Evidence saved: `test-output/v15-phase4/rejection-tests.log`

---

### Phase 5: Grafana LGTM HTTPS + OTLP Ingest TLS [Status: ☐ TODO]

**Phase Objective**: Enable Grafana HTTPS UI (D1: grafana.ini) and OTLP ingest TLS (D6: mTLS assumed).

#### Task 5.1: Create grafana.ini with HTTPS Config

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 4 complete
- **Description**: Create custom `grafana.ini` with `[server]` HTTPS configuration.
- **Acceptance Criteria**:
  - [ ] `deployments/shared-telemetry/grafana-otel-lgtm/grafana.ini` created
  - [ ] `[server]` section: `protocol = https`, `cert_file`, `cert_key` = Cat 2 Grafana cert paths
  - [ ] File uses `__PS_ID__` placeholder (template compliance)
- **Files**: `deployments/shared-telemetry/grafana-otel-lgtm/grafana.ini`

#### Task 5.2: Grafana OTLP Ingest TLS Config

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 5.1
- **Description**: Configure Grafana OTLP ingest TLS (D6 verification — confirm grafana/otel-lgtm
  supports mTLS OTLP ingest).
- **Acceptance Criteria**:
  - [ ] Grafana OTLP ingest (:14317/:14318) configured for TLS with Cat 8 truststore
  - [ ] If D6=A supported: config applied
  - [ ] If NOT supported: finding documented; proceed HTTPS-only; create fix task for D6=C sidecar
- **Files**: Grafana config at `deployments/shared-telemetry/grafana-otel-lgtm/`

#### Task 5.3: Grafana Compose Volume Mounts and Healthcheck

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 5.2
- **Description**: Mount cert dirs and grafana.ini; update healthcheck to HTTPS.
- **Acceptance Criteria**:
  - [ ] `grafana.ini` mounted at `/etc/grafana/grafana.ini:ro`
  - [ ] Cat 2 `public-https-server-entity-grafana-otel-lgtm/` keystore mounted
  - [ ] Cat 8 `grafana-otel-lgtm-https-client-issuing-ca/truststore/` mounted
  - [ ] `__PS_ID__-certs` named volume referenced (D5 include-merged)
  - [ ] Healthcheck: `https://127.0.0.1:3000/api/health`; `start_period` (underscore in YAML)
  - [ ] NOT mounted: Cat 9 dirs (Grafana does not present client certs)
- **Files**: `deployments/shared-telemetry/compose.yml`

---

### Phase 6: OTel→Grafana Client mTLS [Status: ☐ TODO]

**Phase Objective**: Configure OTel Collector exporter to present Cat 9 infra cert to Grafana.

#### Task 6.1: OTel Exporter Client Cert Config

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 5 complete
- **Description**: Update OTel Collector exporter TLS config with Cat 9 infra client cert.
- **Acceptance Criteria**:
  - [ ] `exporters.otlp.tls.ca_file` = Cat 1 truststore (verify Grafana server cert)
  - [ ] `exporters.otlp.tls.cert_file` = Cat 9 infra `otel-collector-contrib-https-client-entity-infra/SAME-AS-DIR-NAME.crt`
  - [ ] `exporters.otlp.tls.key_file` = Cat 9 infra key path
  - [ ] `exporters.otlp.endpoint` = `https://grafana-otel-lgtm:14317`
- **Files**: `deployments/shared-telemetry/otel-collector-contrib/config.yml`

#### Task 6.2: OTel Compose Add Cat 9 Infra + Cat 1 Mounts

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 6.1
- **Description**: Add Cat 9 infra keystore + Cat 1 truststore to OTel compose.
- **Acceptance Criteria**:
  - [ ] Cat 9 infra `otel-collector-contrib-https-client-entity-infra/` keystore mounted
  - [ ] Cat 1 `public-https-server-issuing-ca/truststore/` mounted (verify Grafana server cert)
  - [ ] Phase 2 mounts retained: Cat 2 keystore + Cat 8 truststore
  - [ ] Total OTel mounts: **exactly 4 dirs**: Cat 1 + Cat 2 + Cat 8 + Cat 9 infra
- **Files**: `deployments/shared-telemetry/compose.yml`

---

### Phase 7: Verify OTel→Grafana Pipeline [Status: ☐ TODO]

**Phase Objective**: Full pipeline verification: app→OTel→Grafana mTLS chain.

#### Task 7.1: Verify Full Pipeline (app → OTel → Grafana)

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 6 complete
- **Description**: Verify traces, metrics, logs flow through full mTLS chain.
- **Acceptance Criteria**:
  - [ ] `docker compose build` run BEFORE `docker compose up` (V14 lesson)
  - [ ] All 4 PS-ID variants send OTLP via mTLS to OTel (Cat 9 app certs)
  - [ ] OTel forwards to Grafana via mTLS (Cat 9 infra cert)
  - [ ] Traces visible in Grafana Tempo
  - [ ] Metrics visible in Grafana Mimir/Prometheus
  - [ ] Logs visible in Grafana Loki

#### Task 7.2: Verify Grafana HTTPS UI

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 7.1
- **Description**: Verify Grafana UI accessible via HTTPS with correct cert.
- **Acceptance Criteria**:
  - [ ] `curl --cacert <Cat 1 truststore> https://localhost:3000/api/health` returns 200
  - [ ] Certificate matches Cat 2 `public-https-server-entity-grafana-otel-lgtm`

#### Task 7.3: Verify OTel→Grafana mTLS Rejection

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 7.2
- **Description**: Verify OTel→Grafana connection fails without client cert.
- **Acceptance Criteria**:
  - [ ] OTel exporter without `cert_file`/`key_file`: TLS handshake failure logged
  - [ ] OTel exporter with wrong client cert (not from Cat 8 CA): handshake failure

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
- **Files**: `deployments/{PS-ID}/compose.yml` (10 files)

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

#### Task 9.2: Update shared-telemetry OTel Compose Template

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 9.1
- **Description**: Update shared-telemetry OTel Collector compose template with V15 mounts.
- **Acceptance Criteria**:
  - [ ] OTel mounts: Cat 1 truststore + Cat 2 `public-https-server-entity-otel-collector-contrib/` + Cat 8 `otel-collector-contrib-https-client-issuing-ca/truststore/` + Cat 9 infra `otel-collector-contrib-https-client-entity-infra/`
  - [ ] `__PS_ID__` placeholders in all paths
  - [ ] Template compliance linter accepts the file
- **Files**: `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml`

#### Task 9.3: Update shared-telemetry Grafana Compose Template

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 9.1
- **Description**: Update shared-telemetry Grafana compose template with cert mounts + grafana.ini.
- **Acceptance Criteria**:
  - [ ] Grafana mounts: Cat 2 `public-https-server-entity-grafana-otel-lgtm/` + Cat 8 `grafana-otel-lgtm-https-client-issuing-ca/truststore/`
  - [ ] `grafana.ini` mounted at `/etc/grafana/grafana.ini:ro`
  - [ ] Healthcheck uses `https://`; `start_period` (underscore in YAML)
  - [ ] `__PS_ID__` placeholders used
- **Files**: `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml`

#### Task 9.4: Update PS-ID Compose Template (V15 additions)

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 9.1
- **Description**: Add Cat 3, Cat 4, Cat 9 app cert mounts to PS-ID compose template per variant.
- **Acceptance Criteria**:
  - [ ] sqlite-1: Cat 3 `public-https-server-entity-__PS_ID__-sqlite-1/` + Cat 4 sqlite-domain truststore + Cat 9 app sqlite-1 keystore
  - [ ] sqlite-2: same for sqlite-2
  - [ ] postgres-1: Cat 3 + Cat 4 postgres-domain + Cat 9 app postgres-1 (in addition to V12 mounts)
  - [ ] postgres-2: same for postgres-2
  - [ ] `__PS_ID__` placeholders consistent throughout
- **Files**: `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml`

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

#### Task 11.2: Verify App→OTel→Grafana mTLS Pipeline

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 11.1
- **Description**: Verify telemetry flows end-to-end through full mTLS chain.
- **Acceptance Criteria**:
  - [ ] All 4 PS-ID app variants send OTLP via mTLS to OTel (Cat 9 app certs)
  - [ ] OTel forwards to Grafana via mTLS (Cat 9 infra cert)
  - [ ] Traces visible in Grafana Tempo
  - [ ] Non-mTLS OTLP rejected by OTel
  - [ ] Grafana HTTPS UI accessible at `https://localhost:3000`

#### Task 11.3: Verify App Public HTTPS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 11.2
- **Description**: Verify each PS-ID app variant serves its public API over HTTPS with Cat 3 cert.
- **Acceptance Criteria**:
  - [ ] `GET /service/api/v1/health` via HTTPS returns 200 for all 4 variants
  - [ ] Server cert matches Cat 3 `public-https-server-entity-{PS-ID}-{variant}`
  - [ ] Connections without valid client cert rejected (Cat 4 CA enforcement)

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
# Tasks - Framework V13: OTel/Grafana mTLS + Public PS-ID App TLS Trust

**Status**: 0 of 37 tasks complete (0%)
**Last Updated**: 2026-04-16
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

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** - All failing tests/builds/lints are BLOCKING
- ✅ **Treat as BLOCKING**: ALL issues block progress to next task
- ✅ **Document root causes** — root cause analysis required; planning blockers resolved during planning
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task or step complete with known issues
- ✅ **NEVER de-prioritize quality**

---

## Task Checklist

### Phase 0: pki-init Patch — Cat 2, Cat 3, Cat 4, Cat 8, Cat 9 app [Status: ☐ TODO]

**Phase Objective**: Add V13 cert category generation to pki-init generator.

#### Task 0.1: Add Cat 2 Server Entity Generation

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: V12 Phase 0 complete
- **Description**: Add `public-https-server-entity-otel-collector-contrib` and `public-https-server-entity-grafana-otel-lgtm` cert generation to pki-init.
- **Acceptance Criteria**:
  - [ ] Cat 2 OTel entity generated with Server Auth EKU
  - [ ] Cat 2 Grafana entity generated with Server Auth EKU
  - [ ] SAME-AS-DIR-NAME file naming convention used
  - [ ] `go test ./...` passes with ≥98% coverage on generator code

#### Task 0.2: Add Cat 3 + Cat 4 Per-PS-ID Entity Generation

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 0.1
- **Description**: Add PS-ID public server cert (Cat 3) and PS-ID public client CA (Cat 4) generation.
- **Acceptance Criteria**:
  - [ ] Cat 3: `public-https-server-entity-{PS-ID}-{sqlite-1,sqlite-2,postgres-1,postgres-2}` generated per PS-ID with Server Auth EKU
  - [ ] Cat 4: `public-https-client-issuing-ca-{PS-ID}-{sqlite,postgres}/truststore/` CA generated per PS-ID
  - [ ] pki-domain parameter (sqlite, postgres) correctly maps to variant sets
  - [ ] Tests updated: 4 new Cat 3 dirs + 2 new Cat 4 dirs per PS-ID asserted

#### Task 0.3: Add Cat 8 + Cat 9 app Entity Generation

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 0.2
- **Description**: Add OTel/Grafana client CA (Cat 8) and app→OTel client certs (Cat 9 app) generation.
- **Acceptance Criteria**:
  - [ ] Cat 8: `otel-collector-contrib-https-client-issuing-ca/truststore/` CA generated
  - [ ] Cat 8: `grafana-otel-lgtm-https-client-issuing-ca/truststore/` CA generated
  - [ ] Cat 9 app: `otel-collector-contrib-https-client-entity-{PS-ID}-{sqlite-1,sqlite-2,postgres-1,postgres-2}` per PS-ID with Client Auth EKU
  - [ ] Tests cover all new dirs; `go test ./...` clean; coverage ≥98%

#### Task 0.4: Integration Verify pki-init Output

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 0.3
- **Description**: Run pki-init and verify all expected V13 cert dirs are created.
- **Acceptance Criteria**:
  - [ ] All Cat 2, 3, 4, 8, 9 app dirs present in output volume
  - [ ] No V12 cert dirs missing (regression check)
  - [ ] `golangci-lint run` clean on pki-init package

---

### Phase 1: OTel Collector Server TLS [Status: ☐ TODO]

**Phase Objective**: Configure OTel Collector OTLP receivers to serve TLS and require client certs.

#### Task 1.1: OTLP gRPC Receiver TLS Config

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 0 complete
- **Description**: Update otel-collector-contrib YAML config for TLS on gRPC receiver.
- **Acceptance Criteria**:
  - [ ] `receivers.otlp.protocols.grpc.tls.cert_file` = Cat 2 `public-https-server-entity-otel-collector-contrib/SAME-AS-DIR-NAME.crt`
  - [ ] `receivers.otlp.protocols.grpc.tls.key_file` = Cat 2 key path
  - [ ] `receivers.otlp.protocols.grpc.tls.client_ca_file` = Cat 8 `otel-collector-contrib-https-client-issuing-ca/truststore/` path
  - [ ] `insecure: false` (or field removed — default is false)
- **Files**: `deployments/shared-telemetry/otel-collector-contrib/config.yml`

#### Task 1.2: OTLP HTTP Receiver TLS Config

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 1.1
- **Description**: Apply identical TLS config to HTTP receiver.
- **Acceptance Criteria**:
  - [ ] `receivers.otlp.protocols.http.tls.cert_file`, `tls.key_file`, `tls.client_ca_file` configured
  - [ ] Same Cat 2 / Cat 8 cert paths as gRPC receiver
- **Files**: `deployments/shared-telemetry/otel-collector-contrib/config.yml`

#### Task 1.3: OTel Compose Cert Volume Mounts

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Tasks 1.1, 1.2
- **Description**: Mount cert dirs in otel-collector-contrib service with least privilege.
- **Acceptance Criteria**:
  - [ ] `__PS_ID__-certs` named volume referenced (D5 include-merged, not re-declared)
  - [ ] OTel Collector mounts: Cat 2 `public-https-server-entity-otel-collector-contrib/` + Cat 8 `otel-collector-contrib-https-client-issuing-ca/truststore/` ONLY
  - [ ] NO Cat 9 infra mount yet (added in Phase 5)
  - [ ] NO Cat 1, Cat 3, Cat 4 dirs mounted (OTel server does not need these)
  - [ ] Healthcheck updated: endpoint uses `https://` scheme
- **Files**: `deployments/shared-telemetry/compose.yml`

---

### Phase 2: App→OTel Client mTLS [Status: ☐ TODO]

**Phase Objective**: Configure app OTLP exporters to present client certs to OTel Collector.

#### Task 2.1: Framework OTLP Exporter TLS Config Fields

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Phase 1 complete
- **Description**: Add TLS config fields to framework OTLP exporter settings.
- **Acceptance Criteria**:
  - [ ] Framework config struct: `otlp.tls.cert-file` (Cat 9 app client cert path)
  - [ ] Framework config struct: `otlp.tls.key-file` (Cat 9 app client key path)
  - [ ] Framework config struct: `otlp.tls.ca-file` (Cat 1 server CA truststore path)
  - [ ] OTLP exporter uses `grpc.WithTransportCredentials(credentials.NewTLS(...))` when cert fields set
  - [ ] `otlp.insecure` field defaults to `false` when TLS fields present
  - [ ] `go build ./...` clean; `golangci-lint run` clean
- **Files**: `internal/apps/framework/service/config/`, `internal/shared/telemetry/`

#### Task 2.2: Deployment Config Per Variant

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 2.1
- **Description**: Add OTLP TLS config to per-variant deployment config files.
- **Acceptance Criteria**:
  - [ ] sqlite-1: `otlp.tls.cert-file` = Cat 9 app `otel-collector-contrib-https-client-entity-{PS-ID}-sqlite-1/SAME-AS-DIR-NAME.crt`
  - [ ] sqlite-1: `otlp.tls.key-file`, `otlp.tls.ca-file` = Cat 1 truststore
  - [ ] sqlite-2: same for sqlite-2
  - [ ] postgres-1: same for postgres-1 (combined with PG SSL fields from V12 Phase 4)
  - [ ] postgres-2: same for postgres-2
  - [ ] `otlp.endpoint` changed from `http://` to `https://` for all variants
- **Files**: `deployments/{PS-ID}/config/*-app-framework-{variant}.yml`

#### Task 2.3: App Compose Cert Volume Mounts

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 2.2
- **Description**: Mount Cat 9 app cert dir and Cat 1 truststore in app compose per variant.
- **Acceptance Criteria**:
  - [ ] sqlite-1: mounts Cat 9 app `otel-collector-contrib-https-client-entity-{PS-ID}-sqlite-1/` + Cat 1 `public-https-server-issuing-ca/truststore/`
  - [ ] sqlite-2: same for sqlite-2
  - [ ] postgres-1: mounts Cat 9 app for postgres-1 + Cat 1 (in addition to V12 Cat 14 + Cat 10 mounts)
  - [ ] postgres-2: same for postgres-2
  - [ ] NO extra cert dirs mounted beyond minimum required
- **Files**: `deployments/{PS-ID}/compose.yml`

---

### Phase 3: Verify OTel Standalone [Status: ☐ TODO]

**Phase Objective**: Confirm app→OTel mTLS works end-to-end before wiring Grafana.

#### Task 3.1: Verify OTel Server TLS

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 2 complete
- **Description**: Verify OTel server TLS alone.
- **Acceptance Criteria**:
  - [ ] `openssl s_client -connect localhost:4317 -servername otel-collector-contrib` shows Cat 2 cert
  - [ ] TLS 1.3 negotiated

#### Task 3.2: Verify App→OTel mTLS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 3.1
- **Description**: Verify telemetry flows from app to OTel Collector via mTLS.
- **Acceptance Criteria**:
  - [ ] App instances start; OTLP exporter connects to OTel successfully
  - [ ] Traces visible in OTel pipeline (logs from collector show data received)
  - [ ] OTel healthcheck passes

#### Task 3.3: Verify Non-mTLS Export Rejected

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 3.2
- **Description**: Verify plaintext OTLP and missing-client-cert connections are rejected.
- **Acceptance Criteria**:
  - [ ] OTLP export with `insecure: true` rejected (connection refused or TLS error)
  - [ ] OTLP export with TLS but no client cert rejected (TLS handshake failure)

---

### Phase 4: Grafana LGTM HTTPS + OTLP Ingest TLS [Status: ☐ TODO]

**Phase Objective**: Enable Grafana HTTPS UI (D1: grafana.ini) and OTLP ingest TLS (D6: mTLS assumed).

#### Task 4.1: Create grafana.ini with HTTPS Config

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 3 complete
- **Description**: Create custom `grafana.ini` with `[server]` HTTPS configuration.
- **Acceptance Criteria**:
  - [ ] `deployments/shared-telemetry/grafana-otel-lgtm/grafana.ini` created
  - [ ] `[server]` section: `protocol = https`, `cert_file = /certs/__PS_ID__/public-https-server-entity-grafana-otel-lgtm/public-https-server-entity-grafana-otel-lgtm.crt`
  - [ ] `cert_key = /certs/__PS_ID__/public-https-server-entity-grafana-otel-lgtm/public-https-server-entity-grafana-otel-lgtm.key`
  - [ ] File uses `__PS_ID__` placeholder (template compliance)
- **Files**: `deployments/shared-telemetry/grafana-otel-lgtm/grafana.ini`

#### Task 4.2: Grafana OTLP Ingest TLS Config

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 4.1
- **Description**: Configure Grafana OTLP ingest TLS (D6 verification step: confirm grafana/otel-lgtm supports this).
- **Acceptance Criteria**:
  - [ ] Grafana OTLP ingest ports (:14317/:14318) configured for TLS
  - [ ] `client_ca_file` = Cat 8 `grafana-otel-lgtm-https-client-issuing-ca/truststore/` path
  - [ ] If D6=A (mTLS supported): config applied. If not supported: document finding; proceed with Phase 4 HTTPS-only; create fix task for OTel sidecar approach
- **Files**: `deployments/shared-telemetry/grafana-otel-lgtm/` config

#### Task 4.3: Grafana Compose Volume Mounts and healthcheck

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 4.2
- **Description**: Mount cert dirs and grafana.ini; update healthcheck.
- **Acceptance Criteria**:
  - [ ] `grafana.ini` mounted at `/etc/grafana/grafana.ini:ro`
  - [ ] Cat 2 `public-https-server-entity-grafana-otel-lgtm/` keystore mounted
  - [ ] Cat 8 `grafana-otel-lgtm-https-client-issuing-ca/truststore/` mounted
  - [ ] `__PS_ID__-certs` named volume referenced (D5 include-merged)
  - [ ] Healthcheck updated: `https://127.0.0.1:3000/api/health`
  - [ ] NO Cat 9 dirs mounted (Grafana does not need them)
- **Files**: `deployments/shared-telemetry/compose.yml`

---

### Phase 5: OTel→Grafana Client mTLS [Status: ☐ TODO]

**Phase Objective**: Configure OTel Collector exporter to present Cat 9 infra client cert to Grafana.

#### Task 5.1: OTel Exporter Client Cert Config

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 4 complete
- **Description**: Update OTel Collector exporter TLS config with Cat 9 infra client cert.
- **Acceptance Criteria**:
  - [ ] `exporters.otlp.tls.ca_file` = Cat 1 `public-https-server-issuing-ca/truststore/` (verify Grafana server cert)
  - [ ] `exporters.otlp.tls.cert_file` = Cat 9 infra `otel-collector-contrib-https-client-entity-infra/SAME-AS-DIR-NAME.crt`
  - [ ] `exporters.otlp.tls.key_file` = Cat 9 infra key path
  - [ ] `exporters.otlp.endpoint` uses `https://grafana-otel-lgtm:14317` (or 14318 for HTTP)
- **Files**: `deployments/shared-telemetry/otel-collector-contrib/config.yml`

#### Task 5.2: OTel Compose Add Cat 9 Infra Mount

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 5.1
- **Description**: Add Cat 9 infra cert dir mount to OTel Collector compose (and Cat 1 truststore).
- **Acceptance Criteria**:
  - [ ] Cat 9 infra `otel-collector-contrib-https-client-entity-infra/` keystore mounted
  - [ ] Cat 1 `public-https-server-issuing-ca/truststore/` mounted (OTel needs to verify Grafana server cert)
  - [ ] Previous mounts (Cat 2 + Cat 8 from Phase 1) retained
  - [ ] Total OTel mounts: Cat 1 truststore + Cat 2 keystore + Cat 8 truststore + Cat 9 infra keystore (exactly 4 dirs)
- **Files**: `deployments/shared-telemetry/compose.yml`

---

### Phase 6: Verify OTel→Grafana Pipeline [Status: ☐ TODO]

**Phase Objective**: End-to-end telemetry pipeline verification with full mTLS chain.

#### Task 6.1: Verify Full Pipeline (app → OTel → Grafana)

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 5 complete
- **Description**: Verify traces, metrics, and logs flow from app to Grafana dashboards.
- **Acceptance Criteria**:
  - [ ] App instances send OTLP data to OTel; OTel forwards to Grafana
  - [ ] Traces visible in Grafana Tempo
  - [ ] Metrics visible in Grafana Mimir/Prometheus
  - [ ] Logs visible in Grafana Loki

#### Task 6.2: Verify Grafana HTTPS UI

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 6.1
- **Description**: Verify Grafana UI accessible via HTTPS.
- **Acceptance Criteria**:
  - [ ] `curl --cacert <Cat 1> https://localhost:3000/api/health` returns 200
  - [ ] Dashboards load in browser via HTTPS
  - [ ] Certificate matches Cat 2 `public-https-server-entity-grafana-otel-lgtm`

#### Task 6.3: Verify OTel→Grafana mTLS Rejection

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 6.2
- **Description**: Verify OTel→Grafana connection fails without client cert.
- **Acceptance Criteria**:
  - [ ] OTel exporter without `cert_file`/`key_file`: TLS handshake failure
  - [ ] OTel exporter with wrong client cert (not from Cat 8 CA): handshake failure

---

### Phase 7: Public PS-ID App Server TLS [Status: ☐ TODO]

**Phase Objective**: Configure framework to load public server cert (Cat 3) and public client CA (Cat 4).

#### Task 7.1: Framework Public Server Cert Config Fields

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Phase 6 complete
- **Description**: Add public server cert and client CA config fields to framework ServerSettings.
- **Acceptance Criteria**:
  - [ ] Framework config struct: `server.public-tls-cert-file` (path to Cat 3 `.crt`)
  - [ ] Framework config struct: `server.public-tls-key-file` (path to Cat 3 `.key`)
  - [ ] Framework config struct: `server.public-tls-client-ca-file` (path to Cat 4 truststore)
  - [ ] Public listener uses provided cert+key; `tls.RequireAndVerifyClientCert` when `client-ca-file` set
  - [ ] Falls back to auto-TLS (existing behavior) when fields absent
  - [ ] `go build ./...` clean; `golangci-lint run` clean
- **Files**: `internal/apps/framework/service/config/`, framework server builder

#### Task 7.2: Deployment Config Templates for Public TLS

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 7.1
- **Description**: Add `server.public-tls-*` fields to per-variant deployment config files.
- **Acceptance Criteria**:
  - [ ] sqlite-1: `server.public-tls-cert-file` = Cat 3 `public-https-server-entity-{PS-ID}-sqlite-1/SAME-AS-DIR-NAME.crt`
  - [ ] sqlite-1: `server.public-tls-key-file`, `server.public-tls-client-ca-file` = Cat 4 sqlite-domain truststore
  - [ ] sqlite-2: same for sqlite-2
  - [ ] postgres-1: same for postgres-1 (combined with OTLP TLS fields from Phase 2 + PG SSL fields from V12)
  - [ ] postgres-2: same for postgres-2
- **Files**: `deployments/{PS-ID}/config/*-app-framework-{variant}.yml`

#### Task 7.3: App Compose Cert Volume Mounts for Public TLS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 7.2
- **Description**: Mount Cat 3 and Cat 4 cert dirs in app compose per variant (least privilege).
- **Acceptance Criteria**:
  - [ ] sqlite-1: Cat 3 `public-https-server-entity-{PS-ID}-sqlite-1/` + Cat 4 sqlite-domain truststore
  - [ ] sqlite-2: same for sqlite-2
  - [ ] postgres-1: Cat 3 + Cat 4 postgres-domain (combined with V12 Cat 6+7+10+14 mounts)
  - [ ] postgres-2: same for postgres-2
- **Files**: `deployments/{PS-ID}/compose.yml`

---

### Phase 8: Deployment Templates [Status: ☐ TODO]

**Phase Objective**: Update canonical deployment templates to encode V13 cert mounts and config fields.

#### Task 8.1: Update deployment-templates.md

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 7 complete
- **Description**: Document V13 cert mount rules and combined V12+V13 least privilege table.
- **Acceptance Criteria**:
  - [ ] Combined V12+V13 least privilege table: service row × cert-dir column for all 4 variants + shared-telemetry services
  - [ ] OTel Collector cert dirs: Cat 1 truststore + Cat 2 keystore + Cat 8 truststore + Cat 9 infra keystore
  - [ ] Grafana cert dirs: Cat 2 keystore + Cat 8 truststore
  - [ ] App variants: Cat 3 + Cat 4 + Cat 6 + Cat 7 + Cat 9 app + Cat 10 + Cat 14 (all mounts combined)
- **Files**: `docs/deployment-templates.md`

#### Task 8.2: Update shared-telemetry OTel Compose Template

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 8.1
- **Description**: Update shared-telemetry OTel Collector compose template.
- **Acceptance Criteria**:
  - [ ] OTel mounts: Cat 1 truststore + Cat 2 `public-https-server-entity-otel-collector-contrib/` + Cat 8 `otel-collector-contrib-https-client-issuing-ca/truststore/` + Cat 9 infra `otel-collector-contrib-https-client-entity-infra/`
  - [ ] `__PS_ID__` placeholders in all paths
  - [ ] Template compliance linter accepts the file
- **Files**: `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml`

#### Task 8.3: Update shared-telemetry Grafana Compose Template

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 8.1
- **Description**: Update shared-telemetry Grafana compose template with cert mounts and grafana.ini.
- **Acceptance Criteria**:
  - [ ] Grafana mounts: Cat 2 `public-https-server-entity-grafana-otel-lgtm/` + Cat 8 `grafana-otel-lgtm-https-client-issuing-ca/truststore/`
  - [ ] `grafana.ini` mounted at `/etc/grafana/grafana.ini:ro`
  - [ ] Healthcheck uses `https://`
  - [ ] `__PS_ID__` placeholders used
- **Files**: `api/cryptosuite-registry/templates/deployments/shared-telemetry/compose.yml`

#### Task 8.4: Update PS-ID Compose Template (V13 additions)

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 8.1
- **Description**: Add Cat 3, Cat 4, Cat 9 app cert mounts to PS-ID compose template per variant.
- **Acceptance Criteria**:
  - [ ] sqlite-1: Cat 3 `public-https-server-entity-__PS_ID__-sqlite-1/` + Cat 4 sqlite-domain truststore + Cat 9 app sqlite-1 keystore
  - [ ] sqlite-2: same for sqlite-2
  - [ ] postgres-1: Cat 3 + Cat 4 postgres-domain + Cat 9 app postgres-1 (in addition to V12 mounts)
  - [ ] postgres-2: same for postgres-2
  - [ ] `__PS_ID__` placeholders consistent
- **Files**: `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml`

---

### Phase 9: Deployment Linting [Status: ☐ TODO]

**Phase Objective**: All updated deployment files pass lint-deployments validators.

#### Task 9.1: Run lint-deployments

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 8 complete
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-deployments` exits 0
  - [ ] All 8 validators pass

#### Task 9.2: Lint Deployments Code

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 9.1
- **Acceptance Criteria**:
  - [ ] `golangci-lint run ./internal/apps/tools/cicd_lint/lint_deployments/...` passes
  - [ ] `go test ./internal/apps/tools/cicd_lint/lint_deployments/...` passes

---

### Phase 10: Deployment Verification — Full Telemetry Stack [Status: ☐ TODO]

**Phase Objective**: Start full deployment and verify telemetry mTLS chain and app public TLS.

#### Task 10.1: Docker Compose Up (pki-init + shared-telemetry + one PS-ID)

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 9 complete
- **Description**: Start pki-init, shared-telemetry (OTel + Grafana), and one PS-ID deployment.
- **Acceptance Criteria**:
  - [ ] pki-init completes; all cert dirs present in `__PS_ID__-certs` volume
  - [ ] OTel Collector healthy (TLS active)
  - [ ] Grafana healthy (HTTPS active)
  - [ ] PS-ID app instances (all 4 variants) healthy

#### Task 10.2: Verify App→OTel→Grafana mTLS Pipeline

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 10.1
- **Description**: Verify telemetry flows end-to-end through mTLS chain.
- **Acceptance Criteria**:
  - [ ] All 4 PS-ID app variants send OTLP via mTLS to OTel (Cat 9 app certs)
  - [ ] OTel forwards to Grafana via mTLS (Cat 9 infra cert)
  - [ ] Traces visible in Grafana Tempo
  - [ ] Non-mTLS OTLP rejected by OTel
  - [ ] Grafana HTTPS UI accessible at `https://localhost:3000`

#### Task 10.3: Verify App Public HTTPS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 10.2
- **Description**: Verify each PS-ID app variant serves its public API over HTTPS with Cat 3 cert.
- **Acceptance Criteria**:
  - [ ] `GET /service/api/v1/health` via HTTPS returns 200 for all 4 variants
  - [ ] Server cert matches Cat 3 `public-https-server-entity-{PS-ID}-{variant}`
  - [ ] Connections without valid client cert rejected (Cat 4 CA enforcement)

---

### Phase 11: Knowledge Propagation [Status: ☐ TODO]

**Phase Objective**: Apply lessons learned to permanent artifacts.

#### Task 11.1: Review Lessons

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phases 1-10 complete
- **Description**: Review lessons.md entries from all prior phases.
- **Acceptance Criteria**:
  - [ ] All lessons reviewed
  - [ ] Actionable items identified for ENG-HANDBOOK.md

#### Task 11.2: Update ENG-HANDBOOK.md

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 11.1
- **Description**: Update ENG-HANDBOOK.md with OTel/Grafana mTLS and public app TLS patterns.
- **Acceptance Criteria**:
  - [ ] OTel Collector mTLS receiver config documented (§9.4 or new subsection)
  - [ ] Grafana HTTPS grafana.ini approach documented (D1)
  - [ ] OTel→Grafana mTLS forwarding documented
  - [ ] Public PS-ID app server TLS (Cat 3/Cat 4) documented
  - [ ] Combined V12+V13 cert mount least privilege table referenced

#### Task 11.3: Update deployment-templates.md

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 11.2
- **Acceptance Criteria**:
  - [ ] Final combined V12+V13 cert mount table verified accurate
  - [ ] grafana.ini template content documented

#### Task 11.4: Verify Propagation and Final Commit

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 11.3
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes
  - [ ] Clean working tree
  - [ ] All quality gates pass

---

## Cross-Cutting Tasks

### Testing

- [ ] Phase 0: pki-init generator tests pass (≥98% coverage)
- [ ] Phase 2: Framework OTLP exporter config unit tests pass (≥95%)
- [ ] Phase 7: Framework public server cert loading unit tests pass (≥95%)
- [ ] Phases 3, 6, 10: verification tasks pass in Docker Compose

### Code Quality

- [ ] Linting passes: `golangci-lint run ./...` and `golangci-lint run --build-tags e2e,integration ./...`
- [ ] No new TODOs without tracking
- [ ] Formatting clean

### Documentation

- [ ] `deployment-templates.md` updated with combined V12+V13 least privilege table
- [ ] `tls-structure.md` cross-referenced for cert category numbers
- [ ] ENG-HANDBOOK.md updated with OTel/Grafana/App TLS wiring patterns

### Deployment

- [ ] `lint-deployments` passes after Phase 8 template updates (Phase 9)
- [ ] Docker Compose health checks pass (Phase 10)
- [ ] `./configs/` unchanged (auto-TLS mode only — no changes to configs/)

---

## Notes

- **Least Privilege Enforcement**: Every compose template task MUST list exactly which Cat dirs are mounted and explicitly note what is NOT mounted.
- **./configs/ isolation**: No changes to `./configs/` files; they continue using auto-TLS only.
- **V13 depends on V12 Phase 0**: Cat 9 infra cert (`otel-collector-contrib-https-client-entity-infra`) generated in V12 Phase 0. V13 Phase 0 is independent (can run in parallel with V12 Phases 1-9).
- **D6 contingency**: If grafana/otel-lgtm does not support OTLP ingest TLS, pivot to OTel sidecar (D6=C) in Phase 4 Task 4.2.
- **V14 carry-forward (admin port YAML config)**: `internal/apps/framework/service/cli/` subcommands (`livez`, `readyz`, `shutdown`) are currently CLI-args-only. They should be extended to support YAML config files following the config-priority pattern (Docker secrets > YAML > CLI), allowing PS-ID instances to reuse their existing config files instead of passing all cert paths as CLI args. Track as a dedicated Phase in V15 or a separate V16 plan.

---

## Evidence Archive

- `test-output/v13-phase0/` — pki-init V13 cert generation verification
- `test-output/v13-phase3/` — OTel standalone mTLS verification
- `test-output/v13-phase6/` — OTel→Grafana pipeline verification
- `test-output/v13-phase10/` — Full deployment stack verification
