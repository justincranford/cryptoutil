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
			name: requestPolicyNoCAName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequest,
			},
			wantClientAuth: tls.RequestClientCert,
		},
		{
			name: requireAnyPolicyNoCAName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequireAny,
			},
			wantClientAuth: tls.RequireAnyClientCert,
		},
		{
			name: verifyIfGivenNoCAName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyVerifyIfGiven,
			},
			wantErr: requiresCAFileError,
		},
		{
			name: requireAndVerifyNoCAName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequireAndVerify,
			},
			wantErr: requiresCAFileError,
		},
		{
			name: certReadErrorName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCertFile: testCertPath,
				PublicTLSKeyFile:  testKeyPath,
			},
			stubFiles: stubReadFile{},
			wantErr:   "failed to read public TLS cert file",
		},
		{
			name: keyReadErrorName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCertFile: testCertPath,
				PublicTLSKeyFile:  testKeyPath,
			},
			stubFiles: stubReadFile{
				testCertPath: []byte("fake-cert"),
			},
			wantErr: "failed to read public TLS key file",
		},
		{
			name: invalidCertKeyPairName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCertFile: testCertPath,
				PublicTLSKeyFile:  testKeyPath,
			},
			stubFiles: stubReadFile{
				testCertPath: []byte("not-valid-cert-pem"),
				testKeyPath:  []byte("not-valid-key-pem"),
			},
			wantErr: "failed to parse public TLS cert+key pair",
		},
		{
			name: caReadErrorName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCAFile: testCAPath,
			},
			stubFiles: stubReadFile{},
			wantErr:   "failed to read public TLS CA file",
		},
		{
			name: caTrustOnlyName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCAFile: testCAPath,
			},
			stubFiles: stubReadFile{
				testCAPath: []byte("not-a-valid-cert-pem-but-decoded-ok"),
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
			name: certKeyOverrideAppliedName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCertFile: testCertPath,
				PublicTLSKeyFile:  testKeyPath,
			},
			stubFiles: stubReadFile{
				testCertPath: certPEM,
				testKeyPath:  keyPEM,
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
			name: verifyIfGivenPolicyEnforcementName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCAFile:       testCAPath,
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyVerifyIfGiven,
			},
			stubFiles: stubReadFile{
				testCAPath: caPEM,
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
			name: requireAndVerifyPolicyEnforcementName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				PublicTLSCAFile:       testCAPath,
				PublicTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequireAndVerify,
			},
			stubFiles: stubReadFile{
				testCAPath: caPEM,
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
