// Copyright (c) 2025-2026 Justin Cranford.

package test_help_tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	http "net/http"
	"os"
	"path/filepath"
	"testing"

	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps-framework/service/config/tls_generator"

	"github.com/stretchr/testify/require"
)

func TestNewTestTLSSettings_InternalHelperTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		injectErr    error
		wantErrMatch string
	}{
		{name: "success path"},
		{name: "injected generator error", injectErr: errors.New("injected"), wantErrMatch: "generate auto TLS settings"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			generatorFn := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings
			if tc.injectErr != nil {
				generatorFn = func(_ []string, _ []string, _ int) (*cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings, error) {
					return nil, tc.injectErr
				}
			}

			settings, err := newTestTLSSettingsWithGenerator(generatorFn)
			if tc.wantErrMatch != "" {
				require.Nil(t, settings)
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantErrMatch)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, settings)
			require.NotEmpty(t, settings.StaticCertPEM)
			require.NotEmpty(t, settings.StaticKeyPEM)
		})
	}
}

func TestNewTestTLSSettings_Success(t *testing.T) {
	t.Parallel()

	tlsSettings := NewTestTLSSettings(t)
	require.NotNil(t, tlsSettings)
	require.NotEmpty(t, tlsSettings.StaticCertPEM)
	require.NotEmpty(t, tlsSettings.StaticKeyPEM)

	material, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial(tlsSettings)
	require.NoError(t, err)
	require.NotNil(t, material)
	require.NotNil(t, material.Config)
	require.NotNil(t, material.RootCAPool)
}

func TestNewTestTLSSettings_PanicsOnGeneratorError(t *testing.T) {
	t.Parallel()

	errGenerator := func(_ []string, _ []string, _ int) (*cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings, error) {
		return nil, errors.New("injected generator error")
	}

	require.Panics(t, func() {
		_ = mustTestTLSSettings(errGenerator)
	})
}

func TestNewInsecureHTTPSClient_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "returns client with insecure skip verify and keep-alive disabled"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := NewInsecureHTTPSClient(t)
			require.NotNil(t, client)

			transport, ok := client.Transport.(*http.Transport)
			require.True(t, ok)
			require.NotNil(t, transport.TLSClientConfig)
			require.True(t, transport.TLSClientConfig.InsecureSkipVerify)
			require.True(t, transport.DisableKeepAlives)
		})
	}
}

func TestNewMTLSClient_Table(t *testing.T) {
	t.Parallel()

	tlsSettings := NewTestTLSSettings(t)
	require.NotNil(t, tlsSettings)

	tempDir := t.TempDir()
	certPath := filepath.Join(tempDir, "client.crt")
	keyPath := filepath.Join(tempDir, "client.key")

	require.NoError(t, os.WriteFile(certPath, tlsSettings.StaticCertPEM, 0o600))
	require.NoError(t, os.WriteFile(keyPath, tlsSettings.StaticKeyPEM, 0o600))

	material, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial(tlsSettings)
	require.NoError(t, err)
	require.NotNil(t, material)

	tests := []struct {
		name   string
		caPool *x509.CertPool
	}{
		{name: "with root ca pool", caPool: material.RootCAPool},
		{name: "with nil ca pool", caPool: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, newErr := newMTLSClient(certPath, keyPath, tc.caPool)
			require.NoError(t, newErr)
			require.NotNil(t, client)

			transport, ok := client.Transport.(*http.Transport)
			require.True(t, ok)
			require.True(t, transport.DisableKeepAlives)
			require.NotNil(t, transport.TLSClientConfig)
			require.Len(t, transport.TLSClientConfig.Certificates, 1)
			require.Equal(t, uint16(tls.VersionTLS13), transport.TLSClientConfig.MinVersion)
			require.Equal(t, tc.caPool, transport.TLSClientConfig.RootCAs)
		})
	}

	t.Run("public wrapper delegates to helper", func(t *testing.T) {
		t.Parallel()

		client := NewMTLSClient(t, certPath, keyPath, material.RootCAPool)
		require.NotNil(t, client)

		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		require.NotNil(t, transport.TLSClientConfig)
		require.Len(t, transport.TLSClientConfig.Certificates, 1)
	})
}

func TestNewMTLSClient_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		certPath string
		keyPath  string
		wantErr  string
	}{
		{name: "empty cert path", certPath: "", keyPath: "key.pem", wantErr: "certPath must be non-empty"},
		{name: "empty key path", certPath: "cert.pem", keyPath: "", wantErr: "keyPath must be non-empty"},
		{name: "missing key pair files", certPath: "missing.crt", keyPath: "missing.key", wantErr: "load client certificate/key pair"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, err := newMTLSClient(tc.certPath, tc.keyPath, nil)
			require.Nil(t, client)
			require.Error(t, err)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestNewMTLSClient_PanicsOnError(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		_ = NewMTLSClient(t, "missing.crt", "missing.key", nil)
	})
}
