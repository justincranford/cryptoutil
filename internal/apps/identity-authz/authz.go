// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package authz provides the Authorization Server service entry point.
package authz

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsIdentityAuthzServer "cryptoutil/internal/apps/identity-authz/server"
	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Authz implements the Authorization Server service subcommand handler.
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
		authzServiceInit,
	)
}

// authzServerStart implements the server subcommand.
func authzServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsIdentityAuthzServerConfig.IdentityAuthzServerSettings]{
			UsageServer:  AUTHZUsageServer,
			ServiceLabel: cryptoutilSharedMagic.IdentityAuthzServiceID,
			FlagSetName:  "identity-authz-server",
			ParseConfig:  cryptoutilAppsIdentityAuthzServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsIdentityAuthzServerConfig.IdentityAuthzServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsIdentityAuthzServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsIdentityAuthzServerConfig.IdentityAuthzServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// authzClient implements the client subcommand.
// CLI wrapper for client operations.
func authzClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, AUTHZUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the Authorization Server service")

	return 1
}

// authzServiceInit implements the init subcommand.
// Generates PKI certificates for identity-authz TLS endpoints via the framework PKI init.
func authzServiceInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, AUTHZUsageInit)

		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentityAuthzServiceID, args, stdout, stderr)
}
