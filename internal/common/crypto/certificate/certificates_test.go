package certificate

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	cryptoutilDateTime "cryptoutil/internal/common/util/datetime"

	"github.com/stretchr/testify/require"
)

func TestMutualTLS(t *testing.T) {
	tlsServerSubjectsKeyPairs, err := getKeyPairs(4, testKeyGenPool) // End Entity + 2 Intermediate CAs + Root CA
	require.NoError(t, err, "Failed to get key pairs for CA subjects")
	tlsClientSubjectsKeyPairs, err := getKeyPairs(3, testKeyGenPool)
	require.NoError(t, err, "Failed to get key pairs for CA subjects") // End Entity + 1 Intermediate CA + Root CA

	tlsServerCASubjects, err := CreateCASubjects(tlsServerSubjectsKeyPairs[1:], "Test TLS Server CA")
	verifyCASubjects(t, err, tlsServerCASubjects)
	tlsClientCASubjects, err := CreateCASubjects(tlsClientSubjectsKeyPairs[1:], "Test TLS Client CA")
	verifyCASubjects(t, err, tlsClientCASubjects)

	tlsServerEndEntitySubject, err := CreateEndEntitySubject(tlsServerCASubjects[0], tlsServerSubjectsKeyPairs[0], "Test TLS Server End Entity", 397*cryptoutilDateTime.Days1, []string{"localhost", "tlsserver.example.com"}, []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, nil, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
	verifyEndEntitySubject(t, err, tlsServerEndEntitySubject)
	tlsClientEndEntitySubject, err := CreateEndEntitySubject(tlsClientCASubjects[0], tlsClientSubjectsKeyPairs[0], "Test TLS Client End Entity", 30*cryptoutilDateTime.Days1, nil, nil, []string{"client1@tlsclient.example.com"}, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
	verifyEndEntitySubject(t, err, tlsClientEndEntitySubject)

	tlsServerCertChain, tlsServerRootCAs, _, err := BuildTLSCertificate(tlsServerEndEntitySubject)
	require.NoError(t, err, "Failed to build TLS server certificate")
	tlsClientCertChain, tlsClientRootCAs, _, err := BuildTLSCertificate(tlsClientEndEntitySubject)
	require.NoError(t, err, "Failed to build TLS client certificate")

	// TLS configuration instances are reusable for both of the Raw mTLS and HTTP mTLS tests
	serverTLSConfig := &tls.Config{Certificates: []tls.Certificate{tlsServerCertChain}, ClientCAs: tlsClientRootCAs, ClientAuth: tls.RequireAndVerifyClientCert}
	clientTLSConfig := &tls.Config{Certificates: []tls.Certificate{tlsClientCertChain}, RootCAs: tlsServerRootCAs, InsecureSkipVerify: false}

	const clientConnections = 10
	t.Run("Raw mTLS", func(t *testing.T) {
		callerShutdownSignalCh := make(chan struct{})
		tlsListenerAddress, err := startTlsEchoServer("127.0.0.1:0", 100*time.Millisecond, 100*time.Millisecond, serverTLSConfig, callerShutdownSignalCh) // or "0.0.0.0:0" for all interfaces
		require.NoError(t, err, "failed to start TLS Echo Server")
		defer close(callerShutdownSignalCh)
		tlsClientRequestBody := []byte("Hello Mutual TLS!")
		for i := 1; i <= clientConnections; i++ {
			func() {
				tlsClientConnection, err := tls.Dial("tcp", tlsListenerAddress, clientTLSConfig)
				require.NoError(t, err, "client failed to connect to TLS Echo Server")
				defer tlsClientConnection.Close()

				_, err = tlsClientConnection.Write(tlsClientRequestBody)
				require.NoError(t, err, "client failed to write to TLS Echo Server (%d of %d)", i, clientConnections)

				tlsServerResponseBody := make([]byte, len(tlsClientRequestBody))
				_, err = tlsClientConnection.Read(tlsServerResponseBody)
				require.NoError(t, err, "client failed to read from TLS Echo Server (%d of %d)", i, clientConnections)
				require.Equal(t, tlsClientRequestBody, tlsServerResponseBody, "echo message mismatch (%d of %d)", i, clientConnections)
			}()
		}
	})

	t.Run("HTTP mTLS", func(t *testing.T) {
		httpsServer, serverURL, err := startHTTPSEchoServer("127.0.0.1:0", 100*time.Millisecond, 100*time.Millisecond, serverTLSConfig) // or "0.0.0.0:0" for all interfaces
		require.NoError(t, err, "failed to start HTTPS Echo Server")
		defer httpsServer.Close()
		httpsClientRequestBody := []byte("Hello Mutual HTTPS!")
		httpsClient := &http.Client{Transport: &http.Transport{TLSClientConfig: clientTLSConfig}}
		for i := 1; i <= clientConnections; i++ {
			httpsServerResponse, err := httpsClient.Post(serverURL, "text/plain", bytes.NewReader(httpsClientRequestBody))
			require.NoError(t, err, "client failed to POST to HTTPS Echo Server (%d of %d)", i, clientConnections)
			require.Equal(t, http.StatusOK, httpsServerResponse.StatusCode, "Unexpected HTTP status (%d of %d)", i, clientConnections)
			func() {
				defer httpsServerResponse.Body.Close()
				httpServerResponseBody, err := io.ReadAll(httpsServerResponse.Body)
				require.NoError(t, err, "client failed to read response body (%d of %d)", i, clientConnections)
				require.Equal(t, httpsClientRequestBody, httpServerResponseBody, "Echoed message mismatch (%d of %d)", i, clientConnections)
			}()
		}
	})
}

// testSubjectSerializationRoundTrip is a helper function to test subject serialization/deserialization
// with or without private keys, eliminating code duplication between test cases
func testSubjectSerializationRoundTrip(t *testing.T, originalSubjects []*Subject, includePrivateKey bool) {
	// Serialize subjects
	serializedSubjects, err := SerializeSubjects(originalSubjects, includePrivateKey)
	require.NoError(t, err, "Failed to serialize subjects (includePrivateKey=%t)", includePrivateKey)
	require.NotEmpty(t, serializedSubjects, "Serialized data should not be empty (includePrivateKey=%t)", includePrivateKey)

	// Verify serialized JSON content
	for i, serializedSubject := range serializedSubjects {
		require.Contains(t, string(serializedSubject), "der_private_key", "Serialization should contain der_private_key field for subject %d (includePrivateKey=%t)", i, includePrivateKey)
		require.Contains(t, string(serializedSubject), "pem_private_key", "Serialization should contain pem_private_key field for subject %d (includePrivateKey=%t)", i, includePrivateKey)

		if includePrivateKey {
			// With private key: first subject (leaf CA at index 0) should have private key, others should be null
			if i == 0 {
				require.NotContains(t, string(serializedSubject), "\"der_private_key\":null", "Serialization with private key should not have der_private_key as null for subject %d", i)
				require.NotContains(t, string(serializedSubject), "\"pem_private_key\":null", "Serialization with private key should not have pem_private_key as null for subject %d", i)
			} else {
				require.Contains(t, string(serializedSubject), "\"der_private_key\":null", "Serialization with private key should have der_private_key as null for subject %d", i)
				require.Contains(t, string(serializedSubject), "\"pem_private_key\":null", "Serialization with private key should have pem_private_key as null for subject %d", i)
			}
		} else {
			// Without private key: all subjects should have null private keys
			require.Contains(t, string(serializedSubject), "\"der_private_key\":null", "Serialization without private key should have der_private_key as null for subject %d", i)
			require.Contains(t, string(serializedSubject), "\"pem_private_key\":null", "Serialization without private key should have pem_private_key as null for subject %d", i)
		}
	}

	// Deserialize subjects
	deserializedSubjects, err := DeserializeSubjects(serializedSubjects)
	require.NoError(t, err, "Failed to deserialize subjects (includePrivateKey=%t)", includePrivateKey)
	require.Len(t, deserializedSubjects, len(originalSubjects), "Deserialized Subject count should match original (includePrivateKey=%t)", includePrivateKey)

	// Verify round-trip correctness
	for i, originalSubject := range originalSubjects {
		deserializedSubject := deserializedSubjects[i]
		originalKeyMaterial := originalSubject.KeyMaterial
		deserializedKeyMaterial := deserializedSubject.KeyMaterial

		// Verify private key handling based on includePrivateKey flag
		if includePrivateKey {
			require.Equal(t, originalKeyMaterial.PrivateKey, deserializedKeyMaterial.PrivateKey, "PrivateKey mismatch at index %d", i)
		} else {
			require.Nil(t, deserializedKeyMaterial.PrivateKey, "PrivateKey should be nil at index %d when includePrivateKey=false", i)
		}

		// Verify common fields (always preserved regardless of includePrivateKey)
		require.Equal(t, originalKeyMaterial.PublicKey, deserializedKeyMaterial.PublicKey, "PublicKey mismatch at index %d (includePrivateKey=%t)", i, includePrivateKey)
		require.Len(t, deserializedKeyMaterial.CertChain, len(originalKeyMaterial.CertChain), "CertChain length mismatch at index %d (includePrivateKey=%t)", i, includePrivateKey)
		for j, originalCert := range originalKeyMaterial.CertChain {
			deserializedCert := deserializedKeyMaterial.CertChain[j]
			require.Equal(t, originalCert.Raw, deserializedCert.Raw, "Certificate Raw data mismatch at index %d, cert %d (includePrivateKey=%t)", i, j, includePrivateKey)
		}
		require.Equal(t, originalSubject.SubjectName, deserializedSubject.SubjectName, "SubjectName mismatch at index %d (includePrivateKey=%t)", i, includePrivateKey)
		require.Equal(t, originalSubject.IssuerName, deserializedSubject.IssuerName, "IssuerName mismatch at index %d (includePrivateKey=%t)", i, includePrivateKey)
		require.Equal(t, originalSubject.IsCA, deserializedSubject.IsCA, "IsCA mismatch at index %d (includePrivateKey=%t)", i, includePrivateKey)
		if originalSubject.IsCA {
			require.Equal(t, originalSubject.MaxPathLen, deserializedSubject.MaxPathLen, "MaxPathLen mismatch at index %d (includePrivateKey=%t)", i, includePrivateKey)
		} else {
			require.Equal(t, originalSubject.DNSNames, deserializedSubject.DNSNames, "DNSNames mismatch at index %d (includePrivateKey=%t)", i, includePrivateKey)
			require.Equal(t, originalSubject.IPAddresses, deserializedSubject.IPAddresses, "IPAddresses mismatch at index %d (includePrivateKey=%t)", i, includePrivateKey)
			require.Equal(t, originalSubject.EmailAddresses, deserializedSubject.EmailAddresses, "EmailAddresses mismatch at index %d (includePrivateKey=%t)", i, includePrivateKey)
			require.Equal(t, originalSubject.URIs, deserializedSubject.URIs, "URIs mismatch at index %d (includePrivateKey=%t)", i, includePrivateKey)
		}
	}
}

func TestSerializeSubjectsWithAndWithoutPrivateKey(t *testing.T) {
	subjectsKeyPairs, err := getKeyPairs(2, testKeyGenPool)
	require.NoError(t, err, "Failed to get key pairs for CA subjects")

	originalSubjects, err := CreateCASubjects(subjectsKeyPairs, "Test PrivateKey CA")
	verifyCASubjects(t, err, originalSubjects)

	t.Run("includePrivateKey = false", func(t *testing.T) {
		testSubjectSerializationRoundTrip(t, originalSubjects, false)
	})

	t.Run("includePrivateKey = true", func(t *testing.T) {
		testSubjectSerializationRoundTrip(t, originalSubjects, true)
	})
}

func TestCompleteSubjectRoundTripSerialization(t *testing.T) {
	subjectsKeyPairs, err := getKeyPairs(2, testKeyGenPool)
	require.NoError(t, err, "Failed to get key pairs for CA subjects")

	originalCASubjects, err := CreateCASubjects(subjectsKeyPairs[1:], "Round Trip CA")
	verifyCASubjects(t, err, originalCASubjects)

	endEntitySubject, err := CreateEndEntitySubject(originalCASubjects[0], subjectsKeyPairs[0], "Round Trip End Entity", 365*cryptoutilDateTime.Days1, []string{"example.com"}, []net.IP{net.ParseIP("127.0.0.1")}, []string{"test@example.com"}, []*url.URL{{Scheme: "https", Host: "example.com"}}, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
	verifyEndEntitySubject(t, err, endEntitySubject)

	originalSubjects := append([]*Subject{endEntitySubject}, originalCASubjects...)

	serialized, err := SerializeSubjects(originalSubjects, true)
	require.NoError(t, err, "Failed to serialize subjects")

	deserializedSubjects, err := DeserializeSubjects(serialized)
	require.NoError(t, err, "Failed to deserialize subjects")
	require.Len(t, deserializedSubjects, len(originalSubjects), "Deserialized count should match original")

	// Verify full round-trip: []*Subject -> [][]byte -> []*Subject
	for i, originalSubject := range originalSubjects {
		deserializedSubject := deserializedSubjects[i]

		require.Equal(t, originalSubject.SubjectName, deserializedSubject.SubjectName, "SubjectName mismatch at index %d", i)
		require.Equal(t, originalSubject.IssuerName, deserializedSubject.IssuerName, "IssuerName mismatch at index %d", i)

		require.NotNil(t, deserializedSubject.KeyMaterial.PrivateKey, "PrivateKey should not be nil at index %d", i)
		require.NotNil(t, deserializedSubject.KeyMaterial.PublicKey, "PublicKey should not be nil at index %d", i)
		require.NotEmpty(t, deserializedSubject.KeyMaterial.CertChain, "CertChain should not be empty at index %d", i)

		if len(originalSubject.KeyMaterial.CertChain) > 1 {
			tlsCert, rootCACertsPool, intermediateCertsPool, err := BuildTLSCertificate(deserializedSubject)
			require.NoError(t, err, "BuildTLSCertificate should not fail at index %d", i)
			require.NotNil(t, rootCACertsPool, "Root CA cert pool should be reconstructed at index %d", i)
			require.NotEmpty(t, tlsCert.Certificate, "TLS certificate should have cert chain at index %d", i)

			verifyOptions := x509.VerifyOptions{
				Roots:         rootCACertsPool,
				Intermediates: intermediateCertsPool,
				KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
			}
			_, err = deserializedSubject.KeyMaterial.CertChain[0].Verify(verifyOptions)
			require.NoError(t, err, "Certificate chain verification failed for subject %d (%s)", i, originalSubject.SubjectName)
		}

		require.Equal(t, originalSubject.IsCA, deserializedSubject.IsCA, "IsCA mismatch at index %d", i)
		if originalSubject.IsCA {
			require.Equal(t, originalSubject.MaxPathLen, deserializedSubject.MaxPathLen, "MaxPathLen mismatch at index %d", i)
		} else {
			require.Equal(t, originalSubject.DNSNames, deserializedSubject.DNSNames, "DNSNames mismatch at index %d", i)

			require.Len(t, deserializedSubject.IPAddresses, len(originalSubject.IPAddresses), "IPAddresses length mismatch at index %d", i)
			for j, originalIP := range originalSubject.IPAddresses {
				deserializedIPAddress := deserializedSubject.IPAddresses[j]
				require.True(t, originalIP.Equal(deserializedIPAddress), "IPAddresses[%d] mismatch at index %d: expected %v, got %v", j, i, originalIP, deserializedIPAddress)
			}

			require.Equal(t, originalSubject.EmailAddresses, deserializedSubject.EmailAddresses, "EmailAddresses mismatch at index %d", i)
			require.Equal(t, originalSubject.URIs, deserializedSubject.URIs, "URIs mismatch at index %d", i)
		}

		t.Logf("Subject %d (%s) successfully round-tripped", i, originalSubject.SubjectName)
	}

	t.Logf("Successfully round-tripped %d subjects (including %d CAs and %d end entities)",
		len(originalSubjects), len(originalSubjects)-1, 1)
}

func TestSerializeSubjectsSadPaths(t *testing.T) {
	t.Run("nil subjects slice", func(t *testing.T) {
		_, err := SerializeSubjects(nil, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "subjects cannot be nil")
	})
	t.Run("empty SubjectName", func(t *testing.T) {
		subjects := []*Subject{{
			SubjectName: "", // Empty subject name should cause error
			IssuerName:  "Test Issuer",
		}}
		_, err := SerializeSubjects(subjects, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "has empty SubjectName")
	})
	t.Run("end-entity with CA MaxPathLen", func(t *testing.T) {
		subjects := []*Subject{{
			SubjectName: "Test Subject",
			IssuerName:  "Test Issuer",
			Duration:    time.Hour,
			IsCA:        false,
			MaxPathLen:  1, // Invalid for end entity
			KeyMaterial: KeyMaterial{
				CertChain: []*x509.Certificate{{}}, // Mock cert
				PublicKey: []byte("mock-key"),
			},
		}}
		_, err := SerializeSubjects(subjects, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is not a CA but has MaxPathLen populated")
	})
	t.Run("CA with end-entity fields", func(t *testing.T) {
		subjects := []*Subject{{
			SubjectName: "Test Subject",
			IssuerName:  "Test Issuer",
			Duration:    time.Hour,
			IsCA:        true,
			MaxPathLen:  0,
			DNSNames:    []string{"example.com"}, // Invalid for CA
		}}
		_, err := SerializeSubjects(subjects, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is a CA but has end-entity fields")
	})
	t.Run("invalid CA MaxPathLen", func(t *testing.T) {
		subjects := []*Subject{{
			SubjectName: "Test CA",
			IssuerName:  "Test Issuer",
			Duration:    time.Hour,
			IsCA:        true,
			MaxPathLen:  -1, // Invalid negative MaxPathLen
		}}
		_, err := SerializeSubjects(subjects, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "has invalid MaxPathLen (-1), must be >= 0")
	})
}

func TestSerializeKeyMaterialSadPaths(t *testing.T) {
	t.Run("nil keyMaterial", func(t *testing.T) {
		_, err := serializeKeyMaterial(nil, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "keyMaterial cannot be nil")
	})
	t.Run("nil PublicKey", func(t *testing.T) {
		keyPair := testKeyGenPool.Get()

		certTemplate, err := CertificateTemplateCA("Test Issuer", "Test CA", 10*365*cryptoutilDateTime.Days1, 0)
		verifyCertificateTemplate(t, err, certTemplate)

		cert, _, _, err := SignCertificate(nil, keyPair.Private, certTemplate, keyPair.Public, x509.ECDSAWithSHA256)
		require.NoError(t, err)

		keyMaterial := &KeyMaterial{
			CertChain: []*x509.Certificate{cert},
			PublicKey: nil, // Should cause error
		}
		_, err = serializeKeyMaterial(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "PublicKey cannot be nil")
	})
	t.Run("empty cert chain", func(t *testing.T) {
		keyPair := testKeyGenPool.Get()
		keyMaterial := &KeyMaterial{
			CertChain: []*x509.Certificate{}, // Empty chain should cause error
			PublicKey: keyPair.Public,
		}
		_, err := serializeKeyMaterial(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate chain cannot be empty")
	})
	t.Run("nil cert in chain", func(t *testing.T) {
		keyPair := testKeyGenPool.Get()
		keyMaterial := &KeyMaterial{
			CertChain: []*x509.Certificate{nil}, // Nil cert should cause error
			PublicKey: keyPair.Public,
		}
		_, err := serializeKeyMaterial(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate at index 0 in chain cannot be nil")
	})
}

func TestNegativeDuration(t *testing.T) {
	_, err := CertificateTemplateCA("Root CA", "Root CA", -1, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}
