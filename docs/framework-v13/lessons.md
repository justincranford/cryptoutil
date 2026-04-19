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

*(To be filled during Phase 2 execution using the 4-section structure above)*

---

## Phase 3: v11 Mutation & Race Testing

*(To be filled during Phase 3 execution using the 4-section structure above)*

---

## Phase 4: Documentation Synchronization

*(To be filled during Phase 4 execution using the 4-section structure above)*

---

## Phase 5: Template Docker Validation

*(To be filled during Phase 5 execution using the 4-section structure above)*

---

## Phase 6: Knowledge Propagation

*(To be filled during Phase 6 execution using the 4-section structure above)*
