// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package authz provides the Authorization Server service entry point.
package authz

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
	cryptoutilAppsIdentityAuthzServer "cryptoutil/internal/apps/identity-authz/server"
	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Authz implements the Authorization Server service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Authz(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.IdentityAuthzServiceID,
			ProductName:       cryptoutilSharedMagic.IdentityProductName,
			ServiceName:       cryptoutilSharedMagic.AuthzServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.IdentityAuthzServicePort),
			UsageMain:         AUTHZUsageMain,
			UsageServer:       AUTHZUsageServer,
			UsageClient:       AUTHZUsageClient,
			UsageInit:         AUTHZUsageInit,
			UsageHealth:       AUTHZUsageHealth,
			UsageLivez:        AUTHZUsageLivez,
			UsageReadyz:       AUTHZUsageReadyz,
			UsageShutdown:     AUTHZUsageShutdown,
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

	// Parse configuration using ParseWithFlagSet with a fresh FlagSet.
	// Uses ContinueOnError for proper error handling (no os.Exit on bad flags).
	// Note: We prepend "start" as the subcommand for Parse() to validate.
	argsWithSubcommand := append([]string{"start"}, args...)

	fs := pflag.NewFlagSet("identity-authz-server", pflag.ContinueOnError)

	cfg, err := cryptoutilAppsIdentityAuthzServerConfig.ParseWithFlagSet(fs, argsWithSubcommand, true)
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

	_, _ = fmt.Fprintf(stdout, "🚀 Starting identity-authz service...\n")
	_, _ = fmt.Fprintf(stdout, "   Public Server: https://%s:%d\n", cfg.BindPublicAddress, cfg.BindPublicPort)
	_, _ = fmt.Fprintf(stdout, "   Admin Server:  https://%s:%d\n", cfg.BindPrivateAddress, cfg.BindPrivatePort)

	exitCode := cryptoutilLifecycle.RunService(ctx, stdout, stderr, srv)

	_, _ = fmt.Fprintln(stdout, "✅ identity-authz service stopped")

	return exitCode
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
// Generates PKI certificates for identity-authz TLS endpoints via the framework PKI init.
func authzServiceInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, AUTHZUsageInit)

		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentityAuthzServiceID, args, stdout, stderr)
}
