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
	tlsServerSubjectsKeyPairs, err := getKeyPairs(4, testKeyGenPool)
	require.NoError(t, err, "Failed to get key pairs for CA subjects")
	tlsClientSubjectsKeyPairs, err := getKeyPairs(3, testKeyGenPool)
	require.NoError(t, err, "Failed to get key pairs for CA subjects")

	tlsServerCASubjects, err := CreateCASubjects(tlsServerSubjectsKeyPairs[1:], "Test TLS Server CA") // Root CA + 2 Intermediate CAs
	verifyCASubjects(t, err, tlsServerCASubjects)
	tlsClientCASubjects, err := CreateCASubjects(tlsClientSubjectsKeyPairs[1:], "Test TLS Client CA") // Root CA + 1 Intermediate CA
	verifyCASubjects(t, err, tlsClientCASubjects)

	tlsServerEndEntitySubject, err := CreateEndEntitySubject(tlsServerSubjectsKeyPairs[0], "Test TLS Server End Entity", 397*cryptoutilDateTime.Days1, []string{"localhost", "tlsserver.example.com"}, []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, nil, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, tlsServerCASubjects)
	verifyEndEntitySubject(t, err, tlsServerEndEntitySubject)
	tlsClientEndEntitySubject, err := CreateEndEntitySubject(tlsClientSubjectsKeyPairs[0], "Test TLS Client End Entity", 30*cryptoutilDateTime.Days1, nil, nil, []string{"client1@tlsclient.example.com"}, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, tlsClientCASubjects)
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

func TestSerializeSubjectsWithAndWithoutPrivateKey(t *testing.T) {
	subjectsKeyPairs, err := getKeyPairs(2, testKeyGenPool)
	require.NoError(t, err, "Failed to get key pairs for CA subjects")

	originalSubjects, err := CreateCASubjects(subjectsKeyPairs, "Test PrivateKey CA")
	verifyCASubjects(t, err, originalSubjects)

	t.Run("includePrivateKey = false", func(t *testing.T) {
		// Test with
		serializedSubjectsWithoutPrivateKey, err := SerializeSubjects(originalSubjects, false)
		require.NoError(t, err, "Failed to serialize subjects without private key")
		require.NotEmpty(t, serializedSubjectsWithoutPrivateKey, "Serialized data without private key should not be empty")

		// Verify that the serialization without private key doesn't contain private key data
		for i, originalSubjectWithoutPrivateKey := range serializedSubjectsWithoutPrivateKey {
			require.Contains(t, string(originalSubjectWithoutPrivateKey), "\"der_private_key\":null", "Serialization without private key should have der_private_key as null for subject %d", i)
			require.Contains(t, string(originalSubjectWithoutPrivateKey), "\"pem_private_key\":null", "Serialization without private key should have pem_private_key as null for subject %d", i)
		}

		// Test deserialization of subjects without private key
		deserializedSubjectsWithoutPrivateKey, err := DeserializeSubjects(serializedSubjectsWithoutPrivateKey)
		require.NoError(t, err, "Failed to deserialize subjects without private key")
		require.Len(t, deserializedSubjectsWithoutPrivateKey, len(originalSubjects), "Deserialized Subject count should match original")

		// Verify round-trip for subjects without private key (private key should be nil)
		for i, originalSubject := range originalSubjects {
			deserializedSubject := deserializedSubjectsWithoutPrivateKey[i]
			originalKeyMaterial := originalSubject.KeyMaterial
			deserializedKeyMaterial := deserializedSubject.KeyMaterial

			// Private key should be nil for deserialized without private key
			require.Nil(t, deserializedKeyMaterial.PrivateKey, "PrivateKey should be nil at index %d", i)
			require.Equal(t, originalKeyMaterial.PublicKey, deserializedKeyMaterial.PublicKey, "PublicKey mismatch at index %d", i)
			require.Len(t, deserializedKeyMaterial.CertChain, len(originalKeyMaterial.CertChain), "CertChain length mismatch at index %d", i)
			for j, originalCert := range originalKeyMaterial.CertChain {
				deserializedCert := deserializedKeyMaterial.CertChain[j]
				require.Equal(t, originalCert.Raw, deserializedCert.Raw, "Certificate Raw data mismatch at index %d, cert %d", i, j)
			}
			require.Equal(t, originalSubject.SubjectName, deserializedSubject.SubjectName, "SubjectName mismatch at index %d", i)
			require.Equal(t, originalSubject.IssuerName, deserializedSubject.IssuerName, "IssuerName mismatch at index %d", i)
			require.Equal(t, originalSubject.IsCA, deserializedSubject.IsCA, "IsCA mismatch at index %d", i)
			if originalSubject.IsCA {
				require.Equal(t, originalSubject.MaxPathLen, deserializedSubject.MaxPathLen, "MaxPathLen mismatch at index %d", i)
			} else {
				require.Equal(t, originalSubject.DNSNames, deserializedSubject.DNSNames, "DNSNames mismatch at index %d", i)
				require.Equal(t, originalSubject.IPAddresses, deserializedSubject.IPAddresses, "IPAddresses mismatch at index %d", i)
				require.Equal(t, originalSubject.EmailAddresses, deserializedSubject.EmailAddresses, "EmailAddresses mismatch at index %d", i)
				require.Equal(t, originalSubject.URIs, deserializedSubject.URIs, "URIs mismatch at index %d", i)
			}
		}
	})

	t.Run("includePrivateKey = true", func(t *testing.T) {
		serializedSubjectsWithPrivateKey, err := SerializeSubjects(originalSubjects, true)
		require.NoError(t, err, "Failed to serialize subjects with private key")
		require.NotEmpty(t, serializedSubjectsWithPrivateKey, "Serialized data with private key should not be empty")

		// Verify that the serialization with private key does contain actual private key data
		for i, serializedSubjectWithPrivateKey := range serializedSubjectsWithPrivateKey {
			require.NotContains(t, string(serializedSubjectWithPrivateKey), "\"der_private_key\":null", "Serialization with private key should not have der_private_key as null for subject %d", i)
			require.NotContains(t, string(serializedSubjectWithPrivateKey), "\"pem_private_key\":null", "Serialization with private key should not have pem_private_key as null for subject %d", i)
			require.Contains(t, string(serializedSubjectWithPrivateKey), "der_private_key", "Serialization with private key should contain der_private_key field for subject %d", i)
			require.Contains(t, string(serializedSubjectWithPrivateKey), "pem_private_key", "Serialization with private key should contain pem_private_key field for subject %d", i)
		}

		// Test deserialization of subjects with private key
		deserializedSubjectsWithPrivateKey, err := DeserializeSubjects(serializedSubjectsWithPrivateKey)
		require.NoError(t, err, "Failed to deserialize subjects with private key")
		require.Len(t, deserializedSubjectsWithPrivateKey, len(originalSubjects), "Deserialized Subject count should match original")

		// Verify full round-trip for subjects with private key (comprehensive validation from TestSerializeSubjects)
		for i, originalSubject := range originalSubjects {
			deserializedSubject := deserializedSubjectsWithPrivateKey[i]
			originalKeyMaterial := originalSubject.KeyMaterial
			deserializedKeyMaterial := deserializedSubject.KeyMaterial

			require.Equal(t, originalKeyMaterial.PrivateKey, deserializedKeyMaterial.PrivateKey, "PrivateKey mismatch at index %d", i)
			require.Equal(t, originalKeyMaterial.PublicKey, deserializedKeyMaterial.PublicKey, "PublicKey mismatch at index %d", i)
			require.Len(t, deserializedKeyMaterial.CertChain, len(originalKeyMaterial.CertChain), "CertChain length mismatch at index %d", i)
			for j, originalCert := range originalKeyMaterial.CertChain {
				deserializedCert := deserializedKeyMaterial.CertChain[j]
				require.Equal(t, originalCert.Raw, deserializedCert.Raw, "Certificate Raw data mismatch at index %d, cert %d", i, j)
			}
			require.Equal(t, originalSubject.SubjectName, deserializedSubject.SubjectName, "SubjectName mismatch at index %d", i)
			require.Equal(t, originalSubject.IssuerName, deserializedSubject.IssuerName, "IssuerName mismatch at index %d", i)
			require.Equal(t, originalSubject.IsCA, deserializedSubject.IsCA, "IsCA mismatch at index %d", i)
			if originalSubject.IsCA {
				require.Equal(t, originalSubject.MaxPathLen, deserializedSubject.MaxPathLen, "MaxPathLen mismatch at index %d", i)
			} else {
				require.Equal(t, originalSubject.DNSNames, deserializedSubject.DNSNames, "DNSNames mismatch at index %d", i)
				require.Equal(t, originalSubject.IPAddresses, deserializedSubject.IPAddresses, "IPAddresses mismatch at index %d", i)
				require.Equal(t, originalSubject.EmailAddresses, deserializedSubject.EmailAddresses, "EmailAddresses mismatch at index %d", i)
				require.Equal(t, originalSubject.URIs, deserializedSubject.URIs, "URIs mismatch at index %d", i)
			}
		}
	})
}

func TestCompleteSubjectRoundTripSerialization(t *testing.T) {
	subjectsKeyPairs, err := getKeyPairs(2, testKeyGenPool)
	require.NoError(t, err, "Failed to get key pairs for CA subjects")

	// Create test data
	originalSubjects, err := CreateCASubjects(subjectsKeyPairs[1:], "Round Trip CA")
	verifyCASubjects(t, err, originalSubjects)

	// Add an end entity subject
	endEntitySubject, err := CreateEndEntitySubject(subjectsKeyPairs[0], "Round Trip End Entity", 365*cryptoutilDateTime.Days1, []string{"example.com"}, []net.IP{net.ParseIP("127.0.0.1")}, []string{"test@example.com"}, []*url.URL{{Scheme: "https", Host: "example.com"}}, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, originalSubjects)
	verifyEndEntitySubject(t, err, endEntitySubject)

	originalSubjects = append(originalSubjects, endEntitySubject)

	// Test full round-trip: []Subject -> [][]byte -> []Subject
	serialized, err := SerializeSubjects(originalSubjects, true)
	require.NoError(t, err, "Failed to serialize subjects")

	deserializedSubjects, err := DeserializeSubjects(serialized)
	require.NoError(t, err, "Failed to deserialize subjects")
	require.Len(t, deserializedSubjects, len(originalSubjects), "Deserialized count should match original")

	// Verify each subject was correctly round-tripped
	for i, originalSubject := range originalSubjects {
		deserializedSubject := deserializedSubjects[i]

		// Verify basic metadata
		require.Equal(t, originalSubject.SubjectName, deserializedSubject.SubjectName, "SubjectName mismatch at index %d", i)
		require.Equal(t, originalSubject.IssuerName, deserializedSubject.IssuerName, "IssuerName mismatch at index %d", i)

		// Verify key material
		require.NotNil(t, deserializedSubject.KeyMaterial.PrivateKey, "PrivateKey should not be nil at index %d", i)
		require.NotNil(t, deserializedSubject.KeyMaterial.PublicKey, "PublicKey should not be nil at index %d", i)
		require.NotEmpty(t, deserializedSubject.KeyMaterial.CertChain, "CertChain should not be empty at index %d", i)

		// Verify cert pools can be reconstructed correctly through BuildTLSCertificate
		// For CA subjects, verify that we can build TLS certificate and verify the cert chain
		if len(originalSubject.KeyMaterial.CertChain) > 1 {
			// Test that we can verify the cert chain using BuildTLSCertificate
			tlsCert, rootCACertsPool, intermediateCertsPool, err := BuildTLSCertificate(deserializedSubject)
			require.NoError(t, err, "BuildTLSCertificate should not fail at index %d", i)
			require.NotNil(t, rootCACertsPool, "Root CA cert pool should be reconstructed at index %d", i)
			require.NotEmpty(t, tlsCert.Certificate, "TLS certificate should have cert chain at index %d", i)

			leafCert := deserializedSubject.KeyMaterial.CertChain[0]

			// Use the intermediate certificate pool returned by BuildTLSCertificate
			// (No need to manually create it anymore - BuildTLSCertificate provides it)

			// Verify the certificate chain using the reconstructed root pool
			verifyOptions := x509.VerifyOptions{
				Roots:         rootCACertsPool,
				Intermediates: intermediateCertsPool,
				KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
			}

			_, verifyErr := leafCert.Verify(verifyOptions)
			require.NoError(t, verifyErr, "Certificate chain verification failed for subject %d (%s)", i, originalSubject.SubjectName)
		}

		// Verify subject type
		require.Equal(t, originalSubject.IsCA, deserializedSubject.IsCA, "IsCA mismatch at index %d", i)
		if originalSubject.IsCA {
			require.Equal(t, originalSubject.MaxPathLen, deserializedSubject.MaxPathLen, "MaxPathLen mismatch at index %d", i)
		} else {
			require.Equal(t, originalSubject.DNSNames, deserializedSubject.DNSNames, "DNSNames mismatch at index %d", i)

			// Compare IP addresses with tolerance for IPv4/IPv6 format differences
			require.Len(t, deserializedSubject.IPAddresses, len(originalSubject.IPAddresses), "IPAddresses length mismatch at index %d", i)
			for j, originalIP := range originalSubject.IPAddresses {
				restoredIP := deserializedSubject.IPAddresses[j]
				// Compare the actual IP addresses, allowing for IPv4/IPv6 representation differences
				require.True(t, originalIP.Equal(restoredIP), "IPAddresses[%d] mismatch at index %d: expected %v, got %v", j, i, originalIP, restoredIP)
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
		_, err := toKeyMaterialEncoded(nil, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "keyMaterial cannot be nil")
	})

	t.Run("nil PublicKey", func(t *testing.T) {
		keyPair := testKeyGenPool.Get()

		// Create a certificate template and sign it to get a proper certificate
		certTemplate, err := CertificateTemplateCA("Test Issuer", "Test CA", 10*365*cryptoutilDateTime.Days1, 0)
		verifyCertificateTemplate(t, err, certTemplate)

		cert, _, _, err := SignCertificate(nil, keyPair.Private, certTemplate, keyPair.Public, x509.ECDSAWithSHA256)
		require.NoError(t, err)

		keyMaterial := &KeyMaterial{
			CertChain: []*x509.Certificate{cert},
			PublicKey: nil, // Should cause error
		}
		_, err = toKeyMaterialEncoded(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "PublicKey cannot be nil")
	})

	t.Run("empty cert chain", func(t *testing.T) {
		keyPair := testKeyGenPool.Get()
		keyMaterial := &KeyMaterial{
			CertChain: []*x509.Certificate{}, // Empty chain should cause error
			PublicKey: keyPair.Public,
		}
		_, err := toKeyMaterialEncoded(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate chain cannot be empty")
	})

	t.Run("nil cert in chain", func(t *testing.T) {
		keyPair := testKeyGenPool.Get()
		keyMaterial := &KeyMaterial{
			CertChain: []*x509.Certificate{nil}, // Nil cert should cause error
			PublicKey: keyPair.Public,
		}
		_, err := toKeyMaterialEncoded(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate at index 0 in chain cannot be nil")
	})
}

func TestNegativeDuration(t *testing.T) {
	_, err := CertificateTemplateCA("Root CA", "Root CA", -1, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}
