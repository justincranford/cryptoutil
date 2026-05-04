// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package authz provides the identity-authz service entry point.
package authz

import (
	"context"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/identity-authz/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
