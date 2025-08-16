package certificate

import (
	"crypto"
	"crypto/tls"
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
	tlsServerRootsPool := x509.NewCertPool()
	tlsServerIntermediatesPool := x509.NewCertPool()

	t.Run("PKI Chain 1", func(t *testing.T) {
		t.Run("Root CA 1", func(t *testing.T) {
			rootCert1Template, err := CertificateTemplateCA(rootCert1.SubjectName, rootCert1.SubjectName, rootCert1.Duration, rootCert1.MaxPathLen)
			verifyCertificateTemplate(t, err, rootCert1Template)
			rootCert1.KeyMaterial.Cert, rootCert1.KeyMaterial.DER, rootCert1.KeyMaterial.PEM, err = SignCertificate(nil, rootCert1.KeyMaterial.KeyPair.Private.(crypto.Signer), rootCert1Template, rootCert1.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, rootCert1.KeyMaterial.Cert, rootCert1.KeyMaterial.DER, rootCert1.KeyMaterial.PEM, rootCert1.SubjectName, rootCert1.SubjectName, rootCert1.Duration, rootCert1.MaxPathLen)
			tlsServerRootsPool.AddCert(rootCert1.KeyMaterial.Cert) // subsequent verify cert chain needs the root CA
		})
		t.Run("Intermediate CA 1", func(t *testing.T) {
			intermediateCert1Template, err := CertificateTemplateCA(rootCert1.SubjectName, intermediateCert1.SubjectName, intermediateCert1.Duration, intermediateCert1.MaxPathLen)
			verifyCertificateTemplate(t, err, intermediateCert1Template)
			intermediateCert1.KeyMaterial.Cert, intermediateCert1.KeyMaterial.DER, intermediateCert1.KeyMaterial.PEM, err = SignCertificate(rootCert1.KeyMaterial.Cert, rootCert1.KeyMaterial.KeyPair.Private.(crypto.Signer), intermediateCert1Template, intermediateCert1.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, intermediateCert1.KeyMaterial.Cert, intermediateCert1.KeyMaterial.DER, intermediateCert1.KeyMaterial.PEM, rootCert1.SubjectName, intermediateCert1.SubjectName, intermediateCert1.Duration, intermediateCert1.MaxPathLen)
			verifyCertChain(t, intermediateCert1.KeyMaterial.Cert, tlsServerRootsPool, tlsServerIntermediatesPool)
			tlsServerIntermediatesPool.AddCert(intermediateCert1.KeyMaterial.Cert) // subsequent verify cert chain needs the intermediate CA
		})
		t.Run("Issuing CA 1", func(t *testing.T) {
			issuingCert1Template, err := CertificateTemplateCA(intermediateCert1.SubjectName, issuingCert1.SubjectName, issuingCert1.Duration, issuingCert1.MaxPathLen)
			verifyCertificateTemplate(t, err, issuingCert1Template)
			issuingCert1.KeyMaterial.Cert, issuingCert1.KeyMaterial.DER, issuingCert1.KeyMaterial.PEM, err = SignCertificate(intermediateCert1.KeyMaterial.Cert, intermediateCert1.KeyMaterial.KeyPair.Private.(crypto.Signer), issuingCert1Template, issuingCert1.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, issuingCert1.KeyMaterial.Cert, issuingCert1.KeyMaterial.DER, issuingCert1.KeyMaterial.PEM, intermediateCert1.SubjectName, issuingCert1.SubjectName, issuingCert1.Duration, issuingCert1.MaxPathLen)
			verifyCertChain(t, issuingCert1.KeyMaterial.Cert, tlsServerRootsPool, tlsServerIntermediatesPool)
			tlsServerIntermediatesPool.AddCert(issuingCert1.KeyMaterial.Cert) // subsequent verify cert chain needs the issuing CA
		})
		t.Run("TLS Server 1", func(t *testing.T) {
			tlsServerCert1Template, err := CertificateTemplateTLSServer(issuingCert1.SubjectName, tlsServerCert1.SubjectName, tlsServerCert1.Duration, tlsServerCert1.DNSNames, tlsServerCert1.IPAddresses, tlsServerCert1.EmailAddresses, tlsServerCert1.URIs)
			verifyCertificateTemplate(t, err, tlsServerCert1Template)
			tlsServerCert1.KeyMaterial.Cert, tlsServerCert1.KeyMaterial.DER, tlsServerCert1.KeyMaterial.PEM, err = SignCertificate(issuingCert1.KeyMaterial.Cert, issuingCert1.KeyMaterial.KeyPair.Private.(crypto.Signer), tlsServerCert1Template, tlsServerCert1.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyEndEntityCertificate(t, err, tlsServerCert1.KeyMaterial.Cert, tlsServerCert1.KeyMaterial.DER, tlsServerCert1.KeyMaterial.PEM, issuingCert1.SubjectName, tlsServerCert1.SubjectName, tlsServerCert1.Duration, tlsServerCert1.DNSNames, tlsServerCert1.IPAddresses, tlsServerCert1.EmailAddresses, tlsServerCert1.URIs)
			verifyCertChain(t, tlsServerCert1.KeyMaterial.Cert, tlsServerRootsPool, tlsServerIntermediatesPool)
		})
	})

	rootCert2 := testCASubject{SubjectName: "Test Root CA 2", Duration: 10 * cryptoutilDateTime.Days365, MaxPathLen: 2, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	intermediateCert2 := testCASubject{SubjectName: "Test Intermediate CA 2", Duration: 5 * cryptoutilDateTime.Days365, MaxPathLen: 1, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	issuingCert2 := testCASubject{SubjectName: "Test Issuing CA 2", Duration: 2 * cryptoutilDateTime.Days365, MaxPathLen: 0, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientCert2 := testEndEntitySubject{SubjectName: "TLS Client 2", Duration: 30 * cryptoutilDateTime.Days1, EmailAddresses: []string{"client2@client.example.com"}, KeyMaterial: testKeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientRootsPool := x509.NewCertPool()
	tlsClientIntermediatesPool := x509.NewCertPool()
	t.Run("PKI Chain 2", func(t *testing.T) {
		t.Run("Root CA 2", func(t *testing.T) {
			rootCert2Template, err := CertificateTemplateCA(rootCert2.SubjectName, rootCert2.SubjectName, rootCert2.Duration, rootCert2.MaxPathLen)
			verifyCertificateTemplate(t, err, rootCert2Template)
			rootCert2.KeyMaterial.Cert, rootCert2.KeyMaterial.DER, rootCert2.KeyMaterial.PEM, err = SignCertificate(nil, rootCert2.KeyMaterial.KeyPair.Private.(crypto.Signer), rootCert2Template, rootCert2.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, rootCert2.KeyMaterial.Cert, rootCert2.KeyMaterial.DER, rootCert2.KeyMaterial.PEM, rootCert2.SubjectName, rootCert2.SubjectName, rootCert2.Duration, rootCert2.MaxPathLen)
			tlsClientRootsPool.AddCert(rootCert2.KeyMaterial.Cert) // subsequent verify cert chain needs the root CA
		})
		t.Run("Intermediate CA 2", func(t *testing.T) {
			intermediateCert2Template, err := CertificateTemplateCA(rootCert2.SubjectName, intermediateCert2.SubjectName, intermediateCert2.Duration, intermediateCert2.MaxPathLen)
			verifyCertificateTemplate(t, err, intermediateCert2Template)
			intermediateCert2.KeyMaterial.Cert, intermediateCert2.KeyMaterial.DER, intermediateCert2.KeyMaterial.PEM, err = SignCertificate(rootCert2.KeyMaterial.Cert, rootCert2.KeyMaterial.KeyPair.Private.(crypto.Signer), intermediateCert2Template, intermediateCert2.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, intermediateCert2.KeyMaterial.Cert, intermediateCert2.KeyMaterial.DER, intermediateCert2.KeyMaterial.PEM, rootCert2.SubjectName, intermediateCert2.SubjectName, intermediateCert2.Duration, intermediateCert2.MaxPathLen)
			verifyCertChain(t, intermediateCert2.KeyMaterial.Cert, tlsClientRootsPool, tlsClientIntermediatesPool)
			tlsClientIntermediatesPool.AddCert(intermediateCert2.KeyMaterial.Cert) // subsequent verify cert chain needs the intermediate CA
		})
		t.Run("Issuing CA 2", func(t *testing.T) {
			issuingCert2Template, err := CertificateTemplateCA(intermediateCert2.SubjectName, issuingCert2.SubjectName, issuingCert2.Duration, issuingCert2.MaxPathLen)
			verifyCertificateTemplate(t, err, issuingCert2Template)
			issuingCert2.KeyMaterial.Cert, issuingCert2.KeyMaterial.DER, issuingCert2.KeyMaterial.PEM, err = SignCertificate(intermediateCert2.KeyMaterial.Cert, intermediateCert2.KeyMaterial.KeyPair.Private.(crypto.Signer), issuingCert2Template, issuingCert2.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, issuingCert2.KeyMaterial.Cert, issuingCert2.KeyMaterial.DER, issuingCert2.KeyMaterial.PEM, intermediateCert2.SubjectName, issuingCert2.SubjectName, issuingCert2.Duration, issuingCert2.MaxPathLen)
			verifyCertChain(t, issuingCert2.KeyMaterial.Cert, tlsClientRootsPool, tlsClientIntermediatesPool)
			tlsClientIntermediatesPool.AddCert(issuingCert2.KeyMaterial.Cert) // subsequent verify cert chain needs the issuing CA
		})
		t.Run("TLS Client 2", func(t *testing.T) {
			tlsClientCert2Template, err := CertificateTemplateTLSClient(issuingCert2.SubjectName, tlsClientCert2.SubjectName, tlsClientCert2.Duration, tlsClientCert2.DNSNames, tlsClientCert2.IPAddresses, tlsClientCert2.EmailAddresses, tlsClientCert2.URIs)
			verifyCertificateTemplate(t, err, tlsClientCert2Template)

			tlsClientCert2.KeyMaterial.Cert, tlsClientCert2.KeyMaterial.DER, tlsClientCert2.KeyMaterial.PEM, err = SignCertificate(issuingCert2.KeyMaterial.Cert, issuingCert2.KeyMaterial.KeyPair.Private.(crypto.Signer), tlsClientCert2Template, tlsClientCert2.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyEndEntityCertificate(t, err, tlsClientCert2.KeyMaterial.Cert, tlsClientCert2.KeyMaterial.DER, tlsClientCert2.KeyMaterial.PEM, issuingCert2.SubjectName, tlsClientCert2.SubjectName, tlsClientCert2.Duration, tlsClientCert2.DNSNames, tlsClientCert2.IPAddresses, tlsClientCert2.EmailAddresses, tlsClientCert2.URIs)
			verifyCertChain(t, tlsClientCert2.KeyMaterial.Cert, tlsClientRootsPool, tlsClientIntermediatesPool)
		})
	})

	t.Run("TLS Mutual Authentication", func(t *testing.T) {
		serverTLSCert := tls.Certificate{
			Certificate: [][]byte{tlsServerCert1.KeyMaterial.DER, issuingCert1.KeyMaterial.DER, intermediateCert1.KeyMaterial.DER, rootCert1.KeyMaterial.DER},
			PrivateKey:  tlsServerCert1.KeyMaterial.KeyPair.Private,
		}
		clientTLSCert := tls.Certificate{
			Certificate: [][]byte{tlsClientCert2.KeyMaterial.DER, issuingCert2.KeyMaterial.DER, intermediateCert2.KeyMaterial.DER, rootCert2.KeyMaterial.DER},
			PrivateKey:  tlsClientCert2.KeyMaterial.KeyPair.Private,
		}

		tlsServer := &tls.Config{
			Certificates: []tls.Certificate{serverTLSCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    tlsClientRootsPool,
		}
		ln, err := tls.Listen("tcp", "127.0.0.1:0", tlsServer)
		require.NoError(t, err, "Failed to start TLS server")
		defer ln.Close()
		serverErrCh := make(chan error, 1)
		go func() {
			conn, err := ln.Accept()
			if err != nil {
				serverErrCh <- err
				return
			}
			defer conn.Close()
			buf := make([]byte, 512)
			n, err := conn.Read(buf)
			if err != nil {
				serverErrCh <- err
				return
			}
			_, err = conn.Write(buf[:n])
			serverErrCh <- err
		}()

		clientTLSConfig := &tls.Config{
			Certificates:       []tls.Certificate{clientTLSCert},
			RootCAs:            tlsServerRootsPool,
			InsecureSkipVerify: false,
		}
		addr := ln.Addr().String()
		conn, err := tls.Dial("tcp", addr, clientTLSConfig)
		require.NoError(t, err, "Client failed to connect to TLS server")
		defer conn.Close()

		testMsg := []byte("hello mutual tls")
		_, err = conn.Write(testMsg)
		require.NoError(t, err, "Client failed to write to server")
		resp := make([]byte, len(testMsg))
		_, err = conn.Read(resp)
		require.NoError(t, err, "Client failed to read from server")
		require.Equal(t, testMsg, resp, "Echoed message mismatch")

		// Ensure server goroutine completed without error
		require.NoError(t, <-serverErrCh, "Server goroutine error")
	})
}
