package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/url"
	"testing"
	"time"

	"bytes"
	"io"
	"net/http"
	"net/http/httptest"

	cryptoutilDateTime "cryptoutil/internal/common/util/datetime"

	"github.com/stretchr/testify/require"
)

func TestNegativeDuration(t *testing.T) {
	_, err := CertificateTemplateCA("Root CA", "Root CA", -1, 1)
	require.Error(t, err, "Creating a certificate with negative duration should fail")
}

type testCASubject struct {
	SubjectName string
	Duration    time.Duration
	MaxPathLen  int
	KeyMaterial KeyMaterial
}

type testEndEntitySubject struct {
	SubjectName    string
	Duration       time.Duration
	DNSNames       []string
	IPAddresses    []net.IP
	EmailAddresses []string
	URIs           []*url.URL
	KeyMaterial    KeyMaterial
}

func TestMutualTLS(t *testing.T) {
	tlsServerRootCert := testCASubject{SubjectName: "Test TLS Server Root CA", Duration: 10 * cryptoutilDateTime.Days365, MaxPathLen: 2, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsServerIntermediateCert := testCASubject{SubjectName: "Test TLS Server Intermediate CA", Duration: 5 * cryptoutilDateTime.Days365, MaxPathLen: 1, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsServerIssuingCert := testCASubject{SubjectName: "Test TLS Server Issuing CA", Duration: 2 * cryptoutilDateTime.Days365, MaxPathLen: 0, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsServerCert := testEndEntitySubject{SubjectName: "Test TLS Server End Entity", Duration: 397 * cryptoutilDateTime.Days1, DNSNames: []string{"localhost", "tlsserver.example.com"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsServerRootsPool := x509.NewCertPool()
	tlsServerIntermediatesPool := x509.NewCertPool()

	t.Run("TLS Server Chain", func(t *testing.T) {
		t.Run("TLS Server Root CA", func(t *testing.T) {
			rootCert1Template, err := CertificateTemplateCA(tlsServerRootCert.SubjectName, tlsServerRootCert.SubjectName, tlsServerRootCert.Duration, tlsServerRootCert.MaxPathLen)
			verifyCertificateTemplate(t, err, rootCert1Template)
			tlsServerRootCert.KeyMaterial.Cert, tlsServerRootCert.KeyMaterial.DER, tlsServerRootCert.KeyMaterial.PEM, err = SignCertificate(nil, tlsServerRootCert.KeyMaterial.KeyPair.Private, rootCert1Template, tlsServerRootCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsServerRootCert.KeyMaterial.Cert, tlsServerRootCert.KeyMaterial.DER, tlsServerRootCert.KeyMaterial.PEM, tlsServerRootCert.SubjectName, tlsServerRootCert.SubjectName, tlsServerRootCert.Duration, tlsServerRootCert.MaxPathLen)
			tlsServerRootsPool.AddCert(tlsServerRootCert.KeyMaterial.Cert) // subsequent verify cert chain needs the root CA
		})
		t.Run("TLS Server Intermediate CA", func(t *testing.T) {
			intermediateCert1Template, err := CertificateTemplateCA(tlsServerRootCert.SubjectName, tlsServerIntermediateCert.SubjectName, tlsServerIntermediateCert.Duration, tlsServerIntermediateCert.MaxPathLen)
			verifyCertificateTemplate(t, err, intermediateCert1Template)
			tlsServerIntermediateCert.KeyMaterial.Cert, tlsServerIntermediateCert.KeyMaterial.DER, tlsServerIntermediateCert.KeyMaterial.PEM, err = SignCertificate(tlsServerRootCert.KeyMaterial.Cert, tlsServerRootCert.KeyMaterial.KeyPair.Private, intermediateCert1Template, tlsServerIntermediateCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsServerIntermediateCert.KeyMaterial.Cert, tlsServerIntermediateCert.KeyMaterial.DER, tlsServerIntermediateCert.KeyMaterial.PEM, tlsServerRootCert.SubjectName, tlsServerIntermediateCert.SubjectName, tlsServerIntermediateCert.Duration, tlsServerIntermediateCert.MaxPathLen)
			verifyCertChain(t, tlsServerIntermediateCert.KeyMaterial.Cert, tlsServerRootsPool, tlsServerIntermediatesPool)
			tlsServerIntermediatesPool.AddCert(tlsServerIntermediateCert.KeyMaterial.Cert) // subsequent verify cert chain needs the intermediate CA
		})
		t.Run("TLS Server Issuing CA", func(t *testing.T) {
			issuingCert1Template, err := CertificateTemplateCA(tlsServerIntermediateCert.SubjectName, tlsServerIssuingCert.SubjectName, tlsServerIssuingCert.Duration, tlsServerIssuingCert.MaxPathLen)
			verifyCertificateTemplate(t, err, issuingCert1Template)
			tlsServerIssuingCert.KeyMaterial.Cert, tlsServerIssuingCert.KeyMaterial.DER, tlsServerIssuingCert.KeyMaterial.PEM, err = SignCertificate(tlsServerIntermediateCert.KeyMaterial.Cert, tlsServerIntermediateCert.KeyMaterial.KeyPair.Private, issuingCert1Template, tlsServerIssuingCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsServerIssuingCert.KeyMaterial.Cert, tlsServerIssuingCert.KeyMaterial.DER, tlsServerIssuingCert.KeyMaterial.PEM, tlsServerIntermediateCert.SubjectName, tlsServerIssuingCert.SubjectName, tlsServerIssuingCert.Duration, tlsServerIssuingCert.MaxPathLen)
			verifyCertChain(t, tlsServerIssuingCert.KeyMaterial.Cert, tlsServerRootsPool, tlsServerIntermediatesPool)
			tlsServerIntermediatesPool.AddCert(tlsServerIssuingCert.KeyMaterial.Cert) // subsequent verify cert chain needs the issuing CA
		})
		t.Run("TLS Server End Entity", func(t *testing.T) {
			tlsServerCert1Template, err := CertificateTemplateTLSServer(tlsServerIssuingCert.SubjectName, tlsServerCert.SubjectName, tlsServerCert.Duration, tlsServerCert.DNSNames, tlsServerCert.IPAddresses, tlsServerCert.EmailAddresses, tlsServerCert.URIs)
			verifyCertificateTemplate(t, err, tlsServerCert1Template)
			tlsServerCert.KeyMaterial.Cert, tlsServerCert.KeyMaterial.DER, tlsServerCert.KeyMaterial.PEM, err = SignCertificate(tlsServerIssuingCert.KeyMaterial.Cert, tlsServerIssuingCert.KeyMaterial.KeyPair.Private, tlsServerCert1Template, tlsServerCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyEndEntityCertificate(t, err, tlsServerCert.KeyMaterial.Cert, tlsServerCert.KeyMaterial.DER, tlsServerCert.KeyMaterial.PEM, tlsServerIssuingCert.SubjectName, tlsServerCert.SubjectName, tlsServerCert.Duration, tlsServerCert.DNSNames, tlsServerCert.IPAddresses, tlsServerCert.EmailAddresses, tlsServerCert.URIs)
			verifyCertChain(t, tlsServerCert.KeyMaterial.Cert, tlsServerRootsPool, tlsServerIntermediatesPool)
		})
	})

	tlsClientRootCert := testCASubject{SubjectName: "Test TLS Client Root CA", Duration: 10 * cryptoutilDateTime.Days365, MaxPathLen: 2, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientIntermediateCert := testCASubject{SubjectName: "Test TLS Client Intermediate CA", Duration: 5 * cryptoutilDateTime.Days365, MaxPathLen: 1, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientIssuingCert := testCASubject{SubjectName: "Test TLS Client Issuing CA", Duration: 2 * cryptoutilDateTime.Days365, MaxPathLen: 0, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientCertEndEntity := testEndEntitySubject{SubjectName: "TLS Client", Duration: 30 * cryptoutilDateTime.Days1, EmailAddresses: []string{"client1@tlsclient.example.com"}, KeyMaterial: KeyMaterial{KeyPair: testKeyGenPool.Get()}}
	tlsClientRootsPool := x509.NewCertPool()
	tlsClientIntermediatesPool := x509.NewCertPool()
	t.Run("TLS Client Chain", func(t *testing.T) {
		t.Run("TLS Client Root CA", func(t *testing.T) {
			rootCert2Template, err := CertificateTemplateCA(tlsClientRootCert.SubjectName, tlsClientRootCert.SubjectName, tlsClientRootCert.Duration, tlsClientRootCert.MaxPathLen)
			verifyCertificateTemplate(t, err, rootCert2Template)
			tlsClientRootCert.KeyMaterial.Cert, tlsClientRootCert.KeyMaterial.DER, tlsClientRootCert.KeyMaterial.PEM, err = SignCertificate(nil, tlsClientRootCert.KeyMaterial.KeyPair.Private, rootCert2Template, tlsClientRootCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsClientRootCert.KeyMaterial.Cert, tlsClientRootCert.KeyMaterial.DER, tlsClientRootCert.KeyMaterial.PEM, tlsClientRootCert.SubjectName, tlsClientRootCert.SubjectName, tlsClientRootCert.Duration, tlsClientRootCert.MaxPathLen)
			tlsClientRootsPool.AddCert(tlsClientRootCert.KeyMaterial.Cert) // subsequent verify cert chain needs the root CA
		})
		t.Run("TLS Client Intermediate CA", func(t *testing.T) {
			intermediateCert2Template, err := CertificateTemplateCA(tlsClientRootCert.SubjectName, tlsClientIntermediateCert.SubjectName, tlsClientIntermediateCert.Duration, tlsClientIntermediateCert.MaxPathLen)
			verifyCertificateTemplate(t, err, intermediateCert2Template)
			tlsClientIntermediateCert.KeyMaterial.Cert, tlsClientIntermediateCert.KeyMaterial.DER, tlsClientIntermediateCert.KeyMaterial.PEM, err = SignCertificate(tlsClientRootCert.KeyMaterial.Cert, tlsClientRootCert.KeyMaterial.KeyPair.Private, intermediateCert2Template, tlsClientIntermediateCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsClientIntermediateCert.KeyMaterial.Cert, tlsClientIntermediateCert.KeyMaterial.DER, tlsClientIntermediateCert.KeyMaterial.PEM, tlsClientRootCert.SubjectName, tlsClientIntermediateCert.SubjectName, tlsClientIntermediateCert.Duration, tlsClientIntermediateCert.MaxPathLen)
			verifyCertChain(t, tlsClientIntermediateCert.KeyMaterial.Cert, tlsClientRootsPool, tlsClientIntermediatesPool)
			tlsClientIntermediatesPool.AddCert(tlsClientIntermediateCert.KeyMaterial.Cert) // subsequent verify cert chain needs the intermediate CA
		})
		t.Run("TLS Client Issuing CA", func(t *testing.T) {
			issuingCert2Template, err := CertificateTemplateCA(tlsClientIntermediateCert.SubjectName, tlsClientIssuingCert.SubjectName, tlsClientIssuingCert.Duration, tlsClientIssuingCert.MaxPathLen)
			verifyCertificateTemplate(t, err, issuingCert2Template)
			tlsClientIssuingCert.KeyMaterial.Cert, tlsClientIssuingCert.KeyMaterial.DER, tlsClientIssuingCert.KeyMaterial.PEM, err = SignCertificate(tlsClientIntermediateCert.KeyMaterial.Cert, tlsClientIntermediateCert.KeyMaterial.KeyPair.Private, issuingCert2Template, tlsClientIssuingCert.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyCACertificate(t, err, tlsClientIssuingCert.KeyMaterial.Cert, tlsClientIssuingCert.KeyMaterial.DER, tlsClientIssuingCert.KeyMaterial.PEM, tlsClientIntermediateCert.SubjectName, tlsClientIssuingCert.SubjectName, tlsClientIssuingCert.Duration, tlsClientIssuingCert.MaxPathLen)
			verifyCertChain(t, tlsClientIssuingCert.KeyMaterial.Cert, tlsClientRootsPool, tlsClientIntermediatesPool)
			tlsClientIntermediatesPool.AddCert(tlsClientIssuingCert.KeyMaterial.Cert) // subsequent verify cert chain needs the issuing CA
		})
		t.Run("TLS Client End Entity", func(t *testing.T) {
			tlsClientCert2Template, err := CertificateTemplateTLSClient(tlsClientIssuingCert.SubjectName, tlsClientCertEndEntity.SubjectName, tlsClientCertEndEntity.Duration, tlsClientCertEndEntity.DNSNames, tlsClientCertEndEntity.IPAddresses, tlsClientCertEndEntity.EmailAddresses, tlsClientCertEndEntity.URIs)
			verifyCertificateTemplate(t, err, tlsClientCert2Template)

			tlsClientCertEndEntity.KeyMaterial.Cert, tlsClientCertEndEntity.KeyMaterial.DER, tlsClientCertEndEntity.KeyMaterial.PEM, err = SignCertificate(tlsClientIssuingCert.KeyMaterial.Cert, tlsClientIssuingCert.KeyMaterial.KeyPair.Private, tlsClientCert2Template, tlsClientCertEndEntity.KeyMaterial.KeyPair.Public, x509.ECDSAWithSHA256)
			verifyEndEntityCertificate(t, err, tlsClientCertEndEntity.KeyMaterial.Cert, tlsClientCertEndEntity.KeyMaterial.DER, tlsClientCertEndEntity.KeyMaterial.PEM, tlsClientIssuingCert.SubjectName, tlsClientCertEndEntity.SubjectName, tlsClientCertEndEntity.Duration, tlsClientCertEndEntity.DNSNames, tlsClientCertEndEntity.IPAddresses, tlsClientCertEndEntity.EmailAddresses, tlsClientCertEndEntity.URIs)
			verifyCertChain(t, tlsClientCertEndEntity.KeyMaterial.Cert, tlsClientRootsPool, tlsClientIntermediatesPool)
		})
	})

	// These TLS certificate chain instances are reusable for both the Raw mTLS and HTTP mTLS tests
	serverTLSCertChain := tls.Certificate{
		Certificate: [][]byte{tlsServerCert.KeyMaterial.DER, tlsServerIssuingCert.KeyMaterial.DER, tlsServerIntermediateCert.KeyMaterial.DER, tlsServerRootCert.KeyMaterial.DER},
		PrivateKey:  tlsServerCert.KeyMaterial.KeyPair.Private,
	}
	clientTLSCertChain := tls.Certificate{
		Certificate: [][]byte{tlsClientCertEndEntity.KeyMaterial.DER, tlsClientIssuingCert.KeyMaterial.DER, tlsClientIntermediateCert.KeyMaterial.DER, tlsClientRootCert.KeyMaterial.DER},
		PrivateKey:  tlsClientCertEndEntity.KeyMaterial.KeyPair.Private,
	}

	// These TLS configuration instances are reusable for both the Raw mTLS and HTTP mTLS tests
	serverTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSCertChain},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    tlsClientRootsPool,
	}
	clientTLSConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientTLSCertChain},
		RootCAs:            tlsServerRootsPool,
		InsecureSkipVerify: false,
	}

	t.Run("Raw mTLS", func(t *testing.T) {
		ln, err := tls.Listen("tcp", "127.0.0.1:0", serverTLSConfig)
		require.NoError(t, err, "Failed to start TLS server")
		defer ln.Close()
		serverErrCh := make(chan error, 1)
		go func() {
			conn, err := ln.Accept()
			if err != nil {
				serverErrCh <- err
				return
			}
			defer conn.Close()
			buf := make([]byte, 512)
			n, err := conn.Read(buf)
			if err != nil {
				serverErrCh <- err
				return
			}
			_, err = conn.Write(buf[:n])
			serverErrCh <- err
		}()

		addr := ln.Addr().String()
		conn, err := tls.Dial("tcp", addr, clientTLSConfig)
		require.NoError(t, err, "Client failed to connect to TLS server")
		defer conn.Close()

		testMsg := []byte("hello mutual tls")
		_, err = conn.Write(testMsg)
		require.NoError(t, err, "Client failed to write to server")
		resp := make([]byte, len(testMsg))
		_, err = conn.Read(resp)
		require.NoError(t, err, "Client failed to read from server")
		require.Equal(t, testMsg, resp, "Echoed message mismatch")

		// Ensure server goroutine completed without error
		require.NoError(t, <-serverErrCh, "Server goroutine error")
	})

	t.Run("HTTP mTLS", func(t *testing.T) {
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, err := io.ReadAll(r.Body)
			require.NoError(t, err, "Server failed to read request body")
			_, err = w.Write(data)
			require.NoError(t, err, "Server failed to write response")
		})
		httpsServer := httptest.NewUnstartedServer(httpHandler)
		httpsServer.TLS = serverTLSConfig
		httpsServer.StartTLS()
		defer httpsServer.Close()

		httpsClient := &http.Client{Transport: &http.Transport{TLSClientConfig: clientTLSConfig}}
		testMsg := []byte("hello mutual https")
		resp, err := httpsClient.Post(httpsServer.URL, "text/plain", bytes.NewReader(testMsg))
		require.NoError(t, err, "Client failed to POST to HTTPS server")
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, "Unexpected HTTP status")
		echoed, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Client failed to read response body")
		require.Equal(t, testMsg, echoed, "Echoed message mismatch")
	})
}
