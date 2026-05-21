# Framework V23 — Followup Items

All Framework V23 code changes are **complete and committed**. All 10 PS-ID e2e test
stacks were fixed for a post-V23 regression (distroless bash error) in a follow-up commit.
One item remains open: the `postgres-leader` crash in sm-im and other services that still
needs root-cause evidence from a live container run.

---

## Numbered Followup Items

### 0. Post-V23 Regression: All 10 PS-ID e2e Tests Fail — Bash in Distroless Container

**Status**: **FIXED** (committed `fix(deployments): replace bash entrypoint in distroless
otel-collector-contrib with Alpine init container`)

**Origin**: V23 named-volume cert architecture change introduced a bash entrypoint override
in the distroless `otel/opentelemetry-collector-contrib` container to symlink certs from
the new `{PS-ID}-certs` named volume into `/etc/pki-init/certs`.

**All 5 tested services failed with identical error**:
```
Error response from daemon: failed to create task for container: failed to create shim task:
OCI runtime create failed: runc create failed: unable to start container process:
error during container init: exec: "/bin/bash": stat /bin/bash: no such file or directory
```
(Verified: skeleton-template, sm-kms, jose-ja, pki-ca, sm-im — all same root cause.
The sm-im postgres error documented in item #1 below is a *separate* secondary issue
that was obscured by this bash error.)

**Root cause**: `otel/opentelemetry-collector-contrib` is a distroless image with no shell.
The V23 change added `entrypoint: /bin/bash` with a symlink setup script — this fails on
every container start.

**Fix applied** (all 10 PS-ID compose files + canonical template):
- Added `otel-certs-init` Alpine init container per PS-ID that:
  1. Runs after `pki-init` completes (`depends_on: service_completed_successfully`)
  2. Mounts `{PS-ID}-certs` volume read-only at `/mnt/ps-certs-src`
  3. Copies (via `cp -r`) 4 cert directories into new `{PS-ID}-otel-certs` volume (writable)
  4. Exits 0
- `opentelemetry-collector-contrib` now depends on `otel-certs-init` and mounts
  `{PS-ID}-otel-certs` read-only — no bash, no entrypoint override needed
- `cp -r` used instead of `ln -sf` because symlinks pointing to `/mnt/ps-certs-src/` would
  be dangling in the distroless container (that volume is only mounted in the init container)
- New `{PS-ID}-otel-certs:` volume added to volumes section in each compose file

**Validation**: `lint-deployments` 65/65 validators pass. `lint-fitness` passes.

---

### 1. sm-im E2E Test Fails: `cryptoutil-postgres-leader exited (1)` During Stack Startup

**Current status**: Open blocker (runtime, infrastructure). **Now unmasked** — the bash/distroless error (item #0) was previously hiding this failure for sm-im. With item #0 fixed, this postgres-leader crash is the next blocker for sm-im e2e.

**Origin**: Task 4.2 runtime pass criterion (V23 tasks.md)

**Failing command**:
```
go test -tags e2e ./internal/apps/sm-im/e2e/...
```

**Failure message** (from `test-output/v23-phase4/go-test-sm-im-e2e.txt`, around line 250):
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
- Despite correct cert paths and secrets, `cryptoutil-postgres-leader` still exits with code 1 at runtime.

**Root cause still unknown**. Working theories:
- a) PostgreSQL SSL setup error (cert file permissions inside container, ownership issue)
- b) `init-leader-databases.sql` initdb script fails (creates 10+ databases; may exceed
     `start_period: 30s` + 5 retries × 10s = 80s healthcheck window)
- c) The bash script's `-c` string ends with `exec docker-entrypoint.sh postgres -c
     config_file=...` BUT the base `command:` from shared-postgres
     (`["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf"]`) is appended as
     argv to the bash `-c` script, potentially causing argument collision
- d) Volume timing: pki-init completes (exit 0) but certs not yet flushed to shared named
     volume before postgres-leader reads them

**NOTE**: This error was previously obscured by the bash/distroless error (item #0). Now
that item #0 is fixed, the sm-im e2e test may reveal a *new* error message for the postgres
failure that supersedes theories (a)–(d) above. Capture fresh container logs before
assuming one of the above theories.

**Diagnostic steps for the next session**:
1. Manually start the stack:
   ```
   docker compose -f deployments/sm-im/compose.yml up --no-build pki-init postgres-leader
   ```
2. Immediately capture postgres-leader logs:
   ```
   docker logs cryptoutil-postgres-leader 2>&1
   ```
3. Capture compose event context if needed:
  ```
  docker compose -f deployments/sm-im/compose.yml logs --no-color postgres-leader pki-init
  ```
1. The exact PostgreSQL error should appear (for example, "could not open certificate file",
   "FATAL:  SSL error", initdb failure, etc.)
2. Cross-reference the error against theories (a)–(d) above

**Closure criteria for this followup item**:
- Root cause identified with concrete error output saved under `test-output/v23-phase4/`.
- Corrective change implemented in deployment/runtime configuration.
- `go test -tags e2e ./internal/apps/sm-im/e2e/...` passes.

**Code location**: `deployments/sm-im/compose.yml` postgres-leader service override
(search for `exec docker-entrypoint.sh`)

**Scope note**: The skip-constant refactor (the actual V23 Task 4.2 code change) is complete and
correct. The remaining failure is in e2e runtime/deployment behavior and must be resolved with
container-level evidence from the failing `postgres-leader` startup path.

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
- Build and static quality gates pass: `go build ./...`, `lint-deployments` (65/65),
  `lint-fitness`, `lint-docs`
- Runtime e2e gate still open: `go test -tags e2e ./internal/apps/sm-im/e2e/...`
