# Identity V2 Remediation Program Master Plan

## Purpose

- Deliver a repeatable remediation program that brings the identity stack to production quality with reliable APIs, resilient UIs, and automated orchestration.
- Audit the previously completed Tasks 1-15 from the legacy plan (`docs/identity/identity_master.md`) and reconcile them with the actual repository state from commit `15cd829760f6bd6baf147cd953f8a7759e0800f4` through `HEAD`.
- Close the gaps identified in historical commits that marked several tasks as "partial" (`d784ca6`, `74dcf14`) and ensure each remediation activity lands with executable automation, documentation, and telemetry.
- Establish deterministic end-to-end execution: local workflows, Docker Compose orchestration, and Go-based CLIs must all share the same validated configuration contracts.

## Historical Findings

- **Legacy plan baseline**: The original 15-task program in `docs/identity/identity_master.md` was implemented across commits `1974b06` through `2514fef`, but multiple follow-up commits (for example `dc68619`, `5c04e44`) patched issues piecemeal, confirming the need for a holistic remediation sweep.
- **Partial completions**: Commits `d784ca6` (Task 8) and `74dcf14` (Task 10) explicitly recorded partial implementations; newer fixes did not provide integrated verification, so these areas require full re-validation.
- **Documentation drift**: Commits `80d4e00`, `a6884d3`, and `d91791b` introduced overlapping documentation sets that no longer align with the repository. The new remediation tasks re-orient around the current codebase instead of historic assumptions.
- **Workflow integration gaps**: The addition of mock services (`5c04e44`) resolved e2e test setup friction but also highlighted configuration inconsistencies across identity services that this plan must normalize.

## Execution Guidelines

- Execute tasks strictly in numerical order to maintain dependency flow.
- Commit after **every** task (`git commit` per task) before moving to the next. Never combine multiple tasks into one commit.
- Each task document defines objective, scope, deliverables, validation, dependencies, and known risks. Treat the documents as contracts; update them only through follow-up change control.
- Maintain at least 95% test coverage for the identity packages touched by each task. Add regression tests whenever remediation fixes a historical defect.
- Keep documentation synchronized. When a task updates behaviour or configuration, update both the task deliverables and any affected higher-level docs (for example, service READMEs or runbooks).

## Task Index

| # | Task | Document | Summary |
|---|------|----------|---------|
| 01 | Historical Baseline Assessment | `docs/identityV2/task-01-historical-baseline-assessment.md` | Capture the precise identity state from commit `15cd8297` through `HEAD` and produce comparison matrices. |
| 02 | Requirements and Success Criteria Registry | `docs/identityV2/task-02-requirements-success-criteria.md` | Translate user and client flows into measurable requirements with traceability. |
| 03 | Configuration Inventory and Normalization | `docs/identityV2/task-03-configuration-normalization.md` | Create canonical configuration templates and fixtures across services. |
| 04 | Identity Package Dependency Audit | `docs/identityV2/task-04-dependency-audit.md` | Detect and enforce domain boundaries within `internal/identity/**`. |
| 05 | Storage Layer Verification | `docs/identityV2/task-05-storage-verification.md` | Validate migrations, schemas, and CRUD operations across SQLite/PostgreSQL. |
| 06 | OAuth 2.1 Authorization Server Core Rehab | `docs/identityV2/task-06-authz-core-rehab.md` | Align the authorization server with OAuth 2.1 draft 15 and strengthen telemetry. |
| 07 | Client Authentication Enhancements | `docs/identityV2/task-07-client-auth-enhancements.md` | Harden client authentication methods including mTLS and policy controls. |
| 08 | Token Service Hardening | `docs/identityV2/task-08-token-service-hardening.md` | Introduce deterministic key rotation and expand token validation coverage. |
| 09 | SPA Relying Party UX Repair | `docs/identityV2/task-09-spa-ux-repair.md` | Restore SPA usability, add telemetry, and align API contracts. |
| 10 | Integration Layer Completion | `docs/identityV2/task-10-integration-layer-completion.md` | Finish service composition, queue listeners, and Docker orchestration. |
| 11 | Client MFA Chains Stabilization | `docs/identityV2/task-11-client-mfa-stabilization.md` | Ensure multi-factor flows are concurrency-safe with observability hooks. |
| 12 | OTP and Magic Link Services | `docs/identityV2/task-12-otp-magic-link.md` | Secure provider abstractions with rate limiting and auditing. |
| 13 | Adaptive Authentication Engine | `docs/identityV2/task-13-adaptive-engine.md` | Externalize risk policies and add simulation support. |
| 14 | Biometric + WebAuthn Path | `docs/identityV2/task-14-biometric-webauthn.md` | Bring WebAuthn features to production readiness. |
| 15 | Hardware Credential Support | `docs/identityV2/task-15-hardware-credential-support.md` | Deliver end-to-end hardware credential enrolment and validation. |
| 16 | OpenAPI 3.0 Spec Modernization | `docs/identityV2/task-16-openapi-modernization.md` | Synchronize OpenAPI specs and generation configs with current services. |
| 17 | Gap Analysis and Remediation Plan | `docs/identityV2/task-17-gap-analysis.md` | Produce remediation tracker aligned with compliance obligations. |
| 18 | Docker Compose Orchestration Suite | `docs/identityV2/task-18-orchestration-suite.md` | Build deterministic orchestration bundles and tooling. |
| 19 | Integration and E2E Testing Fabric | `docs/identityV2/task-19-integration-e2e-fabric.md` | Create comprehensive automated test coverage and reporting. |
| 20 | Final Verification and Delivery Readiness | `docs/identityV2/task-20-final-verification.md` | Execute regression, DR drills, and delivery sign-off. |

## Dependencies and Cross-Cutting Concerns

- **Configuration Consistency**: Tasks 3, 6, 7, 8, 10, and 18 share configuration artefacts. Coordinate template changes through Task 3 before downstream tasks consume them.
- **Testing Infrastructure**: Tasks 1, 5, 6, 8, 11, 12, 13, and 19 all depend on stable test utilities. Consolidate helpers under `internal/identity/testutil` as part of Task 5 deliverables.
- **Documentation**: Task outputs that land in `docs/identity/**` must reference the legacy plan for traceability until the new artefacts fully replace it.
- **Tooling**: Any new lint rules or static analysis introduced in Task 4 must be reflected in `.golangci.yml` and documented within `docs/pre-commit-hooks.md` to maintain tooling parity.

## Exit Criteria

- All twenty task documents have their deliverables satisfied, validated, and linked within this master plan.
- Identity services achieve consistent startup via Go CLI, Docker Compose, and automated workflows with 95%+ coverage on touched packages.
- Historical gaps documented in commits `d784ca6`, `74dcf14`, and subsequent fixes are resolved with regression tests preventing recurrence.
- Stakeholders receive updated runbooks, architectural diagrams, and remediation trackers demonstrating production readiness.
