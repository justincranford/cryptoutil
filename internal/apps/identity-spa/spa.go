// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package spa provides the identity-spa service entry point.
package spa

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/identity-spa/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/identity-spa/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// SPAUsageMain is the main usage message for the identity-spa command.
	SPAUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
		cryptoutilSharedMagic.SPADisplayName,
	)

	// SPAUsageServer is the usage message for the server subcommand.
	SPAUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
		cryptoutilSharedMagic.SPADisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentitySPAServiceID, cryptoutilSharedMagic.IdentitySPAServiceID),
	)

	// SPAUsageClient is the usage message for the client subcommand.
	SPAUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
		cryptoutilSharedMagic.SPADisplayName,
	)

	// SPAUsageInit is the usage message for the init subcommand.
	SPAUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
		cryptoutilSharedMagic.SPADisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentitySPAServiceID, cryptoutilSharedMagic.IdentitySPAServiceID),
	)

	// SPAUsageHealth is the usage message for the health subcommand.
	SPAUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.IdentitySPAServicePort),
	)

	// SPAUsageLivez is the usage message for the livez subcommand.
	SPAUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
	)

	// SPAUsageReadyz is the usage message for the readyz subcommand.
	SPAUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
	)

	// SPAUsageShutdown is the usage message for the shutdown subcommand.
	SPAUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
	)
)

// Spa implements the identity-spa service subcommand handler.
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
		spaInit,
	)
}

// spaServerStart implements the server subcommand.
func spaServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.IdentitySPAServerSettings]{
			UsageServer:  SPAUsageServer,
			ServiceLabel: cryptoutilSharedMagic.IdentitySPAServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.IdentitySPAServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.IdentitySPAServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.IdentitySPAServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// spaClient implements the client subcommand.
// CLI wrapper for client operations.
func spaClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.IdentitySPAServiceID, UsageText: SPAUsageClient}) {
		return 0
	}

	return 1
}

// spaInit implements the init subcommand.
// Generates PKI certificates for identity-spa TLS endpoints via the framework PKI init.
func spaInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: SPAUsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentitySPAServiceID, args, stdout, stderr)
}
