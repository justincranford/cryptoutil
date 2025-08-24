package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/url"
	"testing"
	"time"

	"bytes"
	"io"
	"net/http"

	cryptoutilDateTime "cryptoutil/internal/common/util/datetime"

	"github.com/stretchr/testify/require"
)

func TestNegativeDuration(t *testing.T) {
	_, err := CertificateTemplateCA("Root CA", "Root CA", -1, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}

type testCASubject struct {
	SubjectName string
	Duration    time.Duration
	MaxPathLen  int
	KeyMaterial KeyMaterial
}

type testEndEntitySubject struct {
	SubjectName    string
	Duration       time.Duration
	DNSNames       []string
	IPAddresses    []net.IP
	EmailAddresses []string
	URIs           []*url.URL
	KeyMaterial    KeyMaterial
}

func TestMutualTLS(t *testing.T) {
	tlsServerRootCACert := testCASubject{SubjectName: "Test TLS Server Root CA", Duration: 10 * cryptoutilDateTime.Days365, MaxPathLen: 2, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsServerSubordinateCA1Cert := testCASubject{SubjectName: "Test TLS Server Subordinate CA 1", Duration: 5 * cryptoutilDateTime.Days365, MaxPathLen: 1, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsServerSubordinateCA2Cert := testCASubject{SubjectName: "Test TLS Server Subordinate CA 2", Duration: 2 * cryptoutilDateTime.Days365, MaxPathLen: 0, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsServerRootCAsPool := x509.NewCertPool()
	tlsServerSubordinateCAsPool := x509.NewCertPool()
	tlsServerEndEntityCert := testEndEntitySubject{SubjectName: "Test TLS Server End Entity", Duration: 397 * cryptoutilDateTime.Days1, DNSNames: []string{"localhost", "tlsserver.example.com"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}

	t.Run("TLS Server Chain", func(t *testing.T) {
		t.Run("TLS Server Root CA", func(t *testing.T) {
			tlsServerRootCACertTemplate, err := CertificateTemplateCA(tlsServerRootCACert.SubjectName, tlsServerRootCACert.SubjectName, tlsServerRootCACert.Duration, tlsServerRootCACert.MaxPathLen)
			verifyCertificateTemplate(t, err, tlsServerRootCACertTemplate)
			tlsServerRootCACert.KeyMaterial.Cert, tlsServerRootCACert.KeyMaterial.DER, tlsServerRootCACert.KeyMaterial.PEM, err = SignCertificate(nil, tlsServerRootCACert.KeyMaterial.KeyPair.Private, tlsServerRootCACertTemplate, tlsServerRootCACert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsServerRootCACert.KeyMaterial.Cert, tlsServerRootCACert.KeyMaterial.DER, tlsServerRootCACert.KeyMaterial.PEM, tlsServerRootCACert.SubjectName, tlsServerRootCACert.SubjectName, tlsServerRootCACert.Duration, tlsServerRootCACert.MaxPathLen)
			tlsServerRootCAsPool.AddCert(tlsServerRootCACert.KeyMaterial.Cert) // subsequent verify cert chain needs the root CA
		})
		t.Run("TLS Server Subordinate CA 1", func(t *testing.T) {
			tlsServerSubordinateCA1CertTemplate, err := CertificateTemplateCA(tlsServerRootCACert.SubjectName, tlsServerSubordinateCA1Cert.SubjectName, tlsServerSubordinateCA1Cert.Duration, tlsServerSubordinateCA1Cert.MaxPathLen)
			verifyCertificateTemplate(t, err, tlsServerSubordinateCA1CertTemplate)
			tlsServerSubordinateCA1Cert.KeyMaterial.Cert, tlsServerSubordinateCA1Cert.KeyMaterial.DER, tlsServerSubordinateCA1Cert.KeyMaterial.PEM, err = SignCertificate(tlsServerRootCACert.KeyMaterial.Cert, tlsServerRootCACert.KeyMaterial.KeyPair.Private, tlsServerSubordinateCA1CertTemplate, tlsServerSubordinateCA1Cert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsServerSubordinateCA1Cert.KeyMaterial.Cert, tlsServerSubordinateCA1Cert.KeyMaterial.DER, tlsServerSubordinateCA1Cert.KeyMaterial.PEM, tlsServerRootCACert.SubjectName, tlsServerSubordinateCA1Cert.SubjectName, tlsServerSubordinateCA1Cert.Duration, tlsServerSubordinateCA1Cert.MaxPathLen)
			verifyCertChain(t, tlsServerSubordinateCA1Cert.KeyMaterial.Cert, tlsServerRootCAsPool, tlsServerSubordinateCAsPool)
			tlsServerSubordinateCAsPool.AddCert(tlsServerSubordinateCA1Cert.KeyMaterial.Cert) // subsequent verify cert chain needs the intermediate CA
		})
		t.Run("TLS Server Subordinate CA 2", func(t *testing.T) {
			tlsServerSubordinateCA2CertTemplate, err := CertificateTemplateCA(tlsServerSubordinateCA1Cert.SubjectName, tlsServerSubordinateCA2Cert.SubjectName, tlsServerSubordinateCA2Cert.Duration, tlsServerSubordinateCA2Cert.MaxPathLen)
			verifyCertificateTemplate(t, err, tlsServerSubordinateCA2CertTemplate)
			tlsServerSubordinateCA2Cert.KeyMaterial.Cert, tlsServerSubordinateCA2Cert.KeyMaterial.DER, tlsServerSubordinateCA2Cert.KeyMaterial.PEM, err = SignCertificate(tlsServerSubordinateCA1Cert.KeyMaterial.Cert, tlsServerSubordinateCA1Cert.KeyMaterial.KeyPair.Private, tlsServerSubordinateCA2CertTemplate, tlsServerSubordinateCA2Cert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsServerSubordinateCA2Cert.KeyMaterial.Cert, tlsServerSubordinateCA2Cert.KeyMaterial.DER, tlsServerSubordinateCA2Cert.KeyMaterial.PEM, tlsServerSubordinateCA1Cert.SubjectName, tlsServerSubordinateCA2Cert.SubjectName, tlsServerSubordinateCA2Cert.Duration, tlsServerSubordinateCA2Cert.MaxPathLen)
			verifyCertChain(t, tlsServerSubordinateCA2Cert.KeyMaterial.Cert, tlsServerRootCAsPool, tlsServerSubordinateCAsPool)
			tlsServerSubordinateCAsPool.AddCert(tlsServerSubordinateCA2Cert.KeyMaterial.Cert) // subsequent verify cert chain needs the issuing CA
		})
		t.Run("TLS Server End Entity", func(t *testing.T) {
			tlsServerEndEntityCertTemplate, err := CertificateTemplateTLSServer(tlsServerSubordinateCA2Cert.SubjectName, tlsServerEndEntityCert.SubjectName, tlsServerEndEntityCert.Duration, tlsServerEndEntityCert.DNSNames, tlsServerEndEntityCert.IPAddresses, tlsServerEndEntityCert.EmailAddresses, tlsServerEndEntityCert.URIs)
			verifyCertificateTemplate(t, err, tlsServerEndEntityCertTemplate)
			tlsServerEndEntityCert.KeyMaterial.Cert, tlsServerEndEntityCert.KeyMaterial.DER, tlsServerEndEntityCert.KeyMaterial.PEM, err = SignCertificate(tlsServerSubordinateCA2Cert.KeyMaterial.Cert, tlsServerSubordinateCA2Cert.KeyMaterial.KeyPair.Private, tlsServerEndEntityCertTemplate, tlsServerEndEntityCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyEndEntityCertificate(t, err, tlsServerEndEntityCert.KeyMaterial.Cert, tlsServerEndEntityCert.KeyMaterial.DER, tlsServerEndEntityCert.KeyMaterial.PEM, tlsServerSubordinateCA2Cert.SubjectName, tlsServerEndEntityCert.SubjectName, tlsServerEndEntityCert.Duration, tlsServerEndEntityCert.DNSNames, tlsServerEndEntityCert.IPAddresses, tlsServerEndEntityCert.EmailAddresses, tlsServerEndEntityCert.URIs)
			verifyCertChain(t, tlsServerEndEntityCert.KeyMaterial.Cert, tlsServerRootCAsPool, tlsServerSubordinateCAsPool)
		})
	})

	tlsClientRootCACert := testCASubject{SubjectName: "Test TLS Client Root CA", Duration: 10 * cryptoutilDateTime.Days365, MaxPathLen: 2, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientSubordinateCA1Cert := testCASubject{SubjectName: "Test TLS Client Subordinate CA 1", Duration: 5 * cryptoutilDateTime.Days365, MaxPathLen: 1, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientSubordinateCA2Cert := testCASubject{SubjectName: "Test TLS Client Subordinate CA 2", Duration: 2 * cryptoutilDateTime.Days365, MaxPathLen: 0, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientRootCAsPool := x509.NewCertPool()
	tlsClientSubordinateCAsPool := x509.NewCertPool()
	tlsClientEndEntityCert := testEndEntitySubject{SubjectName: "TLS Client", Duration: 30 * cryptoutilDateTime.Days1, EmailAddresses: []string{"client1@tlsclient.example.com"}, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	t.Run("TLS Client Chain", func(t *testing.T) {
		t.Run("TLS Client Root CA", func(t *testing.T) {
			tlsClientRootCACertTemplate, err := CertificateTemplateCA(tlsClientRootCACert.SubjectName, tlsClientRootCACert.SubjectName, tlsClientRootCACert.Duration, tlsClientRootCACert.MaxPathLen)
			verifyCertificateTemplate(t, err, tlsClientRootCACertTemplate)
			tlsClientRootCACert.KeyMaterial.Cert, tlsClientRootCACert.KeyMaterial.DER, tlsClientRootCACert.KeyMaterial.PEM, err = SignCertificate(nil, tlsClientRootCACert.KeyMaterial.KeyPair.Private, tlsClientRootCACertTemplate, tlsClientRootCACert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsClientRootCACert.KeyMaterial.Cert, tlsClientRootCACert.KeyMaterial.DER, tlsClientRootCACert.KeyMaterial.PEM, tlsClientRootCACert.SubjectName, tlsClientRootCACert.SubjectName, tlsClientRootCACert.Duration, tlsClientRootCACert.MaxPathLen)
			tlsClientRootCAsPool.AddCert(tlsClientRootCACert.KeyMaterial.Cert) // subsequent verify cert chain needs the root CA
		})
		t.Run("TLS Client Subordinate CA", func(t *testing.T) {
			tlsClientSubordinateCA1CertTemplate, err := CertificateTemplateCA(tlsClientRootCACert.SubjectName, tlsClientSubordinateCA1Cert.SubjectName, tlsClientSubordinateCA1Cert.Duration, tlsClientSubordinateCA1Cert.MaxPathLen)
			verifyCertificateTemplate(t, err, tlsClientSubordinateCA1CertTemplate)
			tlsClientSubordinateCA1Cert.KeyMaterial.Cert, tlsClientSubordinateCA1Cert.KeyMaterial.DER, tlsClientSubordinateCA1Cert.KeyMaterial.PEM, err = SignCertificate(tlsClientRootCACert.KeyMaterial.Cert, tlsClientRootCACert.KeyMaterial.KeyPair.Private, tlsClientSubordinateCA1CertTemplate, tlsClientSubordinateCA1Cert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsClientSubordinateCA1Cert.KeyMaterial.Cert, tlsClientSubordinateCA1Cert.KeyMaterial.DER, tlsClientSubordinateCA1Cert.KeyMaterial.PEM, tlsClientRootCACert.SubjectName, tlsClientSubordinateCA1Cert.SubjectName, tlsClientSubordinateCA1Cert.Duration, tlsClientSubordinateCA1Cert.MaxPathLen)
			verifyCertChain(t, tlsClientSubordinateCA1Cert.KeyMaterial.Cert, tlsClientRootCAsPool, tlsClientSubordinateCAsPool)
			tlsClientSubordinateCAsPool.AddCert(tlsClientSubordinateCA1Cert.KeyMaterial.Cert) // subsequent verify cert chain needs the intermediate CA
		})
		t.Run("TLS Client Subordinate CA", func(t *testing.T) {
			tlsClientSubordinateCA2CertTemplate, err := CertificateTemplateCA(tlsClientSubordinateCA1Cert.SubjectName, tlsClientSubordinateCA2Cert.SubjectName, tlsClientSubordinateCA2Cert.Duration, tlsClientSubordinateCA2Cert.MaxPathLen)
			verifyCertificateTemplate(t, err, tlsClientSubordinateCA2CertTemplate)
			tlsClientSubordinateCA2Cert.KeyMaterial.Cert, tlsClientSubordinateCA2Cert.KeyMaterial.DER, tlsClientSubordinateCA2Cert.KeyMaterial.PEM, err = SignCertificate(tlsClientSubordinateCA1Cert.KeyMaterial.Cert, tlsClientSubordinateCA1Cert.KeyMaterial.KeyPair.Private, tlsClientSubordinateCA2CertTemplate, tlsClientSubordinateCA2Cert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsClientSubordinateCA2Cert.KeyMaterial.Cert, tlsClientSubordinateCA2Cert.KeyMaterial.DER, tlsClientSubordinateCA2Cert.KeyMaterial.PEM, tlsClientSubordinateCA1Cert.SubjectName, tlsClientSubordinateCA2Cert.SubjectName, tlsClientSubordinateCA2Cert.Duration, tlsClientSubordinateCA2Cert.MaxPathLen)
			verifyCertChain(t, tlsClientSubordinateCA2Cert.KeyMaterial.Cert, tlsClientRootCAsPool, tlsClientSubordinateCAsPool)
			tlsClientSubordinateCAsPool.AddCert(tlsClientSubordinateCA2Cert.KeyMaterial.Cert) // subsequent verify cert chain needs the issuing CA
		})
		t.Run("TLS Client End Entity", func(t *testing.T) {
			tlsClientEndEntityCertTemplate, err := CertificateTemplateTLSClient(tlsClientSubordinateCA2Cert.SubjectName, tlsClientEndEntityCert.SubjectName, tlsClientEndEntityCert.Duration, tlsClientEndEntityCert.DNSNames, tlsClientEndEntityCert.IPAddresses, tlsClientEndEntityCert.EmailAddresses, tlsClientEndEntityCert.URIs)
			verifyCertificateTemplate(t, err, tlsClientEndEntityCertTemplate)
			tlsClientEndEntityCert.KeyMaterial.Cert, tlsClientEndEntityCert.KeyMaterial.DER, tlsClientEndEntityCert.KeyMaterial.PEM, err = SignCertificate(tlsClientSubordinateCA2Cert.KeyMaterial.Cert, tlsClientSubordinateCA2Cert.KeyMaterial.KeyPair.Private, tlsClientEndEntityCertTemplate, tlsClientEndEntityCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyEndEntityCertificate(t, err, tlsClientEndEntityCert.KeyMaterial.Cert, tlsClientEndEntityCert.KeyMaterial.DER, tlsClientEndEntityCert.KeyMaterial.PEM, tlsClientSubordinateCA2Cert.SubjectName, tlsClientEndEntityCert.SubjectName, tlsClientEndEntityCert.Duration, tlsClientEndEntityCert.DNSNames, tlsClientEndEntityCert.IPAddresses, tlsClientEndEntityCert.EmailAddresses, tlsClientEndEntityCert.URIs)
			verifyCertChain(t, tlsClientEndEntityCert.KeyMaterial.Cert, tlsClientRootCAsPool, tlsClientSubordinateCAsPool)
		})
	})

	// These TLS certificate chain instances are reusable for both the Raw mTLS and HTTP mTLS tests
	serverTLSCertChain := tls.Certificate{
		Certificate: [][]byte{tlsServerEndEntityCert.KeyMaterial.DER, tlsServerSubordinateCA2Cert.KeyMaterial.DER, tlsServerSubordinateCA1Cert.KeyMaterial.DER, tlsServerRootCACert.KeyMaterial.DER},
		PrivateKey:  tlsServerEndEntityCert.KeyMaterial.KeyPair.Private,
	}
	clientTLSCertChain := tls.Certificate{
		Certificate: [][]byte{tlsClientEndEntityCert.KeyMaterial.DER, tlsClientSubordinateCA2Cert.KeyMaterial.DER, tlsClientSubordinateCA1Cert.KeyMaterial.DER, tlsClientRootCACert.KeyMaterial.DER},
		PrivateKey:  tlsClientEndEntityCert.KeyMaterial.KeyPair.Private,
	}

	// These TLS configuration instances are reusable for both the Raw mTLS and HTTP mTLS tests
	serverTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSCertChain},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    tlsClientRootCAsPool,
	}
	clientTLSConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientTLSCertChain},
		RootCAs:            tlsServerRootCAsPool,
		InsecureSkipVerify: false,
	}

	t.Run("Raw mTLS", func(t *testing.T) {
		callerShutdownSignalCh := make(chan struct{})
		tlsListenerAddress, err := startTlsEchoServer("127.0.0.1:0", 100*time.Millisecond, serverTLSConfig, callerShutdownSignalCh) // or "0.0.0.0:0" for all interfaces
		require.NoError(t, err, "failed to start TLS Echo Server")
		defer close(callerShutdownSignalCh)
		const tlsClientConnections = 10
		tlsClientRequestBody := []byte("Hello Mutual TLS!")
		for i := 1; i <= tlsClientConnections; i++ {
			func() {
				tlsClientConnection, err := tls.Dial("tcp", tlsListenerAddress, clientTLSConfig)
				require.NoError(t, err, "client failed to connect to TLS Echo Server")
				defer tlsClientConnection.Close()

				_, err = tlsClientConnection.Write(tlsClientRequestBody)
				require.NoError(t, err, "client failed to write to TLS Echo Server (%d of %d)", i, tlsClientConnections)

				tlsServerResponseBody := make([]byte, len(tlsClientRequestBody))
				_, err = tlsClientConnection.Read(tlsServerResponseBody)
				require.NoError(t, err, "client failed to read from TLS Echo Server (%d of %d)", i, tlsClientConnections)
				require.Equal(t, tlsClientRequestBody, tlsServerResponseBody, "echo message mismatch (%d of %d)", i, tlsClientConnections)
			}()
		}
	})

	t.Run("HTTP mTLS", func(t *testing.T) {
		httpsServer, serverURL := startHTTPSEchoServer(serverTLSConfig, t)
		defer httpsServer.Close()
		httpsClientRequestBody := []byte("Hello Mutual HTTPS!")
		httpsClient := &http.Client{Transport: &http.Transport{TLSClientConfig: clientTLSConfig}}
		const httpsClientConnections = 10
		for i := 1; i <= httpsClientConnections; i++ {
			httpsServerResponse, err := httpsClient.Post(serverURL, "text/plain", bytes.NewReader(httpsClientRequestBody))
			require.NoError(t, err, "client failed to POST to HTTPS server (%d of %d)", i, httpsClientConnections)
			require.Equal(t, http.StatusOK, httpsServerResponse.StatusCode, "Unexpected HTTP status (%d of %d)", i, httpsClientConnections)
			func() {
				defer httpsServerResponse.Body.Close()
				httpServerResponseBody, err := io.ReadAll(httpsServerResponse.Body)
				require.NoError(t, err, "client failed to read response body (%d of %d)", i, httpsClientConnections)
				require.Equal(t, httpsClientRequestBody, httpServerResponseBody, "Echoed message mismatch (%d of %d)", i, httpsClientConnections)
			}()
		}
	})
}
