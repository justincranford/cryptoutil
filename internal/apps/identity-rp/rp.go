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

	"github.com/spf13/pflag"

	cryptoutilTemplateCli "cryptoutil/internal/apps/framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps/framework/tls"
	cryptoutilAppsIdentityRpServer "cryptoutil/internal/apps/identity-rp/server"
	cryptoutilAppsIdentityRpServerConfig "cryptoutil/internal/apps/identity-rp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Rp implements the Relying Party service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Rp(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.IdentityRPServiceID,
			ProductName:       cryptoutilSharedMagic.IdentityProductName,
			ServiceName:       cryptoutilSharedMagic.RPServiceName,
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
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, RPUsageServer)

		return 0
	}

	ctx := context.Background()

	argsWithSubcommand := append([]string{"start"}, args...)

	fs := pflag.NewFlagSet("identity-rp-server", pflag.ContinueOnError)

	cfg, err := cryptoutilAppsIdentityRpServerConfig.ParseWithFlagSet(fs, argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilAppsIdentityRpServer.NewFromConfig(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	srv.SetReady(true)

	errChan := make(chan error, 1)

	go func() {
		_, _ = fmt.Fprintf(stdout, "🚀 Starting identity-rp service...\n")
		_, _ = fmt.Fprintf(stdout, "   Public Server: https://%s:%d\n", cfg.BindPublicAddress, cfg.BindPublicPort)
		_, _ = fmt.Fprintf(stdout, "   Admin Server:  https://%s:%d\n", cfg.BindPrivateAddress, cfg.BindPrivatePort)

		errChan <- srv.Start(ctx)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "❌ Server error: %v\n", err)

			return 1
		}
	case sig := <-sigChan:
		_, _ = fmt.Fprintf(stdout, "\n⏹️  Received signal %v, shutting down gracefully...\n", sig)

		shutdownCtx, cancel := context.WithTimeout(ctx, cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
		defer cancel()

		if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
			_, _ = fmt.Fprintf(stderr, "⚠️  Shutdown error: %v\n", shutdownErr)
		}
	}

	signal.Stop(sigChan)

	_, _ = fmt.Fprintln(stdout, "✅ identity-rp service stopped")

	return 0
}

// rpClient implements the client subcommand.
// CLI wrapper for client operations.
func rpClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, RPUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the Relying Party service")

	return 1
}

// rpServiceInit implements the init subcommand.
// Generates PKI certificates for identity-rp TLS endpoints via the framework PKI init.
func rpServiceInit(args []string, stdout, stderr io.Writer) int {
	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentityRPServiceID, args, stdout, stderr)

}
