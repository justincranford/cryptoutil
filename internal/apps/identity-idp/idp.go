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

	"github.com/spf13/pflag"

	cryptoutilTemplateCli "cryptoutil/internal/apps/framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps/framework/tls"
	cryptoutilAppsIdentityIdpServer "cryptoutil/internal/apps/identity-idp/server"
	cryptoutilAppsIdentityIdpServerConfig "cryptoutil/internal/apps/identity-idp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Idp implements the Identity Provider service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Idp(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.IdentityIDPServiceID,
			ProductName:       cryptoutilSharedMagic.IdentityProductName,
			ServiceName:       cryptoutilSharedMagic.IDPServiceName,
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
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, IDPUsageServer)

		return 0
	}

	ctx := context.Background()

	argsWithSubcommand := append([]string{"start"}, args...)

	fs := pflag.NewFlagSet("identity-idp-server", pflag.ContinueOnError)

	cfg, err := cryptoutilAppsIdentityIdpServerConfig.ParseWithFlagSet(fs, argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilAppsIdentityIdpServer.NewFromConfig(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	srv.SetReady(true)

	errChan := make(chan error, 1)

	go func() {
		_, _ = fmt.Fprintf(stdout, "🚀 Starting identity-idp service...\n")
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

	_, _ = fmt.Fprintln(stdout, "✅ identity-idp service stopped")

	return 0
}

// idpClient implements the client subcommand.
// CLI wrapper for client operations.
func idpClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, IDPUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the Identity Provider service")

	return 1
}

// idpServiceInit implements the init subcommand.
// Generates PKI certificates for identity-idp TLS endpoints via the framework PKI init.
func idpServiceInit(args []string, stdout, stderr io.Writer) int {
	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentityIDPServiceID, args, stdout, stderr)
}
