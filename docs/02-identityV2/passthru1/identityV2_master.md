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

### Task Reflection Requirement

**CRITICAL**: Every task MUST begin with a reflection section evaluating prior work:

```markdown
## Task Reflection

### What Went Well
- [List successes from previous tasks]

### At Risk Items
- [Identify risks discovered during implementation]

### Could Be Improved
- [Note areas for enhancement or technical debt]

### Dependencies and Blockers
- [Any dependencies from prior tasks needing resolution]
```

This reflection informs the current task's approach and ensures continuous improvement.

### Execution Rules

1. **Strictly Sequential Execution**: Complete tasks in numerical order (01 → 02 → ... → 19)
   - Exception: If a task discovers blocking issues in prior tasks, pause and remediate before continuing
   - Document any out-of-order fixes in task reflection sections

2. **Commit After Every Task**: Each task completion requires a git commit with:
   - Conventional commit format: `feat(identity): complete task NN - <summary>`
   - Detailed commit message referencing task document
   - All linting passing (`golangci-lint run`)
   - All tests passing (`go test ./...`)

3. **Quality Gates**: Before marking task complete:
   - ✅ All deliverables from task document implemented
   - ✅ Tests written and passing (unit, integration, e2e as appropriate)
   - ✅ Documentation updated (code comments, README, runbooks)
   - ✅ Linting passes with zero violations
   - ✅ Task reflection completed and added to task document

4. **Incremental Progress**: Break large tasks into subtasks but commit the complete task atomically
   - Use `manage_todo_list` tool for subtask tracking during implementation
   - Final commit represents complete, tested, documented task

5. **Cross-Platform Validation**: All code must work on:
   - Windows PowerShell development environments
   - Linux GitHub Actions runners
   - Docker containers (Alpine, Ubuntu bases)

6. **Bootstrap Simplicity Goal**: Maintain focus on delivering one-liner commands:
   - `./identity start --profile demo` (local development)
   - `docker compose -f compose.advanced.yml up` (containerized)
   - `./identity test --suite e2e` (testing)

## Task Index

### ✅ PHASE 1: Foundation (Tasks 01-10.4) - COMPLETED

| # | Task | Document | Summary | Status |
|---|------|----------|---------|--------|
| 01 | Historical Baseline Assessment | `docs/identityV2/task-01-historical-baseline-assessment.md` | Capture the precise identity state from commit `15cd8297` through `HEAD` and produce comparison matrices. | ✅ Complete |
| 02 | Requirements and Success Criteria Registry | `docs/identityV2/task-02-requirements-success-criteria.md` | Translate user and client flows into measurable requirements with traceability. | ✅ Complete |
| 03 | Configuration Inventory and Normalization | `docs/identityV2/task-03-configuration-normalization.md` | Create canonical configuration templates and fixtures across services. | ✅ Complete |
| 04 | Identity Package Dependency Audit | `docs/identityV2/task-04-dependency-audit.md` | Detect and enforce domain boundaries within `internal/identity/**`. | ✅ Complete |
| 05 | Storage Layer Verification | `docs/identityV2/task-05-storage-verification.md` | Validate migrations, schemas, and CRUD operations across SQLite/PostgreSQL. | ✅ Complete |
| 06 | OAuth 2.1 Authorization Server Core Rehab | `docs/identityV2/task-06-authz-core-rehab.md` | Align the authorization server with OAuth 2.1 draft 15 and strengthen telemetry. | ✅ Complete |
| 07 | Client Authentication Enhancements | `docs/identityV2/task-07-client-auth-enhancements.md` | Harden client authentication methods including mTLS and policy controls. | ✅ Complete |
| 08 | Token Service Hardening | `docs/identityV2/task-08-token-service-hardening.md` | Introduce deterministic key rotation and expand token validation coverage. | ✅ Complete |
| 09 | SPA Relying Party UX Repair | `docs/identityV2/task-09-spa-ux-repair.md` | Restore SPA usability, add telemetry, and align API contracts. | ✅ Complete |
| 10.1-10.4 | Integration Layer Infrastructure | `docs/identityV2/task-10-integration-layer-completion.md` | Integration tests, background jobs, queue decision, architecture docs. | ✅ Complete |

### 🔴 PHASE 2: Critical Path - Foundation Completion (Tasks 10.5-10.7) - IN PROGRESS

**CRITICAL**: Integration tests revealed AuthZ/IdP servers lack core OAuth/OIDC endpoints. Must complete before feature additions.

| # | Task | Document | Summary | Status |
|---|------|----------|---------|--------|
| 10.5 | AuthZ/IdP Core Endpoints | `docs/identityV2/task-10.5-authz-idp-endpoints.md` | Implement `/oauth2/v1/authorize`, `/oauth2/v1/token`, `/health`, `/oidc/v1/login` to make integration tests pass. | ✅ Complete |
| 10.6 | Unified Identity CLI | `docs/identityV2/task-10.6-unified-cli.md` | Create `./identity start --profile <name>` for one-liner bootstrap of all service combinations. | 🔴 **NEXT** |
| 10.7 | OpenAPI Synchronization | `docs/identityV2/task-10.7-openapi-sync.md` | Synchronize OpenAPI specs with implemented endpoints, generate client libraries, update Swagger UI. | ⏳ Pending |

### 🟡 PHASE 3: Enhanced Features (Tasks 11-15)

| # | Task | Document | Summary | Status |
|---|------|----------|---------|--------|
| 11 | Client MFA Chains Stabilization | `docs/identityV2/task-11-client-mfa-stabilization.md` | Ensure multi-factor flows are concurrency-safe with observability hooks. | ⏳ Pending |
| 12 | OTP and Magic Link Services | `docs/identityV2/task-12-otp-magic-link.md` | Secure provider abstractions with rate limiting and auditing. | ⏳ Pending |
| 13 | Adaptive Authentication Engine | `docs/identityV2/task-13-adaptive-engine.md` | Externalize risk policies and add simulation support. | ⏳ Pending |
| 14 | Biometric + WebAuthn Path | `docs/identityV2/task-14-biometric-webauthn.md` | Bring WebAuthn features to production readiness. | ⏳ Pending |
| 15 | Hardware Credential Support | `docs/identityV2/task-15-hardware-credential-support.md` | Deliver end-to-end hardware credential enrolment and validation. | ⏳ Pending |

### 🟢 PHASE 4: Quality & Delivery (Tasks 16-20)

| # | Task | Document | Summary | Status |
|---|------|----------|---------|--------|
| 16 | Gap Analysis and Remediation Plan | `docs/identityV2/task-16-gap-analysis.md` | Produce remediation tracker aligned with compliance obligations. | ⏳ Pending |
| 17 | Docker Compose Orchestration Suite | `docs/identityV2/task-17-orchestration-suite.md` | Build deterministic orchestration bundles and Docker profiles. | ⏳ Pending |
| 18 | Integration and E2E Testing Fabric | `docs/identityV2/task-18-integration-e2e-fabric.md` | Create comprehensive automated test coverage and reporting. | ⏳ Pending |
| 19 | Final Verification and Delivery Readiness | `docs/identityV2/task-19-final-verification.md` | Execute regression, DR drills, and delivery sign-off. | ⏳ Pending |

**Note**: Original Tasks 16-20 renumbered to 16-19. Task 18 (Orchestration) merged into 10.6-10.7 and 17.

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
