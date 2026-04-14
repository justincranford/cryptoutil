# Tasks — Framework v11: TLS Integration for Shared Services

**Status**: 0 of 37 tasks complete (0%)
**Created**: 2026-04-14
**Last Updated**: 2026-04-14

---

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥98% coverage/mutation for infrastructure)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions:**
- ✅ **Fix issues immediately** — when unknowns discovered, blockers identified, any tests fail, or
  quality gates not met, STOP and address
- ✅ **Treat as BLOCKING**: ALL issues block progress to next task
- ✅ **NEVER defer**: No "fix later", no "non-critical", no shortcuts

---

## Task Checklist

### Phase 1: Architecture Resolution

**Phase Objective**: Resolve the 6 open architectural decisions from quizme-v1 before any
implementation begins. Answers from user recorded in plan.md Decisions section.

**Status**: ☐ TODO — BLOCKED on quizme-v1 user answers

#### Task 1.1: Resolve Decision 1 — /certs Volume Strategy

- **Status**: ❌
- **Owner**: LLM Agent (after quizme-v1 answered)
- **Estimated**: 0.1h
- **Dependencies**: quizme-v1 Q1 answer from user
- **Description**: Record user's chosen volume strategy in plan.md; update all subsequent
  phase tasks to use the correct bind-mount or named-volume approach.
- **Acceptance Criteria**:
  - [ ] Decision 1 updated in plan.md with chosen option (A/B/C/D)
  - [ ] All phase tasks below aligned to chosen approach

#### Task 1.2: Resolve Decision 2 — PostgreSQL TLS Scope

- **Status**: ❌
- **Owner**: LLM Agent (after quizme-v1 answered)
- **Estimated**: 0.1h
- **Dependencies**: quizme-v1 Q2 answer from user
- **Description**: Record user's choice: mTLS, one-way TLS, or defer. Update Phase 4 tasks
  accordingly (mark Phase 4 deferred if Option C selected).
- **Acceptance Criteria**:
  - [ ] Decision 2 updated in plan.md
  - [ ] Phase 4 tasks activated or marked deferred

#### Task 1.3: Resolve Decision 3 — OTel mTLS vs One-Way TLS

- **Status**: ❌
- **Owner**: LLM Agent (after quizme-v1 answered)
- **Estimated**: 0.1h
- **Dependencies**: quizme-v1 Q3 answer from user
- **Description**: If mTLS: mark Task 2.4 (generator enhancement) as required. If one-way:
  mark Task 2.4 as skipped.
- **Acceptance Criteria**:
  - [ ] Decision 3 updated in plan.md

#### Task 1.4: Resolve Decisions 4, 5, 6

- **Status**: ❌
- **Owner**: LLM Agent (after quizme-v1 answered)
- **Estimated**: 0.1h
- **Dependencies**: quizme-v1 Q4/Q5/Q6 answers from user
- **Description**: Record OTel→Grafana transport, grafana/otel-lgtm stance, and pki-init
  placement decisions.
- **Acceptance Criteria**:
  - [ ] Decisions 4, 5, 6 updated in plan.md

---

### Phase 2: OTel Collector TLS Wiring

**Phase Objective**: Update `otel-collector-config.yaml` to enable TLS on gRPC 4317 and HTTP
4318 receivers; enable TLS on the OTel → Grafana exporter; add `/certs` mount to compose.

**⚠️ BLOCKED on Phase 1 (Decisions 1, 3, 4)**

#### Task 2.1: Update otel-collector-config.yaml — OTLP gRPC receiver TLS

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 1.3 (mTLS vs one-way TLS), Task 1.1 (/certs path)
- **Description**: Uncomment and populate `grpc:` receiver `tls:` stanza in
  `deployments/shared-telemetry/otel/otel-collector-config.yaml`.
  - `cert_file`: `ALL-telemetry-otel-private-server/ALL-telemetry-otel-receiver-private-server/ALL-telemetry-otel-receiver-private-server-crt.pem`
  - `key_file`: `...ALL-telemetry-otel-receiver-private-server-key.pem`
  - If mTLS (Decision 3A): add `client_ca_file: ALL-telemetry-otel-private-client/ALL-telemetry-otel-private-client-issuing-crt.pem`
- **Acceptance Criteria**:
  - [ ] gRPC TLS enabled with correct cert/key paths
  - [ ] `client_ca_file` added iff Decision 3 = mTLS
  - [ ] YAML valid (no extra indentation, correct nesting)

#### Task 2.2: Update otel-collector-config.yaml — OTLP HTTP receiver TLS

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Dependencies**: Task 2.1
- **Description**: Uncomment and populate `http:` receiver `tls:` stanza. Uses SAME cert/key
  as gRPC receiver (one cert serves both protocols simultaneously — confirmed from OTel docs).
- **Acceptance Criteria**:
  - [ ] HTTP TLS enabled with identical cert/key paths as gRPC
  - [ ] `client_ca_file` added iff Decision 3 = mTLS

#### Task 2.3: Update otel-collector-config.yaml — OTel→Grafana exporter TLS

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 1.4 (Decision 4 transport choice)
- **Description**: Update `otlphttp` exporter (or rename to `otlp` if Decision 4B):
  - Change endpoint from `http://grafana-otel-lgtm:4318` to HTTPS (or gRPC 4317)
  - Add `tls:` stanza with:
    - `ca_file`: `ALL-telemetry-grafana-private-server/ALL-telemetry-grafana-private-server-issuing-crt.pem`
    - `cert_file`: `ALL-telemetry-grafana-private-client/ALL-telemetry-otel-grafana-private-client/ALL-telemetry-otel-grafana-private-client-crt.pem`
    - `key_file`: `...ALL-telemetry-otel-grafana-private-client-key.pem`
  - Update service `pipelines` section if exporter name changed
- **Acceptance Criteria**:
  - [ ] Exporter TLS configured
  - [ ] Service pipelines reference correct exporter name

#### Task 2.4: [CONDITIONAL] Update generator.go — OTel mTLS client cert leaves

- **Status**: ❌ (Activate only if Decision 3 = mTLS)
- **Owner**: LLM Agent
- **Estimated**: 1.5h (includes tests)
- **Dependencies**: Task 1.3 (Decision 3 = mTLS)
- **Description**: If mTLS chosen: update `internal/apps/framework/tls/generator.go` to
  generate per-PS-ID OTel client certificate leaves under
  `ALL-telemetry-otel-private-client/` hierarchy. Each app instance (4 per PS-ID =
  40 leaves total) gets its own client cert for mTLS with OTel Collector.
  Update `docs/tls-structure.md` to document new leaves.
  Add tests in generator_test.go. Coverage ≥98%.
- **Acceptance Criteria**:
  - [ ] Generator creates leaves for all 40 app instances
  - [ ] Tests pass; coverage ≥98%
  - [ ] `docs/tls-structure.md` updated

#### Task 2.5: Update shared-telemetry compose.yml — OTel /certs mount + pki-init dep

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 1.1 (volume strategy), Task 1.4 (Decision 6 pki-init placement)
- **Description**: Update `deployments/shared-telemetry/compose.yml` for the
  `opentelemetry-collector-contrib` service:
  - Add `/certs` volume mount (bind mount or named volume per Decision 1)
  - Add pki-init service or `depends_on` per Decision 6
- **Acceptance Criteria**:
  - [ ] OTel container mounts `/certs`
  - [ ] pki-init dependency wired

---

### Phase 3: Grafana LGTM TLS Wiring

**Phase Objective**: Enable HTTPS for Grafana UI (port 3000); optionally enable TLS for
bundled OTel receiver inside the lgtm image.

**⚠️ BLOCKED on Phase 1 (Decisions 1, 4, 5)**

#### Task 3.1: Update shared-telemetry compose.yml — Grafana TLS env vars + /certs

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 1.1 (volume strategy), Task 1.4 (Decision 5 grafana stance)
- **Description**: Update `grafana-otel-lgtm` service in `shared-telemetry/compose.yml`:
  - Add env vars: `GF_SERVER_PROTOCOL: https`, `GF_SERVER_CERT_FILE`, `GF_SERVER_CERT_KEY`
  - Cert path: `ALL-telemetry-grafana-lgtm-public-server/ALL-telemetry-grafana-lgtm-public-server-crt.pem`
  - Add `/certs` volume mount (per Decision 1)
  - Update healthcheck to HTTPS: `curl -fk https://127.0.0.1:3000/api/health`
- **Acceptance Criteria**:
  - [ ] GF_SERVER_* env vars added
  - [ ] Volume mount for /certs added
  - [ ] Healthcheck updated to HTTPS with -k flag

#### Task 3.2: [CONDITIONAL] Create grafana-otel-lgtm/otelcol-config.yaml override

- **Status**: ❌ (Activate only if Decision 4 = TLS for OTel→Grafana receiver)
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 1.4 (Decision 4 = TLS enabled for OTel→Grafana link)
- **Description**: Create `deployments/shared-telemetry/grafana-otel-lgtm/otelcol-config.yaml`
  with TLS-enabled OTLP receiver config for the bundled OTel inside grafana-otel-lgtm image.
  Mount in compose at `/otel-lgtm/otelcol-config.yaml:ro`.
  Cert: `ALL-telemetry-grafana-private-server/ALL-telemetry-grafana-lgtm-private-server/`.
  `client_ca_file`: `ALL-telemetry-grafana-private-client` (for verifying OTel Collector client cert).
- **Acceptance Criteria**:
  - [ ] Custom YAML matches bundled OTel config format exactly
  - [ ] Compose mounts the file at correct path
  - [ ] Container starts successfully with override

#### Task 3.3: Add pki-init to shared-telemetry (if Decision 6A)

- **Status**: ❌ (Activate only if Decision 6 = standalone pki-init in shared-telemetry)
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 1.4 (Decision 6 = standalone)
- **Description**: Add `pki-init` service to `shared-telemetry/compose.yml` using suite
  binary with `--domain=cryptoutil --output-dir=/certs`.
  Reference suite binary from `cryptoutil:local` image.
  OTel and Grafana services `depends_on: pki-init → service_completed_successfully`.
- **Acceptance Criteria**:
  - [ ] pki-init service added
  - [ ] OTel and Grafana depend on pki-init completion

---

### Phase 4: PostgreSQL TLS Wiring

**Phase Objective**: Update shared-postgres and all PS-ID postgres-url.secret files to
use TLS instead of `sslmode=disable`.

**⚠️ STATUS DEPENDS ON Decision 2 (defer to v12 if Option C)**

#### Task 4.1: Update shared-postgres postgresql.conf — enable SSL

- **Status**: ❌ (Activate if Decision 2 = A or B)
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 1.2 (Decision 2 ≠ defer)
- **Description**: Update `deployments/shared-postgres/postgresql-leader.conf` and
  `postgresql-follower.conf` to enable SSL:
  - `ssl = on`
  - `ssl_cert_file = '/certs/ALL-db-postgres-private-server/.../...-crt.pem'`
  - `ssl_key_file = '/certs/ALL-db-postgres-private-server/.../...-key.pem'`
  - `ssl_ca_file` for follower and leader client CA (if Decision 2A = mTLS)
- **Acceptance Criteria**:
  - [ ] SSL enabled in both conf files

#### Task 4.2: Update shared-postgres compose.yml — /certs mount + pki-init dep

- **Status**: ❌ (Activate if Decision 2 = A or B)
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 4.1
- **Description**: Add `/certs` volume mount to `postgres-leader` and `postgres-follower`
  services. Add pki-init dependency. Volume strategy from Decision 1.
- **Acceptance Criteria**:
  - [ ] Both postgres containers mount /certs
  - [ ] pki-init dependency added

#### Task 4.3: Update pg_hba.conf — require SSL (if mTLS, Decision 2A)

- **Status**: ❌ (Activate ONLY if Decision 2 = full mTLS option A)
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 4.1
- **Description**: Update `pg_hba.conf` to `hostssl ... clientcert=verify-full` for all
  connectios to require client certificates.
- **Acceptance Criteria**:
  - [ ] pg_hba.conf entries use `hostssl` or `hostnossl`

#### Task 4.4: Update all 10 postgres-url.secret files

- **Status**: ❌ (Activate if Decision 2 = A or B)
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 4.1, Task 1.2 (TLS mode)
- **Description**: Update `deployments/<PS-ID>/secrets/postgres-url.secret` for all 10
  PS-IDs. Remove `?sslmode=disable`. Add:
  - Decision 2B (server cert only): `?sslmode=verify-ca&sslrootcert=/certs/ALL-db-postgres-private-server/ALL-db-postgres-private-server-ca-crt.pem`
  - Decision 2A (mTLS): add `&sslcert=...&sslkey=...` for client cert based on PS-ID domain
- **Files**: 10 × `deployments/<PS-ID>/secrets/postgres-url.secret`
- **Acceptance Criteria**:
  - [ ] All 10 files updated
  - [ ] `?sslmode=disable` removed from all

---

### Phase 5: PS-ID Service Compose TLS Trust

**Phase Objective**: Update all 10 PS-ID service compose.yml files to add `/certs` CA trust
paths for OTel, Grafana, and PostgreSQL to app service environment or config.

**⚠️ BLOCKED on Phase 1 (all decisions)**

#### Task 5.1: Identify OTel client cert strategy per app instance

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 1.3 (Decision 3)
- **Description**: Based on Decision 3 (mTLS or one-way), determine what each app service
  needs from `/certs` for OTel connectivity. Document exact env var or config key needed
  to pass CA cert path (otel.yml or `--otlp-insecure=false`).
- **Acceptance Criteria**:
  - [ ] OTel trust approach documented for app services

#### Task 5.2: Update PS-ID service compose.yml — OTel TLS config

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Task 5.1
- **Description**: For all 10 PS-ID compose.yml files, update the 4 app instance services
  (`-sqlite-1`, `-sqlite-2`, `-postgresql-1`, `-postgresql-2`) to:
  - Pass correct OTel endpoint with TLS (`--otlp-insecure=false` or config file)
  - Trust `ALL-telemetry-otel-private-server` issuing CA
  - If mTLS: include client cert path from `/certs/<PS-ID>-...` per app instance
- **Files**: 10 × `deployments/<PS-ID>/compose.yml`
- **Acceptance Criteria**:
  - [ ] All 40 app service instances updated
  - [ ] No `--otlp-insecure=true` remaining in any compose file

---

### Phase 6: Template Registry Updates

**Phase Objective**: Mirror ALL changes from Phases 2–5 in
`api/cryptosuite-registry/templates/`.

**⚠️ BLOCKED on Phases 2–5**

#### Task 6.1: Update templates/deployments/shared-telemetry/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Phases 2, 3
- **Description**: Update `templates/deployments/shared-telemetry/compose.yml` and
  `templates/deployments/shared-telemetry/otel/otel-collector-config.yaml` to exactly match
  the actual deployment files.
  Create `templates/deployments/shared-telemetry/grafana-otel-lgtm/otelcol-config.yaml` if
  Phase 3 Task 3.2 created this file.
- **Acceptance Criteria**:
  - [ ] Template files identical to actual deployment files
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes

#### Task 6.2: Update templates/deployments/**PS_ID**/

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Phase 5
- **Description**: Update `templates/deployments/__PS_ID__/compose.yml` to match PS-ID
  compose changes for OTel TLS config.
- **Acceptance Criteria**:
  - [ ] Template matches actual PS-ID compose structure
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes

#### Task 6.3: Update templates postgres-url.secret

- **Status**: ❌ (Activate if Decision 2 ≠ defer)
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Dependencies**: Phase 4
- **Description**: Update `templates/deployments/__PS_ID__/secrets/postgres-url.secret` to
  match new format without `?sslmode=disable`.
- **Acceptance Criteria**:
  - [ ] Template secret matches updated format

#### Task 6.4: Update templates/deployments/shared-postgres/ (if Decision 2 ≠ defer)

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Phase 4
- **Description**: Mirror shared-postgres compose and conf file changes in templates.
- **Acceptance Criteria**:
  - [ ] Template files match actual shared-postgres changes

---

### Phase 7: Documentation Updates

**Phase Objective**: Update `docs/tls-structure.md`, `docs/target-structure.md`, and
`docs/deployment-templates.md` to reflect v11 changes.

**⚠️ BLOCKED on Phases 2–5**

#### Task 7.1: Update docs/tls-structure.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Phases 2–5
- **Description**: Add/update:
  - Note that OTel gRPC (4317) and HTTP (4318) receivers share the same cert/key pair
  - Note that ports 13133/1777/55679 are plain HTTP only — no TLS ever for these extension ports
  - Add any new cert entries from Task 2.4 (if mTLS OTel client leaves added)
  - Confirm all `ALL-*` shared domain cert paths match actual generator output
- **Acceptance Criteria**:
  - [ ] OTel single-cert note added
  - [ ] Extension plain-HTTP note added
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 7.2: Update docs/target-structure.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Phases 2–5
- **Description**: Add `/certs` volume directory tree section showing the full layout pki-init
  produces. Include all `ALL-*` directories and sample PS-ID directories.
- **Acceptance Criteria**:
  - [ ] `/certs` tree section added
  - [ ] All shared `ALL-*` dirs documented

#### Task 7.3: Update docs/deployment-templates.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Phases 2–5
- **Description**: Update the deployment-templates.md sections covering shared-telemetry and
  shared-postgres templates:
  - Document new `GF_SERVER_*` env vars
  - Document postgres-url.secret format change
  - Document pki-init dependency in shared stacks
  - Document OTel TLS exporter configuration
- **Acceptance Criteria**:
  - [ ] All new template patterns documented
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

---

### Phase 8: cicd-lint Evaluation

**Phase Objective**: Evaluate and update cicd-lint validators for any v11 changes.

#### Task 8.1: Evaluate template-compliance linter for new files

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Phase 6
- **Description**: Run `go run ./cmd/cicd-lint lint-fitness` after all template changes.
  Verify `template-compliance` linter passes for:
  - New `grafana-otel-lgtm/otelcol-config.yaml` (if created)
  - Updated `otel-collector-config.yaml`
  - All PS-ID compose changes
- **Acceptance Criteria**:
  - [ ] `lint-fitness` passes with zero errors

#### Task 8.2: Evaluate lint-deployments for postgres-url changes

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Phase 4
- **Description**: Run `go run ./cmd/cicd-lint lint-deployments`. Verify `validate-secrets`
  does not flag `sslrootcert=...` path in postgres-url.secret as an inline secret (it's a
  cert path, not a credential value).
  If false positive triggered: update the allowlist pattern in `validate_secrets.go`.
- **Acceptance Criteria**:
  - [ ] `lint-deployments` passes with zero errors
  - [ ] No false-positive secret detection on cert file paths

---

### Phase 9: Quality Gates

**Phase Objective**: Run full quality gate suite and confirm zero failures.

#### Task 9.1: Full build check

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Dependencies**: All prior phases
- **Description**:
  `go build ./...` AND `go build -tags e2e,integration ./...`
- **Acceptance Criteria**:
  - [ ] Zero build errors

#### Task 9.2: Linting check

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Dependencies**: Task 9.1
- **Description**:
  `golangci-lint run ./...` AND `golangci-lint run --build-tags e2e,integration ./...`
- **Acceptance Criteria**:
  - [ ] Zero lint warnings

#### Task 9.3: cicd-lint full suite check

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Dependencies**: Tasks 8.1, 8.2
- **Description**:
  `go run ./cmd/cicd-lint lint-fitness lint-deployments lint-docs`
- **Acceptance Criteria**:
  - [ ] All three linters pass with zero errors

---

### Phase 10: Knowledge Propagation

**Phase Objective**: Apply all lessons from phase post-mortems to permanent artifacts.

#### Task 10.1: Review lessons.md and identify propagation targets

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Phase 9
- **Description**: Read `docs/framework-v11/lessons.md` in full. Identify specific
  ENG-HANDBOOK.md sections, instruction files, and code that need updates.
- **Acceptance Criteria**:
  - [ ] Propagation target list compiled

#### Task 10.2: Update ENG-HANDBOOK.md with OTel TLS patterns

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Dependencies**: Task 10.1
- **Description**: Add/update Section 9.4.1 with:
  - OTel gRPC+HTTP single cert pattern (confirmed)
  - Health-check extension plain HTTP invariant (13133 never gets TLS)
  - Grafana `GF_SERVER_*` env var TLS configuration pattern
  - Volume strategy decision outcome (from Decision 1)
  - PostgreSQL TLS connection string pattern (from Decision 2 outcome)
- **Acceptance Criteria**:
  - [ ] ENG-HANDBOOK.md sections updated
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes

#### Task 10.3: Update instruction files

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Dependencies**: Task 10.2
- **Description**: Update relevant instruction files per `@propagate` system:
  - `02-03.observability.instructions.md` — OTel TLS patterns
  - `02-05.security.instructions.md` — if PostgreSQL TLS established new pattern
  - `04-01.deployment.instructions.md` — shared-telemetry TLS wiring patterns
- **Acceptance Criteria**:
  - [ ] Instruction files updated
  - [ ] Propagation validated: `go run ./cmd/cicd-lint lint-docs`

#### Task 10.4: Final git commit

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Dependencies**: Task 10.3
- **Description**: Stage all changes; separate semantic commits; verify clean worktree.
- **Acceptance Criteria**:
  - [ ] `git status --porcelain` returns empty
  - [ ] All commits use conventional format

---

## Cross-Cutting Tasks

### Testing

- [ ] `go test ./...` passes — 100% passing, zero skips
- [ ] `go test ./internal/apps/framework/tls/...` passes (if generator changes)
- [ ] Coverage ≥98% for `internal/apps/framework/tls/` (if generator changed)
- [ ] No race conditions: `go test -race ./internal/apps/framework/tls/...`

### Code Quality

- [ ] `golangci-lint run ./...` — zero warnings
- [ ] `golangci-lint run --build-tags e2e,integration ./...` — zero warnings
- [ ] YAML files are valid (use `python -c "import yaml; yaml.safe_load(open('...'))"` to verify)
- [ ] No inline secrets in any compose file (validated by `lint-deployments validate-secrets`)

### Documentation

- [ ] `docs/tls-structure.md` updated with new cert notes
- [ ] `docs/target-structure.md` updated with /certs tree
- [ ] `docs/deployment-templates.md` updated with TLS patterns
- [ ] All doc changes pass `go run ./cmd/cicd-lint lint-docs`

---

## Notes / Deferred Work

- **Phase 4 (PostgreSQL TLS)**: May be deferred entirely to v12 if user selects Decision 2C.
  If deferred, create a tracking item in `docs/` or the tracker to pick it up as v12 scope.
- **grafana/otel-lgtm production migration**: If user selects Decision 5C (keep but plan
  migration), create a v12 scope doc referencing the production stack design.
- **Task 2.4 (OTel mTLS generator change)**: Only activates for Decision 3A (full mTLS).
  If one-way TLS selected, this task is permanently skipped in v11.
- **Standalone grafana/otel-lgtm bundled OTel TLS**: The bundled OTel collector inside
  grafana/otel-lgtm uses the same config format as standalone OTel Collector Contrib.
  The `/otel-lgtm/otelcol-config.yaml` override path is confirmed from docker-otel-lgtm README.
