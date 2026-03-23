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
# Copy entire skeleton directory
cp -r internal/apps/skeleton/template internal/apps/PRODUCT/SERVICE

# Copy skeleton cmd/
cp cmd/skeleton-template/main.go cmd/PRODUCT-SERVICE/main.go
```

### Step 2: Rename identifiers

```bash
# Replace all skeleton-template identifiers
find internal/apps/PRODUCT/SERVICE cmd/PRODUCT-SERVICE -type f -name "*.go" | xargs sed -i 's/skeleton-template/PRODUCT-SERVICE/g'
find internal/apps/PRODUCT/SERVICE cmd/PRODUCT-SERVICE -type f -name "*.go" | xargs sed -i 's/skeletonTemplate/PRODUCTService/g'
```

### Step 3: Assign port range

- Service: `PRODUCT-SERVICE` → host ports 8XX0-8XX9 (see service catalog)
- Public: `0.0.0.0:8080` (container) / `127.0.0.1:8XX0` (dev)
- Admin: `127.0.0.1:9090` (container) / `127.0.0.1:8XX1` (dev)
- PostgreSQL: `localhost:5432X` (container) / `localhost:5432X` (dev)

### Step 4: Create domain migrations

```bash
# Start from 2001 (template uses 1001-1999)
touch internal/apps/PRODUCT/SERVICE/repository/migrations/2001_init.up.sql
touch internal/apps/PRODUCT/SERVICE/repository/migrations/2001_init.down.sql
```

### Step 5: Add config files

```bash
cp configs/skeleton/template/config-development.yml configs/PRODUCT/SERVICE/config-development.yml
cp configs/skeleton/template/config-production.yml  configs/PRODUCT/SERVICE/config-production.yml
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
go build ./cmd/PRODUCT-SERVICE/...
go test ./internal/apps/PRODUCT/SERVICE/...
go run ./cmd/cicd-lint lint-deployments
```

### Step 9: Update Documentation

```bash
# Update ARCHITECTURE.md Section 3.4 Service Catalog table
# Add row: | PRODUCT | SERVICE | PRODUCT-SERVICE | HOST_PORT_RANGE | 0.0.0.0:8080 | 127.0.0.1:9090 |
```

- Update service catalog in `docs/ARCHITECTURE.md` Section 3.4 Port Assignments & Networking
- Update service catalog table in `.github/instructions/02-01.architecture.instructions.md`
- Update `README.md` if it lists services

## Port Assignment Rules

- **Service deployment**: PORT (8000–8999 range)
- **Product deployment**: PORT + 10000 (18000–18999)
- **Suite deployment**: PORT + 20000 (28000–28999)

## References

Read [ARCHITECTURE.md Section 3.4 Port Assignments](../../../docs/ARCHITECTURE.md#34-port-assignments--networking) for port catalog — select the next available port range from this table when assigning host ports for the new service.
Read [ARCHITECTURE.md Section 5.1 Service Framework Pattern](../../../docs/ARCHITECTURE.md#51-service-framework-pattern) for framework components — validate that all required components (dual HTTPS, health checks, migrations, telemetry) are present in the new service.
Read [ARCHITECTURE.md Section 5.2 Service Builder Pattern](../../../docs/ARCHITECTURE.md#52-service-builder-pattern) for builder usage — follow the builder registration flow and `ServiceResources` pattern exactly as specified.
