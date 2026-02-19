// Copyright (c) 2025 Justin Cranford
//
//

// Package authz provides the Authorization Server service entry point.
package authz

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilAppsIdentityAuthzServer "cryptoutil/internal/apps/identity/authz/server"
	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity/authz/server/config"
	cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)


// Authz implements the Authorization Server service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Authz(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:        cryptoutilSharedMagic.IdentityAuthzServiceID,
			ProductName:      cryptoutilSharedMagic.IdentityProductName,
			ServiceName:      cryptoutilSharedMagic.AuthzServiceName,
			DefaultPublicPort: cryptoutilSharedMagic.IdentityAuthzServicePort,
			UsageMain:        AUTHZUsageMain,
			UsageServer:      AUTHZUsageServer,
			UsageClient:      AUTHZUsageClient,
			UsageInit:        AUTHZUsageInit,
			UsageHealth:      AUTHZUsageHealth,
			UsageLivez:       AUTHZUsageLivez,
			UsageReadyz:      AUTHZUsageReadyz,
			UsageShutdown:    AUTHZUsageShutdown,
		},
		args, stdout, stderr,
		authzServerStart,
		authzClient,
		authzServiceInit,
	)
}

// authzServerStart implements the server subcommand.
func authzServerStart(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, AUTHZUsageServer)

		return 0
	}

	ctx := context.Background()

	// Parse configuration using config.Parse() which leverages viper+pflag.
	// Note: We prepend "start" as the subcommand for Parse() to validate.
	argsWithSubcommand := append([]string{"start"}, args...)

	cfg, err := cryptoutilAppsIdentityAuthzServerConfig.Parse(argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilAppsIdentityAuthzServer.NewFromConfig(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	// Mark server as ready after successful initialization.
	// This enables /admin/api/v1/readyz to return 200 OK instead of 503 Service Unavailable.
	srv.SetReady(true)

	// Start server with graceful shutdown.
	errChan := make(chan error, 1)

	go func() {
		_, _ = fmt.Fprintf(stdout, "🚀 Starting identity-authz service...\n")
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

	fmt.Println("✅ identity-authz service stopped")

	return 0
}

// authzClient implements the client subcommand.
// CLI wrapper for client operations.
func authzClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, AUTHZUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the Authorization Server service")

	return 1
}

// authzServiceInit implements the init subcommand.
// CLI wrapper for database and configuration initialization.
func authzServiceInit(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, AUTHZUsageInit)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Init subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will initialize database schema and configuration")

	return 1
}
