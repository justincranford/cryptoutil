// Copyright (c) 2025 Justin Cranford
//
//

// Package idp provides the Identity Provider service entry point.
package idp

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilAppsIdentityIdpServer "cryptoutil/internal/apps/identity/idp/server"
	cryptoutilAppsIdentityIdpServerConfig "cryptoutil/internal/apps/identity/idp/server/config"
	cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	helpCommand   = "help"
	helpFlag      = "--help"
	helpShortFlag = "-h"
)

// Idp implements the Identity Provider service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Idp(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         "identity-idp",
			ProductName:       "identity",
			ServiceName:       "idp",
			DefaultPublicPort: cryptoutilSharedMagic.IdentityIDPServicePort,
			UsageMain:         IDPUsageMain,
			UsageServer:       IDPUsageServer,
			UsageClient:       IDPUsageClient,
			UsageInit:         IDPUsageInit,
			UsageHealth:       IDPUsageHealth,
			UsageLivez:        IDPUsageLivez,
			UsageReadyz:       IDPUsageReadyz,
			UsageShutdown:     IDPUsageShutdown,
		},
		args, stdout, stderr,
		idpServerStart,
		idpClient,
		idpServiceInit,
	)
}

// idpServerStart implements the server subcommand.
func idpServerStart(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, IDPUsageServer)

		return 0
	}

	ctx := context.Background()

	// Parse configuration using config.Parse() which leverages viper+pflag.
	// Note: We prepend "start" as the subcommand for Parse() to validate.
	argsWithSubcommand := append([]string{"start"}, args...)

	cfg, err := cryptoutilAppsIdentityIdpServerConfig.Parse(argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "\u274c Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilAppsIdentityIdpServer.NewFromConfig(ctx, cfg)
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
		_, _ = fmt.Fprintf(stdout, "\U0001f680 Starting identity-idp service...\n")
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

	fmt.Println("\u2705 identity-idp service stopped")

	return 0
}

// idpClient implements the client subcommand.
// CLI wrapper for client operations.
func idpClient(args []string, _, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, IDPUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "\u274c Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the Identity Provider service")

	return 1
}

// idpServiceInit implements the init subcommand.
// CLI wrapper for database and configuration initialization.
func idpServiceInit(args []string, _, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, IDPUsageInit)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "\u274c Init subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will initialize database schema and configuration")

	return 1
}
