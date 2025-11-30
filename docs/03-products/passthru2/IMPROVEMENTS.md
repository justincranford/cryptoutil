# Passthru2: Improvement & PR Task Suggestions

This file lists suggested quick PR tasks and improvements prioritized for immediate work:

## Quick Fixes (High Impact, Low Risk)
- PR-0001: Add `demo` flag to KMS server with seeding logic for users, demo keys, and demo clients
- PR-0002: Add `demo` Docker Compose profile for KMS (deployments/kms/compose.demo.yml) and Identity
- PR-0003: Extract telemetry compose to `deployments/telemetry/compose.yml` and update per-product compose files
- PR-0004: Replace Identity inline secrets with Docker secrets
- PR-0005: Consolidate KMS README / TASK-LIST mismatch into a single authoritative checklist

## Medium PRs
- PR-0101: Add KMS handler/businesslogic unit tests to reach 85% coverage
- PR-0102: Implement Identity `/authorize` endpoint, PKCE, redirect flow and tests
- PR-0103: Implement identity seed data and `--demo` mode
- PR-0104: Add E2E docker-compose tests that run `demo` profile and validate flows

## Larger PRs
- PR-1001: Extract `internal/common/*` into `internal/infra/*` (magic, apperr, crypto, telemetry) with per-package commits
- PR-1002: Create JOSE Authority product extraction and JOSE centralization
- PR-1003: Implement per-product CLI demo (`cmd/demo`), or `make demo` wrapper for quick demonstration across OSes

## Migration & CI
- CI-001: Add demo-run CI job to run `deployments/<product>/compose.demo.yml` health checks and run the demo script
- CI-002: Add coverage gating per package and coverage reporting to `test-output/`

## Notes
- PR-1001 should be done one package at a time with build + tests + lint checks after each move
- PR-0101/0102 are high priority for PASSTHRU2 acceptance

