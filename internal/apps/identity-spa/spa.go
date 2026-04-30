// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package spa provides the Single Page Application service entry point.
package spa

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	"github.com/spf13/pflag"

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilLifecycle "cryptoutil/internal/apps-framework/service/lifecycle"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsIdentitySpaServer "cryptoutil/internal/apps/identity-spa/server"
	cryptoutilAppsIdentitySpaServerConfig "cryptoutil/internal/apps/identity-spa/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Spa implements the Single Page Application service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Spa(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.IdentitySPAServiceID,
			ProductName:       cryptoutilSharedMagic.IdentityProductName,
			ServiceName:       cryptoutilSharedMagic.SPAServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.IdentitySPAServicePort),
			UsageMain:         SPAUsageMain,
			UsageServer:       SPAUsageServer,
			UsageClient:       SPAUsageClient,
			UsageInit:         SPAUsageInit,
			UsageHealth:       SPAUsageHealth,
			UsageLivez:        SPAUsageLivez,
			UsageReadyz:       SPAUsageReadyz,
			UsageShutdown:     SPAUsageShutdown,
		},
		args, stdout, stderr,
		spaServerStart,
		spaClient,
		spaServiceInit,
	)
}

// spaServerStart implements the server subcommand.
func spaServerStart(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, SPAUsageServer)

		return 0
	}

	ctx := context.Background()

	argsWithSubcommand := append([]string{"start"}, args...)

	fs := pflag.NewFlagSet("identity-spa-server", pflag.ContinueOnError)

	cfg, err := cryptoutilAppsIdentitySpaServerConfig.ParseWithFlagSet(fs, argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilAppsIdentitySpaServer.NewFromConfig(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	srv.SetReady(true)

	_, _ = fmt.Fprintf(stdout, "🚀 Starting identity-spa service...\n")
	_, _ = fmt.Fprintf(stdout, "   Public Server: https://%s:%d\n", cfg.BindPublicAddress, cfg.BindPublicPort)
	_, _ = fmt.Fprintf(stdout, "   Admin Server:  https://%s:%d\n", cfg.BindPrivateAddress, cfg.BindPrivatePort)

	exitCode := cryptoutilLifecycle.RunService(ctx, stdout, stderr, srv)

	_, _ = fmt.Fprintln(stdout, "✅ identity-spa service stopped")

	return exitCode
}

// spaClient implements the client subcommand.
// CLI wrapper for client operations.
func spaClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, SPAUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the Single Page Application service")

	return 1
}

// spaServiceInit implements the init subcommand.
// Generates PKI certificates for identity-spa TLS endpoints via the framework PKI init.
func spaServiceInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, SPAUsageInit)

		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentitySPAServiceID, args, stdout, stderr)
}
