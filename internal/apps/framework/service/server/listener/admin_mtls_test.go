// Copyright (c) 2025 Justin Cranford
//
//

package listener

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps/framework/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// stubReadFile returns predefined file contents by filename.
type stubReadFile map[string][]byte

func (s stubReadFile) read(name string) ([]byte, error) {
	data, ok := s[name]
	if !ok {
		return nil, fmt.Errorf("stub: file not found: %q", name)
	}

	return data, nil
}

func newTestTLSMaterial(t *testing.T) *cryptoutilAppsFrameworkServiceConfig.TLSMaterial {
	t.Helper()

	tlsCfg, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	mat, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial(tlsCfg)
	require.NoError(t, err)

	return mat
}

// Test cases for applyAdminMTLS.
func TestApplyAdminMTLS(t *testing.T) {
	t.Parallel()

	// Generate real cert+key PEM bytes from auto-TLS for use in table tests.
	mat := newTestTLSMaterial(t)
	require.NotEmpty(t, mat.Config.Certificates)

	// Extract the raw cert and key PEM from the generated material by regenerating (simplest approach).
	tlsCfg, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	// Use the static cert PEM bytes from TLSGeneratedSettings if available.
	certPEM := tlsCfg.StaticCertPEM
	keyPEM := tlsCfg.StaticKeyPEM
	caPEM := tlsCfg.MixedCACertPEM

	tests := []struct {
		name           string
		settings       *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
		stubFiles      stubReadFile
		wantErr        string
		wantClientCA   bool
		wantClientAuth tls.ClientAuthType
	}{
		{
			name:     "no admin TLS config - no-op",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{},
		},
		{
			name: "cert file read error",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCertFile: "/certs/cert.crt",
				AdminTLSKeyFile:  "/certs/cert.key",
			},
			stubFiles: stubReadFile{},
			wantErr:   "failed to read admin TLS cert file",
		},
		{
			name: "key file read error",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCertFile: "/certs/cert.crt",
				AdminTLSKeyFile:  "/certs/cert.key",
			},
			stubFiles: stubReadFile{
				"/certs/cert.crt": []byte("fake-cert"),
			},
			wantErr: "failed to read admin TLS key file",
		},
		{
			name: "invalid cert+key pair",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCertFile: "/certs/cert.crt",
				AdminTLSKeyFile:  "/certs/cert.key",
			},
			stubFiles: stubReadFile{
				"/certs/cert.crt": []byte("not-valid-cert-pem"),
				"/certs/cert.key": []byte("not-valid-key-pem"),
			},
			wantErr: "failed to parse admin TLS cert+key pair",
		},
		{
			name: "CA file read error",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCAFile: "/certs/ca.crt",
			},
			stubFiles: stubReadFile{},
			wantErr:   "failed to read admin TLS CA file",
		},
		{
			name: "only CA file set - enables client auth with no-op cert pool",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCAFile: "/certs/ca.crt",
			},
			stubFiles: stubReadFile{
				"/certs/ca.crt": []byte("not-a-valid-cert-pem-but-decoded-ok"),
			},
			wantClientCA:   true,
			wantClientAuth: tls.RequireAndVerifyClientCert,
		},
	}

	if len(certPEM) > 0 && len(keyPEM) > 0 {
		tests = append(tests, struct {
			name           string
			settings       *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
			stubFiles      stubReadFile
			wantErr        string
			wantClientCA   bool
			wantClientAuth tls.ClientAuthType
		}{
			name: "cert+key override applied",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCertFile: "/certs/cert.crt",
				AdminTLSKeyFile:  "/certs/cert.key",
			},
			stubFiles: stubReadFile{
				"/certs/cert.crt": certPEM,
				"/certs/cert.key": keyPEM,
			},
		})
	}

	if len(caPEM) > 0 {
		tests = append(tests, struct {
			name           string
			settings       *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
			stubFiles      stubReadFile
			wantErr        string
			wantClientCA   bool
			wantClientAuth tls.ClientAuthType
		}{
			name: "CA cert enables client auth",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCAFile: "/certs/ca.crt",
			},
			stubFiles: stubReadFile{
				"/certs/ca.crt": caPEM,
			},
			wantClientCA:   true,
			wantClientAuth: tls.RequireAndVerifyClientCert,
		})
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mat := newTestTLSMaterial(t)

			err := applyAdminMTLS(tc.settings, mat, tc.stubFiles.read)

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)

				return
			}

			require.NoError(t, err)

			if tc.wantClientCA {
				require.NotNil(t, mat.Config.ClientCAs)
				require.IsType(t, &x509.CertPool{}, mat.Config.ClientCAs)
			}

			if tc.wantClientAuth != 0 {
				require.Equal(t, tc.wantClientAuth, mat.Config.ClientAuth)
			}
		})
	}
}
