// Copyright (c) 2025-2026 Justin Cranford.

// Package test_help_tls provides TLS material creation, certificate/client construction,
// and secure/insecure test client helpers for integration and E2E test suites.
// It handles test TLS certificate generation, mTLS client setup, and client configuration.
//
// Consumed by:
//   - test_orch_e2e: TLS material for E2E tests
//   - test_orch_integration: TLS clients and certificates
//   - TLS test suites: certificate validation and mTLS testing
package test_help_tls
