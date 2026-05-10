// Copyright (c) 2025-2026 Justin Cranford.

// Package test_help_barrier provides barrier and unseal key fixture composition helpers
// for integration and E2E test suites that need encryption-at-rest (barrier layer) support.
// It handles barrier service setup, unseal key derivation, and elastic key ring initialization.
//
// Consumed by:
//   - test_orch_integration: optional fixture for barrier-heavy tests
//   - Repository test suites: barrier service fixtures
//   - API test suites: barrier-protected resource fixtures
package test_help_barrier
