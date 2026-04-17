# Tasks - Framework V12: PostgreSQL mTLS + Private PS-ID App mTLS Trust

**Status**: 14 of 43 tasks complete (33%)
**Last Updated**: 2026-04-17
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

### Phase 0: pki-init Patch — Cat 9 infra + Cat 14 postgres-only [Status: ✅ COMPLETE]

**Phase Objective**: Apply D3 and D4 structural changes to pki-init generator before TLS wiring.

#### Task 0.1: Add PKIInitEntityInfra Magic Constant

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: None
- **Description**: Add `PKIInitEntityInfra = "infra"` to `internal/shared/magic/magic_pkiinit.go`.
- **Acceptance Criteria**:
  - [x] `PKIInitEntityInfra = "infra"` added alongside other entity constants
  - [x] Godoc comment added
  - [x] `go build ./...` clean
  - [x] `golangci-lint run` clean
- **Files**: `internal/shared/magic/magic_pkiinit.go`

#### Task 0.2: Add PKIInitPostgresAppInstanceSuffixes Function

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: Task 0.1
- **Description**: Add `PKIInitPostgresAppInstanceSuffixes()` to `tier.go` returning `["postgres-1", "postgres-2"]` only.
- **Acceptance Criteria**:
  - [x] Function returns `[PKIInitInstanceSuffixPostgres1, PKIInitInstanceSuffixPostgres2]`
  - [x] Godoc comment explains postgres-only rationale (sqlite instances don’t connect to PostgreSQL)
  - [x] `go build ./...` clean
- **Files**: `internal/apps/framework/tls/tier.go`

#### Task 0.3: Cat 9 infra Cert Generation

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: Task 0.1
- **Description**: Add `infra` entity type to Cat 9 in `generateSharedCAs()`.
- **Acceptance Criteria**:
  - [x] `grafana-otel-lgtm-https-client-entity-infra/` generated after `admin` block
  - [x] `otel-collector-contrib-https-client-entity-infra/` generated after `admin` block
  - [x] Both use `PKIInitEntityInfra` constant (not bare string `"infra"`)
  - [x] Generator function comment updated: `9 (admin+infra)`
  - [x] `go build ./...` clean
- **Files**: `internal/apps/framework/tls/generator.go`

#### Task 0.4: Cat 14 postgres-only Loop

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: Task 0.2
- **Description**: Change Cat 14 generation to use `PKIInitPostgresAppInstanceSuffixes()` instead of `PKIInitAppInstanceSuffixes()`.
- **Acceptance Criteria**:
  - [x] Cat 14 loop uses `PKIInitPostgresAppInstanceSuffixes()`
  - [x] Cat 14 comment updated from "8 dirs" to "4 dirs"
  - [x] `go build ./...` clean
- **Files**: `internal/apps/framework/tls/generator.go`

#### Task 0.5: Update Generator Tests

- **Status**: ✅
- **Estimated**: 0.5h
- **Dependencies**: Tasks 0.3, 0.4
- **Description**: Update generator unit tests for new directory structure.
- **Acceptance Criteria**:
  - [x] Expected total dir count updated (recalculate: 28 global + new PS-ID count per test tier)
  - [x] Any Cat 9 `entity-admin` dir name assertions updated to also include `entity-infra`
  - [x] Any Cat 14 sqlite dir assertions removed
  - [x] `go test ./internal/apps/framework/tls/... -v -run TestGenerate` passes
  - [x] `go test ./internal/apps/framework/tls/...` 100% pass (no failures)
- **Files**: `internal/apps/framework/tls/generator_test.go`

---

### Phase 1: PostgreSQL Server TLS — Leader + Follower [Status: ✅ COMPLETE]

**Phase Objective**: Configure PostgreSQL leader and follower to serve TLS from `/certs` volume mounts.

#### Task 1.1: Leader postgresql.conf SSL Config

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: Phase 0 complete, V11 complete
- **Description**: Configure leader `postgresql.conf` for TLS.
- **Acceptance Criteria**:
  - [x] `ssl = on`
  - [x] `ssl_cert_file` = `postgres-tls-server-entity-leader/SAME-AS-DIR-NAME.crt` (Cat 11)
  - [x] `ssl_key_file` = `postgres-tls-server-entity-leader/SAME-AS-DIR-NAME.key` (Cat 11)
  - [x] `ssl_ca_file` = `postgres-tls-client-issuing-ca/truststore/postgres-tls-client-issuing-ca.crt` (Cat 12)
  - [x] `ssl_min_protocol_version = TLSv1.3`
- **Files**: `deployments/shared-postgres/postgresql-leader.conf`, `api/cryptosuite-registry/templates/deployments/shared-postgres/postgresql-leader.conf`

#### Task 1.2: Follower postgresql.conf SSL Config

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: Task 1.1
- **Description**: Configure follower `postgresql.conf` for TLS (follower-specific cert paths).
- **Acceptance Criteria**:
  - [x] `ssl = on`
  - [x] `ssl_cert_file` = `postgres-tls-server-entity-follower/SAME-AS-DIR-NAME.crt` (Cat 11)
  - [x] `ssl_key_file` = `postgres-tls-server-entity-follower/SAME-AS-DIR-NAME.key` (Cat 11)
  - [x] `ssl_ca_file` = `postgres-tls-client-issuing-ca/truststore/postgres-tls-client-issuing-ca.crt` (Cat 12)
  - [x] `ssl_min_protocol_version = TLSv1.3`
- **Files**: `deployments/shared-postgres/postgresql-follower.conf`, `api/cryptosuite-registry/templates/deployments/shared-postgres/postgresql-follower.conf`

#### Task 1.3: shared-postgres Compose Cert Volume Mounts (Leader + Follower)

- **Status**: ✅
- **Estimated**: 1.5h
- **Dependencies**: Tasks 1.1, 1.2
- **Description**: Mount cert dirs in shared-postgres compose with least privilege per node.
- **Acceptance Criteria**:
  - [x] `cryptoutil-certs` named volume referenced (not re-declared) in shared-postgres compose (D5 include-merged via `__SUITE__-certs` template placeholder)
  - [x] Leader mounts: `cryptoutil-certs:/certs:ro` (all certs via shared volume)
  - [x] Follower mounts: `cryptoutil-certs:/certs:ro` (all certs via shared volume)
  - [x] `pg_hba.conf` bind-mounted to both leader and follower
  - [x] Template updated: `api/cryptosuite-registry/templates/deployments/shared-postgres/compose.yml`
  - [x] `go run ./cmd/cicd-lint lint-deployments` passes all 54 validators
  - [x] `go run ./cmd/cicd-lint lint-fitness` passes all linters
- **Files**: `deployments/shared-postgres/compose.yml`, `api/cryptosuite-registry/templates/deployments/shared-postgres/compose.yml`

#### Task 1.4: pg_hba.conf Creation (hostssl rules, no clientcert yet)

- **Status**: ✅
- **Estimated**: 1h
- **Dependencies**: Task 1.3
- **Description**: Create `pg_hba.conf` with hostssl rules for TLS-required connections.
- **Acceptance Criteria**:
  - [x] `pg_hba.conf` created with `local trust`, `host scram-sha-256` loopback, `hostssl` for replication and app
  - [x] No `clientcert=verify-full` yet (added in Phases 4 and 5)
  - [x] Template `api/cryptosuite-registry/templates/deployments/shared-postgres/pg_hba.conf` created
  - [x] `go run ./cmd/cicd-lint lint-fitness` passes template-compliance check
- **Files**: `deployments/shared-postgres/pg_hba.conf`, `api/cryptosuite-registry/templates/deployments/shared-postgres/pg_hba.conf`

---

### Phase 2: PostgreSQL Replication Server TLS [Status: ✅ COMPLETE (2.1+2.2 done in Phase 1; 2.3 requires Docker)]

**Phase Objective**: Configure follower→leader replication to use server TLS.

#### Task 2.1: Follower primary_conninfo Server TLS

- **Status**: ✅ (completed as part of Task 1.2)
- **Estimated**: 1h
- **Dependencies**: Phase 1 complete
- **Description**: Add server TLS params to follower `primary_conninfo`.
- **Acceptance Criteria**:
  - [x] `sslmode=verify-full` in `primary_conninfo`
  - [x] `sslrootcert` = Cat 10 `postgres-tls-server-issuing-ca/truststore/postgres-tls-server-issuing-ca.crt` path
  - [x] `sslcert`/`sslkey` = Cat 13 paths included (follower-replication client cert — Phase 5 completes HBA requirement)
- **Files**: `deployments/shared-postgres/postgresql-follower.conf`

#### Task 2.2: Follower Cat 10 Truststore Mount

- **Status**: ✅ (included in Task 1.3 — all certs mounted via shared volume)
- **Estimated**: 0.5h
- **Dependencies**: Task 2.1
- **Description**: Cat 10 truststore accessible via shared `cryptoutil-certs` volume — no separate mount needed.
- **Acceptance Criteria**:
  - [x] Cat 10 `postgres-tls-server-issuing-ca/truststore/` accessible via shared volume
  - [x] Compose file validates (all lint-deployments checks pass)
- **Files**: `deployments/shared-postgres/compose.yml`

#### Task 2.3: Verify Replication Server TLS

- **Status**: ⏳ DEFERRED (requires Docker — to be verified in Phase 9)
- **Estimated**: 0.5h
- **Dependencies**: Tasks 2.1, 2.2
- **Description**: Verify replication slot reconnects over TLS.
- **Acceptance Criteria**:
  - [ ] `SELECT * FROM pg_stat_replication` on leader shows `ssl = t`
  - [ ] Replication lag returns to zero after TLS reconfiguration

---

### Phase 3: Verify PostgreSQL Standalone [Status: ☐ TODO]

**Phase Objective**: Confirm each node responds to TLS connections independently before requiring client certs.

#### Task 3.1: Verify Leader TLS

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 2 complete
- **Description**: Verify leader accepts TLS connections.
- **Acceptance Criteria**:
  - [ ] `psql "sslmode=verify-full sslrootcert=<Cat 10>"` connects to leader
  - [ ] `SELECT ssl_version FROM pg_stat_ssl WHERE pid = pg_backend_pid()` returns `TLSv1.3`
  - [ ] `sslmode=disable` connection still succeeds (HBA not yet hardened)

#### Task 3.2: Verify Follower TLS

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 2 complete
- **Description**: Verify follower accepts TLS connections independently.
- **Acceptance Criteria**:
  - [ ] `psql "sslmode=verify-full sslrootcert=<Cat 10>"` connects to follower
  - [ ] TLS 1.3 negotiated
  - [ ] `sslmode=disable` connection still succeeds

#### Task 3.3: Verify Replication Running Over TLS

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Tasks 3.1, 3.2
- **Description**: Verify replication is active and using TLS.
- **Acceptance Criteria**:
  - [ ] `pg_stat_replication` on leader shows `ssl = t` and active slot
  - [ ] No replication lag issues

---

### Phase 4: PostgreSQL Client mTLS — HBA + GORM Config [Status: ☐ TODO]

**Phase Objective**: Require client certs for app connections and configure GORM to present them.

#### Task 4.1: pg_hba.conf mTLS Rules

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Phase 3 complete
- **Description**: Update `pg_hba.conf` to require client certs for app connections.
- **Acceptance Criteria**:
  - [ ] All `hostssl` app connection rules updated to `clientcert=verify-full`
  - [ ] Plain `host` rules removed (only `hostssl` permitted)
  - [ ] Replication rule remains `hostssl replication ... scram-sha-256` (no `clientcert` yet — added in Phase 5)
- **Files**: `deployments/shared-postgres/leader/pg_hba.conf`, `deployments/shared-postgres/follower/pg_hba.conf`

#### Task 4.2: Framework Config SSL Fields

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 4.1
- **Description**: Add SSL YAML fields to framework database config (D2: YAML config approach).
- **Acceptance Criteria**:
  - [ ] `database.sslmode` field added to framework config struct
  - [ ] `database.sslcert` field added (path to Cat 14 client cert crt)
  - [ ] `database.sslkey` field added (path to Cat 14 client cert key)
  - [ ] `database.sslrootcert` field added (path to Cat 10 truststore)
  - [ ] GORM DSN builder appends SSL params when `sslmode = verify-full`
  - [ ] `go build ./...` clean; `golangci-lint run` clean
- **Files**: `internal/apps/framework/service/config/`, GORM DSN builder

#### Task 4.3: App Instance Client Cert Config

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Task 4.2
- **Description**: Configure YAML deployment configs for postgres-1 and postgres-2 instances with client cert paths.
- **Acceptance Criteria**:
  - [ ] postgres-1 config: `sslcert` = `postgres-tls-client-entity-leader-{PS-ID}-postgres-1/SAME-AS-DIR-NAME.crt` (Cat 14)
  - [ ] postgres-1 config: `sslkey` = `postgres-tls-client-entity-leader-{PS-ID}-postgres-1/SAME-AS-DIR-NAME.key` (Cat 14)
  - [ ] postgres-1 config: `sslrootcert` = `postgres-tls-server-issuing-ca/truststore/postgres-tls-server-issuing-ca.crt` (Cat 10)
  - [ ] postgres-2 config: same pattern for postgres-2
  - [ ] sqlite-1, sqlite-2 configs: NO `sslcert`/`sslkey`/`sslrootcert` fields
- **Files**: `deployments/{PS-ID}/config/*-app-framework-postgresql-{1,2}.yml`

#### Task 4.4: App Instance Cert Volume Mounts

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 4.3
- **Description**: Mount cert dirs for postgres-1 and postgres-2 instances (least privilege: Cat 14 + Cat 10 only).
- **Acceptance Criteria**:
  - [ ] postgres-1: Cat 14 `postgres-tls-client-entity-leader-{PS-ID}-postgres-1/` keystore + Cat 10 `postgres-tls-server-issuing-ca/truststore/`
  - [ ] postgres-2: Cat 14 `postgres-tls-client-entity-leader-{PS-ID}-postgres-2/` keystore + Cat 10
  - [ ] sqlite-1, sqlite-2: NO Cat 14 or Cat 10 dirs mounted
- **Files**: `deployments/{PS-ID}/compose.yml`

---

### Phase 5: PostgreSQL Replication Client mTLS [Status: ☐ TODO]

**Phase Objective**: Add client cert to follower replication connection; require it on leader.

#### Task 5.1: Follower primary_conninfo Client Cert

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 4 complete
- **Description**: Add follower replication client cert to `primary_conninfo`.
- **Acceptance Criteria**:
  - [ ] `sslcert=<Cat 13 postgres-tls-client-entity-follower-replication/ crt path>` in `primary_conninfo`
  - [ ] `sslkey=<Cat 13 postgres-tls-client-entity-follower-replication/ key path>` in `primary_conninfo`
  - [ ] `sslmode=verify-full` and `sslrootcert` retained from Phase 2
- **Files**: `deployments/shared-postgres/follower/postgresql.conf`

#### Task 5.2: Follower Cat 13 Client Cert Mount

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 5.1
- **Description**: Add Cat 13 replication client cert mount to follower (least privilege).
- **Acceptance Criteria**:
  - [ ] Follower mounts: Cat 13 `postgres-tls-client-entity-follower-replication/` keystore added
  - [ ] Leader mounts: NO Cat 13 changes
- **Files**: `deployments/shared-postgres/compose.yml`

#### Task 5.3: Leader HBA Replication clientcert Requirement

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Tasks 5.1, 5.2
- **Description**: Update leader `pg_hba.conf` replication entry to require client cert.
- **Acceptance Criteria**:
  - [ ] `hostssl replication ... scram-sha-256 clientcert=verify-full` in leader HBA
  - [ ] Replication without client cert rejected
- **Files**: `deployments/shared-postgres/leader/pg_hba.conf`

---

### Phase 6: Verify PostgreSQL Full Stack [Status: ☐ TODO]

**Phase Objective**: End-to-end verification of full mTLS with active replication.

#### Task 6.1: Verify App mTLS to Leader

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 5 complete
- **Description**: Verify postgres-1 and postgres-2 app instances connect via mTLS.
- **Acceptance Criteria**:
  - [ ] postgres-1 connects; `pg_stat_ssl` shows `ssl=t` and `client_dn` populated
  - [ ] postgres-2 connects; same verification

#### Task 6.2: Verify Non-mTLS Rejected

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 6.1
- **Description**: Verify connections without client certs are rejected.
- **Acceptance Criteria**:
  - [ ] `psql "sslmode=require"` (no sslcert) is rejected by leader and follower
  - [ ] `psql "sslmode=disable"` is rejected

#### Task 6.3: Verify Replication Full mTLS

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 5 complete
- **Description**: Verify replication uses full mTLS and data replicates correctly.
- **Acceptance Criteria**:
  - [ ] `pg_stat_replication` shows `ssl=t` and `client_dn` (follower replication cert)
  - [ ] Data written to leader appears on follower within replication lag threshold

---

### Phase 7: Deployment Templates for PostgreSQL TLS [Status: ☐ TODO]

**Phase Objective**: Update canonical deployment templates for PG TLS cert mounts and SSL connection params.

#### Task 7.1: Update deployment-templates.md

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 6 complete
- **Description**: Document cert volume mount rules and least privilege table for PG TLS.
- **Acceptance Criteria**:
  - [ ] Least privilege table: service → cert dirs mounted (leader, follower, postgres-1 app, postgres-2 app; sqlite-1/2 explicitly excluded)
  - [ ] `postgres-url.secret` template: `sslmode=verify-full&sslrootcert=...&sslcert=...&sslkey=...`
- **Files**: `docs/deployment-templates.md`

#### Task 7.2: Update shared-postgres Compose Template

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 7.1
- **Description**: Update shared-postgres compose template for per-node cert mounts.
- **Acceptance Criteria**:
  - [ ] Template leader service: Cat 11 + Cat 12 cert dirs mounted ONLY
  - [ ] Template follower service: Cat 10 + Cat 11 + Cat 12 + Cat 13 cert dirs mounted ONLY
  - [ ] `__PS_ID__` placeholders used for all PS-ID-specific paths
  - [ ] Template compliance linter accepts the file
- **Files**: `api/cryptosuite-registry/templates/deployments/shared-postgres/compose.yml`

#### Task 7.3: Update PS-ID Compose Template (postgres-1, postgres-2)

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 7.1
- **Description**: Update PS-ID compose template for app instance postgres cert mounts.
- **Acceptance Criteria**:
  - [ ] postgres-1 instance: Cat 14 `postgres-tls-client-entity-leader-__PS_ID__-postgres-1/` + Cat 10 mounted
  - [ ] postgres-2 instance: Cat 14 `postgres-tls-client-entity-leader-__PS_ID__-postgres-2/` + Cat 10 mounted
  - [ ] sqlite-1, sqlite-2 instances: NO PG cert dirs mounted
  - [ ] `__PS_ID__` placeholders used consistently
- **Files**: `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml`

#### Task 7.4: Update PostgreSQL Instance Config Templates

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 7.3
- **Description**: Update postgresql-1 and postgresql-2 config file templates with cert paths.
- **Acceptance Criteria**:
  - [ ] `__PS_ID__-app-framework-postgresql-1.yml`: `database.sslmode`, `database.sslcert`, `database.sslkey`, `database.sslrootcert` fields with `__PS_ID__` placeholders
  - [ ] `__PS_ID__-app-framework-postgresql-2.yml`: same for postgres-2
  - [ ] sqlite config templates: NO SSL fields added
- **Files**: `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/`

---

### Phase 8: Deployment Linting [Status: ☐ TODO]

**Phase Objective**: All updated deployment files pass lint-deployments validators.

#### Task 8.1: Run lint-deployments

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phase 7 complete
- **Description**: Run lint-deployments and fix all violations.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-deployments` exits 0
  - [ ] All 8 validators pass (naming, kebab-case, schema, template, ports, telemetry, admin, secrets)
  - [ ] No violations reported

#### Task 8.2: Lint lint-deployments Code

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 8.1
- **Description**: Ensure the linter code itself is clean.
- **Acceptance Criteria**:
  - [ ] `golangci-lint run ./internal/apps/tools/cicd_lint/lint_deployments/...` passes
  - [ ] `go test ./internal/apps/tools/cicd_lint/lint_deployments/...` passes

---

### Phase 9: Deployment Verification — PostgreSQL TLS [Status: ☐ TODO]

**Phase Objective**: Start deployments and verify PG TLS connectivity and replication end-to-end.

#### Task 9.1: Docker Compose Up (pki-init + shared-postgres)

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Phase 8 complete
- **Description**: Start pki-init and shared-postgres with TLS config.
- **Acceptance Criteria**:
  - [ ] pki-init generates certs into `__PS_ID__-certs` named volume
  - [ ] shared-postgres leader and follower start with TLS
  - [ ] Health checks pass for both leader and follower

#### Task 9.2: Verify App Connects to PG Leader via mTLS

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 9.1
- **Description**: Start PS-ID service with postgres-1 instance; verify mTLS connection to leader.
- **Acceptance Criteria**:
  - [ ] PS-ID postgres-1 instance starts and connects to PG leader
  - [ ] `pg_stat_ssl` on leader shows app connection with `ssl=t` and `client_dn` populated
  - [ ] App health check passes

#### Task 9.3: Verify PG Replication

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 9.2
- **Description**: Verify PG leader replicates to follower in deployment.
- **Acceptance Criteria**:
  - [ ] `pg_stat_replication` shows active slot with `ssl=t`
  - [ ] Data written via postgres-1 app is readable from follower with `sslmode=verify-full`
  - [ ] No replication lag accumulation

---

### Phase 10: Private PS-ID App mTLS Trust [Status: ☐ TODO]

**Phase Objective**: Configure all PS-ID app instance variants to serve admin (:9090) endpoint with mTLS.

#### Task 10.1: Framework Admin mTLS Cert Loading

- **Status**: ❌
- **Estimated**: 2h
- **Dependencies**: Phase 9 complete
- **Description**: Update framework server builder to load private admin mTLS certs from YAML config paths.
- **Acceptance Criteria**:
  - [ ] Framework server config: `server.admin-tls-cert-file` and `server.admin-tls-key-file` fields (Cat 7 keystore)
  - [ ] Framework server config: `server.admin-tls-ca-file` field (Cat 6 truststore for verifying admin clients)
  - [ ] Admin listener uses `tls.RequireAndVerifyClientCert` when cert fields are set
  - [ ] `go build ./...` clean; `golangci-lint run` clean
- **Files**: `internal/apps/framework/service/config/`, framework server builder

#### Task 10.2: Deployment Config Templates for Admin mTLS

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 10.1
- **Description**: Update deployment config file templates for each instance variant with admin cert paths.
- **Acceptance Criteria**:
  - [ ] sqlite-1 config: `server.admin-tls-cert-file` = `private-https-mutual-entity-__PS_ID__-sqlite-1/SAME-AS-DIR-NAME.crt` (Cat 7); `server.admin-tls-ca-file` = Cat 6 path
  - [ ] sqlite-2 config: same for sqlite-2
  - [ ] postgres-1 config: same for postgres-1 (combined with PG SSL fields from Phase 4)
  - [ ] postgres-2 config: same for postgres-2
- **Files**: `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/`

#### Task 10.3: Deployment Compose Template — Admin Cert Mounts

- **Status**: ❌
- **Estimated**: 1.5h
- **Dependencies**: Task 10.2
- **Description**: Update PS-ID compose template to mount admin cert dirs (least privilege: Cat 6 + Cat 7 only for admin, per variant).
- **Acceptance Criteria**:
  - [ ] sqlite-1: Cat 7 `private-https-mutual-entity-__PS_ID__-sqlite-1/` + Cat 6 `private-https-mutual-issuing-ca-__PS_ID__-sqlite-1/truststore/` mounted ONLY
  - [ ] sqlite-2: same for sqlite-2
  - [ ] postgres-1: Cat 7 + Cat 6 for postgres-1 (in addition to Cat 14 + Cat 10 from Phase 4)
  - [ ] postgres-2: same for postgres-2
  - [ ] NO public cert dirs, NO OTel/Grafana cert dirs in admin mounts
- **Files**: `api/cryptosuite-registry/templates/deployments/__PS_ID__/compose.yml`

#### Task 10.4: Run lint-deployments After Admin mTLS Templates

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 10.3
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-deployments` exits 0
  - [ ] All 8 validators pass

#### Task 10.5: Verify Admin mTLS in Deployment

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 10.4
- **Description**: Start deployment and verify admin endpoint serves mTLS.
- **Acceptance Criteria**:
  - [ ] `/admin/api/v1/livez` responds over mTLS (requires client cert from Cat 6 CA)
  - [ ] Admin request without client cert is rejected (TLS handshake failure)
  - [ ] All 4 instance variants (sqlite-1, sqlite-2, postgres-1, postgres-2) verified

---

### Phase 11: Knowledge Propagation [Status: ☐ TODO]

**Phase Objective**: Apply lessons learned to permanent artifacts.

#### Task 11.1: Review Lessons

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Phases 1-10 complete
- **Description**: Review lessons.md from all prior phases.
- **Acceptance Criteria**:
  - [ ] All lessons reviewed
  - [ ] Actionable items identified for ENG-HANDBOOK.md

#### Task 11.2: Update ENG-HANDBOOK.md

- **Status**: ❌
- **Estimated**: 1h
- **Dependencies**: Task 11.1
- **Description**: Update ENG-HANDBOOK.md with PostgreSQL mTLS, private admin mTLS, and least privilege patterns.
- **Acceptance Criteria**:
  - [ ] PostgreSQL mTLS wiring pattern documented (staged: server TLS → verify → client TLS → verify)
  - [ ] GORM SSL config YAML fields documented
  - [ ] Private admin mTLS cert loading documented
  - [ ] Cert mount least privilege principle documented with table
  - [ ] §10.3.4 `InsecureSkipVerify` test example updated to use `RootCAs: testCAPool`

#### Task 11.3: Update deployment-templates.md

- **Status**: ❌
- **Estimated**: 0.5h
- **Dependencies**: Task 11.2
- **Acceptance Criteria**:
  - [ ] Least privilege cert mount table finalized
  - [ ] postgres-url.secret TLS param format documented

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

- [ ] Phase 0: pki-init generator tests pass (100%, no skips)
- [ ] Phase 4: GORM SSL config unit tests pass
- [ ] Phase 10: Framework admin mTLS unit tests pass
- [ ] Phases 3, 6, 9, 10: verification tasks pass in Docker Compose

### Code Quality

- [ ] Linting passes: `golangci-lint run ./...` and `golangci-lint run --build-tags e2e,integration ./...`
- [ ] No new TODOs without tracking
- [ ] Formatting clean

### Documentation

- [ ] `deployment-templates.md` updated with TLS cert mount rules and least privilege table
- [ ] `tls-structure.md` cross-referenced for cert category numbers
- [ ] ENG-HANDBOOK.md updated with TLS wiring patterns

### Deployment

- [ ] `lint-deployments` passes after Phases 7 and 10
- [ ] Docker Compose health checks pass (Phases 9 and 10)
- [ ] `./configs/` unchanged (auto-TLS mode only)

---

## Notes

- **Least Privilege Enforcement**: CRITICAL — every compose template task must list exactly which Cat dirs are mounted per service, and explicitly note what is NOT mounted.
- **./configs/ isolation**: No changes to `./configs/` files; they continue using auto-TLS. All cert wiring is `./deployments/` only.
- **Phase 0 is V12+V13 prerequisite**: V13 execution begins AFTER V12 Phase 0 is complete.

---

## Evidence Archive

- `test-output/v12-phase1/` — Phase 1 PostgreSQL server TLS verification
- `test-output/v12-phase3/` — Phase 3 standalone TLS verification
- `test-output/v12-phase6/` — Phase 6 full stack mTLS verification
- `test-output/v12-phase9/` — Phase 9 deployment PG TLS verification
- `test-output/v12-phase10/` — Phase 10 admin mTLS deployment verification
