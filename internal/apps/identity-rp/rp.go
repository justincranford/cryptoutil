// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package rp provides the identity-rp service entry point.
package rp

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/identity-rp/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/identity-rp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// RPUsageMain is the main usage message for the identity-rp command.
	RPUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
		cryptoutilSharedMagic.RPDisplayName,
	)

	// RPUsageServer is the usage message for the server subcommand.
	RPUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
		cryptoutilSharedMagic.RPDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityRPServiceID, cryptoutilSharedMagic.IdentityRPServiceID),
	)

	// RPUsageClient is the usage message for the client subcommand.
	RPUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
		cryptoutilSharedMagic.RPDisplayName,
	)

	// RPUsageInit is the usage message for the init subcommand.
	RPUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
		cryptoutilSharedMagic.RPDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityRPServiceID, cryptoutilSharedMagic.IdentityRPServiceID),
	)

	// RPUsageHealth is the usage message for the health subcommand.
	RPUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.IdentityRPServicePort),
	)

	// RPUsageLivez is the usage message for the livez subcommand.
	RPUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
	)

	// RPUsageReadyz is the usage message for the readyz subcommand.
	RPUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
	)

	// RPUsageShutdown is the usage message for the shutdown subcommand.
	RPUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
	)
)

// Rp implements the identity-rp service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Rp(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.IdentityRPServiceID,
			ProductName:       cryptoutilSharedMagic.IdentityProductName,
			ServiceName:       cryptoutilSharedMagic.RPServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.IdentityRPServicePort),
			UsageMain:         RPUsageMain,
			UsageServer:       RPUsageServer,
			UsageClient:       RPUsageClient,
			UsageInit:         RPUsageInit,
			UsageHealth:       RPUsageHealth,
			UsageLivez:        RPUsageLivez,
			UsageReadyz:       RPUsageReadyz,
			UsageShutdown:     RPUsageShutdown,
		},
		args, stdout, stderr,
		rpServerStart,
		rpClient,
		rpInit,
	)
}

// rpServerStart implements the server subcommand.
func rpServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.IdentityRPServerSettings]{
			UsageServer:  RPUsageServer,
			ServiceLabel: cryptoutilSharedMagic.IdentityRPServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.IdentityRPServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.IdentityRPServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.IdentityRPServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// rpClient implements the client subcommand.
// CLI wrapper for client operations.
func rpClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.IdentityRPServiceID, UsageText: RPUsageClient}) {
		return 0
	}

	return 1
}

// rpInit implements the init subcommand.
// Generates PKI certificates for identity-rp TLS endpoints via the framework PKI init.
func rpInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: RPUsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentityRPServiceID, args, stdout, stderr)
}
