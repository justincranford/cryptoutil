//go:build e2e

package test

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilOpenapiClient "cryptoutil/api/client"
	cryptoutilClient "cryptoutil/internal/client"
	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

// TestFixture provides shared test infrastructure and utilities.
type TestFixture struct {
	t         *testing.T
	startTime time.Time
	ctx       context.Context
	cancel    context.CancelFunc

	// Logging
	logFile *os.File

	// Infrastructure
	infraMgr *InfrastructureManager

	// Service URLs
	sqliteURL    string
	postgres1URL string
	postgres2URL string
	grafanaURL   string
	otelURL      string

	// API Clients
	sqliteClient    *cryptoutilOpenapiClient.ClientWithResponses
	postgres1Client *cryptoutilOpenapiClient.ClientWithResponses
	postgres2Client *cryptoutilOpenapiClient.ClientWithResponses

	// Test configuration
	rootCAsPool *x509.CertPool
}

// NewTestFixture creates a new test fixture.
func NewTestFixture(t *testing.T) *TestFixture {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	startTime := time.Now()

	// Create log file
	logFileName := filepath.Join("..", "..", "test", "e2e", "e2e-reports", fmt.Sprintf("e2e-test-%s.log", startTime.Format("2006-01-02_15-04-05")))

	// Ensure the directory exists
	logDir := filepath.Dir(logFileName)
	if err := os.MkdirAll(logDir, cryptoutilMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute); err != nil {
		t.Fatalf("Failed to create log directory %s: %v", logDir, err)
	}

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, cryptoutilMagic.FilePermOwnerReadWriteGroupRead)
	if err != nil {
		t.Fatalf("Failed to create log file %s: %v", logFileName, err)
	}

	return &TestFixture{
		t:         t,
		startTime: startTime,
		ctx:       ctx,
		cancel:    cancel,
		logFile:   logFile,
		infraMgr:  NewInfrastructureManager(startTime, logFile),
	}
}

// Setup initializes the test infrastructure.
func (f *TestFixture) Setup() {
	f.log("🚀 Setting up test fixture")

	// Initialize service URLs
	f.initializeServiceURLs()

	// Load test certificates
	f.loadTestCertificates()

	// Setup infrastructure
	f.setupInfrastructure()

	f.log("✅ Test fixture setup complete")
}

// Teardown cleans up the test infrastructure.
func (f *TestFixture) Teardown() {
	f.log("🧹 Tearing down test fixture")

	if f.cancel != nil {
		f.cancel()
	}

	// Stop infrastructure
	if err := f.infraMgr.TeardownServices(context.Background()); err != nil {
		f.log("⚠️ Warning: failed to stop infrastructure: %v", err)
	}

	// Close log file
	if f.logFile != nil {
		if err := f.logFile.Close(); err != nil {
			f.log("⚠️ Warning: failed to close log file: %v", err)
		}
	}

	f.log("✅ Test fixture teardown complete")
}

// initializeServiceURLs sets up all service URLs.
func (f *TestFixture) initializeServiceURLs() {
	f.sqliteURL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortCryptoutilCompose0)
	f.postgres1URL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortCryptoutilCompose1)
	f.postgres2URL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortCryptoutilCompose2)
	f.grafanaURL = cryptoutilMagic.URLPrefixLocalhostHTTP + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortGrafana)
	f.otelURL = cryptoutilMagic.URLPrefixLocalhostHTTP + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortInternalMetrics)
}

// loadTestCertificates configures TLS settings for tests.
func (f *TestFixture) loadTestCertificates() {
	// Using InsecureSkipVerify for e2e tests
	f.rootCAsPool = nil
	f.log("✅ Test certificates configured (InsecureSkipVerify)")
}

// setupInfrastructure initializes Docker and services.
func (f *TestFixture) setupInfrastructure() {
	// Ensure clean environment
	require.NoError(f.t, f.infraMgr.StopServices(f.ctx), "Failed to ensure clean environment")

	// Start services
	require.NoError(f.t, f.infraMgr.StartServices(f.ctx), "Failed to start services")

	// Wait for services to be ready
	require.NoError(f.t, f.infraMgr.WaitForServicesReady(f.ctx), "Failed to wait for services ready")
}

// InitializeAPIClients creates API clients for all services.
func (f *TestFixture) InitializeAPIClients() {
	f.sqliteClient = cryptoutilClient.RequireClientWithResponses(f.t, &f.sqliteURL, f.rootCAsPool)
	f.postgres1Client = cryptoutilClient.RequireClientWithResponses(f.t, &f.postgres1URL, f.rootCAsPool)
	f.postgres2Client = cryptoutilClient.RequireClientWithResponses(f.t, &f.postgres2URL, f.rootCAsPool)
	f.log("✅ API clients initialized")
}

// GetClient returns the API client for the specified instance.
func (f *TestFixture) GetClient(instanceName string) *cryptoutilOpenapiClient.ClientWithResponses {
	switch instanceName {
	case cryptoutilMagic.TestDatabaseSQLite:
		return f.sqliteClient
	case cryptoutilMagic.TestDatabasePostgres1:
		return f.postgres1Client
	case cryptoutilMagic.TestDatabasePostgres2:
		return f.postgres2Client
	default:
		require.Fail(f.t, "Unknown instance name", "Instance %s not found", instanceName)

		return nil
	}
}

// GetServiceURL returns the service URL for the specified instance.
func (f *TestFixture) GetServiceURL(instanceName string) string {
	switch instanceName {
	case cryptoutilMagic.TestDatabaseSQLite:
		return f.sqliteURL
	case cryptoutilMagic.TestDatabasePostgres1:
		return f.postgres1URL
	case cryptoutilMagic.TestDatabasePostgres2:
		return f.postgres2URL
	case "grafana":
		return f.grafanaURL
	case "otel":
		return f.otelURL
	default:
		require.Fail(f.t, "Unknown service name", "Service %s not found", instanceName)

		return ""
	}
}

// log provides structured logging for the fixture.
func (f *TestFixture) log(format string, args ...interface{}) {
	message := fmt.Sprintf("[%s] [%v] %s\n",
		time.Now().Format("15:04:05"),
		time.Since(f.startTime).Round(time.Second),
		fmt.Sprintf(format, args...))

	// Write to console
	fmt.Print(message)

	// Write to log file if available
	if f.logFile != nil {
		if _, err := f.logFile.WriteString(message); err != nil {
			// If we can't write to the log file, at least write to console
			fmt.Printf("⚠️ Failed to write to log file: %v\n", err)
		}
	}
}
