// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package authz provides the identity-authz service entry point.
package authz

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/identity-authz/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// AUTHZUsageMain is the main usage message for the identity-authz command.
	AUTHZUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
		cryptoutilSharedMagic.AuthzDisplayName,
	)

	// AUTHZUsageServer is the usage message for the server subcommand.
	AUTHZUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
		cryptoutilSharedMagic.AuthzDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityAuthzServiceID, cryptoutilSharedMagic.IdentityAuthzServiceID),
	)

	// AUTHZUsageClient is the usage message for the client subcommand.
	AUTHZUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
		cryptoutilSharedMagic.AuthzDisplayName,
	)

	// AUTHZUsageInit is the usage message for the init subcommand.
	AUTHZUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
		cryptoutilSharedMagic.AuthzDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityAuthzServiceID, cryptoutilSharedMagic.IdentityAuthzServiceID),
	)

	// AUTHZUsageHealth is the usage message for the health subcommand.
	AUTHZUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.IdentityAuthzServicePort),
	)

	// AUTHZUsageLivez is the usage message for the livez subcommand.
	AUTHZUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
	)

	// AUTHZUsageReadyz is the usage message for the readyz subcommand.
	AUTHZUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
	)

	// AUTHZUsageShutdown is the usage message for the shutdown subcommand.
	AUTHZUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
	)
)

// Authz implements the identity-authz service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Authz(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.IdentityAuthzServiceID,
			ProductName:       cryptoutilSharedMagic.IdentityProductName,
			ServiceName:       cryptoutilSharedMagic.AuthzServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.IdentityAuthzServicePort),
			UsageMain:         AUTHZUsageMain,
			UsageServer:       AUTHZUsageServer,
			UsageClient:       AUTHZUsageClient,
			UsageInit:         AUTHZUsageInit,
			UsageHealth:       AUTHZUsageHealth,
			UsageLivez:        AUTHZUsageLivez,
			UsageReadyz:       AUTHZUsageReadyz,
			UsageShutdown:     AUTHZUsageShutdown,
		},
		args, stdout, stderr,
		authzServerStart,
		authzClient,
		authzInit,
	)
}

// authzServerStart implements the server subcommand.
func authzServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.IdentityAuthzServerSettings]{
			UsageServer:  AUTHZUsageServer,
			ServiceLabel: cryptoutilSharedMagic.IdentityAuthzServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.IdentityAuthzServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.IdentityAuthzServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.IdentityAuthzServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// authzClient implements the client subcommand.
// CLI wrapper for client operations.
func authzClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.IdentityAuthzServiceID, UsageText: AUTHZUsageClient}) {
		return 0
	}

	return 1
}

// authzInit implements the init subcommand.
// Generates PKI certificates for identity-authz TLS endpoints via the framework PKI init.
func authzInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: AUTHZUsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentityAuthzServiceID, args, stdout, stderr)
}
