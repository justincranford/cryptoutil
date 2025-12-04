# Task 04 – Identity Package Dependency Audit

## Objective

Enforce clean architecture boundaries by detecting and eliminating hidden dependencies between `internal/identity/**` and other domains (KMS, CA, etc.), strengthening maintainability.

## Historical Context

- The legacy plan warned about cross-domain coupling, but commits after Task 07 introduced utilities without formal boundary checks.
- No dedicated static analysis currently prevents regressions.

## Scope

- Inventory direct and indirect imports used by identity packages.
- Introduce lint/staticcheck rules (for example, `importas`, `depguard`, or custom analyzers) that enforce boundaries.
- Produce a visual dependency graph to help engineers reason about cross-package interactions.

## Deliverables

- `docs/identityV2/dependency-graph.md` with diagrams (plantuml or mermaid) and narrative commentary.
- Updates to `.golangci.yml` and supporting tooling to enforce newly defined rules.
- Documentation of allowed exceptions (if any) with justification.

## Validation

- Run `golangci-lint run` to confirm the new rules operate without false positives.
- Peer review the dependency graph for accuracy and readability.

## Dependencies

- Requires outputs from Task 01 (baseline) and Task 02 (requirements) to understand functional boundaries.
- Coordinate with Task 03 to ensure configuration utilities remain accessible without reintroducing coupling.

## Risks & Notes

- Overly strict lint rules can block legitimate cross-cutting concerns; iterate with small allowlists where necessary.
