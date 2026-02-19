// Copyright (c) 2025 Justin Cranford
//
//

// Package rp provides the Relying Party service entry point.
package rp

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilAppsIdentityRpServer "cryptoutil/internal/apps/identity/rp/server"
	cryptoutilAppsIdentityRpServerConfig "cryptoutil/internal/apps/identity/rp/server/config"
	cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	helpCommand   = "help"
	helpFlag      = "--help"
	helpShortFlag = "-h"
)

// Rp implements the Relying Party service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Rp(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         "identity-rp",
			ProductName:       "identity",
			ServiceName:       "rp",
			DefaultPublicPort: cryptoutilSharedMagic.IdentityRPServicePort,
			UsageMain:         RPUsageMain,
			UsageServer:       RPUsageServer,
			UsageClient:       RPUsageClient,
			UsageInit:         RPUsageInit,
			UsageHealth:       RPUsageHealth,
			UsageLivez:        RPUsageLivez,
			UsageReadyz:       RPUsageReadyz,
			UsageShutdown:     RPUsageShutdown,
		},
		args, stdout, stderr,
		rpServerStart,
		rpClient,
		rpServiceInit,
	)
}

// rpServerStart implements the server subcommand.
func rpServerStart(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, RPUsageServer)

		return 0
	}

	ctx := context.Background()

	// Parse configuration using config.Parse() which leverages viper+pflag.
	// Note: We prepend "start" as the subcommand for Parse() to validate.
	argsWithSubcommand := append([]string{"start"}, args...)

	cfg, err := cryptoutilAppsIdentityRpServerConfig.Parse(argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "\u274c Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilAppsIdentityRpServer.NewFromConfig(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "\u274c Failed to create server: %v\n", err)

		return 1
	}

	// Mark server as ready after successful initialization.
	// This enables /admin/api/v1/readyz to return 200 OK instead of 503 Service Unavailable.
	srv.SetReady(true)

	// Start server with graceful shutdown.
	errChan := make(chan error, 1)

	go func() {
		_, _ = fmt.Fprintf(stdout, "\U0001f680 Starting identity-rp service...\n")
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
			_, _ = fmt.Fprintf(stderr, "\u274c Server error: %v\n", err)

			return 1
		}
	case sig := <-sigChan:
		fmt.Printf("\n\u23f9\ufe0f  Received signal %v, shutting down gracefully...\n", sig)
	}

	fmt.Println("\u2705 identity-rp service stopped")

	return 0
}

// rpClient implements the client subcommand.
// CLI wrapper for client operations.
func rpClient(args []string, _, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, RPUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "\u274c Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the Relying Party service")

	return 1
}

// rpServiceInit implements the init subcommand.
// CLI wrapper for database and configuration initialization.
func rpServiceInit(args []string, _, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, RPUsageInit)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "\u274c Init subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will initialize database schema and configuration")

	return 1
}
