package certificate

import (
	"crypto"
	"crypto/x509"
	"net"
	"net/url"
	"testing"
	"time"

	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"

	"github.com/stretchr/testify/require"
)

const (
	testRrootCertDuration          = 10 * 365 * 24 * time.Hour // 10 years
	testIntermediateCertDuration   = 5 * 365 * 24 * time.Hour  // 5 years
	testIssuingCertDuration        = 2 * 365 * 24 * time.Hour  // 2 years
	testTlsServerCertDuration      = 397 * 24 * time.Hour      // 398 days max
	testTlsClientCertDuration      = 397 * 24 * time.Hour      // 398 days max
	testRootCertMaxPathLen         = 2
	testIntermediateCertMaxPathLen = testRootCertMaxPathLen - 1
	testIssuingCertMaxPathLen      = testRootCertMaxPathLen - 2
)

func TestCreateInvalidRootCA_WithNegativeDuration(t *testing.T) {
	_, err := CertificateTemplateCA("Root CA", "Root CA", -1, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}

func TestCertificateChain(t *testing.T) {
	rootCert1SubjectName := "Test Root CA 1"
	intermediateCert1SubjectName := "Test Intermediate CA 1"
	issuingCert1SubjectName := "Test Issuing CA 1"
	tlsServerCert1SubjectName := "TLS Server 1"
	tlsClientCert1SubjectName := "TLS Client 1"

	tlsServerCert1DnsNames := []string{"localhost", "example.com"}
	tlsServerCert1IpAddresses := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
	tlsServerCert1EmailAddresses := []string{}
	tlsServerCert1URIs := []*url.URL{}

	tlsClientCert1DnsNames := []string{}
	tlsClientCert1IpAddresses := []net.IP{}
	tlsClientCert1EmailAddresses := []string{"client@example.com"}
	tlsClientCert1URIs := []*url.URL{}

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

	var tlsServerCert1KeyPair *cryptoutilKeyGen.KeyPair
	var tlsServerCert1 *x509.Certificate
	var tlsServerCert1DER []byte
	var tlsServerCert1PEM []byte

	var tlsClientCert1KeyPair *cryptoutilKeyGen.KeyPair
	var tlsClientCert1 *x509.Certificate
	var tlsClientCert1DER []byte
	var tlsClientCert1PEM []byte

	verify1RootsPool := x509.NewCertPool()
	verify1IntermediatesPool := x509.NewCertPool()

	t.Run("Root CA 1", func(t *testing.T) {
		rootCert1Template, err := CertificateTemplateCA(rootCert1SubjectName, rootCert1SubjectName, testRrootCertDuration, testRootCertMaxPathLen)
		verifyCertificateTemplate(t, err, rootCert1Template)

		rootCert1KeyPair = testKeyGenPool.Get()
		rootCert1, rootCert1DER, rootCert1PEM, err = SignCertificate(nil, rootCert1KeyPair.Private.(crypto.Signer), rootCert1Template, rootCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, rootCert1, rootCert1DER, rootCert1PEM, rootCert1SubjectName, rootCert1SubjectName, testRrootCertDuration, testRootCertMaxPathLen)

		verify1RootsPool.AddCert(rootCert1) // subsequent verify cert chain calls need the root CA
	})

	t.Run("Intermediate CA 1", func(t *testing.T) {
		intermediateCert1Template, err := CertificateTemplateCA(rootCert1SubjectName, intermediateCert1SubjectName, testIntermediateCertDuration, testIntermediateCertMaxPathLen)
		verifyCertificateTemplate(t, err, intermediateCert1Template)

		intermediateCert1KeyPair = testKeyGenPool.Get()
		intermediateCert1, intermediateCert1DER, intermediateCert1PEM, err = SignCertificate(rootCert1, rootCert1KeyPair.Private.(crypto.Signer), intermediateCert1Template, intermediateCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, intermediateCert1, intermediateCert1DER, intermediateCert1PEM, rootCert1SubjectName, intermediateCert1SubjectName, testIntermediateCertDuration, testIntermediateCertMaxPathLen)

		verifyCertChain(t, intermediateCert1, verify1RootsPool, verify1IntermediatesPool)
		verify1IntermediatesPool.AddCert(intermediateCert1) // subsequent verify cert chain calls need the intermediate CA
	})

	t.Run("Issuing CA 1", func(t *testing.T) {
		issuingCert1Template, err := CertificateTemplateCA(intermediateCert1SubjectName, issuingCert1SubjectName, testIssuingCertDuration, testIssuingCertMaxPathLen)
		verifyCertificateTemplate(t, err, issuingCert1Template)

		issuingCert1KeyPair = testKeyGenPool.Get()
		issuingCert1, issuingCert1DER, issuingCert1PEM, err = SignCertificate(intermediateCert1, intermediateCert1KeyPair.Private.(crypto.Signer), issuingCert1Template, issuingCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, issuingCert1, issuingCert1DER, issuingCert1PEM, intermediateCert1SubjectName, issuingCert1SubjectName, testIssuingCertDuration, testIssuingCertMaxPathLen)

		verifyCertChain(t, issuingCert1, verify1RootsPool, verify1IntermediatesPool)
		verify1IntermediatesPool.AddCert(issuingCert1) // subsequent verify cert chain calls need the issuing CA
	})

	t.Run("TLS Server 1", func(t *testing.T) {
		tlsServerCert1Template, err := CertificateTemplateTLSServer(issuingCert1SubjectName, tlsServerCert1SubjectName, testTlsServerCertDuration, tlsServerCert1DnsNames, tlsServerCert1IpAddresses, tlsServerCert1EmailAddresses, tlsServerCert1URIs)
		verifyCertificateTemplate(t, err, tlsServerCert1Template)

		tlsServerCert1KeyPair = testKeyGenPool.Get()
		tlsServerCert1, tlsServerCert1DER, tlsServerCert1PEM, err = SignCertificate(issuingCert1, issuingCert1KeyPair.Private.(crypto.Signer), tlsServerCert1Template, tlsServerCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyEndEntityCertificate(t, err, tlsServerCert1, tlsServerCert1DER, tlsServerCert1PEM, issuingCert1SubjectName, tlsServerCert1SubjectName, testTlsServerCertDuration, tlsServerCert1DnsNames, tlsServerCert1IpAddresses, tlsServerCert1EmailAddresses, tlsServerCert1URIs)

		verifyCertChain(t, tlsServerCert1, verify1RootsPool, verify1IntermediatesPool)
	})

	t.Run("TLS Client 1", func(t *testing.T) {
		tlsClientCert1Template, err := CertificateTemplateTLSClient(issuingCert1SubjectName, tlsClientCert1SubjectName, testTlsClientCertDuration, tlsClientCert1DnsNames, tlsClientCert1IpAddresses, tlsClientCert1EmailAddresses, tlsClientCert1URIs)
		verifyCertificateTemplate(t, err, tlsClientCert1Template)

		tlsClientCert1KeyPair = testKeyGenPool.Get()
		tlsClientCert1, tlsClientCert1DER, tlsClientCert1PEM, err = SignCertificate(issuingCert1, issuingCert1KeyPair.Private.(crypto.Signer), tlsClientCert1Template, tlsClientCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyEndEntityCertificate(t, err, tlsClientCert1, tlsClientCert1DER, tlsClientCert1PEM, issuingCert1SubjectName, tlsClientCert1SubjectName, testTlsClientCertDuration, tlsClientCert1DnsNames, tlsClientCert1IpAddresses, tlsClientCert1EmailAddresses, tlsClientCert1URIs)

		verifyCertChain(t, tlsClientCert1, verify1RootsPool, verify1IntermediatesPool)
	})
}
