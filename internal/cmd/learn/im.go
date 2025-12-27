// Copyright (c) 2025 Justin Cranford
//
//

package learn

import (
	"context"
	"crypto/tls"
	"database/sql"
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

	"cryptoutil/internal/learn/domain"
	"cryptoutil/internal/learn/server"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
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

	// Check for version flags.
	if args[0] == versionCommand || args[0] == versionFlag || args[0] == versionShortFlag {
		printIMVersion()

		return 0
	}

	// Route to subcommand.
	switch args[0] {
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

// printIMUsage prints the instant messaging service usage information.
func printIMUsage() {
	fmt.Fprintln(os.Stderr, `Usage: learn im <subcommand> [options]

Available subcommands:
  server      Start the instant messaging server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "learn im <subcommand> help" for subcommand-specific help.
Use "learn im version" for version information.`)
}

// printIMVersion prints the instant messaging service version information.
func printIMVersion() {
	// Version information should be injected from the calling binary.
	fmt.Println("learn-im service (cryptoutil learn product)")
}

// imServer implements the server subcommand.
func imServer(args []string) int {
	ctx := context.Background()

	// Initialize SQLite in-memory database for demonstration.
	db, err := initDatabase(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to initialize database: %v\n", err)

		return 1
	}

	// Create learn-im server.
	cfg := &server.Config{
		PublicPort: int(cryptoutilMagic.DefaultPublicPortLearnIM),
		AdminPort:  cryptoutilMagic.DefaultPrivatePortLearnIM,
		DB:         db,
		JWTSecret:  "learn-im-dev-secret-change-in-production", // TODO: Load from configuration file
	}

	srv, err := server.New(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	// Start server with graceful shutdown.
	errChan := make(chan error, 1)

	go func() {
		fmt.Printf("🚀 Starting learn-im service...\n")
		fmt.Printf("   Public Server: https://127.0.0.1:%d\n", cryptoutilMagic.DefaultPublicPortLearnIM)
		fmt.Printf("   Admin Server:  https://127.0.0.1:%d\n", cryptoutilMagic.DefaultPrivatePortLearnIM)

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

	fmt.Println("✅ learn-im service stopped")

	return 0
}

// imClient implements the client subcommand.
// CLI wrapper for client operations.
func imClient(args []string) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		fmt.Fprintln(os.Stderr, `Usage: learn im client [options]

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
		fmt.Fprintln(os.Stderr, `Usage: learn im init [options]

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
		fmt.Fprintln(os.Stderr, `Usage: learn im health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:8888)
  --help, -h     Show this help message

Examples:
  learn im health
  learn im health --url https://localhost:8888`)

		return 0
	}

	// Parse URL flag.
	url := "https://127.0.0.1:8888/health"

	for i := 0; i < len(args); i++ {
		if args[i] == urlFlag && i+1 < len(args) {
			baseURL := args[i+1]
			if !strings.HasSuffix(baseURL, "/health") {
				url = baseURL + "/health"
			} else {
				url = baseURL
			}

			break
		}
	}

	// Call health endpoint.
	statusCode, body, err := httpGet(url)
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
		fmt.Fprintln(os.Stderr, `Usage: learn im livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --help, -h     Show this help message

Examples:
  learn im livez
  learn im livez --url https://localhost:9090`)

		return 0
	}

	// Parse URL flag.
	url := "https://127.0.0.1:9090/admin/v1/livez"

	for i := 0; i < len(args); i++ {
		if args[i] == urlFlag && i+1 < len(args) {
			baseURL := args[i+1]
			if !strings.HasSuffix(baseURL, "/admin/v1/livez") {
				url = baseURL + "/admin/v1/livez"
			} else {
				url = baseURL
			}

			break
		}
	}

	// Call livez endpoint.
	statusCode, body, err := httpGet(url)
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
		fmt.Fprintln(os.Stderr, `Usage: learn im readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --help, -h     Show this help message

Examples:
  learn im readyz
  learn im readyz --url https://localhost:9090`)

		return 0
	}

	// Parse URL flag.
	url := "https://127.0.0.1:9090/admin/v1/readyz"

	for i := 0; i < len(args); i++ {
		if args[i] == urlFlag && i+1 < len(args) {
			baseURL := args[i+1]
			if !strings.HasSuffix(baseURL, "/admin/v1/readyz") {
				url = baseURL + "/admin/v1/readyz"
			} else {
				url = baseURL
			}

			break
		}
	}

	// Call readyz endpoint.
	statusCode, body, err := httpGet(url)
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
		fmt.Fprintln(os.Stderr, `Usage: learn im shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  learn im shutdown
  learn im shutdown --url https://localhost:9090
  learn im shutdown --force`)

		return 0
	}

	// Parse URL flag.
	url := "https://127.0.0.1:9090/admin/v1/shutdown"

	for i := 0; i < len(args); i++ {
		if args[i] == urlFlag && i+1 < len(args) {
			baseURL := args[i+1]
			if !strings.HasSuffix(baseURL, "/admin/v1/shutdown") {
				url = baseURL + "/admin/v1/shutdown"
			} else {
				url = baseURL
			}

			break
		}
	}

	// Call shutdown endpoint.
	statusCode, body, err := httpPost(url)
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

// httpGet performs an HTTP GET request with TLS certificate validation disabled.
// Used by health check CLI wrappers to call API endpoints.
func httpGet(url string) (int, string, error) {
	// Create HTTP client that accepts self-signed certificates.
	// TODO: Add proper certificate validation with --cacert flag.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Allow self-signed certificates for dev/testing.
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

// httpPost performs an HTTP POST request with TLS certificate validation disabled.
// Used by shutdown CLI wrapper to call admin API endpoint.
func httpPost(url string) (int, string, error) {
	// Create HTTP client that accepts self-signed certificates.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Allow self-signed certificates for dev/testing.
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
func initDatabase(ctx context.Context) (*gorm.DB, error) {
	// Determine database URL from flags or environment.
	// Priority: CLI flag > environment variable > default (SQLite in-memory).
	databaseURL := "file::memory:?cache=shared" // Default: SQLite in-memory

	// TODO: Parse --database-url flag when flag parsing is added.
	// For now, check environment variable.
	if envURL := os.Getenv("DATABASE_URL"); envURL != "" {
		databaseURL = envURL
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
	// Open PostgreSQL database.
	sqlDB, err := sql.Open("postgres", databaseURL)
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
