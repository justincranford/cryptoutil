---
name: new-service
disable-model-invocation: true
description: "Guide service creation from skeleton-template including copy, rename, port assignment, migration, and test setup. Use when creating a new cryptoutil service to cover all steps from copying the skeleton source to CI/CD registration."
argument-hint: "[PRODUCT SERVICE port-prefix]"
---

Guide service creation from skeleton-template: copy, rename, register, migrate, test.

## Purpose

Use when creating a new cryptoutil service from the template. Covers all steps
from copying the skeleton to registering with CI/CD.

## Key Rules

- ALWAYS copy from `skeleton-template` — NEVER create from scratch
- Port block: assign from registry.yaml (4 ports per PS-ID: sqlite-1=+0, sqlite-2=+1, postgresql-1=+2, postgresql-2=+3)
- Register PS-ID in `internal/apps-tools/cicd_lint/lint_fitness/registry/registry.go`
- Add magic constants to `internal/shared/magic/magic_psids.go`
- Compose.yml MUST have 4 service instances (2 SQLite + 2 PostgreSQL)
- Migration numbers MUST use PS-ID range from `api/cryptosuite-registry/registry.yaml`
- TLS client policy: ALWAYS add `server-*-tls-client-policy` alongside any `server-*-tls-ca-file` in deployment overlays

## Service Catalog

| Product | Service ID | Host Port Range |
|---------|-----------|----------------|
| SM | sm-kms | 8000-8099 |
| PKI | pki-ca | 8100-8199 |
| Identity | identity-authz | 8200-8299 |
| ... | ... | ... |
| Skeleton | skeleton-template | 8900-8999 |

## Step-by-Step Process

### Step 1: Copy skeleton-template

```bash
# Copy entire skeleton app directory and cmd entry point
cp -r internal/apps/skeleton-template internal/apps/PS-ID

# Copy skeleton cmd/
cp cmd/skeleton-template/main.go cmd/PS-ID/main.go
```

### Step 2: Rename identifiers

```bash
# Replace all skeleton-template identifiers
find internal/apps/PS-ID cmd/PS-ID -type f -name "*.go" | xargs sed -i 's/skeleton-template/PS-ID/g'
find internal/apps/PS-ID cmd/PS-ID -type f -name "*.go" | xargs sed -i 's/skeletonTemplate/PSIDCamelCase/g'
```

### Step 3: Assign port range

- Service: `PRODUCT-SERVICE` → host ports 8XX0-8XX9 (see service catalog)
- Public: `0.0.0.0:8080` (container) / `127.0.0.1:8XX0` (dev)
- Admin: `127.0.0.1:9090` (container) / `127.0.0.1:8XX1` (dev)
- PostgreSQL: `localhost:5432X` (container) / `localhost:5432X` (dev)

### Step 4: Create domain migrations

```bash
# Start from 2001 (template uses 1001-1999)
touch internal/apps/PS-ID/repository/migrations/2001_init.up.sql
touch internal/apps/PS-ID/repository/migrations/2001_init.down.sql
```

### Step 5: Add config files

```bash
# Copy domain config from skeleton-template and rename
cp -r configs/skeleton-template configs/PS-ID
mv configs/PS-ID/skeleton-template.yml configs/PS-ID/PS-ID.yml

# Copy deployment variant configs and rename
cp -r deployments/skeleton-template/config deployments/PS-ID/config
for f in deployments/PS-ID/config/skeleton-template-*.yml; do
  mv "$f" "${f/skeleton-template/PS-ID}"
done
# Update port numbers, service name, database config
```

### Step 6: Add Docker Compose deployment

```bash
cp -r deployments/skeleton-template deployments/PRODUCT-SERVICE
# Update port bindings, service name, secrets references
```

### TLS Configuration (Two-Axis Model)

Cryptoutil uses a two-axis TLS model. Understand both axes before editing deployment configs.

**Axis 1 — TLSProvisionMode** (`auto` / `mixed` / `static`): controls certificate sourcing.
This is **automatic** — no manual configuration needed for new services:
- `auto`: no secrets provided → framework generates ephemeral certs in memory (local dev, tests)
- `mixed`: issuing CA key provided → framework generates a leaf cert at startup
- `static`: cert chain + private key provided → framework uses the pre-generated cert as-is

**Axis 2 — TLSClientPolicy** (`none` / `request` / `require-any` / `verify-if-given` / `require-and-verify`): controls runtime client-certificate enforcement.
This **must be set explicitly** in deployment overlay configs:
- Default (framework config): `none` — no client certificates requested
- Skeleton-template overlays: `require-and-verify` for both `server-public-tls-client-policy`
  and `server-admin-tls-client-policy` — already set correctly when you copy them

**Rule when copying skeleton-template overlays (Steps 5–6)**:
The `server-*-tls-client-policy` keys come with the copy — do not remove them.

**Rule when adding new `server-*-tls-ca-file` keys**:
ALWAYS add the corresponding `server-*-tls-client-policy` key alongside it.
The `config-tls-ca-policy-coupling` fitness linter enforces this and blocks commit.

Example (from any overlay in `deployments/skeleton-template/config/`):
```yaml
server-admin-tls-ca-file: /certs/.../truststore/issuing-ca.crt
server-admin-tls-client-policy: require-and-verify   # MANDATORY when ca-file present

server-public-tls-ca-file: /certs/.../truststore/issuing-ca.crt
server-public-tls-client-policy: require-and-verify  # MANDATORY when ca-file present
```

For a transitional rollout where some clients don't yet present certificates, use
`verify-if-given` until all clients are migrated, then switch to `require-and-verify`.

### Step 7: Register in CI/CD

- Add service to `.github/workflows/ci-*.yml` matrix
- Add service to `docker-compose.yml` (root-level if suite)
- Run `go run ./cmd/cicd-lint lint-deployments` to validate

### Step 8: Test

```bash
go build ./cmd/PS-ID/...
go test ./internal/apps/PS-ID/...
go run ./cmd/cicd-lint lint-deployments
```

### Step 9: Update Documentation

```bash
# Update ENG-HANDBOOK.md Section 3.4 Service Catalog table
# Add row: | PRODUCT | SERVICE | PRODUCT-SERVICE | HOST_PORT_RANGE | 0.0.0.0:8080 | 127.0.0.1:9090 |
```

- Update service catalog in `docs/ENG-HANDBOOK.md` Section 3.4 Port Assignments & Networking
- Update service catalog table in `.github/instructions/02-01.architecture.instructions.md`
- Update `README.md` if it lists services

## Port Assignment Rules

- **Service deployment**: PORT (8000–8999 range)
- **Product deployment**: PORT + 10000 (18000–18999)
- **Suite deployment**: PORT + 20000 (28000–28999)

## References

Read [ENG-HANDBOOK.md Section 3.4 Port Assignments](../../../docs/ENG-HANDBOOK.md#34-port-assignments--networking) for port catalog — select the next available port range from this table when assigning host ports for the new service.
Read [ENG-HANDBOOK.md Section 5.1 Service Framework Pattern](../../../docs/ENG-HANDBOOK.md#51-service-framework-pattern) for framework components — validate that all required components (dual HTTPS, health checks, migrations, telemetry) are present in the new service.
Read [ENG-HANDBOOK.md Section 5.2 Service Builder Pattern](../../../docs/ENG-HANDBOOK.md#52-service-builder-pattern) for builder usage — follow the builder registration flow and `ServiceResources` pattern exactly as specified.
Read [ENG-HANDBOOK.md Section 5.6 PS-ID Entry Point Patterns](../../../docs/ENG-HANDBOOK.md#56-ps-id-entry-point-patterns) for `lifecycle.RunService()` (signal handling) and `BuildUsage*()` (usage strings) — the skeleton-template already uses these; ensure copied entry point is NOT modified to use raw `signal.Notify` or inline usage strings.
