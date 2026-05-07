// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package ca provides the pki-ca service entry point.
package ca

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/pki-ca/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/pki-ca/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// CAUsageMain is the main usage message for the pki-ca command.
	CAUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
		cryptoutilSharedMagic.PKICADisplayName,
	)

	// CAUsageServer is the usage message for the server subcommand.
	CAUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
		cryptoutilSharedMagic.PKICADisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.PKICAServiceID, cryptoutilSharedMagic.PKICAServiceID),
	)

	// CAUsageClient is the usage message for the client subcommand.
	CAUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
		cryptoutilSharedMagic.PKICADisplayName,
	)

	// CAUsageInit is the usage message for the init subcommand.
	CAUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
		cryptoutilSharedMagic.PKICADisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.PKICAServiceID, cryptoutilSharedMagic.PKICAServiceID),
	)

	// CAUsageHealth is the usage message for the health subcommand.
	CAUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.PKICAServicePort),
	)

	// CAUsageLivez is the usage message for the livez subcommand.
	CAUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
	)

	// CAUsageReadyz is the usage message for the readyz subcommand.
	CAUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
	)

	// CAUsageShutdown is the usage message for the shutdown subcommand.
	CAUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
	)
)

// Ca implements the pki-ca service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Ca(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.PKICAServiceID,
			ProductName:       cryptoutilSharedMagic.PKIProductName,
			ServiceName:       cryptoutilSharedMagic.PKICAServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.PKICAServicePort),
			UsageMain:         CAUsageMain,
			UsageServer:       CAUsageServer,
			UsageClient:       CAUsageClient,
			UsageInit:         CAUsageInit,
			UsageHealth:       CAUsageHealth,
			UsageLivez:        CAUsageLivez,
			UsageReadyz:       CAUsageReadyz,
			UsageShutdown:     CAUsageShutdown,
		},
		args, stdout, stderr,
		caServerStart,
		caClient,
		caInit,
	)
}

// caServerStart implements the server subcommand.
func caServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.CAServerSettings]{
			UsageServer:  CAUsageServer,
			ServiceLabel: cryptoutilSharedMagic.PKICAServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.PKICAServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.CAServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.CAServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// caClient implements the client subcommand.
// CLI wrapper for client operations.
func caClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.PKICAServiceID, UsageText: CAUsageClient}) {
		return 0
	}

	return 1
}

// caInit implements the init subcommand.
// Generates PKI certificates for pki-ca TLS endpoints via the framework PKI init.
func caInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: CAUsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.PKICAServiceID, args, stdout, stderr)
}
