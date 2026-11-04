// Copyright (c) 2025-2026 Justin Cranford.
package builder

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"database/sql/driver"
	"encoding/pem"
	"errors"
	"io"
	http "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsFrameworkServiceServerReadiness "cryptoutil/internal/apps-framework/service/server/readiness"
	cryptoutilAppsFrameworkServiceTestHelpDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testCollectorHealthEndpoint           = "https://collector.local/health"
	testCollectorHealthURLValidationError = "must be an HTTPS URL"
)

func TestDatabaseReadinessCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		createTable bool
		wantErr     string
	}{
		{name: "healthy application pool", createTable: true},
		{name: "missing framework schema", wantErr: "no such table"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			db := cryptoutilAppsFrameworkServiceTestHelpDB.NewInMemorySQLiteDB(t)

			if testCase.createTable {
				require.NoError(t, db.Exec("CREATE TABLE barrier_root_keys (uuid TEXT PRIMARY KEY)").Error)
			}

			err := databaseReadinessCheck(db)(context.Background())
			if testCase.wantErr == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, testCase.wantErr)
		})
	}
}

func TestDatabaseReadinessCheckUnavailableDatabase(t *testing.T) {
	t.Parallel()

	t.Run("invalid database handle", func(t *testing.T) {
		t.Parallel()

		err := databaseReadinessCheck(&gorm.DB{Config: &gorm.Config{}})(context.Background())
		require.ErrorContains(t, err, "get application database pool")
	})

	t.Run("canceled database ping", func(t *testing.T) {
		t.Parallel()

		db := cryptoutilAppsFrameworkServiceTestHelpDB.NewInMemorySQLiteDB(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := databaseReadinessCheck(db)(ctx)
		require.ErrorContains(t, err, "ping application database")
	})
}

func TestDatabaseReadinessCheckRejectsSchemaReadErrors(t *testing.T) {
	t.Parallel()

	readErr := errors.New("injected schema read error")
	sqlDB := sql.OpenDB(readinessQueryConnector{nextErr: readErr})

	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	db := &gorm.DB{Config: &gorm.Config{ConnPool: sqlDB}}
	err := databaseReadinessCheck(db)(context.Background())
	require.ErrorIs(t, err, readErr)
	require.ErrorContains(t, err, "read framework schema observation")
}

func TestDatabaseReadinessCheckReportsSchemaCloseErrors(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("injected schema close error")
	sqlDB := sql.OpenDB(readinessQueryConnector{closeErr: closeErr, returnRow: true})

	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	db := &gorm.DB{Config: &gorm.Config{ConnPool: sqlDB}}
	err := databaseReadinessCheck(db)(context.Background())
	require.ErrorIs(t, err, closeErr)
	require.ErrorContains(t, err, "close framework schema observation")
}

func TestNewReadinessMonitorRequiresConfiguredDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		db              *gorm.DB
		config          *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
		nilPublicServer bool
		wantErr         string
	}{
		{name: "database required", wantErr: "database is required"},
		{name: "service settings required", db: &gorm.DB{}, wantErr: "service settings are required"},
		{
			name:            "public server required",
			db:              &gorm.DB{Config: &gorm.Config{}},
			config:          &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{},
			nilPublicServer: true,
			wantErr:         "public server is required",
		},
		{
			name:    "collector health endpoint required",
			db:      &gorm.DB{},
			config:  &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{OTLPEnabled: true},
			wantErr: "collector health URL",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			publicServer := &cryptoutilAppsFrameworkServiceServer.PublicServerBase{}
			if testCase.nilPublicServer {
				publicServer = nil
			}

			_, err := newReadinessMonitor(testCase.db, publicServer, testCase.config, nil)
			require.ErrorContains(t, err, testCase.wantErr)
		})
	}
}

func TestNewReadinessMonitorRejectsInvalidAdditionalChecks(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkServiceTestHelpDB.NewInMemorySQLiteDB(t)
	_, err := newReadinessMonitor(
		db,
		&cryptoutilAppsFrameworkServiceServer.PublicServerBase{},
		&cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{},
		[]cryptoutilAppsFrameworkServiceServerReadiness.Check{
			{Name: "INVALID", ReasonCode: "invalid_reason", Run: func(context.Context) error { return nil }},
		},
	)
	require.ErrorContains(t, err, "sanitized identifiers")
}

func TestNewCollectorHealthCheckRequiresValidConfiguration(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	invalidCAPath := filepath.Join(tempDir, "invalid-ca.pem")
	require.NoError(t, os.WriteFile(invalidCAPath, []byte("not a certificate"), 0o600))

	tests := []struct {
		name     string
		endpoint string
		caPath   string
		certPath string
		keyPath  string
		wantErr  string
	}{
		{
			name:     "HTTP endpoint rejected",
			endpoint: "http://collector.local/health",
			wantErr:  testCollectorHealthURLValidationError,
		},
		{
			name:     "root path rejected",
			endpoint: "https://collector.local/",
			wantErr:  testCollectorHealthURLValidationError,
		},
		{
			name:     "URL credentials rejected",
			endpoint: "https://user:password@collector.local/health",
			wantErr:  testCollectorHealthURLValidationError,
		},
		{
			name:     "empty query rejected",
			endpoint: testCollectorHealthEndpoint + "?",
			wantErr:  testCollectorHealthURLValidationError,
		},
		{
			name:     "missing CA rejected",
			endpoint: testCollectorHealthEndpoint,
			caPath:   filepath.Join(tempDir, "missing-ca.pem"),
			wantErr:  "read collector health CA certificate",
		},
		{
			name:     "invalid CA rejected",
			endpoint: testCollectorHealthEndpoint,
			caPath:   invalidCAPath,
			wantErr:  "parse collector health CA certificate",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := newCollectorHealthCheck(testCase.endpoint, testCase.caPath, testCase.certPath, testCase.keyPath)
			require.ErrorContains(t, err, testCase.wantErr)
			require.NotContains(t, err.Error(), tempDir)
		})
	}
}

func TestCollectorHealthCheckUsesVerifiedMutualTLSAndRequiresHTTP200(t *testing.T) {
	t.Parallel()

	var responseStatus atomic.Int32
	responseStatus.Store(http.StatusOK)

	var truncateResponse atomic.Bool

	clientCertificateObserved := make(chan bool, 8)

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		clientCertificateObserved <- request.TLS != nil && len(request.TLS.PeerCertificates) > 0

		if truncateResponse.Load() {
			writer.Header().Set("Content-Length", "20")
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte("truncated"))

			return
		}

		if responseStatus.Load() == http.StatusFound {
			writer.Header().Set("Location", "/redirect")
		}

		writer.WriteHeader(int(responseStatus.Load()))
	}))
	server.TLS = &tls.Config{ClientAuth: tls.RequireAnyClientCert}
	server.StartTLS()
	t.Cleanup(server.Close)

	certificatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: server.Certificate().Raw,
	})
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(server.TLS.Certificates[0].PrivateKey)
	require.NoError(t, err)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: privateKeyDER})

	tempDir := t.TempDir()
	caPath := filepath.Join(tempDir, "ca.pem")
	certPath := filepath.Join(tempDir, "client.pem")
	keyPath := filepath.Join(tempDir, "client-key.pem")

	require.NoError(t, os.WriteFile(caPath, certificatePEM, 0o600))
	require.NoError(t, os.WriteFile(certPath, certificatePEM, 0o600))
	require.NoError(t, os.WriteFile(keyPath, privateKeyPEM, 0o600))

	healthCheck, err := newCollectorHealthCheckWithTimeout(server.URL+"/health", caPath, certPath, keyPath, cryptoutilSharedMagic.TestTimeout)
	require.NoError(t, err)
	_, err = newCollectorHealthCheck(
		server.URL+"/health",
		caPath,
		filepath.Join(tempDir, "missing-client.pem"),
		filepath.Join(tempDir, "missing-client-key.pem"),
	)
	require.ErrorContains(t, err, "load collector health client certificate")

	ctx, cancel := context.WithTimeout(t.Context(), cryptoutilSharedMagic.TestTimeout)
	defer cancel()

	require.NoError(t, healthCheck(ctx))
	require.True(t, <-clientCertificateObserved, "collector health request must present the configured client certificate")

	responseStatus.Store(http.StatusServiceUnavailable)
	require.ErrorContains(t, healthCheck(ctx), "HTTP 503")
	require.True(t, <-clientCertificateObserved, "unhealthy response must still use the configured mTLS transport")

	responseStatus.Store(http.StatusFound)
	require.ErrorContains(t, healthCheck(ctx), "HTTP 302")
	require.True(t, <-clientCertificateObserved, "redirect response must still use the configured mTLS transport")

	responseStatus.Store(http.StatusOK)
	truncateResponse.Store(true)
	require.ErrorContains(t, healthCheck(ctx), "read collector health response")
	require.True(t, <-clientCertificateObserved, "truncated response must still use the configured mTLS transport")
	truncateResponse.Store(false)

	require.ErrorContains(t, healthCheck(nil), "create collector health request")

	settings := &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
		OTLPEnabled:        true,
		OTLPHealthEndpoint: server.URL + "/health",
		OTLPTLSCAFile:      caPath,
		OTLPTLSCertFile:    certPath,
		OTLPTLSKeyFile:     keyPath,
	}
	db := cryptoutilAppsFrameworkServiceTestHelpDB.NewInMemorySQLiteDB(t)
	monitor, err := newReadinessMonitor(db, &cryptoutilAppsFrameworkServiceServer.PublicServerBase{}, settings, nil)
	require.NoError(t, err)
	require.NotNil(t, monitor)

	server.Close()
	require.ErrorContains(t, healthCheck(ctx), "collector health endpoint request failed")
}

type readinessQueryConnector struct {
	nextErr   error
	closeErr  error
	returnRow bool
}

func (c readinessQueryConnector) Connect(context.Context) (driver.Conn, error) {
	return readinessQueryConnection{connector: c}, nil
}

func (c readinessQueryConnector) Driver() driver.Driver {
	return readinessQueryDriver{}
}

type readinessQueryDriver struct{}

func (readinessQueryDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("database readiness connector must use Connect")
}

type readinessQueryConnection struct {
	connector readinessQueryConnector
}

func (readinessQueryConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("database readiness connection does not support Prepare")
}

func (readinessQueryConnection) Close() error {
	return nil
}

func (readinessQueryConnection) Begin() (driver.Tx, error) {
	return nil, errors.New("database readiness connection does not support transactions")
}

func (readinessQueryConnection) Ping(context.Context) error {
	return nil
}

func (c readinessQueryConnection) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &readinessQueryRows{connector: c.connector}, nil
}

type readinessQueryRows struct {
	connector readinessQueryConnector
	returned  bool
}

func (readinessQueryRows) Columns() []string {
	return []string{"value"}
}

func (r readinessQueryRows) Close() error {
	return r.connector.closeErr
}

func (r *readinessQueryRows) Next(values []driver.Value) error {
	if r.connector.nextErr != nil {
		return r.connector.nextErr
	}

	if r.connector.returnRow && !r.returned {
		r.returned = true
		values[0] = int64(1)

		return nil
	}

	return io.EOF
}
