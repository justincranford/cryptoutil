// Copyright (c) 2025 Justin Cranford
//
//

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

// PublicHTTPServer implements the PublicServer interface for business logic APIs and UIs.
// Binds to configurable address and port from ServiceFrameworkServerSettings.
//
// Request Path Prefixes:
// - /service/** : Service-to-service APIs (headless clients, IP allowlist, rate limiting)
// - /browser/** : Browser-to-service APIs/UI (sessions, CSRF, CORS, CSP headers)
//
// Both paths serve the SAME OpenAPI specification but with different middleware stacks.
type PublicHTTPServer struct {
	app           *fiber.App
	settings      *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
	actualPort    int
	listener      net.Listener
	tlsMaterial   *cryptoutilAppsFrameworkServiceConfig.TLSMaterial
	listenFn      func(ctx context.Context, network, address string) (net.Listener, error)
	appListenerFn func(app *fiber.App, ln net.Listener) error
	mu            sync.RWMutex
	shutdown      bool
}

// NewPublicHTTPServer creates a new public HTTPS server instance.
//
// The server starts in shutdown=false state and ready=false state.
// Applications must call SetReady(true) after initializing dependencies (database, cache, etc.).
//
// Parameters:
// - ctx: Context for initialization (must not be nil)
// - settings: ServiceFrameworkServerSettings containing bind address, port, and paths (must not be nil)
// - tlsCfg: TLS configuration (mode, certificates, parameters)
//
// Returns:
// - *PublicHTTPServer: Server instance ready to Start()
// - error: Non-nil if initialization fails (nil context, TLS generation failure, Fiber setup failure).
func NewPublicHTTPServer(ctx context.Context, settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings, tlsCfg *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings) (*PublicHTTPServer, error) {
	return newPublicHTTPServerInternal(ctx, settings, tlsCfg,
		cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial,
		func(ctx context.Context, network, address string) (net.Listener, error) {
			return (&net.ListenConfig{}).Listen(ctx, network, address)
		},
		func(app *fiber.App, ln net.Listener) error {
			return app.Listener(ln) //nolint:wrapcheck // Pass-through to Fiber framework.
		},
		os.ReadFile,
	)
}

func newPublicHTTPServerInternal(
	ctx context.Context,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	tlsCfg *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings,
	generateTLSMaterialFn func(cfg *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings) (*cryptoutilAppsFrameworkServiceConfig.TLSMaterial, error),
	listenFn func(ctx context.Context, network, address string) (net.Listener, error),
	appListenerFn func(app *fiber.App, ln net.Listener) error,
	osReadFileFn func(name string) ([]byte, error),
) (*PublicHTTPServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	} else if tlsCfg == nil {
		return nil, fmt.Errorf("TLS config cannot be nil")
	}

	// Generate TLS material based on configured mode.
	tlsMaterial, err := generateTLSMaterialFn(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS material: %w", err)
	}

	// Apply public mTLS cert overrides from file paths (Cat 3 server cert + Cat 4 client CA).
	// When set, overrides auto-generated TLS with production certs and requires client certs.
	if err := applyPublicMTLS(settings, tlsMaterial, osReadFileFn); err != nil {
		return nil, fmt.Errorf("failed to apply public mTLS configuration: %w", err)
	}

	server := &PublicHTTPServer{
		settings:      settings,
		tlsMaterial:   tlsMaterial,
		listenFn:      listenFn,
		appListenerFn: appListenerFn,
	}

	server.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Public API",
		ReadTimeout:           cryptoutilSharedMagic.DefaultHTTPServerTimeoutSeconds * time.Second,
		WriteTimeout:          cryptoutilSharedMagic.DefaultHTTPServerTimeoutSeconds * time.Second,
		IdleTimeout:           cryptoutilSharedMagic.DefaultHTTPServerTimeoutSeconds * time.Second,
	})

	// Register public routes (placeholder - to be implemented by services).
	server.registerRoutes()

	return server, nil
}

// registerRoutes registers public HTTP endpoints.
// This is a placeholder - services will inject their own route handlers.
func (s *PublicHTTPServer) registerRoutes() {
	// Service-to-service paths.
	s.app.Get(cryptoutilSharedMagic.IdentityE2EHealthEndpoint, s.handleServiceHealth)

	// Browser-to-service paths.
	s.app.Get("/browser/api/v1/health", s.handleBrowserHealth)
}

// handleServiceHealth returns health status for service-to-service clients.
func (s *PublicHTTPServer) handleServiceHealth(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			cryptoutilSharedMagic.StringStatus: "shutting down",
		}); err != nil {
			return fmt.Errorf("failed to send service health shutdown response: %w", err)
		}

		return nil
	}

	if err := c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: cryptoutilSharedMagic.DockerServiceHealthHealthy,
	}); err != nil {
		return fmt.Errorf("failed to send service health response: %w", err)
	}

	return nil
}

// handleBrowserHealth returns health status for browser clients.
func (s *PublicHTTPServer) handleBrowserHealth(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			cryptoutilSharedMagic.StringStatus: "shutting down",
		}); err != nil {
			return fmt.Errorf("failed to send browser health shutdown response: %w", err)
		}

		return nil
	}

	if err := c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: cryptoutilSharedMagic.DockerServiceHealthHealthy,
	}); err != nil {
		return fmt.Errorf("failed to send browser health response: %w", err)
	}

	return nil
}

// Start starts the public HTTPS server and blocks until shutdown or error.
//
// The server:
// 1. Uses TLS material generated during NewPublicHTTPServer (configured mode)
// 2. Creates TCP listener on configured address and port from ServiceFrameworkServerSettings
// 3. Starts HTTPS server with Fiber app
// 4. Blocks until context cancelled or server error
// 5. Triggers graceful shutdown on context cancellation
//
// Parameters:
// - ctx: Context for server lifecycle (cancellation triggers shutdown)
//
// Returns:
// - error: Non-nil if server fails to start or encounters runtime error.
func (s *PublicHTTPServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Create TCP listener using address and port from ServiceFrameworkServerSettings.
	listener, err := s.listenFn(ctx, "tcp", net.JoinHostPort(s.settings.BindPublicAddress, strconv.Itoa(int(s.settings.BindPublicPort))))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.mu.Lock()
	s.listener = listener

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		s.mu.Unlock()

		return fmt.Errorf("listener address is not *net.TCPAddr")
	}

	s.actualPort = tcpAddr.Port
	s.mu.Unlock()

	// Create TLS listener using configured TLS material.
	tlsListener := tls.NewListener(listener, s.tlsMaterial.Config)

	// Start server in goroutine.
	errChan := make(chan error, 1)

	go func() {
		if err := s.appListenerFn(s.app, tlsListener); err != nil {
			errChan <- fmt.Errorf("public server error: %w", err)
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

		return fmt.Errorf("public server stopped: %w", ctx.Err())
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully shuts down the public server.
func (s *PublicHTTPServer) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	s.mu.Lock()

	if s.shutdown {
		s.mu.Unlock()

		return fmt.Errorf("public server already shutdown")
	}

	s.shutdown = true
	s.mu.Unlock()

	// Shutdown Fiber app.
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("failed to shutdown public server: %w", err)
	}

	return nil
}

// ActualPort returns the actual port the server is listening on (after dynamic allocation).
func (s *PublicHTTPServer) ActualPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.actualPort
}

// PublicBaseURL returns the base URL for public API access.
func (s *PublicHTTPServer) PublicBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("%s://%s:%d", s.settings.BindPublicProtocol, s.settings.BindPublicAddress, s.actualPort)
}

// App returns the underlying fiber.App for in-memory testing.
// This allows tests to use app.Test() without starting an HTTPS listener.
func (s *PublicHTTPServer) App() *fiber.App {
	return s.app
}

// PublicTLSRootCAPool returns the root CA certificate pool for the public server's TLS chain.
// Used by test infrastructure to configure secure HTTP clients without InsecureSkipVerify.
func (s *PublicHTTPServer) PublicTLSRootCAPool() *x509.CertPool {
	if s.tlsMaterial == nil {
		return nil
	}

	return s.tlsMaterial.RootCAPool
}

// applyPublicMTLS applies public mTLS configuration from file paths in settings.
// When PublicTLSCertFile and PublicTLSKeyFile are set, the auto-generated TLS cert is replaced
// with the static cert from files (Cat 3: public-https-server-entity-{PS-ID}).
// When PublicTLSCAFile is set, client certificate verification is enabled using the CA
// truststore (Cat 4: public-https-client-issuing-ca) with tls.RequireAndVerifyClientCert.
// Both fields are independent: cert override and client auth can be configured separately.
func applyPublicMTLS(
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	tlsMaterial *cryptoutilAppsFrameworkServiceConfig.TLSMaterial,
	osReadFileFn func(name string) ([]byte, error),
) error {
	// Override server cert from Cat 3 file paths when both cert and key are configured.
	if settings.PublicTLSCertFile != "" && settings.PublicTLSKeyFile != "" {
		certPEM, err := osReadFileFn(settings.PublicTLSCertFile)
		if err != nil {
			return fmt.Errorf("failed to read public TLS cert file %q: %w", settings.PublicTLSCertFile, err)
		}

		keyPEM, err := osReadFileFn(settings.PublicTLSKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read public TLS key file %q: %w", settings.PublicTLSKeyFile, err)
		}

		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return fmt.Errorf("failed to parse public TLS cert+key pair: %w", err)
		}

		tlsMaterial.Config.Certificates = []tls.Certificate{cert}
	}

	// Enable client cert verification from Cat 4 CA truststore when CA file is configured.
	if settings.PublicTLSCAFile != "" {
		caPEM, err := osReadFileFn(settings.PublicTLSCAFile)
		if err != nil {
			return fmt.Errorf("failed to read public TLS CA file %q: %w", settings.PublicTLSCAFile, err)
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
				return fmt.Errorf("failed to parse CA certificate from %q: %w", settings.PublicTLSCAFile, parseErr)
			}

			clientCAPool.AddCert(caCert)
		}

		tlsMaterial.Config.ClientAuth = tls.RequireAndVerifyClientCert
		tlsMaterial.Config.ClientCAs = clientCAPool
	}

	return nil
}
