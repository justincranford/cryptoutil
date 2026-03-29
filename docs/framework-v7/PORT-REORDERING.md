# PORT-REORDERING.md — Port Range Reassignment Plan

Port ranges should match canonical service order so the numeric sequence
reflects the migration priority and product grouping.

## Current vs Target Port Assignments

| # | Service | Current Host Range | Target Host Range | Change |
|---|---------|-------------------|-------------------|--------|
| 1 | sm-kms | 8000-8099 | 8000-8099 | none |
| 2 | sm-im | 8700-8799 | 8100-8199 | move |
| 3 | jose-ja | 8800-8899 | 8200-8299 | move |
| 4 | pki-ca | 8100-8199 | 8300-8399 | move |
| 5 | identity-authz | 8200-8299 | 8400-8499 | move |
| 6 | identity-idp | 8300-8399 | 8500-8599 | move |
| 7 | identity-rs | 8400-8499 | 8600-8699 | move |
| 8 | identity-rp | 8500-8599 | 8700-8799 | move |
| 9 | identity-spa | 8600-8699 | 8800-8899 | move |
| 10 | skeleton-template | 8900-8999 | 8900-8999 | none |

Derived port tiers (per architecture: PRODUCT = SERVICE + 10000, SUITE = SERVICE + 20000):

| # | Service | SERVICE | PRODUCT | SUITE |
|---|---------|---------|---------|-------|
| 1 | sm-kms | 8000 | 18000 | 28000 |
| 2 | sm-im | 8100 | 18100 | 28100 |
| 3 | jose-ja | 8200 | 18200 | 28200 |
| 4 | pki-ca | 8300 | 18300 | 28300 |
| 5 | identity-authz | 8400 | 18400 | 28400 |
| 6 | identity-idp | 8500 | 18500 | 28500 |
| 7 | identity-rs | 8600 | 18600 | 28600 |
| 8 | identity-rp | 8700 | 18700 | 28700 |
| 9 | identity-spa | 8800 | 18800 | 28800 |
| 10 | skeleton-template | 8900 | 18900 | 28900 |

PostgreSQL ports follow the same canonical order (54320-54329):

| # | Service | Current PG Port | Target PG Port |
|---|---------|----------------|----------------|
| 1 | sm-kms | 54323 | 54320 |
| 2 | sm-im | 54322 | 54321 |
| 3 | jose-ja | 54321 | 54322 |
| 4 | pki-ca | 54320 | 54323 |
| 5 | identity-authz | 54324 | 54324 |
| 6 | identity-idp | 54325 | 54325 |
| 7 | identity-rs | 54326 | 54326 |
| 8 | identity-rp | 54327 | 54327 |
| 9 | identity-spa | 54328 | 54328 |
| 10 | skeleton-template | 54329 | 54329 |

## Steps

1. **Update `internal/shared/magic/magic_ports.go`** (or equivalent magic constants) — change port
   constants to target values. Ensure both SERVICE, PRODUCT, and SUITE tiers are updated.

2. **Update `internal/apps/tools/cicd_lint/lint_ports/common/common.go`** — change `ServicePorts`
   map values to target port ranges. Update any range validation constants.

3. **Update all `deployments/*/compose.yml`** files — change host port mappings for each service,
   product, and suite compose file. Each deployment tier has its own compose.yml.

4. **Update all `configs/*/config-*.yml`** files — change any hardcoded port references in service
   framework configs (bind-public-port, etc.). Container ports (8080/9090) stay unchanged.

5. **Update `docs/ARCHITECTURE.md`** — update Section 3.4 Port Assignments table, Section 5.3
   Dual HTTPS Endpoint Pattern examples, and any other port references.

6. **Update `.github/instructions/02-01.architecture.instructions.md`** — update Service Catalog
   table port assignments and PostgreSQL ports list.

7. **Update `docs/framework-v7/target-structure.md`** — update any port references in the
   target structure documentation.

8. **Update test fixtures and E2E configs** — search for old port numbers in test files,
   E2E compose files, and testdata. Replace with target port numbers.

9. **Run full validation suite**:
   - `go build ./...`
   - `go test ./... -shuffle=on`
   - `go run ./cmd/cicd-lint lint-ports` — validates port range enforcement
   - `go run ./cmd/cicd-lint lint-deployments` — validates deployment structure
   - `go run ./cmd/cicd-lint lint-fitness` — validates architecture fitness

10. **Verify no port conflicts** — ensure no two services share the same host port range
    after reassignment. The `lint-ports` validator should catch conflicts automatically.

## Risk Mitigation

- Ports 8000 (sm-kms) and 8900 (skeleton-template) do not move.
- All other 8 services shift to new ranges — do all changes atomically in one commit.
- Container-internal ports (8080 public, 9090 admin) are unchanged.
- PostgreSQL container port (5432) is unchanged; only host-mapped ports shift.
- Docker Desktop users may need to clear cached port bindings after the change.
