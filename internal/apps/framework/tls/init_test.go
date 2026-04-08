// Copyright (c) 2025 Justin Cranford
//

package tls_test

import (
	"bytes"
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps/framework/tls"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

func TestInit_WrongArgCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "zero args", args: []string{}},
		{name: "one arg", args: []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault}},
		{name: "three args", args: []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, "/certs", "extra"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			code := cryptoutilAppsFrameworkTls.ExportedInitRun(tc.args, nil, &stdout, &stderr, nil, nil)
			require.Equal(t, 1, code)
			require.Contains(t, stderr.String(), "usage:")
			require.Empty(t, stdout.String())
		})
	}
}

func TestInit_SeamInjection(t *testing.T) {
	t.Parallel()

	t.Run("invalid tier ID", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, nil, nil, nil, nil, nil)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"nonexistent-tier", t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), "unknown tier ID")
		require.Empty(t, stdout.String())
	})

	t.Run("telemetry failure", func(t *testing.T) {
		t.Parallel()

		telemetryFn := func(_ context.Context) (*cryptoutilSharedTelemetry.TelemetryService, error) {
			return nil, fmt.Errorf("injected telemetry error")
		}

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, t.TempDir()}, nil, &stdout, &stderr, telemetryFn, nil)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), "injected telemetry error")
		require.Empty(t, stdout.String())
	})

	t.Run("generator creation failure", func(t *testing.T) {
		t.Parallel()

		telemetryFn := func(_ context.Context) (*cryptoutilSharedTelemetry.TelemetryService, error) {
			return &cryptoutilSharedTelemetry.TelemetryService{}, nil
		}

		generatorFn := func(_ context.Context, _ *cryptoutilSharedTelemetry.TelemetryService) (*cryptoutilAppsFrameworkTls.Generator, error) {
			return nil, fmt.Errorf("injected generator error")
		}

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), "injected generator error")
		require.Empty(t, stdout.String())
	})

	t.Run("generation failure mkdir", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t,
			func(_ string, _ os.FileMode) error { return fmt.Errorf("injected mkdir error") },
			stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.OTLPServiceSMKMS, t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), "pki-init:")
		require.Empty(t, stdout.String())
	})

	t.Run("non-empty target dir", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		outputDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(outputDir, "existing-file.txt"), []byte("data"), cryptoutilSharedMagic.PKIInitCertFileMode))

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.OTLPServiceSMKMS, outputDir}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), "not empty")
		require.Empty(t, stdout.String())
	})

	t.Run("happy path suite", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Contains(t, stdout.String(), cryptoutilSharedMagic.DefaultOTLPServiceDefault)
		require.Empty(t, stderr.String())
	})

	t.Run("happy path product sm", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.SMProductName, t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})

	t.Run("happy path product jose", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.JoseProductName, t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})

	t.Run("happy path service sm-kms", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.OTLPServiceSMKMS, t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})

	t.Run("happy path service jose-ja", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.OTLPServiceJoseJA, t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})
}

func TestInitForSuite_HappyPath(t *testing.T) {
	t.Parallel()

	telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
	require.Equal(t, 0, code, "stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
}

func TestInitForProduct_HappyPath(t *testing.T) {
	t.Parallel()

	telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.JoseProductName, t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
	require.Equal(t, 0, code, "stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
}

func TestInitForService_HappyPath(t *testing.T) {
	t.Parallel()

	telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{cryptoutilSharedMagic.OTLPServiceSMKMS, t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
	require.Equal(t, 0, code, "stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
}

// buildStubSeams builds telemetry and generator seam functions with the given injectable functions.
func buildStubSeams(
	t *testing.T,
	mkdirAllFn func(string, os.FileMode) error,
	writeFileFn func(string, []byte, os.FileMode) error,
	createCAFn func(*cryptoutilSharedCryptoCertificate.Subject, any, string, *cryptoutilSharedCryptoKeygen.KeyPair, time.Duration, int) (*cryptoutilSharedCryptoCertificate.Subject, error),
	createLeafFn func(*cryptoutilSharedCryptoCertificate.Subject, *cryptoutilSharedCryptoKeygen.KeyPair, string, time.Duration, []string, []net.IP, []string, x509.KeyUsage, []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error),
	getKeyPairFn func() *cryptoutilSharedCryptoKeygen.KeyPair,
) (cryptoutilAppsFrameworkTls.TelemetryFnType, cryptoutilAppsFrameworkTls.GeneratorFnType) {
	t.Helper()

	telemetryFn := func(_ context.Context) (*cryptoutilSharedTelemetry.TelemetryService, error) {
		return &cryptoutilSharedTelemetry.TelemetryService{}, nil
	}

	generatorFn := func(_ context.Context, _ *cryptoutilSharedTelemetry.TelemetryService) (*cryptoutilAppsFrameworkTls.Generator, error) {
		return cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(mkdirAllFn, writeFileFn, createCAFn, createLeafFn, getKeyPairFn), nil
	}

	return telemetryFn, generatorFn
}

// stubMkdirAll is a no-op mkdir for testing.
func stubMkdirAll(_ string, _ os.FileMode) error { return nil }

// stubWriteFile is a no-op file write for testing.
func stubWriteFile(_ string, _ []byte, _ os.FileMode) error { return nil }

// stubGetKeyPair returns a test ECDSA P-256 key pair (fast, not production P-384).
func stubGetKeyPair() *cryptoutilSharedCryptoKeygen.KeyPair {
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		panic(fmt.Sprintf("stub key gen failed: %v", err))
	}

	return &cryptoutilSharedCryptoKeygen.KeyPair{Private: key, Public: &key.PublicKey}
}

// stubCreateCA returns a minimal self-signed CA Subject for testing.
func stubCreateCA(
	_ *cryptoutilSharedCryptoCertificate.Subject,
	_ any,
	name string,
	kp *cryptoutilSharedCryptoKeygen.KeyPair,
	_ time.Duration,
	_ int,
) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, template, kp.Public, kp.Private)
	if err != nil {
		return nil, fmt.Errorf("stub CA cert: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("stub CA parse: %w", err)
	}

	return &cryptoutilSharedCryptoCertificate.Subject{
		SubjectName: name,
		IsCA:        true,
		KeyMaterial: cryptoutilSharedCryptoCertificate.KeyMaterial{
			CertificateChain: []*x509.Certificate{cert},
			PublicKey:        kp.Public,
			PrivateKey:       kp.Private,
		},
	}, nil
}

// stubCreateLeaf returns a minimal leaf Subject for testing.
func stubCreateLeaf(
	issuer *cryptoutilSharedCryptoCertificate.Subject,
	kp *cryptoutilSharedCryptoKeygen.KeyPair,
	name string,
	_ time.Duration,
	_ []string,
	_ []net.IP,
	_ []string,
	_ x509.KeyUsage,
	_ []x509.ExtKeyUsage,
) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	issuerCert := issuer.KeyMaterial.CertificateChain[0]

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, issuerCert, kp.Public, issuer.KeyMaterial.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("stub leaf cert: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("stub leaf parse: %w", err)
	}

	return &cryptoutilSharedCryptoCertificate.Subject{
		SubjectName: name,
		KeyMaterial: cryptoutilSharedCryptoCertificate.KeyMaterial{
			CertificateChain: []*x509.Certificate{cert, issuerCert},
			PublicKey:        kp.Public,
			PrivateKey:       kp.Private,
		},
	}, nil
}
