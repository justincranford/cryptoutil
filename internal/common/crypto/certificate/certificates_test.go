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

	serializedCASubjects, err := SerializeCASubjects(caSubjects)
	require.NoError(t, err, "Failed to serialize CA subjects")
	require.NotEmpty(t, serializedCASubjects, "Serialized data should not be empty")
	t.Logf("Serialized CA Subjects count: %d", len(serializedCASubjects))
	for i, caSubject := range serializedCASubjects {
		t.Logf("Serialized CA Subject %d: %s", i, string(caSubject))
	}

	deserializedKeyMaterials, err := DeserializeCASubjects(serializedCASubjects)
	require.NoError(t, err, "Failed to deserialize CA subjects")
	require.Len(t, deserializedKeyMaterials, len(caSubjects), "Deserialized KeyMaterials count should match original")

	for i, original := range caSubjects {
		deserializedKM := deserializedKeyMaterials[i]
		originalKM := original.KeyMaterial

		// Compare private keys
		require.Equal(t, originalKM.PrivateKey, deserializedKM.PrivateKey, "PrivateKey mismatch at index %d", i)

		// Compare public keys
		require.Equal(t, originalKM.PublicKey, deserializedKM.PublicKey, "PublicKey mismatch at index %d", i)

		// Compare certificate chains
		require.Len(t, deserializedKM.CertChain, len(originalKM.CertChain), "CertChain length mismatch at index %d", i)
		for j, originalCert := range originalKM.CertChain {
			deserializedCert := deserializedKM.CertChain[j]
			require.Equal(t, originalCert.Raw, deserializedCert.Raw, "Certificate Raw data mismatch at index %d, cert %d", i, j)
		}
	}
}

func TestSerializeKeyMaterial(t *testing.T) {
	// Create a KeyMaterial with test data
	keyPair := testKeyGenPool.Get()
	keyMaterial := KeyMaterial{
		PrivateKey: keyPair.Private,
		PublicKey:  keyPair.Public,
		CertChain:  []*x509.Certificate{}, // Empty for this test to avoid certificate parsing issues
	}

	// Test serialization with private key
	serializedWithPrivateKey, err := SerializeKeyMaterial(&keyMaterial, true)
	require.NoError(t, err, "Failed to serialize KeyMaterial with private key")
	require.NotEmpty(t, serializedWithPrivateKey, "Serialized data should not be empty")

	// Test serialization without private key
	serializedWithoutPrivateKey, err := SerializeKeyMaterial(&keyMaterial, false)
	require.NoError(t, err, "Failed to serialize KeyMaterial without private key")
	require.NotEmpty(t, serializedWithoutPrivateKey, "Serialized data should not be empty")

	// Verify that serialization without private key excludes private key data
	require.NotEqual(t, serializedWithPrivateKey, serializedWithoutPrivateKey, "Serializations with and without private key should be different")

	// Test deserialization
	deserializedKeyMaterial, err := DeserializeKeyMaterial(serializedWithPrivateKey)
	require.NoError(t, err, "Failed to deserialize KeyMaterial")

	// Verify the reconstructed data matches the original
	require.Equal(t, len(keyMaterial.CertChain), len(deserializedKeyMaterial.CertChain), "CertChain length should match")
	require.NotNil(t, deserializedKeyMaterial.PrivateKey, "PrivateKey should be reconstructed")
	require.NotNil(t, deserializedKeyMaterial.PublicKey, "PublicKey should be reconstructed")
}

func createCASubjects(t *testing.T, caSubjectNamePrefix string, numCAs int) []CASubject {
	require.Greater(t, numCAs, 0, "numCAs must be greater than 0")
	caSubjects := make([]CASubject, numCAs)
	for i := 0; i < numCAs; i++ {
		keyPair := testKeyGenPool.Get()
		require.NotNil(t, keyPair, "keyPair should not be nil for CA %d", i)
		require.NotNil(t, keyPair.Private, "keyPair.Private should not be nil for CA %d", i)
		require.NotNil(t, keyPair.Public, "keyPair.Public should not be nil for CA %d", i)

		// Determine issuer name - root CA issues itself, others are issued by previous CA
		issuerName := fmt.Sprintf("%s %d", caSubjectNamePrefix, i) // Self-signed for root CA
		if i > 0 {
			issuerName = fmt.Sprintf("%s %d", caSubjectNamePrefix, i-1) // Previous CA for intermediate CAs
		}

		currentCASubject := CASubject{
			Subject: Subject{
				SubjectName: fmt.Sprintf("%s %d", caSubjectNamePrefix, i),
				IssuerName:  issuerName,
				Duration:    10 * 365 * cryptoutilDateTime.Days1,
				KeyMaterial: KeyMaterial{
					PrivateKey:             keyPair.Private,
					PublicKey:              keyPair.Public,
					CertChain:              []*x509.Certificate{},
					RootCACertsPool:        x509.NewCertPool(),
					SubordinateCACertsPool: x509.NewCertPool(),
				},
			},
			MaxPathLen: numCAs - i - 1,
			IsCA:       true,
		}
		previousCASubject := currentCASubject
		var previousCACert *x509.Certificate
		if i > 0 {
			previousCASubject = caSubjects[i-1]
			previousCACert = previousCASubject.KeyMaterial.CertChain[0]
		}
		t.Run(currentCASubject.SubjectName, func(t *testing.T) {
			currentCACertTemplate, err := CertificateTemplateCA(previousCASubject.SubjectName, currentCASubject.SubjectName, currentCASubject.Duration, currentCASubject.MaxPathLen)
			verifyCertificateTemplate(t, err, currentCACertTemplate)
			cert, _, pemBytes, err := SignCertificate(previousCACert, previousCASubject.KeyMaterial.PrivateKey, currentCACertTemplate, currentCASubject.KeyMaterial.PublicKey, x509.ECDSAWithSHA256)
			currentCASubject.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, previousCASubject.KeyMaterial.CertChain...)

			// Create DER and PEM chains locally for verification
			derChain := make([][]byte, len(currentCASubject.KeyMaterial.CertChain))
			pemChain := make([][]byte, len(currentCASubject.KeyMaterial.CertChain))
			for j, c := range currentCASubject.KeyMaterial.CertChain {
				derChain[j] = c.Raw
				pemChain[j] = pemBytes // Use the pemBytes from SignCertificate for the first cert
			}
			verifyCACertificate(t, err, currentCASubject.KeyMaterial.CertChain, derChain, pemChain, previousCASubject.SubjectName, currentCASubject.SubjectName, currentCASubject.Duration, currentCACertTemplate.MaxPathLen)

			currentCASubject.KeyMaterial.RootCACertsPool = previousCASubject.KeyMaterial.RootCACertsPool.Clone()
			currentCASubject.KeyMaterial.SubordinateCACertsPool = previousCASubject.KeyMaterial.SubordinateCACertsPool.Clone()
			if i == 0 {
				currentCASubject.KeyMaterial.RootCACertsPool.AddCert(cert)
			} else {
				currentCASubject.KeyMaterial.SubordinateCACertsPool.AddCert(cert)
			}
		})
		caSubjects[i] = currentCASubject
	}
	return caSubjects
}

func createEndEntitySubject(t *testing.T, subjectName string, duration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage, caSubjects []CASubject) EndEntitySubject {
	keyPair := testKeyGenPool.Get()
	require.NotNil(t, keyPair, "keyPair should not be nil")
	require.NotNil(t, keyPair.Private, "keyPair.Private should not be nil")
	require.NotNil(t, keyPair.Public, "keyPair.Public should not be nil")

	// The issuing CA is the last one in the chain (leaf CA)
	require.NotEmpty(t, caSubjects, "caSubjects should not be empty")
	issuingCA := caSubjects[len(caSubjects)-1]
	require.NotEmpty(t, issuingCA.SubjectName, "issuingCA.SubjectName should not be empty")

	endEntityCert := EndEntitySubject{
		Subject: Subject{
			SubjectName: subjectName,
			IssuerName:  issuingCA.SubjectName,
			Duration:    duration,
			KeyMaterial: KeyMaterial{
				PrivateKey:             keyPair.Private,
				PublicKey:              keyPair.Public,
				CertChain:              []*x509.Certificate{},
				RootCACertsPool:        x509.NewCertPool(),
				SubordinateCACertsPool: x509.NewCertPool(),
			},
		},
		DNSNames:       dnsNames,
		IPAddresses:    ipAddresses,
		EmailAddresses: emailAddresses,
		URIs:           uris,
	}
	t.Run(subjectName, func(t *testing.T) {
		endEntityCertTemplate, err := CertificateTemplateEndEntity(issuingCA.SubjectName, endEntityCert.SubjectName, endEntityCert.Duration, endEntityCert.DNSNames, endEntityCert.IPAddresses, endEntityCert.EmailAddresses, endEntityCert.URIs, keyUsage, extKeyUsage)
		verifyCertificateTemplate(t, err, endEntityCertTemplate)
		cert, derBytes, pemBytes, err := SignCertificate(issuingCA.KeyMaterial.CertChain[0], issuingCA.KeyMaterial.PrivateKey, endEntityCertTemplate, endEntityCert.KeyMaterial.PublicKey, x509.ECDSAWithSHA256)
		endEntityCert.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, issuingCA.KeyMaterial.CertChain...)
		verifyEndEntityCertificate(t, err, cert, derBytes, pemBytes, issuingCA.SubjectName, endEntityCert.SubjectName, endEntityCert.Duration, endEntityCert.DNSNames, endEntityCert.IPAddresses, endEntityCert.EmailAddresses, endEntityCert.URIs)
		verifyCertChain(t, cert, issuingCA.KeyMaterial.RootCACertsPool, issuingCA.KeyMaterial.SubordinateCACertsPool)
		endEntityCert.KeyMaterial.RootCACertsPool = issuingCA.KeyMaterial.RootCACertsPool.Clone()
		endEntityCert.KeyMaterial.SubordinateCACertsPool = issuingCA.KeyMaterial.SubordinateCACertsPool.Clone()
	})
	return endEntityCert
}

func buildTLSCertificate(endEntitySubject EndEntitySubject) (tls.Certificate, *x509.CertPool) {
	// Convert certificate chain to DER format for TLS
	derCertChain := make([][]byte, len(endEntitySubject.KeyMaterial.CertChain))
	for i, cert := range endEntitySubject.KeyMaterial.CertChain {
		derCertChain[i] = cert.Raw
	}

	return tls.Certificate{Certificate: derCertChain, PrivateKey: endEntitySubject.KeyMaterial.PrivateKey, Leaf: endEntitySubject.KeyMaterial.CertChain[0]}, endEntitySubject.KeyMaterial.RootCACertsPool
}

func TestNewFieldsPopulated(t *testing.T) {
	// Test CA subjects have proper fields populated
	caSubjects := createCASubjects(t, "Test Fields CA", 2)

	// Root CA (index 0)
	rootCA := caSubjects[0]
	require.Equal(t, "Test Fields CA 0", rootCA.SubjectName, "Root CA subject name should match")
	require.Equal(t, "Test Fields CA 0", rootCA.IssuerName, "Root CA issuer name should be self-signed")
	require.True(t, rootCA.IsCA, "Root CA should have IsCA=true")
	require.Equal(t, 1, rootCA.MaxPathLen, "Root CA should have expected MaxPathLen")

	// Intermediate CA (index 1)
	intermediateCA := caSubjects[1]
	require.Equal(t, "Test Fields CA 1", intermediateCA.SubjectName, "Intermediate CA subject name should match")
	require.Equal(t, "Test Fields CA 0", intermediateCA.IssuerName, "Intermediate CA should be issued by root CA")
	require.True(t, intermediateCA.IsCA, "Intermediate CA should have IsCA=true")
	require.Equal(t, 0, intermediateCA.MaxPathLen, "Intermediate CA should have expected MaxPathLen")

	// Test End Entity subjects have proper fields populated
	endEntitySubject := createEndEntitySubject(t, "Test Fields End Entity", 30*cryptoutilDateTime.Days1,
		[]string{"test.example.com"}, []net.IP{net.ParseIP("127.0.0.1")}, nil, nil,
		x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, caSubjects)

	require.Equal(t, "Test Fields End Entity", endEntitySubject.SubjectName, "End entity subject name should match")
	require.Equal(t, "Test Fields CA 1", endEntitySubject.IssuerName, "End entity should be issued by leaf CA")
}
