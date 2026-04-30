// Copyright (c) 2025-2026 Justin Cranford.
//
// TEMPLATE: Copy and rename 'skeleton' → your-service-name before use.

// Package template provides the Skeleton Template service entry point.
package template

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/pflag"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilLifecycle "cryptoutil/internal/apps-framework/service/lifecycle"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsSkeletonTemplateServer "cryptoutil/internal/apps/skeleton-template/server"
	cryptoutilAppsSkeletonTemplateServerConfig "cryptoutil/internal/apps/skeleton-template/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Template implements the Skeleton Template service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Template(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.SkeletonTemplateServiceID,
			ProductName:       cryptoutilSharedMagic.SkeletonProductName,
			ServiceName:       cryptoutilSharedMagic.SkeletonTemplateServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.SkeletonTemplateServicePort),
			UsageMain:         TemplateUsageMain,
			UsageServer:       TemplateUsageServer,
			UsageClient:       TemplateUsageClient,
			UsageInit:         TemplateUsageInit,
			UsageHealth:       TemplateUsageHealth,
			UsageLivez:        TemplateUsageLivez,
			UsageReadyz:       TemplateUsageReadyz,
			UsageShutdown:     TemplateUsageShutdown,
		},
		args, stdout, stderr,
		templateServerStart,
		templateClient,
		templateInit,
	)
}

// templateServerStart implements the server subcommand.
func templateServerStart(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, TemplateUsageServer)

		return 0
	}

	ctx := context.Background()

	// Parse configuration using ParseWithFlagSet with a fresh FlagSet.
	// Uses ContinueOnError for proper error handling (no os.Exit on bad flags).
	// Note: We prepend "start" as the subcommand for Parse() to validate.
	argsWithSubcommand := append([]string{"start"}, args...)

	fs := pflag.NewFlagSet("skeleton-template-server", pflag.ContinueOnError)

	cfg, err := cryptoutilAppsSkeletonTemplateServerConfig.ParseWithFlagSet(fs, argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilAppsSkeletonTemplateServer.NewFromConfig(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	// Mark server as ready after successful initialization.
	// This enables /admin/api/v1/readyz to return 200 OK instead of 503 Service Unavailable.
	srv.SetReady(true)

	_, _ = fmt.Fprintf(stdout, "🚀 Starting skeleton-template service...\n")
	_, _ = fmt.Fprintf(stdout, "   Public Server: https://%s:%d\n", cfg.BindPublicAddress, cfg.BindPublicPort)
	_, _ = fmt.Fprintf(stdout, "   Admin Server:  https://%s:%d\n", cfg.BindPrivateAddress, cfg.BindPrivatePort)

	exitCode := cryptoutilLifecycle.RunService(ctx, stdout, stderr, srv)

	_, _ = fmt.Fprintln(stdout, "✅ skeleton-template service stopped")

	return exitCode
}

// templateClient implements the client subcommand.
// CLI wrapper for client operations.
func templateClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, TemplateUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the Skeleton Template service")

	return 1
}

// templateInit implements the init subcommand.
// Generates PKI certificates for skeleton-template TLS endpoints via the framework PKI init.
func templateInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, TemplateUsageInit)

		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.SkeletonTemplateServiceID, args, stdout, stderr)
}
