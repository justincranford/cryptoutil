package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func verifyCertificateTemplate(t *testing.T, err error, certTemplate *x509.Certificate) {
	require.NoError(t, err, "Failed to create certificate template")
	require.NotNil(t, certTemplate, "Certificate template should not be nil")
}

func verifyCACertificate(t *testing.T, err error, certChain []*x509.Certificate, DERChain [][]byte, PEMChain [][]byte, expectedIssuerName string, expectedSubjectName string, expectedDuration time.Duration, expectedMaxPathLen int) {
	require.NoError(t, err, "Failed to sign certificate")
	require.NotNil(t, certChain, "Signed certificate should not be nil")
	require.NotEmpty(t, DERChain, "Certificate bytes should not be empty")
	require.NotEmpty(t, PEMChain, "Certificate PEM should not be empty")
	now := time.Now().UTC()
	require.Equal(t, expectedIssuerName, certChain[0].Issuer.CommonName, "Issuer name mismatch")
	require.Equal(t, expectedSubjectName, certChain[0].Subject.CommonName, "Subject name mismatch")
	require.True(t, certChain[0].IsCA, "Certificate should be a CA certificate")
	require.True(t, certChain[0].BasicConstraintsValid, "Basic constraints should be valid")
	require.Equal(t, expectedMaxPathLen, certChain[0].MaxPathLen, "MaxPathLen mismatch")
	require.Equal(t, expectedMaxPathLen == 0, certChain[0].MaxPathLenZero, "MaxPathLenZero mismatch")
	require.Equal(t, x509.KeyUsageCertSign|x509.KeyUsageCRLSign, certChain[0].KeyUsage, "Key usage mismatch")
	require.Nil(t, certChain[0].ExtKeyUsage, "Extended key usage should be nil")
	require.True(t, certChain[0].NotBefore.Before(now), "NotBefore should be in the past")
	require.True(t, certChain[0].NotAfter.After(now), "NotAfter should be in the future")
	require.True(t, certChain[0].NotAfter.Sub(certChain[0].NotBefore) >= expectedDuration, "Certificate validity period should be >= duration")
}

func verifyEndEntityCertificate(t *testing.T, err error, cert *x509.Certificate, certDER []byte, certPEM []byte, expectedIssuerName string, expectedSubjectName string, expectedDuration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL) {
	require.NoError(t, err, "Failed to sign certificate")
	require.NotNil(t, cert, "Signed certificate should not be nil")
	require.NotEmpty(t, certDER, "Certificate bytes should not be empty")
	require.NotEmpty(t, certPEM, "Certificate PEM should not be empty")
	now := time.Now().UTC()
	require.Equal(t, expectedIssuerName, cert.Issuer.CommonName, "Issuer name mismatch")
	require.Equal(t, expectedSubjectName, cert.Subject.CommonName, "Subject name mismatch")
	require.False(t, cert.IsCA, "Certificate should not be a CA certificate")
	require.False(t, cert.BasicConstraintsValid, "Basic constraints should be invalid")
	require.Equal(t, cert.KeyUsage, x509.KeyUsageDigitalSignature, "Key usage mismatch")
	// require.ElementsMatch(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, cert.ExtKeyUsage, "Extended key usage mismatch")
	// require.ElementsMatch(t, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, cert.ExtKeyUsage, "Extended key usage mismatch")
	require.True(t, cert.NotBefore.Before(now), "NotBefore should be in the past")
	require.True(t, cert.NotAfter.After(now), "NotAfter should be in the future")
	require.True(t, cert.NotAfter.Sub(cert.NotBefore) >= expectedDuration, "Certificate validity period should be >= duration")
}

func verifyCertChain(t *testing.T, certificate *x509.Certificate, roots *x509.CertPool, intermediates *x509.CertPool) {
	x509VerifyOptions := x509.VerifyOptions{
		CurrentTime:   time.Now().UTC(),
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	chains, err := certificate.Verify(x509VerifyOptions)
	require.NoError(t, err, "Failed to verify intermediate certificate using root certificate")
	require.NotEmpty(t, chains, "Certificate chains should not be empty")
}

func startTlsEchoServer(tlsServerListener string, readTimeout time.Duration, writeTimeout time.Duration, serverTLSConfig *tls.Config, callerShutdownSignalCh <-chan struct{}) (string, error) {
	netListener, err := net.Listen("tcp", tlsServerListener)
	if err != nil {
		return "", fmt.Errorf("failed to start TCP Listener: %w", err)
	}
	netTCPListener, ok := netListener.(*net.TCPListener)
	if !ok {
		return "", fmt.Errorf("failed to cast net.Listener to *net.TCPListener")
	}

	// Configure TCP-level settings for robustness
	netTCPListener.SetDeadline(time.Time{}) // Clear any existing deadline

	tlsListener := tls.NewListener(netListener, serverTLSConfig)

	go func() {
		defer tlsListener.Close()
		osShutdownSignalCh := make(chan os.Signal, 1)
		signal.Notify(osShutdownSignalCh, os.Interrupt, syscall.SIGTERM)
		for {
			select {
			case <-callerShutdownSignalCh:
				log.Printf("stopping TLS Echo Server, caller shutdown signal received")
				return
			case <-osShutdownSignalCh:
				log.Printf("stopping TLS Echo Server, OS shutdown signal received")
				return
			default:
				netTCPListener.SetDeadline(time.Now().Add(readTimeout))
				tlsClientConnection, err := tlsListener.Accept()
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						continue // Timeout errors are expected and recoverable
					}
					// For other errors, check if they're likely recoverable
					// Connection refused, reset by peer, etc. might be recoverable
					switch {
					case err.Error() == "use of closed network connection":
						// Server is shutting down
						log.Printf("server shutting down: %v", err)
						return
					default:
						// For other errors, log and retry with backoff
						log.Printf("error accepting connection (will retry): %v", err)
						time.Sleep(100 * time.Millisecond) // Brief backoff on errors
						continue
					}
				}
				go func(conn net.Conn) {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("panic in TLS connection handler: %v", r)
						}
						conn.Close()
					}()

					// Set both read and write deadlines upfront
					conn.SetReadDeadline(time.Now().Add(readTimeout))
					conn.SetWriteDeadline(time.Now().Add(writeTimeout))

					tlsClientRequestBodyBuffer := make([]byte, 1024) // Increased buffer size
					bytesRead, err := conn.Read(tlsClientRequestBodyBuffer)
					if err != nil {
						if ne, ok := err.(net.Error); ok && ne.Timeout() {
							log.Printf("read timeout on TLS connection")
						} else {
							log.Printf("failed to read from TLS connection: %v", err)
						}
						return
					}
					// Do not treat empty request as shutdown; just ignore
					if bytesRead > 0 {
						// Refresh write deadline before writing
						conn.SetWriteDeadline(time.Now().Add(writeTimeout))
						_, err = conn.Write(tlsClientRequestBodyBuffer[:bytesRead])
						if err != nil {
							if ne, ok := err.(net.Error); ok && ne.Timeout() {
								log.Printf("write timeout on TLS connection")
							} else {
								log.Printf("failed to write to TLS connection: %v", err)
							}
						}
					}
				}(tlsClientConnection)
			}
		}
	}()

	return tlsListener.Addr().String(), nil
}

func startHTTPSEchoServer(httpsServerListener string, readTimeout time.Duration, writeTimeout time.Duration, serverTLSConfig *tls.Config) (*http.Server, string, error) {
	netListener, err := net.Listen("tcp", httpsServerListener)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start TCP Listener for HTTPS Server: %w", err)
	}
	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic in HTTP handler: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Limit request body size to prevent memory exhaustion
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit

		data, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("failed to read request body: %v", err)
			http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))

		_, err = w.Write(data)
		if err != nil {
			log.Printf("failed to write response: %v", err)
			// Don't call http.Error here as headers are already written
		}
	})
	server := &http.Server{
		Handler:           httpHandler,
		TLSConfig:         serverTLSConfig,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       30 * time.Second, // Close idle connections after 30s
		ReadHeaderTimeout: 10 * time.Second, // Timeout for reading headers (prevents slowloris)
		MaxHeaderBytes:    1 << 20,          // 1MB max header size (prevents large header attacks)
		ErrorLog:          log.New(os.Stderr, "https-server: ", log.LstdFlags),
	}
	go server.ServeTLS(netListener, "", "")
	url := "https://" + netListener.Addr().String()
	return server, url, nil
}
