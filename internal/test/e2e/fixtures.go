// Copyright (c) 2025 Justin Cranford

//go:build e2e

package test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilCAClient "cryptoutil/api/ca/client"
	cryptoutilOpenapiClient "cryptoutil/api/client"
	cryptoutilJOSEClient "cryptoutil/api/jose/client"
	cryptoutilClient "cryptoutil/internal/apps/sm/kms/client"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestFixture provides shared test infrastructure and utilities.
type TestFixture struct {
	t         *testing.T
	startTime time.Time
	ctx       context.Context
	cancel    context.CancelFunc

	// Logging
	logger *Logger

	// Infrastructure
	infraMgr *InfrastructureManager

	// Service URLs
	sqliteURL    string
	postgres1URL string
	postgres2URL string
	caURL        string
	joseURL      string
	authzURL     string
	idpURL       string
	grafanaURL   string
	otelURL      string

	// API Clients
	sqliteClient    *cryptoutilOpenapiClient.ClientWithResponses
	postgres1Client *cryptoutilOpenapiClient.ClientWithResponses
	postgres2Client *cryptoutilOpenapiClient.ClientWithResponses
	caClient        *cryptoutilCAClient.ClientWithResponses
	joseClient      *cryptoutilJOSEClient.ClientWithResponses
	authzClient     *http.Client
	idpClient       *http.Client

	// Test configuration
	rootCAsPool *x509.CertPool
}

// NewTestFixture creates a new test fixture.
func NewTestFixture(t *testing.T) *TestFixture {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	startTime := time.Now().UTC()

	// Create log file
	logFileName := filepath.Join("..", "..", "..", "workflow-reports", "e2e", fmt.Sprintf("e2e-test-%s.log", startTime.Format("2006-01-02_15-04-05")))

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
		logger:    NewLogger(startTime, logFile),
		infraMgr:  NewInfrastructureManager(startTime, logFile),
	}
}

// Setup initializes the test infrastructure.
func (f *TestFixture) Setup() {
	Log(f.logger, "🚀 Setting up test fixture")

	// Initialize service URLs
	f.initializeServiceURLs()

	// Load test certificates
	f.loadTestCertificates()

	// Setup infrastructure
	f.setupInfrastructure()

	// Initialize API clients
	f.InitializeAPIClients()

	Log(f.logger, "✅ Test fixture setup complete")
}

// Teardown cleans up the test infrastructure.
func (f *TestFixture) Teardown() {
	f.cancel()
	// Note: Log file is NOT closed here to support multiple subtests.
	// The log file will be closed when the test process exits.
	Log(f.logger, "Teardown: Context canceled (log file remains open for subsequent subtests)")
}

// initializeServiceURLs sets up all service URLs.
func (f *TestFixture) initializeServiceURLs() {
	f.sqliteURL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortCryptoutilCompose0) + cryptoutilMagic.DefaultPublicServiceAPIContextPath
	f.postgres1URL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortCryptoutilCompose1) + cryptoutilMagic.DefaultPublicServiceAPIContextPath
	f.postgres2URL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortCryptoutilCompose2) + cryptoutilMagic.DefaultPublicServiceAPIContextPath
	f.caURL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortCAServer)     // CA E2E service uses standardized port
	f.joseURL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortJOSEServer) // JOSE E2E service uses standardized port
	f.authzURL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.IdentityE2EAuthzPublicPort) // Identity AuthZ uses standardized port
	f.idpURL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.IdentityE2EIDPPublicPort)     // Identity IdP uses standardized port
	f.grafanaURL = cryptoutilMagic.URLPrefixLocalhostHTTP + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortGrafana)
	f.otelURL = cryptoutilMagic.URLPrefixLocalhostHTTP + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortOtelCollectorHealth)
}

// loadTestCertificates configures TLS settings for tests.
func (f *TestFixture) loadTestCertificates() {
	// Using InsecureSkipVerify for e2e tests
	f.rootCAsPool = nil
	Log(f.logger, "✅ Test certificates configured (InsecureSkipVerify)")
}

// setupInfrastructure initializes Docker and services.
func (f *TestFixture) setupInfrastructure() {
	// Ensure clean environment
	require.NoError(f.t, f.infraMgr.StopServices(f.ctx), "Failed to ensure clean environment")

	// Start services
	if err := f.infraMgr.StartServices(f.ctx); err != nil {
		// Capture container logs before failing
		logOutputDir := getContainerLogsOutputDir()
		if logErr := CaptureAndZipContainerLogs(f.ctx, f.logger, logOutputDir); logErr != nil {
			Log(f.logger, "⚠️ Failed to capture container logs after startup failure: %v", logErr)
		}

		require.NoError(f.t, err, "Failed to start services")
	}

	// Wait for services to be ready
	Log(f.logger, "⏳ Waiting for Docker Compose services to initialize...")
	time.Sleep(cryptoutilMagic.TestTimeoutDockerComposeInit)

	if err := f.infraMgr.WaitForDockerServicesHealthy(f.ctx); err != nil {
		// Capture container logs before failing
		logOutputDir := getContainerLogsOutputDir()
		if logErr := CaptureAndZipContainerLogs(f.ctx, f.logger, logOutputDir); logErr != nil {
			Log(f.logger, "⚠️ Failed to capture container logs after health check failure: %v", logErr)
		}

		require.NoError(f.t, err, "Failed to wait for docker services healthy")
	}

	if err := f.infraMgr.WaitForServicesReachable(f.ctx); err != nil {
		// Capture container logs before failing
		logOutputDir := getContainerLogsOutputDir()
		if logErr := CaptureAndZipContainerLogs(f.ctx, f.logger, logOutputDir); logErr != nil {
			Log(f.logger, "⚠️ Failed to capture container logs after reachability check failure: %v", logErr)
		}

		require.NoError(f.t, err, "Failed to wait for services reachable")
	}

	Log(f.logger, "✅ All services are ready")
}

// InitializeAPIClients creates API clients for all services.
func (f *TestFixture) InitializeAPIClients() {
	f.sqliteClient = cryptoutilClient.RequireClientWithResponses(f.t, &f.sqliteURL, f.rootCAsPool)
	f.postgres1Client = cryptoutilClient.RequireClientWithResponses(f.t, &f.postgres1URL, f.rootCAsPool)
	f.postgres2Client = cryptoutilClient.RequireClientWithResponses(f.t, &f.postgres2URL, f.rootCAsPool)

	// Initialize CA client
	f.caClient = f.requireCAClientWithResponses(&f.caURL, f.rootCAsPool)

	// Initialize JOSE client
	f.joseClient = f.requireJOSEClientWithResponses(&f.joseURL, f.rootCAsPool)

	// Initialize Identity HTTP clients
	f.authzClient = f.requireHTTPClient(f.rootCAsPool)
	f.idpClient = f.requireHTTPClient(f.rootCAsPool)

	Log(f.logger, "✅ API clients initialized")
}

// requireCAClientWithResponses creates a CA API client with TLS configuration.
func (f *TestFixture) requireCAClientWithResponses(baseURL *string, rootCAsPool *x509.CertPool) *cryptoutilCAClient.ClientWithResponses {
	f.t.Helper()

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if rootCAsPool != nil {
		tlsConfig.RootCAs = rootCAsPool
	} else {
		// Skip verification for self-signed certificates
		tlsConfig.InsecureSkipVerify = true //nolint:gosec // G402: TLS InsecureSkipVerify set true for E2E testing with self-signed certs
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	caClient, err := cryptoutilCAClient.NewClientWithResponses(*baseURL, cryptoutilCAClient.WithHTTPClient(httpClient))
	require.NoError(f.t, err)
	require.NotNil(f.t, caClient)

	return caClient
}

// requireJOSEClientWithResponses creates a JOSE API client with TLS configuration.
func (f *TestFixture) requireJOSEClientWithResponses(baseURL *string, rootCAsPool *x509.CertPool) *cryptoutilJOSEClient.ClientWithResponses {
	f.t.Helper()

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if rootCAsPool != nil {
		tlsConfig.RootCAs = rootCAsPool
	} else {
		// Skip verification for self-signed certificates
		tlsConfig.InsecureSkipVerify = true //nolint:gosec // G402: TLS InsecureSkipVerify set true for E2E testing with self-signed certs
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	joseClient, err := cryptoutilJOSEClient.NewClientWithResponses(*baseURL, cryptoutilJOSEClient.WithHTTPClient(httpClient))
	require.NoError(f.t, err)
	require.NotNil(f.t, joseClient)

	return joseClient
}

// requireHTTPClient creates an HTTP client with TLS configuration.
func (f *TestFixture) requireHTTPClient(rootCAsPool *x509.CertPool) *http.Client {
	f.t.Helper()

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if rootCAsPool != nil {
		tlsConfig.RootCAs = rootCAsPool
	} else {
		// Skip verification for self-signed certificates
		tlsConfig.InsecureSkipVerify = true //nolint:gosec // G402: TLS InsecureSkipVerify set true for E2E testing with self-signed certs
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}

// GetCAClient returns the CA API client.
func (f *TestFixture) GetCAClient() *cryptoutilCAClient.ClientWithResponses {
	return f.caClient
}

// GetJOSEClient returns the JOSE API client.
func (f *TestFixture) GetJOSEClient() *cryptoutilJOSEClient.ClientWithResponses {
	return f.joseClient
}

// GetAuthZClient returns the Identity AuthZ HTTP client.
func (f *TestFixture) GetAuthZClient() *http.Client {
	return f.authzClient
}

// GetIdPClient returns the Identity IdP HTTP client.
func (f *TestFixture) GetIdPClient() *http.Client {
	return f.idpClient
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
	case "ca":
		return f.caURL
	case "jose":
		return f.joseURL
	case "authz":
		return f.authzURL
	case "idp":
		return f.idpURL
	case "grafana":
		return f.grafanaURL
	case "otel":
		return f.otelURL
	default:
		require.Fail(f.t, "Unknown service name", "Service %s not found", instanceName)

		return ""
	}
}
