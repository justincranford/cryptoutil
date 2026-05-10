// Copyright (c) 2025-2026 Justin Cranford.

// Package test_help_bootstrap provides configuration, environment, and startup wiring helpers
// for integration and E2E test suites. It handles config loading, environment variable setup,
// and bootstrap orchestration needed before starting test servers or compose stacks.
//
// Consumed by:
//   - test_orch_e2e: compose environment and config setup
//   - test_orch_integration: server startup config wiring
//   - Integration/E2E test suites: config loading and env setup
package test_help_bootstrap
