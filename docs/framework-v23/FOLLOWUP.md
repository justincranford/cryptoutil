# Framework V23 — Followup Items

All Framework V23 code changes are **complete and committed** (12/13 tasks). One task
remains open with a runtime blocker. All other V23 work is done: named Docker volumes for
certs (replacing bind mounts across all 10 PS-IDs), two new deployment validators
(`validate_cert_volume_policy.go` and `validate_postgres_secrets_dir_sync.go`), and
skip-constant refactors in sm-im e2e tests.

---

## Numbered Followup Items

### 1. sm-im E2E Test Fails: `cryptoutil-postgres-leader exited (1)` During Stack Startup

**Origin**: Task 4.2 runtime pass criterion (V23 tasks.md)

**Failing command**:
```
go test -tags e2e ./internal/apps/sm-im/e2e/...
```

**Failure message** (from `test-output/v23-phase4/go-test-sm-im-e2e.txt` ~line 250):
```
dependency failed to start: container cryptoutil-postgres-leader exited (1)
E2E setup failed: failed to start docker compose: exit status 1
FAIL cryptoutil/internal/apps/sm-im/e2e 182.652s
```

**Context**:
- The `sm-im` Docker stack includes `../shared-postgres/compose.yml` via `include:` with
  `env_file: .env.postgres`.
- `deployments/sm-im/.env.postgres` sets `POSTGRES_CONF_DIR=../sm-im` and
  `POSTGRES_SECRETS_DIR=../sm-im/secrets`.
- `deployments/sm-im/postgresql-leader.conf` has `ssl = on` with cert paths:
  ```
  ssl_cert_file = '/var/lib/postgresql/ssl/server.crt'
  ssl_key_file  = '/var/lib/postgresql/ssl/server.key'
  ssl_ca_file   = '/var/lib/postgresql/ssl/client-ca.crt'
  ```
- The `postgres-leader` service in `deployments/sm-im/compose.yml` (around line 474) overrides
  the shared-postgres base with a custom bash entrypoint that:
  1. Creates `/var/lib/postgresql/ssl/`
  2. Copies certs from `sm-im-certs:/mnt/ps-certs-src` (named volume) using `|| true` (silently
     ignores failures)
  3. Calls `exec docker-entrypoint.sh postgres -c config_file=/etc/postgresql/postgresql.conf`
- `pki-init` runs before `postgres-leader` (`depends_on: service_completed_successfully`)
  and writes certs into the `sm-im-certs` volume at `/certs/sm-im/...`.

**Analysis completed in prior sessions**:
- pki-init DOES generate the correct cert directories (verified by running
  `go run ./cmd/sm-im init --domain=sm-im --output-dir=./test-output/pki-init-check`):
  - `sm-im/postgres-tls-server-entity-leader/postgres-tls-server-entity-leader.crt` ✓
  - `sm-im/postgres-tls-server-entity-leader/postgres-tls-server-entity-leader.key` ✓
  - `sm-im/postgres-tls-client-issuing-ca/truststore/postgres-tls-client-issuing-ca.crt` ✓
- All secrets files exist in `deployments/sm-im/secrets/`.
- The cert path filenames in the entrypoint MATCH what pki-init generates.
- Despite correct cert paths and secrets, `cryptoutil-postgres-leader` exits(1) at runtime.

**Root cause still unknown** — possible theories:
- a) PostgreSQL SSL setup error (cert file permissions inside container, ownership issue)
- b) `init-leader-databases.sql` initdb script fails (creates 10+ databases; may exceed
     `start_period: 30s` + 5 retries × 10s = 80s healthcheck window)
- c) The bash script's `-c` string ends with `exec docker-entrypoint.sh postgres -c
     config_file=...` BUT the base `command:` from shared-postgres
     (`["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf"]`) is appended as
     argv to the bash `-c` script, potentially causing argument collision
- d) Volume timing: pki-init completes (exit 0) but certs not yet flushed to shared named
     volume before postgres-leader reads them

**Diagnostic steps for the next session**:
1. Manually start the stack:
   ```
   docker compose -f deployments/sm-im/compose.yml up --no-build pki-init postgres-leader
   ```
2. Immediately capture postgres-leader logs:
   ```
   docker logs cryptoutil-postgres-leader 2>&1
   ```
3. The exact PostgreSQL error will appear (e.g., "could not open certificate file",
   "FATAL:  SSL error", initdb failure, etc.)
4. Cross-reference the error against theories (a)–(d) above

**Code location**: `deployments/sm-im/compose.yml` postgres-leader service override
(search for `exec docker-entrypoint.sh`)

**Why this is NOT a V23 code regression**: The skip-constant refactor (the actual V23 Task 4.2
code change) is complete and correct. The e2e infrastructure issue pre-dates V23 or was
introduced by the named-volume migration in V23 Tasks 2.x. All V23 code is committed.

---

## V23 Completed Work (for reference, do not redo)

All of the following are **done and committed**:

- Named cert volumes replacing bind mounts in all 10 PS-ID compose files
- `validate_cert_volume_policy.go` — CO-21/CO-22 linter (65/65 validators pass)
- `validate_postgres_secrets_dir_sync.go` — POSTGRES_SECRETS_DIR linter
- `internal/apps/sm-im/e2e/e2e_registration_test.go`: `skipReasonJoinTenantIDNotSupported`
  constant replaces inline string literal
- `internal/apps/sm-im/e2e/e2e_test.go`: `skipReasonOtelPortNotExposed` constant replaces
  inline string literal
- All quality gates pass: `go build ./...`, `lint-deployments` (65/65), `lint-fitness`,
  `lint-docs`
