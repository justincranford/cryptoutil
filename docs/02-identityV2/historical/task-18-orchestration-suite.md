# Task 18 – Docker Compose Orchestration Suite
## ⚠️ NOTE: Core orchestration functionality **PARTIALLY MOVED** to Task 10.6 (Unified CLI) in refactored plan.

**Original Task 18 content retained. This task now focuses on advanced orchestration patterns (scaling, templating, Docker profiles).**

**See also**: `docs/identityV2/task-10.6-unified-cli.md` - unified CLI provides foundation for orchestration.

---

## Task Reflection

### What Went Well

- ✅ **Task 10.6 Unified CLI**: One-liner bootstrap (`./identity start --profile demo`) achieved
- ✅ **Existing Compose Infrastructure**: `identity-compose.yml` provides Docker orchestration foundation
- ✅ **Task 10.7 OpenAPI**: Well-documented APIs simplify client integration

### At Risk Items

- ⚠️ **Scaling Templates Missing**: No Nx/Mx/Xx patterns for multiple service instances
- ⚠️ **Secret Handling Incomplete**: Current Compose files only partially follow Docker secrets best practices
- ⚠️ **Troubleshooting Gaps**: Limited documentation for common Docker issues (networking, health checks)

### Could Be Improved

- **Profile Variety**: Need demo, development, CI, production-like Compose configurations
- **Health Check Integration**: Better integration with `./identity health` command for Docker mode
- **Developer Experience**: Simplified workflows for common scenarios (single-service testing, debugging)

### Dependencies and Blockers

- **Dependency on Task 10.6**: Unified CLI provides orchestration foundation
- **Dependency on Task 10**: Integration tests validate Docker Compose health checks
- **Enables Task 18**: E2E testing requires deterministic Docker orchestration

---

## Objective

Deliver deterministic Docker Compose orchestration bundles and supporting tooling that reflect the final identity architecture and simplify local/CI operations.

## Historical Context

- Commit `f3e1f34` added infrastructure but relied on manual steps and lacked scaling templates (Nx/Mx/Xx/Yx/Zx patterns).
- Security instructions mandate consistent secret handling that current Compose files only partially satisfy.

## Scope

- Build templated Compose bundles covering development, demo, and CI scenarios with scalable service counts.
- Provide orchestration CLI enhancements (Go-based) for start/stop, health checks, and diagnostics.
- Document troubleshooting workflows, including interaction with the OTEL collector and Grafana stack.

## Deliverables

- `deployments/compose/identity-demo.yml` (and variants) with relative paths and secret reuse per `.github/instructions/02-02.docker.instructions.md`.
- Updated CLI tooling or scripts under `cmd/workflow` or `scripts/` for orchestration control.
- Runbooks or quick-start guides for developers and QA.

## Validation

- Automated smoke tests (`go test ./internal/identity/demo/...` or equivalent) plus manual verification of scaling scenarios.
- Ensure Compose health checks pass and respect IPv4 loopback requirements.

## Dependencies

- Builds upon integration work (Task 10) and configuration normalization (Task 03).
- Supports testing fabric (Task 19) and final verification (Task 20).

## Risks & Notes

- Watch for cross-platform path incompatibilities; rely on relative paths only.
- Avoid modifying shared Docker secrets as mandated by project instructions.
