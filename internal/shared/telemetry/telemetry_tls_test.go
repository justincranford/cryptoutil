// Copyright (c) 2025-2026 Justin Cranford.
//

package telemetry

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	stdoutLogExporter "log/slog"

	"github.com/stretchr/testify/require"
)

func TestBuildOTLPTLSConfig_ErrorAndSuccessPaths(t *testing.T) {
	t.Parallel()

	certFile, keyFile, caFile := writeTestOTLPTLSFiles(t)
	missingFile := filepath.Join(t.TempDir(), "missing.pem")
	malformedCAFile := filepath.Join(t.TempDir(), "malformed-ca.pem")
	require.NoError(t, os.WriteFile(malformedCAFile, []byte("not a certificate"), 0o600))

	tests := []struct {
		name      string
		settings  *TelemetrySettings
		wantError string
	}{
		{
			name: "client certificate missing",
			settings: &TelemetrySettings{
				OTLPTLSCertFile: missingFile,
				OTLPTLSKeyFile:  missingFile,
			},
			wantError: "failed to load OTLP client cert/key",
		},
		{
			name: "CA certificate missing",
			settings: &TelemetrySettings{
				OTLPTLSCertFile: certFile,
				OTLPTLSKeyFile:  keyFile,
				OTLPTLSCAFile:   missingFile,
			},
			wantError: "failed to read OTLP CA cert",
		},
		{
			name: "CA certificate malformed",
			settings: &TelemetrySettings{
				OTLPTLSCertFile: certFile,
				OTLPTLSKeyFile:  keyFile,
				OTLPTLSCAFile:   malformedCAFile,
			},
			wantError: "failed to parse OTLP CA cert PEM",
		},
		{
			name: "valid mTLS material",
			settings: &TelemetrySettings{
				OTLPTLSCertFile: certFile,
				OTLPTLSKeyFile:  keyFile,
				OTLPTLSCAFile:   caFile,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, err := buildOTLPTLSConfig(tc.settings)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				require.Nil(t, config)

				return
			}

			require.NoError(t, err)
			require.Equal(t, uint16(tls.VersionTLS13), config.MinVersion)
			require.False(t, config.InsecureSkipVerify)
			require.Len(t, config.Certificates, 1)
			require.NotNil(t, config.RootCAs)
		})
	}
}

func TestBuildOTLPGRPCCredentials_ErrorAndSuccessPaths(t *testing.T) {
	t.Parallel()

	certFile, keyFile, caFile := writeTestOTLPTLSFiles(t)
	missingFile := filepath.Join(t.TempDir(), "missing.pem")

	tests := []struct {
		name     string
		settings *TelemetrySettings
		wantNil  bool
	}{
		{name: "certificate omitted", settings: &TelemetrySettings{}, wantNil: true},
		{
			name: "invalid certificate material",
			settings: &TelemetrySettings{
				OTLPTLSCertFile: missingFile,
				OTLPTLSKeyFile:  missingFile,
			},
			wantNil: true,
		},
		{
			name: "valid certificate material",
			settings: &TelemetrySettings{
				OTLPTLSCertFile: certFile,
				OTLPTLSKeyFile:  keyFile,
				OTLPTLSCAFile:   caFile,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			credentials := buildOTLPGRPCCredentials(tc.settings)
			if tc.wantNil {
				require.Nil(t, credentials)

				return
			}

			require.NotNil(t, credentials)
		})
	}
}

func TestTelemetryInitializers_ApplyMTLSCredentials(t *testing.T) {
	t.Parallel()

	certFile, keyFile, caFile := writeTestOTLPTLSFiles(t)

	tests := []struct {
		name        string
		endpoint    string
		otlpEnabled bool
	}{
		{name: testTelemetryProtocolHTTPS, endpoint: testHTTPSEndpoint, otlpEnabled: true},
		{name: testTelemetryProtocolGRPCS, endpoint: testGRPCSEndpoint},
		{name: "gRPCS mTLS", endpoint: "grpcs://localhost:4317", otlpEnabled: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			settings := NewTestTelemetrySettings("test_mtls_" + tc.name)
			settings.OTLPEnabled = tc.otlpEnabled
			settings.OTLPEndpoint = tc.endpoint
			settings.OTLPTLSCertFile = certFile
			settings.OTLPTLSKeyFile = keyFile
			settings.OTLPTLSCAFile = caFile

			logger, logsProvider, err := initLogger(ctx, settings)
			require.NoError(t, err)
			require.NotNil(t, logger)

			slogger := stdoutLogExporter.New(stdoutLogExporter.NewTextHandler(io.Discard, nil))
			metricsProvider, err := initMetrics(ctx, slogger, settings, nil)
			require.NoError(t, err)
			require.NotNil(t, metricsProvider)

			tracesProvider, err := initTraces(ctx, slogger, settings, nil)
			require.NoError(t, err)
			require.NotNil(t, tracesProvider)

			if tc.otlpEnabled {
				// Exporter shutdown tries the configured endpoint, where no test collector is listening.
				_ = logsProvider.Shutdown(ctx)
				_ = metricsProvider.Shutdown(ctx)
				_ = tracesProvider.Shutdown(ctx)

				return
			}

			require.NoError(t, logsProvider.Shutdown(ctx))
			require.NoError(t, metricsProvider.Shutdown(ctx))
			require.NoError(t, tracesProvider.Shutdown(ctx))
		})
	}
}

func writeTestOTLPTLSFiles(t *testing.T) (string, string, string) {
	t.Helper()

	now := time.Now().UTC()
	maxSerial := new(big.Int).Lsh(big.NewInt(1), 128)
	caSerial, err := crand.Int(crand.Reader, maxSerial)
	require.NoError(t, err)
	caSerial.Add(caSerial, big.NewInt(1))

	clientSerial, err := crand.Int(crand.Reader, maxSerial)
	require.NoError(t, err)
	clientSerial.Add(clientSerial, big.NewInt(1))

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	caTemplate := &x509.Certificate{
		SerialNumber:          caSerial,
		Subject:               pkix.Name{CommonName: "telemetry test CA"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caCertificateDER, err := x509.CreateCertificate(crand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	clientTemplate := &x509.Certificate{
		SerialNumber: clientSerial,
		Subject:      pkix.Name{CommonName: "telemetry test client"},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientCertificateDER, err := x509.CreateCertificate(crand.Reader, clientTemplate, caTemplate, &clientKey.PublicKey, caKey)
	require.NoError(t, err)

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(clientKey)
	require.NoError(t, err)

	directory := t.TempDir()
	certFile := filepath.Join(directory, "client.crt")
	keyFile := filepath.Join(directory, "client.key")
	caFile := filepath.Join(directory, "ca.crt")

	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: clientCertificateDER}), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: privateKeyDER}), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(caFile, pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: caCertificateDER}), cryptoutilSharedMagic.CacheFilePermissions))

	return certFile, keyFile, caFile
}
