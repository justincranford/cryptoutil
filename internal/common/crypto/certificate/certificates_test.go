package certificate

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
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
	tlsServerCASubjects := createCASubjects(t, "Test TLS Server CA", 3) // Root CA + 2 Intermediate CAs
	tlsClientCASubjects := createCASubjects(t, "Test TLS Client CA", 2) // Root CA + 1 Intermediate CA
	tlsServerEndEntitySubject := createEndEntitySubject(t, "Test TLS Server End Entity", 397*cryptoutilDateTime.Days1, []string{"localhost", "tlsserver.example.com"}, []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, nil, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, tlsServerCASubjects)
	tlsClientEndEntitySubject := createEndEntitySubject(t, "Test TLS Client End Entity", 30*cryptoutilDateTime.Days1, nil, nil, []string{"client1@tlsclient.example.com"}, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, tlsClientCASubjects)

	// The TLS certificate chain instances are reusable for both the Raw mTLS and HTTP mTLS tests
	tlsServerCertChain, tlsServerRootCAs := buildTLSCertificate(tlsServerEndEntitySubject)
	tlsClientCertChain, tlsClientRootCAs := buildTLSCertificate(tlsClientEndEntitySubject)

	// These TLS configuration instances are reusable for both the Raw mTLS and HTTP mTLS tests
	serverTLSConfig := &tls.Config{Certificates: []tls.Certificate{tlsServerCertChain}, ClientCAs: tlsClientRootCAs, ClientAuth: tls.RequireAndVerifyClientCert}
	clientTLSConfig := &tls.Config{Certificates: []tls.Certificate{tlsClientCertChain}, RootCAs: tlsServerRootCAs, InsecureSkipVerify: false}

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

func TestSerializeCASubjects(t *testing.T) {
	caSubjects := createCASubjects(t, "Test Serialize CA", 2)

	serializedData, err := SerializeCASubjects(caSubjects)
	require.NoError(t, err, "Failed to serialize CA subjects")
	require.NotEmpty(t, serializedData, "Serialized data should not be empty")

	deserializedSubjects, err := DeserializeCASubjects(serializedData)
	require.NoError(t, err, "Failed to deserialize CA subjects")
	require.Len(t, deserializedSubjects, len(caSubjects), "Deserialized subjects count should match original")

	for i, original := range caSubjects {
		deserialized := deserializedSubjects[i]
		require.Equal(t, original.SubjectName, deserialized.SubjectName, "Subject name mismatch at index %d", i)
		require.Equal(t, original.Duration, deserialized.Duration, "Duration mismatch at index %d", i)
		require.Equal(t, original.MaxPathLen, deserialized.MaxPathLen, "MaxPathLen mismatch at index %d", i)
		require.Equal(t, original.KeyMaterial.DERChain, deserialized.DERChain, "DERChain mismatch at index %d", i)
		require.Equal(t, original.KeyMaterial.PEMChain, deserialized.PEMChain, "PEMChain mismatch at index %d", i)
	}
}

func createCASubjects(t *testing.T, caSubjectNamePrefix string, numCAs int) []CASubject {
	caSubjects := make([]CASubject, 0, numCAs)
	for i := range cap(caSubjects) {
		currentCASubject := CASubject{SubjectName: fmt.Sprintf("%s %d", caSubjectNamePrefix, i), Duration: 10 * 365 * cryptoutilDateTime.Days1, MaxPathLen: cap(caSubjects) - i - 1, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get(), CertChain: []*x509.Certificate{}, DERChain: [][]byte{}, PEMChain: [][]byte{}, RootCACertsPool: x509.NewCertPool(), SubordinateCACertsPool: x509.NewCertPool()}}
		previousCASubject := currentCASubject
		var previousCACert *x509.Certificate
		if i > 0 {
			previousCASubject = caSubjects[i-1]
			previousCACert = previousCASubject.KeyMaterial.CertChain[0]
		}
		t.Run(currentCASubject.SubjectName, func(t *testing.T) {
			currentCACertTemplate, err := CertificateTemplateCA(previousCASubject.SubjectName, currentCASubject.SubjectName, currentCASubject.Duration, currentCASubject.MaxPathLen)
			verifyCertificateTemplate(t, err, currentCACertTemplate)
			cert, der, pem, err := SignCertificate(previousCACert, previousCASubject.KeyMaterial.KeyPair.Private, currentCACertTemplate, currentCASubject.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			currentCASubject.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, previousCASubject.KeyMaterial.CertChain...)
			currentCASubject.KeyMaterial.DERChain = append([][]byte{der}, previousCASubject.KeyMaterial.DERChain...)
			currentCASubject.KeyMaterial.PEMChain = append([][]byte{pem}, previousCASubject.KeyMaterial.PEMChain...)
			verifyCACertificate(t, err, currentCASubject.KeyMaterial.CertChain, currentCASubject.KeyMaterial.DERChain, currentCASubject.KeyMaterial.PEMChain, previousCASubject.SubjectName, currentCASubject.SubjectName, currentCASubject.Duration, currentCACertTemplate.MaxPathLen)
			currentCASubject.KeyMaterial.RootCACertsPool = previousCASubject.KeyMaterial.RootCACertsPool.Clone()
			currentCASubject.KeyMaterial.SubordinateCACertsPool = previousCASubject.KeyMaterial.SubordinateCACertsPool.Clone()
			if i == 0 {
				currentCASubject.KeyMaterial.RootCACertsPool.AddCert(cert)
			} else {
				currentCASubject.KeyMaterial.SubordinateCACertsPool.AddCert(cert)
			}
		})
		caSubjects = append(caSubjects, currentCASubject)
	}
	return caSubjects
}

func createEndEntitySubject(t *testing.T, subjectName string, duration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage, caSubjects []CASubject) EndEntitySubject {
	endEntityCert := EndEntitySubject{SubjectName: subjectName, Duration: duration, DNSNames: dnsNames, IPAddresses: ipAddresses, EmailAddresses: emailAddresses, URIs: uris, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get(), CertChain: []*x509.Certificate{}, DERChain: [][]byte{}, PEMChain: [][]byte{}, RootCACertsPool: x509.NewCertPool(), SubordinateCACertsPool: x509.NewCertPool()}}
	t.Run(subjectName, func(t *testing.T) {
		issuingCA := caSubjects[cap(caSubjects)-1]
		endEntityCertTemplate, err := CertificateTemplateEndEntity(issuingCA.SubjectName, endEntityCert.SubjectName, endEntityCert.Duration, endEntityCert.DNSNames, endEntityCert.IPAddresses, endEntityCert.EmailAddresses, endEntityCert.URIs, keyUsage, extKeyUsage)
		verifyCertificateTemplate(t, err, endEntityCertTemplate)
		cert, der, pem, err := SignCertificate(issuingCA.KeyMaterial.CertChain[0], issuingCA.KeyMaterial.KeyPair.Private, endEntityCertTemplate, endEntityCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
		endEntityCert.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, issuingCA.KeyMaterial.CertChain...)
		endEntityCert.KeyMaterial.DERChain = append([][]byte{der}, issuingCA.KeyMaterial.DERChain...)
		endEntityCert.KeyMaterial.PEMChain = append([][]byte{pem}, issuingCA.KeyMaterial.PEMChain...)
		verifyEndEntityCertificate(t, err, cert, der, pem, issuingCA.SubjectName, endEntityCert.SubjectName, endEntityCert.Duration, endEntityCert.DNSNames, endEntityCert.IPAddresses, endEntityCert.EmailAddresses, endEntityCert.URIs)
		verifyCertChain(t, cert, issuingCA.KeyMaterial.RootCACertsPool, issuingCA.KeyMaterial.SubordinateCACertsPool)
		endEntityCert.KeyMaterial.RootCACertsPool = issuingCA.KeyMaterial.RootCACertsPool.Clone()
		endEntityCert.KeyMaterial.SubordinateCACertsPool = issuingCA.KeyMaterial.SubordinateCACertsPool.Clone()
	})
	return endEntityCert
}

func buildTLSCertificate(endEntitySubject EndEntitySubject) (tls.Certificate, *x509.CertPool) {
	return tls.Certificate{Certificate: endEntitySubject.KeyMaterial.DERChain, PrivateKey: endEntitySubject.KeyMaterial.KeyPair.Private, Leaf: endEntitySubject.KeyMaterial.CertChain[0]}, endEntitySubject.KeyMaterial.RootCACertsPool
}
