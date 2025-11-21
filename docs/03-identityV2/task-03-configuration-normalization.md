# Task 03 – Configuration Inventory and Normalization

## Objective

Create a unified inventory of identity-related configuration, flags, and secrets, then normalize them into versioned templates to eliminate drift across services and environments.

## Historical Context

- Configuration changes landed piecemeal during the original implementation and subsequent fixes (`dc68619`, `5c04e44`), resulting in inconsistent defaults.
- Docker Compose overrides and Go CLI flags diverged from documented expectations in `docs/identity/identity_master.md`.

## Scope

- Catalogue all configuration touchpoints (Go structs, YAML files, Docker Compose, workflow inputs).
- Define canonical templates under `configs/identity/` that distinguish development, test, and production personas.
- Document secrets sourcing (files vs. CLI flags) while adhering to existing security instructions (prefer file-based secrets).

## Deliverables

- Versioned YAML templates: `configs/identity/<persona>.yml` with clearly annotated defaults.
- Configuration diff report summarizing changes from current state to normalized templates.
- Automated fixtures in `internal/identity/config/testdata` for use in unit and integration tests.

## Validation

- Execute `go test ./internal/identity/config/...` using the new fixtures.
- Run Docker Compose smoke tests to verify compatibility with normalized templates.
- Confirm alignment with security guidance (`.github/instructions/01-05.security.instructions.md`).

## Dependencies

- Task 01 baseline highlights drift candidates; Task 02 requirements provide necessary configuration hooks.
- Downstream tasks (06, 07, 08, 10, 18) depend on the standardized templates.

## Risks & Notes

- Ensure magic values extracted during normalization live in the approved `magic_*.go` files.
- Communicate configuration changes early to avoid blocking concurrent remediation work.
