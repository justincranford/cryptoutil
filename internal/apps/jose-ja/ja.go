// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package ja provides the jose-ja service entry point.
package ja

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/jose-ja/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/jose-ja/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// JAUsageMain is the main usage message for the jose-ja command.
	JAUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
		cryptoutilSharedMagic.JoseJADisplayName,
	)

	// JAUsageServer is the usage message for the server subcommand.
	JAUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
		cryptoutilSharedMagic.JoseJADisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.JoseJAServiceID, cryptoutilSharedMagic.JoseJAServiceID),
	)

	// JAUsageClient is the usage message for the client subcommand.
	JAUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
		cryptoutilSharedMagic.JoseJADisplayName,
	)

	// JAUsageInit is the usage message for the init subcommand.
	JAUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
		cryptoutilSharedMagic.JoseJADisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.JoseJAServiceID, cryptoutilSharedMagic.JoseJAServiceID),
	)

	// JAUsageHealth is the usage message for the health subcommand.
	JAUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.JoseJAServicePort),
	)

	// JAUsageLivez is the usage message for the livez subcommand.
	JAUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
	)

	// JAUsageReadyz is the usage message for the readyz subcommand.
	JAUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
	)

	// JAUsageShutdown is the usage message for the shutdown subcommand.
	JAUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
	)
)

// Ja implements the jose-ja service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Ja(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.JoseJAServiceID,
			ProductName:       cryptoutilSharedMagic.JoseProductName,
			ServiceName:       cryptoutilSharedMagic.JoseJAServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.JoseJAServicePort),
			UsageMain:         JAUsageMain,
			UsageServer:       JAUsageServer,
			UsageClient:       JAUsageClient,
			UsageInit:         JAUsageInit,
			UsageHealth:       JAUsageHealth,
			UsageLivez:        JAUsageLivez,
			UsageReadyz:       JAUsageReadyz,
			UsageShutdown:     JAUsageShutdown,
		},
		args, stdout, stderr,
		jaServerStart,
		jaClient,
		jaInit,
	)
}

// jaServerStart implements the server subcommand.
func jaServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.JoseJAServerSettings]{
			UsageServer:  JAUsageServer,
			ServiceLabel: cryptoutilSharedMagic.JoseJAServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.JoseJAServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.JoseJAServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.JoseJAServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// jaClient implements the client subcommand.
// CLI wrapper for client operations.
func jaClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.JoseJAServiceID, UsageText: JAUsageClient}) {
		return 0
	}

	return 1
}

// jaInit implements the init subcommand.
// Generates PKI certificates for jose-ja TLS endpoints via the framework PKI init.
func jaInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: JAUsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.JoseJAServiceID, args, stdout, stderr)
}
