---
name: new-service
disable-model-invocation: true
description: "Guide service creation from skeleton-template including copy, rename, port assignment, migration, and test setup. Use when creating a new cryptoutil service to cover all steps from copying the skeleton source to CI/CD registration."
argument-hint: "[PRODUCT SERVICE port-prefix]"
---

Guide service creation from skeleton-template: copy, rename, register, migrate, test.

## Purpose

Use when creating a new cryptoutil service from the template. Covers all steps
from cloning the skeleton to registering the service in validation and documentation.

Use `migration-create` for the migration file details, `openapi-codegen` for API scaffolding, and `copilot-customization` for new repo-local agent or skill artifacts created during the service rollout.

## Key Rules

- ALWAYS copy from `skeleton-template` — NEVER create from scratch
- Port block: assign from `api/cryptosuite-registry/registry.yaml` and the service catalog in `docs/ENG-HANDBOOK.md`
- Register PS-ID in `internal/apps-tools/cicd_lint/lint_fitness/registry/registry.go`
- Add magic constants to `internal/shared/magic/magic_psids.go`
- Compose.yml MUST have 4 service instances (2 SQLite + 2 PostgreSQL)
- Migration numbers MUST use PS-ID range from `api/cryptosuite-registry/registry.yaml`
- TLS client policy: ALWAYS add `server-*-tls-client-policy` alongside any `server-*-tls-ca-file` in deployment overlays
- Prefer repo-aware file operations and targeted edits; do not rely on Bash-only copy or mass-replace snippets

## Service Catalog

| Product | Service ID | Host Port Range |
|---------|-----------|----------------|
| SM | sm-kms | 8000-8099 |
| JOSE | jose-ja | 8200-8299 |
| PKI | pki-ca | 8300-8399 |
| Identity | identity-authz | 8400-8499 |
| ... | ... | ... |
| Skeleton | skeleton-template | 8900-8999 |

## Step-by-Step Process

### Step 1: Clone the skeleton surfaces

- Copy `internal/apps/skeleton-template/` to the new PS-ID location under `internal/apps/`
- Copy `cmd/skeleton-template/` to the new PS-ID entry-point directory under `cmd/`
- Copy the deployment and config directories from `configs/skeleton-template/` and `deployments/skeleton-template/`

### Step 2: Rename identifiers

- Replace `skeleton-template` and the template-specific Go identifiers with the new PS-ID consistently across copied files
- Re-check usage strings, generated-code config, module-local README files, and deployment filenames after the rename

### Step 3: Assign port range

- Reserve the next available service host-port block from the registry and handbook catalog
- Keep container bindings on `0.0.0.0:8080` (public) and `127.0.0.1:9090` (admin)
- Keep deployment formulas aligned across service, product, and suite overlays

### Step 4: Create domain migrations

- Start domain migrations at the service range defined in `api/cryptosuite-registry/registry.yaml`
- Create paired `.up.sql` and `.down.sql` files and register them via `WithDomainMigrations`
- Use `migration-create` if the main task in front of you is the migration content itself

### Step 5: Add config files

- Rename the standalone service config in `configs/PS-ID/`
- Rename the deployment overlay files in `deployments/PS-ID/config/`
- Update PS-ID-specific names, port values, OTLP service names, and database settings in every variant file

### Step 6: Add Docker Compose deployment

- Update the copied compose deployment with the new service name, secrets, cert mount references, and four-instance topology
- Keep the `/certs:/certs:ro` bind mount and admin healthcheck conventions intact

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
- Run `go run ./cmd/cicd-lint lint-deployments` to validate
- Run `go run ./cmd/cicd-lint lint-fitness` when registry or template-instantiated files changed

### Step 8: Test

```bash
go build ./cmd/PS-ID/...
go test ./internal/apps/PS-ID/...
go run ./cmd/cicd-lint lint-deployments
```

### Step 9: Update Documentation

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
