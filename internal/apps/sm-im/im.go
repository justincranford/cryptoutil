// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package im provides the sm-im service entry point.
package im

import (
	"context"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/sm-im/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/sm-im/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Im implements the sm-im service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Im(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.IMServiceID,
			ProductName:       cryptoutilSharedMagic.IMProductName,
			ServiceName:       cryptoutilSharedMagic.IMServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.IMServicePort),
			UsageMain:         IMUsageMain,
			UsageServer:       IMUsageServer,
			UsageClient:       IMUsageClient,
			UsageInit:         IMUsageInit,
			UsageHealth:       IMUsageHealth,
			UsageLivez:        IMUsageLivez,
			UsageReadyz:       IMUsageReadyz,
			UsageShutdown:     IMUsageShutdown,
		},
		args, stdout, stderr,
		imServerStart,
		imClient,
		imInit,
	)
}

// imServerStart implements the server subcommand.
func imServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.SmIMServerSettings]{
			UsageServer:  IMUsageServer,
			ServiceLabel: cryptoutilSharedMagic.IMServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.IMServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.SmIMServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewIMServerFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.SmIMServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// imClient implements the client subcommand.
// CLI wrapper for client operations.
func imClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.IMServiceID, UsageText: IMUsageClient}) {
		return 0
	}

	return 1
}

// imInit implements the init subcommand.
// Generates PKI certificates for sm-im TLS endpoints via the framework PKI init.
func imInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: IMUsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IMServiceID, args, stdout, stderr)
}
