# Orphaned Config Files

Config files in this directory have no corresponding deployment in `deployments/`.
They were moved here during the deployment/config restructuring to maintain a clean
mirror between `deployments/` and `configs/`.

## Contents

### observability/

- `prometheus/adaptive-auth-alerts.yml` - Prometheus alert rules for adaptive authentication
- `grafana/adaptive-auth-dashboard.json` - Grafana dashboard for adaptive auth monitoring
- **Reason**: `deployments/shared-telemetry/` has its own embedded configs; these are unreferenced

### template/

- `config-pg-1.yml` - Template PostgreSQL instance 1 config
- `config-pg-2.yml` - Template PostgreSQL instance 2 config
- `config-sqlite.yml` - Template SQLite config
- **Reason**: `deployments/template/config/` contains authoritative placeholder templates; these are duplicates with hardcoded values

### test/

- `config.yml` - Test config (uses http protocol and inline credentials; intentionally invalid)
- `dast-simple-unseal-secret-value-1.secret` - DAST test secret file
- **Reason**: No `test` deployment exists; these are test infrastructure files

## Restoration

To restore a file, move it back to `configs/<product>/` and ensure a corresponding
deployment exists in `deployments/`.

## Validation

Run `go run ./cmd/cicd lint-deployments validate-mirror` to verify the mirror
structure. Orphaned directories appear as warnings (not errors).
