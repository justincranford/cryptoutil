# Repository Refactor and CLI Enablement Plan

## Purpose

- Restructure the repository to support modular service groups (kms, identity, ca, and beyond).
- Deliver consistent CLI tooling for every service API and administrative surface.
- Align documentation, workflows, and automation with the new structure without breaking compatibility.

## Guiding Principles

- Preserve build reproducibility; all existing workflows must keep passing during incremental refactors.
- Provide migration guidance for each structural change (imports, configs, docs).
- Prefer automation for repetitive updates (code generation, import rewrites, documentation sync).
- Maintain clear boundaries between service domains while allowing shared utilities where appropriate.

## Task Breakdown

### Task 1: Service Group Taxonomy and Roadmap

- Objective: Define mandatory groups (kms, identity, ca) plus 10 additional repo-driven groups and 30 adjacent-market groups.
- Actions: Analyze current directories (`internal`, `cmd`, `api`, `scripts`, `docs`) to draft taxonomy and adjacency list.
- Deliverables: `docs/refactor/service-groups.md` with rationales and review checklist.
- Validation: Stakeholder approval; cross-reference with LONGER-TERM-IDEAS requirements.

### Task 2: Repository Inventory and Coupling Analysis

- Objective: Map current package dependencies to identify cross-group coupling risks.
- Actions: Use `go list` metadata, custom scripts, and documentation review.
- Deliverables: Dependency graph, coupling risk log, proposed mitigation strategies.
- Validation: Updated documentation; static analysis confirming accuracy.

### Task 3: Group Directory Blueprint

- Objective: Design target directory layout for groups, ensuring import stability.
- Actions: Draft layout in `docs/refactor/blueprint.md`, including migration steps.
- Deliverables: Blueprint doc, migration sequence diagrams.
- Validation: Architecture review sign-off.

### Task 4: Import Alias Policy Update

- Objective: Extend `.golangci.yml` `importas` rules to reflect new groups.
- Actions: Update instructions and config; add lint tests.
- Deliverables: Proposed alias map, unit tests verifying enforcement.
- Validation: `golangci-lint run` passes with new rules.

### Task 5: CLI Strategy Framework

- Objective: Define CLI patterns for service APIs and administration.
- Actions: Document command structure, flag conventions, JSON/YAML output expectations.
- Deliverables: `docs/refactor/cli-strategy.md`, shared CLI helper package plan.
- Validation: CLI framework review; prototype command skeleton builds.

### Task 6: Shared Utility Extraction

- Objective: Identify utilities suitable for shared packages (`internal/common`, `pkg`, etc.).
- Actions: Audit for duplication; plan extraction strategy.
- Deliverables: Refactoring backlog, risk assessment, staged extraction plan.
- Validation: Staticcheck and go test runs for impacted packages.

### Task 7: Build Pipeline Impact Assessment

- Objective: Analyze how refactor affects GitHub workflows, pre-commit hooks, and scripts.
- Actions: Map dependencies, list updates required per workflow.
- Deliverables: `docs/refactor/pipeline-impact.md`, checklist per workflow.
- Validation: Dry-run of affected workflows using `cmd/workflow`.

### Task 8: Workspace and Tooling Alignment

- Objective: Update VS Code settings, launch configs, and tasks to match new structure.
- Actions: Consolidate autoApprove patterns, add new CLI tasks.
- Deliverables: Revised `.vscode/settings.json`, `tasks.json`, documentation updates.
- Validation: Manual verification in VS Code; lint checks for JSON.

### Task 9: Documentation Restructuring

- Objective: Align README files and docs hierarchy with new service groups.
- Actions: Update `README.md`, `docs/README.md`, add group-specific quick starts.
- Deliverables: Documentation PR plan, editorial checklist.
- Validation: Docs review; link checker run.

### Task 10: Code Migration Phase 1 (Identity)

- Objective: Move identity packages/assets into new group layout.
- Actions: Update imports, go files, tests, docs.
- Deliverables: Migration scripts, refactor PR.
- Validation: `go test ./...`, `golangci-lint run`, full workflow run for identity.

### Task 11: Code Migration Phase 2 (KMS/Core)

- Objective: Relocate KMS-related packages, ensuring minimal downtime.
- Actions: Apply same process as Task 10 with focus on key management services.
- Deliverables: Scripts, documentation updates, test adjustments.
- Validation: Full test suite; inspection of CLI compatibility.

### Task 12: Code Migration Phase 3 (CA)

- Objective: Apply new structure to CA components once built.
- Actions: Coordinate with CA plan tasks; ensure imports match blueprint.
- Deliverables: PR plan, documentation updates, integration tests.
- Validation: `go test` for CA packages; CA-specific workflows passing.

### Task 13: CLI Implementation Sprint 1 (Identity)

- Objective: Deliver identity client/admin CLIs aligned with strategy framework.
- Actions: Implement commands, integrate with orchestrator demo, add tests.
- Deliverables: `cmd/identity` CLI updates, docs, examples.
- Validation: CLI integration tests; manual smoke tests.

### Task 14: CLI Implementation Sprint 2 (KMS)

- Objective: Implement KMS client/admin CLIs.
- Actions: Build commands, tie into key management APIs.
- Deliverables: CLI modules, documentation, usage examples.
- Validation: CLI tests; workflow runs including CLI smoke suite.

### Task 15: CLI Implementation Sprint 3 (CA)

- Objective: Provide CA client/admin CLIs once CA services are available.
- Actions: Extend CLI helpers; add issuance, revocation, audit commands.
- Deliverables: CLI modules, documentation.
- Validation: CLI integration tests; manual issuance demo.

### Task 16: Workflow and Automation Updates

- Objective: Update GitHub workflows, `cmd/workflow`, and scripts to cover new groups.
- Actions: Add targeted workflows, adjust artifact paths, ensure caching works.
- Deliverables: Updated workflow files, `cmd/workflow` enhancements.
- Validation: act dry-runs, CI green.

### Task 17: Backward Compatibility & Deprecation Strategy

- Objective: Provide shims or alias packages for old import paths.
- Actions: Implement temporary wrappers, log deprecation warnings.
- Deliverables: Compatibility layer, migration guide.
- Validation: `go test` across consumers; ensure warnings appear.

### Task 18: Telemetry and Observability Alignment

- Objective: Ensure new structure propagates telemetry labels, dashboards, alerts.
- Actions: Update OTEL instrumentation and Grafana dashboards for groups.
- Deliverables: Telemetry updates, dashboard migrations, alert tune-ups.
- Validation: Observability smoke test; metrics and traces visible.

### Task 19: Final Integration and Regression Testing

- Objective: Run full regression across services, CLIs, and workflows.
- Actions: Execute `go test ./...`, `golangci-lint run`, all workflows via `cmd/workflow`.
- Deliverables: Regression report, highlights of issues found/resolved.
- Validation: Documented sign-off; zero failing tests.

### Task 20: Handoff, Training, and Continuous Improvement

- Objective: Document final architecture, provide training, and establish iteration backlog.
- Actions: Produce training materials, record sessions, curate backlog for next iteration.
- Deliverables: Training pack, video references, continuous improvement backlog.
- Validation: Stakeholder acknowledgment; backlog prioritized for future cycles.
