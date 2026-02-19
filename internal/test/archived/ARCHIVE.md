# Archived Test Framework

## E2E Legacy Framework (e2e-legacy/)

**Archived**: 2025-12-24
**Reason**: Migrated to service template pattern
**Replacement**: `internal/apps/template/service/testing/e2e_helpers/` (service-level) + `internal/apps/template/service/testing/e2e_infra/` (Docker Compose orchestration)

### Migration Summary

**Fully Migrated**:
- `docker_health.go` → `internal/apps/template/service/testing/e2e_infra/docker_health.go` (3-use-case health checking with ServiceAndJob struct)

**Partially Migrated**:
- `docker_utils.go` → Simplified as `internal/apps/template/service/testing/e2e_infra/compose_manager.go` (ComposeManager, lighter design)
- `http_utils.go` → Similar functionality in `internal/apps/template/service/testing/e2e_helpers/http_helpers.go`

**Not Migrated (Intentionally Archived)**:
- `log_utils.go` - Elaborate dual-output logging replaced by simpler testing.T.Log()
- `assertions.go` - Service-specific assertions (not reusable)
- `fixtures.go` - Legacy test fixture pattern with elaborate infrastructure dependencies
- `infrastructure.go` - Heavy InfrastructureManager replaced by simpler ComposeManager
- `test_suite.go` - testify/suite pattern replaced by TestMain pattern
- `*_workflow_test.go` - Legacy compose stack tests, not applicable to new services

### Design Improvements

**Old Framework**: Elaborate infrastructure with Logger, Asserter, InfrastructureManager, testify/suite
**New Framework**: Simpler TestMain pattern, testing.T.Log(), direct ComposeManager usage

**Two Directory Strategy**:
1. `internal/apps/template/service/testing/e2e_helpers/` - Service-level testing helpers (in-process, no Docker)
2. `internal/apps/template/service/testing/e2e_infra/` - Docker Compose orchestration (docker_health.go, compose_manager.go)

### References

- Migration Analysis: `docs/e2e-migration-analysis.md`
- Architecture Documentation: `docs/ARCHITECTURE.md` Section 10.4 (E2E Testing Strategy)
- Healthcheck Patterns: `docs/ARCHITECTURE.md` Section 10.4.2 (3 use cases documented)
