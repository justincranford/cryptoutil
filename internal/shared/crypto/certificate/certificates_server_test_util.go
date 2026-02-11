// Copyright (c) 2025 Justin Cranford
//
//

package certificate

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	http "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func startTLSEchoServer(tlsServerListener string, readTimeout, writeTimeout time.Duration, serverTLSConfig *tls.Config, callerShutdownSignalCh <-chan struct{}) (string, error) {
	netListener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", tlsServerListener)
	if err != nil {
		return "", fmt.Errorf("failed to start TCP Listener: %w", err)
	}

	netTCPListener, ok := netListener.(*net.TCPListener)
	if !ok {
		return "", fmt.Errorf("failed to cast net.Listener to *net.TCPListener")
	}

	// Configure TCP-level settings for robustness
	if err := netTCPListener.SetDeadline(time.Time{}); err != nil {
		log.Printf("warning: failed to clear TCP deadline: %v", err)
	} // Clear any existing deadline

	tlsListener := tls.NewListener(netListener, serverTLSConfig)

	go func() {
		defer func() {
			if err := tlsListener.Close(); err != nil {
				log.Printf("warning: failed to close TLS listener: %v", err)
			}
		}()

		osShutdownSignalCh := make(chan os.Signal, 1)
		signal.Notify(osShutdownSignalCh, os.Interrupt, syscall.SIGTERM) //nolint:errcheck

		for {
			select {
			case <-callerShutdownSignalCh:
				log.Printf("stopping TLS Echo Server, caller shutdown signal received")

				return
			case <-osShutdownSignalCh:
				log.Printf("stopping TLS Echo Server, OS shutdown signal received")

				return
			default:
				if err := netTCPListener.SetDeadline(time.Now().UTC().Add(readTimeout)); err != nil {
					log.Printf("warning: failed to set TCP deadline: %v", err)
				}

				tlsClientConnection, err := tlsListener.Accept()
				if err != nil {
					var ne net.Error
					if errors.As(err, &ne) && ne.Timeout() {
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
						time.Sleep(cryptoutilSharedMagic.TestTLSClientRetryWait) // Brief backoff on errors

						continue
					}
				}

				go func(conn net.Conn) {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("panic in TLS connection handler: %v", r)
						}

						if err := conn.Close(); err != nil {
							log.Printf("warning: failed to close connection: %v", err)
						}
					}()

					// Set both read and write deadlines upfront
					if err := conn.SetReadDeadline(time.Now().UTC().Add(readTimeout)); err != nil {
						log.Printf("warning: failed to set read deadline: %v", err)
					}

					if err := conn.SetWriteDeadline(time.Now().UTC().Add(writeTimeout)); err != nil {
						log.Printf("warning: failed to set write deadline: %v", err)
					}

					tlsClientRequestBodyBuffer := make([]byte, 1024) // Increased buffer size

					bytesRead, err := conn.Read(tlsClientRequestBodyBuffer)
					if err != nil {
						var ne net.Error
						if errors.As(err, &ne) && ne.Timeout() {
							log.Printf("read timeout on TLS connection")
						} else {
							log.Printf("failed to read from TLS connection: %v", err)
						}

						return
					}
					// Do not treat empty request as shutdown; just ignore
					if bytesRead > 0 {
						// Refresh write deadline before writing
						if err := conn.SetWriteDeadline(time.Now().UTC().Add(writeTimeout)); err != nil {
							log.Printf("warning: failed to set write deadline: %v", err)
						}

						_, err = conn.Write(tlsClientRequestBodyBuffer[:bytesRead])
						if err != nil {
							var ne net.Error
							if errors.As(err, &ne) && ne.Timeout() {
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

func startHTTPSEchoServer(httpsServerListener string, readTimeout, writeTimeout time.Duration, serverTLSConfig *tls.Config) (*http.Server, string, error) {
	netListener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", httpsServerListener)
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
		r.Body = http.MaxBytesReader(w, r.Body, cryptoutilSharedMagic.DefaultHTTPRequestBodyLimit)

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
			// Don't call http.Error here as headers are already written
			log.Printf("failed to write response: %v", err)
		}
	})
	server := &http.Server{
		Handler:           httpHandler,
		TLSConfig:         serverTLSConfig,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       cryptoutilSharedMagic.TestDefaultServerIdleTimeout,       // Close idle connections after 30s
		ReadHeaderTimeout: cryptoutilSharedMagic.TestDefaultServerReadHeaderTimeout, // Timeout for reading headers (prevents slowloris)
		MaxHeaderBytes:    cryptoutilSharedMagic.TestDefaultServerMaxHeaderBytes,    // 1MB max header size (prevents large header attacks)
		ErrorLog:          log.New(os.Stderr, "https-server: ", log.LstdFlags),
	}

	go func() {
		if err := server.ServeTLS(netListener, "", ""); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTPS server error: %v", err)
		}
	}()

	url := "https://" + netListener.Addr().String()

	return server, url, nil
}
