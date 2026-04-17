# pki-init Phase Ordering Principles

**Purpose**: Document why TLS wiring phases are ordered the way they are in V12 and V13.
**Created**: 2026-04-16

---

## Ordering Principles

### 1. Server TLS Before Client mTLS

A service must serve TLS before any client can verify its cert. Adding client cert requirements to a
service that does not yet have server TLS will cause connection failures before the cert exchange
even begins. The staged approach:

```
Server TLS → Verify Server TLS → Client mTLS → Verify Full mTLS
```

### 2. Verify After Each Server TLS Stage

Before adding client cert requirements (which change `pg_hba.conf`, `ssl_ca_file`, etc.), verify
the server-only TLS works. This isolates failures: if the verify step fails, the root cause is
server TLS config, not client cert wiring.

### 3. Infrastructure Before Application

Infrastructure services (PostgreSQL, OTel Collector, Grafana LGTM) run their TLS first. Application
instances connect TO infrastructure — they cannot be configured for mTLS until the infrastructure
endpoint exists and serves TLS.

**Order within infrastructure**:
- Persistent storage (PostgreSQL) before ephemeral pipeline (OTel/Grafana)
- Rationale: a broken PG connection causes data loss; a broken telemetry pipeline is observable but safe

### 4. Replication/Internal Channels Before External Connections

PG leader→follower replication uses a server-only TLS channel in Phase 2 of V12. The full
replication mTLS (with follower client cert) is added in Phase 5, AFTER app client mTLS is
verified in Phase 6. This prevents replication disruption from blocking app connectivity work.

For OTel/Grafana: the intra-infra channel (OTel→Grafana) is handled separately from the app→OTel
channel. OTel server TLS is established first (Phase 1 of V13), then app clients connect (Phase 2),
then the OTel→Grafana forwarding channel gets client cert (Phase 5).

### 5. Config Before Deployment Templates

Implementation config correctness is validated in a running Docker Compose stack BEFORE writing
canonical deployment templates. This prevents encoding a broken config into the canonical templates
that all 10 services copy from.

### 6. Lint Before Deployment Verification

Run `lint-deployments` after each template update, before starting Docker Compose. This catches
structural errors (port conflicts, missing secrets, admin bind violations) without needing a full
cluster start. Cost is low; value is catching errors early.

### 7. Knowledge Propagation Last

After all implementation phases complete and quality gates pass, capture lessons in `lessons.md`
and propagate findings to ENG-HANDBOOK.md, agents, skills, and instructions. Doing this last
ensures lessons reflect the full implementation experience, not speculation.

---

## V12 Phase Ordering Rationale

**V12 Scope**: PostgreSQL mTLS (leader/follower) + Private PS-ID App Admin mTLS Trust

```
Phase 0: pki-init Patch (Cat 9 infra + Cat 14 postgres-only)
Phase 1: PG Server TLS — Leader + Follower
Phase 2: PG Replication Server TLS
Phase 3: Verify PG Standalone
Phase 4: PG Client mTLS — HBA + GORM
Phase 5: PG Replication Client mTLS
Phase 6: Verify PG Full Stack
Phase 7: Deployment Templates for PG TLS
Phase 8: Deployment Linting
Phase 9: Deployment Verification — PG TLS
Phase 10: Private App Admin mTLS Trust
Phase 11: Knowledge Propagation
```

**Why Phase 0 generates Cat 9 infra + Cat 14**:
- Cat 14 (postgres-only client certs) is a new entity type that pki-init doesn't yet generate
- Cat 9 infra (OTel→Grafana client cert) is a V13 prerequisite generated in V12 Phase 0 so that
  V13 can start immediately after V12 without another pki-init patch cycle blocking progress

**Why leader server TLS (Phase 1) before replication server TLS (Phase 2)**:
- Leader is the write target; all app connections go to leader first
- Replication is non-blocking (follower reconnects automatically when leader TLS is enabled)
- Leader TLS independently verifiable before touching replication config

**Why standalone verify (Phase 3) before client mTLS (Phase 4)**:
- `pg_hba.conf clientcert=verify-full` makes connections require a client cert immediately
- If server TLS is broken and we add clientcert simultaneously, impossible to isolate which failed
- Phase 3 confirms server TLS works; Phase 4 can then be the only variable introduced

**Why replication client mTLS (Phase 5) after app client mTLS (Phase 4)**:
- Replication disruption is more disruptive than app connectivity issues
- App client mTLS verified working in Phase 6 before replication mTLS adds complexity

**Why Phase 10 (admin mTLS) after PG full stack verified (Phase 6 + Phases 7-9)**:
- Admin mTLS is orthogonal to PostgreSQL TLS; they don't interact
- Placing admin mTLS after PG deployment verification keeps PG scope clean
- If PG deployment has issues (Phases 7-9), admin mTLS work proceeds in parallel conceptually
  but is sequenced after to avoid interleaved failures

---

## V13 Phase Ordering Rationale

**V13 Scope**: OTel Collector mTLS + Grafana LGTM HTTPS/mTLS + Public PS-ID App TLS Trust

```
Phase 0: pki-init Patch (Cat 2, Cat 3, Cat 4, Cat 8, Cat 9 app)
Phase 1: OTel Collector Server TLS (OTLP receiver)
Phase 2: App→OTel Client mTLS (OTLP exporter)
Phase 3: Verify OTel Standalone
Phase 4: Grafana LGTM HTTPS UI + OTLP Ingest TLS
Phase 5: OTel→Grafana Client mTLS
Phase 6: Verify OTel→Grafana Pipeline
Phase 7: Public PS-ID App Server TLS
Phase 8: Deployment Templates for OTel/Grafana/App TLS
Phase 9: Deployment Linting
Phase 10: Deployment Verification — Full Telemetry Stack
Phase 11: Knowledge Propagation
```

**Why OTel server TLS (Phase 1) before app→OTel client mTLS (Phase 2)**:
- Apps cannot present client certs to an OTel receiver that doesn't yet serve TLS
- Same principle as V12: server first, then verify, then client

**Why verify OTel standalone (Phase 3) before Grafana (Phase 4)**:
- OTel→Grafana forwarding depends on OTel server TLS being stable
- Grafana config introduces new variables (grafana.ini, D1 decision); isolate from OTel issues

**Why Grafana HTTPS UI (Phase 4) includes both HTTPS UI AND OTLP ingest TLS**:
- Both Grafana features use the same Cat 2 server cert (`public-https-server-entity-grafana-otel-lgtm`)
- Combining into one phase reduces the number of Grafana container restarts and cert mount changes
- D1 (grafana.ini approach) and D6 (mTLS assumed) are both Grafana-specific decisions

**Why OTel→Grafana client mTLS (Phase 5) after Grafana HTTPS is verified (Phase 4)**:
- Same staged pattern: Grafana serves TLS (Phase 4), OTel presents client cert (Phase 5)
- If Phase 4 is broken, Phase 5 failures would be misattributed

**Why Phase 7 (Public App Server TLS) is separate from OTel/Grafana phases**:
- App public server TLS (Cat 3) is independent of the telemetry pipeline
- Certs don't interact: Cat 3 is served by apps; Cat 2 is served by OTel/Grafana
- Phase 7 is placed after telemetry pipeline is verified to keep concerns cleanly separated

**Why Phase 8 (Deployment Templates) precedes Phase 9 (Linting) precedes Phase 10 (Verification)**:
- Same principle as V12: lint catches template errors without Docker overhead
- Templates encode the "correct" config; linting validates structure; deployment verifies runtime

---

## Cross-Plan Dependency

V13 Phase 0 (pki-init patch) generates Cat 2, Cat 3, Cat 4, Cat 8, and Cat 9 app entities.
V12 Phase 0 generates Cat 9 infra and Cat 14 entities.

**Execution order**: V12 → V13 (V13's Phase 0 can begin in parallel with V12 Phases 1-9, but
V13 Phases 1+ require V12 Phase 0 complete for the Cat 9 infra cert used in OTel→Grafana).

**V13 Phase 0 is independent of V12 Phases 1-9**: The pki-init generator changes for V13 cert
categories do not depend on the PostgreSQL TLS wiring in V12 Phases 1-9. V13 Phase 0 may be
executed concurrently with V12 Phases 1-9 when two parallel workstreams are available.
