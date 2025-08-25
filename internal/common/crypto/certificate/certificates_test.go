package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
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

func TestMutualTLS(t *testing.T) {
	tlsServerCAs := make([]CASubject, 0, 4) // Root CA + Subordinate CA 1 (Intermediate) + Subordinate CA 2 (Intermediate) + Subordinate CA 3 (Issuing)
	tlsServerRootCAsPool := x509.NewCertPool()
	tlsServerSubordinateCAsPool := x509.NewCertPool()
	var previous CASubject
	for i := range cap(tlsServerCAs) {
		subjectName := fmt.Sprintf("Test TLS Server CA %d", i)
		current := CASubject{SubjectName: subjectName, Duration: 10 * 365 * cryptoutilDateTime.Days1, MaxPathLen: cap(tlsServerCAs) - i - 1, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
		if i == 0 {
			previous = current
		} else {
			previous = tlsServerCAs[i-1]
		}
		t.Run(subjectName, func(t *testing.T) {
			currentCACertTemplate, err := CertificateTemplateCA(previous.SubjectName, current.SubjectName, current.Duration, current.MaxPathLen)
			verifyCertificateTemplate(t, err, currentCACertTemplate)
			current.KeyMaterial.Cert, current.KeyMaterial.DER, current.KeyMaterial.PEM, err = SignCertificate(previous.KeyMaterial.Cert, previous.KeyMaterial.KeyPair.Private, currentCACertTemplate, current.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, current.KeyMaterial.Cert, current.KeyMaterial.DER, current.KeyMaterial.PEM, previous.SubjectName, current.SubjectName, current.Duration, currentCACertTemplate.MaxPathLen)
			if i == 0 {
				tlsServerRootCAsPool.AddCert(current.KeyMaterial.Cert)
			} else {
				tlsServerSubordinateCAsPool.AddCert(current.KeyMaterial.Cert)
			}
		})
		tlsServerCAs = append(tlsServerCAs, current)
	}

	tlsServerEndEntityCert := EndEntitySubject{SubjectName: "Test TLS Server End Entity", Duration: 397 * cryptoutilDateTime.Days1, DNSNames: []string{"localhost", "tlsserver.example.com"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	t.Run("TLS Server", func(t *testing.T) {
		t.Run("TLS Server End Entity", func(t *testing.T) {
			tlsServerIssuingCA := tlsServerCAs[cap(tlsServerCAs)-1]
			tlsServerEndEntityCertTemplate, err := CertificateTemplateTLSServer(tlsServerIssuingCA.SubjectName, tlsServerEndEntityCert.SubjectName, tlsServerEndEntityCert.Duration, tlsServerEndEntityCert.DNSNames, tlsServerEndEntityCert.IPAddresses, tlsServerEndEntityCert.EmailAddresses, tlsServerEndEntityCert.URIs)
			verifyCertificateTemplate(t, err, tlsServerEndEntityCertTemplate)
			tlsServerEndEntityCert.KeyMaterial.Cert, tlsServerEndEntityCert.KeyMaterial.DER, tlsServerEndEntityCert.KeyMaterial.PEM, err = SignCertificate(tlsServerIssuingCA.KeyMaterial.Cert, tlsServerIssuingCA.KeyMaterial.KeyPair.Private, tlsServerEndEntityCertTemplate, tlsServerEndEntityCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyEndEntityCertificate(t, err, tlsServerEndEntityCert.KeyMaterial.Cert, tlsServerEndEntityCert.KeyMaterial.DER, tlsServerEndEntityCert.KeyMaterial.PEM, tlsServerIssuingCA.SubjectName, tlsServerEndEntityCert.SubjectName, tlsServerEndEntityCert.Duration, tlsServerEndEntityCert.DNSNames, tlsServerEndEntityCert.IPAddresses, tlsServerEndEntityCert.EmailAddresses, tlsServerEndEntityCert.URIs)
			verifyCertChain(t, tlsServerEndEntityCert.KeyMaterial.Cert, tlsServerRootCAsPool, tlsServerSubordinateCAsPool)
		})
	})
	tlsServerCerts := make([][]byte, 0, cap(tlsServerCAs)+1)
	tlsServerCerts = append(tlsServerCerts, tlsServerEndEntityCert.KeyMaterial.DER)
	for i := cap(tlsServerCAs) - 1; i >= 0; i-- {
		tlsServerCerts = append(tlsServerCerts, tlsServerCAs[i].KeyMaterial.DER)
	}
	serverTLSCertChain := tls.Certificate{Certificate: tlsServerCerts, PrivateKey: tlsServerEndEntityCert.KeyMaterial.KeyPair.Private}

	tlsClientRootCACert := CASubject{SubjectName: "Test TLS Client Root CA", Duration: 10 * cryptoutilDateTime.Days365, MaxPathLen: 2, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientSubordinateCA1Cert := CASubject{SubjectName: "Test TLS Client Subordinate CA 1", Duration: 5 * cryptoutilDateTime.Days365, MaxPathLen: 1, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientSubordinateCA2Cert := CASubject{SubjectName: "Test TLS Client Subordinate CA 2", Duration: 2 * cryptoutilDateTime.Days365, MaxPathLen: 0, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientRootCAsPool := x509.NewCertPool()
	tlsClientSubordinateCAsPool := x509.NewCertPool()
	t.Run("TLS Client Chain", func(t *testing.T) {
		t.Run("TLS Client Root CA", func(t *testing.T) {
			tlsClientRootCACertTemplate, err := CertificateTemplateCA(tlsClientRootCACert.SubjectName, tlsClientRootCACert.SubjectName, tlsClientRootCACert.Duration, tlsClientRootCACert.MaxPathLen)
			verifyCertificateTemplate(t, err, tlsClientRootCACertTemplate)
			tlsClientRootCACert.KeyMaterial.Cert, tlsClientRootCACert.KeyMaterial.DER, tlsClientRootCACert.KeyMaterial.PEM, err = SignCertificate(nil, tlsClientRootCACert.KeyMaterial.KeyPair.Private, tlsClientRootCACertTemplate, tlsClientRootCACert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsClientRootCACert.KeyMaterial.Cert, tlsClientRootCACert.KeyMaterial.DER, tlsClientRootCACert.KeyMaterial.PEM, tlsClientRootCACert.SubjectName, tlsClientRootCACert.SubjectName, tlsClientRootCACert.Duration, tlsClientRootCACert.MaxPathLen)
			tlsClientRootCAsPool.AddCert(tlsClientRootCACert.KeyMaterial.Cert) // subsequent verify cert chain needs the root CA
		})
		t.Run("TLS Client Subordinate CA 1", func(t *testing.T) {
			tlsClientSubordinateCA1CertTemplate, err := CertificateTemplateCA(tlsClientRootCACert.SubjectName, tlsClientSubordinateCA1Cert.SubjectName, tlsClientSubordinateCA1Cert.Duration, tlsClientSubordinateCA1Cert.MaxPathLen)
			verifyCertificateTemplate(t, err, tlsClientSubordinateCA1CertTemplate)
			tlsClientSubordinateCA1Cert.KeyMaterial.Cert, tlsClientSubordinateCA1Cert.KeyMaterial.DER, tlsClientSubordinateCA1Cert.KeyMaterial.PEM, err = SignCertificate(tlsClientRootCACert.KeyMaterial.Cert, tlsClientRootCACert.KeyMaterial.KeyPair.Private, tlsClientSubordinateCA1CertTemplate, tlsClientSubordinateCA1Cert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsClientSubordinateCA1Cert.KeyMaterial.Cert, tlsClientSubordinateCA1Cert.KeyMaterial.DER, tlsClientSubordinateCA1Cert.KeyMaterial.PEM, tlsClientRootCACert.SubjectName, tlsClientSubordinateCA1Cert.SubjectName, tlsClientSubordinateCA1Cert.Duration, tlsClientSubordinateCA1Cert.MaxPathLen)
			verifyCertChain(t, tlsClientSubordinateCA1Cert.KeyMaterial.Cert, tlsClientRootCAsPool, tlsClientSubordinateCAsPool)
			tlsClientSubordinateCAsPool.AddCert(tlsClientSubordinateCA1Cert.KeyMaterial.Cert) // subsequent verify cert chain needs the intermediate CA
		})
		t.Run("TLS Client Subordinate CA 2", func(t *testing.T) {
			tlsClientSubordinateCA2CertTemplate, err := CertificateTemplateCA(tlsClientSubordinateCA1Cert.SubjectName, tlsClientSubordinateCA2Cert.SubjectName, tlsClientSubordinateCA2Cert.Duration, tlsClientSubordinateCA2Cert.MaxPathLen)
			verifyCertificateTemplate(t, err, tlsClientSubordinateCA2CertTemplate)
			tlsClientSubordinateCA2Cert.KeyMaterial.Cert, tlsClientSubordinateCA2Cert.KeyMaterial.DER, tlsClientSubordinateCA2Cert.KeyMaterial.PEM, err = SignCertificate(tlsClientSubordinateCA1Cert.KeyMaterial.Cert, tlsClientSubordinateCA1Cert.KeyMaterial.KeyPair.Private, tlsClientSubordinateCA2CertTemplate, tlsClientSubordinateCA2Cert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsClientSubordinateCA2Cert.KeyMaterial.Cert, tlsClientSubordinateCA2Cert.KeyMaterial.DER, tlsClientSubordinateCA2Cert.KeyMaterial.PEM, tlsClientSubordinateCA1Cert.SubjectName, tlsClientSubordinateCA2Cert.SubjectName, tlsClientSubordinateCA2Cert.Duration, tlsClientSubordinateCA2Cert.MaxPathLen)
			verifyCertChain(t, tlsClientSubordinateCA2Cert.KeyMaterial.Cert, tlsClientRootCAsPool, tlsClientSubordinateCAsPool)
			tlsClientSubordinateCAsPool.AddCert(tlsClientSubordinateCA2Cert.KeyMaterial.Cert) // subsequent verify cert chain needs the issuing CA
		})
	})

	tlsClientEndEntityCert := EndEntitySubject{SubjectName: "Test TLS Client End Entity", Duration: 30 * cryptoutilDateTime.Days1, DNSNames: []string{"localhost", "tlsclient.example.com"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, EmailAddresses: []string{"client1@tlsclient.example.com"}, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	t.Run("TLS Client", func(t *testing.T) {
		t.Run("TLS Client End Entity", func(t *testing.T) {
			tlsClientEndEntityCertTemplate, err := CertificateTemplateTLSClient(tlsClientSubordinateCA2Cert.SubjectName, tlsClientEndEntityCert.SubjectName, tlsClientEndEntityCert.Duration, tlsClientEndEntityCert.DNSNames, tlsClientEndEntityCert.IPAddresses, tlsClientEndEntityCert.EmailAddresses, tlsClientEndEntityCert.URIs)
			verifyCertificateTemplate(t, err, tlsClientEndEntityCertTemplate)
			tlsClientEndEntityCert.KeyMaterial.Cert, tlsClientEndEntityCert.KeyMaterial.DER, tlsClientEndEntityCert.KeyMaterial.PEM, err = SignCertificate(tlsClientSubordinateCA2Cert.KeyMaterial.Cert, tlsClientSubordinateCA2Cert.KeyMaterial.KeyPair.Private, tlsClientEndEntityCertTemplate, tlsClientEndEntityCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyEndEntityCertificate(t, err, tlsClientEndEntityCert.KeyMaterial.Cert, tlsClientEndEntityCert.KeyMaterial.DER, tlsClientEndEntityCert.KeyMaterial.PEM, tlsClientSubordinateCA2Cert.SubjectName, tlsClientEndEntityCert.SubjectName, tlsClientEndEntityCert.Duration, tlsClientEndEntityCert.DNSNames, tlsClientEndEntityCert.IPAddresses, tlsClientEndEntityCert.EmailAddresses, tlsClientEndEntityCert.URIs)
			verifyCertChain(t, tlsClientEndEntityCert.KeyMaterial.Cert, tlsClientRootCAsPool, tlsClientSubordinateCAsPool)
		})
	})

	// These TLS certificate chain instances are reusable for both the Raw mTLS and HTTP mTLS tests

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
