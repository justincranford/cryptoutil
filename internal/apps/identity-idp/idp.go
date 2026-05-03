// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package idp provides the Identity Provider service entry point.
package idp

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsIdentityIdpServer "cryptoutil/internal/apps/identity-idp/server"
	cryptoutilAppsIdentityIdpServerConfig "cryptoutil/internal/apps/identity-idp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Idp implements the Identity Provider service subcommand handler.
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
		idpServiceInit,
	)
}

// idpServerStart implements the server subcommand.
func idpServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsIdentityIdpServerConfig.IdentityIDPServerSettings]{
			UsageServer:  IDPUsageServer,
			ServiceLabel: cryptoutilSharedMagic.IdentityIDPServiceID,
			FlagSetName:  "identity-idp-server",
			ParseConfig:  cryptoutilAppsIdentityIdpServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsIdentityIdpServerConfig.IdentityIDPServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsIdentityIdpServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsIdentityIdpServerConfig.IdentityIDPServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// idpClient implements the client subcommand.
// CLI wrapper for client operations.
func idpClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, IDPUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the Identity Provider service")

	return 1
}

// idpServiceInit implements the init subcommand.
// Generates PKI certificates for identity-idp TLS endpoints via the framework PKI init.
func idpServiceInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, IDPUsageInit)

		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentityIDPServiceID, args, stdout, stderr)
}
