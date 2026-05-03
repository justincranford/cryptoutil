// Copyright (c) 2025-2026 Justin Cranford.
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
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/identity-spa/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/identity-spa/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.IdentitySPAServiceID}) {
		_, _ = fmt.Fprintln(stderr, SPAUsageClient)

		return 0
	}

	return 1
}

// spaInit implements the init subcommand.
// Generates PKI certificates for identity-spa TLS endpoints via the framework PKI init.
func spaInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, SPAUsageInit)

		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentitySPAServiceID, args, stdout, stderr)
}
