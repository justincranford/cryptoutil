// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package im provides the Instant Messaging Service entry point.
package im

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkServiceLifecycle "cryptoutil/internal/apps-framework/service/lifecycle"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsSmImServer "cryptoutil/internal/apps/sm-im/server"
	cryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm-im/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
)

// Im implements the instant messaging service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Im(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.IMServiceID,
			ProductName:       cryptoutilSharedMagic.IMProductName,
			ServiceName:       cryptoutilSharedMagic.IMServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.IMServicePort),
			UsageMain:         IMUsageMain,
			UsageServer:       IMUsageServer,
			UsageClient:       IMUsageClient,
			UsageInit:         IMUsageInit,
			UsageHealth:       IMUsageHealth,
			UsageLivez:        IMUsageLivez,
			UsageReadyz:       IMUsageReadyz,
			UsageShutdown:     IMUsageShutdown,
		},
		args, stdout, stderr,
		imServerStart,
		imClient,
		imInit,
	)
}

// imServerStart implements the server subcommand.
func imServerStart(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, IMUsageServer)

		return 0
	}

	ctx := context.Background()

	// Parse configuration using config.Parse() which leverages viper+pflag.
	// Note: We prepend "start" as the subcommand for Parse() to validate.
	argsWithSubcommand := append([]string{"start"}, args...)

	fs := pflag.NewFlagSet("sm-im-server", pflag.ContinueOnError)

	cfg, err := cryptoutilAppsSmImServerConfig.ParseWithFlagSet(fs, argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilAppsSmImServer.NewIMServerFromConfig(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	// Mark server as ready so /admin/api/v1/readyz return 200 OK instead of 503 Service Unavailable.
	srv.SetReady(true)

	_, _ = fmt.Fprintf(stdout, "🚀 Starting sm-im service...\n")
	_, _ = fmt.Fprintf(stdout, "   Public Server: https://%s:%d\n", cfg.BindPublicAddress, cfg.BindPublicPort)
	_, _ = fmt.Fprintf(stdout, "   Admin Server:  https://%s:%d\n", cfg.BindPrivateAddress, cfg.BindPrivatePort)

	exitCode := cryptoutilAppsFrameworkServiceLifecycle.RunService(ctx, stdout, stderr, srv)

	_, _ = fmt.Fprintln(stdout, "✅ sm-im service stopped")

	return exitCode
}

// imClient implements the client subcommand.
// CLI wrapper for client operations.
func imClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, IMUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the IM service")

	return 1
}

// imInit implements the init subcommand.
// Generates PKI certificates for sm-im TLS endpoints via the framework PKI init.
func imInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, IMUsageInit)

		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IMServiceID, args, stdout, stderr)
}
