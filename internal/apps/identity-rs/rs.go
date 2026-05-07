// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package rs provides the identity-rs service entry point.
package rs

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/identity-rs/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/identity-rs/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// RSUsageMain is the main usage message for the identity-rs command.
	RSUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
		cryptoutilSharedMagic.RSDisplayName,
	)

	// RSUsageServer is the usage message for the server subcommand.
	RSUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
		cryptoutilSharedMagic.RSDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityRSServiceID, cryptoutilSharedMagic.IdentityRSServiceID),
	)

	// RSUsageClient is the usage message for the client subcommand.
	RSUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
		cryptoutilSharedMagic.RSDisplayName,
	)

	// RSUsageInit is the usage message for the init subcommand.
	RSUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
		cryptoutilSharedMagic.RSDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityRSServiceID, cryptoutilSharedMagic.IdentityRSServiceID),
	)

	// RSUsageHealth is the usage message for the health subcommand.
	RSUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.IdentityRSServicePort),
	)

	// RSUsageLivez is the usage message for the livez subcommand.
	RSUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
	)

	// RSUsageReadyz is the usage message for the readyz subcommand.
	RSUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
	)

	// RSUsageShutdown is the usage message for the shutdown subcommand.
	RSUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
	)
)

// Rs implements the identity-rs service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Rs(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.IdentityRSServiceID,
			ProductName:       cryptoutilSharedMagic.IdentityProductName,
			ServiceName:       cryptoutilSharedMagic.RSServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.IdentityRSServicePort),
			UsageMain:         RSUsageMain,
			UsageServer:       RSUsageServer,
			UsageClient:       RSUsageClient,
			UsageInit:         RSUsageInit,
			UsageHealth:       RSUsageHealth,
			UsageLivez:        RSUsageLivez,
			UsageReadyz:       RSUsageReadyz,
			UsageShutdown:     RSUsageShutdown,
		},
		args, stdout, stderr,
		rsServerStart,
		rsClient,
		rsInit,
	)
}

// rsServerStart implements the server subcommand.
func rsServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.IdentityRSServerSettings]{
			UsageServer:  RSUsageServer,
			ServiceLabel: cryptoutilSharedMagic.IdentityRSServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.IdentityRSServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.IdentityRSServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.IdentityRSServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// rsClient implements the client subcommand.
// CLI wrapper for client operations.
func rsClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.IdentityRSServiceID, UsageText: RSUsageClient}) {
		return 0
	}

	return 1
}

// rsInit implements the init subcommand.
// Generates PKI certificates for identity-rs TLS endpoints via the framework PKI init.
func rsInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: RSUsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentityRSServiceID, args, stdout, stderr)
}
