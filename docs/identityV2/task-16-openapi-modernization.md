# Task 16 – OpenAPI 3.0 Spec Modernization

## Objective

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
