// Copyright (c) 2025-2026 Justin Cranford.
//

package application

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestStartTLSListener_FailurePaths(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected TLS listener failure")
	tests := []struct {
		name       string
		operations func(*testing.T) startTLSListenerOperations
		wantErr    string
	}{
		{
			name: "core startup",
			operations: func(*testing.T) startTLSListenerOperations {
				return startTLSListenerOperations{
					startCoreFn: func(context.Context, *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*Core, error) {
						return nil, injectedErr
					},
				}
			},
			wantErr: "failed to start application core",
		},
		{
			name: "basic services startup",
			operations: func(t *testing.T) startTLSListenerOperations {
				t.Helper()

				return startTLSListenerOperations{
					startCoreFn: func(context.Context, *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*Core, error) {
						return &Core{}, nil
					},
					startBasicFn: func(context.Context, *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*Basic, error) {
						return nil, injectedErr
					},
				}
			},
			wantErr: "failed to start basic application services",
		},
		{
			name: "TLS subject generation",
			operations: func(*testing.T) startTLSListenerOperations {
				return startTLSListenerOperations{
					startCoreFn: func(context.Context, *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*Core, error) {
						return &Core{}, nil
					},
					startBasicFn: func(context.Context, *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*Basic, error) {
						return &Basic{}, nil
					},
					generateTLSFn: func(*cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings, *Basic) (*cryptoutilSharedCryptoCertificate.Subject, *cryptoutilSharedCryptoCertificate.Subject, error) {
						return nil, nil, injectedErr
					},
				}
			},
			wantErr: "failed to generate TLS server subjects",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tlsListener, err := startTLSListener(&cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{}, tc.operations(t))

			require.Nil(t, tlsListener)
			require.ErrorContains(t, err, tc.wantErr)
			require.ErrorIs(t, err, injectedErr)
		})
	}
}

func TestStartTLSListener_NilSettings(t *testing.T) {
	t.Parallel()

	tlsListener, err := StartTLSListener(nil)

	require.Nil(t, tlsListener)
	require.ErrorContains(t, err, valueSettingsCannotBeNil)
}

func TestGenerateTLSServerSubjectsInMemory_GenerationErrors(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected certificate generation failure")
	tests := []struct {
		name       string
		failOnCall int
		wantErr    string
	}{
		{name: "public subject", failOnCall: 1, wantErr: "failed to create TLS public server certs in memory"},
		{name: "private subject", failOnCall: 2, wantErr: "failed to create TLS private server certs in memory"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			callCount := 0
			settings := &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				TLSPublicIPAddresses:  []string{cryptoutilSharedMagic.IPv4Loopback},
				TLSPrivateIPAddresses: []string{cryptoutilSharedMagic.IPv4Loopback},
			}
			public, private, err := generateTLSServerSubjectsInMemory(settings, &Basic{}, func(*Basic, []string, []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error) {
				callCount++
				if callCount == tc.failOnCall {
					return nil, injectedErr
				}

				return &cryptoutilSharedCryptoCertificate.Subject{}, nil
			})

			require.Nil(t, public)
			require.Nil(t, private)
			require.ErrorContains(t, err, tc.wantErr)
			require.ErrorIs(t, err, injectedErr)
		})
	}
}

func TestGenerateTLSServerSubjectsInMemory_ParseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		publicIPs  []string
		privateIPs []string
		wantErr    string
	}{
		{
			name:      testInvalidPublicIP,
			publicIPs: []string{testInvalidIP},
			wantErr:   testFailedParsePublicTLSServerIPAddresses,
		},
		{
			name:       testInvalidPrivateIP,
			privateIPs: []string{testInvalidIP},
			wantErr:    testFailedParsePrivateTLSServerIPAddresses,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := generateTLSServerSubjectsInMemory(
				&cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
					TLSPublicIPAddresses:  tc.publicIPs,
					TLSPrivateIPAddresses: tc.privateIPs,
				},
				&Basic{},
				func(*Basic, []string, []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					t.Fatal("subject generation must not run when an IP address is invalid")

					return nil, nil
				},
			)

			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestGenerateTLSServerSubjectsToDisk_ErrorPaths(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected disk subject generation failure")
	tests := []struct {
		name        string
		settings    *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
		failPrefix  string
		wantErr     string
		wantCalls   int
		wantSuccess bool
	}{
		{
			name: "both subjects",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				TLSPublicIPAddresses:  []string{cryptoutilSharedMagic.IPv4Loopback},
				TLSPrivateIPAddresses: []string{cryptoutilSharedMagic.IPv4Loopback},
			},
			wantCalls:   2,
			wantSuccess: true,
		},
		{
			name: "invalid public IP",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				TLSPublicIPAddresses: []string{testInvalidIP},
			},
			wantErr: testFailedParsePublicTLSServerIPAddresses,
		},
		{
			name: "invalid private IP",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				TLSPrivateIPAddresses: []string{testInvalidIP},
			},
			wantErr: testFailedParsePrivateTLSServerIPAddresses,
		},
		{
			name: "public subject generation",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				TLSPublicIPAddresses:  []string{cryptoutilSharedMagic.IPv4Loopback},
				TLSPrivateIPAddresses: []string{cryptoutilSharedMagic.IPv4Loopback},
			},
			failPrefix: "tls_public_server_",
			wantErr:    "failed to create TLS public server certs",
			wantCalls:  1,
		},
		{
			name: "private subject generation",
			settings: &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
				TLSPublicIPAddresses:  []string{cryptoutilSharedMagic.IPv4Loopback},
				TLSPrivateIPAddresses: []string{cryptoutilSharedMagic.IPv4Loopback},
			},
			failPrefix: "tls_private_server_",
			wantErr:    "failed to create TLS private server certs",
			wantCalls:  2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			callCount := 0
			public, private, err := generateTLSServerSubjectsToDisk(
				tc.settings,
				&Basic{},
				func(_ *Basic, prefix string, _ []string, _ []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					callCount++

					if prefix == tc.failPrefix {
						return nil, injectedErr
					}

					return &cryptoutilSharedCryptoCertificate.Subject{}, nil
				},
			)

			if tc.wantSuccess {
				require.NoError(t, err)
				require.NotNil(t, public)
				require.NotNil(t, private)
				require.Equal(t, tc.wantCalls, callCount)
			} else {
				require.Nil(t, public)
				require.Nil(t, private)
				require.ErrorContains(t, err, tc.wantErr)

				if tc.failPrefix != "" {
					require.Equal(t, tc.wantCalls, callCount)
					require.ErrorIs(t, err, injectedErr)
				} else {
					require.Zero(t, callCount)
				}
			}
		})
	}
}

func TestGenerateTLSServerSubjectInMemory_CertificateErrors(t *testing.T) {
	t.Parallel()

	basic, err := StartBasic(context.Background(), cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(basic.Shutdown)

	injectedErr := errors.New("injected certificate factory failure")
	tests := []struct {
		name       string
		operations tlsSubjectOperations
		wantErr    string
	}{
		{
			name: "CA subject creation",
			operations: tlsSubjectOperations{
				createCASubjectsFn: func([]*cryptoutilSharedCryptoKeygen.KeyPair, string, time.Duration) ([]*cryptoutilSharedCryptoCertificate.Subject, error) {
					return nil, injectedErr
				},
				createEndEntitySubjectFn: cryptoutilSharedCryptoCertificate.CreateEndEntitySubject,
			},
			wantErr: "failed to create TLS server CA subjects",
		},
		{
			name: "end-entity subject creation",
			operations: tlsSubjectOperations{
				createCASubjectsFn: func([]*cryptoutilSharedCryptoKeygen.KeyPair, string, time.Duration) ([]*cryptoutilSharedCryptoCertificate.Subject, error) {
					return []*cryptoutilSharedCryptoCertificate.Subject{{}}, nil
				},
				createEndEntitySubjectFn: func(
					*cryptoutilSharedCryptoCertificate.Subject,
					*cryptoutilSharedCryptoKeygen.KeyPair,
					string,
					time.Duration,
					[]string,
					[]net.IP,
					[]string,
					[]*url.URL,
					x509.KeyUsage,
					[]x509.ExtKeyUsage,
				) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					return nil, injectedErr
				},
			},
			wantErr: "failed to create TLS server end entity subject",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			subject, err := generateTLSServerSubjectInMemory(basic, nil, nil, tc.operations)

			require.Nil(t, subject)
			require.ErrorContains(t, err, tc.wantErr)
			require.ErrorIs(t, err, injectedErr)
		})
	}
}

func TestGenerateTLSServerSubjectToDisk_ErrorPaths(t *testing.T) {
	t.Parallel()

	basic, err := StartBasic(context.Background(), cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(basic.Shutdown)

	subject, err := GenerateTLSServerSubjectInMemory(basic, nil, nil)
	require.NoError(t, err)

	injectedErr := errors.New("injected TLS file generation failure")
	tests := []struct {
		name    string
		setup   func(*tlsDiskOperations)
		wantErr string
		success bool
	}{
		{
			name:    "successful certificate and encrypted key writes",
			success: true,
		},
		{
			name: "subject generation",
			setup: func(operations *tlsDiskOperations) {
				operations.generateSubjectFn = func(*Basic, []string, []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					return nil, injectedErr
				}
			},
			wantErr: "failed to generate TLS server subject",
		},
		{
			name: "certificate-chain PEM encoding",
			setup: func(operations *tlsDiskOperations) {
				operations.pemEncodesFn = func(any) ([][]byte, error) {
					return nil, injectedErr
				}
			},
			wantErr: "failed to encode certificate chain as PEM",
		},
		{
			name: "certificate file write",
			setup: func(operations *tlsDiskOperations) {
				operations.writeFileFn = func(string, []byte, os.FileMode) error {
					return injectedErr
				}
			},
			wantErr: "failed to write TLS server certificate PEM file",
		},
		{
			name: "private-key PEM encoding",
			setup: func(operations *tlsDiskOperations) {
				operations.pemEncodeFn = func(any) ([]byte, error) {
					return nil, injectedErr
				}
			},
			wantErr: "failed to encode private key as PEM",
		},
		{
			name: "private-key encryption",
			setup: func(operations *tlsDiskOperations) {
				operations.encryptFn = func([]byte) ([]byte, error) {
					return nil, injectedErr
				}
			},
			wantErr: "failed to encrypt TLS server private key PEM",
		},
		{
			name: "encrypted private-key file write",
			setup: func(operations *tlsDiskOperations) {
				operations.writeFileFn = func(filename string, _ []byte, _ os.FileMode) error {
					if filename == "test_private_key.pem" {
						return injectedErr
					}

					return nil
				}
			},
			wantErr: "failed to write encrypted TLS server private key PEM file",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			operations := tlsDiskOperations{
				generateSubjectFn: func(*Basic, []string, []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					return subject, nil
				},
				pemEncodesFn: func(any) ([][]byte, error) {
					return [][]byte{[]byte("certificate")}, nil
				},
				writeFileFn: func(string, []byte, os.FileMode) error { return nil },
				pemEncodeFn: func(any) ([]byte, error) { return []byte("private-key"), nil },
				encryptFn:   func(data []byte) ([]byte, error) { return data, nil },
			}
			if tc.setup != nil {
				tc.setup(&operations)
			}

			result, err := generateTLSServerSubjectToDisk(basic, "test_", nil, nil, operations)

			if tc.success {
				require.NoError(t, err)
				require.Same(t, subject, result)
			} else {
				require.Nil(t, result)
				require.ErrorContains(t, err, tc.wantErr)
				require.ErrorIs(t, err, injectedErr)
			}
		})
	}
}

func TestInitTLSServerCerts_NilSettings(t *testing.T) {
	t.Parallel()

	err := InitTLSServerCerts(nil)

	require.ErrorContains(t, err, "failed to initialize server application core")
}
