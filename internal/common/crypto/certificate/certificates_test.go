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
	tlsServerCASubjects, err := CreateCASubjects("Test TLS Server CA", 3) // Root CA + 2 Intermediate CAs
	require.NoError(t, err, "Failed to create TLS Server CA subjects")
	tlsClientCASubjects, err := CreateCASubjects("Test TLS Client CA", 2) // Root CA + 1 Intermediate CA
	require.NoError(t, err, "Failed to create TLS Client CA subjects")
	tlsServerEndEntitySubject, err := CreateEndEntitySubject("Test TLS Server End Entity", 397*cryptoutilDateTime.Days1, []string{"localhost", "tlsserver.example.com"}, []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, nil, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, tlsServerCASubjects)
	require.NoError(t, err, "Failed to create TLS Server End Entity subject")
	tlsClientEndEntitySubject, err := CreateEndEntitySubject("Test TLS Client End Entity", 30*cryptoutilDateTime.Days1, nil, nil, []string{"client1@tlsclient.example.com"}, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, tlsClientCASubjects)
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
	subjects, err := CreateCASubjects("Test Serialize CA", 2)
	require.NoError(t, err, "Failed to create CA subjects")

	serializedSubjects, err := SerializeSubjects(subjects)
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

func CreateCASubjects(caSubjectNamePrefix string, numCAs int) ([]Subject, error) {
	if numCAs <= 0 {
		return nil, fmt.Errorf("numCAs must be greater than 0")
	}
	subjects := make([]Subject, numCAs)
	for i := range numCAs {
		keyPair := testKeyGenPool.Get()
		if keyPair == nil {
			return nil, fmt.Errorf("keyPair should not be nil for CA %d", i)
		}
		if keyPair.Private == nil {
			return nil, fmt.Errorf("keyPair.Private should not be nil for CA %d", i)
		}
		if keyPair.Public == nil {
			return nil, fmt.Errorf("keyPair.Public should not be nil for CA %d", i)
		}

		// Determine issuer name - root CA issues itself, others are issued by previous CA
		issuerName := fmt.Sprintf("%s %d", caSubjectNamePrefix, i) // Self-signed for root CA
		if i > 0 {
			issuerName = fmt.Sprintf("%s %d", caSubjectNamePrefix, i-1) // Previous CA for intermediate CAs
		}

		currentSubject := Subject{
			SubjectName: fmt.Sprintf("%s %d", caSubjectNamePrefix, i),
			IssuerName:  issuerName,
			Duration:    10 * 365 * cryptoutilDateTime.Days1,
			KeyMaterial: KeyMaterialDecoded{
				PrivateKey:             keyPair.Private,
				PublicKey:              keyPair.Public,
				CertChain:              []*x509.Certificate{},
				RootCACertsPool:        x509.NewCertPool(),
				SubordinateCACertsPool: x509.NewCertPool(),
			},
			CASubject: &CASubject{
				MaxPathLen: numCAs - i - 1,
				IsCA:       true,
			},
		}
		previousSubject := currentSubject
		var previousCACert *x509.Certificate
		if i > 0 {
			previousSubject = subjects[i-1]
			previousCACert = previousSubject.KeyMaterial.CertChain[0]
		}

		currentCACertTemplate, err := CertificateTemplateCA(previousSubject.IssuerName, currentSubject.SubjectName, currentSubject.Duration, currentSubject.CASubject.MaxPathLen)
		if err != nil {
			return nil, fmt.Errorf("failed to create CA certificate template for %s: %w", currentSubject.SubjectName, err)
		}

		cert, _, pemBytes, err := SignCertificate(previousCACert, previousSubject.KeyMaterial.PrivateKey, currentCACertTemplate, currentSubject.KeyMaterial.PublicKey, x509.ECDSAWithSHA256)
		if err != nil {
			return nil, fmt.Errorf("failed to sign CA certificate for %s: %w", currentSubject.SubjectName, err)
		}

		currentSubject.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, previousSubject.KeyMaterial.CertChain...)

		// Create DER and PEM chains locally for verification
		derChain := make([][]byte, len(currentSubject.KeyMaterial.CertChain))
		pemChain := make([][]byte, len(currentSubject.KeyMaterial.CertChain))
		for j, c := range currentSubject.KeyMaterial.CertChain {
			derChain[j] = c.Raw
			pemChain[j] = pemBytes // Use the pemBytes from SignCertificate for the first cert
		}

		currentSubject.KeyMaterial.RootCACertsPool = previousSubject.KeyMaterial.RootCACertsPool.Clone()
		currentSubject.KeyMaterial.SubordinateCACertsPool = previousSubject.KeyMaterial.SubordinateCACertsPool.Clone()
		if i == 0 {
			currentSubject.KeyMaterial.RootCACertsPool.AddCert(cert)
		} else {
			currentSubject.KeyMaterial.SubordinateCACertsPool.AddCert(cert)
		}

		subjects[i] = currentSubject
	}
	return subjects, nil
}

func CreateEndEntitySubject(subjectName string, duration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage, caSubjects []Subject) (Subject, error) {
	keyPair := testKeyGenPool.Get()
	if keyPair == nil {
		return Subject{}, fmt.Errorf("keyPair should not be nil")
	}
	if keyPair.Private == nil {
		return Subject{}, fmt.Errorf("keyPair.Private should not be nil")
	}
	if keyPair.Public == nil {
		return Subject{}, fmt.Errorf("keyPair.Public should not be nil")
	}

	// The issuing CA is the last one in the chain (leaf CA)
	if len(caSubjects) == 0 {
		return Subject{}, fmt.Errorf("caSubjects should not be empty")
	}
	issuingCA := caSubjects[len(caSubjects)-1]
	if issuingCA.SubjectName == "" {
		return Subject{}, fmt.Errorf("issuingCA.SubjectName should not be empty")
	}

	endEntitySubject := Subject{
		SubjectName: subjectName,
		IssuerName:  issuingCA.SubjectName,
		Duration:    duration,
		KeyMaterial: KeyMaterialDecoded{
			PrivateKey:             keyPair.Private,
			PublicKey:              keyPair.Public,
			CertChain:              []*x509.Certificate{},
			RootCACertsPool:        x509.NewCertPool(),
			SubordinateCACertsPool: x509.NewCertPool(),
		},
		EndEntitySubject: &EndEntitySubject{
			DNSNames:       dnsNames,
			IPAddresses:    ipAddresses,
			EmailAddresses: emailAddresses,
			URIs:           uris,
		},
	}

	endEntityCertTemplate, err := CertificateTemplateEndEntity(issuingCA.SubjectName, endEntitySubject.SubjectName, endEntitySubject.Duration, endEntitySubject.EndEntitySubject.DNSNames, endEntitySubject.EndEntitySubject.IPAddresses, endEntitySubject.EndEntitySubject.EmailAddresses, endEntitySubject.EndEntitySubject.URIs, keyUsage, extKeyUsage)
	if err != nil {
		return Subject{}, fmt.Errorf("failed to create end entity certificate template for %s: %w", subjectName, err)
	}

	cert, _, _, err := SignCertificate(issuingCA.KeyMaterial.CertChain[0], issuingCA.KeyMaterial.PrivateKey, endEntityCertTemplate, endEntitySubject.KeyMaterial.PublicKey, x509.ECDSAWithSHA256)
	if err != nil {
		return Subject{}, fmt.Errorf("failed to sign end entity certificate for %s: %w", subjectName, err)
	}

	endEntitySubject.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, issuingCA.KeyMaterial.CertChain...)
	endEntitySubject.KeyMaterial.RootCACertsPool = issuingCA.KeyMaterial.RootCACertsPool.Clone()
	endEntitySubject.KeyMaterial.SubordinateCACertsPool = issuingCA.KeyMaterial.SubordinateCACertsPool.Clone()

	return endEntitySubject, nil
}

func BuildTLSCertificate(endEntitySubject Subject) (tls.Certificate, *x509.CertPool, error) {
	if len(endEntitySubject.KeyMaterial.CertChain) == 0 {
		return tls.Certificate{}, nil, fmt.Errorf("certificate chain is empty")
	}
	if endEntitySubject.KeyMaterial.PrivateKey == nil {
		return tls.Certificate{}, nil, fmt.Errorf("private key is nil")
	}
	if endEntitySubject.KeyMaterial.RootCACertsPool == nil {
		return tls.Certificate{}, nil, fmt.Errorf("root CA certs pool is nil")
	}

	// Convert certificate chain to DER format for TLS
	derCertChain := make([][]byte, len(endEntitySubject.KeyMaterial.CertChain))
	for i, cert := range endEntitySubject.KeyMaterial.CertChain {
		derCertChain[i] = cert.Raw
	}

	return tls.Certificate{Certificate: derCertChain, PrivateKey: endEntitySubject.KeyMaterial.PrivateKey, Leaf: endEntitySubject.KeyMaterial.CertChain[0]}, endEntitySubject.KeyMaterial.RootCACertsPool, nil
}

func TestNewFieldsPopulated(t *testing.T) {
	// Test CA subjects have proper fields populated
	subjects, err := CreateCASubjects("Test Fields CA", 2)
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
	endEntitySubject, err := CreateEndEntitySubject("Test Fields End Entity", 30*cryptoutilDateTime.Days1,
		[]string{"test.example.com"}, []net.IP{net.ParseIP("127.0.0.1")}, nil, nil,
		x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, subjects)
	require.NoError(t, err, "Failed to create End Entity subject")

	require.Equal(t, "Test Fields End Entity", endEntitySubject.SubjectName, "End entity subject name should match")
	require.Equal(t, "Test Fields CA 1", endEntitySubject.IssuerName, "End entity should be issued by leaf CA")
	require.NotNil(t, endEntitySubject.EndEntitySubject, "End entity should have EndEntitySubject populated")
	require.Equal(t, []string{"test.example.com"}, endEntitySubject.EndEntitySubject.DNSNames, "End entity DNS names should match")
}
