// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package ja provides the jose-ja service entry point.
package ja

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/jose-ja/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/jose-ja/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Ja implements the jose-ja service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Ja(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.JoseJAServiceID,
			ProductName:       cryptoutilSharedMagic.JoseProductName,
			ServiceName:       cryptoutilSharedMagic.JoseJAServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.JoseJAServicePort),
			UsageMain:         JAUsageMain,
			UsageServer:       JAUsageServer,
			UsageClient:       JAUsageClient,
			UsageInit:         JAUsageInit,
			UsageHealth:       JAUsageHealth,
			UsageLivez:        JAUsageLivez,
			UsageReadyz:       JAUsageReadyz,
			UsageShutdown:     JAUsageShutdown,
		},
		args, stdout, stderr,
		jaServerStart,
		jaClient,
		jaInit,
	)
}

// jaServerStart implements the server subcommand.
func jaServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.JoseJAServerSettings]{
			UsageServer:  JAUsageServer,
			ServiceLabel: cryptoutilSharedMagic.JoseJAServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.JoseJAServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.JoseJAServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.JoseJAServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// jaClient implements the client subcommand.
// CLI wrapper for client operations.
func jaClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.JoseJAServiceID}) {
		_, _ = fmt.Fprintln(stderr, JAUsageClient)

		return 0
	}

	return 1
}

// jaInit implements the init subcommand.
// Generates PKI certificates for jose-ja TLS endpoints via the framework PKI init.
func jaInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, JAUsageInit)

		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.JoseJAServiceID, args, stdout, stderr)
}
