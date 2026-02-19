// Copyright (c) 2025 Justin Cranford
//
//

package im

import (
	"context"
	"fmt"
	"io"
	http "net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilAppsCipherImServer "cryptoutil/internal/apps/cipher/im/server"
	cryptoutilAppsCipherImServerConfig "cryptoutil/internal/apps/cipher/im/server/config"
)

const (
	helpCommand     = "help"
	helpFlag        = "--help"
	helpShortFlag   = "-h"
	urlFlag         = "--url"
	cacertFlag      = "--cacert"
	databaseURLFlag = "--database-url"

	// Default URLs for health check endpoints.
	defaultHealthURL   = "https://127.0.0.1:8070/health"
	defaultLivezURL    = "https://127.0.0.1:9090/admin/api/v1/livez"
	defaultReadyzURL   = "https://127.0.0.1:9090/admin/v1/readyz"
	defaultShutdownURL = "https://127.0.0.1:9090/admin/v1/shutdown"

	// SQLite in-memory database URL for shared cache.
	sqliteInMemoryURL = "file::memory:?cache=shared"
	dialectPostgres   = "postgres"
	dialectPostgresPG = "pgx"
)

// Im implements the instant messaging service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Im(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
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

	cfg, err := cryptoutilAppsCipherImServerConfig.Parse(argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilAppsCipherImServer.NewFromConfig(ctx, cfg)
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
		_, _ = fmt.Fprintf(stdout, "   Public Server: https://%s:%d\n", cfg.BindPublicAddress, cfg.BindPublicPort)
		_, _ = fmt.Fprintf(stdout, "   Admin Server:  https://%s:%d\n", cfg.BindPrivateAddress, cfg.BindPrivatePort)

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
func imServiceClient(args []string, _, stderr io.Writer) int {
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
func imServiceInit(args []string, _, stderr io.Writer) int {
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
			_, _ = fmt.Fprintln(stdout, body)
		}

		return 0
	}

	_, _ = fmt.Fprintf(stderr, "❌ Service is unhealthy (HTTP %d)\n", statusCode)

	if body != "" {
		_, _ = fmt.Fprintln(stderr, body)
	}

	return 1
}

// imServiceLivez implements the livez subcommand.
// CLI wrapper calling the admin liveness check API.
