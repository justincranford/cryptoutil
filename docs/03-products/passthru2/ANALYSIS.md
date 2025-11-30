# Deep Analysis: What Was Done in `passthru1` and Recommendations for `passthru2`

This document distills the status and gaps found in `docs/03-products/passthru1` and maps required improvements for `passthru2`.

## 1) Completed Items (from passthru1)

- KMS server startup, Swagger UI, `/livez` and `/readyz` endpoints verified
- KMS: encrypt/decrypt, sign/verify functionality verified
- Identity: token endpoint, client_credentials flow, basic discovery partially working
- Identity DB & migrations validated for SQLite/Postgres
- Coverage baseline collected for KMS/Identity and prioritized test gaps
- PRAGMA & SQLite connection pool settings implemented across packages for concurrency (WAL + busy_timeout)
- PBKDF2 migration completed in Identity for FIPS compliance

---

## 2) Not Completed / Pending Items

- Identity: Authorization Code flow (/authorize) and full PKCE support
- Identity: Session management, refresh token rotation, some revocation/inspect tests incomplete
- Identity: Domain & repository test coverage gaps especially in `idp` & `idp/userauth`
- KMS: Demo parity for seed accounts and UI demo `Try it out` examples (inconsistent status between README and TASK-LIST)
- KMS: Add UI-based key pool creation and demo-friendly walkthrough
- Cross-product Integration: End-to-end flows verifying scopes and embedded Identity still pending
- E2E: Cross-product E2E tests not implemented; per-product E2E exists but needs consistent naming & guards

---

## 3) Bugs, Mistakes & Inconsistencies

- Inconsistent documentation: `README.md` vs `TASK-LIST.md` in KMS regarding "Create key pool" status (blocked vs working). Consolidate authoritative checklist.
- Duplicate code: Identity contains copies of `apperr/`, `config/`, `magic/`, etc. Strategy is to extract and consolidate in infra.
- Telemetry coupling: `deployments/kms/compose.yml` contains telemetry; Identity references telemetry but doesn't include a `telemetry` compose file. This prevents running Identity independently.
- Secrets: KMS uses docker secrets, Identity still uses inline envs. Standardize to Docker secrets for all deploys.
- Demo parity: Identity has `demo` concept; KMS lacks equivalent `demo` startup & seeding (the plan mentions a `demo` mode but is not consistent across docs)

---

## 4) Missed Requirements & Non-Functional Gaps

- Non-functional: No per-product `demo` compose profile for KMS (Identity has this in practice), lacking scripted demo experiences.
- Non-functional: Insufficient test coverage in KMS handler & businesslogic packages, especially for behavior and errors.
- Non-functional: No consistent audit logging approach; KMS and Identity have different levels of audit detail.
- Scalability: Telemetry & compose splitting needed to avoid monolith coupling for new products.
- Parallel testability: Verified PRAGMA and MaxOpenConns patterns for SQLite exist, but some unit tests may still rely on single connection settings; confirm all test code uses shared cache for `:memory:` DB.

---

## 5) Copilot / Project Instruction Compliance

Key checks against repo instruction files:

- Coding & Patterns: `passthru1` proposals use switch statements and default values; KMS code follows the `internal/` split; Identity replication duplicates some infra which will be consolidated as per instructions.
- Testing: Table-driven tests and `t.Parallel()` are used across the project but the identity packages show deficiencies in coverage and need to adhere to `t.Parallel()` patterns and table-driven tests (see coverage gaps in TASK-LIST).
- Database: GORM SQLite setup and `PRAGMA` settings for concurrency are implemented in both KMS and Identity; Identity uses `sql.Open` + `sqlite.Dialector` pattern per instructions.
- Security: Initial bcrypt usage in Identity was removed (replaced with PBKDF2), aligning with FIPS constraints. JWT signing and JOSE extraction are being reworked into a JOSE Authority in the plan.
- Linting: Plans to maintain zero-lint errors and coverage gates exist; enforcement was not fully implemented prior to passthru1 but must be in passthru2.

---

## 6) Identity vs KMS Best Practices Comparison

What identity lacks compared to KMS best practices:

- KMS: Strong unit test coverage for core logic; Identity has low coverage in important packages (handler, idp).
- KMS: Demonstrated stable deployments and UI demo; Identity lacks seeded demo users/clients account parity and certain flows.
- KMS: Consistent DB settings and GORM usage for SQLite in-memory; Identity follows these best practices but earlier had issues (fixed) and some duplicated helpers that differ.
- KMS: Barrier pattern and business logic layering; Identity's code sometimes mixes logic in handlers and needs to enforce same separation of business logic vs handler vs repository.

Risk: If Identity remains inconsistent, KMS may end up reusing identity-specific code instead of infra code, creating duplication and divergence.

---

## 7) Design, Security & Implementation Observations

- JOSE Authority: Good idea to extract from Identity issuer; centralizing JWT/JWS operations improves consistency and eliminates duplicated implementations.
- Key Management: KMS demonstrates 3-tier key hierarchy; ensure Identity's JWK rotation uses KMS where needed, or be explicit in design for usage when JOSE Authority is KMS-backed vs standalone.
- Token Validation: KMS should support both introspection and local JWT verification with caching for performance (choose mix per Q17 in GROOMING Qs).
- Audit & Logs: Define a cross-product audit log format and tamper-resistant storage for any regulatory needs. Implement in infra telemetry and optionally integrate with otel collector.

---

## 8) Scalability & Parallel Testability

- The PRAGMA + MaxOpenConns pattern for SQLite is used correctly in both KMS and Identity packages, enabling parallel test runs with WAL+busy_timeout and proper connection pool sizing.
- However, ensure that GORM with transactions (SkipDefaultTransaction) config usage is consistent; GORM transactions with MaxOpenConns settings must be tested across packages.
- Suggest: Run a focused set of parallel tests for Identity and KMS to find any deadlocks.

---

## 9) E2E & Integration

- E2E tests need per-product demo orchestration and cross-product orchestration in `internal/product/e2e/` or `test/e2e/`.
- Compose merging: implement `deployments/telemetry/compose.yml` for central telemetry and optional local telemetry runs.
- Compose `demo` profile should seed data to ensure deterministic E2E runs.

---

## 10) Recommended Improvement Tasks (map to passthru2 grooming: RESEARCH.md)

Immediate (Phase 0):

- Add `demo` mode and seeding for KMS and Identity
- Extract telemetry to `deployments/telemetry/compose.yml` and update product compose files to reference it
- Standardize secret usage: Docker secrets across products
- Standardize config locations under `deployments/<product>/config/`
- Remove empty directories and document expected structure

Short term:

- Add missing flows: Identity `/authorize`, PKCE validation, redirect handling, token rotation logic, refresh tests
- KMS UI demo: ensure `Try it out` and seeded accounts work in Swagger UI
- Fix coverage gaps: KMS handler/businesslogic and Identity `idp/userauth` need + tests
- Consolidate duplicate infra code in `internal/infra/` and update import paths

Medium term:

- Implement JOSE Authority extraction and define JWK/JWS profiles
- E2E tests that run entire demo workflow with seeded data (integration mode)
- CI updates: coverage gates & demo-run CI job

---

## 11) Suggested Next Steps & Quick Wins

- Answer 25 grooming questions in `docs/03-products/passthru2/grooming/GROOMING-QUESTIONS.md`
- Implement Phase 0 tasks: demo seeding + telemetry extraction
- Add per-product `demo` orchestration scripts
- Start writing tests for KMS and Identity gaps prioritized by coverage impact

---

**Final Note:** `passthru1` provided a pragmatic and aggressive plan to get working demos. `passthru2` should retain that pragmatism while emphasizing developer experience, consistent infra, and parity between product demos. The first concrete step is to align demos via `demo` compose profiles and seed data so UX reviewers can validate the experience for all products.
