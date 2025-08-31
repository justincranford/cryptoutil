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

func TestMutualTLS(t *testing.T) {
	tlsServerSubjectsKeyPairs, err := getKeyPairs(4, testKeyGenPool) // End Entity + 2 Intermediate CAs + Root CA
	require.NoError(t, err, "Failed to get key pairs for CA subjects")
	tlsClientSubjectsKeyPairs, err := getKeyPairs(3, testKeyGenPool)
	require.NoError(t, err, "Failed to get key pairs for CA subjects") // End Entity + 1 Intermediate CA + Root CA

	tlsServerCASubjects, err := CreateCASubjects(tlsServerSubjectsKeyPairs[1:], "Test TLS Server CA", 10*365*cryptoutilDateTime.Days1)
	verifyCASubjects(t, err, tlsServerCASubjects)
	tlsClientCASubjects, err := CreateCASubjects(tlsClientSubjectsKeyPairs[1:], "Test TLS Client CA", 10*365*cryptoutilDateTime.Days1)
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

func TestSerializeCASubjects(t *testing.T) {
	subjectsKeyPairs, err := getKeyPairs(3, testKeyGenPool)
	require.NoError(t, err, "Failed to get key pairs for CA subjects")

	rootCASubject, err := CreateCASubject(nil, nil, "Round Trip Root CA", subjectsKeyPairs[0], 20*365*cryptoutilDateTime.Days1, 1)
	verifyCASubjects(t, err, []*Subject{rootCASubject})
	testSerializeDeserialize(t, []*Subject{rootCASubject})

	rootCASubject.KeyMaterial.PrivateKey = nil
	subCASubject, err := CreateCASubject(rootCASubject, subjectsKeyPairs[0].Private, "Round Trip Sub CA", subjectsKeyPairs[1], 20*365*cryptoutilDateTime.Days1, 0)
	verifyCASubjects(t, err, []*Subject{subCASubject, rootCASubject})
	testSerializeDeserialize(t, []*Subject{subCASubject, rootCASubject})
	subCASubject.KeyMaterial.PrivateKey = nil
}

func TestSerializeEndEntitySubjects(t *testing.T) {
	subjectsKeyPairs, err := getKeyPairs(3, testKeyGenPool)
	require.NoError(t, err, "Failed to get key pairs for CA subjects")

	originalCASubjects, err := CreateCASubjects(subjectsKeyPairs[1:], "Round Trip CA", 10*365*cryptoutilDateTime.Days1)
	verifyCASubjects(t, err, originalCASubjects)

	endEntitySubject, err := CreateEndEntitySubject(originalCASubjects[0], subjectsKeyPairs[0], "Round Trip End Entity", 365*cryptoutilDateTime.Days1, []string{"example.com"}, []net.IP{net.ParseIP("127.0.0.1")}, []string{"test@example.com"}, []*url.URL{{Scheme: "https", Host: "example.com"}}, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
	verifyEndEntitySubject(t, err, endEntitySubject)

	originalCASubjects[0].KeyMaterial.PrivateKey = nil
	originalSubjects := append([]*Subject{endEntitySubject}, originalCASubjects...)

	testSerializeDeserialize(t, originalSubjects)
}

func testSerializeDeserialize(t *testing.T, originalSubjects []*Subject) {
	for _, includePrivateKey := range []bool{false, true} {
		t.Run(fmt.Sprintf("includePrivateKey = %t", includePrivateKey), func(t *testing.T) {
			serializedSubjects, err := SerializeSubjects(originalSubjects, includePrivateKey)
			require.NoError(t, err, "Failed to serialize subjects (includePrivateKey=%t)", includePrivateKey)
			require.NotEmpty(t, serializedSubjects, "Serialized data should not be empty (includePrivateKey=%t)", includePrivateKey)

			deserializedSubjects, err := DeserializeSubjects(serializedSubjects)
			require.NoError(t, err, "Failed to deserialize subjects (includePrivateKey=%t)", includePrivateKey)
			require.Len(t, deserializedSubjects, len(originalSubjects), "Deserialized count should match original (includePrivateKey=%t)", includePrivateKey)

			// Verify full round-trip: []*Subject -> [][]byte -> []*Subject
			for i, originalSubject := range originalSubjects {
				deserializedSubject := deserializedSubjects[i]
				originalKeyMaterial := originalSubject.KeyMaterial
				deserializedKeyMaterial := deserializedSubject.KeyMaterial

				require.Equal(t, originalSubject.SubjectName, deserializedSubject.SubjectName, "SubjectName mismatch %d (includePrivateKey=%t)", i, includePrivateKey)
				require.Equal(t, originalSubject.IssuerName, deserializedSubject.IssuerName, "IssuerName mismatch %d (includePrivateKey=%t)", i, includePrivateKey)
				require.NotNil(t, deserializedKeyMaterial.PublicKey, "PublicKey should not be nil %d (includePrivateKey=%t)", i, includePrivateKey)
				require.NotEmpty(t, deserializedKeyMaterial.CertificateChain, "CertChain should not be empty %d (includePrivateKey=%t)", i, includePrivateKey)
				require.Equal(t, originalKeyMaterial.PublicKey, deserializedKeyMaterial.PublicKey, "PublicKey mismatch %d (includePrivateKey=%t)", i, includePrivateKey)
				require.Len(t, deserializedKeyMaterial.CertificateChain, len(originalKeyMaterial.CertificateChain), "CertChain length mismatch %d (includePrivateKey=%t)", i, includePrivateKey)
				for j, originalCertificate := range originalKeyMaterial.CertificateChain {
					require.Equal(t, originalCertificate.Raw, deserializedKeyMaterial.CertificateChain[j].Raw, "Certificate Raw data mismatch %d, cert %d (includePrivateKey=%t)", i, j, includePrivateKey)
				}
				require.Equal(t, originalSubject.IsCA, deserializedSubject.IsCA, "IsCA mismatch %d (includePrivateKey=%t)", i, includePrivateKey)
				require.Equal(t, originalSubject.MaxPathLen, deserializedSubject.MaxPathLen, "MaxPathLen mismatch %d", i)
				if !originalSubject.IsCA {
					require.Equal(t, originalSubject.DNSNames, deserializedSubject.DNSNames, "DNSNames mismatch %d", i)
					require.Len(t, deserializedSubject.IPAddresses, len(originalSubject.IPAddresses), "IPAddresses length mismatch %d", i)
					for j, originalIP := range originalSubject.IPAddresses {
						deserializedIPAddress := deserializedSubject.IPAddresses[j]
						require.True(t, originalIP.Equal(deserializedIPAddress), "IPAddresses[%d] mismatch %d: expected %v, got %v", j, i, originalIP, deserializedIPAddress)
					}
					require.Equal(t, originalSubject.EmailAddresses, deserializedSubject.EmailAddresses, "EmailAddresses mismatch %d", i)
					require.Equal(t, originalSubject.URIs, deserializedSubject.URIs, "URIs mismatch %d", i)
				}
				if includePrivateKey && i == 0 {
					require.NotNil(t, deserializedKeyMaterial.PrivateKey, "PrivateKey should not be nil %d when includePrivateKey=true", i)
					require.Equal(t, originalKeyMaterial.PrivateKey, deserializedKeyMaterial.PrivateKey, "PrivateKey mismatch %d", i)
				} else {
					require.Nil(t, deserializedKeyMaterial.PrivateKey, "PrivateKey should be nil %d when includePrivateKey=false", i)
				}
			}

			if len(deserializedSubjects[0].KeyMaterial.CertificateChain) > 1 && includePrivateKey {
				tlsCertificate, rootCACertificatesPool, intermediateCertificatesPool, err := BuildTLSCertificate(deserializedSubjects[0])
				require.NoError(t, err, "BuildTLSCertificate should not fail (includePrivateKey=%t)", includePrivateKey)
				require.NotNil(t, rootCACertificatesPool, "Root CA cert pool should be reconstructed (includePrivateKey=%t)", includePrivateKey)
				require.NotEmpty(t, tlsCertificate.Certificate, "TLS certificate should have cert chain (includePrivateKey=%t)", includePrivateKey)
				verifyOptions := x509.VerifyOptions{
					Roots:         rootCACertificatesPool,
					Intermediates: intermediateCertificatesPool,
					KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
				}
				_, err = deserializedSubjects[0].KeyMaterial.CertificateChain[0].Verify(verifyOptions)
				require.NoError(t, err, "Certificate chain verification failed for subject (%s) (includePrivateKey=%t)", originalSubjects[0].SubjectName, includePrivateKey)
			}
		})
	}
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
				CertificateChain: []*x509.Certificate{{}}, // Mock cert
				PublicKey:        []byte("mock-key"),
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
			CertificateChain: []*x509.Certificate{cert},
			PublicKey:        nil, // Should cause error
		}
		_, err = serializeKeyMaterial(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "PublicKey cannot be nil")
	})
	t.Run("empty cert chain", func(t *testing.T) {
		keyPair := testKeyGenPool.Get()
		keyMaterial := &KeyMaterial{
			CertificateChain: []*x509.Certificate{}, // Empty chain should cause error
			PublicKey:        keyPair.Public,
		}
		_, err := serializeKeyMaterial(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate chain cannot be empty")
	})
	t.Run("nil cert in chain", func(t *testing.T) {
		keyPair := testKeyGenPool.Get()
		keyMaterial := &KeyMaterial{
			CertificateChain: []*x509.Certificate{nil}, // Nil cert should cause error
			PublicKey:        keyPair.Public,
		}
		_, err := serializeKeyMaterial(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate 0 in chain cannot be nil")
	})
}

func TestNegativeDuration(t *testing.T) {
	_, err := CertificateTemplateCA("Root CA", "Root CA", -1, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}
