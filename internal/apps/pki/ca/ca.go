// Copyright (c) 2025 Justin Cranford
//
//

// Package ca provides the Certificate Authority service entry point.
package ca

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilAppsCaServer "cryptoutil/internal/apps/pki/ca/server"
	cryptoutilAppsCaServerConfig "cryptoutil/internal/apps/pki/ca/server/config"
	cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)


// Ca implements the Certificate Authority service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Ca(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.PKICAServiceID,
			ProductName:       cryptoutilSharedMagic.PKIProductName,
			ServiceName:       cryptoutilSharedMagic.PKICAServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.PKICAServicePort),
			UsageMain:         CAUsageMain,
			UsageServer:       CAUsageServer,
			UsageClient:       CAUsageClient,
			UsageInit:         CAUsageInit,
			UsageHealth:       CAUsageHealth,
			UsageLivez:        CAUsageLivez,
			UsageReadyz:       CAUsageReadyz,
			UsageShutdown:     CAUsageShutdown,
		},
		args, stdout, stderr,
		caServerStart,
		caClient,
		caInit,
	)
}

// caServerStart implements the server subcommand.
func caServerStart(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, CAUsageServer)

		return 0
	}

	ctx := context.Background()

	// Parse configuration using ParseWithFlagSet with a fresh FlagSet.
	// Uses ContinueOnError for proper error handling (no os.Exit on bad flags).
	// Note: We prepend "start" as the subcommand for Parse() to validate.
	argsWithSubcommand := append([]string{"start"}, args...)

	fs := pflag.NewFlagSet("pki-ca-server", pflag.ContinueOnError)

	cfg, err := cryptoutilAppsCaServerConfig.ParseWithFlagSet(fs, argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilAppsCaServer.NewFromConfig(ctx, cfg)
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
		_, _ = fmt.Fprintf(stdout, "🚀 Starting pki-ca service...\n")
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

	fmt.Println("✅ pki-ca service stopped")

	return 0
}

// caClient implements the client subcommand.
// CLI wrapper for client operations.
func caClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, CAUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the CA service")

	return 1
}

// caInit implements the init subcommand.
// CLI wrapper for database and configuration initialization.
func caInit(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, CAUsageInit)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Init subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will initialize database schema and configuration")

	return 1
}
