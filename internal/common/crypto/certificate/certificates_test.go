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

	tlsServerCertChain, tlsServerRootCAs, err := BuildTLSCertificate(tlsServerEndEntitySubject)
	require.NoError(t, err, "Failed to build TLS server certificate")
	tlsClientCertChain, tlsClientRootCAs, err := BuildTLSCertificate(tlsClientEndEntitySubject)
	require.NoError(t, err, "Failed to build TLS client certificate")

	// TLS configuration instances are reusable for both of the Raw mTLS and HTTP mTLS tests
	serverTLSConfig := &tls.Config{Certificates: []tls.Certificate{tlsServerCertChain}, ClientCAs: tlsClientRootCAs, ClientAuth: tls.RequireAndVerifyClientCert}
	clientTLSConfig := &tls.Config{Certificates: []tls.Certificate{tlsClientCertChain}, RootCAs: tlsServerRootCAs, InsecureSkipVerify: false}

	t.Run("Raw mTLS", func(t *testing.T) {
		callerShutdownSignalCh := make(chan struct{})
		tlsListenerAddress, err := startTlsEchoServer("127.0.0.1:0", 100*time.Millisecond, 100*time.Millisecond, serverTLSConfig, callerShutdownSignalCh) // or "0.0.0.0:0" for all interfaces
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
		httpsServer, serverURL, err := startHTTPSEchoServer("127.0.0.1:0", 100*time.Millisecond, 100*time.Millisecond, serverTLSConfig) // or "0.0.0.0:0" for all interfaces
		require.NoError(t, err, "failed to start HTTPS Echo Server")
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
	originalSubjects, err := CreateCASubjects(testKeyGenPool, "Test Serialize CA", 2)
	require.NoError(t, err, "Failed to create CA subjects")

	serializedSubjects, err := SerializeSubjects(originalSubjects, true)
	require.NoError(t, err, "Failed to serialize subjects")
	require.NotEmpty(t, serializedSubjects, "Serialized data should not be empty")
	t.Logf("Serialized Subjects count: %d", len(serializedSubjects))
	for i, subject := range serializedSubjects {
		t.Logf("Serialized Subject %d: %s", i, string(subject))
	}

	deserializedSubjects, err := DeserializeSubjects(serializedSubjects)
	require.NoError(t, err, "Failed to deserialize subjects")
	require.Len(t, deserializedSubjects, len(originalSubjects), "Deserialized Subject count should match original")

	for i, originalSubject := range originalSubjects {
		deserializedSubject := deserializedSubjects[i]
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

func TestNewFieldsPopulated(t *testing.T) {
	// Test CA subjects have proper fields populated
	subjects, err := CreateCASubjects(testKeyGenPool, "Test Fields CA", 2)
	require.NoError(t, err, "Failed to create CA subjects")

	// Root CA (index 0)
	rootCA := subjects[0]
	require.Equal(t, "Test Fields CA 0", rootCA.SubjectName, "Root CA subject name should match")
	require.Equal(t, "Test Fields CA 0", rootCA.IssuerName, "Root CA issuer name should be self-signed")
	require.True(t, rootCA.IsCA, "Root CA should have IsCA=true")
	require.Equal(t, 1, rootCA.MaxPathLen, "Root CA should have expected MaxPathLen")

	// Intermediate CA (index 1)
	intermediateCA := subjects[1]
	require.Equal(t, "Test Fields CA 1", intermediateCA.SubjectName, "Intermediate CA subject name should match")
	require.Equal(t, "Test Fields CA 0", intermediateCA.IssuerName, "Intermediate CA should be issued by root CA")
	require.True(t, intermediateCA.IsCA, "Intermediate CA should have IsCA=true")
	require.Equal(t, 0, intermediateCA.MaxPathLen, "Intermediate CA should have expected MaxPathLen")

	// Test End Entity subjects have proper fields populated
	endEntitySubject, err := CreateEndEntitySubject(testKeyGenPool, "Test Fields End Entity", 30*cryptoutilDateTime.Days1,
		[]string{"test.example.com"}, []net.IP{net.ParseIP("127.0.0.1")}, nil, nil,
		x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, subjects)
	require.NoError(t, err, "Failed to create End Entity subject")

	require.Equal(t, "Test Fields End Entity", endEntitySubject.SubjectName, "End entity subject name should match")
	require.Equal(t, "Test Fields CA 1", endEntitySubject.IssuerName, "End entity should be issued by leaf CA")
	require.False(t, endEntitySubject.IsCA, "End entity should have IsCA=false")
	require.Equal(t, []string{"test.example.com"}, endEntitySubject.DNSNames, "End entity DNS names should match")
}

func TestCompleteSubjectRoundTripSerialization(t *testing.T) {
	// Create test data
	subjects, err := CreateCASubjects(testKeyGenPool, "Round Trip CA", 2)
	require.NoError(t, err, "Failed to create CA subjects")

	// Add an end entity subject
	endEntitySubject, err := CreateEndEntitySubject(testKeyGenPool, "Round Trip End Entity", 365*cryptoutilDateTime.Days1, []string{"example.com"}, []net.IP{net.ParseIP("127.0.0.1")}, []string{"test@example.com"}, []*url.URL{{Scheme: "https", Host: "example.com"}}, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, subjects)
	require.NoError(t, err, "Failed to create end entity subject")
	subjects = append(subjects, endEntitySubject)

	// Test full round-trip: []Subject -> [][]byte -> []Subject
	serialized, err := SerializeSubjects(subjects, true)
	require.NoError(t, err, "Failed to serialize subjects")

	deserialized, err := DeserializeSubjects(serialized)
	require.NoError(t, err, "Failed to deserialize subjects")
	require.Len(t, deserialized, len(subjects), "Deserialized count should match original")

	// Verify each subject was correctly round-tripped
	for i, original := range subjects {
		restored := deserialized[i]

		// Verify basic metadata
		require.Equal(t, original.SubjectName, restored.SubjectName, "SubjectName mismatch at index %d", i)
		require.Equal(t, original.IssuerName, restored.IssuerName, "IssuerName mismatch at index %d", i)

		// Verify key material
		require.NotNil(t, restored.KeyMaterial.PrivateKey, "PrivateKey should not be nil at index %d", i)
		require.NotNil(t, restored.KeyMaterial.PublicKey, "PublicKey should not be nil at index %d", i)
		require.NotEmpty(t, restored.KeyMaterial.CertChain, "CertChain should not be empty at index %d", i)

		// Verify cert pools can be reconstructed correctly through BuildTLSCertificate
		// For CA subjects, verify that we can build TLS certificate and verify the cert chain
		if len(original.KeyMaterial.CertChain) > 1 {
			// Test that we can verify the cert chain using BuildTLSCertificate
			tlsCert, rootCACertsPool, err := BuildTLSCertificate(restored)
			require.NoError(t, err, "BuildTLSCertificate should not fail at index %d", i)
			require.NotNil(t, rootCACertsPool, "Root CA cert pool should be reconstructed at index %d", i)
			require.NotEmpty(t, tlsCert.Certificate, "TLS certificate should have cert chain at index %d", i)

			leafCert := restored.KeyMaterial.CertChain[0]
			intermediates := x509.NewCertPool()

			// Add intermediate CAs to the intermediates pool for verification
			for j := 1; j < len(restored.KeyMaterial.CertChain)-1; j++ {
				intermediates.AddCert(restored.KeyMaterial.CertChain[j])
			}

			// Verify the certificate chain using the reconstructed root pool
			verifyOptions := x509.VerifyOptions{
				Roots:         rootCACertsPool,
				Intermediates: intermediates,
				KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
			}

			_, verifyErr := leafCert.Verify(verifyOptions)
			require.NoError(t, verifyErr, "Certificate chain verification failed for subject %d (%s)", i, original.SubjectName)
		}

		// Verify subject type
		require.Equal(t, original.IsCA, restored.IsCA, "IsCA mismatch at index %d", i)
		if original.IsCA {
			require.Equal(t, original.MaxPathLen, restored.MaxPathLen, "MaxPathLen mismatch at index %d", i)
		} else {
			require.Equal(t, original.DNSNames, restored.DNSNames, "DNSNames mismatch at index %d", i)

			// Compare IP addresses with tolerance for IPv4/IPv6 format differences
			require.Len(t, restored.IPAddresses, len(original.IPAddresses), "IPAddresses length mismatch at index %d", i)
			for j, originalIP := range original.IPAddresses {
				restoredIP := restored.IPAddresses[j]
				// Compare the actual IP addresses, allowing for IPv4/IPv6 representation differences
				require.True(t, originalIP.Equal(restoredIP), "IPAddresses[%d] mismatch at index %d: expected %v, got %v", j, i, originalIP, restoredIP)
			}

			require.Equal(t, original.EmailAddresses, restored.EmailAddresses, "EmailAddresses mismatch at index %d", i)
			require.Equal(t, original.URIs, restored.URIs, "URIs mismatch at index %d", i)
		}

		t.Logf("Subject %d (%s) successfully round-tripped", i, original.SubjectName)
	}

	t.Logf("Successfully round-tripped %d subjects (including %d CAs and %d end entities)",
		len(subjects), len(subjects)-1, 1)
}

func TestSerializeSubjectsValidation(t *testing.T) {
	t.Run("nil subjects slice", func(t *testing.T) {
		_, err := SerializeSubjects(nil, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "subjects cannot be nil")
	})

	t.Run("empty SubjectName", func(t *testing.T) {
		subjects := []Subject{{
			SubjectName: "", // Empty subject name should cause error
			IssuerName:  "Test Issuer",
		}}
		_, err := SerializeSubjects(subjects, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "has empty SubjectName")
	})

	t.Run("end-entity with CA MaxPathLen", func(t *testing.T) {
		subjects := []Subject{{
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
		subjects := []Subject{{
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
		subjects := []Subject{{
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

func TestToKeyMaterialEncodedValidation(t *testing.T) {
	t.Run("nil keyMaterial", func(t *testing.T) {
		_, err := toKeyMaterialEncoded(nil, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "keyMaterial cannot be nil")
	})

	t.Run("nil PublicKey", func(t *testing.T) {
		keyPair := testKeyGenPool.Get()

		// Create a certificate template and sign it to get a proper certificate
		certTemplate, err := CertificateTemplateCA("Test Issuer", "Test CA", 10*365*cryptoutilDateTime.Days1, 0)
		require.NoError(t, err)

		cert, _, _, err := SignCertificate(nil, keyPair.Private, certTemplate, keyPair.Public, x509.ECDSAWithSHA256)
		require.NoError(t, err)

		keyMaterial := &KeyMaterial{
			PublicKey: nil, // Should cause error
			CertChain: []*x509.Certificate{cert},
		}
		_, err = toKeyMaterialEncoded(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "PublicKey cannot be nil")
	})

	t.Run("empty cert chain", func(t *testing.T) {
		keyPair := testKeyGenPool.Get()
		keyMaterial := &KeyMaterial{
			PublicKey: keyPair.Public,
			CertChain: []*x509.Certificate{}, // Empty chain should cause error
		}
		_, err := toKeyMaterialEncoded(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate chain cannot be empty")
	})

	t.Run("nil cert in chain", func(t *testing.T) {
		keyPair := testKeyGenPool.Get()
		keyMaterial := &KeyMaterial{
			PublicKey: keyPair.Public,
			CertChain: []*x509.Certificate{nil}, // Nil cert should cause error
		}
		_, err := toKeyMaterialEncoded(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate at index 0 in chain cannot be nil")
	})
}
