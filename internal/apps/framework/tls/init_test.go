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

func TestInit_MissingFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "no flags", args: []string{}},
		{name: "domain only", args: []string{"--domain=" + cryptoutilSharedMagic.DefaultOTLPServiceDefault}},
		{name: "output-dir only", args: []string{"--output-dir=/certs"}},
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

func TestInit_UnknownFlag(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--unknown-flag=value"}, nil, &stdout, &stderr, nil, nil)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "unknown flag")
	require.Empty(t, stdout.String())
}

func TestInit_SeamInjection(t *testing.T) {
	t.Parallel()

	t.Run("invalid tier ID", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, nil, nil, nil, nil, nil)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=nonexistent-tier", "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
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

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.DefaultOTLPServiceDefault, "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, nil)
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

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.DefaultOTLPServiceDefault, "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
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

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.OTLPServiceSMKMS, "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), "pki-init:")
		require.Empty(t, stdout.String())
	})

	t.Run("non-empty target dir", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		outputDir := t.TempDir()
		// The basePath is filepath.Join(outputDir, tierID), so the file must be inside the tierID subdir.
		subDir := filepath.Join(outputDir, cryptoutilSharedMagic.OTLPServiceSMKMS)
		require.NoError(t, os.MkdirAll(subDir, cryptoutilSharedMagic.CICDTempDirPermissions))
		require.NoError(t, os.WriteFile(filepath.Join(subDir, "existing-file.txt"), []byte("data"), cryptoutilSharedMagic.PKIInitCertFileMode))

		var stdout, stderr bytes.Buffer

		// Non-empty target dir is cleaned and regenerated — pki-init is idempotent.
		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.OTLPServiceSMKMS, "--output-dir=" + outputDir}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})

	t.Run("happy path suite", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.DefaultOTLPServiceDefault, "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Contains(t, stdout.String(), cryptoutilSharedMagic.DefaultOTLPServiceDefault)
		require.Empty(t, stderr.String())
	})

	t.Run("happy path product sm", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.SMProductName, "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})

	t.Run("happy path product jose", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.JoseProductName, "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})

	t.Run("happy path service sm-kms", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.OTLPServiceSMKMS, "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})

	t.Run("happy path service jose-ja", func(t *testing.T) {
		t.Parallel()

		telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

		var stdout, stderr bytes.Buffer

		code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.OTLPServiceJoseJA, "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
		require.Equal(t, 0, code, "stderr=%s", stderr.String())
		require.Contains(t, stdout.String(), "certificates written")
		require.Empty(t, stderr.String())
	})
}

func TestInitForSuite_HappyPath(t *testing.T) {
	t.Parallel()

	telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.DefaultOTLPServiceDefault, "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
	require.Equal(t, 0, code, "stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
}

func TestInitForProduct_HappyPath(t *testing.T) {
	t.Parallel()

	telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.JoseProductName, "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
	require.Equal(t, 0, code, "stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
}

func TestInitForService_HappyPath(t *testing.T) {
	t.Parallel()

	telemetryFn, generatorFn := buildStubSeams(t, stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair)

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.ExportedInitRun([]string{"--domain=" + cryptoutilSharedMagic.OTLPServiceSMKMS, "--output-dir=" + t.TempDir()}, nil, &stdout, &stderr, telemetryFn, generatorFn)
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
		return cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
			mkdirAllFn, writeFileFn, createCAFn, createLeafFn, getKeyPairFn,
			stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
		), nil
	}

	return telemetryFn, generatorFn
}

// stubEncodePKCS12 is a no-op PKCS#12 keystore encoder for testing.
func stubEncodePKCS12(_ any, _ *x509.Certificate, _ []*x509.Certificate) ([]byte, error) {
	return []byte{}, nil
}

// stubEncodeTrustPKCS12 is a no-op PKCS#12 truststore encoder for testing.
func stubEncodeTrustPKCS12(_ []*x509.Certificate) ([]byte, error) {
	return []byte{}, nil
}

// stubGetRealmsForPSID returns the default realms for testing.
func stubGetRealmsForPSID(_ string) ([]string, error) {
	return []string{"file", "db"}, nil
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
