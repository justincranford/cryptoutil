// Copyright (c) 2025 Justin Cranford
//
//

// Package listener provides a reusable template for dual HTTPS server pattern used across all cryptoutil services.
// AdminServer implements the private admin API server with health check endpoints and graceful shutdown.
package listener

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps/framework/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// AdminServer represents the private admin API server for health checks and graceful shutdown.
// Binds to address and port from ServiceFrameworkServerSettings.
type AdminServer struct {
	app           *fiber.App
	listener      net.Listener
	settings      *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
	actualPort    uint16
	tlsMaterial   *cryptoutilAppsFrameworkServiceConfig.TLSMaterial
	listenFn      func(ctx context.Context, network, address string) (net.Listener, error)
	appListenerFn func(app *fiber.App, ln net.Listener) error
	mu            sync.RWMutex
	ready         bool
	shutdown      bool
}

// NewAdminHTTPServer creates a new admin server instance for private administrative operations.
// settings: ServiceFrameworkServerSettings containing bind address, port, and paths (MUST NOT be nil).
// tlsCfg: TLS configuration (mode + parameters) for HTTPS server. MUST NOT be nil.
func NewAdminHTTPServer(ctx context.Context, settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings, tlsCfg *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings) (*AdminServer, error) {
	return newAdminHTTPServerInternal(ctx, settings, tlsCfg,
		cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial,
		func(ctx context.Context, network, address string) (net.Listener, error) {
			return (&net.ListenConfig{}).Listen(ctx, network, address)
		},
		func(app *fiber.App, ln net.Listener) error {
			//nolint:wrapcheck // Pass-through to Fiber framework.
			return app.Listener(ln)
		},
		os.ReadFile,
	)
}

func newAdminHTTPServerInternal(
	ctx context.Context,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	tlsCfg *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings,
	generateTLSMaterialFn func(cfg *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings) (*cryptoutilAppsFrameworkServiceConfig.TLSMaterial, error),
	listenFn func(ctx context.Context, network, address string) (net.Listener, error),
	appListenerFn func(app *fiber.App, ln net.Listener) error,
	osReadFileFn func(name string) ([]byte, error),
) (*AdminServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	if tlsCfg == nil {
		return nil, fmt.Errorf("TLS configuration cannot be nil")
	}

	// Generate TLS material based on configured mode.
	tlsMaterial, err := generateTLSMaterialFn(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS material: %w", err)
	}

	// Apply admin mTLS cert overrides from file paths (Cat 7 server cert + Cat 6 client CA).
	// When set, overrides auto-generated TLS with production certs and requires client certs.
	if err := applyAdminMTLS(settings, tlsMaterial, osReadFileFn); err != nil {
		return nil, fmt.Errorf("failed to apply admin mTLS configuration: %w", err)
	}

	server := &AdminServer{
		settings:      settings,
		tlsMaterial:   tlsMaterial,
		listenFn:      listenFn,
		appListenerFn: appListenerFn,
		ready:         false,
		shutdown:      false,
	}

	// Create Fiber app with minimal configuration.
	server.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Admin API",
		ReadTimeout:           cryptoutilSharedMagic.DefaultHTTPServerTimeoutSeconds * time.Second,
		WriteTimeout:          cryptoutilSharedMagic.DefaultHTTPServerTimeoutSeconds * time.Second,
		IdleTimeout:           cryptoutilSharedMagic.DefaultHTTPServerTimeoutSeconds * time.Second,
	})

	// Register admin routes.
	server.registerRoutes()

	return server, nil
}

// registerRoutes sets up admin API endpoints.
func (s *AdminServer) registerRoutes() {
	api := s.app.Group(cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath)

	// Health check endpoints.
	api.Get(cryptoutilSharedMagic.PrivateAdminLivezRequestPath, s.handleLivez)
	api.Get(cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, s.handleReadyz)

	// Graceful shutdown endpoint.
	api.Post(cryptoutilSharedMagic.PrivateAdminShutdownRequestPath, s.handleShutdown)
}

// handleLivez returns liveness status (200 if server is running).
// Liveness check: Is the process alive? Failure action: restart container.
func (s *AdminServer) handleLivez(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			cryptoutilSharedMagic.StringStatus: "shutting down",
		}); err != nil {
			return fmt.Errorf("failed to send livez shutdown response: %w", err)
		}

		return nil
	}

	if err := c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: "alive",
	}); err != nil {
		return fmt.Errorf("failed to send livez response: %w", err)
	}

	return nil
}

// handleReadyz returns readiness status (200 if server is ready to accept traffic).
// Readiness check: Is the service ready? Failure action: remove from load balancer (do NOT restart).
func (s *AdminServer) handleReadyz(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			cryptoutilSharedMagic.StringStatus: "shutting down",
		}); err != nil {
			return fmt.Errorf("failed to send readyz shutdown response: %w", err)
		}

		return nil
	}

	if !s.ready {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			cryptoutilSharedMagic.StringStatus: "not ready",
		}); err != nil {
			return fmt.Errorf("failed to send readyz not-ready response: %w", err)
		}

		return nil
	}

	if err := c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: "ready",
	}); err != nil {
		return fmt.Errorf("failed to send readyz response: %w", err)
	}

	return nil
}

// handleShutdown initiates graceful shutdown of the admin server.
func (s *AdminServer) handleShutdown(c *fiber.Ctx) error {
	s.mu.Lock()
	s.shutdown = true
	s.mu.Unlock()

	// Acknowledge shutdown request.
	_ = c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: "shutdown initiated",
	})

	// Trigger shutdown in background to avoid blocking response.
	go func() {
		// Wait for response to be sent.
		time.Sleep(cryptoutilSharedMagic.DefaultAdminServerShutdownDelay)

		// Shutdown server gracefully.
		ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultServerShutdownTimeout)
		defer cancel()

		_ = s.Shutdown(ctx)
	}()

	return nil
}

// Start begins listening on configured address and port from ServiceFrameworkServerSettings for admin API requests.
// This method blocks until shutdown is called or context is cancelled.
func (s *AdminServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Bind to address and port from ServiceFrameworkServerSettings.
	addr := net.JoinHostPort(s.settings.BindPrivateAddress, strconv.Itoa(int(s.settings.BindPrivatePort)))

	// Create listener.
	listener, err := s.listenFn(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create admin listener: %w", err)
	}

	s.listener = listener

	// Store actual port if dynamic allocation was used (port 0).
	// Use mutex to protect actualPort writes for concurrent access safety.
	s.mu.Lock()

	if s.settings.BindPrivatePort == 0 {
		tcpAddr, ok := listener.Addr().(*net.TCPAddr)
		if !ok {
			s.mu.Unlock()

			_ = listener.Close()

			return fmt.Errorf("listener address is not a TCP address")
		}

		if tcpAddr.Port < 0 || tcpAddr.Port > int(cryptoutilSharedMagic.MaxPortNumber) {
			s.mu.Unlock()

			_ = listener.Close()

			return fmt.Errorf("invalid port number: %d", tcpAddr.Port)
		}

		s.actualPort = uint16(tcpAddr.Port) //nolint:gosec // Port range validated above.
	} else {
		s.actualPort = s.settings.BindPrivatePort
	}

	s.mu.Unlock()

	// Wrap with TLS using pre-generated TLS configuration.
	tlsListener := tls.NewListener(listener, s.tlsMaterial.Config)

	// Note: Server starts with ready=false. Application should call SetReady(true) after initializing dependencies.

	// Start Fiber server in goroutine and monitor context cancellation.
	errChan := make(chan error, 1)

	go func() {
		if err := s.appListenerFn(s.app, tlsListener); err != nil {
			errChan <- fmt.Errorf("admin server error: %w", err)
		} else {
			errChan <- nil
		}
	}()

	// Wait for either context cancellation or server error.
	select {
	case <-ctx.Done():
		// Context cancelled - trigger graceful shutdown.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultServerShutdownTimeout)
		defer cancel()

		_ = s.Shutdown(shutdownCtx)

		return fmt.Errorf("admin server stopped: %w", ctx.Err())
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully shuts down the admin server.
func (s *AdminServer) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	s.mu.Lock()

	if s.shutdown {
		s.mu.Unlock()

		return nil // Already shutdown, return success.
	}

	s.shutdown = true
	s.ready = false
	s.mu.Unlock()

	// Shutdown Fiber app (this automatically closes the listener).
	if s.app != nil {
		if err := s.app.ShutdownWithContext(ctx); err != nil {
			return fmt.Errorf("failed to shutdown admin app: %w", err)
		}
	}

	// Do NOT explicitly close listener - Fiber's Shutdown already did this.
	// Attempting to close again causes "use of closed network connection" errors.

	return nil
}

// ActualPort returns the actual port the admin server is listening on.
// Returns 0 before Start() is called, or the dynamically allocated port after Start().
func (s *AdminServer) ActualPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return int(s.actualPort)
}

// AdminBaseURL returns the base URL for admin API access.
func (s *AdminServer) AdminBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("%s://%s:%d", s.settings.BindPrivateProtocol, s.settings.BindPrivateAddress, s.actualPort)
}

// App returns the underlying fiber.App for custom route registration.
// This allows callers to register additional admin endpoints before calling Start().
// Thread-safe with read lock.
func (s *AdminServer) App() *fiber.App {
	return s.app
}

// SetReady marks the server as ready to accept traffic.
// This is called by the application after dependencies are initialized.
// Thread-safe with full Lock.
func (s *AdminServer) SetReady(ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ready = ready
}

// AdminTLSRootCAPool returns the root CA certificate pool for the admin server's TLS chain.
// Used by test infrastructure to configure secure HTTP clients without InsecureSkipVerify.
func (s *AdminServer) AdminTLSRootCAPool() *x509.CertPool {
	if s.tlsMaterial == nil {
		return nil
	}

	return s.tlsMaterial.RootCAPool
}

// applyAdminMTLS applies admin mTLS configuration from file paths in settings.
// When AdminTLSCertFile and AdminTLSKeyFile are set, the auto-generated TLS cert is replaced
// with the static cert from files (Cat 7: private-https-mutual-entity).
// When AdminTLSCAFile is set, client certificate verification is enabled using the CA
// truststore (Cat 6: private-https-mutual-issuing-ca) with tls.RequireAndVerifyClientCert.
// Both fields are independent: cert override and client auth can be configured separately.
func applyAdminMTLS(
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	tlsMaterial *cryptoutilAppsFrameworkServiceConfig.TLSMaterial,
	osReadFileFn func(name string) ([]byte, error),
) error {
	// Override server cert from Cat 7 file paths when both cert and key are configured.
	if settings.AdminTLSCertFile != "" && settings.AdminTLSKeyFile != "" {
		certPEM, err := osReadFileFn(settings.AdminTLSCertFile)
		if err != nil {
			return fmt.Errorf("failed to read admin TLS cert file %q: %w", settings.AdminTLSCertFile, err)
		}

		keyPEM, err := osReadFileFn(settings.AdminTLSKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read admin TLS key file %q: %w", settings.AdminTLSKeyFile, err)
		}

		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return fmt.Errorf("failed to parse admin TLS cert+key pair: %w", err)
		}

		tlsMaterial.Config.Certificates = []tls.Certificate{cert}
	}

	// Enable client cert verification from Cat 6 CA truststore when CA file is configured.
	if settings.AdminTLSCAFile != "" {
		caPEM, err := osReadFileFn(settings.AdminTLSCAFile)
		if err != nil {
			return fmt.Errorf("failed to read admin TLS CA file %q: %w", settings.AdminTLSCAFile, err)
		}

		clientCAPool := x509.NewCertPool()

		for rest := caPEM; len(rest) > 0; {
			var block *pem.Block

			block, rest = pem.Decode(rest)
			if block == nil {
				break
			}

			if block.Type != cryptoutilSharedMagic.StringPEMTypeCertificate {
				continue
			}

			caCert, parseErr := x509.ParseCertificate(block.Bytes)
			if parseErr != nil {
				return fmt.Errorf("failed to parse CA certificate from %q: %w", settings.AdminTLSCAFile, parseErr)
			}

			clientCAPool.AddCert(caCert)
		}

		tlsMaterial.Config.ClientAuth = tls.RequireAndVerifyClientCert
		tlsMaterial.Config.ClientCAs = clientCAPool
	}

	return nil
}
