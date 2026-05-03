// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package template provides the skeleton-template service entry point.
package template

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/skeleton-template/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/skeleton-template/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Template implements the skeleton-template service subcommand handler.
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
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.SkeletonTemplateServerSettings]{
			UsageServer:  TemplateUsageServer,
			ServiceLabel: cryptoutilSharedMagic.SkeletonTemplateServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.SkeletonTemplateServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.SkeletonTemplateServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.SkeletonTemplateServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// templateClient implements the client subcommand.
// CLI wrapper for client operations.
func templateClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.SkeletonTemplateServiceID}) {
		_, _ = fmt.Fprintln(stderr, TemplateUsageClient)

		return 0
	}

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
