package certificate

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateRootCA_ECDSA_P521(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	require.NoError(t, err, "Failed to generate ECDSA P521 key pair")
	require.NotNil(t, privateKey, "Generated private key should not be nil")

	issuerName := "Test Root CA ECDSA P521"
	subjectName := "Test Root CA ECDSA P521"
	duration := 10 * 365 * 24 * time.Hour // 10 years
	maxPathLen := 2

	// Create root CA certificate template
	certTemplate, err := CertificateTemplateRootCA(issuerName, subjectName, duration, maxPathLen)
	require.NoError(t, err, "Failed to create certificate template")
	require.NotNil(t, certTemplate, "Certificate template should not be nil")

	// Sign root CA certificate
	certTemplate.SignatureAlgorithm = x509.ECDSAWithSHA512
	cert, certBytes, err := SignCertificate(nil, privateKey, certTemplate, &privateKey.PublicKey)
	require.NoError(t, err, "Failed to sign certificate")
	require.NotNil(t, cert, "Signed certificate should not be nil")
	require.NotEmpty(t, certBytes, "Certificate bytes should not be empty")

	// Verify root CA certificate
	now := time.Now().UTC()
	require.Equal(t, issuerName, cert.Issuer.CommonName, "Issuer name mismatch")
	require.Equal(t, subjectName, cert.Subject.CommonName, "Subject name mismatch")
	require.True(t, cert.IsCA, "Certificate should be a CA certificate")
	require.True(t, cert.BasicConstraintsValid, "Basic constraints should be valid")
	require.Equal(t, maxPathLen, cert.MaxPathLen, "MaxPathLen should be 2 for root CA")
	require.False(t, cert.MaxPathLenZero, "MaxPathLenZero should be false")
	require.Equal(t, x509.KeyUsageCertSign|x509.KeyUsageCRLSign, cert.KeyUsage, "Key usage mismatch")
	require.ElementsMatch(t, []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping, x509.ExtKeyUsageOCSPSigning}, cert.ExtKeyUsage, "Extended key usage mismatch")
	require.True(t, cert.NotBefore.Before(now), "NotBefore should be in the past")
	require.True(t, cert.NotAfter.After(now), "NotAfter should be in the future")
	require.True(t, cert.NotAfter.Sub(cert.NotBefore) >= duration, "Certificate validity period should be >= duration")

	// Verify CA certificate PEM encoding
	pemBlock := &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}
	pemData := pem.EncodeToMemory(pemBlock)
	require.NotEmpty(t, pemData, "PEM encoded certificate should not be empty")
	t.Logf("PEM encoded certificate:\n%s", pemData)
}

func TestCreateAndVerifyCertificateChain_ECDSA_P521(t *testing.T) {
	// Generate ECDSA P521 key pair for the root CA
	rootPrivateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	require.NoError(t, err, "Failed to generate ECDSA P521 key pair for root CA")

	// Generate ECDSA P521 key pair for the intermediate CA
	intermediatePrivateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	require.NoError(t, err, "Failed to generate ECDSA P521 key pair for intermediate CA")

	// Create root CA certificate
	rootIssuerName := "Test Root CA ECDSA P521"
	rootSubjectName := "Test Root CA ECDSA P521"
	rootDuration := 10 * 365 * 24 * time.Hour // 10 years
	rootMaxPathLen := 2

	rootTemplate, err := CertificateTemplateRootCA(rootIssuerName, rootSubjectName, rootDuration, rootMaxPathLen)
	require.NoError(t, err, "Failed to create root CA template")

	rootCert, _, err := SignCertificate(nil, rootPrivateKey, rootTemplate, &rootPrivateKey.PublicKey)
	require.NoError(t, err, "Failed to sign root CA certificate")

	// Create intermediate CA certificate
	intermediateIssuerName := rootSubjectName
	intermediateSubjectName := "Test Intermediate CA ECDSA P521"
	intermediateDuration := 5 * 365 * 24 * time.Hour // 5 years
	intermediateMaxPathLen := rootMaxPathLen - 1

	intermediateTemplate, err := CertificateTemplateIntermediateCA(intermediateIssuerName, intermediateSubjectName, intermediateDuration, intermediateMaxPathLen)
	require.NoError(t, err, "Failed to create intermediate CA template")

	intermediateCert, _, err := SignCertificate(rootCert, rootPrivateKey, intermediateTemplate, &intermediatePrivateKey.PublicKey)
	require.NoError(t, err, "Failed to sign intermediate CA certificate")

	// Create certificate pool and add root certificate
	roots := x509.NewCertPool()
	roots.AddCert(rootCert)

	// Create certificate pool and add intermediate certificate
	intermediates := x509.NewCertPool()
	intermediates.AddCert(intermediateCert)

	// Verify intermediate certificate using root certificate
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: x509.NewCertPool(),
		CurrentTime:   time.Now().UTC(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	chains, err := intermediateCert.Verify(opts)
	require.NoError(t, err, "Failed to verify intermediate certificate using root certificate")
	require.NotEmpty(t, chains, "Certificate chains should not be empty")
}

func TestCreateInvalidRootCA_ECDSA_P521_WithNegativeDuration(t *testing.T) {
	// We don't need the private key for this test as we're just testing template creation
	_, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	require.NoError(t, err, "Failed to generate ECDSA P521 key pair")

	// Try to create CA certificate with negative duration
	issuerName := "Test Root CA ECDSA P521"
	subjectName := "Test Root CA ECDSA P521"
	negativeDuration := -1 * time.Hour

	// This should cause an error in randomizedNotBeforeNotAfterCA
	_, err = CertificateTemplateRootCA(issuerName, subjectName, negativeDuration, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}
