// Copyright (c) 2025-2026 Justin Cranford.
//
//

package listener

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps-framework/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Test cases for applyPublicMTLS.
func TestApplyPublicMTLS(t *testing.T) {
	t.Parallel()

	// Generate real cert+key PEM bytes from auto-TLS for use in table tests.
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
			name:           "no public TLS config - no-op",
			settings:       &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{},
			wantClientAuth: tls.NoClientCert,
		},
		{
			name: "request policy without CA file",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequest,
			},
			wantClientAuth: tls.RequestClientCert,
		},
		{
			name: "require-any policy without CA file",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequireAny,
			},
			wantClientAuth: tls.RequireAnyClientCert,
		},
		{
			name: "verify-if-given without CA file fails",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyVerifyIfGiven,
			},
			wantErr: "requires a CA file",
		},
		{
			name: "require-and-verify without CA file fails",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequireAndVerify,
			},
			wantErr: "requires a CA file",
		},
		{
			name: "cert file read error",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCertFile: "/certs/cert.crt",
				PublicTLSKeyFile:  "/certs/cert.key",
			},
			stubFiles: stubReadFile{},
			wantErr:   "failed to read public TLS cert file",
		},
		{
			name: "key file read error",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCertFile: "/certs/cert.crt",
				PublicTLSKeyFile:  "/certs/cert.key",
			},
			stubFiles: stubReadFile{
				"/certs/cert.crt": []byte("fake-cert"),
			},
			wantErr: "failed to read public TLS key file",
		},
		{
			name: "invalid cert+key pair",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCertFile: "/certs/cert.crt",
				PublicTLSKeyFile:  "/certs/cert.key",
			},
			stubFiles: stubReadFile{
				"/certs/cert.crt": []byte("not-valid-cert-pem"),
				"/certs/cert.key": []byte("not-valid-key-pem"),
			},
			wantErr: "failed to parse public TLS cert+key pair",
		},
		{
			name: "CA file read error",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCAFile: "/certs/ca.crt",
			},
			stubFiles: stubReadFile{},
			wantErr:   "failed to read public TLS CA file",
		},
		{
			name: "only CA file set - trust material only",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCAFile: "/certs/ca.crt",
			},
			stubFiles: stubReadFile{
				"/certs/ca.crt": []byte("not-a-valid-cert-pem-but-decoded-ok"),
			},
			wantClientCA:   true,
			wantClientAuth: tls.NoClientCert,
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
				PublicTLSCertFile: "/certs/cert.crt",
				PublicTLSKeyFile:  "/certs/cert.key",
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
			name: "CA cert with verify-if-given policy enables verification when presented",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCAFile:       "/certs/ca.crt",
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyVerifyIfGiven,
			},
			stubFiles: stubReadFile{
				"/certs/ca.crt": caPEM,
			},
			wantClientCA:   true,
			wantClientAuth: tls.VerifyClientCertIfGiven,
		})
		tests = append(tests, struct {
			name           string
			settings       *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
			stubFiles      stubReadFile
			wantErr        string
			wantClientCA   bool
			wantClientAuth tls.ClientAuthType
		}{
			name: "CA cert with require-and-verify policy enforces verification",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCAFile:       "/certs/ca.crt",
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequireAndVerify,
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

			err := applyPublicMTLS(tc.settings, mat, tc.stubFiles.read)

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

			require.Equal(t, tc.wantClientAuth, mat.Config.ClientAuth)
		})
	}
}
