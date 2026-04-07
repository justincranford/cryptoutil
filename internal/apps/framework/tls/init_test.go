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

			code := cryptoutilAppsFrameworkTls.Init(tc.args, nil, &stdout, &stderr)
			require.Equal(t, 1, code)
			require.Contains(t, stderr.String(), "usage:")
			require.Empty(t, stdout.String())
		})
	}
}

// Sequential: mutates newTelemetryServiceFn and newGeneratorFn package-level state.
func TestInit_SeamInjection(t *testing.T) {
	t.Run("invalid tier ID", func(t *testing.T) {
		restore := setStubSeams(t, nil, nil, nil, nil, nil)
		defer restore()

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.Init([]string{"nonexistent-tier", t.TempDir()}, nil, &stdout, &stderr)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), "unknown tier ID")
		require.Empty(t, stdout.String())
	})

	t.Run("telemetry failure", func(t *testing.T) {
		restoreTelemetry := cryptoutilAppsFrameworkTls.ExportedSetNewTelemetryServiceFn(
			func(_ context.Context) (*cryptoutilSharedTelemetry.TelemetryService, error) {
				return nil, fmt.Errorf("injected telemetry error")
			},
		)
		defer restoreTelemetry()

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.Init([]string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, t.TempDir()}, nil, &stdout, &stderr)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), "injected telemetry error")
		require.Empty(t, stdout.String())
	})

	t.Run("generator creation failure", func(t *testing.T) {
		restoreTelemetry := cryptoutilAppsFrameworkTls.ExportedSetNewTelemetryServiceFn(
			func(_ context.Context) (*cryptoutilSharedTelemetry.TelemetryService, error) {
				return &cryptoutilSharedTelemetry.TelemetryService{}, nil
			},
		)
		defer restoreTelemetry()

		restoreGen := cryptoutilAppsFrameworkTls.ExportedSetNewGeneratorFn(
			func(_ context.Context, _ *cryptoutilSharedTelemetry.TelemetryService) (*cryptoutilAppsFrameworkTls.Generator, error) {
				return nil, fmt.Errorf("injected generator error")
			},
		)
		defer restoreGen()

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.Init([]string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, t.TempDir()}, nil, &stdout, &stderr)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), "injected generator error")
		require.Empty(t, stdout.String())
	})

	t.Run("generation failure mkdir", func(t *testing.T) {
		restore := setStubSeams(t,
			func(_ string, _ os.FileMode) error { return fmt.Errorf("injected mkdir error") },
			stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		)
		defer restore()

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.Init([]string{cryptoutilSharedMagic.OTLPServiceSMKMS, t.TempDir()}, nil, &stdout, &stderr)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), "pki-init:")
		require.Empty(t, stdout.String())
	})

	t.Run("non-empty target dir", func(t *testing.T) {
		restore := setStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)
		defer restore()

		outputDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(outputDir, "existing-file.txt"), []byte("data"), cryptoutilSharedMagic.PKIInitCertFileMode))

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.Init([]string{cryptoutilSharedMagic.OTLPServiceSMKMS, outputDir}, nil, &stdout, &stderr)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), "not empty")
		require.Empty(t, stdout.String())
	})

	t.Run("happy path suite", func(t *testing.T) {
		restore := setStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)
		defer restore()

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.Init([]string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, t.TempDir()}, nil, &stdout, &stderr)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Contains(t, stdout.String(), cryptoutilSharedMagic.DefaultOTLPServiceDefault)
		require.Empty(t, stderr.String())
	})

	t.Run("happy path product sm", func(t *testing.T) {
		restore := setStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)
		defer restore()

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.Init([]string{cryptoutilSharedMagic.SMProductName, t.TempDir()}, nil, &stdout, &stderr)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})

	t.Run("happy path product jose", func(t *testing.T) {
		restore := setStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)
		defer restore()

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.Init([]string{cryptoutilSharedMagic.JoseProductName, t.TempDir()}, nil, &stdout, &stderr)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})

	t.Run("happy path service sm-kms", func(t *testing.T) {
		restore := setStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)
		defer restore()

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.Init([]string{cryptoutilSharedMagic.OTLPServiceSMKMS, t.TempDir()}, nil, &stdout, &stderr)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})

	t.Run("happy path service jose-ja", func(t *testing.T) {
		restore := setStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)
		defer restore()

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.Init([]string{cryptoutilSharedMagic.OTLPServiceJoseJA, t.TempDir()}, nil, &stdout, &stderr)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})
}

// Sequential: mutates newTelemetryServiceFn and newGeneratorFn package-level state.
func TestInitForSuite_HappyPath(t *testing.T) {
	restore := setStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)
	defer restore()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.InitForSuite(cryptoutilSharedMagic.DefaultOTLPServiceDefault, []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, t.TempDir()}, &stdout, &stderr)
	require.Equal(t, 0, code, "stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
}

// Sequential: mutates newTelemetryServiceFn and newGeneratorFn package-level state.
func TestInitForProduct_HappyPath(t *testing.T) {
	restore := setStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)
	defer restore()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.InitForProduct(cryptoutilSharedMagic.JoseProductName, []string{cryptoutilSharedMagic.JoseProductName, t.TempDir()}, &stdout, &stderr)
	require.Equal(t, 0, code, "stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
}

// Sequential: mutates newTelemetryServiceFn and newGeneratorFn package-level state.
func TestInitForService_HappyPath(t *testing.T) {
	restore := setStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)
	defer restore()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.OTLPServiceSMKMS, []string{cryptoutilSharedMagic.OTLPServiceSMKMS, t.TempDir()}, &stdout, &stderr)
	require.Equal(t, 0, code, "stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
}

// setStubSeams configures both telemetry and generator seams with the given functions.
// Returns a restore function that restores both original seams.
func setStubSeams(
	t *testing.T,
	mkdirAllFn func(string, os.FileMode) error,
	writeFileFn func(string, []byte, os.FileMode) error,
	createCAFn func(*cryptoutilSharedCryptoCertificate.Subject, any, string, *cryptoutilSharedCryptoKeygen.KeyPair, time.Duration, int) (*cryptoutilSharedCryptoCertificate.Subject, error),
	createLeafFn func(*cryptoutilSharedCryptoCertificate.Subject, *cryptoutilSharedCryptoKeygen.KeyPair, string, time.Duration, []string, []net.IP, []string, x509.KeyUsage, []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error),
	getKeyPairFn func() *cryptoutilSharedCryptoKeygen.KeyPair,
) func() {
	t.Helper()

	restoreTelemetry := cryptoutilAppsFrameworkTls.ExportedSetNewTelemetryServiceFn(
		func(_ context.Context) (*cryptoutilSharedTelemetry.TelemetryService, error) {
			return &cryptoutilSharedTelemetry.TelemetryService{}, nil
		},
	)

	restoreGen := cryptoutilAppsFrameworkTls.ExportedSetNewGeneratorFn(
		func(_ context.Context, _ *cryptoutilSharedTelemetry.TelemetryService) (*cryptoutilAppsFrameworkTls.Generator, error) {
			return cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(mkdirAllFn, writeFileFn, createCAFn, createLeafFn, getKeyPairFn), nil
		},
	)

	return func() {
		restoreGen()
		restoreTelemetry()
	}
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
