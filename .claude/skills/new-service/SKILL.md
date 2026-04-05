---
name: new-service
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
- Register PS-ID in `internal/apps/tools/cicd_lint/lint_fitness/registry/registry.go`
- Add magic constants to `internal/shared/magic/magic_psids.go`
- Compose.yml MUST have 4 service instances (2 SQLite + 2 PostgreSQL)
- Migration numbers MUST use PS-ID range from `api/cryptosuite-registry/registry.yaml`

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
