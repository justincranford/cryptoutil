// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilNetwork "cryptoutil/internal/common/util/network"
	cryptoutilTLS "cryptoutil/internal/common/crypto/tls"
)

// AdminServer represents the private admin API server for OAuth 2.1 authorization service.
type AdminServer struct {
	config      *cryptoutilIdentityConfig.Config
	app         *fiber.App
	tlsConfig   *tls.Config
	certPool    *x509.CertPool
	certificate *tls.Certificate
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory
	actualPort  uint16
}

// NewAdminServer creates a new admin server for internal health checks and management.
func NewAdminServer(
	config *cryptoutilIdentityConfig.Config,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) (*AdminServer, error) {
	// Create TLS configuration for private server (TLS 1.3 only, optional mTLS).
	var (
		tlsConfig   *tls.Config
		certPool    *x509.CertPool
		certificate *tls.Certificate
	)

	if config.AuthZ.TLSEnabled {
		privateClientAuth := tls.RequireAndVerifyClientCert
		if config.DevMode {
			privateClientAuth = tls.NoClientCert
		}

		// Generate self-signed certificate for admin server.
		subject := &cryptoutilTLS.CertificateSubject{
			CommonName:   "identity-authz-admin",
			Organization: []string{"CryptoUtil"},
			Country:      []string{"US"},
			Locality:     []string{"Local"},
			DNSNames:     []string{"localhost", "authz-admin"},
			IPAddresses:  []string{"127.0.0.1", "::1"},
		}

		privateTLSConfig, err := cryptoutilTLS.NewServerConfig(&cryptoutilTLS.ServerConfigOptions{
			Subject:    subject,
			ClientAuth: privateClientAuth,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to build private TLS server config: %w", err)
		}

		tlsConfig = privateTLSConfig.TLSConfig
		certPool = privateTLSConfig.RootCAsPool
		certificate = privateTLSConfig.Certificate
	}

	// Create Fiber app for admin endpoints.
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(cryptoutilIdentityMagic.FiberReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cryptoutilIdentityMagic.FiberWriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cryptoutilIdentityMagic.FiberIdleTimeoutSeconds) * time.Second,
	})

	// Register middleware.
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	app.Use(requestid.New())
	app.Use(compress.New())

	adminServer := &AdminServer{
		config:      config,
		app:         app,
		tlsConfig:   tlsConfig,
		certPool:    certPool,
		certificate: certificate,
		repoFactory: repoFactory,
	}

	// Register admin routes.
	adminServer.registerRoutes()

	return adminServer, nil
}

// registerRoutes registers admin API routes.
func (s *AdminServer) registerRoutes() {
	adminPath := s.config.PrivateAdminAPIContextPath

	// Health check endpoints.
	s.app.Get(adminPath+cryptoutilMagic.PrivateAdminLivezRequestPath, s.handleLiveness)
	s.app.Get(adminPath+cryptoutilMagic.PrivateAdminReadyzRequestPath, s.handleReadiness)
	s.app.Get(adminPath+cryptoutilMagic.PrivateAdminHealthzRequestPath, s.handleHealth)
	s.app.Post(adminPath+cryptoutilMagic.PrivateAdminShutdownRequestPath, s.handleShutdown)
}

// handleLiveness handles liveness probe requests.
func (s *AdminServer) handleLiveness(c *fiber.Ctx) error {
	healthStatus := map[string]any{
		cryptoutilMagic.StringStatus: cryptoutilMagic.StringStatusOK,
		"timestamp":                  time.Now().UTC().Format(time.RFC3339),
		"service":                    "identity-authz",
		"version":                    cryptoutilMagic.ServiceVersion,
		"probe":                      "liveness",
	}

	return c.Status(fiber.StatusOK).JSON(healthStatus)
}

// handleReadiness handles readiness probe requests.
func (s *AdminServer) handleReadiness(c *fiber.Ctx) error {
	healthStatus := map[string]any{
		cryptoutilMagic.StringStatus: cryptoutilMagic.StringStatusOK,
		"timestamp":                  time.Now().UTC().Format(time.RFC3339),
		"service":                    "identity-authz",
		"version":                    cryptoutilMagic.ServiceVersion,
		"probe":                      "readiness",
	}

	// Perform readiness checks concurrently.
	readinessResults := s.performConcurrentReadinessChecks()

	// Add results to health status.
	for checkName, result := range readinessResults {
		healthStatus[checkName] = result
	}

	// Check if any component is unhealthy for readiness.
	if dbStatus, ok := healthStatus["database"].(map[string]any); ok {
		if status, ok := dbStatus[cryptoutilMagic.StringStatus].(string); ok && status != cryptoutilMagic.StringStatusOK {
			healthStatus[cryptoutilMagic.StringStatus] = cryptoutilMagic.StringStatusDegraded
		}
	}

	statusCode := fiber.StatusOK
	if healthStatus[cryptoutilMagic.StringStatus] != cryptoutilMagic.StringStatusOK {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(healthStatus)
}

// handleHealth handles health check requests (combined liveness + readiness).
func (s *AdminServer) handleHealth(c *fiber.Ctx) error {
	// Health check is combination of liveness and readiness.
	return s.handleReadiness(c)
}

// handleShutdown handles graceful shutdown requests.
func (s *AdminServer) handleShutdown(c *fiber.Ctx) error {
	response := map[string]any{
		cryptoutilMagic.StringStatus: cryptoutilMagic.StringStatusOK,
		"message":                    "shutdown initiated",
		"timestamp":                  time.Now().UTC().Format(time.RFC3339),
	}

	// Send response before shutdown.
	if err := c.Status(fiber.StatusOK).JSON(response); err != nil {
		return fmt.Errorf("failed to send shutdown response: %w", err)
	}

	// Trigger graceful shutdown in background.
	go func() {
		time.Sleep(100 * time.Millisecond) // Allow response to be sent
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = s.Stop(ctx) //nolint:errcheck // Best-effort shutdown
	}()

	return nil
}

// performConcurrentReadinessChecks performs readiness checks concurrently.
func (s *AdminServer) performConcurrentReadinessChecks() map[string]any {
	results := make(map[string]any)

	// Channel to collect results.
	resultsChan := make(chan struct {
		name   string
		result any
	})

	// WaitGroup to wait for all checks to complete.
	var wg sync.WaitGroup

	// Helper function to perform a check and send the result to the channel.
	doCheck := func(name string, checkFunc func() any) {
		defer wg.Done()

		result := checkFunc()
		resultsChan <- struct {
			name   string
			result any
		}{name: name, result: result}
	}

	// Database check.
	wg.Add(1)

	go doCheck("database", func() any {
		db := s.repoFactory.DB()

		sqlDB, err := db.DB()
		if err != nil {
			return map[string]any{
				cryptoutilMagic.StringStatus: "error",
				"error":                      err.Error(),
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			return map[string]any{
				cryptoutilMagic.StringStatus: "error",
				"error":                      err.Error(),
			}
		}

		stats := sqlDB.Stats()

		return map[string]any{
			cryptoutilMagic.StringStatus: cryptoutilMagic.StringStatusOK,
			"db_type":                    "gorm",
			"open_connections":           stats.OpenConnections,
			"in_use_connections":         stats.InUse,
			"idle_connections":           stats.Idle,
			"max_open_connections":       stats.MaxOpenConnections,
			"wait_count":                 stats.WaitCount,
			"wait_duration":              stats.WaitDuration.String(),
		}
	})

	// Memory check.
	wg.Add(1)

	go doCheck("memory", func() any {
		return map[string]any{
			cryptoutilMagic.StringStatus: cryptoutilMagic.StringStatusOK,
			"note":                       "Memory metrics available via OTLP metrics",
		}
	})

	// Close results channel after all checks complete.
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results.
	for result := range resultsChan {
		results[result.name] = result.result
	}

	return results
}

// Start starts the admin server.
func (s *AdminServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", cryptoutilMagic.IPv4Loopback, s.config.PrivateAdminPort)

	// Start TLS server if enabled.
	if s.config.AuthZ.TLSEnabled && s.tlsConfig != nil {
		// Create TLS listener.
		ln, err := tls.Listen("tcp", addr, s.tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to create TLS listener: %w", err)
		}

		// Extract actual port (important for port 0 dynamic allocation).
		s.actualPort = uint16(ln.Addr().(*tls.Conn).NetConn().LocalAddr().(*net.TCPAddr).Port) //nolint:forcetypeassert // Type guaranteed by TLS listener

		// Start server with TLS listener.
		if err := s.app.Listener(ln); err != nil {
			return fmt.Errorf("failed to start admin HTTPS server: %w", err)
		}

		return nil
	}

	// Start HTTP server (dev mode only).
	if err := s.app.Listen(addr); err != nil {
		return fmt.Errorf("failed to start admin HTTP server: %w", err)
	}

	return nil
}

// Stop stops the admin server.
func (s *AdminServer) Stop(ctx context.Context) error {
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("failed to shutdown admin server: %w", err)
	}

	return nil
}

// SendLivenessCheck sends a liveness check request to the admin server.
func SendLivenessCheck(config *cryptoutilIdentityConfig.Config) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilMagic.ClientLivenessRequestTimeout)
	defer cancel()

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, config.PrivateAdminPort)

	_, _, result, err := cryptoutilNetwork.HTTPGetLivez(ctx, baseURL, config.PrivateAdminAPIContextPath, 0, nil, config.DevMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get liveness check: %w", err)
	}

	return result, nil
}

// SendReadinessCheck sends a readiness check request to the admin server.
func SendReadinessCheck(config *cryptoutilIdentityConfig.Config) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilMagic.ClientReadinessRequestTimeout)
	defer cancel()

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, config.PrivateAdminPort)

	_, _, result, err := cryptoutilNetwork.HTTPGetReadyz(ctx, baseURL, config.PrivateAdminAPIContextPath, 0, nil, config.DevMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get readiness check: %w", err)
	}

	return result, nil
}

// SendShutdownRequest sends a shutdown request to the admin server.
func SendShutdownRequest(config *cryptoutilIdentityConfig.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilMagic.ClientShutdownRequestTimeout)
	defer cancel()

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, config.PrivateAdminPort)

	_, _, _, err := cryptoutilNetwork.HTTPPostShutdown(ctx, baseURL, config.PrivateAdminAPIContextPath, 0, nil, config.DevMode)
	if err != nil {
		return fmt.Errorf("failed to send shutdown request: %w", err)
	}

	return nil
}
