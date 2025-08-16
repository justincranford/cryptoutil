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
	rootCertSubjectName := "Test Root CA"
	rootCertDuration := 10 * 365 * 24 * time.Hour // 10 years
	rootCertMaxPathLen := 2

	intermediateCertSubjectName := "Test Intermediate CA"
	intermediateCertDuration := 5 * 365 * 24 * time.Hour // 5 years
	intermediateCertMaxPathLen := rootCertMaxPathLen - 1

	issuingCertSubjectName := "Test Issuing CA"
	issuingCertDuration := 2 * 365 * 24 * time.Hour // 2 years
	issuingCertMaxPathLen := rootCertMaxPathLen - 2

	var rootCertKeyPair *cryptoutilKeyGen.KeyPair
	var rootCert *x509.Certificate
	var rootCertDER []byte
	var rootCertPEM []byte

	var intermediateCertKeyPair *cryptoutilKeyGen.KeyPair
	var intermediateCert *x509.Certificate
	var intermediateCertDER []byte
	var intermediateCertPEM []byte

	var issuingCertKeyPair *cryptoutilKeyGen.KeyPair
	var issuingCert *x509.Certificate
	var issuingCertDER []byte
	var issuingCertPEM []byte

	rootsPool := x509.NewCertPool()
	intermediatesPool := x509.NewCertPool()

	t.Run("RootCA", func(t *testing.T) {
		rootCertTemplate, err := CertificateTemplateCA(x509.ECDSAWithSHA256, rootCertSubjectName, rootCertSubjectName, rootCertDuration, rootCertMaxPathLen)
		verifyCertificateTemplate(t, err, rootCertTemplate)

		rootCertKeyPair = testKeyGenPool.Get()
		rootCert, rootCertDER, rootCertPEM, err = SignCertificate(nil, rootCertKeyPair.Private.(crypto.Signer), rootCertTemplate, rootCertKeyPair.Public)
		verifyCACertificate(t, err, rootCert, rootCertDER, rootCertPEM, rootCertSubjectName, rootCertSubjectName, rootCertMaxPathLen, rootCertDuration)

		rootsPool.AddCert(rootCert)
	})

	t.Run("IntermediateCA", func(t *testing.T) {
		intermediateCertTemplate, err := CertificateTemplateCA(x509.ECDSAWithSHA256, rootCertSubjectName, intermediateCertSubjectName, intermediateCertDuration, intermediateCertMaxPathLen)
		verifyCertificateTemplate(t, err, intermediateCertTemplate)

		intermediateCertKeyPair = testKeyGenPool.Get()
		intermediateCert, intermediateCertDER, intermediateCertPEM, err = SignCertificate(rootCert, rootCertKeyPair.Private.(crypto.Signer), intermediateCertTemplate, intermediateCertKeyPair.Public)
		verifyCACertificate(t, err, intermediateCert, intermediateCertDER, intermediateCertPEM, rootCertSubjectName, intermediateCertSubjectName, intermediateCertMaxPathLen, intermediateCertDuration)

		verifyCertChain(t, intermediateCert, rootsPool, intermediatesPool)
		intermediatesPool.AddCert(intermediateCert)
	})

	t.Run("IssuingCA", func(t *testing.T) {
		issuingCertTemplate, err := CertificateTemplateCA(x509.ECDSAWithSHA256, rootCertSubjectName, issuingCertSubjectName, issuingCertDuration, issuingCertMaxPathLen)
		verifyCertificateTemplate(t, err, issuingCertTemplate)

		issuingCertKeyPair = testKeyGenPool.Get()
		issuingCert, issuingCertDER, issuingCertPEM, err = SignCertificate(rootCert, rootCertKeyPair.Private.(crypto.Signer), issuingCertTemplate, issuingCertKeyPair.Public)
		verifyCACertificate(t, err, issuingCert, issuingCertDER, issuingCertPEM, rootCertSubjectName, issuingCertSubjectName, issuingCertMaxPathLen, issuingCertDuration)

		verifyCertChain(t, issuingCert, rootsPool, x509.NewCertPool())
		intermediatesPool.AddCert(issuingCert)
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
	_, err = CertificateTemplateCA(x509.ECDSAWithSHA256, issuerName, subjectName, negativeDuration, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}

func verifyCertificateTemplate(t *testing.T, err error, rootTemplate *x509.Certificate) {
	require.NoError(t, err, "Failed to create certificate template")
	require.NotNil(t, rootTemplate, "Certificate template should not be nil")
}

func verifyCACertificate(t *testing.T, err error, cert *x509.Certificate, certDER []byte, certPEM []byte, expectedIssuerName string, expectedSubjectName string, expectedMaxPathLen int, expectedDuration time.Duration) {
	require.NoError(t, err, "Failed to sign certificate")
	require.NotNil(t, cert, "Signed certificate should not be nil")
	require.NotEmpty(t, certDER, "Certificate bytes should not be empty")
	require.NotEmpty(t, certPEM, "Certificate PEM should not be empty")
	now := time.Now().UTC()
	require.Equal(t, expectedIssuerName, cert.Issuer.CommonName, "Issuer name mismatch")
	require.Equal(t, expectedSubjectName, cert.Subject.CommonName, "Subject name mismatch")
	require.True(t, cert.IsCA, "Certificate should be a CA certificate")
	require.True(t, cert.BasicConstraintsValid, "Basic constraints should be valid")
	require.Equal(t, expectedMaxPathLen, cert.MaxPathLen, "MaxPathLen mismatch")
	require.Equal(t, expectedMaxPathLen == 0, cert.MaxPathLenZero, "MaxPathLenZero mismatch")
	require.Equal(t, x509.KeyUsageCertSign|x509.KeyUsageCRLSign, cert.KeyUsage, "Key usage mismatch")
	require.ElementsMatch(t, []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping, x509.ExtKeyUsageOCSPSigning}, cert.ExtKeyUsage, "Extended key usage mismatch")
	require.True(t, cert.NotBefore.Before(now), "NotBefore should be in the past")
	require.True(t, cert.NotAfter.After(now), "NotAfter should be in the future")
	require.True(t, cert.NotAfter.Sub(cert.NotBefore) >= expectedDuration, "Certificate validity period should be >= duration")
}

func verifyCertChain(t *testing.T, certificate *x509.Certificate, roots *x509.CertPool, intermediates *x509.CertPool) {
	x509VerifyOptions := x509.VerifyOptions{
		CurrentTime:   time.Now().UTC(),
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	chains, err := certificate.Verify(x509VerifyOptions)
	require.NoError(t, err, "Failed to verify intermediate certificate using root certificate")
	require.NotEmpty(t, chains, "Certificate chains should not be empty")
}
