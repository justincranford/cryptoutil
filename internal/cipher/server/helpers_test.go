// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"cryptoutil/internal/cipher/repository"
	"cryptoutil/internal/cipher/server"
	"cryptoutil/internal/cipher/server/config"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
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

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilMagic.CipherDefaultTimeout, // Increased for concurrent test execution.
	}
}

// createTestCipherIMServer creates a full CipherIMServer for testing using shared resources.
// Returns the server instance, public URL, and admin URL.
func createTestCipherIMServer(t *testing.T, db *gorm.DB) (*server.CipherIMServer, string, string) {
	t.Helper()

	ctx := context.Background()

	// Clean database for test isolation.
	cleanTestDB(t)

	// Create AppConfig with test settings.
	cfg := &config.AppConfig{
		ServerSettings: cryptoutilConfig.ServerSettings{
			BindPublicProtocol:    cryptoutilMagic.ProtocolHTTPS,
			BindPublicAddress:     cryptoutilMagic.IPv4Loopback,
			BindPublicPort:        0, // Dynamic allocation
			BindPrivateProtocol:   cryptoutilMagic.ProtocolHTTPS,
			BindPrivateAddress:    cryptoutilMagic.IPv4Loopback,
			BindPrivatePort:       0, // Dynamic allocation
			TLSPublicDNSNames:     []string{cryptoutilMagic.HostnameLocalhost},
			TLSPublicIPAddresses:  []string{cryptoutilMagic.IPv4Loopback},
			TLSPrivateDNSNames:    []string{cryptoutilMagic.HostnameLocalhost},
			TLSPrivateIPAddresses: []string{cryptoutilMagic.IPv4Loopback},
			CORSAllowedOrigins:    []string{},
			OTLPService:           "cipher-im-server-test",
			OTLPEndpoint:          "",
			LogLevel:              "error",
		},
		JWTSecret: testJWTSecret,
	}

	// Create full server.
	cipherServer, err := server.New(ctx, cfg, testDB, repository.DatabaseTypeSQLite)
	require.NoError(t, err)

	// Start server in background.
	errChan := make(chan error, 1)

	go func() {
		if startErr := cipherServer.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for both servers to bind to ports.
	const (
		maxWaitAttempts = 50
		waitInterval    = 100 * time.Millisecond
	)

	var publicPort int
	var adminPort int

	for i := 0; i < maxWaitAttempts; i++ {
		publicPort = cipherServer.PublicPort()

		adminPortValue, _ := cipherServer.AdminPort()
		adminPort = adminPortValue

		if publicPort > 0 && adminPort > 0 {
			break
		}

		select {
		case err := <-errChan:
			require.NoError(t, err)
		case <-time.After(waitInterval):
		}
	}

	if publicPort == 0 {
		t.Fatal("createTestCipherIMServer: public server did not bind to port")
	}

	if adminPort == 0 {
		t.Fatal("createTestCipherIMServer: admin server did not bind to port")
	}

	publicURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, publicPort)
	adminURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, adminPort)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := cipherServer.Shutdown(ctx); err != nil {
			t.Logf("createTestCipherIMServer cleanup: failed to shutdown server: %v", err)
		}
	})

	return cipherServer, publicURL, adminURL
}
