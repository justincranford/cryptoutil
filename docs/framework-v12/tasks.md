# Tasks - Framework V12: TLS Wiring

**Status**: 0 of 32 tasks complete (0%)
**Last Updated**: 2025-06-26
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

### Phase 1: PostgreSQL Server TLS [Status: ☐ TODO]

**Phase Objective**: Configure PostgreSQL leader and follower to serve TLS connections.

#### Task 1.1: PostgreSQL Server Cert Loading

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: V11 complete
- **Description**: Configure shared-postgres to load server certs from named Docker volumes.
- **Acceptance Criteria**:
  - [ ] Leader loads `public-postgres-leader-https-server-keystore/` cert+key
  - [ ] Follower loads `public-postgres-follower-https-server-keystore/` cert+key
  - [ ] `postgresql.conf`: `ssl = on`, `ssl_cert_file`, `ssl_key_file` configured

#### Task 1.2: PostgreSQL SSL Config

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 1.1
- **Description**: Update shared-postgres `postgresql.conf` template for SSL parameters.
- **Acceptance Criteria**:
  - [ ] `ssl = on`
  - [ ] `ssl_ca_file` points to server CA truststore `.crt`
  - [ ] `ssl_min_protocol_version = TLSv1.3`
  - [ ] Compose volumes mount cert directories

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
- **Description**: Configure GORM connection strings with client certs.
- **Acceptance Criteria**:
  - [ ] DSN includes `sslmode=verify-full&sslcert=...&sslkey=...&sslrootcert=...`
  - [ ] Each PS-ID uses its own client cert
  - [ ] postgres-{1,2} connect to leader; read replicas to follower

#### Task 2.3: Replication mTLS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 2.1
- **Description**: Configure leader↔follower replication with mTLS.
- **Acceptance Criteria**:
  - [ ] Follower replication uses `public-postgres-follower-https-client-keystore/` cert
  - [ ] Leader accepts replication with `public-postgres-leader-https-client-keystore/` cert

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
  - [ ] `receivers.otlp.protocols.grpc.tls.cert_file` and `key_file` configured
  - [ ] `receivers.otlp.protocols.http.tls.cert_file` and `key_file` configured
  - [ ] Client CA configured for client verification

#### Task 3.2: OTel Compose Volume Mounts

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 3.1
- **Description**: Mount cert volumes in otel-collector-contrib service.
- **Acceptance Criteria**:
  - [ ] Server keystore volume mounted
  - [ ] Client CA truststore volume mounted
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
  - [ ] Each PS-ID loads its OTel client cert
  - [ ] CA trust configured for OTel server cert verification
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

**Phase Objective**: Configure OTel-to-Grafana forwarding with mTLS.

#### Task 5.1: Grafana OTLP Ingest TLS

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Phase 3 complete
- **Description**: Configure grafana/otel-lgtm to accept mTLS on OTLP ingest ports.
- **Acceptance Criteria**:
  - [ ] Grafana OTLP ingest (:14317/:14318) accepts mTLS
  - [ ] Server cert loaded
  - [ ] Client CA configured

#### Task 5.2: OTel Exporter Client Cert

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 5.1
- **Description**: Configure OTel Collector exporter with client cert for Grafana.
- **Acceptance Criteria**:
  - [ ] `exporters.otlp.tls` section configured
  - [ ] Client cert from `public-grafana-otel-lgtm-*-https-client-keystore/`
  - [ ] Pipeline verified

---

### Phase 6: Grafana LGTM HTTPS UI [Status: ☐ TODO]

**Phase Objective**: Serve Grafana UI over HTTPS.

#### Task 6.1: Grafana HTTPS Config

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: V11 complete
- **Description**: Configure grafana/otel-lgtm for HTTPS serving on port 3000.
- **Acceptance Criteria**:
  - [ ] Server cert loaded from `public-grafana-otel-lgtm-https-server-keystore/`
  - [ ] HTTPS enabled on :3000
  - [ ] Admin client cert for access

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
  - [ ] Each instance loads from `public-{PS-ID}-{variant}-https-server-keystore/`
  - [ ] ServerSettings updated with cert paths

#### Task 7.2: Private Admin mTLS Cert Loading

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 7.1
- **Description**: Configure apps to load private admin mTLS certs.
- **Acceptance Criteria**:
  - [ ] Each instance loads from `private-{PS-ID}-mutual-https-client-server-{variant}-keystore/`
  - [ ] Admin endpoint verifies client certs

#### Task 7.3: Client Cert Loading

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 7.1
- **Description**: Configure apps to accept client certs per API path and realm.
- **Acceptance Criteria**:
  - [ ] Browser path uses browseruser realm certs
  - [ ] Service path uses serviceuser realm certs
  - [ ] Client CAs loaded from truststore directories

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
