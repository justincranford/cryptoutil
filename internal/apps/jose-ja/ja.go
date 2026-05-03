// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package ja provides the JWK Authority service entry point.
package ja

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsJoseJaServer "cryptoutil/internal/apps/jose-ja/server"
	cryptoutilAppsJoseJaServerConfig "cryptoutil/internal/apps/jose-ja/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Ja implements the JWK Authority service subcommand handler.
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
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsJoseJaServerConfig.JoseJAServerSettings]{
			UsageServer:  JAUsageServer,
			ServiceLabel: cryptoutilSharedMagic.JoseJAServiceID,
			FlagSetName:  "jose-ja-server",
			ParseConfig:  cryptoutilAppsJoseJaServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsJoseJaServerConfig.JoseJAServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsJoseJaServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsJoseJaServerConfig.JoseJAServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// jaClient implements the client subcommand.
// CLI wrapper for client operations.
func jaClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, JAUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the JWK Authority service")

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
