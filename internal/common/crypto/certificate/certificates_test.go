package certificate

import (
	"context"
	"crypto"
	"crypto/elliptic"
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

func TestCreateInvalidRootCA_WithNegativeDuration(t *testing.T) {
	_, err := CertificateTemplateCA("Root CA", "Root CA", -1, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}

func TestCertificateChain(t *testing.T) {
	rootCert1SubjectName := "Test Root CA 1"
	rootCert1Duration := 10 * 365 * 24 * time.Hour // 10 years
	rootCert1MaxPathLen := 2

	intermediateCert1SubjectName := "Test Intermediate CA 1"
	intermediateCert1Duration := 5 * 365 * 24 * time.Hour // 5 years
	intermediateCert1MaxPathLen := rootCert1MaxPathLen - 1

	issuingCert1SubjectName := "Test Issuing CA 1"
	issuingCert1Duration := 2 * 365 * 24 * time.Hour // 2 years
	issuingCert1MaxPathLen := rootCert1MaxPathLen - 2

	var rootCert1KeyPair *cryptoutilKeyGen.KeyPair
	var rootCert1 *x509.Certificate
	var rootCert1DER []byte
	var rootCert1PEM []byte

	var intermediateCert1KeyPair *cryptoutilKeyGen.KeyPair
	var intermediateCert1 *x509.Certificate
	var intermediateCert1DER []byte
	var intermediateCert1PEM []byte

	var issuingCert1KeyPair *cryptoutilKeyGen.KeyPair
	var issuingCert1 *x509.Certificate
	var issuingCert1DER []byte
	var issuingCert1PEM []byte

	roots1Pool := x509.NewCertPool()
	intermediates1Pool := x509.NewCertPool()

	t.Run("Root1 CA", func(t *testing.T) {
		rootCert1Template, err := CertificateTemplateCA(rootCert1SubjectName, rootCert1SubjectName, rootCert1Duration, rootCert1MaxPathLen)
		verifyCertificateTemplate(t, err, rootCert1Template)

		rootCert1KeyPair = testKeyGenPool.Get()
		rootCert1, rootCert1DER, rootCert1PEM, err = SignCertificate(nil, rootCert1KeyPair.Private.(crypto.Signer), rootCert1Template, rootCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, rootCert1, rootCert1DER, rootCert1PEM, rootCert1SubjectName, rootCert1SubjectName, rootCert1MaxPathLen, rootCert1Duration)

		roots1Pool.AddCert(rootCert1)
	})

	t.Run("Intermediate1 CA", func(t *testing.T) {
		intermediateCert1Template, err := CertificateTemplateCA(rootCert1SubjectName, intermediateCert1SubjectName, intermediateCert1Duration, intermediateCert1MaxPathLen)
		verifyCertificateTemplate(t, err, intermediateCert1Template)

		intermediateCert1KeyPair = testKeyGenPool.Get()
		intermediateCert1, intermediateCert1DER, intermediateCert1PEM, err = SignCertificate(rootCert1, rootCert1KeyPair.Private.(crypto.Signer), intermediateCert1Template, intermediateCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, intermediateCert1, intermediateCert1DER, intermediateCert1PEM, rootCert1SubjectName, intermediateCert1SubjectName, intermediateCert1MaxPathLen, intermediateCert1Duration)

		verifyCertChain(t, intermediateCert1, roots1Pool, intermediates1Pool)
		intermediates1Pool.AddCert(intermediateCert1)
	})

	t.Run("Issuing1 CA", func(t *testing.T) {
		issuingCert1Template, err := CertificateTemplateCA(rootCert1SubjectName, issuingCert1SubjectName, issuingCert1Duration, issuingCert1MaxPathLen)
		verifyCertificateTemplate(t, err, issuingCert1Template)

		issuingCert1KeyPair = testKeyGenPool.Get()
		issuingCert1, issuingCert1DER, issuingCert1PEM, err = SignCertificate(rootCert1, rootCert1KeyPair.Private.(crypto.Signer), issuingCert1Template, issuingCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, issuingCert1, issuingCert1DER, issuingCert1PEM, rootCert1SubjectName, issuingCert1SubjectName, issuingCert1MaxPathLen, issuingCert1Duration)

		verifyCertChain(t, issuingCert1, roots1Pool, x509.NewCertPool())
		intermediates1Pool.AddCert(issuingCert1)
	})
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
