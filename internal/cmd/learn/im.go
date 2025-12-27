// Copyright (c) 2025 Justin Cranford
//
//

package learn

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

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
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printIMUsage()

		return 0
	}

	// Check for version flags.
	if args[0] == "version" || args[0] == "--version" || args[0] == "-v" {
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
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
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
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
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
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
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

	fmt.Fprintln(os.Stderr, "❌ Health subcommand not yet implemented")
	fmt.Fprintln(os.Stderr, "   This will call GET /health on the public server")
	fmt.Fprintln(os.Stderr, "   Example: curl -k https://127.0.0.1:8888/health")

	return 1
}

// imLivez implements the livez subcommand.
// CLI wrapper calling the admin liveness check API.
func imLivez(args []string) int {
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
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

	fmt.Fprintln(os.Stderr, "❌ Livez subcommand not yet implemented")
	fmt.Fprintln(os.Stderr, "   This will call GET /admin/v1/livez on the admin server")
	fmt.Fprintln(os.Stderr, "   Example: curl -k https://127.0.0.1:9090/admin/v1/livez")

	return 1
}

// imReadyz implements the readyz subcommand.
// CLI wrapper calling the admin readiness check API.
func imReadyz(args []string) int {
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
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

	fmt.Fprintln(os.Stderr, "❌ Readyz subcommand not yet implemented")
	fmt.Fprintln(os.Stderr, "   This will call GET /admin/v1/readyz on the admin server")
	fmt.Fprintln(os.Stderr, "   Example: curl -k https://127.0.0.1:9090/admin/v1/readyz")

	return 1
}

// imShutdown implements the shutdown subcommand.
// CLI wrapper calling the admin graceful shutdown API.
func imShutdown(args []string) int {
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
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

	fmt.Fprintln(os.Stderr, "❌ Shutdown subcommand not yet implemented")
	fmt.Fprintln(os.Stderr, "   This will call POST /admin/v1/shutdown on the admin server")
	fmt.Fprintln(os.Stderr, "   Example: curl -k -X POST https://127.0.0.1:9090/admin/v1/shutdown")

	return 1
}

// initDatabase initializes SQLite database with schema.
func initDatabase(ctx context.Context) (*gorm.DB, error) {
	// Open SQLite in-memory database.
	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
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
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	// Configure connection pool for GORM transactions.
	sqlDB, err = db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections) // 5
	sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections) // 5
	sqlDB.SetConnMaxLifetime(0)                                     // In-memory: never close

	// Auto-migrate schema.
	if err := db.WithContext(ctx).AutoMigrate(&domain.User{}, &domain.Message{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return db, nil
}
