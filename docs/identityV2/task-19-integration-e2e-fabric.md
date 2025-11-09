# Task 19 – Integration and E2E Testing Fabric

## Objective

Establish a comprehensive integration and end-to-end testing fabric that continuously validates identity flows across services, environments, and orchestration modes.

## Historical Context

- Commit `f3e1f34` introduced partial testing frameworks, while `5c04e44` improved mock orchestration; however, coverage remains inconsistent.
- CI workflows documented in `d91791b` reference tests that no longer exist or have drifted.

## Scope

- Build or enhance Go-based integration/E2E test suites covering critical identity scenarios (auth flows, SPA journeys, MFA, adaptive decisions).
- Integrate tests with `cmd/workflow` for deterministic execution in local and CI environments.
- Produce coverage dashboards and artefacts stored under `workflow-reports/`.

## Deliverables

- New or updated Go test packages (`internal/identity/...`) with clear naming and tagging (`-tags=e2e`).
- CI workflow updates (or documentation) detailing how to run the suites and collect artefacts.
- Coverage dashboards (HTML/JSON) archived for review.

## Validation

- Execute `go test ./internal/identity/... -tags=e2e -run Integration` (or equivalent) with documented success criteria.
- Dry run workflows using `go run ./cmd/workflow -workflows=e2e` to ensure automation parity.

## Dependencies

- Depends on orchestration suite (Task 18) and upstream remediation tasks that stabilize individual components.
- Provides foundational evidence for final verification (Task 20).

## Risks & Notes

- Avoid introducing flaky tests; leverage deterministic fixtures and mock services.
- Ensure artefact upload patterns align with CI/CD instructions in `.github/instructions/02-01.github.instructions.md`.
