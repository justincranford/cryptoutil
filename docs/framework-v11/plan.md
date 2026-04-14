# Implementation Plan — Framework v11: TLS Integration for Shared Services

**Status**: Planning
**Created**: 2026-04-14
**Last Updated**: 2026-04-14
**Purpose**: Wire pki-init–generated TLS certificates into the shared-telemetry (OTel Collector
Contrib + Grafana OTEL LGTM) and shared-postgres compose stacks; update all PS-ID service
compose files to trust the shared service CAs and use TLS for OTel and database connections;
mirror all changes in the canonical template registry; update documentation.

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
- ✅ **Treat as BLOCKING**: ALL issues block progress to next phase or task
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no shortcuts

---

## Overview

Framework v10 established the canonical template registry and pki-init CLI. The pki-init generator
(`internal/apps/framework/tls/generator.go`) already generates **all** required TLS material for
shared services: OTel Collector Contrib, Grafana OTEL LGTM, and PostgreSQL leader/follower. The
`generateSharedDomains()` function is called unconditionally for every domain, so TLS certs for
shared services are always produced regardless of which `--domain` value is passed to pki-init.

Framework v11 closes the wiring gap: the certs exist but are not yet mounted or configured in any
compose file. v11 connects the generated certs to the services that need them.

**Key principle**: Every pki-init domain (16 total: 1 suite + 5 products + 10 PS-IDs) generates
the same `ALL-*` shared service certs in addition to its domain-specific ones. Shared services
(OTel Collector, Grafana LGTM, PostgreSQL) must use the `ALL-*` certs because they serve all
domains simultaneously.

---

## Background

**Completed in v10**:
- `api/cryptosuite-registry/templates/` canonical template registry (~63 files)
- pki-init CLI with pflag — `--domain` and `--output-dir` flags
- `internal/apps/framework/tls/generator.go` — full cert generation for all shared + PS-ID domains
- `docs/tls-structure.md` — documents full `/certs` volume layout
- `docs/deployment-templates.md` — documents template structure and placeholders
- All PS-ID compose templates reference pki-init and `./certs/` bind mount (PS-ID-local)
- .gitkeep files in `api/cryptosuite-registry/templates/` dirs with siblings already deleted

**Outstanding v10 debt carried into v11**:
- `deployments/shared-telemetry/compose.yml` has NO pki-init dependency and NO `/certs` volume
- `deployments/shared-telemetry/otel/otel-collector-config.yaml` has TLS commented out
- `deployments/shared-postgres/compose.yml` uses password auth only (no TLS)
- `api/cryptosuite-registry/templates/deployments/shared-telemetry/` mirrors this gap
- All PS-ID `postgres-url.secret` files still contain `?sslmode=disable`

---

## Technical Context

- **Language**: Go 1.26.1
- **Framework**: `internal/apps/framework/tls/`
- **Key magic constants**:
  - `cryptoutilSharedMagic.DockerServiceOtelCollector` = `"opentelemetry-collector-contrib"`
  - `cryptoutilSharedMagic.DockerServiceGrafanaOtelLgtm` = `"grafana-otel-lgtm"`
  - `cryptoutilSharedMagic.DefaultOTLPServiceDefault` = `"cryptoutil"` (suite domain ID)
- **OTel Collector image**: `otel/opentelemetry-collector-contrib:latest`
  - Standalone service in `deployments/shared-telemetry/`; config at
    `deployments/shared-telemetry/otel/otel-collector-config.yaml`
  - TLS is commented out for gRPC 4317 and HTTP 4318 receivers
  - Ports 13133 (health_check), 1777 (pprof), 55679 (zpages) — these extensions use **plain HTTP
    only** (no TLS; compose healthcheck on 13133 correctly stays HTTP forever)
  - Exporter: `otlphttp` to `http://grafana-otel-lgtm:4318` (no TLS currently)
- **Grafana image**: `grafana/otel-lgtm:latest` (bundles OTel Collector + Prometheus + Tempo
  - Loki + Pyroscope + Grafana)
  - Grafana port 3000 TLS: env vars `GF_SERVER_PROTOCOL=https`, `GF_SERVER_CERT_FILE`,
    `GF_SERVER_CERT_KEY`
  - Internal bundled OTel OTLP receiver TLS: override `/otel-lgtm/otelcol-config.yaml` by
    bind-mounting a custom config file
  - Healthcheck currently: `curl -f http://127.0.0.1:3000/api/health` → must change to HTTPS
    after TLS enabled
- **PostgreSQL image**: `postgres:18`
  - Docker service names: `postgres-leader`, `postgres-follower` (match cert SANs exactly)
  - Current postgres-url.secret format:
    `postgres://user:pass@shared-postgres-leader:5432/db?sslmode=disable`
- **Cert paths in /certs (relevant to shared services)**:
  - OTel server cert:
    `ALL-telemetry-otel-private-server/ALL-telemetry-otel-receiver-private-server/`
    (SAN: `opentelemetry-collector-contrib`, `localhost`)
  - OTel server issuing CA:
    `ALL-telemetry-otel-private-server/ALL-telemetry-otel-private-server-issuing-crt.pem`
  - OTel client CA (for mTLS):
    `ALL-telemetry-otel-private-client/ALL-telemetry-otel-private-client-issuing-crt.pem`
  - OTel→Grafana client cert:
    `ALL-telemetry-grafana-private-client/ALL-telemetry-otel-grafana-private-client/`
  - Grafana public server cert (port 3000):
    `ALL-telemetry-grafana-lgtm-public-server/ALL-telemetry-grafana-lgtm-public-server-crt.pem`
  - Grafana private server cert (internal OTel receiver port):
    `ALL-telemetry-grafana-private-server/ALL-telemetry-grafana-lgtm-private-server/`
  - PostgreSQL leader server cert:
    `ALL-db-postgres-private-server/` (leader + follower leaves)
  - PostgreSQL client CA (leader):
    `ALL-db-postgresql-leader-private-client/ALL-db-postgresql-leader-private-client-issuing-crt.pem`
  - PostgreSQL client CA (follower):
    `ALL-db-postgresql-follower-private-client/ALL-db-postgresql-follower-private-client-issuing-crt.pem`
  - Public app server CA (issuing CA for PS-ID HTTPS endpoints; PS-ID services trust this):
    `ALL-app-public-server/ALL-app-public-server-issuing-crt.pem`

---

## Architecture Decisions

### Decision 1: /certs Volume Strategy

**Problem**: The PS-ID compose template uses a bind mount `./certs/:/certs/:rw` for pki-init
and `./certs/:/certs/:ro` for app containers. When `shared-telemetry/compose.yml` is included
into a PS-ID compose, paths in the included file resolve relative to `shared-telemetry/` — a
DIFFERENT directory. OTel Collector and Grafana would look in `deployments/shared-telemetry/certs/`
while pki-init writes to `deployments/sm-kms/certs/`.

**Options**:
- A: Named Docker volume `certs` — pki-init: `certs:/certs:rw`; all consumers: `certs:/certs:ro`
- B: Bind mount to shared canonical path — pki-init: `../shared-telemetry/certs:/certs:rw`;
  shared-telemetry services: `./certs:/certs:ro`; all app containers: `../shared-telemetry/certs:/certs:ro`
- C: pki-init stays PS-ID-local; shared-telemetry pki-init runs separately with `--domain=cryptoutil`
  writing to `./certs/`; each PS-ID's pki-init only serves PS-ID-specific certs
- D: Other (user specifies)

**Status**: ⚠️ **OPEN — see quizme-v1 Q1**

### Decision 2: PostgreSQL TLS Scope in v11

**Problem**: Enabling PostgreSQL mTLS requires: (a) updating `shared-postgres`'s `pg_hba.conf`
and `postgresql.conf` to require SSL/client certs, (b) changing postgres-url.secret format for
all 10 PS-IDs, (c) potentially adding pki-init dependency to shared-postgres compose. This is
non-trivial and a separate concern from OTel/Grafana TLS.

**Options**:
- A: Full mTLS — update shared-postgres + all postgres-url.secret + trust CA chain in GORM
- B: One-way TLS only — postgres requires server cert, clients verify but don't present client cert
  (`sslmode=verify-ca`); simpler; removes `?sslmode=disable` without requiring client cert management
- C: Defer entirely to v12 — only OTel + Grafana TLS in v11; postgres stays `?sslmode=disable`

**Status**: ⚠️ **OPEN — see quizme-v1 Q2**

### Decision 3: OTel Collector mTLS vs One-Way TLS (PS-ID → OTel path)

**Problem**: The generator creates `ALL-telemetry-otel-private-client` CA but **no per-PS-ID
app OTel client certificate leaves**. For mTLS, each app instance would need a client cert issued
by this CA. One-way TLS requires no client certs but all connections are still encrypted.

**Options**:
- A: mTLS — generator must also create per-PS-ID OTel client cert leaves; includes code change to
  `generator.go`; each app instance needs its own client cert; compose configs become more complex
- B: One-way TLS — PS-ID apps verify OTel server cert; OTel no `client_ca_file`; generator unchanged;
  simpler compose configs
- C: Skip OTel TLS for now (not recommended — user request explicitly includes OTel TLS)

**Status**: ⚠️ **OPEN — see quizme-v1 Q3**

### Decision 4: OTel Collector → Grafana LGTM TLS Protocol

**Problem**: The current exporter uses `otlphttp` to `http://grafana-otel-lgtm:4318`. The OTel
Collector image supports both HTTP and gRPC. The `grafana/otel-lgtm` bundled OTel receiver
supports both ports (4317 gRPC, 4318 HTTP) — but internal TLS config requires mounting a custom
`/otel-lgtm/otelcol-config.yaml`.

**Options**:
- A: `otlphttp` with HTTPS to port 4318 — minimal change from current (just add `tls:` stanza
  and `https://` URL); HTTP/1.1 overhead minor for internal traffic
- B: `otlp` gRPC to port 4317 with TLS — more efficient; rename exporter; change service pipeline
- C: No TLS for OTel→Grafana link — both are internal Docker network services; TLS only on the
  external PS-ID→OTel path; simpler; note: Grafana UI (port 3000) still gets TLS

**Status**: ⚠️ **OPEN — see quizme-v1 Q4**

### Decision 5: grafana/otel-lgtm Image in Production Context

**Problem**: `grafana/otel-lgtm` README explicitly states: *"intended for development, demo,
and testing environments."* It bundles multiple services in a single container (OTel Collector,
Prometheus, Tempo, Loki, Pyroscope, Grafana). This is not a production-grade design.

**Options**:
- A: Keep `grafana/otel-lgtm` — acceptable for this project's scope; add documentation caveat
- B: Migrate to separate services in v11 — `grafana/grafana`, `prom/prometheus`, `grafana/tempo`,
  `grafana/loki`, standalone OTel Collector only; v11 scope grows substantially
- C: Keep `grafana/otel-lgtm` for v11 but plan migration as v12 scope; document the concern

**Status**: ⚠️ **OPEN — see quizme-v1 Q5**

### Decision 6: pki-init in shared-telemetry compose.yml

**Problem**: Regardless of the volume strategy (Decision 1), the question is whether a pki-init
invocation should live *inside* `shared-telemetry/compose.yml` or only in PS-ID compose files.

**Context**: The PS-ID template already has a `pki-init` service. When PS-ID compose includes
shared-telemetry, both run. But for standalone shared-telemetry use (e.g., `docker compose -f
shared-telemetry/compose.yml up`), there is no pki-init unless shared-telemetry has its own.

**Options**:
- A: Add standalone `pki-init` service to `shared-telemetry/compose.yml` using the suite
  binary (`--domain=cryptoutil`) — independent; always generates shared `ALL-*` certs
- B: Rely exclusively on PS-ID compose includes — shared-telemetry cannot be started standalone
  with TLS (acceptable if it's always included)
- C: Add `pki-init` to shared-telemetry that generates shared-only certs; PS-ID init only
  generates PS-ID-specific certs; total generation is union

**Status**: ⚠️ **OPEN — see quizme-v1 Q6**

---

## Phases

### Phase 0: Prerequisites (Completed Before v11 Begins)

**Status**: ✅ DONE
- Template .gitkeep files in directories with siblings already deleted
- `docs/framework-v10` complete (100% tasks done per tasks.md)
- pki-init generates all shared service TLS material in `generateSharedDomains()`

---

### Phase 1: Architecture Resolution (0.5h)

**Status**: ☐ TODO — **BLOCKED on quizme-v1**

**Objective**: Resolve the 6 open architectural decisions (see Decisions 1–6 above) from
quizme-v1 answers before any implementation begins. Record decisions in this plan.

**Success**: All 6 decisions resolved; implementation phases can proceed in parallel.

**Post-Mortem**: After quality gates pass, update lessons.md — what worked, what didn't,
root causes, patterns. Evaluate for contradictions/omissions; create fix tasks immediately.

---

### Phase 2: OTel Collector TLS Wiring (2h)

**Status**: ☐ TODO — **BLOCKED on Phase 1**

**Objective**: Enable TLS for OTel Collector Contrib's gRPC receiver (port 4317) and HTTP
receiver (port 4318), and TLS for the OTel → Grafana LGTM exporter connection.

**Implementation**:
- Update `deployments/shared-telemetry/otel/otel-collector-config.yaml`:
  - Uncomment and fill TLS stanzas for both `grpc:` and `http:` OTLP receivers
  - Same cert/key pair for both: `ALL-telemetry-otel-receiver-private-server`
  - Add `client_ca_file` if Decision 3 = mTLS; omit if one-way TLS
  - Change `otlphttp` exporter endpoint to HTTPS (or change to `otlp` gRPC per Decision 4)
  - Add `tls:` stanza to exporter using `ALL-telemetry-otel-grafana-private-client` cert
- Update `deployments/shared-telemetry/compose.yml`:
  - Add `/certs` volume mount to `opentelemetry-collector-contrib` container (per Decision 1)
  - Add pki-init dependency (per Decision 6)
- Mirror changes in `api/cryptosuite-registry/templates/deployments/shared-telemetry/`

**OTel healthcheck note**: Port 13133 (health_check extension) is plain HTTP ONLY — no TLS
config available for this extension. Compose healthcheck `wget http://127.0.0.1:13133/` remains
HTTP. This is correct and expected; do NOT attempt to convert this to HTTPS.

**Success**: OTel Collector starts with TLS-enabled receivers; accepts OTLP/TLS from a test client.

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 3: Grafana LGTM TLS Wiring (1.5h)

**Status**: ☐ TODO — **BLOCKED on Phase 1**

**Objective**: Enable HTTPS for Grafana UI (port 3000) and optionally TLS for the bundled OTel
Collector receiver (ports 4317/4318 inside the lgtm image) per Decision 4/5.

**Implementation**:
- Update `deployments/shared-telemetry/compose.yml` for `grafana-otel-lgtm` service:
  - Add `GF_SERVER_PROTOCOL: https`
  - Add `GF_SERVER_CERT_FILE: /certs/ALL-telemetry-grafana-lgtm-public-server/ALL-telemetry-grafana-lgtm-public-server-crt.pem`
  - Add `GF_SERVER_CERT_KEY: /certs/ALL-telemetry-grafana-lgtm-public-server/ALL-telemetry-grafana-lgtm-public-server-key.pem`
  - Add `/certs` volume mount (per Decision 1)
  - Update healthcheck: `curl -fk https://127.0.0.1:3000/api/health` (use `-k` since local CA)
  - If Decision 4 = TLS for OTel→Grafana link: create
    `deployments/shared-telemetry/grafana-otel-lgtm/otelcol-config.yaml` override and mount
    it at `/otel-lgtm/otelcol-config.yaml:ro` in the container
  - If Decision 4B (gRPC): change `14317:4317` port mapping label comment
- Mirror changes in `api/cryptosuite-registry/templates/deployments/shared-telemetry/`

**Success**: Grafana UI accessible via HTTPS; browser shows valid TLS (trusted by `ALL-app-public-server` CA).

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 4: PostgreSQL TLS Wiring (2.5h) [Decision 2 determines scope]

**Status**: ☐ TODO — **BLOCKED on Phase 1 (Decision 2)**

**Option A (mTLS)**: Update `deployments/shared-postgres/` (`pg_hba.conf` / `postgresql.conf`
to require SSL), add pki-init dependency to `shared-postgres/compose.yml`, update all 10
`deployments/<PS-ID>/secrets/postgres-url.secret` files, update 10 template equivalents.

**Option B (one-way TLS)**: Update `shared-postgres/postgresql.conf` to enable SSL with server
cert only; update `postgres-url.secret` files to `sslmode=verify-ca&sslrootcert=...`; no
client cert required; simpler.

**Option C (defer)**: No changes in v11; ticket created for v12.

**Success (A or B)**: `go test ./...` passes; PostgreSQL connections use verified TLS.

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 5: PS-ID Service Compose TLS Trust (1.5h)

**Status**: ☐ TODO — **BLOCKED on Phase 1**

**Objective**: Update all 10 PS-ID `deployments/<PS-ID>/compose.yml` files and their template
counterpart to: (a) trust shared service CAs (public app server CA, OTel private CA, Grafana
private CA), (b) mount `/certs` (per Decision 1 volume strategy), and (c) pass the correct
`--otlp-insecure=false` flag with CA trust path in the OTel config.

**Note on OTel client cert (Decision 3)**:
- If mTLS: Each app instance mounting `/certs` needs its own client cert leaf generated by
  the OTel mTLS generator enhancement (from Decision 3A). PS-ID OTel config must specify
  client cert path.
- If one-way TLS: App just needs CA cert for server verification; no client cert changes.

**Note on pki-init dependency**: App services already `depend_on: pki-init`. The pki-init
service in the PS-ID template already runs before apps start. No structural change needed;
only the `/certs` volume content expands to include `ALL-*` paths (already true, always was).

**Success**: PS-ID services start, connect to OTel via TLS; linting clean.

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 6: Template Registry Updates (1h)

**Status**: ☐ TODO — **BLOCKED on Phases 2–5**

**Objective**: Update `api/cryptosuite-registry/templates/` to mirror ALL changes from Phases
2–5 exactly. The canonical template is always in sync with the actual deployment file.

**Affected template files**:
- `templates/deployments/shared-telemetry/compose.yml`
- `templates/deployments/shared-telemetry/otel/otel-collector-config.yaml`
- `templates/deployments/shared-telemetry/grafana-otel-lgtm/otelcol-config.yaml` (if Phase 3B)
- `templates/deployments/__PS_ID__/compose.yml`
- `templates/deployments/__PS_ID__/secrets/postgres-url.secret`
- `templates/deployments/shared-postgres/compose.yml` (if Phase 4A or 4B)

**Success**: `go run ./cmd/cicd-lint lint-fitness` passes (template-compliance linter).

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 7: Documentation Updates (1.5h)

**Status**: ☐ TODO — **BLOCKED on Phases 2–5**

**Objective**: Update docs to reflect v11 changes:
1. `docs/tls-structure.md` — Add notes: OTel gRPC+HTTP share same cert/key pair; ports 13133/1777/55679
   are plain HTTP only (no TLS ever); confirm all `ALL-*` shared domain cert paths
2. `docs/target-structure.md` — Add `/certs` directory tree section showing layout that
   `pki-init --domain=<any>` produces
3. `docs/deployment-templates.md` — Add TLS cert reference sections for shared-telemetry and
   shared-postgres; document new GF_SERVER_* env vars; document postgres-url.secret format change

**Success**: `go run ./cmd/cicd-lint lint-docs` passes.

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 8: cicd-lint Updates (1h)

**Status**: ☐ TODO

**Objective**: Evaluate whether any cicd-lint fitness functions need updates for v11:
1. `lint-fitness` — `template_drift` linter: ensure it validates new shared-telemetry template
   files match actual compose; no behavior change needed if template = actual
2. `lint-deployments` — `validate-secrets`: check if new `sslrootcert=` path in postgres-url
   is detected as inline secret (it's a path, not a value — should not trigger)
3. `lint_compose` — `check_admin_port` and other validators: verify no false positives

**Note**: The `template_drift.go` linter already explicitly skips `.gitkeep` files (confirmed
no blocking issue from .gitkeep deletion). No `.gitkeep`-related cicd-lint changes needed.

**Success**: All cicd-lint commands pass with zero errors.

**Post-Mortem**: After quality gates pass, update lessons.md with lessons learned.

---

### Phase 9: Quality Gates (0.5h)

**Status**: ☐ TODO — **BLOCKED on Phases 2–8**

**Checklist**:
- [ ] `go build ./...` — zero errors
- [ ] `go build -tags e2e,integration ./...` — zero errors
- [ ] `golangci-lint run ./...` — zero warnings
- [ ] `golangci-lint run --build-tags e2e,integration ./...` — zero warnings
- [ ] `go run ./cmd/cicd-lint lint-fitness` — zero errors
- [ ] `go run ./cmd/cicd-lint lint-deployments` — zero errors
- [ ] `go run ./cmd/cicd-lint lint-docs` — zero errors
- [ ] All 10 PS-ID compose.yml files validated
- [ ] `api/cryptosuite-registry/templates/` in sync with actual deployments

---

### Phase 10: Knowledge Propagation (0.5h)

**Status**: ☐ TODO — **BLOCKED on Phase 9**

**Objective**: Apply lessons from all prior phases permanently.

- Review `docs/framework-v11/lessons.md` from all phase post-mortems
- Update `docs/ENG-HANDBOOK.md` with new patterns discovered (OTel TLS config, Grafana TLS env
  vars, volume strategy decision, PostgreSQL mTLS approach)
- Update `.github/instructions/02-03.observability.instructions.md` with confirmed OTel TLS
  patterns (gRPC+HTTP same cert, health_check plain HTTP invariant)
- Update `.github/instructions/02-05.security.instructions.md` if PostgreSQL TLS approach
  establishes a new pattern
- Update `.github/instructions/04-01.deployment.instructions.md` with shared-telemetry TLS
  wiring patterns
- Verify propagation: `go run ./cmd/cicd-lint lint-docs`
- Commit all artifact updates with separate semantic commits per artifact type

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Volume path mismatch (pki-init vs shared-telemetry) | High | High | Decision 1 resolves; named volume or canonical bind path |
| grafana/otel-lgtm bundled OTel TLS config format unknown | Medium | Medium | Web research confirmed `/otel-lgtm/otelcol-config.yaml` override; same YAML format as standalone OTel Collector |
| `grafana/otel-lgtm` dev/test image stability | Low | Medium | Keep for v11; document caveat; plan migration in v12 if needed |
| PostgreSQL TLS scope bloat | Medium | High | Decision 2 hard-copes; defer if scope is too large |
| OTel mTLS requires generator code change | Medium | Medium | Decision 3: one-way TLS avoids generator change entirely |
| compose healthcheck 13133 HTTP always | Known | Low | OTel health_check extension has no TLS support; accept and document |
| Grafana HTTPS healthcheck curl certificate | Low | Low | Use `curl -fk` (insecure) in healthcheck; cert is self-signed with our CA |

---

## Quality Gates - MANDATORY

**Per-Action Quality Gates**:
- ✅ All tests pass (`go test ./...`) — 100% passing, zero skips
- ✅ Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`) — zero errors
- ✅ Linting clean (`golangci-lint run` AND `--build-tags e2e,integration`) — zero warnings
- ✅ No new TODOs without tracking in tasks.md

**Per-Phase Quality Gates**:
- ✅ `go run ./cmd/cicd-lint lint-fitness` passes after every compose/template change
- ✅ `go run ./cmd/cicd-lint lint-deployments` passes after every compose/config change
- ✅ `go run ./cmd/cicd-lint lint-docs` passes after every doc change

---

## Success Criteria

- [ ] All phases complete with evidence
- [ ] `deployments/shared-telemetry/` compose starts with TLS (no plain HTTP for OTel/Grafana)
- [ ] All PS-ID services trust shared service CAs via `/certs` volume
- [ ] `postgres-url.secret` format updated (per Decision 2)
- [ ] `api/cryptosuite-registry/templates/` reflects all changes
- [ ] Documentation updated (tls-structure.md, target-structure.md, deployment-templates.md)
- [ ] All cicd-lint validators pass
- [ ] CI/CD workflows green

---

## ENG-HANDBOOK.md Cross-References

| Topic | Section |
|-------|---------|
| Service Framework Pattern | [Section 5.1](../../docs/ENG-HANDBOOK.md#51-service-framework-pattern) |
| Dual HTTPS Endpoint Pattern | [Section 5.3](../../docs/ENG-HANDBOOK.md#53-dual-https-endpoint-pattern) |
| TLS Certificate Configuration | [ENG-HANDBOOK.md Section 6](../../docs/ENG-HANDBOOK.md#6-security-architecture) |
| Secret Management | [Section 6 (MANDATORY)](../../docs/ENG-HANDBOOK.md#6-security-architecture) |
| Telemetry Strategy | [Section 9.4](../../docs/ENG-HANDBOOK.md#94-telemetry-strategy) |
| OTel Collector Constraints | [Section 9.4.1](../../docs/ENG-HANDBOOK.md#941-otel-collector-processor-constraints) |
| Testing Strategy | [Section 10](../../docs/ENG-HANDBOOK.md#10-testing-architecture) |
| Quality Gates | [Section 11.2](../../docs/ENG-HANDBOOK.md#112-quality-gates) |
| Deployment Architecture | [Section 12](../../docs/ENG-HANDBOOK.md#12-deployment-architecture) |
| Config File Architecture | [Section 13.2](../../docs/ENG-HANDBOOK.md#132-config-file-architecture) |
| Secrets Management in Deployments | [Section 13.3](../../docs/ENG-HANDBOOK.md#133-secrets-management-in-deployments) |
| Template Enforcement | [Section 13.6](../../docs/ENG-HANDBOOK.md#136-template-enforcement--drift-detection) |
| Infrastructure Blocker Escalation | [Section 14.7](../../docs/ENG-HANDBOOK.md#147-infrastructure-blocker-escalation) |
| Phase Post-Mortem | [Section 14.8](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation) |
