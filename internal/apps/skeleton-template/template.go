// Copyright (c) 2025-2026 Justin Cranford.
//
// TEMPLATE: Copy and rename 'skeleton' → your-service-name before use.

// Package template provides the Skeleton Template service entry point.
package template

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
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
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsSkeletonTemplateServerConfig.SkeletonTemplateServerSettings]{
			UsageServer:  TemplateUsageServer,
			ServiceLabel: cryptoutilSharedMagic.SkeletonTemplateServiceID,
			FlagSetName:  "skeleton-template-server",
			ParseConfig:  cryptoutilAppsSkeletonTemplateServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsSkeletonTemplateServerConfig.SkeletonTemplateServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsSkeletonTemplateServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsSkeletonTemplateServerConfig.SkeletonTemplateServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
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
