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
	dialectPostgres   = "postgres"
	dialectPostgresPG = "pgx"
)

// IM implements the instant messaging service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func IM(args []string) int {
	// Default to "server" subcommand if no args provided (backward compatibility).
	if len(args) == 0 {
		args = []string{"server"}
	}

	// Check for help flags.
	if args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag {
		printIMUsage()

		return 0
	}

	// Route to subcommand.
	switch args[0] {
	case "version":
		printIMVersion()

		return 0
	case "server":
		return imServer(args[1:])
	case "client":
		return imClient(args[1:])
	case "init":
		return imInit(args[1:])
	case "health":
		return imHealth(args[1:])
	case "livez":
		return imLivez(args[1:])
	case "readyz":
		return imReadyz(args[1:])
	case "shutdown":
		return imShutdown(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", args[0])
		printIMUsage()

		return 1
	}
}

// printIMVersion prints the instant messaging service version information.
func printIMVersion() {
	fmt.Println("cipher-im service")
	fmt.Println("Part of cryptoutil cipher product")
	fmt.Println("Version information available via Docker image tags")
}

// printIMUsage prints the instant messaging service usage information.
func printIMUsage() {
	fmt.Fprintln(os.Stderr, `Usage: cipher im <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the instant messaging server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "learn im <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`)
}

// imServer implements the server subcommand.
func imServer(args []string) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		fmt.Fprintln(os.Stderr, `Usage: cipher im server [options]

Description:
  Start the instant messaging server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL    Database URL (default: SQLite in-memory)
                        SQLite: file::memory:?cache=shared
                        PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --help, -h            Show this help message

Examples:
  learn im server
  learn im server --database-url file:/tmp/cipher.db
  learn im server --database-url postgres://user:pass@localhost:5432/cipher`)

		return 0
	}

	ctx := context.Background()

	// Parse flags.
	databaseURL := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case databaseURLFlag:
			if i+1 < len(args) && databaseURL == "" { // Only set if not already set
				databaseURL = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Initialize database (PostgreSQL or SQLite).
	db, err := initDatabase(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to initialize database: %v\n", err)

		return 1
	}

	// Create cipher-im server configuration using AppConfig.
	// AppConfig embeds ServerSettings and adds cipher-im-specific settings.
	cfg := config.DefaultAppConfig()
	cfg.BindPublicPort = cryptoutilMagic.DefaultPublicPortCipherIM
	cfg.BindPrivatePort = cryptoutilMagic.DefaultPrivatePortCipherIM
	cfg.OTLPService = "cipher-im"
	cfg.OTLPEnabled = false // Demo service uses in-process telemetry only.

	srv, err := server.New(ctx, cfg, db, determineDatabaseType(db))
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	// Mark server as ready after successful initialization.
	// This enables /admin/v1/readyz to return 200 OK instead of 503 Service Unavailable.
	srv.SetReady(true)

	// Start server with graceful shutdown.
	errChan := make(chan error, 1)

	go func() {
		fmt.Printf("🚀 Starting cipher-im service...\n")
		fmt.Printf("   Public Server: https://127.0.0.1:%d\n", cryptoutilMagic.DefaultPublicPortCipherIM)
		fmt.Printf("   Admin Server:  https://127.0.0.1:%d\n", cryptoutilMagic.DefaultPrivatePortCipherIM)

		errChan <- srv.Start(ctx)
	}()

	// Wait for interrupt signal.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Server error: %v\n", err)

			return 1
		}
	case sig := <-sigChan:
		fmt.Printf("\n⏹️  Received signal %v, shutting down gracefully...\n", sig)
	}

	fmt.Println("✅ cipher-im service stopped")

	return 0
}

// imClient implements the client subcommand.
// CLI wrapper for client operations.
func imClient(args []string) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		fmt.Fprintln(os.Stderr, `Usage: cipher im client [options]

Description:
  Run client operations for instant messaging service.

Options:
  --help, -h    Show this help message

Examples:
  learn im client`)

		return 0
	}

	fmt.Fprintln(os.Stderr, "❌ Client subcommand not yet implemented")
	fmt.Fprintln(os.Stderr, "   This will provide CLI tools for interacting with the IM service")

	return 1
}

// imInit implements the init subcommand.
// CLI wrapper for database and configuration initialization.
func imInit(args []string) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		fmt.Fprintln(os.Stderr, `Usage: cipher im init [options]

Description:
  Initialize database schema and configuration for instant messaging service.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  learn im init
  learn im init --config configs/learn/im/config.yml`)

		return 0
	}

	fmt.Fprintln(os.Stderr, "❌ Init subcommand not yet implemented")
	fmt.Fprintln(os.Stderr, "   This will initialize database schema and configuration")

	return 1
}

// imHealth implements the health subcommand.
// CLI wrapper calling the public health check API.
func imHealth(args []string) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		fmt.Fprintln(os.Stderr, `Usage: cipher im health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:8888)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  learn im health
  learn im health --url https://localhost:8888
  learn im health --cacert /path/to/ca.pem`)

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
		fmt.Fprintf(os.Stderr, "❌ Health check failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK {
		fmt.Printf("✅ Service is healthy (HTTP %d)\n", statusCode)

		if body != "" {
			fmt.Println(body)
		}

		return 0
	}

	fmt.Fprintf(os.Stderr, "❌ Service is unhealthy (HTTP %d)\n", statusCode)

	if body != "" {
		fmt.Fprintln(os.Stderr, body)
	}

	return 1
}

// imLivez implements the livez subcommand.
// CLI wrapper calling the admin liveness check API.
func imLivez(args []string) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		fmt.Fprintln(os.Stderr, `Usage: cipher im livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  learn im livez
  learn im livez --url https://localhost:9090`)

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
		fmt.Fprintf(os.Stderr, "❌ Liveness check failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK {
		fmt.Printf("✅ Service is alive (HTTP %d)\n", statusCode)

		if body != "" {
			fmt.Println(body)
		}

		return 0
	}

	fmt.Fprintf(os.Stderr, "❌ Service is not alive (HTTP %d)\n", statusCode)

	if body != "" {
		fmt.Fprintln(os.Stderr, body)
	}

	return 1
}

// imReadyz implements the readyz subcommand.
// CLI wrapper calling the admin readiness check API.
func imReadyz(args []string) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		fmt.Fprintln(os.Stderr, `Usage: cipher im readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  learn im readyz
  learn im readyz --url https://localhost:9090`)

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
		fmt.Fprintf(os.Stderr, "❌ Readiness check failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK {
		fmt.Printf("✅ Service is ready (HTTP %d)\n", statusCode)

		if body != "" {
			fmt.Println(body)
		}

		return 0
	}

	fmt.Fprintf(os.Stderr, "❌ Service is not ready (HTTP %d)\n", statusCode)

	if body != "" {
		fmt.Fprintln(os.Stderr, body)
	}

	return 1
}

// imShutdown implements the shutdown subcommand.
// CLI wrapper calling the admin graceful shutdown API.
func imShutdown(args []string) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		fmt.Fprintln(os.Stderr, `Usage: cipher im shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  learn im shutdown
  learn im shutdown --url https://localhost:9090
  learn im shutdown --force`)

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
		fmt.Fprintf(os.Stderr, "❌ Shutdown request failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK || statusCode == http.StatusAccepted {
		fmt.Printf("✅ Shutdown initiated (HTTP %d)\n", statusCode)

		if body != "" {
			fmt.Println(body)
		}

		return 0
	}

	fmt.Fprintf(os.Stderr, "❌ Shutdown request failed (HTTP %d)\n", statusCode)

	if body != "" {
		fmt.Fprintln(os.Stderr, body)
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
// SQLite: file::memory:?cache=shared or file:/path/to/data.db?cache=shared
// PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
func initDatabase(ctx context.Context, databaseURL string) (*gorm.DB, error) {
	// Use provided database URL, or fall back to environment variable or default.
	if databaseURL == "" {
		if envURL := os.Getenv("DATABASE_URL"); envURL != "" {
			databaseURL = envURL
		} else {
			databaseURL = "file::memory:?cache=shared" // Default: SQLite in-memory
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
	if err := db.WithContext(ctx).AutoMigrate(&domain.User{}, &domain.Message{}); err != nil {
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
