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

	tlsServerCert1SubjectName := "TLS Server 1"
	tlsServerCert1Duration := 397 * 24 * time.Hour // 398 days max
	tlsServerCert1DnsNames := []string{}
	tlsServerCert1IpAddresses := []net.IP{}
	tlsServerCert1EmailAddresses := []string{}
	tlsServerCert1URIs := []*url.URL{}

	tlsClientCert1SubjectName := "TLS Client 1"
	tlsClientCert1Duration := 397 * 24 * time.Hour // 398 days max
	tlsClientCert1DnsNames := []string{}
	tlsClientCert1IpAddresses := []net.IP{}
	tlsClientCert1EmailAddresses := []string{}
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

	roots1Pool := x509.NewCertPool()
	intermediates1Pool := x509.NewCertPool()

	t.Run("Root CA 1", func(t *testing.T) {
		rootCert1Template, err := CertificateTemplateCA(rootCert1SubjectName, rootCert1SubjectName, rootCert1Duration, rootCert1MaxPathLen)
		verifyCertificateTemplate(t, err, rootCert1Template)

		rootCert1KeyPair = testKeyGenPool.Get()
		rootCert1, rootCert1DER, rootCert1PEM, err = SignCertificate(nil, rootCert1KeyPair.Private.(crypto.Signer), rootCert1Template, rootCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, rootCert1, rootCert1DER, rootCert1PEM, rootCert1SubjectName, rootCert1SubjectName, rootCert1Duration, rootCert1MaxPathLen)

		roots1Pool.AddCert(rootCert1)
	})

	t.Run("Intermediate CA 1", func(t *testing.T) {
		intermediateCert1Template, err := CertificateTemplateCA(rootCert1SubjectName, intermediateCert1SubjectName, intermediateCert1Duration, intermediateCert1MaxPathLen)
		verifyCertificateTemplate(t, err, intermediateCert1Template)

		intermediateCert1KeyPair = testKeyGenPool.Get()
		intermediateCert1, intermediateCert1DER, intermediateCert1PEM, err = SignCertificate(rootCert1, rootCert1KeyPair.Private.(crypto.Signer), intermediateCert1Template, intermediateCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, intermediateCert1, intermediateCert1DER, intermediateCert1PEM, rootCert1SubjectName, intermediateCert1SubjectName, intermediateCert1Duration, intermediateCert1MaxPathLen)

		verifyCertChain(t, intermediateCert1, roots1Pool, intermediates1Pool)
		intermediates1Pool.AddCert(intermediateCert1)
	})

	t.Run("Issuing CA 1", func(t *testing.T) {
		issuingCert1Template, err := CertificateTemplateCA(rootCert1SubjectName, issuingCert1SubjectName, issuingCert1Duration, issuingCert1MaxPathLen)
		verifyCertificateTemplate(t, err, issuingCert1Template)

		issuingCert1KeyPair = testKeyGenPool.Get()
		issuingCert1, issuingCert1DER, issuingCert1PEM, err = SignCertificate(rootCert1, rootCert1KeyPair.Private.(crypto.Signer), issuingCert1Template, issuingCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, issuingCert1, issuingCert1DER, issuingCert1PEM, rootCert1SubjectName, issuingCert1SubjectName, issuingCert1Duration, issuingCert1MaxPathLen)

		verifyCertChain(t, issuingCert1, roots1Pool, x509.NewCertPool())
		intermediates1Pool.AddCert(issuingCert1)
	})

	t.Run("TLS Server 1", func(t *testing.T) {
		tlsServerCert1Template, err := CertificateTemplateTLSServer(rootCert1SubjectName, tlsServerCert1SubjectName, tlsServerCert1Duration, tlsServerCert1DnsNames, tlsServerCert1IpAddresses, tlsServerCert1EmailAddresses, tlsServerCert1URIs)
		verifyCertificateTemplate(t, err, tlsServerCert1Template)

		tlsServerCert1KeyPair = testKeyGenPool.Get()
		tlsServerCert1, tlsServerCert1DER, tlsServerCert1PEM, err = SignCertificate(rootCert1, rootCert1KeyPair.Private.(crypto.Signer), tlsServerCert1Template, tlsServerCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyEndEntityCertificate(t, err, tlsServerCert1, tlsServerCert1DER, tlsServerCert1PEM, rootCert1SubjectName, tlsServerCert1SubjectName, tlsServerCert1Duration, tlsServerCert1DnsNames, tlsServerCert1IpAddresses, tlsServerCert1EmailAddresses, tlsServerCert1URIs)

		verifyCertChain(t, tlsServerCert1, roots1Pool, x509.NewCertPool())
	})

	t.Run("TLS Client 1", func(t *testing.T) {
		tlsClientCert1Template, err := CertificateTemplateTLSClient(rootCert1SubjectName, tlsClientCert1SubjectName, tlsClientCert1Duration, tlsClientCert1DnsNames, tlsClientCert1IpAddresses, tlsClientCert1EmailAddresses, tlsClientCert1URIs)
		verifyCertificateTemplate(t, err, tlsClientCert1Template)

		tlsClientCert1KeyPair = testKeyGenPool.Get()
		tlsClientCert1, tlsClientCert1DER, tlsClientCert1PEM, err = SignCertificate(rootCert1, rootCert1KeyPair.Private.(crypto.Signer), tlsClientCert1Template, tlsClientCert1KeyPair.Public, x509.ECDSAWithSHA256)
		verifyEndEntityCertificate(t, err, tlsClientCert1, tlsClientCert1DER, tlsClientCert1PEM, rootCert1SubjectName, tlsClientCert1SubjectName, tlsClientCert1Duration, tlsClientCert1DnsNames, tlsClientCert1IpAddresses, tlsClientCert1EmailAddresses, tlsClientCert1URIs)

		verifyCertChain(t, tlsClientCert1, roots1Pool, x509.NewCertPool())
	})
}
