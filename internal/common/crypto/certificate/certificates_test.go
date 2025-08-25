package certificate

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	cryptoutilDateTime "cryptoutil/internal/common/util/datetime"

	"github.com/stretchr/testify/require"
)

func TestNegativeDuration(t *testing.T) {
	_, err := CertificateTemplateCA("Root CA", "Root CA", -1, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}

func TestMutualTLS(t *testing.T) {
	tlsServerCASubjects := make([]CASubject, 0, 4) // Root CA + Subordinate CA 1 (Intermediate) + Subordinate CA 2 (Intermediate) + Subordinate CA 3 (Issuing)
	tlsServerRootCACertsPool := x509.NewCertPool()
	tlsServerSubordinateCACertsPool := x509.NewCertPool()
	for i := range cap(tlsServerCASubjects) {
		currentTLSServerCASubject := CASubject{SubjectName: fmt.Sprintf("Test TLS Server CA %d", i), Duration: 10 * 365 * cryptoutilDateTime.Days1, MaxPathLen: cap(tlsServerCASubjects) - i - 1, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get(), CertChain: []*x509.Certificate{}, DERChain: [][]byte{}, PEMChain: [][]byte{}, RootCACertsPool: x509.NewCertPool(), SubordinateCACertsPool: x509.NewCertPool()}}
		previousTLSServerCASubject := currentTLSServerCASubject
		var previousTLSServerCACert *x509.Certificate
		if i > 0 {
			previousTLSServerCASubject = tlsServerCASubjects[i-1]
			previousTLSServerCACert = previousTLSServerCASubject.KeyMaterial.CertChain[0]
		}
		t.Run(currentTLSServerCASubject.SubjectName, func(t *testing.T) {
			currentCACertTemplate, err := CertificateTemplateCA(previousTLSServerCASubject.SubjectName, currentTLSServerCASubject.SubjectName, currentTLSServerCASubject.Duration, currentTLSServerCASubject.MaxPathLen)
			verifyCertificateTemplate(t, err, currentCACertTemplate)
			cert, der, pem, err := SignCertificate(previousTLSServerCACert, previousTLSServerCASubject.KeyMaterial.KeyPair.Private, currentCACertTemplate, currentTLSServerCASubject.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			currentTLSServerCASubject.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, previousTLSServerCASubject.KeyMaterial.CertChain...)
			currentTLSServerCASubject.KeyMaterial.DERChain = append([][]byte{der}, previousTLSServerCASubject.KeyMaterial.DERChain...)
			currentTLSServerCASubject.KeyMaterial.PEMChain = append([][]byte{pem}, previousTLSServerCASubject.KeyMaterial.PEMChain...)
			verifyCACertificate(t, err, currentTLSServerCASubject.KeyMaterial.CertChain, currentTLSServerCASubject.KeyMaterial.DERChain, currentTLSServerCASubject.KeyMaterial.PEMChain, previousTLSServerCASubject.SubjectName, currentTLSServerCASubject.SubjectName, currentTLSServerCASubject.Duration, currentCACertTemplate.MaxPathLen)
			if i == 0 {
				tlsServerRootCACertsPool.AddCert(cert)
				currentTLSServerCASubject.KeyMaterial.RootCACertsPool.AddCert(cert)
			} else {
				tlsServerSubordinateCACertsPool.AddCert(cert)
				currentTLSServerCASubject.KeyMaterial.SubordinateCACertsPool.AddCert(cert)
			}
		})
		tlsServerCASubjects = append(tlsServerCASubjects, currentTLSServerCASubject)
	}

	tlsServerEndEntityCert := EndEntitySubject{SubjectName: "Test TLS Server End Entity", Duration: 397 * cryptoutilDateTime.Days1, DNSNames: []string{"localhost", "tlsserver.example.com"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get(), CertChain: []*x509.Certificate{}, DERChain: [][]byte{}, PEMChain: [][]byte{}, RootCACertsPool: x509.NewCertPool(), SubordinateCACertsPool: x509.NewCertPool()}}
	t.Run("TLS Server", func(t *testing.T) {
		t.Run("TLS Server End Entity", func(t *testing.T) {
			tlsServerIssuingCA := tlsServerCASubjects[cap(tlsServerCASubjects)-1]
			tlsServerEndEntityCertTemplate, err := CertificateTemplateTLSServer(tlsServerIssuingCA.SubjectName, tlsServerEndEntityCert.SubjectName, tlsServerEndEntityCert.Duration, tlsServerEndEntityCert.DNSNames, tlsServerEndEntityCert.IPAddresses, tlsServerEndEntityCert.EmailAddresses, tlsServerEndEntityCert.URIs)
			verifyCertificateTemplate(t, err, tlsServerEndEntityCertTemplate)
			cert, der, pem, err := SignCertificate(tlsServerIssuingCA.KeyMaterial.CertChain[0], tlsServerIssuingCA.KeyMaterial.KeyPair.Private, tlsServerEndEntityCertTemplate, tlsServerEndEntityCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			tlsServerEndEntityCert.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, tlsServerIssuingCA.KeyMaterial.CertChain...)
			tlsServerEndEntityCert.KeyMaterial.DERChain = append([][]byte{der}, tlsServerIssuingCA.KeyMaterial.DERChain...)
			tlsServerEndEntityCert.KeyMaterial.PEMChain = append([][]byte{pem}, tlsServerIssuingCA.KeyMaterial.PEMChain...)
			verifyEndEntityCertificate(t, err, cert, der, pem, tlsServerIssuingCA.SubjectName, tlsServerEndEntityCert.SubjectName, tlsServerEndEntityCert.Duration, tlsServerEndEntityCert.DNSNames, tlsServerEndEntityCert.IPAddresses, tlsServerEndEntityCert.EmailAddresses, tlsServerEndEntityCert.URIs)
			verifyCertChain(t, cert, tlsServerRootCACertsPool, tlsServerSubordinateCACertsPool)
		})
	})
	serverTLSCertChain := tls.Certificate{Certificate: tlsServerEndEntityCert.KeyMaterial.DERChain, PrivateKey: tlsServerEndEntityCert.KeyMaterial.KeyPair.Private}

	tlsClientCASubjects := make([]CASubject, 0, 2) // Root CA + Subordinate CA 1 (Issuing)
	tlsClientRootCACertsPool := x509.NewCertPool()
	tlsClientSubordinateCACertsPool := x509.NewCertPool()
	var previousTLSClientCASubject CASubject
	var previousTLSClientCACert *x509.Certificate
	for i := range cap(tlsClientCASubjects) {
		currentTLSClientCASubject := CASubject{SubjectName: fmt.Sprintf("Test TLS Client CA %d", i), Duration: 10 * 365 * cryptoutilDateTime.Days1, MaxPathLen: cap(tlsClientCASubjects) - i - 1, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get(), CertChain: []*x509.Certificate{}, DERChain: [][]byte{}, PEMChain: [][]byte{}, RootCACertsPool: x509.NewCertPool(), SubordinateCACertsPool: x509.NewCertPool()}}
		if i == 0 {
			previousTLSClientCASubject = currentTLSClientCASubject
			previousTLSClientCACert = nil
		} else {
			previousTLSClientCASubject = tlsClientCASubjects[i-1]
			previousTLSClientCACert = tlsClientCASubjects[i-1].KeyMaterial.CertChain[0]
		}
		t.Run(currentTLSClientCASubject.SubjectName, func(t *testing.T) {
			currentCACertTemplate, err := CertificateTemplateCA(previousTLSClientCASubject.SubjectName, currentTLSClientCASubject.SubjectName, currentTLSClientCASubject.Duration, currentTLSClientCASubject.MaxPathLen)
			verifyCertificateTemplate(t, err, currentCACertTemplate)
			cert, der, pem, err := SignCertificate(previousTLSClientCACert, previousTLSClientCASubject.KeyMaterial.KeyPair.Private, currentCACertTemplate, currentTLSClientCASubject.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			currentTLSClientCASubject.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, previousTLSClientCASubject.KeyMaterial.CertChain...)
			currentTLSClientCASubject.KeyMaterial.DERChain = append([][]byte{der}, previousTLSClientCASubject.KeyMaterial.DERChain...)
			currentTLSClientCASubject.KeyMaterial.PEMChain = append([][]byte{pem}, previousTLSClientCASubject.KeyMaterial.PEMChain...)
			verifyCACertificate(t, err, currentTLSClientCASubject.KeyMaterial.CertChain, currentTLSClientCASubject.KeyMaterial.DERChain, currentTLSClientCASubject.KeyMaterial.PEMChain, previousTLSClientCASubject.SubjectName, currentTLSClientCASubject.SubjectName, currentTLSClientCASubject.Duration, currentCACertTemplate.MaxPathLen)
			if i == 0 {
				tlsClientRootCACertsPool.AddCert(cert)
				currentTLSClientCASubject.KeyMaterial.RootCACertsPool.AddCert(cert)
			} else {
				tlsClientSubordinateCACertsPool.AddCert(cert)
				currentTLSClientCASubject.KeyMaterial.SubordinateCACertsPool.AddCert(cert)
			}
		})
		tlsClientCASubjects = append(tlsClientCASubjects, currentTLSClientCASubject)
	}

	tlsClientEndEntityCert := EndEntitySubject{SubjectName: "Test TLS Client End Entity", Duration: 30 * cryptoutilDateTime.Days1, DNSNames: nil, IPAddresses: nil, EmailAddresses: []string{"client1@tlsclient.example.com"}, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get(), CertChain: []*x509.Certificate{}, DERChain: [][]byte{}, PEMChain: [][]byte{}, RootCACertsPool: x509.NewCertPool(), SubordinateCACertsPool: x509.NewCertPool()}}
	t.Run("TLS Client", func(t *testing.T) {
		t.Run("TLS Client End Entity", func(t *testing.T) {
			tlsClientIssuingCA := tlsClientCASubjects[cap(tlsClientCASubjects)-1]
			tlsClientEndEntityCertTemplate, err := CertificateTemplateTLSClient(tlsClientIssuingCA.SubjectName, tlsClientEndEntityCert.SubjectName, tlsClientEndEntityCert.Duration, tlsClientEndEntityCert.DNSNames, tlsClientEndEntityCert.IPAddresses, tlsClientEndEntityCert.EmailAddresses, tlsClientEndEntityCert.URIs)
			verifyCertificateTemplate(t, err, tlsClientEndEntityCertTemplate)
			cert, der, pem, err := SignCertificate(tlsClientIssuingCA.KeyMaterial.CertChain[0], tlsClientIssuingCA.KeyMaterial.KeyPair.Private, tlsClientEndEntityCertTemplate, tlsClientEndEntityCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			tlsClientEndEntityCert.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, tlsClientIssuingCA.KeyMaterial.CertChain...)
			tlsClientEndEntityCert.KeyMaterial.DERChain = append([][]byte{der}, tlsClientIssuingCA.KeyMaterial.DERChain...)
			tlsClientEndEntityCert.KeyMaterial.PEMChain = append([][]byte{pem}, tlsClientIssuingCA.KeyMaterial.PEMChain...)
			verifyEndEntityCertificate(t, err, cert, der, pem, tlsClientIssuingCA.SubjectName, tlsClientEndEntityCert.SubjectName, tlsClientEndEntityCert.Duration, tlsClientEndEntityCert.DNSNames, tlsClientEndEntityCert.IPAddresses, tlsClientEndEntityCert.EmailAddresses, tlsClientEndEntityCert.URIs)
			verifyCertChain(t, cert, tlsClientRootCACertsPool, tlsClientSubordinateCACertsPool)
		})
	})
	clientTLSCertChain := tls.Certificate{Certificate: tlsClientEndEntityCert.KeyMaterial.DERChain, PrivateKey: tlsClientEndEntityCert.KeyMaterial.KeyPair.Private}

	// The TLS certificate chain instances are reusable for both the Raw mTLS and HTTP mTLS tests
	// These TLS configuration instances are reusable for both the Raw mTLS and HTTP mTLS tests
	serverTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSCertChain},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    tlsClientRootCACertsPool,
	}
	clientTLSConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientTLSCertChain},
		InsecureSkipVerify: false,
		RootCAs:            tlsServerRootCACertsPool,
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
			require.NoError(t, err, "client failed to POST to HTTPS Echo Server (%d of %d)", i, httpsClientConnections)
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
