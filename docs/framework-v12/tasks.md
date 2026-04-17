# Tasks - Framework V12: TLS Wiring

**Status**: 0 of 36 tasks complete (0%)
**Last Updated**: 2025-07-07
**Created**: 2025-06-26

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

### Phase 0: pki-init Patch — Cat 9 infra + Cat 14 postgres-only [Status: ☐ TODO]

**Phase Objective**: Apply D3 and D4 structural changes to pki-init generator before TLS wiring.

#### Task 0.1: Add PKIInitEntityInfra Magic Constant

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Add `PKIInitEntityInfra = "infra"` to `internal/shared/magic/magic_pkiinit.go`.
- **Acceptance Criteria**:
  - [ ] `PKIInitEntityInfra = "infra"` added alongside other entity constants
  - [ ] Godoc comment added
  - [ ] `go build ./...` clean
  - [ ] `golangci-lint run` clean
- **Files**: `internal/shared/magic/magic_pkiinit.go`

#### Task 0.2: Add PKIInitPostgresAppInstanceSuffixes Function

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 0.1
- **Description**: Add `PKIInitPostgresAppInstanceSuffixes()` to `tier.go` returning `["postgres-1", "postgres-2"]` only.
- **Acceptance Criteria**:
  - [ ] Function returns `[PKIInitInstanceSuffixPostgres1, PKIInitInstanceSuffixPostgres2]`
  - [ ] Godoc comment explains postgres-only rationale (sqlite instances don’t connect to PostgreSQL)
  - [ ] `go build ./...` clean
- **Files**: `internal/apps/framework/tls/tier.go`

#### Task 0.3: Cat 9 infra Cert Generation

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 0.1
- **Description**: Add `infra` entity type to Cat 9 in `generateSharedCAs()`.
- **Acceptance Criteria**:
  - [ ] `grafana-otel-lgtm-https-client-entity-infra/` generated after `admin` block
  - [ ] `otel-collector-contrib-https-client-entity-infra/` generated after `admin` block
  - [ ] Both use `PKIInitEntityInfra` constant (not bare string `"infra"`)
  - [ ] Generator function comment updated: `9 (admin+infra)`
  - [ ] `go build ./...` clean
- **Files**: `internal/apps/framework/tls/generator.go`

#### Task 0.4: Cat 14 postgres-only Loop

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 0.2
- **Description**: Change Cat 14 generation to use `PKIInitPostgresAppInstanceSuffixes()` instead of `PKIInitAppInstanceSuffixes()`.
- **Acceptance Criteria**:
  - [ ] Cat 14 loop uses `PKIInitPostgresAppInstanceSuffixes()`
  - [ ] Cat 14 comment updated from "8 dirs" to "4 dirs"
  - [ ] `go build ./...` clean
- **Files**: `internal/apps/framework/tls/generator.go`

#### Task 0.5: Update Generator Tests

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Tasks 0.3, 0.4
- **Description**: Update generator unit tests for new directory structure.
- **Acceptance Criteria**:
  - [ ] Expected total dir count updated (recalculate: 28 global + new PS-ID count per test tier)
  - [ ] Any Cat 9 `entity-admin` dir name assertions updated to also include `entity-infra`
  - [ ] Any Cat 14 sqlite dir assertions removed
  - [ ] `go test ./internal/apps/framework/tls/... -v -run TestGenerate` passes
  - [ ] `go test ./internal/apps/framework/tls/...` 100% pass (no failures)
- **Files**: `internal/apps/framework/tls/generator_test.go`

---

### Phase 1: PostgreSQL Server TLS [Status: ☐ TODO]

**Phase Objective**: Configure PostgreSQL leader and follower to serve TLS connections.

#### Task 1.1: PostgreSQL Server Cert Loading

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: V11 complete
- **Description**: Configure shared-postgres to load server certs from named Docker volumes.
- **Acceptance Criteria**:
  - [ ] Leader loads `postgres-tls-server-entity-leader/` cert+key (Cat 11 keystore)
  - [ ] Follower loads `postgres-tls-server-entity-follower/` cert+key (Cat 11 keystore)
  - [ ] `postgresql.conf`: `ssl = on`, `ssl_cert_file`, `ssl_key_file` configured
  - [ ] `ssl_ca_file` points to `postgres-tls-client-issuing-ca/truststore/` (Cat 12)
  - [ ] Compose volume: `__PS_ID__-certs` named volume defined in PS-ID compose (include-merged, D5)

#### Task 1.2: PostgreSQL SSL Config

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 1.1
- **Description**: Update shared-postgres `postgresql.conf` template for SSL parameters.
- **Acceptance Criteria**:
  - [ ] `ssl = on`
  - [ ] `ssl_cert_file` = `postgres-tls-server-entity-{leader,follower}/SAME-AS-DIR-NAME.crt`
  - [ ] `ssl_key_file` = `postgres-tls-server-entity-{leader,follower}/SAME-AS-DIR-NAME.key`
  - [ ] `ssl_ca_file` = `postgres-tls-client-issuing-ca/truststore/postgres-tls-client-issuing-ca.crt` (Cat 12)
  - [ ] `ssl_min_protocol_version = TLSv1.3`
  - [ ] Compose volumes mount `__PS_ID__-certs` named volume (include-merged approach, D5)

#### Task 1.3: PostgreSQL Init Script Updates

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 1.2
- **Description**: Update shared-postgres init scripts for TLS-aware startup.
- **Acceptance Criteria**:
  - [ ] Init scripts set correct file permissions on cert files
  - [ ] SSL health check works after startup

#### Task 1.4: Verify Server TLS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 1.3
- **Description**: Verify PostgreSQL accepts TLS connections.
- **Acceptance Criteria**:
  - [ ] `psql "sslmode=verify-full"` connects successfully
  - [ ] TLS 1.3 negotiated

---

### Phase 2: PostgreSQL Client mTLS [Status: ☐ TODO]

**Phase Objective**: Configure mTLS for app-to-PostgreSQL and replication connections.

#### Task 2.1: pg_hba.conf mTLS Rules

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Phase 1 complete
- **Description**: Update `pg_hba.conf` template for client certificate verification.
- **Acceptance Criteria**:
  - [ ] `hostssl` rules with `clientcert=verify-full`
  - [ ] Non-SSL connections rejected
  - [ ] CN mapping configured

#### Task 2.2: App Instance Client Certs

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 2.1
- **Description**: Configure GORM connection with client certs via YAML config fields (D2: YAML cert paths).
- **Acceptance Criteria**:
  - [ ] Framework config struct adds `database.sslmode`, `database.sslcert`, `database.sslkey`, `database.sslrootcert` fields
  - [ ] `sslcert` = `postgres-tls-client-entity-leader-{PS-ID}-postgres-{1,2}/SAME-AS-DIR-NAME.crt` (Cat 14)
  - [ ] `sslkey` = `postgres-tls-client-entity-leader-{PS-ID}-postgres-{1,2}/SAME-AS-DIR-NAME.key` (Cat 14)
  - [ ] `sslrootcert` = `postgres-tls-server-issuing-ca/truststore/postgres-tls-server-issuing-ca.crt` (Cat 10)
  - [ ] Only postgres-1 and postgres-2 instances use PostgreSQL client certs (sqlite-1/sqlite-2 do not)
  - [ ] Config YAML files updated per PS-ID per postgres variant

#### Task 2.3: Replication mTLS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 2.1
- **Description**: Configure leader↔follower replication with mTLS.
- **Acceptance Criteria**:
  - [ ] Follower uses `postgres-tls-client-entity-follower-replication/` cert (Cat 13 keystore)
  - [ ] Leader accepts replication from `postgres-tls-client-entity-leader-replication/` cert (Cat 13 keystore)
  - [ ] `primary_conninfo` in follower uses `sslcert`, `sslkey`, `sslrootcert` parameters

#### Task 2.4: Verify Client mTLS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Tasks 2.2, 2.3
- **Description**: Verify mTLS connections and rejection of non-mTLS.
- **Acceptance Criteria**:
  - [ ] App instances connect via mTLS
  - [ ] Non-client-cert connections rejected
  - [ ] Replication streams verified

---

### Phase 3: OTel Collector Server TLS [Status: ☐ TODO]

**Phase Objective**: Configure OTel Collector to serve mTLS on OTLP endpoints.

#### Task 3.1: OTel Server Cert Config

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: V11 complete
- **Description**: Update otel-collector-contrib YAML config for TLS receiver.
- **Acceptance Criteria**:
  - [ ] `receivers.otlp.protocols.grpc.tls.cert_file` = `public-https-server-entity-otel-collector-contrib/SAME-AS-DIR-NAME.crt` (Cat 2)
  - [ ] `receivers.otlp.protocols.grpc.tls.key_file` = `public-https-server-entity-otel-collector-contrib/SAME-AS-DIR-NAME.key` (Cat 2)
  - [ ] `receivers.otlp.protocols.grpc.tls.client_ca_file` = `otel-collector-contrib-https-client-issuing-ca/truststore/otel-collector-contrib-https-client-issuing-ca.crt` (Cat 8)
  - [ ] HTTP OTLP receiver configured identically
  - [ ] `insecure: false` on all OTLP receiver protocols

#### Task 3.2: OTel Compose Volume Mounts

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 3.1
- **Description**: Mount cert volumes in otel-collector-contrib service.
- **Acceptance Criteria**:
  - [ ] `public-https-server-entity-otel-collector-contrib/` (Cat 2) mounted for server cert (keystore)
  - [ ] `otel-collector-contrib-https-client-issuing-ca/truststore/` (Cat 8) mounted for client CA
  - [ ] Volume: `__PS_ID__-certs` named volume applied per D5 (include-merged)
  - [ ] Compose file validates

#### Task 3.3: Verify OTel Server TLS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 3.2
- **Description**: Verify OTel accepts mTLS and rejects insecure connections.
- **Acceptance Criteria**:
  - [ ] mTLS connection succeeds on :4317 and :4318
  - [ ] Insecure connection rejected

---

### Phase 4: App → OTel Client mTLS [Status: ☐ TODO]

**Phase Objective**: Configure app OTLP exporters to use client certs.

#### Task 4.1: Go OTLP Exporter TLS Config

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Phase 3 complete
- **Description**: Update Go OTLP exporter configuration with client cert and CA.
- **Acceptance Criteria**:
  - [ ] Each PS-ID app instance loads `otel-collector-contrib-https-client-entity-{PS-ID}-{sqlite,postgres}-{1,2}/` (Cat 9 keystore)
  - [ ] Server CA trust from `public-https-server-issuing-ca/truststore/` (Cat 1) for OTel server cert verification
  - [ ] TLS 1.3 minimum

#### Task 4.2: Compose Config Updates

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 4.1
- **Description**: Update compose configs for OTLP endpoint TLS.
- **Acceptance Criteria**:
  - [ ] `otlp.endpoint` uses `https://` scheme
  - [ ] `otlp.insecure: false`
  - [ ] Cert paths in service config

#### Task 4.3: Verify App→OTel mTLS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 4.2
- **Description**: Verify telemetry flows via mTLS.
- **Acceptance Criteria**:
  - [ ] Traces/metrics/logs appear in OTel
  - [ ] Non-mTLS export rejected

---

### Phase 5: OTel → Grafana Client mTLS [Status: ☐ TODO]

**Phase Objective**: Configure OTel-to-Grafana forwarding with full mTLS (D6=mTLS assumed; D3=`infra` entity cert).

#### Task 5.1: Grafana OTLP Ingest mTLS

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Phase 3 complete
- **Description**: Configure grafana/otel-lgtm to accept mTLS on OTLP ingest ports (D6=mTLS assumed; D3=infra entity).
- **Acceptance Criteria**:
  - [ ] Grafana OTLP ingest (:14317/:14318) accepts mTLS connections
  - [ ] Client CA from `grafana-otel-lgtm-https-client-issuing-ca/truststore/` (Cat 8) configured
  - [ ] Server cert from `public-https-server-entity-grafana-otel-lgtm/` (Cat 2) used
  - [ ] Volume: `__PS_ID__-certs` named volume applied per D5 (include-merged)

#### Task 5.2: OTel Exporter Client Cert

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 5.1
- **Description**: Configure OTel Collector exporter with `infra` client cert for Grafana (D3=E: new `infra` entity).
- **Acceptance Criteria**:
  - [ ] `exporters.otlp.tls.ca_file` = `public-https-server-issuing-ca/truststore/` (Cat 1) for Grafana server cert
  - [ ] `exporters.otlp.tls.cert_file` = `grafana-otel-lgtm-https-client-entity-infra/SAME-AS-DIR-NAME.crt` (Cat 9 infra)
  - [ ] `exporters.otlp.tls.key_file` = `grafana-otel-lgtm-https-client-entity-infra/SAME-AS-DIR-NAME.key` (Cat 9 infra)
  - [ ] Pipeline verified: telemetry flows to Grafana

---

### Phase 6: Grafana LGTM HTTPS UI [Status: ☐ TODO]

**Phase Objective**: Serve Grafana UI over HTTPS.

#### Task 6.1: Grafana HTTPS Config

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: V11 complete
- **Description**: Configure grafana/otel-lgtm for HTTPS via custom grafana.ini (D1: grafana.ini approach).
- **Acceptance Criteria**:
  - [ ] `shared-telemetry/grafana-otel-lgtm/grafana.ini` created with `[server]` section
  - [ ] `protocol = https`
  - [ ] `cert_file` = `/certs/__PS_ID__/public-https-server-entity-grafana-otel-lgtm/public-https-server-entity-grafana-otel-lgtm.crt` (Cat 2)
  - [ ] `cert_key` = `/certs/__PS_ID__/public-https-server-entity-grafana-otel-lgtm/public-https-server-entity-grafana-otel-lgtm.key` (Cat 2)
  - [ ] `grafana.ini` mounted as volume at `/etc/grafana/grafana.ini:ro` in compose
  - [ ] Volume: `__PS_ID__-certs` named volume applied per D5 (include-merged)
  - [ ] `healthcheck` updated to `https://127.0.0.1:3000/api/health`

#### Task 6.2: Verify Grafana HTTPS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 6.1
- **Description**: Verify Grafana UI is accessible via HTTPS.
- **Acceptance Criteria**:
  - [ ] `curl --cacert ... https://localhost:3000` succeeds
  - [ ] Dashboards accessible

---

### Phase 7: PS-ID App TLS Trust [Status: ☐ TODO]

**Phase Objective**: Configure all app instances to load certs and trust CAs.

#### Task 7.1: Public Server Cert Loading

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: V11 complete
- **Description**: Configure apps to load public HTTPS server certs.
- **Acceptance Criteria**:
  - [ ] Each instance loads `public-https-server-entity-{PS-ID}-{sqlite,postgres}-{1,2}/` (Cat 3 keystore)
  - [ ] ServerSettings updated with cert file path and key file path

#### Task 7.2: Private Admin mTLS Cert Loading

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 7.1
- **Description**: Configure apps to load private admin mTLS certs.
- **Acceptance Criteria**:
  - [ ] Each instance loads `private-https-mutual-entity-{PS-ID}-{sqlite,postgres}-{1,2}/` (Cat 7 keystore) — this cert has both Client Auth and Server Auth EKU
  - [ ] Admin CA truststore from `private-https-mutual-issuing-ca-{PS-ID}-{sqlite,postgres}-{1,2}/truststore/` (Cat 6) for verifying admin clients

#### Task 7.3: Client Cert Loading

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 7.1
- **Description**: Configure apps to accept client certs per API path and realm.
- **Acceptance Criteria**:
  - [ ] Browser path (`/browser/`) uses `public-https-client-issuing-ca-{PS-ID}-{pki-domain}/truststore/` (Cat 4) where pki-domain=`sqlite-1`,`sqlite-2`, or `postgres`
  - [ ] Service path (`/service/`) uses same Cat 4 truststore per PKI domain
  - [ ] Global server CA truststore `public-https-server-issuing-ca/truststore/` (Cat 1) loaded for outbound TLS verification

---

### Phase 8: E2E Testing [Status: ☐ TODO]

**Phase Objective**: End-to-end TLS verification.

#### Task 8.1: Full Stack Docker Compose

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Phases 1-7 complete
- **Description**: Docker Compose up with all TLS enabled.
- **Acceptance Criteria**:
  - [ ] All services start with TLS
  - [ ] Health checks pass

#### Task 8.2: mTLS Connection Matrix

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 8.1
- **Description**: Verify every mTLS connection pair.
- **Acceptance Criteria**:
  - [ ] App ↔ PostgreSQL mTLS verified
  - [ ] App → OTel mTLS verified
  - [ ] OTel → Grafana mTLS verified
  - [ ] Grafana HTTPS UI verified

#### Task 8.3: Negative Tests

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 8.2
- **Description**: Verify non-TLS and wrong-cert connections are rejected.
- **Acceptance Criteria**:
  - [ ] Non-TLS connections rejected by all services
  - [ ] Wrong client cert rejected
  - [ ] Expired cert rejected

#### Task 8.4: Performance Baseline

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 8.2
- **Description**: Measure TLS overhead.
- **Acceptance Criteria**:
  - [ ] Benchmark before/after TLS
  - [ ] Acceptable overhead (<5% latency increase)

---

### Phase 9: Knowledge Propagation [Status: ☐ TODO]

**Phase Objective**: Apply lessons learned to permanent artifacts.

#### Task 9.1: Review Lessons

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phases 1-8 complete
- **Description**: Review lessons.md.
- **Acceptance Criteria**:
  - [ ] All lessons reviewed
  - [ ] Actionable items identified

#### Task 9.2: Update Artifacts

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 9.1
- **Description**: Update ENG-HANDBOOK.md and other artifacts.
- **Acceptance Criteria**:
  - [ ] TLS wiring patterns documented
  - [ ] Propagation check passes

#### Task 9.3: Final Commit

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 9.2
- **Acceptance Criteria**:
  - [ ] Clean working tree
  - [ ] All quality gates pass

---

## Cross-Cutting Tasks

### Testing

- [ ] E2E tests pass with full TLS stack
- [ ] mTLS connection matrix verified
- [ ] Negative tests pass (rejection of non-TLS)
- [ ] Performance baseline measured

### Code Quality

- [ ] Linting passes
- [ ] No new TODOs without tracking

### Documentation

- [ ] deployment-templates.md updated with TLS config examples
- [ ] tls-structure.md cross-referenced
- [ ] ENG-HANDBOOK.md updated with TLS wiring patterns

---

## Evidence Archive

- `test-output/v12-phase1/` — PostgreSQL TLS verification
- `test-output/v12-phase8/` — E2E TLS connection matrix
