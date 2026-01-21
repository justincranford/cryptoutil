// Copyright (c) 2025 Justin Cranford
//
//

package certificate

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilDateTime "cryptoutil/internal/shared/util/datetime"

	"github.com/stretchr/testify/require"
)

const (
	// Server timeouts for Raw TLS echo server.
	testTLSServerStartupDelay = cryptoutilMagic.TestTLSServerStartupDelay
	testTLSServerWriteTimeout = cryptoutilMagic.TestTLSServerWriteTimeout
	testTLSServerReadTimeout  = cryptoutilMagic.TestTLSServerReadTimeout
	testTLSRetryBaseDelay     = cryptoutilMagic.TestTLSRetryBaseDelay
	testTLSMaxRetries         = cryptoutilMagic.TestTLSMaxRetries

	// Server timeouts for HTTPS echo server.
	testHTTPServerStartupDelay = cryptoutilMagic.TestHTTPServerStartupDelay
	testHTTPServerWriteTimeout = cryptoutilMagic.TestHTTPServerWriteTimeout
	testHTTPServerReadTimeout  = cryptoutilMagic.TestHTTPServerReadTimeout
	testHTTPRetryBaseDelay     = cryptoutilMagic.TestHTTPRetryBaseDelay
	testHTTPMaxRetries         = cryptoutilMagic.TestHTTPMaxRetries

	// Certificate validity durations.
	testCACertValidity10Years        = cryptoutilMagic.TLSDefaultValidityCACertYears * 365 * cryptoutilDateTime.Days1
	testCACertValidity20Years        = cryptoutilMagic.TLSTestCACertValidity20Years * 365 * cryptoutilDateTime.Days1
	testCACertValidity5Years         = cryptoutilMagic.TLSTestCACertValidity5Years * 365 * cryptoutilDateTime.Days1
	testEndEntityCertValidity396Days = cryptoutilMagic.TLSTestEndEntityCertValidity396Days * cryptoutilDateTime.Days1
	testEndEntityCertValidity30Days  = cryptoutilMagic.TLSTestEndEntityCertValidity30Days * cryptoutilDateTime.Days1
	testEndEntityCertValidity1Year   = cryptoutilMagic.TLSTestEndEntityCertValidity1Year * cryptoutilDateTime.Days1

	// Test constants.
	testNegativeDuration = cryptoutilMagic.TestNegativeDuration
	testHourDuration     = cryptoutilMagic.TestHourDuration
)

func TestMutualTLS(t *testing.T) {
	tlsServerSubjectsKeyPairs := testKeyGenPool.GetMany(4) // End Entity + 2 Intermediate CAs + Root CA
	tlsClientSubjectsKeyPairs := testKeyGenPool.GetMany(3) // End Entity + 1 Intermediate CA + Root CA

	tlsServerCASubjects, err := CreateCASubjects(tlsServerSubjectsKeyPairs[1:], "Test TLS Server CA", testCACertValidity10Years)
	verifyCASubjects(t, err, tlsServerCASubjects)
	tlsClientCASubjects, err := CreateCASubjects(tlsClientSubjectsKeyPairs[1:], "Test TLS Client CA", testCACertValidity10Years)
	verifyCASubjects(t, err, tlsClientCASubjects)

	tlsServerEndEntitySubject, err := CreateEndEntitySubject(tlsServerCASubjects[0], tlsServerSubjectsKeyPairs[0], "Test TLS Server End Entity", testEndEntityCertValidity396Days, []string{"localhost", "tlsserver.example.com"}, []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, nil, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
	verifyEndEntitySubject(t, err, tlsServerEndEntitySubject)
	tlsClientEndEntitySubject, err := CreateEndEntitySubject(tlsClientCASubjects[0], tlsClientSubjectsKeyPairs[0], "Test TLS Client End Entity", testEndEntityCertValidity30Days, nil, nil, []string{"client1@tlsclient.example.com"}, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
	verifyEndEntitySubject(t, err, tlsClientEndEntitySubject)

	tlsServerCertChain, tlsServerRootCAs, _, err := BuildTLSCertificate(tlsServerEndEntitySubject)
	require.NoError(t, err, "Failed to build TLS server certificate")
	tlsClientCertChain, tlsClientRootCAs, _, err := BuildTLSCertificate(tlsClientEndEntitySubject)
	require.NoError(t, err, "Failed to build TLS client certificate")

	// TLS configuration instances are reusable for both of the Raw mTLS and HTTP mTLS tests
	serverTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{*tlsServerCertChain},
		ClientCAs:    tlsClientRootCAs,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}
	clientTLSConfig := &tls.Config{
		Certificates:       []tls.Certificate{*tlsClientCertChain},
		RootCAs:            tlsServerRootCAs,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
	}

	const clientConnections = 10

	t.Run("Raw mTLS", func(t *testing.T) {
		callerShutdownSignalCh := make(chan struct{})
		tlsListenerAddress, err := startTLSEchoServer("127.0.0.1:0", testTLSServerReadTimeout, testTLSServerWriteTimeout, serverTLSConfig, callerShutdownSignalCh) // or "0.0.0.0:0" for all interfaces
		require.NoError(t, err, "failed to start TLS Echo Server")

		defer close(callerShutdownSignalCh)

		// Brief delay to ensure server goroutine is ready to accept connections
		time.Sleep(testTLSServerStartupDelay)

		tlsClientRequestBody := []byte("Hello Mutual TLS!")

		for i := 1; i <= clientConnections; i++ {
			func() {
				var tlsClientConnection *tls.Conn

				var err error
				// Retry connection up to testTLSMaxRetries times with backoff to handle timing issues under load
				for retry := 0; retry < testTLSMaxRetries; retry++ {
					conn, dialErr := (&tls.Dialer{Config: clientTLSConfig}).DialContext(context.Background(), "tcp", tlsListenerAddress)
					if dialErr == nil {
						var ok bool

						tlsClientConnection, ok = conn.(*tls.Conn)
						if !ok {
							err = fmt.Errorf("connection is not a TLS connection")
						} else {
							err = nil

							break
						}
					} else {
						err = dialErr
					}

					time.Sleep(time.Duration(retry+1) * testTLSRetryBaseDelay)
				}

				require.NoError(t, err, "client failed to connect to TLS Echo Server after retries")

				defer func() {
					if err := tlsClientConnection.Close(); err != nil {
						t.Logf("warning: failed to close TLS connection: %v", err)
					}
				}()

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
		httpsServer, serverURL, err := startHTTPSEchoServer("127.0.0.1:0", testHTTPServerReadTimeout, testHTTPServerWriteTimeout, serverTLSConfig) // or "0.0.0.0:0" for all interfaces
		require.NoError(t, err, "failed to start HTTPS Echo Server")

		defer func() {
			if err := httpsServer.Close(); err != nil {
				t.Logf("warning: failed to close HTTPS server: %v", err)
			}
		}()

		// Brief delay to ensure server goroutine is ready to accept connections
		time.Sleep(testHTTPServerStartupDelay)

		httpsClientRequestBody := []byte("Hello Mutual HTTPS!")
		httpsClient := &http.Client{
			Transport: &http.Transport{TLSClientConfig: clientTLSConfig},
			Timeout:   5 * time.Second, // Increase client timeout to prevent flaky failures
		}

		for i := 1; i <= clientConnections; i++ {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, serverURL, bytes.NewReader(httpsClientRequestBody))
			require.NoError(t, err, "failed to create POST request (%d of %d)", i, clientConnections)
			req.Header.Set("Content-Type", "text/plain")

			var httpsServerResponse *http.Response
			// Retry HTTP request up to testHTTPMaxRetries times with backoff to handle timing issues under load
			for retry := 0; retry < testHTTPMaxRetries; retry++ {
				httpsServerResponse, err = httpsClient.Do(req)
				if err == nil && httpsServerResponse.StatusCode == http.StatusOK {
					break
				}

				if httpsServerResponse != nil {
					httpsServerResponse.Body.Close() //nolint:errcheck,gosec // G104: Retry loop cleanup, error ignored intentionally
				}

				time.Sleep(time.Duration(retry+1) * testHTTPRetryBaseDelay)
			}

			require.NoError(t, err, "client failed to POST to HTTPS Echo Server after retries (%d of %d)", i, clientConnections)
			require.Equal(t, http.StatusOK, httpsServerResponse.StatusCode, "Unexpected HTTP status (%d of %d)", i, clientConnections)
			func() {
				defer func() {
					if err := httpsServerResponse.Body.Close(); err != nil {
						t.Logf("warning: failed to close response body: %v", err)
					}
				}()

				httpServerResponseBody, err := io.ReadAll(httpsServerResponse.Body)
				require.NoError(t, err, "client failed to read response body (%d of %d)", i, clientConnections)
				require.Equal(t, httpsClientRequestBody, httpServerResponseBody, "Echoed message mismatch (%d of %d)", i, clientConnections)
			}()
		}
	})
}

func TestSerializeCASubjects(t *testing.T) {
	subjectsKeyPairs := testKeyGenPool.GetMany(3)

	rootCASubject, err := CreateCASubject(nil, nil, "Round Trip Root CA", subjectsKeyPairs[0], testCACertValidity20Years, 2)
	rootCASubjects := []*Subject{rootCASubject}
	verifyCASubjects(t, err, rootCASubjects)
	testSerializeDeserialize(t, rootCASubjects)

	intermediateCASubject, err := CreateCASubject(rootCASubject, rootCASubject.KeyMaterial.PrivateKey, "Round Trip Intermediate CA", subjectsKeyPairs[1], testCACertValidity10Years, 1) // pragma: allowlist secret
	rootCASubject.KeyMaterial.PrivateKey = nil
	intermediateCASubjects := []*Subject{intermediateCASubject, rootCASubject} // pragma: allowlist secret
	verifyCASubjects(t, err, intermediateCASubjects)
	testSerializeDeserialize(t, intermediateCASubjects)

	issuingCASubject, err := CreateCASubject(intermediateCASubject, intermediateCASubject.KeyMaterial.PrivateKey, "Round Trip Issuing CA", subjectsKeyPairs[2], testCACertValidity5Years, 0)
	intermediateCASubject.KeyMaterial.PrivateKey = nil
	issuingCASubjects := []*Subject{issuingCASubject, intermediateCASubject, rootCASubject}
	verifyCASubjects(t, err, issuingCASubjects)
	testSerializeDeserialize(t, issuingCASubjects)

	intermediateCASubject.KeyMaterial.PrivateKey = nil
}

func TestSerializeEndEntitySubjects(t *testing.T) {
	subjectsKeyPairs := testKeyGenPool.GetMany(3)
	originalCASubjects, err := CreateCASubjects(subjectsKeyPairs[1:], "Round Trip CA", testCACertValidity10Years)
	verifyCASubjects(t, err, originalCASubjects)

	endEntitySubject, err := CreateEndEntitySubject(originalCASubjects[0], subjectsKeyPairs[0], "Round Trip End Entity", testEndEntityCertValidity1Year, []string{"example.com"}, []net.IP{net.ParseIP("127.0.0.1")}, []string{"test@example.com"}, []*url.URL{{Scheme: "https", Host: "example.com"}}, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}) // pragma: allowlist secret
	verifyEndEntitySubject(t, err, endEntitySubject)

	originalCASubjects[0].KeyMaterial.PrivateKey = nil
	originalSubjects := append([]*Subject{endEntitySubject}, originalCASubjects...)

	testSerializeDeserialize(t, originalSubjects)
}

func testSerializeDeserialize(t *testing.T, originalSubjects []*Subject) {
	t.Helper()

	for _, includePrivateKey := range []bool{false, true} {
		t.Run(fmt.Sprintf("includePrivateKey = %t", includePrivateKey), func(t *testing.T) {
			serializedSubjects, err := SerializeSubjects(originalSubjects, includePrivateKey) // pragma: allowlist secret
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
					require.NotNil(t, deserializedKeyMaterial.PrivateKey, "PrivateKey should not be nil %d when includePrivateKey=true", i) // pragma: allowlist secret
					require.Equal(t, originalKeyMaterial.PrivateKey, deserializedKeyMaterial.PrivateKey, "PrivateKey mismatch %d", i)       // pragma: allowlist secret
				} else {
					require.Nil(t, deserializedKeyMaterial.PrivateKey, "PrivateKey should be nil %d when includePrivateKey=false", i) // pragma: allowlist secret
				}
			}

			if len(deserializedSubjects[0].KeyMaterial.CertificateChain) > 1 && includePrivateKey { // pragma: allowlist secret
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
			Duration:    testHourDuration,
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
			Duration:    testHourDuration,
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
			Duration:    testHourDuration,
			IsCA:        true,
			MaxPathLen:  testNegativeDuration, // Invalid negative MaxPathLen
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
		keyPair := testKeyGenPool.GetMany(1)[0]

		certTemplate, err := CertificateTemplateCA("Test Issuer", "Test CA", testCACertValidity10Years, 0)
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
		keyPair := testKeyGenPool.GetMany(1)[0]
		keyMaterial := &KeyMaterial{
			CertificateChain: []*x509.Certificate{}, // Empty chain should cause error
			PublicKey:        keyPair.Public,
		}
		_, err := serializeKeyMaterial(keyMaterial, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate chain cannot be empty")
	})
	t.Run("nil cert in chain", func(t *testing.T) {
		keyPair := testKeyGenPool.GetMany(1)[0]
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
	_, err := CertificateTemplateCA("Root CA", "Root CA", testNegativeDuration, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}
