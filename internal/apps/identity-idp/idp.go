// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package idp provides the identity-idp service entry point.
package idp

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/identity-idp/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/identity-idp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// IDPUsageMain is the main usage message for the identity-idp command.
	IDPUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
		cryptoutilSharedMagic.IDPDisplayName,
	)

	// IDPUsageServer is the usage message for the server subcommand.
	IDPUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
		cryptoutilSharedMagic.IDPDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityIDPServiceID, cryptoutilSharedMagic.IdentityIDPServiceID),
	)

	// IDPUsageClient is the usage message for the client subcommand.
	IDPUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
		cryptoutilSharedMagic.IDPDisplayName,
	)

	// IDPUsageInit is the usage message for the init subcommand.
	IDPUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
		cryptoutilSharedMagic.IDPDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityIDPServiceID, cryptoutilSharedMagic.IdentityIDPServiceID),
	)

	// IDPUsageHealth is the usage message for the health subcommand.
	IDPUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.IdentityIDPServicePort),
	)

	// IDPUsageLivez is the usage message for the livez subcommand.
	IDPUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
	)

	// IDPUsageReadyz is the usage message for the readyz subcommand.
	IDPUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
	)

	// IDPUsageShutdown is the usage message for the shutdown subcommand.
	IDPUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
	)
)

// Idp implements the identity-idp service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Idp(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.IdentityIDPServiceID,
			ProductName:       cryptoutilSharedMagic.IdentityProductName,
			ServiceName:       cryptoutilSharedMagic.IDPServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.IdentityIDPServicePort),
			UsageMain:         IDPUsageMain,
			UsageServer:       IDPUsageServer,
			UsageClient:       IDPUsageClient,
			UsageInit:         IDPUsageInit,
			UsageHealth:       IDPUsageHealth,
			UsageLivez:        IDPUsageLivez,
			UsageReadyz:       IDPUsageReadyz,
			UsageShutdown:     IDPUsageShutdown,
		},
		args, stdout, stderr,
		idpServerStart,
		idpClient,
		idpInit,
	)
}

// idpServerStart implements the server subcommand.
func idpServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.IdentityIDPServerSettings]{
			UsageServer:  IDPUsageServer,
			ServiceLabel: cryptoutilSharedMagic.IdentityIDPServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.IdentityIDPServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.IdentityIDPServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.IdentityIDPServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// idpClient implements the client subcommand.
// CLI wrapper for client operations.
func idpClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.IdentityIDPServiceID, UsageText: IDPUsageClient}) {
		return 0
	}

	return 1
}

// idpInit implements the init subcommand.
// Generates PKI certificates for identity-idp TLS endpoints via the framework PKI init.
func idpInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: IDPUsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentityIDPServiceID, args, stdout, stderr)
}
