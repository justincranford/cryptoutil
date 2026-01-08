// Copyright (c) 2025 Justin Cranford
//
//

package im

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	"cryptoutil/internal/apps/cipher/im/domain"
	"cryptoutil/internal/apps/cipher/im/repository"
	"cryptoutil/internal/apps/cipher/im/server"
	"cryptoutil/internal/apps/cipher/im/server/config"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

const (
	helpCommand     = "help"
	helpFlag        = "--help"
	helpShortFlag   = "-h"
	urlFlag         = "--url"
	cacertFlag      = "--cacert"
	databaseURLFlag = "--database-url"

	// Default URLs for health check endpoints.
	defaultHealthURL   = "https://127.0.0.1:8888/health"
	defaultLivezURL    = "https://127.0.0.1:9090/admin/v1/livez"
	defaultReadyzURL   = "https://127.0.0.1:9090/admin/v1/readyz"
	defaultShutdownURL = "https://127.0.0.1:9090/admin/v1/shutdown"

	// Database dialector names.
	dialectSQLite     = "sqlite"
	dialectPostgreSQL = "postgres"

	// SQLite in-memory database URL for shared cache.
	sqliteInMemoryURL = "file::memory:?cache=shared"
	dialectPostgres   = "postgres"
	dialectPostgresPG = "pgx"
)

// IM implements the instant messaging service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func IM(args []string, stdout, stderr io.Writer) int {
	return internalIM(args, stdout, stderr)
}

// internalIM implements the instant messaging service subcommand handler with testable writers.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func internalIM(args []string, stdout, stderr io.Writer) int {
	// Default to "server" subcommand if no args provided (backward compatibility).
	if len(args) == 0 {
		args = []string{"server"}
	}

	// Check for help flags.
	if args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag {
		printIMUsage(stdout)

		return 0
	}

	// Route to subcommand.
	switch args[0] {
	case "version":
		printIMVersion(stdout)

		return 0
	case "server":
		return imServiceServerStart(args[1:], stdout, stderr)
	case "client":
		return imServiceClient(args[1:], stdout, stderr)
	case "init":
		return imServiceInit(args[1:], stdout, stderr)
	case "health":
		return imServiceHealth(args[1:], stdout, stderr)
	case "livez":
		return imServiceLivez(args[1:], stdout, stderr)
	case "readyz":
		return imServiceReadyz(args[1:], stdout, stderr)
	case "shutdown":
		return imServiceShutdown(args[1:], stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown subcommand: %s\n\n", args[0])

		printIMUsage(stdout)

		return 1
	}
}

// printIMVersion prints the instant messaging service version information.
func printIMVersion(stdout io.Writer) {
	_, _ = fmt.Fprintln(stdout, "cipher-im service")
	_, _ = fmt.Fprintln(stdout, "Part of cryptoutil cipher product")
	_, _ = fmt.Fprintln(stdout, "Version information available via Docker image tags")
}

// printIMUsage prints the instant messaging service usage information.
func printIMUsage(stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, IMUsageMain)
}

// imServiceServerStart implements the server subcommand.
func imServiceServerStart(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, IMUsageServer)

		return 0
	}

	ctx := context.Background()

	// Parse configuration using config.Parse() which leverages viper+pflag.
	// This replaces the manual flag parsing and DefaultAppConfig() pattern.
	// The Parse() function:
	//   1. Calls parent ServiceTemplateServerSettings.Parse() for base settings
	//   2. Adds cipher-im specific flags (JWE algorithm, message constraints, JWT secret)
	//   3. Merges config files, environment variables, and command-line flags
	//   4. Returns fully populated CipherImServerSettings
	//
	// Note: We prepend "start" as the subcommand for Parse() to validate.
	argsWithSubcommand := append([]string{"start"}, args...)

	cfg, err := config.Parse(argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	// Override demo-specific defaults if not explicitly set.
	// For production deployments, these should be provided via config file or environment variables.
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = sqliteInMemoryURL
	}

	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "cipher-im-demo-secret" // Demo secret for alpha project
	}

	srv, err := server.NewFromConfig(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	// Mark server as ready after successful initialization.
	// This enables /admin/v1/readyz to return 200 OK instead of 503 Service Unavailable.
	srv.SetReady(true)

	// Start server with graceful shutdown.
	errChan := make(chan error, 1)

	go func() {
		_, _ = fmt.Fprintf(stdout, "🚀 Starting cipher-im service...\n")
		_, _ = fmt.Fprintf(stdout, "   Public Server: https://127.0.0.1:%d\n", cfg.BindPublicPort)
		_, _ = fmt.Fprintf(stdout, "   Admin Server:  https://127.0.0.1:%d\n", cfg.BindPrivatePort)

		errChan <- srv.Start(ctx)
	}()

	// Wait for interrupt signal.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "❌ Server error: %v\n", err)

			return 1
		}
	case sig := <-sigChan:
		fmt.Printf("\n⏹️  Received signal %v, shutting down gracefully...\n", sig)
	}

	fmt.Println("✅ cipher-im service stopped")

	return 0
}

// imServiceClient implements the client subcommand.
// CLI wrapper for client operations.
func imServiceClient(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, IMUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the IM service")

	return 1
}

// imServiceInit implements the init subcommand.
// CLI wrapper for database and configuration initialization.
func imServiceInit(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, IMUsageInit)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Init subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will initialize database schema and configuration")

	return 1
}

// imServiceHealth implements the health subcommand.
// CLI wrapper calling the public health check API.
func imServiceHealth(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, IMUsageHealth)

		return 0
	}

	// Parse flags.
	url := defaultHealthURL
	cacertPath := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case urlFlag:
			if i+1 < len(args) && url == defaultHealthURL { // Only set if not already set
				baseURL := args[i+1]
				if !strings.HasSuffix(baseURL, "/health") {
					url = baseURL + "/health"
				} else {
					url = baseURL
				}

				i++ // Skip next arg
			}
		case cacertFlag:
			if i+1 < len(args) && cacertPath == "" { // Only set if not already set
				cacertPath = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Call health endpoint.
	statusCode, body, err := httpGet(url, cacertPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Health check failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK {
		_, _ = fmt.Fprintf(stdout, "✅ Service is healthy (HTTP %d)\n", statusCode)

		if body != "" {
			fmt.Fprintln(stdout, body)
		}

		return 0
	}

	fmt.Fprintf(stderr, "❌ Service is unhealthy (HTTP %d)\n", statusCode)

	if body != "" {
		fmt.Fprintln(stderr, body)
	}

	return 1
}

// imServiceLivez implements the livez subcommand.
// CLI wrapper calling the admin liveness check API.
func imServiceLivez(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		fmt.Fprintln(stderr, IMUsageLivez)

		return 0
	}

	// Parse flags.
	url := defaultLivezURL

	var cacertPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case urlFlag:
			if i+1 < len(args) && url == defaultLivezURL { // Only set if not already set
				baseURL := args[i+1]

				livezPath := cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminLivezRequestPath
				if !strings.HasSuffix(baseURL, livezPath) {
					url = baseURL + livezPath
				} else {
					url = baseURL
				}

				i++ // Skip next arg
			}
		case cacertFlag:
			if i+1 < len(args) && cacertPath == "" { // Only set if not already set
				cacertPath = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Call livez endpoint.
	statusCode, body, err := httpGet(url, cacertPath)
	if err != nil {
		fmt.Fprintf(stderr, "❌ Liveness check failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK {
		fmt.Fprintf(stdout, "✅ Service is alive (HTTP %d)\n", statusCode)

		if body != "" {
			fmt.Fprintln(stdout, body)
		}

		return 0
	}

	fmt.Fprintf(stderr, "❌ Service is not alive (HTTP %d)\n", statusCode)

	if body != "" {
		fmt.Fprintln(stderr, body)
	}

	return 1
}

// imServiceReadyz implements the readyz subcommand.
// CLI wrapper calling the admin readiness check API.
func imServiceReadyz(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		fmt.Fprintln(stderr, IMUsageReadyz)

		return 0
	}

	// Parse flags.
	url := defaultReadyzURL

	var cacertPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case urlFlag:
			if i+1 < len(args) && url == defaultReadyzURL { // Only set if not already set
				baseURL := args[i+1]

				readyzPath := cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminReadyzRequestPath
				if !strings.HasSuffix(baseURL, readyzPath) {
					url = baseURL + readyzPath
				} else {
					url = baseURL
				}

				i++ // Skip next arg
			}
		case cacertFlag:
			if i+1 < len(args) && cacertPath == "" { // Only set if not already set
				cacertPath = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Call readyz endpoint.
	statusCode, body, err := httpGet(url, cacertPath)
	if err != nil {
		fmt.Fprintf(stderr, "❌ Readiness check failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK {
		fmt.Fprintf(stdout, "✅ Service is ready (HTTP %d)\n", statusCode)

		if body != "" {
			fmt.Fprintln(stdout, body)
		}

		return 0
	}

	fmt.Fprintf(stderr, "❌ Service is not ready (HTTP %d)\n", statusCode)

	if body != "" {
		fmt.Fprintln(stderr, body)
	}

	return 1
}

// imServiceShutdown implements the shutdown subcommand.
// CLI wrapper calling the admin graceful shutdown API.
func imServiceShutdown(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		fmt.Fprintln(stderr, IMUsageShutdown)

		return 0
	}

	// Parse flags.
	url := defaultShutdownURL

	var cacertPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case urlFlag:
			if i+1 < len(args) && url == defaultShutdownURL { // Only set if not already set
				baseURL := args[i+1]

				shutdownPath := cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminShutdownRequestPath
				if !strings.HasSuffix(baseURL, shutdownPath) {
					url = baseURL + shutdownPath
				} else {
					url = baseURL
				}

				i++ // Skip next arg
			}
		case cacertFlag:
			if i+1 < len(args) && cacertPath == "" { // Only set if not already set
				cacertPath = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Call shutdown endpoint.
	statusCode, body, err := httpPost(url, cacertPath)
	if err != nil {
		fmt.Fprintf(stderr, "❌ Shutdown request failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK || statusCode == http.StatusAccepted {
		fmt.Fprintf(stdout, "✅ Shutdown initiated (HTTP %d)\n", statusCode)

		if body != "" {
			fmt.Fprintln(stdout, body)
		}

		return 0
	}

	fmt.Fprintf(stderr, "❌ Shutdown request failed (HTTP %d)\n", statusCode)

	if body != "" {
		fmt.Fprintln(stderr, body)
	}

	return 1
}

// loadCACertPool loads a CA certificate from file and returns an x509.CertPool.
func loadCACertPool(cacertPath string) (*x509.CertPool, error) {
	if cacertPath == "" {
		return nil, nil //nolint:nilnil // Valid pattern: no CA cert specified means use system defaults
	}

	// Read CA certificate file.
	caCertPEM, err := os.ReadFile(cacertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
	}

	// Create certificate pool.
	caCertPool := x509.NewCertPool()

	// Parse and add certificates to pool.
	for {
		block, rest := pem.Decode(caCertPEM)
		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
			}

			caCertPool.AddCert(cert)
		}

		caCertPEM = rest
	}

	if len(caCertPool.Subjects()) == 0 { //nolint:staticcheck // Subjects() is safe for manually created CertPools
		return nil, fmt.Errorf("no CA certificates found in file")
	}

	return caCertPool, nil
}

// httpGet performs an HTTP GET request with optional CA certificate validation.
// Used by health check CLI wrappers to call API endpoints.
func httpGet(url, cacertPath string) (int, string, error) {
	// Load CA certificate pool if specified.
	caCertPool, err := loadCACertPool(cacertPath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to load CA certificate: %w", err)
	}

	// Create HTTP client with proper TLS configuration.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				RootCAs:            caCertPool,        // Use CA cert pool if provided, nil = system defaults
				InsecureSkipVerify: caCertPool == nil, // Skip verification if no CA cert provided (backward compatibility)
			},
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("HTTP GET failed: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, string(body), nil
}

// httpPost performs an HTTP POST request with optional CA certificate validation.
// Used by shutdown CLI wrapper to call admin API endpoint.
func httpPost(url, cacertPath string) (int, string, error) {
	// Load CA certificate pool if specified.
	caCertPool, err := loadCACertPool(cacertPath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to load CA certificate: %w", err)
	}

	// Create HTTP client with proper TLS configuration.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				RootCAs:            caCertPool,        // Use CA cert pool if provided, nil = system defaults
				InsecureSkipVerify: caCertPool == nil, // Skip verification if no CA cert provided (backward compatibility)
			},
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("HTTP POST failed: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, string(body), nil
}

// initDatabase initializes database (PostgreSQL or SQLite) with schema.
// Database type determined by --database-url flag or DATABASE_URL env var.
// SQLite: sqliteInMemoryURL or file:/path/to/data.db?cache=shared
// PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
func initDatabase(ctx context.Context, databaseURL string) (*gorm.DB, error) {
	// Use provided database URL, or fall back to environment variable or default.
	if databaseURL == "" {
		if envURL := os.Getenv("DATABASE_URL"); envURL != "" {
			databaseURL = envURL
		} else {
			databaseURL = sqliteInMemoryURL // Default: SQLite in-memory
		}
	}

	// Detect database type from URL scheme.
	var (
		db  *gorm.DB
		err error
	)

	switch {
	case strings.HasPrefix(databaseURL, "postgres://"):
		db, err = initPostgreSQL(ctx, databaseURL)
	case strings.HasPrefix(databaseURL, "file:"):
		db, err = initSQLite(ctx, databaseURL)
	default:
		return nil, fmt.Errorf("unsupported database URL scheme: %s", databaseURL)
	}

	if err != nil {
		return nil, err
	}

	// Auto-migrate schema.
	if err := db.WithContext(ctx).AutoMigrate(
		&domain.User{},
		&domain.Message{},
		&cryptoutilTemplateRepository.BrowserSessionJWK{},
		&cryptoutilTemplateRepository.ServiceSessionJWK{},
		&cryptoutilTemplateRepository.BrowserSession{},
		&cryptoutilTemplateRepository.ServiceSession{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return db, nil
}

// initPostgreSQL initializes PostgreSQL database connection.
func initPostgreSQL(ctx context.Context, databaseURL string) (*gorm.DB, error) {
	// Open PostgreSQL database using pgx driver.
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Verify connection.
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	// Create GORM instance.
	dialector := postgres.New(postgres.Config{
		Conn: sqlDB,
	})

	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM for PostgreSQL: %w", err)
	}

	// Configure connection pool.
	sqlDB, err = db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilMagic.PostgreSQLMaxOpenConns)       // 25
	sqlDB.SetMaxIdleConns(cryptoutilMagic.PostgreSQLMaxIdleConns)       // 10
	sqlDB.SetConnMaxLifetime(cryptoutilMagic.PostgreSQLConnMaxLifetime) // 1 hour

	return db, nil
}

// initSQLite initializes SQLite database connection.
func initSQLite(ctx context.Context, databaseURL string) (*gorm.DB, error) {
	// Open SQLite database.
	sqlDB, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Enable WAL mode for concurrent operations.
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Set busy timeout for concurrent write operations.
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Create GORM instance.
	dialector := sqlite.Dialector{Conn: sqlDB}

	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM for SQLite: %w", err)
	}

	// Configure connection pool for GORM transactions.
	sqlDB, err = db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections) // 5
	sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections) // 5
	sqlDB.SetConnMaxLifetime(0)                                     // In-memory: never close

	return db, nil
}

// determineDatabaseType determines the database type from a GORM instance.
// Inspects the GORM dialector to determine if it's PostgreSQL or SQLite.
func determineDatabaseType(db *gorm.DB) repository.DatabaseType {
	dialectName := db.Name()

	switch dialectName {
	case dialectPostgres, dialectPostgresPG:
		return repository.DatabaseTypePostgreSQL
	case dialectSQLite:
		return repository.DatabaseTypeSQLite
	default:
		// Fallback to SQLite for unknown dialectors (safety default).
		return repository.DatabaseTypeSQLite
	}
}
