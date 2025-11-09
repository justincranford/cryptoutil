# Identity V2 Remediation Plan

## Purpose

- Deliver a complete overhaul of the identity stack with working UI, reliable APIs, and automated orchestration.
- Audit the original 20 committed tasks, document reality, and repair every missing or broken feature.
- Provide deterministic automation for local, CI, and manual validation of identity services.

## Guiding Principles

- Maintain strict separation between identity packages and other domains (KMS, CA, etc.).
- Prefer deterministic, YAML-driven configuration for repeatable orchestration.
- Every task must produce updated documentation, automated tests, and observable outcomes.
- All CLI entry points must be scriptable and support Windows PowerShell and Linux shells.

## Task Breakdown

### Task 1: Historical Baseline Assessment

- Analysis Focus: Capture the exact state of identity-related commits from 15cd829760f6bd6baf147cd953f8a7759e0800f4 through HEAD.
- Issues to Address: Missing documentation on implemented features and broken functionality.
- Implementation Notes: Produce comparison matrices (commit vs. current behavior) for authz, IdP, RS, RP components.
- Deliverables: `docs/identityV2/history-baseline.md`, updated architectural diagrams, gap summary.
- Validation: Peer review of documentation; ensure matrices cover all 20 tasks.

### Task 2: Requirements and Success Criteria Registry

- Analysis Focus: Translate user flows (authz, token issuance, RP UI) into measurable requirements.
- Issues to Address: Ambiguous acceptance criteria for prior tasks.
- Implementation Notes: Map requirements to existing tests and note gaps.
- Deliverables: `docs/identityV2/requirements.yml` enumerating success metrics and dependencies.
- Validation: Review requirements with stakeholders; cross-check with LONGER-TERM-IDEAS mandate.

### Task 3: Configuration Inventory and Normalization

- Analysis Focus: Collect all identity-related config, flags, and secrets.
- Issues to Address: Divergent config between services and missing defaults.
- Implementation Notes: Normalize into versioned YAML templates under `configs/identity/`.
- Deliverables: `configs/identity/*.yml`, config diff report.
- Validation: `go test ./internal/common/config/...` with new fixtures.

### Task 4: Identity Package Dependency Audit

- Analysis Focus: Inspect `internal/identity/**` for hidden dependencies on other domains.
- Issues to Address: Undocumented coupling with crypto or CA packages.
- Implementation Notes: Introduce linter rule or staticcheck configuration enforcing boundaries.
- Deliverables: Dependency graph (`docs/identityV2/dependency-graph.md`), updated `.golangci.yml` if needed.
- Validation: `golangci-lint run` passes with new checks; dependency report shows compliance.

### Task 5: Storage Layer Verification

- Analysis Focus: Re-run migrations and CRUD operations for identity data models.
- Issues to Address: Schema drift, missing migrations, inconsistent GORM models.
- Implementation Notes: Create dedicated SQLite and PostgreSQL fixtures; validate migration history.
- Deliverables: Migration audit log, `internal/identity/storage/tests` package.
- Validation: Integration tests against SQLite and PostgreSQL containers.

### Task 6: OAuth 2.1 Authorization Server Core Rehab

- Analysis Focus: Review Task 4 implementation vs. spec requirements (auth code, PKCE, scopes).
- Issues to Address: Missing flows, incomplete error handling, logging gaps.
- Implementation Notes: Align with draft 15; wire contextual logging.
- Deliverables: Updated server code, spec conformance matrix, logs.
- Validation: Automated grant-type tests; Postman collection or Go integration suite.

### Task 7: Client Authentication Enhancements

- Analysis Focus: Reconcile Tasks 5 and 6 (basic + mTLS) with security requirements.
- Issues to Address: Certificate validation gaps, fallback logic, CLI ergonomics.
- Implementation Notes: Introduce policy-driven auth methods, align with `internal/common/security`.
- Deliverables: CLI samples, mTLS integration tests, config toggles.
- Validation: mTLS end-to-end test harness; `go test ./internal/identity/auth/...`.

### Task 8: Token Service Hardening

- Analysis Focus: Inspect JWT/JWE logic from Task 3 for correctness and claims coverage.
- Issues to Address: Missing key rotation strategy, algorithm negotiation.
- Implementation Notes: Introduce deterministic key source abstraction and rotation schedule.
- Deliverables: Updated key management docs, rotation CLI, storage hooks.
- Validation: Fuzz tests for token parsing; expiration and revocation tests.

### Task 9: SPA Relying Party UX Repair

- Analysis Focus: Diagnose UI failures (button dead-ends, state sync issues).
- Issues to Address: API incompatibilities, missing loader states, lack of telemetry.
- Implementation Notes: Introduce contract tests between SPA and API; add telemetry hooks.
- Deliverables: Updated SPA build, Cypress/E2E script, user journey docs.
- Validation: Manual walkthrough checklist; automated UI smoke tests.

### Task 10: Integration Layer Completion

- Analysis Focus: Finish incomplete Task 10 infrastructure; ensure service composition works.
- Issues to Address: Missing queue listeners, placeholder mocks, Docker Compose drift.
- Implementation Notes: Update compose overrides; align ports, health checks, secrets.
- Deliverables: Updated `deployments/compose/identity.yml`, service topology diagrams.
- Validation: `go test ./internal/identity/integration/...`; Docker Compose smoke run.

### Task 11: Client MFA Chains Stabilization

- Analysis Focus: Revisit Task 11 multi-factor flows and state machines.
- Issues to Address: Edge case handling, concurrency safety, recovery flows.
- Implementation Notes: Implement idempotent MFA session store and telemetry.
- Deliverables: Updated MFA library, state diagrams, retry policy docs.
- Validation: Load tests for MFA flows; concurrency unit tests.

### Task 12: OTP and Magic Link Services

- Analysis Focus: Confirm SMS/email providers mocked vs. production connectors.
- Issues to Address: Rate limiting, token leakage, audit logging.
- Implementation Notes: Introduce provider abstraction with contract tests.
- Deliverables: Provider adapters, rate-limit policy config, audit logs.
- Validation: Contract tests with fake providers; log verification scripts.

### Task 13: Adaptive Authentication Engine

- Analysis Focus: Validate risk scoring, contextual prompts, and extension points.
- Issues to Address: Missing data sources, scoring algorithms, configurability.
- Implementation Notes: Externalize policies to YAML; add simulation CLI.
- Deliverables: Policy schema, simulation reports, documentation.
- Validation: Scenario-based integration tests; manual policy review.

### Task 14: Biometric + WebAuthn Path

- Analysis Focus: Bring WebAuthn stubs to production readiness.
- Issues to Address: Attestation validation, browser compatibility, fallback flows.
- Implementation Notes: Integrate with existing WebAuthn library; add acceptance suite.
- Deliverables: Updated RP APIs, WebAuthn docs, compatibility matrix.
- Validation: WebAuthn conformance tests; browser matrix results recorded.

### Task 15: Hardware Credential Support

- Analysis Focus: Ensure hardware based auth (e.g., smart cards, FIDO keys) is end-to-end.
- Issues to Address: Driver abstraction, error messaging, admin tooling.
- Implementation Notes: Provide CLI to enroll hardware credentials with audit trails.
- Deliverables: Enrollment CLI, admin guide, failure troubleshooting appendix.
- Validation: Manual hardware tests with documented outcomes; automated mocks.

### Task 16: OpenAPI 3.0 Spec Modernization

- Analysis Focus: Align spec with rebuilt services and add missing endpoints.
- Issues to Address: Divergence between spec and implementation, missing enums.
- Implementation Notes: Regenerate clients/servers, add linting for spec drift.
- Deliverables: Updated `api/openapi_spec_*.yaml`, regeneration scripts.
- Validation: `go generate ./api/...`; schema validation via `oapi-codegen validate`.

### Task 17: Gap Analysis and Remediation Plan

- Analysis Focus: Summarize outstanding issues after technical rebuild.
- Issues to Address: Ensure no task regression; align with compliance requirements.
- Implementation Notes: Update gap log, assign severity, attach mitigation timeline.
- Deliverables: `docs/identityV2/gap-analysis.md`, remediation tracker.
- Validation: Stakeholder sign-off; gap tracker linked to issues/PRs.

### Task 18: Docker Compose Orchestration Suite

- Analysis Focus: Build final orchestration layer with Nx/Mx/Xx/Yx/Zx container scaling.
- Issues to Address: Service naming, network isolation, secret reuse.
- Implementation Notes: Provide templated Compose bundles and orchestrator CLI.
- Deliverables: `deployments/compose/identity-demo.yml`, orchestrator commands.
- Validation: Automated smoke tests via `go test ./internal/identity/demo/...`; manual start-stop script.

### Task 19: Integration and E2E Testing Fabric

- Analysis Focus: Establish comprehensive integration + e2e suites for identity flows.
- Issues to Address: Coverage gaps, flaky tests, missing telemetry correlation.
- Implementation Notes: Integrate with `cmd/workflow` to run identity suites; produce artifacts.
- Deliverables: New Go test packages, workflow updates, coverage dashboards.
- Validation: `go test ./internal/identity/... -tags=e2e -run Integration`; workflow dry run using `go run ./cmd/workflow -workflows=e2e`.

### Task 20: Final Verification and Delivery Readiness

- Analysis Focus: Perform regression testing, documentation handoff, and DR drills.
- Issues to Address: Ensure orchestration CLI, UI, and APIs are production-ready.
- Implementation Notes: Conduct blue/green rehearsal, backup/restore validation.
- Deliverables: Release checklist, DR runbook, training materials.
- Validation: Full system test sign-off; readiness review with leadership.
