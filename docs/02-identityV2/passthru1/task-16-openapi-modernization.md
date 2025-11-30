# Task 16 – OpenAPI 3.0 Spec Modernization

## ⚠️ NOTE: This task has been **REPLACED** by Task 10.7 (OpenAPI Synchronization) in the refactored plan

**Original Task 16 content retained for historical reference only.**

**See instead**: `docs/identityV2/task-10.7-openapi-sync.md` - moved earlier in sequence to document working APIs before feature additions.

---

## Task Reflection (Historical)

### What Went Well

- ✅ **OpenAPI Infrastructure Exists**: `api/` directory has generation configs and tooling
- ✅ **Generation Workflow**: `go generate ./api/...` pattern established

### At Risk Items

- ⚠️ **Spec Drift**: Positioned too late (Task 16), specs don't reflect implemented endpoints
- ⚠️ **Feature Bloat**: Adding features (Tasks 11-15) before documenting existing APIs compounds drift

### Could Be Improved

- **Task Ordering**: Should document APIs immediately after implementation (Task 10.7, not Task 16)
- **Continuous Validation**: No automated checks that specs match server behavior

### Refactor Decision

- **MOVED TO Task 10.7**: OpenAPI work repositioned after Task 10.6 (Unified CLI)
- **RATIONALE**: Document working foundation before adding features
- **STATUS**: Original Task 16 content superseded by comprehensive Task 10.7 specification

---

## Objective (Historical - See Task 10.7)

Modernize the identity OpenAPI specifications and generation workflows to match the rebuilt services, eliminating drift and ensuring client/server compatibility.

## Historical Context

- Commit `01e69f4` updated specs, but subsequent service changes and bug fixes were not reflected consistently.
- Build scripts and generation configs require validation after toolchain updates.

## Scope

- Review and update `api/openapi_spec_*.yaml` and associated generator configs.
- Regenerate clients/servers, ensuring casing and import fixes align with `f217bf7` and later adjustments.
- Add linting or CI checks to prevent future spec drift.

## Deliverables

- Updated OpenAPI specs with comprehensive component and path coverage.
- Refreshed generation scripts (Go, TypeScript, etc.) with documented usage.
- Linting/compliance automation (e.g., `oapi-codegen validate`) integrated into CI.

## Validation

- Run `go generate ./api/...` and associated validation commands.
- Execute contract tests against generated code to confirm compatibility.

## Dependencies

- Requires finalized APIs from Tasks 06–10.
- Coordination with documentation tasks to update published API references.

## Risks & Notes

- Ensure generator updates maintain compatibility with downstream consumers; version and document breaking changes carefully.
