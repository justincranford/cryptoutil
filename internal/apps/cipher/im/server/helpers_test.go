// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"net/http"
	"testing"

	"cryptoutil/internal/apps/cipher/im/server/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilE2E "cryptoutil/internal/template/testing/e2e"
)

// initTestConfig creates a properly configured AppConfig for testing.
func initTestConfig() *config.AppConfig {
	cfg := config.DefaultAppConfig()
	cfg.BindPublicPort = 0                                                          // Dynamic port allocation for tests
	cfg.BindPrivatePort = 0                                                         // Dynamic port allocation for tests
	cfg.OTLPService = "cipher-im-test"                                              // Required for telemetry initialization
	cfg.LogLevel = "info"                                                           // Required for logger initialization
	cfg.OTLPEndpoint = "grpc://" + cryptoutilMagic.HostnameLocalhost + ":" + "4317" // Required for OTLP endpoint validation
	cfg.OTLPEnabled = false                                                         // Disable actual OTLP export in tests

	return cfg
}

// createHTTPClient creates an HTTP client that trusts self-signed certificates.
func createHTTPClient(t *testing.T) *http.Client {
	t.Helper()

	return cryptoutilE2E.CreateInsecureHTTPClient(t)
}
