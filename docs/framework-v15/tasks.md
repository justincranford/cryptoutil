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

---

## Evidence Archive

- `test-output/v13-phase0/` — pki-init V13 cert generation verification
- `test-output/v13-phase3/` — OTel standalone mTLS verification
- `test-output/v13-phase6/` — OTel→Grafana pipeline verification
- `test-output/v13-phase10/` — Full deployment stack verification
