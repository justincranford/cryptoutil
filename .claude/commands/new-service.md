---
name: new-service
description: "Guide service creation from skeleton-template including copy, rename, port assignment, migration, and test setup. Use when creating a new cryptoutil service to cover all steps from copying the skeleton source to CI/CD registration."
argument-hint: "[PRODUCT SERVICE port-prefix]"
---

Guide for creating a new PS-ID service from skeleton-template.

**Full Copilot original**: [.github/skills/new-service/SKILL.md](.github/skills/new-service/SKILL.md)

Provide: PS-ID (e.g., `sm-xyz`), product (e.g., `sm`), service name (e.g., `xyz`), base port (from registry.yaml).

## Key Rules

- ALWAYS copy from `skeleton-template` — NEVER create from scratch
- Port block: assign from registry.yaml (4 ports per PS-ID: sqlite-1=+0, sqlite-2=+1, postgresql-1=+2, postgresql-2=+3)
- Register PS-ID in `internal/apps/tools/cicd_lint/lint_fitness/registry/registry.go`
- Add magic constants to `internal/shared/magic/magic_psids.go`
- Compose.yml MUST have 4 service instances (2 SQLite + 2 PostgreSQL)
- Migration numbers MUST use PS-ID range from `api/cryptosuite-registry/registry.yaml`

## 9-Step Creation Process

### Step 1: Copy skeleton-template

```bash
cp -r internal/apps/skeleton-template internal/apps/{ps-id}
cp -r cmd/skeleton-template cmd/{ps-id}
```

### Step 2: Rename all occurrences

Replace in all files:
- `skeleton-template` → `{ps-id}`
- `skeleton_template` → `{ps_id}` (underscores for SQL, Go)
- `SkeletonTemplate` → `{PsId}` (PascalCase)
- `Template` → `{ServiceName}` (Go types)
- Port `8900` → `{base_port}`

### Step 3: Register in registry.yaml

Add entry to `api/cryptosuite-registry/registry.yaml`:
```yaml
- ps_id: {ps-id}
  product: {product}
  service: {service}
  display_name: "{Display Name}"
  internal_apps_dir: {ps-id}/
  base_port: {base_port}
  migration_range_start: {start}
  migration_range_end: {end}
  api_resources:
    - /resources
```

### Step 4: Assign migration range

Allocate next available 1000-number range from registry.yaml.

### Step 5: Create initial migration

```bash
# Name file using assigned range
touch internal/apps/{ps-id}/server/repository/migrations/{NNNN}_init.up.sql
touch internal/apps/{ps-id}/server/repository/migrations/{NNNN}_init.down.sql
```

### Step 6: Create Docker Compose files

Copy from `deployments/skeleton-template/` and update service names, ports, image names.

### Step 7: Register in CI/CD

Add PS-ID to:
- `.github/workflows/ci-quality.yml` (matrix)
- `.github/workflows/ci-coverage.yml` (matrix)
- `.github/workflows/ci-e2e.yml` (service startup)

### Step 8: Verify build and tests

```bash
go build ./cmd/{ps-id}/...
go test ./internal/apps/{ps-id}/...
```

### Step 9: Update docs

- Add entry to `api/cryptosuite-registry/registry.yaml` (already done in step 3)
- Reference new PS-ID in `docs/ARCHITECTURE.md` §3 product suite table
