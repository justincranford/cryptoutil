# Task 05 – Storage Layer Verification

## Objective

Re-validate storage migrations and CRUD behaviour for all identity data models across SQLite (tests) and PostgreSQL (production parity), ensuring schema accuracy and data integrity.

## Historical Context

- Commit `1974b06` introduced the initial storage layer, but no comprehensive migration audit has been performed since.
- Later functionality (Tasks 8–15) added models without confirming backward compatibility.

## Scope

- Re-run migrations from a clean state in both SQLite and PostgreSQL environments.
- Build a dedicated `internal/identity/storage/tests` package that exercises CRUD and transactional behaviour.
- Validate indices, foreign keys, and unique constraints against documented requirements.

## Deliverables

- Migration audit log detailing applied migrations, schema diffs, and any corrective actions.
- Automated test suite covering CRUD operations, rollback scenarios, and concurrency concerns.
- Database fixtures and helpers reusable by downstream tests.

## Validation

- Execute integration tests against SQLite and containerized PostgreSQL (aligned with Docker Compose definitions).
- Include coverage reports to demonstrate breadth (goal: 95%+ for storage packages touched).

## Dependencies

- Relies on configuration templates from Task 03.
- Subsequent tasks (08, 11, 12) depend on reliable storage layers for token and MFA workflows.

## Risks & Notes

- Ensure migrations remain idempotent for CI reuse.
- Capture performance metrics where relevant to guard against regressions.
