// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
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
	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/skeleton-template/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/skeleton-template/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// TemplateUsageMain is the main usage message for the skeleton-template command.
	TemplateUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
		cryptoutilSharedMagic.TemplateDisplayName,
	)

	// TemplateUsageServer is the usage message for the server subcommand.
	TemplateUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
		cryptoutilSharedMagic.TemplateDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.SkeletonTemplateServiceID, cryptoutilSharedMagic.SkeletonTemplateServiceID),
	)

	// TemplateUsageClient is the usage message for the client subcommand.
	TemplateUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
		cryptoutilSharedMagic.TemplateDisplayName,
	)

	// TemplateUsageInit is the usage message for the init subcommand.
	TemplateUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
		cryptoutilSharedMagic.TemplateDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.SkeletonTemplateServiceID, cryptoutilSharedMagic.SkeletonTemplateServiceID),
	)

	// TemplateUsageHealth is the usage message for the health subcommand.
	TemplateUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.SkeletonTemplateServicePort),
	)

	// TemplateUsageLivez is the usage message for the livez subcommand.
	TemplateUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
	)

	// TemplateUsageReadyz is the usage message for the readyz subcommand.
	TemplateUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
	)

	// TemplateUsageShutdown is the usage message for the shutdown subcommand.
	TemplateUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
	)
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
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.SkeletonTemplateServiceID, UsageText: TemplateUsageClient}) {
		return 0
	}

	return 1
}

// templateInit implements the init subcommand.
// Generates PKI certificates for skeleton-template TLS endpoints via the framework PKI init.
func templateInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: TemplateUsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.SkeletonTemplateServiceID, args, stdout, stderr)
}
