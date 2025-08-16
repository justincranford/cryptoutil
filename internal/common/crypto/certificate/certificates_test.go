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

type testCASubject struct {
	SubjectName string
	Duration    time.Duration
	MaxPathLen  int
	KeyPair     *cryptoutilKeyGen.KeyPair
	Cert        *x509.Certificate
	DER         []byte
	PEM         []byte
}

type testEndEntitySubject struct {
	SubjectName    string
	Duration       time.Duration
	DNSNames       []string
	IPAddresses    []net.IP
	EmailAddresses []string
	URIs           []*url.URL
	KeyPair        *cryptoutilKeyGen.KeyPair
	Cert           *x509.Certificate
	DER            []byte
	PEM            []byte
}

func TestCertificateChain(t *testing.T) {
	rootCert1 := testCASubject{SubjectName: "Test Root CA 1", Duration: 10 * 365 * 24 * time.Hour, MaxPathLen: 2}
	intermediateCert1 := testCASubject{SubjectName: "Test Intermediate CA 1", Duration: 5 * 365 * 24 * time.Hour, MaxPathLen: 1}
	issuingCert1 := testCASubject{SubjectName: "Test Issuing CA 1", Duration: 2 * 365 * 24 * time.Hour, MaxPathLen: 0}
	tlsServerCert1 := testEndEntitySubject{SubjectName: "TLS Server 1", Duration: 397 * 24 * time.Hour, DNSNames: []string{"localhost", "example.com"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}}
	tlsClientCert1 := testEndEntitySubject{SubjectName: "TLS Client 1", Duration: 30 * 24 * time.Hour, EmailAddresses: []string{"client@example.com"}}

	verify1RootsPool := x509.NewCertPool()
	verify1IntermediatesPool := x509.NewCertPool()

	t.Run("Root CA 1", func(t *testing.T) {
		rootCert1Template, err := CertificateTemplateCA(rootCert1.SubjectName, rootCert1.SubjectName, rootCert1.Duration, rootCert1.MaxPathLen)
		verifyCertificateTemplate(t, err, rootCert1Template)

		rootCert1.KeyPair = testKeyGenPool.Get()
		rootCert1.Cert, rootCert1.DER, rootCert1.PEM, err = SignCertificate(nil, rootCert1.KeyPair.Private.(crypto.Signer), rootCert1Template, rootCert1.KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, rootCert1.Cert, rootCert1.DER, rootCert1.PEM, rootCert1.SubjectName, rootCert1.SubjectName, rootCert1.Duration, rootCert1.MaxPathLen)

		verify1RootsPool.AddCert(rootCert1.Cert) // subsequent verify cert chain needs the root CA
	})

	t.Run("Intermediate CA 1", func(t *testing.T) {
		intermediateCert1Template, err := CertificateTemplateCA(rootCert1.SubjectName, intermediateCert1.SubjectName, intermediateCert1.Duration, intermediateCert1.MaxPathLen)
		verifyCertificateTemplate(t, err, intermediateCert1Template)

		intermediateCert1.KeyPair = testKeyGenPool.Get()
		intermediateCert1.Cert, intermediateCert1.DER, intermediateCert1.PEM, err = SignCertificate(rootCert1.Cert, rootCert1.KeyPair.Private.(crypto.Signer), intermediateCert1Template, intermediateCert1.KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, intermediateCert1.Cert, intermediateCert1.DER, intermediateCert1.PEM, rootCert1.SubjectName, intermediateCert1.SubjectName, intermediateCert1.Duration, intermediateCert1.MaxPathLen)

		verifyCertChain(t, intermediateCert1.Cert, verify1RootsPool, verify1IntermediatesPool)
		verify1IntermediatesPool.AddCert(intermediateCert1.Cert) // subsequent verify cert chain needs the intermediate CA
	})

	t.Run("Issuing CA 1", func(t *testing.T) {
		issuingCert1Template, err := CertificateTemplateCA(intermediateCert1.SubjectName, issuingCert1.SubjectName, issuingCert1.Duration, issuingCert1.MaxPathLen)
		verifyCertificateTemplate(t, err, issuingCert1Template)

		issuingCert1.KeyPair = testKeyGenPool.Get()
		issuingCert1.Cert, issuingCert1.DER, issuingCert1.PEM, err = SignCertificate(intermediateCert1.Cert, intermediateCert1.KeyPair.Private.(crypto.Signer), issuingCert1Template, issuingCert1.KeyPair.Public, x509.ECDSAWithSHA256)
		verifyCACertificate(t, err, issuingCert1.Cert, issuingCert1.DER, issuingCert1.PEM, intermediateCert1.SubjectName, issuingCert1.SubjectName, issuingCert1.Duration, issuingCert1.MaxPathLen)

		verifyCertChain(t, issuingCert1.Cert, verify1RootsPool, verify1IntermediatesPool)
		verify1IntermediatesPool.AddCert(issuingCert1.Cert) // subsequent verify cert chain needs the issuing CA
	})

	t.Run("TLS Server 1", func(t *testing.T) {
		tlsServerCert1Template, err := CertificateTemplateTLSServer(issuingCert1.SubjectName, tlsServerCert1.SubjectName, tlsServerCert1.Duration, tlsServerCert1.DNSNames, tlsServerCert1.IPAddresses, tlsServerCert1.EmailAddresses, tlsServerCert1.URIs)
		verifyCertificateTemplate(t, err, tlsServerCert1Template)

		tlsServerCert1.KeyPair = testKeyGenPool.Get()
		tlsServerCert1.Cert, tlsServerCert1.DER, tlsServerCert1.PEM, err = SignCertificate(issuingCert1.Cert, issuingCert1.KeyPair.Private.(crypto.Signer), tlsServerCert1Template, tlsServerCert1.KeyPair.Public, x509.ECDSAWithSHA256)
		verifyEndEntityCertificate(t, err, tlsServerCert1.Cert, tlsServerCert1.DER, tlsServerCert1.PEM, issuingCert1.SubjectName, tlsServerCert1.SubjectName, tlsServerCert1.Duration, tlsServerCert1.DNSNames, tlsServerCert1.IPAddresses, tlsServerCert1.EmailAddresses, tlsServerCert1.URIs)

		verifyCertChain(t, tlsServerCert1.Cert, verify1RootsPool, verify1IntermediatesPool)
	})

	t.Run("TLS Client 1", func(t *testing.T) {
		tlsClientCert1Template, err := CertificateTemplateTLSClient(issuingCert1.SubjectName, tlsClientCert1.SubjectName, tlsClientCert1.Duration, tlsClientCert1.DNSNames, tlsClientCert1.IPAddresses, tlsClientCert1.EmailAddresses, tlsClientCert1.URIs)
		verifyCertificateTemplate(t, err, tlsClientCert1Template)

		tlsClientCert1.KeyPair = testKeyGenPool.Get()
		tlsClientCert1.Cert, tlsClientCert1.DER, tlsClientCert1.PEM, err = SignCertificate(issuingCert1.Cert, issuingCert1.KeyPair.Private.(crypto.Signer), tlsClientCert1Template, tlsClientCert1.KeyPair.Public, x509.ECDSAWithSHA256)
		verifyEndEntityCertificate(t, err, tlsClientCert1.Cert, tlsClientCert1.DER, tlsClientCert1.PEM, issuingCert1.SubjectName, tlsClientCert1.SubjectName, tlsClientCert1.Duration, tlsClientCert1.DNSNames, tlsClientCert1.IPAddresses, tlsClientCert1.EmailAddresses, tlsClientCert1.URIs)

		verifyCertChain(t, tlsClientCert1.Cert, verify1RootsPool, verify1IntermediatesPool)
	})
}
