// Copyright (c) 2025-2026 Justin Cranford.
//
//

package listener

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps-framework/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testCertPath                          = "/certs/cert.crt"
	testKeyPath                           = "/certs/cert.key"
	testCAPath                            = "/certs/ca.crt"
	requestPolicyNoCAName                 = "request policy without CA file"
	requireAnyPolicyNoCAName              = "require-any policy without CA file"
	verifyIfGivenNoCAName                 = "verify-if-given without CA file fails"
	requireAndVerifyNoCAName              = "require-and-verify without CA file fails"
	requiresCAFileError                   = "requires a CA file"
	certReadErrorName                     = "cert file read error"
	keyReadErrorName                      = "key file read error"
	invalidCertKeyPairName                = "invalid cert+key pair"
	caReadErrorName                       = "CA file read error"
	caTrustOnlyName                       = "only CA file set - trust material only"
	certKeyOverrideAppliedName            = "cert+key override applied"
	verifyIfGivenPolicyEnforcementName    = "CA cert with verify-if-given policy enables verification when presented"
	requireAndVerifyPolicyEnforcementName = "CA cert with require-and-verify policy enforces verification"
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
			name:           "no admin TLS config - no-op",
			settings:       &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{},
			wantClientAuth: tls.NoClientCert,
		},
		{
			name: requestPolicyNoCAName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequest,
			},
			wantClientAuth: tls.RequestClientCert,
		},
		{
			name: requireAnyPolicyNoCAName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequireAny,
			},
			wantClientAuth: tls.RequireAnyClientCert,
		},
		{
			name: verifyIfGivenNoCAName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyVerifyIfGiven,
			},
			wantErr: requiresCAFileError,
		},
		{
			name: requireAndVerifyNoCAName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequireAndVerify,
			},
			wantErr: requiresCAFileError,
		},
		{
			name: certReadErrorName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCertFile: testCertPath,
				AdminTLSKeyFile:  testKeyPath,
			},
			stubFiles: stubReadFile{},
			wantErr:   "failed to read admin TLS cert file",
		},
		{
			name: keyReadErrorName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCertFile: testCertPath,
				AdminTLSKeyFile:  testKeyPath,
			},
			stubFiles: stubReadFile{
				testCertPath: []byte("fake-cert"),
			},
			wantErr: "failed to read admin TLS key file",
		},
		{
			name: invalidCertKeyPairName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCertFile: testCertPath,
				AdminTLSKeyFile:  testKeyPath,
			},
			stubFiles: stubReadFile{
				testCertPath: []byte("not-valid-cert-pem"),
				testKeyPath:  []byte("not-valid-key-pem"),
			},
			wantErr: "failed to parse admin TLS cert+key pair",
		},
		{
			name: caReadErrorName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCAFile: testCAPath,
			},
			stubFiles: stubReadFile{},
			wantErr:   "failed to read admin TLS CA file",
		},
		{
			name: caTrustOnlyName,
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				AdminTLSCAFile: testCAPath,
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
				AdminTLSCertFile: testCertPath,
				AdminTLSKeyFile:  testKeyPath,
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
				AdminTLSCAFile:       testCAPath,
				AdminTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyVerifyIfGiven,
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
				AdminTLSCAFile:       testCAPath,
				AdminTLSClientPolicy: cryptoutilAppsFrameworkServiceConfig.TLSClientPolicyRequireAndVerify,
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

			require.Equal(t, tc.wantClientAuth, mat.Config.ClientAuth)
		})
	}
}
