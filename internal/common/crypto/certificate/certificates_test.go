package certificate

import (
	"crypto"
	"crypto/x509"
	"net"
	"net/url"
	"testing"
	"time"

	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilDateTime "cryptoutil/internal/common/util/datetime"

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
	KeyMaterial testKeyMaterial
}

type testEndEntitySubject struct {
	SubjectName    string
	Duration       time.Duration
	DNSNames       []string
	IPAddresses    []net.IP
	EmailAddresses []string
	URIs           []*url.URL
	KeyMaterial    testKeyMaterial
}

type testKeyMaterial struct {
	KeyPair *cryptoutilKeyGen.KeyPair
	Cert    *x509.Certificate
	DER     []byte
	PEM     []byte
}

func TestCertificateChain(t *testing.T) {
	rootCert1 := testCASubject{SubjectName: "Test Root CA 1", Duration: 10 * cryptoutilDateTime.Days365, MaxPathLen: 2, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	intermediateCert1 := testCASubject{SubjectName: "Test Intermediate CA 1", Duration: 5 * cryptoutilDateTime.Days365, MaxPathLen: 1, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	issuingCert1 := testCASubject{SubjectName: "Test Issuing CA 1", Duration: 2 * cryptoutilDateTime.Days365, MaxPathLen: 0, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsServerCert1 := testEndEntitySubject{SubjectName: "TLS Server 1", Duration: 397 * cryptoutilDateTime.Days1, DNSNames: []string{"localhost", "server2.example.com"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	verify1RootsPool := x509.NewCertPool()
	verify1IntermediatesPool := x509.NewCertPool()

	t.Run("PKI Chain 1", func(t *testing.T) {
		t.Run("Root CA 1", func(t *testing.T) {
			rootCert1Template, err := CertificateTemplateCA(rootCert1.SubjectName, rootCert1.SubjectName, rootCert1.Duration, rootCert1.MaxPathLen)
			verifyCertificateTemplate(t, err, rootCert1Template)
			rootCert1.KeyMaterial.Cert, rootCert1.KeyMaterial.DER, rootCert1.KeyMaterial.PEM, err = SignCertificate(nil, rootCert1.KeyMaterial.KeyPair.Private.(crypto.Signer), rootCert1Template, rootCert1.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, rootCert1.KeyMaterial.Cert, rootCert1.KeyMaterial.DER, rootCert1.KeyMaterial.PEM, rootCert1.SubjectName, rootCert1.SubjectName, rootCert1.Duration, rootCert1.MaxPathLen)
			verify1RootsPool.AddCert(rootCert1.KeyMaterial.Cert) // subsequent verify cert chain needs the root CA
		})
		t.Run("Intermediate CA 1", func(t *testing.T) {
			intermediateCert1Template, err := CertificateTemplateCA(rootCert1.SubjectName, intermediateCert1.SubjectName, intermediateCert1.Duration, intermediateCert1.MaxPathLen)
			verifyCertificateTemplate(t, err, intermediateCert1Template)
			intermediateCert1.KeyMaterial.Cert, intermediateCert1.KeyMaterial.DER, intermediateCert1.KeyMaterial.PEM, err = SignCertificate(rootCert1.KeyMaterial.Cert, rootCert1.KeyMaterial.KeyPair.Private.(crypto.Signer), intermediateCert1Template, intermediateCert1.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, intermediateCert1.KeyMaterial.Cert, intermediateCert1.KeyMaterial.DER, intermediateCert1.KeyMaterial.PEM, rootCert1.SubjectName, intermediateCert1.SubjectName, intermediateCert1.Duration, intermediateCert1.MaxPathLen)
			verifyCertChain(t, intermediateCert1.KeyMaterial.Cert, verify1RootsPool, verify1IntermediatesPool)
			verify1IntermediatesPool.AddCert(intermediateCert1.KeyMaterial.Cert) // subsequent verify cert chain needs the intermediate CA
		})
		t.Run("Issuing CA 1", func(t *testing.T) {
			issuingCert1Template, err := CertificateTemplateCA(intermediateCert1.SubjectName, issuingCert1.SubjectName, issuingCert1.Duration, issuingCert1.MaxPathLen)
			verifyCertificateTemplate(t, err, issuingCert1Template)
			issuingCert1.KeyMaterial.Cert, issuingCert1.KeyMaterial.DER, issuingCert1.KeyMaterial.PEM, err = SignCertificate(intermediateCert1.KeyMaterial.Cert, intermediateCert1.KeyMaterial.KeyPair.Private.(crypto.Signer), issuingCert1Template, issuingCert1.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, issuingCert1.KeyMaterial.Cert, issuingCert1.KeyMaterial.DER, issuingCert1.KeyMaterial.PEM, intermediateCert1.SubjectName, issuingCert1.SubjectName, issuingCert1.Duration, issuingCert1.MaxPathLen)
			verifyCertChain(t, issuingCert1.KeyMaterial.Cert, verify1RootsPool, verify1IntermediatesPool)
			verify1IntermediatesPool.AddCert(issuingCert1.KeyMaterial.Cert) // subsequent verify cert chain needs the issuing CA
		})
		t.Run("TLS Server 1", func(t *testing.T) {
			tlsServerCert1Template, err := CertificateTemplateTLSServer(issuingCert1.SubjectName, tlsServerCert1.SubjectName, tlsServerCert1.Duration, tlsServerCert1.DNSNames, tlsServerCert1.IPAddresses, tlsServerCert1.EmailAddresses, tlsServerCert1.URIs)
			verifyCertificateTemplate(t, err, tlsServerCert1Template)
			tlsServerCert1.KeyMaterial.Cert, tlsServerCert1.KeyMaterial.DER, tlsServerCert1.KeyMaterial.PEM, err = SignCertificate(issuingCert1.KeyMaterial.Cert, issuingCert1.KeyMaterial.KeyPair.Private.(crypto.Signer), tlsServerCert1Template, tlsServerCert1.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyEndEntityCertificate(t, err, tlsServerCert1.KeyMaterial.Cert, tlsServerCert1.KeyMaterial.DER, tlsServerCert1.KeyMaterial.PEM, issuingCert1.SubjectName, tlsServerCert1.SubjectName, tlsServerCert1.Duration, tlsServerCert1.DNSNames, tlsServerCert1.IPAddresses, tlsServerCert1.EmailAddresses, tlsServerCert1.URIs)
			verifyCertChain(t, tlsServerCert1.KeyMaterial.Cert, verify1RootsPool, verify1IntermediatesPool)
		})
	})

	rootCert2 := testCASubject{SubjectName: "Test Root CA 2", Duration: 10 * cryptoutilDateTime.Days365, MaxPathLen: 2, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	intermediateCert2 := testCASubject{SubjectName: "Test Intermediate CA 2", Duration: 5 * cryptoutilDateTime.Days365, MaxPathLen: 1, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	issuingCert2 := testCASubject{SubjectName: "Test Issuing CA 2", Duration: 2 * cryptoutilDateTime.Days365, MaxPathLen: 0, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientCert2 := testEndEntitySubject{SubjectName: "TLS Client 2", Duration: 30 * cryptoutilDateTime.Days1, EmailAddresses: []string{"client2@client.example.com"}, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	verify2RootsPool := x509.NewCertPool()
	verify2IntermediatesPool := x509.NewCertPool()
	t.Run("PKI Chain 2", func(t *testing.T) {
		t.Run("Root CA 2", func(t *testing.T) {
			rootCert2Template, err := CertificateTemplateCA(rootCert2.SubjectName, rootCert2.SubjectName, rootCert2.Duration, rootCert2.MaxPathLen)
			verifyCertificateTemplate(t, err, rootCert2Template)
			rootCert2.KeyMaterial.Cert, rootCert2.KeyMaterial.DER, rootCert2.KeyMaterial.PEM, err = SignCertificate(nil, rootCert2.KeyMaterial.KeyPair.Private.(crypto.Signer), rootCert2Template, rootCert2.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, rootCert2.KeyMaterial.Cert, rootCert2.KeyMaterial.DER, rootCert2.KeyMaterial.PEM, rootCert2.SubjectName, rootCert2.SubjectName, rootCert2.Duration, rootCert2.MaxPathLen)
			verify2RootsPool.AddCert(rootCert2.KeyMaterial.Cert) // subsequent verify cert chain needs the root CA
		})
		t.Run("Intermediate CA 2", func(t *testing.T) {
			intermediateCert2Template, err := CertificateTemplateCA(rootCert2.SubjectName, intermediateCert2.SubjectName, intermediateCert2.Duration, intermediateCert2.MaxPathLen)
			verifyCertificateTemplate(t, err, intermediateCert2Template)
			intermediateCert2.KeyMaterial.Cert, intermediateCert2.KeyMaterial.DER, intermediateCert2.KeyMaterial.PEM, err = SignCertificate(rootCert2.KeyMaterial.Cert, rootCert2.KeyMaterial.KeyPair.Private.(crypto.Signer), intermediateCert2Template, intermediateCert2.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, intermediateCert2.KeyMaterial.Cert, intermediateCert2.KeyMaterial.DER, intermediateCert2.KeyMaterial.PEM, rootCert2.SubjectName, intermediateCert2.SubjectName, intermediateCert2.Duration, intermediateCert2.MaxPathLen)
			verifyCertChain(t, intermediateCert2.KeyMaterial.Cert, verify2RootsPool, verify2IntermediatesPool)
			verify2IntermediatesPool.AddCert(intermediateCert2.KeyMaterial.Cert) // subsequent verify cert chain needs the intermediate CA
		})
		t.Run("Issuing CA 2", func(t *testing.T) {
			issuingCert2Template, err := CertificateTemplateCA(intermediateCert2.SubjectName, issuingCert2.SubjectName, issuingCert2.Duration, issuingCert2.MaxPathLen)
			verifyCertificateTemplate(t, err, issuingCert2Template)
			issuingCert2.KeyMaterial.Cert, issuingCert2.KeyMaterial.DER, issuingCert2.KeyMaterial.PEM, err = SignCertificate(intermediateCert2.KeyMaterial.Cert, intermediateCert2.KeyMaterial.KeyPair.Private.(crypto.Signer), issuingCert2Template, issuingCert2.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, issuingCert2.KeyMaterial.Cert, issuingCert2.KeyMaterial.DER, issuingCert2.KeyMaterial.PEM, intermediateCert2.SubjectName, issuingCert2.SubjectName, issuingCert2.Duration, issuingCert2.MaxPathLen)
			verifyCertChain(t, issuingCert2.KeyMaterial.Cert, verify2RootsPool, verify2IntermediatesPool)
			verify2IntermediatesPool.AddCert(issuingCert2.KeyMaterial.Cert) // subsequent verify cert chain needs the issuing CA
		})
		t.Run("TLS Client 2", func(t *testing.T) {
			tlsClientCert2Template, err := CertificateTemplateTLSClient(issuingCert2.SubjectName, tlsClientCert2.SubjectName, tlsClientCert2.Duration, tlsClientCert2.DNSNames, tlsClientCert2.IPAddresses, tlsClientCert2.EmailAddresses, tlsClientCert2.URIs)
			verifyCertificateTemplate(t, err, tlsClientCert2Template)

			tlsClientCert2.KeyMaterial.Cert, tlsClientCert2.KeyMaterial.DER, tlsClientCert2.KeyMaterial.PEM, err = SignCertificate(issuingCert2.KeyMaterial.Cert, issuingCert2.KeyMaterial.KeyPair.Private.(crypto.Signer), tlsClientCert2Template, tlsClientCert2.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyEndEntityCertificate(t, err, tlsClientCert2.KeyMaterial.Cert, tlsClientCert2.KeyMaterial.DER, tlsClientCert2.KeyMaterial.PEM, issuingCert2.SubjectName, tlsClientCert2.SubjectName, tlsClientCert2.Duration, tlsClientCert2.DNSNames, tlsClientCert2.IPAddresses, tlsClientCert2.EmailAddresses, tlsClientCert2.URIs)
			verifyCertChain(t, tlsClientCert2.KeyMaterial.Cert, verify2RootsPool, verify2IntermediatesPool)
		})
	})
}
