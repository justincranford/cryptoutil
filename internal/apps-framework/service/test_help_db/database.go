// Copyright (c) 2025-2026 Justin Cranford.

// Package test_help_db provides database fixture creation, schema migrations, and DB failure-path helpers
// for integration and E2E test suites. It handles SQLite in-memory setup, PostgreSQL containers (E2E only),
// and deterministic DB error creation for error-path testing.
//
// Consumed by:
//   - test_orch_integration: database fixture creation and migration
//   - test_orch_e2e: PostgreSQL test container setup
//   - Repository test suites: DB fixtures
package test_help_db
