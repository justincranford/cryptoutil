package certificate

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
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
	tlsServerCASubjects, err := CreateCASubjects(testKeyGenPool, "Test TLS Server CA", 3) // Root CA + 2 Intermediate CAs
	require.NoError(t, err, "Failed to create TLS Server CA subjects")
	tlsClientCASubjects, err := CreateCASubjects(testKeyGenPool, "Test TLS Client CA", 2) // Root CA + 1 Intermediate CA
	require.NoError(t, err, "Failed to create TLS Client CA subjects")
	tlsServerEndEntitySubject, err := CreateEndEntitySubject(testKeyGenPool, "Test TLS Server End Entity", 397*cryptoutilDateTime.Days1, []string{"localhost", "tlsserver.example.com"}, []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, nil, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, tlsServerCASubjects)
	require.NoError(t, err, "Failed to create TLS Server End Entity subject")
	tlsClientEndEntitySubject, err := CreateEndEntitySubject(testKeyGenPool, "Test TLS Client End Entity", 30*cryptoutilDateTime.Days1, nil, nil, []string{"client1@tlsclient.example.com"}, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, tlsClientCASubjects)
	require.NoError(t, err, "Failed to create TLS Client End Entity subject")

	// The TLS certificate chain instances are reusable for both the Raw mTLS and HTTP mTLS tests
	tlsServerCertChain, tlsServerRootCAs, err := BuildTLSCertificate(tlsServerEndEntitySubject)
	require.NoError(t, err, "Failed to build TLS server certificate")
	tlsClientCertChain, tlsClientRootCAs, err := BuildTLSCertificate(tlsClientEndEntitySubject)
	require.NoError(t, err, "Failed to build TLS client certificate")

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

func TestSerializeSubjects(t *testing.T) {
	subjects, err := CreateCASubjects(testKeyGenPool, "Test Serialize CA", 2)
	require.NoError(t, err, "Failed to create CA subjects")

	serializedSubjects, err := SerializeSubjects(subjects, true)
	require.NoError(t, err, "Failed to serialize subjects")
	require.NotEmpty(t, serializedSubjects, "Serialized data should not be empty")
	t.Logf("Serialized Subjects count: %d", len(serializedSubjects))
	for i, subject := range serializedSubjects {
		t.Logf("Serialized Subject %d: %s", i, string(subject))
	}

	deserializedKeyMaterials, err := DeserializeSubjects(serializedSubjects)
	require.NoError(t, err, "Failed to deserialize subjects")
	require.Len(t, deserializedKeyMaterials, len(subjects), "Deserialized KeyMaterialDecoded count should match original")

	for i, original := range subjects {
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

func TestSerializeSubjectsIncludePrivateKey(t *testing.T) {
	subjects, err := CreateCASubjects(testKeyGenPool, "Test PrivateKey CA", 2)
	require.NoError(t, err, "Failed to create CA subjects")

	// Test with includePrivateKey = true
	serializedWithPrivateKey, err := SerializeSubjects(subjects, true)
	require.NoError(t, err, "Failed to serialize subjects with private key")
	require.NotEmpty(t, serializedWithPrivateKey, "Serialized data with private key should not be empty")

	// Test with includePrivateKey = false
	serializedWithoutPrivateKey, err := SerializeSubjects(subjects, false)
	require.NoError(t, err, "Failed to serialize subjects without private key")
	require.NotEmpty(t, serializedWithoutPrivateKey, "Serialized data without private key should not be empty")

	// Verify that the serializations are different
	require.NotEqual(t, serializedWithPrivateKey, serializedWithoutPrivateKey, "Serializations with and without private key should be different")

	// Verify that the serialization without private key doesn't contain private key data
	for i, serialized := range serializedWithoutPrivateKey {
		require.Contains(t, string(serialized), "\"der_private_key\":null", "Serialization without private key should have der_private_key as null for subject %d", i)
		require.Contains(t, string(serialized), "\"pem_private_key\":null", "Serialization without private key should have pem_private_key as null for subject %d", i)
	}

	// Verify that the serialization with private key does contain actual private key data
	for i, serialized := range serializedWithPrivateKey {
		require.NotContains(t, string(serialized), "\"der_private_key\":null", "Serialization with private key should not have der_private_key as null for subject %d", i)
		require.NotContains(t, string(serialized), "\"pem_private_key\":null", "Serialization with private key should not have pem_private_key as null for subject %d", i)
		require.Contains(t, string(serialized), "der_private_key", "Serialization with private key should contain der_private_key field for subject %d", i)
		require.Contains(t, string(serialized), "pem_private_key", "Serialization with private key should contain pem_private_key field for subject %d", i)
	}
}

func TestSerializeKeyMaterial(t *testing.T) {
	// Create a KeyMaterialDecoded with test data
	keyPair := testKeyGenPool.Get()
	keyMaterial := KeyMaterialDecoded{
		PrivateKey: keyPair.Private,
		PublicKey:  keyPair.Public,
		CertChain:  []*x509.Certificate{}, // Empty for this test to avoid certificate parsing issues
	}

	// Test serialization with private key
	serializedWithPrivateKey, err := SerializeKeyMaterial(&keyMaterial, true)
	require.NoError(t, err, "Failed to serialize KeyMaterialDecoded with private key")
	require.NotEmpty(t, serializedWithPrivateKey, "Serialized data should not be empty")

	// Test serialization without private key
	serializedWithoutPrivateKey, err := SerializeKeyMaterial(&keyMaterial, false)
	require.NoError(t, err, "Failed to serialize KeyMaterialDecoded without private key")
	require.NotEmpty(t, serializedWithoutPrivateKey, "Serialized data should not be empty")

	// Verify that serialization without private key excludes private key data
	require.NotEqual(t, serializedWithPrivateKey, serializedWithoutPrivateKey, "Serializations with and without private key should be different")

	// Test deserialization
	deserializedKeyMaterial, err := DeserializeKeyMaterial(serializedWithPrivateKey)
	require.NoError(t, err, "Failed to deserialize KeyMaterialDecoded")

	// Verify the reconstructed data matches the original
	require.Equal(t, len(keyMaterial.CertChain), len(deserializedKeyMaterial.CertChain), "CertChain length should match")
	require.NotNil(t, deserializedKeyMaterial.PrivateKey, "PrivateKey should be reconstructed")
	require.NotNil(t, deserializedKeyMaterial.PublicKey, "PublicKey should be reconstructed")
}

func TestNewFieldsPopulated(t *testing.T) {
	// Test CA subjects have proper fields populated
	subjects, err := CreateCASubjects(testKeyGenPool, "Test Fields CA", 2)
	require.NoError(t, err, "Failed to create CA subjects")

	// Root CA (index 0)
	rootCA := subjects[0]
	require.Equal(t, "Test Fields CA 0", rootCA.SubjectName, "Root CA subject name should match")
	require.Equal(t, "Test Fields CA 0", rootCA.IssuerName, "Root CA issuer name should be self-signed")
	require.NotNil(t, rootCA.CASubject, "Root CA should have CASubject populated")
	require.True(t, rootCA.CASubject.IsCA, "Root CA should have IsCA=true")
	require.Equal(t, 1, rootCA.CASubject.MaxPathLen, "Root CA should have expected MaxPathLen")

	// Intermediate CA (index 1)
	intermediateCA := subjects[1]
	require.Equal(t, "Test Fields CA 1", intermediateCA.SubjectName, "Intermediate CA subject name should match")
	require.Equal(t, "Test Fields CA 0", intermediateCA.IssuerName, "Intermediate CA should be issued by root CA")
	require.NotNil(t, intermediateCA.CASubject, "Intermediate CA should have CASubject populated")
	require.True(t, intermediateCA.CASubject.IsCA, "Intermediate CA should have IsCA=true")
	require.Equal(t, 0, intermediateCA.CASubject.MaxPathLen, "Intermediate CA should have expected MaxPathLen")

	// Test End Entity subjects have proper fields populated
	endEntitySubject, err := CreateEndEntitySubject(testKeyGenPool, "Test Fields End Entity", 30*cryptoutilDateTime.Days1,
		[]string{"test.example.com"}, []net.IP{net.ParseIP("127.0.0.1")}, nil, nil,
		x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, subjects)
	require.NoError(t, err, "Failed to create End Entity subject")

	require.Equal(t, "Test Fields End Entity", endEntitySubject.SubjectName, "End entity subject name should match")
	require.Equal(t, "Test Fields CA 1", endEntitySubject.IssuerName, "End entity should be issued by leaf CA")
	require.NotNil(t, endEntitySubject.EndEntitySubject, "End entity should have EndEntitySubject populated")
	require.Equal(t, []string{"test.example.com"}, endEntitySubject.EndEntitySubject.DNSNames, "End entity DNS names should match")
}
