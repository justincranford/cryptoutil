// Copyright (c) 2025 Justin Cranford
//

package tls_test

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps/framework/tls"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestNewGenerator_NilInputs verifies that NewGenerator rejects nil arguments.
func TestNewGenerator_NilInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		nilCtx  bool
		wantErr string
	}{
		{name: "nil context", nilCtx: true, wantErr: "context must be non-nil"},
		{name: "nil telemetry service", nilCtx: false, wantErr: "telemetry service must be non-nil"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var ctx context.Context
			if !tc.nilCtx {
				ctx = context.Background()
			}

			gen, err := cryptoutilAppsFrameworkTls.NewGenerator(ctx, nil)
			require.Nil(t, gen)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

// TestGenerate_SkeletonTemplate_DirCount verifies that Generate creates exactly 82 leaf
// certificate directories for the skeleton-template PS-ID with 2 realms.
// The count is: 28 shared + 54 per-PS-ID = 82 (see docs/tls-structure.md).
func TestGenerate_SkeletonTemplate_DirCount(t *testing.T) {
	t.Parallel()

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		os.MkdirAll, // real mkdirAll to actually create dirs
		stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	tmpDir := t.TempDir()

	require.NoError(t, gen.Generate(cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, tmpDir))

	basePath := filepath.Join(tmpDir, cryptoutilSharedMagic.OTLPServiceSkeletonTemplate)

	var dirCount int

	require.NoError(t, filepath.WalkDir(basePath, func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() && path != basePath {
			dirCount++
		}

		return nil
	}))

	const expectedDirs = 82

	require.Equal(t, expectedDirs, dirCount,
		"expected %d cert dirs for skeleton-template with 2 realms (28 shared + 54 per-PSID)", expectedDirs)
}

// TestGenerate_BasepathIsFile verifies that Generate returns an error when basePath
// (targetDir/tierID) already exists as a file rather than a directory.
func TestGenerate_BasepathIsFile(t *testing.T) {
	t.Parallel()

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	tmpDir := t.TempDir()
	// Create a FILE at the path where the tier subdirectory should be created.
	filePath := filepath.Join(tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS)
	require.NoError(t, os.WriteFile(filePath, []byte("blocking file"), cryptoutilSharedMagic.CacheFilePermissions))

	err := gen.Generate(cryptoutilSharedMagic.OTLPServiceSMKMS, tmpDir)
	require.ErrorContains(t, err, "not a directory")
}

// TestGenerateCAChain_Errors verifies error propagation from generateCAChain for
// root CA creation failure and issuing CA creation failure.
func TestGenerateCAChain_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		failOnCall  int // createCAFn call number to inject failure (1-indexed)
		wantErrFrag string
	}{
		{name: "root CA create fails", failOnCall: 1, wantErrFrag: "root CA public-global"},
		{name: "issuing CA create fails", failOnCall: 2, wantErrFrag: "issuing CA public-global"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var count atomic.Int32

			gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
				stubMkdirAll, stubWriteFile,
				func(issuer *cryptoutilSharedCryptoCertificate.Subject, issuerKey any, name string, kp *cryptoutilSharedCryptoKeygen.KeyPair, dur time.Duration, maxPath int) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					if int(count.Add(1)) == tc.failOnCall {
						return nil, fmt.Errorf("injected createCA error")
					}

					return stubCreateCA(issuer, issuerKey, name, kp, dur, maxPath)
				},
				stubCreateLeaf, stubGetKeyPair, stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
			)

			_, err := gen.ExportedGenerateCAChain(t.TempDir(), "public-global", "https-server-ca")
			require.ErrorContains(t, err, tc.wantErrFrag)
		})
	}
}

// TestGenerateLeafDirs_CreateLeafError verifies that all three leaf-generation helpers
// correctly propagate createLeafFn errors.
func TestGenerateLeafDirs_CreateLeafError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		runTest func(t *testing.T, gen *cryptoutilAppsFrameworkTls.Generator, basePath string, issuer *cryptoutilSharedCryptoCertificate.Subject)
	}{
		{
			name: "server leaf create fails",
			runTest: func(t *testing.T, gen *cryptoutilAppsFrameworkTls.Generator, basePath string, issuer *cryptoutilSharedCryptoCertificate.Subject) {
				t.Helper()

				err := gen.ExportedGenerateServerLeafDir(basePath, "test-server-keystore", issuer,
					[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)})
				require.ErrorContains(t, err, "server leaf test-server-keystore")
			},
		},
		{
			name: "client leaf create fails",
			runTest: func(t *testing.T, gen *cryptoutilAppsFrameworkTls.Generator, basePath string, issuer *cryptoutilSharedCryptoCertificate.Subject) {
				t.Helper()

				err := gen.ExportedGenerateClientLeafDir(basePath, "test-client-keystore", issuer)
				require.ErrorContains(t, err, "client leaf test-client-keystore")
			},
		},
		{
			name: "mutual leaf create fails",
			runTest: func(t *testing.T, gen *cryptoutilAppsFrameworkTls.Generator, basePath string, issuer *cryptoutilSharedCryptoCertificate.Subject) {
				t.Helper()

				err := gen.ExportedGenerateMutualLeafDir(basePath, "test-mutual-keystore", issuer,
					[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault}, []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)})
				require.ErrorContains(t, err, "mutual leaf test-mutual-keystore")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
				stubMkdirAll, stubWriteFile, stubCreateCA,
				func(issuer *cryptoutilSharedCryptoCertificate.Subject, kp *cryptoutilSharedCryptoKeygen.KeyPair, name string, dur time.Duration, dns []string, ips []net.IP, emails []string, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					return nil, fmt.Errorf("injected createLeaf error")
				},
				stubGetKeyPair, stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
			)

			issuer := makeStubSubject(t)
			tc.runTest(t, gen, t.TempDir(), issuer)
		})
	}
}

// TestWriteKeystore_Errors verifies error paths inside writeKeystore.
func TestWriteKeystore_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		encodePKCS12Fn  func(any, *x509.Certificate, []*x509.Certificate) ([]byte, error)
		writeFileFn     func(string, []byte, os.FileMode) error
		wantErrFragment string
	}{
		{
			name: "encodePKCS12 fails",
			encodePKCS12Fn: func(_ any, _ *x509.Certificate, _ []*x509.Certificate) ([]byte, error) {
				return nil, fmt.Errorf("injected pkcs12 error")
			},
			writeFileFn:     stubWriteFile,
			wantErrFragment: "pkcs12 encode",
		},
		{
			name:           "write .p12 fails",
			encodePKCS12Fn: stubEncodePKCS12,
			writeFileFn: failWriteFileAtCall(1,
				func(path string, _ []byte, _ os.FileMode) error {
					return fmt.Errorf("injected write error for %s", filepath.Base(path))
				}),
			wantErrFragment: "write .p12",
		},
		{
			name:           "write .crt fails",
			encodePKCS12Fn: stubEncodePKCS12,
			writeFileFn: failWriteFileAtCall(2,
				func(path string, _ []byte, _ os.FileMode) error {
					return fmt.Errorf("injected write error for %s", filepath.Base(path))
				}),
			wantErrFragment: "write .crt",
		},
		{
			name:           "write .key fails",
			encodePKCS12Fn: stubEncodePKCS12,
			writeFileFn: failWriteFileAtCall(3,
				func(path string, _ []byte, _ os.FileMode) error {
					return fmt.Errorf("injected write error for %s", filepath.Base(path))
				}),
			wantErrFragment: "write .key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
				stubMkdirAll, tc.writeFileFn, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
				tc.encodePKCS12Fn, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
			)

			err := gen.ExportedWriteKeystore(t.TempDir(), "test-ks", makeStubSubject(t))
			require.ErrorContains(t, err, tc.wantErrFragment)
		})
	}
}

// TestWriteTruststore_Errors verifies error paths inside writeTruststore.
func TestWriteTruststore_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		encodeTrustPKCS12Fn func([]*x509.Certificate) ([]byte, error)
		writeFileFn         func(string, []byte, os.FileMode) error
		wantErrFragment     string
	}{
		{
			name: "encodeTrustPKCS12 fails",
			encodeTrustPKCS12Fn: func(_ []*x509.Certificate) ([]byte, error) {
				return nil, fmt.Errorf("injected trust pkcs12 error")
			},
			writeFileFn:     stubWriteFile,
			wantErrFragment: "pkcs12 trust encode",
		},
		{
			name:                "write .p12 fails",
			encodeTrustPKCS12Fn: stubEncodeTrustPKCS12,
			writeFileFn: failWriteFileAtCall(1,
				func(path string, _ []byte, _ os.FileMode) error {
					return fmt.Errorf("injected write error for %s", filepath.Base(path))
				}),
			wantErrFragment: "write .p12",
		},
		{
			name:                "write .crt fails",
			encodeTrustPKCS12Fn: stubEncodeTrustPKCS12,
			writeFileFn: failWriteFileAtCall(2,
				func(path string, _ []byte, _ os.FileMode) error {
					return fmt.Errorf("injected write error for %s", filepath.Base(path))
				}),
			wantErrFragment: "write .crt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
				stubMkdirAll, tc.writeFileFn, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
				stubEncodePKCS12, tc.encodeTrustPKCS12Fn, stubGetRealmsForPSID,
			)

			err := gen.ExportedWriteTruststore(t.TempDir(), "test-ts", makeStubSubject(t))
			require.ErrorContains(t, err, tc.wantErrFragment)
		})
	}
}

// TestGenerate_PSIDCerts_GetRealmsError verifies that getRealmsForPSIDFn errors are
// propagated correctly out of generatePSIDCerts.
func TestGenerate_PSIDCerts_GetRealmsError(t *testing.T) {
	t.Parallel()

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12,
		func(_ string) ([]string, error) { return nil, fmt.Errorf("injected realms error") },
	)

	err := gen.Generate(cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, t.TempDir())
	require.ErrorContains(t, err, "injected realms error")
}

// TestReadRealmsForPSID_Scenarios verifies readRealmsForPSID against various YAML inputs.
func TestReadRealmsForPSID_Scenarios(t *testing.T) {
	t.Parallel()

	const validPSID = "test-ps-id"

	tests := []struct {
		name        string
		registryFn  func(t *testing.T) string // returns path to registry file
		psID        string
		wantRealms  []string
		wantErrFrag string
	}{
		{
			name: "success: finds realms",
			registryFn: func(t *testing.T) string {
				t.Helper()

				return writeRegistryFile(t, `product_services:
  - ps_id: test-ps-id
    realms:
      - name: file
      - name: db
`)
			},
			psID:       validPSID,
			wantRealms: []string{"file", "db"},
		},
		{
			name: "file not found",
			registryFn: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "missing.yaml")
			},
			psID:        validPSID,
			wantErrFrag: "failed to read registry file",
		},
		{
			name: "invalid YAML",
			registryFn: func(t *testing.T) string {
				t.Helper()

				return writeRegistryFile(t, ": : invalid yaml {{{")
			},
			psID:        validPSID,
			wantErrFrag: "failed to parse registry file",
		},
		{
			name: "PS-ID not found",
			registryFn: func(t *testing.T) string {
				t.Helper()

				return writeRegistryFile(t, `product_services:
  - ps_id: other-ps-id
    realms:
      - name: file
`)
			},
			psID:        validPSID,
			wantErrFrag: "PS-ID",
		},
		{
			name: "PS-ID has empty realms",
			registryFn: func(t *testing.T) string {
				t.Helper()

				return writeRegistryFile(t, `product_services:
  - ps_id: test-ps-id
    realms: []
`)
			},
			psID:        validPSID,
			wantErrFrag: "no realms configured",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := tc.registryFn(t)

			realms, err := cryptoutilAppsFrameworkTls.ExportedReadRealmsForPSID(path, tc.psID)

			if tc.wantErrFrag != "" {
				require.ErrorContains(t, err, tc.wantErrFrag)
				require.Nil(t, realms)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantRealms, realms)
			}
		})
	}
}

// TestDefaultRealms verifies that defaultRealms returns the expected two realm names.
func TestDefaultRealms(t *testing.T) {
	t.Parallel()

	realms := cryptoutilAppsFrameworkTls.ExportedDefaultRealms()
	require.Equal(t, []string{"file", "db"}, realms)
}

// --- helpers ---

// makeStubSubject creates a minimal Subject with a valid self-signed certificate for testing.
func makeStubSubject(t *testing.T) *cryptoutilSharedCryptoCertificate.Subject {
	t.Helper()

	kp := stubGetKeyPair()
	subj, err := stubCreateCA(nil, nil, "test-subject", kp, 0, 0)
	require.NoError(t, err)

	return subj
}

// failWriteFileAtCall returns a writeFileFn that fails on the nth call with the given errFn.
func failWriteFileAtCall(n int32, errFn func(path string, data []byte, mode os.FileMode) error) func(string, []byte, os.FileMode) error {
	var count atomic.Int32

	return func(path string, data []byte, mode os.FileMode) error {
		if count.Add(1) == n {
			return errFn(path, data, mode)
		}

		return nil
	}
}

// writeRegistryFile writes YAML content to a temp file and returns its path.
func writeRegistryFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "registry.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	return path
}

// TestInit_WrapperStatements covers the single return statement in each Init* wrapper
// by providing a missing --output-dir (initRun exits early, no production seams invoked).
func TestInit_WrapperStatements(t *testing.T) {
	t.Parallel()

	missingOutputDir := []string{"--domain=skeleton-template"} // --output-dir deliberately absent

	require.Equal(t, 1, cryptoutilAppsFrameworkTls.Init(missingOutputDir, nil, io.Discard, io.Discard))
	require.Equal(t, 1, cryptoutilAppsFrameworkTls.InitForSuite(cryptoutilSharedMagic.DefaultOTLPServiceDefault, missingOutputDir, io.Discard, io.Discard))
	require.Equal(t, 1, cryptoutilAppsFrameworkTls.InitForProduct("sm", missingOutputDir, io.Discard, io.Discard))
	require.Equal(t, 1, cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, missingOutputDir, io.Discard, io.Discard))
}

// TestGenerate_EmptyBasePath_Succeeds covers the validateTargetDir path where basePath
// already exists as an empty directory (should be treated as valid).
func TestGenerate_EmptyBasePath_Succeeds(t *testing.T) {
	t.Parallel()

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	tmpDir := t.TempDir()
	// Pre-create an empty basePath directory.
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS), cryptoutilSharedMagic.CICDTempDirPermissions))

	require.NoError(t, gen.Generate(cryptoutilSharedMagic.OTLPServiceSMKMS, tmpDir))
}

// TestGenerateCAChain_WriteErrors covers writeTruststore and issuing-CA write error paths
// inside generateCAChain that are not reachable via generateCAChain's createCA-failure path.
func TestGenerateCAChain_WriteErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		encodePKCS12Fn      func(any, *x509.Certificate, []*x509.Certificate) ([]byte, error)
		encodeTrustPKCS12Fn func([]*x509.Certificate) ([]byte, error)
		wantErrFrag         string
	}{
		{
			name:                "root CA truststore encode fails",
			encodePKCS12Fn:      stubEncodePKCS12,
			encodeTrustPKCS12Fn: func(_ []*x509.Certificate) ([]byte, error) { return nil, fmt.Errorf("injected") },
			wantErrFrag:         "root CA truststore",
		},
		{
			name:                "issuing CA keystore encode fails",
			encodePKCS12Fn:      failNthEncode(2),
			encodeTrustPKCS12Fn: stubEncodeTrustPKCS12,
			wantErrFrag:         "issuing CA keystore",
		},
		{
			name:                "issuing CA truststore encode fails",
			encodePKCS12Fn:      stubEncodePKCS12,
			encodeTrustPKCS12Fn: failNthTrustEncode(2),
			wantErrFrag:         "issuing CA truststore",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
				stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
				tc.encodePKCS12Fn, tc.encodeTrustPKCS12Fn, stubGetRealmsForPSID,
			)

			_, err := gen.ExportedGenerateCAChain(t.TempDir(), "public-global", "https-server-ca")
			require.ErrorContains(t, err, tc.wantErrFrag)
		})
	}
}

// TestGenerate_SharedCAs_ErrorBranches covers all error-return statements in generateSharedCAs
// by injecting failures at precise createCA and createLeaf call positions.
//
// Call sequence in generateSharedCAs (skeleton-template, 1 PS-ID):
//
//	createCA:  #1=cat1-root, #2=cat1-issuing, #3=cat8-grafana-root, #4=cat8-grafana-issuing,
//	           #5=cat8-otel-root, #6=cat8-otel-issuing, #7=cat10-root, #8=cat10-issuing,
//	           #9=cat12-root, #10=cat12-issuing
//	createLeaf: #1=cat2-grafana, #2=cat2-otel, #3=cat9-grafana-admin, #4=cat9-otel-admin,
//	            #5=cat11-leader, #6=cat11-follower, #7=cat13-leader, #8=cat13-follower
func TestGenerate_SharedCAs_ErrorBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		failCAAt    int32 // 0=never; Nth createCA call fails
		failLeafAt  int32 // 0=never; Nth createLeaf call fails
		wantErrFrag string
	}{
		{"cat1 CA error", 1, 0, "cat1 global server CA"},
		{"cat2 grafana leaf error", 0, 1, "cat2 grafana server leaf"},
		{"cat2 otel leaf error", 0, 2, "cat2 otel server leaf"},
		{"cat8 grafana CA error", 3, 0, "cat8 grafana client CA"},
		{"cat8 otel CA error", cryptoutilSharedMagic.DefaultEmailOTPRateLimit, 0, "cat8 otel client CA"},
		{"cat9 grafana admin error", 0, 3, "cat9 grafana admin client"},
		{"cat9 otel admin error", 0, 4, "cat9 otel admin client"},
		{"cat10 postgres server CA error", cryptoutilSharedMagic.GitRecentActivityDays, 0, "cat10 postgres server CA"},
		{"cat11 leader leaf error", 0, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, "cat11 postgres leader server leaf"},
		{"cat11 follower leaf error", 0, cryptoutilSharedMagic.Utf8EnforceWorkerPoolSize, "cat11 postgres follower server leaf"},
		{"cat12 postgres client CA error", 9, 0, "cat12 postgres client CA"},
		{"cat13 leader client error", 0, cryptoutilSharedMagic.GitRecentActivityDays, "cat13 postgres leader replication client"},
		{"cat13 follower client error", 0, cryptoutilSharedMagic.GitShortHashLength, "cat13 postgres follower replication client"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var caCount, leafCount atomic.Int32

			gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
				stubMkdirAll, stubWriteFile,
				func(issuer *cryptoutilSharedCryptoCertificate.Subject, issuerKey any, name string, kp *cryptoutilSharedCryptoKeygen.KeyPair, dur time.Duration, maxPath int) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					if n := caCount.Add(1); tc.failCAAt > 0 && n == tc.failCAAt {
						return nil, fmt.Errorf("injected CA error at call %d", n)
					}

					return stubCreateCA(issuer, issuerKey, name, kp, dur, maxPath)
				},
				func(issuer *cryptoutilSharedCryptoCertificate.Subject, kp *cryptoutilSharedCryptoKeygen.KeyPair, name string, dur time.Duration, dns []string, ips []net.IP, emails []string, ku x509.KeyUsage, eku []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					if n := leafCount.Add(1); tc.failLeafAt > 0 && n == tc.failLeafAt {
						return nil, fmt.Errorf("injected leaf error at call %d", n)
					}

					return stubCreateLeaf(issuer, kp, name, dur, dns, ips, emails, ku, eku)
				},
				stubGetKeyPair, stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
			)

			err := gen.Generate(cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, t.TempDir())
			require.ErrorContains(t, err, tc.wantErrFrag)
		})
	}
}

// TestGenerate_PSIDCerts_ErrorBranches covers error-return statements in generatePSIDCerts.
//
// After generateSharedCAs completes (10 CA calls + 8 leaf calls), generatePSIDCerts runs:
//
//	createCA:  #11=cat4-sqlite1-root, #12=cat4-sqlite1-issuing, ... #22=cat6-first-root, ...
//	createLeaf: #9=cat3-sqlite-1, #10=cat3-sqlite-2, #11=cat3-postgres-1, #12=cat3-postgres-2,
//	            #13=cat5-first-leaf, ... #25=cat7-sqlite-1-mutual, ...
func TestGenerate_PSIDCerts_ErrorBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		failCAAt    int32
		failLeafAt  int32
		wantErrFrag string
	}{
		{"cat3 app server leaf error", 0, 9, "cat3 app server leaf"},
		{"cat4 client CA sqlite-1 error", 11, 0, "cat4 client CA"},
		{"cat5 client leaf error", 0, 13, "cat5 client leaf"},
		{"cat6 admin CA error", 23, 0, "cat6 admin CA"},
		{"cat7 mutual leaf error", 0, cryptoutilSharedMagic.IdentityDefaultMaxOpenConns, "cat7 admin mutual leaf"},
		{"cat9 grafana psid client error", 0, 29, "cat9 grafana psid client"},
		{"cat9 otel psid client error", 0, cryptoutilSharedMagic.FiberReadTimeoutSeconds, "cat9 otel psid client"},
		{"cat14 postgres client error", 0, 31, "cat14 postgres app client"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var caCount, leafCount atomic.Int32

			gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
				stubMkdirAll, stubWriteFile,
				func(issuer *cryptoutilSharedCryptoCertificate.Subject, issuerKey any, name string, kp *cryptoutilSharedCryptoKeygen.KeyPair, dur time.Duration, maxPath int) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					if n := caCount.Add(1); tc.failCAAt > 0 && n == tc.failCAAt {
						return nil, fmt.Errorf("injected CA error at call %d", n)
					}

					return stubCreateCA(issuer, issuerKey, name, kp, dur, maxPath)
				},
				func(issuer *cryptoutilSharedCryptoCertificate.Subject, kp *cryptoutilSharedCryptoKeygen.KeyPair, name string, dur time.Duration, dns []string, ips []net.IP, emails []string, ku x509.KeyUsage, eku []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error) {
					if n := leafCount.Add(1); tc.failLeafAt > 0 && n == tc.failLeafAt {
						return nil, fmt.Errorf("injected leaf error at call %d", n)
					}

					return stubCreateLeaf(issuer, kp, name, dur, dns, ips, emails, ku, eku)
				},
				stubGetKeyPair, stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
			)

			err := gen.Generate(cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, t.TempDir())
			require.ErrorContains(t, err, tc.wantErrFrag)
		})
	}
}

// --- additional helpers ---

// failNthEncode returns an encodePKCS12Fn that fails on the Nth call.
func failNthEncode(n int32) func(any, *x509.Certificate, []*x509.Certificate) ([]byte, error) {
	var count atomic.Int32

	return func(_ any, _ *x509.Certificate, _ []*x509.Certificate) ([]byte, error) {
		if count.Add(1) == n {
			return nil, fmt.Errorf("injected pkcs12 encode error at call %d", n)
		}

		return []byte{}, nil
	}
}

// failNthTrustEncode returns an encodeTrustPKCS12Fn that fails on the Nth call.
func failNthTrustEncode(n int32) func([]*x509.Certificate) ([]byte, error) {
	var count atomic.Int32

	return func(_ []*x509.Certificate) ([]byte, error) {
		if count.Add(1) == n {
			return nil, fmt.Errorf("injected trust pkcs12 encode error at call %d", n)
		}

		return []byte{}, nil
	}
}
