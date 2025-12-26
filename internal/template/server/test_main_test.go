// Copyright (c) 2025 Justin Cranford

package server_test

import (
	"os"
	"testing"

	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
)

var (
	testPublicTLS  *cryptoutilTLSGenerator.TLSGeneratedSettings
	testPrivateTLS *cryptoutilTLSGenerator.TLSGeneratedSettings
)

func TestMain(m *testing.M) {
	// Generate shared TLS fixtures for tests (auto-mode, localhost/IPs).
	var err error

	testPublicTLS, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings([]string{"localhost"}, []string{"127.0.0.1"}, 365)
	if err != nil {
		panic("failed to generate public TLS fixtures: " + err.Error())
	}

	testPrivateTLS, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings([]string{"localhost"}, []string{"127.0.0.1"}, 365)
	if err != nil {
		panic("failed to generate private TLS fixtures: " + err.Error())
	}

	os.Exit(m.Run())
}
