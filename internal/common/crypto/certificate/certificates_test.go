package certificate

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"log"
	"os"
	"testing"
	"time"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	"github.com/stretchr/testify/require"
)

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("certificates_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testKeyGenPool       *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair]
)

func TestMain(m *testing.M) {
	var rc int
	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		var err error
		testKeyGenPool, err = cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(testCtx, testTelemetryService, "certificates_test", 1, 4, 4, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P256())))
		if err != nil {
			log.Fatalf("failed to create key pool: %v", err)
		}
		defer testKeyGenPool.Cancel()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestCertificateChain(t *testing.T) {
	// Root CA parameters
	rootSubjectName := "Test Root CA"
	rootDuration := 10 * 365 * 24 * time.Hour // 10 years
	rootMaxPathLen := 2

	// Intermediate CA parameters
	intermediateSubjectName := "Test Intermediate CA"
	intermediateDuration := 5 * 365 * 24 * time.Hour // 5 years
	intermediateMaxPathLen := rootMaxPathLen - 1

	var rootKeyPair *cryptoutilKeyGen.KeyPair
	var rootCert *x509.Certificate
	var rootCertPool *x509.CertPool
	var rootCertDer []byte
	var rootCertPem []byte

	var intermediateKeyPair *cryptoutilKeyGen.KeyPair
	var intermediateCert *x509.Certificate
	var intermediateCertPool *x509.CertPool
	var intermediateCertDer []byte
	var intermediateCertPem []byte

	t.Run("CreateAndVerifyRootCA", func(t *testing.T) {
		// Create root CA certificate template
		rootTemplate, err := CertificateTemplateRootCA(x509.ECDSAWithSHA256, rootSubjectName, rootSubjectName, rootDuration, rootMaxPathLen)
		require.NoError(t, err, "Failed to create certificate template")
		require.NotNil(t, rootTemplate, "Certificate template should not be nil")

		// Sign root CA certificate
		rootKeyPair = testKeyGenPool.Get()
		rootCert, rootCertDer, rootCertPem, err = SignCertificate(nil, rootKeyPair.Private.(crypto.Signer), rootTemplate, rootKeyPair.Public)
		require.NoError(t, err, "Failed to sign certificate")
		require.NotNil(t, rootCert, "Signed certificate should not be nil")
		require.NotEmpty(t, rootCertDer, "Certificate bytes should not be empty")
		require.NotEmpty(t, rootCertPem, "Certificate PEM should not be empty")

		// Create certificate pool and add root certificate
		rootCertPool = x509.NewCertPool()
		rootCertPool.AddCert(rootCert)

		// Verify root CA certificate
		now := time.Now().UTC()
		require.Equal(t, rootSubjectName, rootCert.Issuer.CommonName, "Issuer name mismatch")
		require.Equal(t, rootSubjectName, rootCert.Subject.CommonName, "Subject name mismatch")
		require.True(t, rootCert.IsCA, "Certificate should be a CA certificate")
		require.True(t, rootCert.BasicConstraintsValid, "Basic constraints should be valid")
		require.Equal(t, rootMaxPathLen, rootCert.MaxPathLen, "MaxPathLen should be 2 for root CA")
		require.False(t, rootCert.MaxPathLenZero, "MaxPathLenZero should be false")
		require.Equal(t, x509.KeyUsageCertSign|x509.KeyUsageCRLSign, rootCert.KeyUsage, "Key usage mismatch")
		require.ElementsMatch(t, []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping, x509.ExtKeyUsageOCSPSigning}, rootCert.ExtKeyUsage, "Extended key usage mismatch")
		require.True(t, rootCert.NotBefore.Before(now), "NotBefore should be in the past")
		require.True(t, rootCert.NotAfter.After(now), "NotAfter should be in the future")
		require.True(t, rootCert.NotAfter.Sub(rootCert.NotBefore) >= rootDuration, "Certificate validity period should be >= duration")
	})

	t.Run("CreateAndVerifyIntermediateCA", func(t *testing.T) {
		intermediateTemplate, err := CertificateTemplateIntermediateCA(x509.ECDSAWithSHA256, rootSubjectName, intermediateSubjectName, intermediateDuration, intermediateMaxPathLen)
		require.NoError(t, err, "Failed to create intermediate CA template")
		require.NotNil(t, intermediateTemplate, "Intermediate certificate template should not be nil")

		intermediateKeyPair = testKeyGenPool.Get()
		intermediateCert, intermediateCertDer, intermediateCertPem, err = SignCertificate(rootCert, rootKeyPair.Private.(crypto.Signer), intermediateTemplate, intermediateKeyPair.Public)
		require.NoError(t, err, "Failed to sign intermediate CA certificate")
		require.NotNil(t, intermediateCert, "Signed intermediate certificate should not be nil")
		require.NotEmpty(t, intermediateCertDer, "Intermediate certificate bytes should not be empty")
		require.NotEmpty(t, intermediateCertPem, "Intermediate certificate PEM should not be empty")

		// Create certificate pool and add intermediate certificate
		intermediateCertPool = x509.NewCertPool()
		intermediateCertPool.AddCert(intermediateCert)

		// Verify intermediate certificate using root certificate
		opts := x509.VerifyOptions{
			Roots:         rootCertPool,
			Intermediates: x509.NewCertPool(),
			CurrentTime:   time.Now().UTC(),
			KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		}
		chains, err := intermediateCert.Verify(opts)
		require.NoError(t, err, "Failed to verify intermediate certificate using root certificate")
		require.NotEmpty(t, chains, "Certificate chains should not be empty")

		// Verify intermediate CA certificate properties
		now := time.Now().UTC()
		require.Equal(t, rootSubjectName, intermediateCert.Issuer.CommonName, "Intermediate issuer name mismatch")
		require.Equal(t, intermediateSubjectName, intermediateCert.Subject.CommonName, "Intermediate subject name mismatch")
		require.True(t, intermediateCert.IsCA, "Intermediate certificate should be a CA certificate")
		require.True(t, intermediateCert.BasicConstraintsValid, "Intermediate basic constraints should be valid")
		require.Equal(t, intermediateMaxPathLen, intermediateCert.MaxPathLen, "Intermediate MaxPathLen should be 1")
		require.False(t, intermediateCert.MaxPathLenZero, "Intermediate MaxPathLenZero should be false")
		require.True(t, intermediateCert.NotBefore.Before(now), "Intermediate NotBefore should be in the past")
		require.True(t, intermediateCert.NotAfter.After(now), "Intermediate NotAfter should be in the future")
		require.True(t, intermediateCert.NotAfter.Sub(intermediateCert.NotBefore) >= intermediateDuration, "Intermediate certificate validity period should be >= duration")
	})
}

func TestCreateInvalidRootCA_WithNegativeDuration(t *testing.T) {
	// We don't need the private key for this test as we're just testing template creation
	_, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "Failed to generate ECDSA key pair")

	// Try to create CA certificate with negative duration
	issuerName := "Test Root CA"
	subjectName := "Test Root CA"
	negativeDuration := -1 * time.Hour

	// This should cause an error in randomizedNotBeforeNotAfterCA
	_, err = CertificateTemplateRootCA(x509.ECDSAWithSHA256, issuerName, subjectName, negativeDuration, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}
