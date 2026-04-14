# Quizme v1 — Framework v11: TLS Integration for Shared Services

**Purpose**: These questions surface genuine architectural decisions the user must make before
implementation can begin. Each question has been verified against the codebase — answers are not
discoverable by code reading alone; they require product/project-level judgment.

**Instructions**: Fill in one of A, B, C, D, or E after each **Answer:** line.
- E is a blank option for custom answers not covered by A–D.
- After answering all questions, the agent will merge decisions into plan.md and delete this file.

---

## Question 1: /certs Volume Strategy

**Background**: pki-init generates all TLS material under one `--output-dir=/certs`. The current
PS-ID compose template uses a bind mount: `./certs/:/certs/:rw` for pki-init and `./certs/:/certs/:ro`
for app containers. Both resolve relative to the PS-ID's deployment directory (e.g.,
`deployments/sm-kms/certs/`).

**Problem**: When `shared-telemetry/compose.yml` is included into a PS-ID compose, Docker Compose
resolves paths in the included file relative to `deployments/shared-telemetry/` — NOT the including
file's directory. So OTel Collector and Grafana LGTM in shared-telemetry would look in
`deployments/shared-telemetry/certs/`, while the PS-ID pki-init writes to `deployments/sm-kms/certs/`.
These are different directories; the certs are never seen by the shared services.

**Resolution needed**: How should OTel Collector and Grafana access the pki-init-generated certs?

**A)** Named Docker volume — define `certs:` named volume at compose level; pki-init writes to
   `certs:/certs:rw`; all consumers (OTel, Grafana, app containers, shared-postgres) mount
   `certs:/certs:ro`. A single named volume (e.g., `cryptoutil-certs`) is created once per stack
   and shared by all services in the network. This requires changing the existing PS-ID template
   `./certs/` bind mount to a named volume.

**B)** Canonical shared bind mount path — change pki-init in the PS-ID template to write to
   `../shared-telemetry/certs:/certs:rw` (resolves to `deployments/shared-telemetry/certs/`).
   shared-telemetry services use `./certs:/certs:ro` (same resolved path). All PS-ID app containers
   also change to `../shared-telemetry/certs:/certs:ro`. One canonical directory holds all certs for
   the entire stack; PS-ID-specific certs go in subdirectories of that shared path.

**C)** Separate standalone pki-init in shared-telemetry — add a `pki-init` service inside
   `shared-telemetry/compose.yml` that runs `--domain=cryptoutil --output-dir=/certs` and writes to
   `./certs:/certs:rw` (resolves to `deployments/shared-telemetry/certs/`). PS-ID compose pki-init
   continues to write to `./certs/` (PS-ID directory). The shared services get the `ALL-*` certs
   from their own init; PS-ID-specific certs are only in the PS-ID's local `./certs/`.
   PS-ID app containers still use `./certs/` for their own branch of certs; they get OTel CA trust
   from `./certs/ALL-telemetry-otel-private-server/` (generated in their local run, since all pki-init
   domains always generate shared `ALL-*` certs).

**D)**

**Answer**:

**Rationale for question**: This is the foundational architecture decision for v11. Incorrect choice
causes a Docker Compose path mismatch where shared services cannot access TLS certs at all.

---

## Question 2: PostgreSQL TLS Scope in v11

**Background**: The user request says "change their postgres-url.secret to remove `?sslmode=disable`".
Current format: `postgres://user:pass@shared-postgres-leader:5432/db?sslmode=disable`.
pki-init already generates PostgreSQL server certs (`ALL-db-postgres-private-server/`) AND
client CAs (`ALL-db-postgresql-leader-private-client/`, `ALL-db-postgresql-follower-private-client/`).

**Implementing PostgreSQL TLS requires**:
1. `shared-postgres/postgresql-leader.conf` and `postgresql-follower.conf` — enable `ssl = on`,
   set `ssl_cert_file` and `ssl_key_file` to pki-init cert paths
2. Adding `/certs` volume mount to both postgres containers
3. Adding pki-init dependency to `shared-postgres/compose.yml`
4. Updating all 10 PS-ID `postgres-url.secret` files
5. For full mTLS: also update `pg_hba.conf` to require client certs

**A)** Full mTLS — PostgreSQL requires client certs from connecting apps; update `pg_hba.conf` to
   `hostssl all all all scram-sha-256 clientcert=verify-full`; each PS-ID app instance presents a
   client cert from the `ALL-db-postgresql-leader-private-client` or `follower` CA.
   postgres-url.secret format becomes:
   `postgres://user:pass@...:5432/db?sslmode=verify-full&sslrootcert=/certs/...ca.pem&sslcert=/certs/...crt.pem&sslkey=/certs/...key.pem`

**B)** One-way TLS (server cert only) — PostgreSQL presents its server cert; apps verify but do NOT
   present a client cert. `pg_hba.conf` unchanged (still uses password auth). postgres-url.secret
   format becomes: `postgres://user:pass@...:5432/db?sslmode=verify-ca&sslrootcert=/certs/ALL-db-postgres-private-server/ALL-db-postgres-private-server-ca-crt.pem`

**C)** Defer entirely to v12 — Only OTel and Grafana TLS in v11; `?sslmode=disable` stays;
   a v12 ticket is created for PostgreSQL TLS.

**D)**

**Answer**:

**Rationale for question**: PostgreSQL mTLS (Option A) requires Go-driver connection string parameters
embedded in a Docker secret, plus conf file changes in shared-postgres. Option B is simpler but still
encrypted. Option C avoids scope expansion entirely. The right answer depends on security requirements
vs implementation effort the user wants in v11 vs v12.

---

## Question 3: OTel Collector mTLS vs One-Way TLS (PS-ID apps → OTel path)

**Background**: pki-init generates `ALL-telemetry-otel-private-client/` CA for mTLS client auth.
However, pki-init does NOT currently generate individual app-instance OTel client certificate leaves
(only the CA). For mTLS, generator.go would need to be updated to create ~40 new client cert leaves
(4 app instances × 10 PS-IDs) under this CA. This is a code change to `internal/apps/framework/tls/generator.go`.

For one-way TLS, apps only need to trust the OTel server cert (using the `ALL-telemetry-otel-private-server`
issuing CA). No generator changes required. All connections are still encrypted.

**A)** Full mTLS — update `generator.go` to create per-app-instance OTel client cert leaves;
   update all PS-ID compose files to pass these client certs when connecting to OTel;
   update `otel-collector-config.yaml` to require client certs via `client_ca_file`.
   Higher security; generator code change required; ~40 new cert files in /certs.

**B)** One-way TLS — apps verify OTel server cert; OTel does NOT require client certs;
   `client_ca_file` omitted from `otel-collector-config.yaml`;
   generator unchanged; simpler PS-ID compose configs;
   connections still encrypted but OTel cannot verify which app is connecting.

**C)** No TLS between PS-ID apps and OTel (not recommended; user request explicitly includes OTel TLS)

**D)**

**Answer**:

**Rationale for question**: mTLS prevents unauthorized clients from pushing data to the OTel Collector,
but requires a generator code change and 40 new cert files. One-way TLS is still encrypted with no
code change. Security posture vs implementation complexity tradeoff.

---

## Question 4: OTel Collector → Grafana LGTM TLS Protocol

**Background**: The current `otel-collector-config.yaml` exports via `otlphttp` to
`http://grafana-otel-lgtm:4318`. The `grafana/otel-lgtm` image bundles its own OTel Collector
that receives on both port 4317 (gRPC) and 4318 (HTTP). Enabling TLS on the OTel Collector →
Grafana LGTM link requires:
1. Changing the exporter endpoint to HTTPS (or gRPC TLS)
2. Adding a `tls:` stanza in the exporter config (cert for the OTel→Grafana client, CA for Grafana's server cert)
3. If TLS for the bundled OTel receiver inside grafana-otel-lgtm: creating a custom
   `/otel-lgtm/otelcol-config.yaml` and bind-mounting it into the container

**A)** `otlphttp` HTTPS to port 4318 — minimal change from current; add `tls:` stanza to
   `otlphttp` exporter; change endpoint to `https://grafana-otel-lgtm:4318`; mount custom
   `/otel-lgtm/otelcol-config.yaml` in grafana-otel-lgtm to enable TLS on bundled OTel receiver.

**B)** `otlp` gRPC TLS to port 4317 — rename exporter to `otlp`; change endpoint to
   `grafana-otel-lgtm:4317` (gRPC format, no scheme prefix); add `tls:` stanza;
   mount custom `/otel-lgtm/otelcol-config.yaml` for bundled OTel gRPC TLS receiver.
   More efficient (gRPC multiplexing); requires service pipeline exporter name change.

**C)** No TLS for the OTel→Grafana link only — keep `http://grafana-otel-lgtm:4318` as plain HTTP;
   both services are inside the same Docker telemetry-network; only add TLS for the
   PS-ID→OTel receiver path (gRPC 4317, HTTP 4318) and Grafana UI (port 3000).
   Simpler; no bundled OTel config override needed.

**D)**

**Answer**:

**Rationale for question**: Options A and B are essentially equivalent security-wise (both encrypt the
OTel→Grafana link). Option C accepts that internal Docker network traffic between OTel Collector and
Grafana does not need encryption. The bundled OTel inside grafana/otel-lgtm adds complexity because
overriding its config requires a custom YAML to be bind-mounted.

---

## Question 5: grafana/otel-lgtm Image — Continue Using or Plan Migration?

**Background**: The `grafana/otel-lgtm:latest` README states explicitly:
> "This Docker image is intended for development, demo, and testing environments."

It bundles OTel Collector + Prometheus + Tempo + Loki + Pyroscope + Grafana in a **single container**.
This goes against separation-of-concerns best practices and is not intended for production use.
The cryptoutil Suite is being built as a production-grade system. Currently `deployments/shared-telemetry/`
uses this image for the Grafana + telemetry backend stack.

**Note**: For v11 TLS purposes, `grafana/otel-lgtm` works well enough — Grafana TLS via
`GF_SERVER_*` env vars is standard Grafana configuration and the image passes these through.
The concern is about long-term production readiness, not immediate v11 functionality.

**A)** Keep using `grafana/otel-lgtm` — it is adequate for this project's scope and testing
   environments; add a comment in `shared-telemetry/compose.yml` noting it is a dev/test image;
   no migration plan needed.

**B)** Keep `grafana/otel-lgtm` for v11 but formally plan migration to a separate production
   stack as v12 scope — create a v12 planning note in `docs/` documenting the intended migration to
   `grafana/grafana` + `grafana/loki` + `grafana/tempo` + `prom/prometheus` stack; no code change
   in v11.

**C)** Migrate to separate production stack in v11 — replace `grafana/otel-lgtm` with standalone
   services (`grafana/grafana`, `grafana/loki`, `grafana/tempo`, `prom/prometheus`); v11 scope
   expands significantly; the bundled OTel override approach no longer needed (use standalone
   `otel/opentelemetry-collector-contrib` receiver for the Grafana backend).

**D)**

**Answer**:

**Rationale for question**: This is a strategic product decision. If cryptoutil will be deployed
in production environments, using a dev/test-only image for telemetry is a risk. But migrating
to separate services in v11 would substantially expand scope.

---

## Question 6: pki-init in shared-telemetry compose.yml

**Background**: The PS-ID compose template already has a `pki-init` service that runs before the
app starts, generating `/certs`. When a PS-ID compose includes `shared-telemetry/compose.yml`,
the shared-telemetry services need `/certs` to exist.

**Concern**: If someone runs `docker compose -f deployments/shared-telemetry/compose.yml up`
standalone (without a PS-ID compose including it), there is no pki-init to generate certs.
The OTel and Grafana containers would fail to start because `/certs` is empty.

**A)** Add a standalone `pki-init` service directly in `shared-telemetry/compose.yml` using
   the suite binary (image `cryptoutil:local`) with `--domain=cryptoutil`.
   This generates all `ALL-*` shared certs. shared-telemetry can be started independently.
   When a PS-ID compose includes shared-telemetry, two pki-init services exist (one suite-level,
   one PS-ID-level) — both produce the same `ALL-*` certs idempotently; PS-ID-level adds
   PS-ID-specific certs. Docker Compose will treat them as different named services.

**B)** No pki-init in shared-telemetry — rely entirely on the including PS-ID compose to run
   pki-init first. shared-telemetry cannot be started standalone with TLS; it must always be
   included. Document this constraint. Simpler compose structure; no additional service.

**C)** Add a lightweight `pki-init` service in shared-telemetry that uses `golang:1.26-alpine`
   or a minimal Go build to run pki-init from source — avoids needing the full app image;
   but adds build complexity and is unusual.

**D)**

**Answer**:

**Rationale for question**: Option A allows independent telemetry stack startup but creates two
parallel pki-init runs when included. Option B is simpler but makes shared-telemetry unusable
without a PS-ID include. The choice affects how operators use the compose files.

---

## Summary of Answers Needed

| Q# | Decision | Affects Plan Phases |
|----|----------|---------------------|
| Q1 | /certs volume strategy | Phases 2, 3, 4, 5 |
| Q2 | PostgreSQL TLS scope in v11 | Phase 4 (defer or activate) |
| Q3 | OTel mTLS vs one-way TLS | Phase 2 Tasks 2.1/2.2/2.4 |
| Q4 | OTel→Grafana TLS protocol | Phase 2 Task 2.3, Phase 3 Task 3.2 |
| Q5 | grafana/otel-lgtm stance | Phase 3, docs |
| Q6 | pki-init placement in shared-telemetry | Phase 3 Task 3.3 |
