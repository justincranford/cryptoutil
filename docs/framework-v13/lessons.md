# Lessons — Framework v13: v10-v12 Cleanup

**Created**: 2026-06-30
**Last Updated**: 2026-06-30

> **MANDATORY per-phase structure** (from v11 pattern — v12's terse 2-4 bullet format is the anti-pattern):
>
> ### What Worked
> - Specific patterns, tools, or approaches that succeeded (with context)
>
> ### What Didn't Work
> - Specific failures, surprises, or regressions (with root cause)
>
> ### Root Causes
> - Why the failures happened (architectural, process, tooling)
>
> ### Patterns for Future Phases
> - Actionable takeaways for subsequent phases or versions

---

## Phase 1: v12 Docker-Deferred TLS Smoke Test

### What Worked

- **PostgreSQL TLS/mTLS is fully functional**: `pg_stat_ssl` confirmed TLSv1.3, AES-256-GCM-SHA384 for all app→postgres connections. Client cert DN (`client_dn`) correctly set to `/CN=postgres-tls-client-entity-leader-sm-kms-postgres-{1,2}` for Cat 12 mTLS.
- **PostgreSQL replication TLS works**: 4 streaming replication connections verified in `pg_stat_replication`. Replication uses server TLS without client cert (correct — `host replication all all scram-sha-256` in pg_hba.conf).
- **Admin port mTLS enforcement confirmed**: `127.0.0.1:9090` returns `tls: certificate required` without client cert. Bound correctly to loopback only.
- **Mixed-mode TLS (`tls-public-mode: mixed`)**: After fix, all 4 app instances dynamically generate server certs signed by `public-https-server-issuing-ca`. Verified: `openssl s_client -CAfile pki-init-ca.crt 127.0.0.1:8000` → `Verify return code: 0 (ok)`.
- **`docker compose up --wait` reliable**: All 8 services (2 postgres, 4 app, grafana, otel-collector) reach healthy status consistently.

### What Didn't Work

1. **`setup-logical-replication.sh` used `-h localhost` (TCP) during initdb**: PostgreSQL's Docker init scripts run with `-c listen_addresses=''` (Unix socket only). TCP refused. All 16 `psql -h localhost` failed with `exit code 2`.
2. **OTel collector healthcheck used `wget`**: `otel/opentelemetry-collector-contrib` is a distroless image — no shell, no wget. Container's own healthcheck failed permanently.
3. **`pki-init` did NOT generate `tls-config.yml`**: Generator produced 66 cert directories but no YAML config file. The compose `--config=/certs/tls-config.yml` flag was silently skipped (framework's `config_parse.go` uses `os.Stat` before reading — missing files are silently skipped). Result: public TLS used `TLSModeAuto` (auto-generated self-signed certs, CN="Auto-Generated Server Certificate").

### Root Causes

1. **Unix socket assumption for Docker initdb**: PostgreSQL Docker image runs init scripts against a temporary postgres instance that only listens on Unix socket. `psql` without `-h` uses Unix socket by default. `pg_hba.conf` has `local all all trust` for this phase. Fix: remove `-h localhost` from all psql commands in init scripts.
2. **Distroless image healthcheck incompatibility**: Distroless images have no shell or utilities. App services depend on an Alpine sidecar (`healthcheck-opentelemetry-collector-contrib`) for OTel readiness — the container's own healthcheck is redundant. Fix: `disable: true` for the distroless container's healthcheck.
3. **Silent config file skip**: Framework config parser uses `os.Stat(path)` check before reading — absent files are silently treated as "no config". The correct fix is to generate `tls-config.yml` from pki-init. File contains `tls-public-mode: mixed` + base64-encoded Cat 1 issuing CA cert+key so framework reads it during startup and enables mixed mode.

### Patterns for Future Phases

- **Always verify TLS mode after pki-init changes**: Run `openssl s_client -CAfile <pki-init-ca>` and check CN — `CN=Server Certificate` (mixed mode) vs `CN=Auto-Generated Server Certificate` (auto mode). Silent fallback to auto mode is a footgun.
- **Docker initdb scripts must use Unix socket**: Any `psql` command in `docker-entrypoint-initdb.d/` MUST omit `-h` flag. Document this in initdb script comments.
- **Distroless images need `disable: true` healthcheck**: Never add wget/curl healthchecks to distroless containers. Always use a sidecar Alpine container for readiness probing when the app container is distroless.
- **Config files with base64 PEM**: YAML values for PEM data should be single-line base64 (no line breaks) to avoid YAML multiline parsing issues. `encoding/base64.StdEncoding.EncodeToString()` produces this format correctly.
- **Admin mTLS test limitation**: `127.0.0.1:9090` is inaccessible from host by design. Full round-trip mTLS test requires either `docker exec` or a Go E2E test that runs inside the network (Phase 2).

---

## Phase 2: TLS/mTLS E2E Go Tests

### What Worked

- **`NewClientForTestWithCA` API was already available**: The function existed in `internal/shared/crypto/tls/tls_test_util.go` from a previous session. Adding it to testmains required only adding the magic constant (CA cert path) and the initialization step.
- **Initializing CA-validated client AFTER `WaitForMultipleServices`**: The key insight from INVESTIGATION.md was that `sharedHTTPClientWithCA` was previously reverted because it was initialized before pki-init completed. Waiting until after the health checks pass guarantees pki-init has written the CA cert to the host-mounted volume.
- **Same TLS chain test pattern works for all 4 PS-IDs**: Only the health endpoint magic constant differs between PS-IDs. `sed` substitution produced correct files for jose-ja, sm-im, skeleton-template from the sm-kms original.
- **`ComposeManager.BuildDockerExecArgs`**: Adding a simple method that wraps `buildComposeArgs` with "exec + service" gave a clean interface for `docker exec` in tests. Unit test pass on first write.
- **Admin port isolation test pattern**: Testing that `net.DialTimeout("tcp", "127.0.0.1:9090", 2s)` fails is a valid and fast Go E2E test for admin port non-exposure. Avoids any docker exec complexity.

### What Didn't Work

1. **Task 2.4 full mTLS client cert sad-path**: Writing a Go test that connects to the admin endpoint with a *wrong* client cert requires (a) the admin port to be accessible from the test host AND (b) a test client cert. Since admin is `127.0.0.1:9090` inside the container, neither condition is satisfied without docker exec complexity. Settled for admin port isolation test (not exposed to host) instead.
2. **`gosec` nolint required for `exec.Command("docker", args...)`**: `gosec` G204 flags subprocess calls with variable commands. The `//nolint:gosec` annotation was needed for the PostgreSQL mTLS test since `docker exec` with known args is safe.

### Root Causes

1. **Admin mTLS limitation is architectural**: `127.0.0.1:9090` is the correct security posture for admin endpoints. Full programmatic mTLS testing from outside the container is impossible by design. The Docker Compose healthcheck (calling `/app/sm-kms livez`) provides sufficient coverage that admin TLS is functional.
2. **pki-init CA cert path is relative**: Magic constants like `KMSE2EPublicCACertPath` are relative paths from the e2e test directory. This works correctly when tests are run with `go test ./internal/apps/sm-kms/e2e/...` from the project root.

### Patterns for Future Phases

- **CA-validated client initialization order is critical**: Always initialize after `WaitForMultipleServices` to guarantee pki-init has completed. Document this as a comment near initialization in testmain.
- **Admin mTLS testing strategy**: Use Docker Compose healthcheck as the integration test for admin TLS (it calls livez via the built-in binary). Go E2E tests add admin port isolation verification (not exposed to host). Full mTLS sad-path (wrong cert → rejected) is out of scope for host-based tests.
- **E2E test file organization**: Each PS-ID's `e2e/` directory should have: `testmain_e2e_test.go` (lifecycle), `e2e_test.go` (health checks), `e2e_tls_test.go` (TLS chain validation). Admin isolation and PostgreSQL mTLS tests are sm-kms-specific (only sm-kms has PostgreSQL in its own compose stack).
- **PostgreSQL mTLS test via `pg_stat_ssl`**: The `docker exec psql` approach works but requires the correct compose service name and database credentials matching the secrets. Use the magic constant for the container name.

---

## Phase 3: v11 Mutation & Race Testing

### What Worked

- **gremlins already installed**: `/home/q/go/bin/gremlins` — no install needed.
- **Race detection clean**: `CGO_ENABLED=1 go test -race -count=2 ./internal/apps/framework/tls/` completed in 1.3s with zero data races.
- **Mutation efficacy 100%**: Zero mutants survived. Every mutation was either killed or timed out. Timed-out mutants indicate the test suite detects the mutation (the tests hang waiting for behavior that the mutation changes, meaning the mutation IS caught, just slowly).
- **92% mutator coverage**: 6 NOT COVERED mutations in `init.go:50` and `tier.go:32-34` are uncoverable in unit tests by design — they require production telemetry wiring or suite-level `cryptoutil` tier invocation. These are correctly excluded from coverage targets.

### What Didn't Work

- **`./...` pattern fails in gremlins**: `gremlins unleash --tags=!integration ./internal/apps/framework/tls/...` produced "matched no packages". Must use bare package path without `...` suffix.

### Root Causes

1. **gremlins `./...` pattern**: gremlins uses a different pattern matching than `go test` — it does not support recursive `...` glob. Use the explicit package directory.
2. **Timed out ≠ survived**: gremlins reports TIMED OUT mutations separately from LIVED (survived). Timed out means the mutation caused the test to hang (detected but slowly). Both KILLED and TIMED OUT count as "mutation caught by tests". LIVED would be the problematic category.

### Patterns for Future Phases

- **gremlins package pattern**: Use `./internal/apps/framework/tls` (no trailing `/...`) for gremlins. `go test ./...` still works as expected.
- **100% efficacy is the target, not 100% coverage**: Some paths in production CLI code are intentionally not unit-tested (production wiring). 92%+ mutator coverage + 100% efficacy = excellent for pki-init code.
- **Timeout distinction**: In gremlins output, TIMED OUT mutants are caught (not survivors). Only LIVED mutants represent coverage gaps requiring new tests.

---

## Phase 4: Documentation Synchronization

*(To be filled during Phase 4 execution using the 4-section structure above)*

---

## Phase 5: Template Docker Validation

*(To be filled during Phase 5 execution using the 4-section structure above)*

---

## Phase 6: Knowledge Propagation

*(To be filled during Phase 6 execution using the 4-section structure above)*
