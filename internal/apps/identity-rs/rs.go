// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package rs provides the Resource Server service entry point.
package rs

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsIdentityRsServer "cryptoutil/internal/apps/identity-rs/server"
	cryptoutilAppsIdentityRsServerConfig "cryptoutil/internal/apps/identity-rs/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Rs implements the Resource Server service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Rs(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.IdentityRSServiceID,
			ProductName:       cryptoutilSharedMagic.IdentityProductName,
			ServiceName:       cryptoutilSharedMagic.RSServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.IdentityRSServicePort),
			UsageMain:         RSUsageMain,
			UsageServer:       RSUsageServer,
			UsageClient:       RSUsageClient,
			UsageInit:         RSUsageInit,
			UsageHealth:       RSUsageHealth,
			UsageLivez:        RSUsageLivez,
			UsageReadyz:       RSUsageReadyz,
			UsageShutdown:     RSUsageShutdown,
		},
		args, stdout, stderr,
		rsServerStart,
		rsClient,
		rsServiceInit,
	)
}

// rsServerStart implements the server subcommand.
func rsServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsIdentityRsServerConfig.IdentityRSServerSettings]{
			UsageServer:  RSUsageServer,
			ServiceLabel: cryptoutilSharedMagic.IdentityRSServiceID,
			FlagSetName:  "identity-rs-server",
			ParseConfig:  cryptoutilAppsIdentityRsServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsIdentityRsServerConfig.IdentityRSServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsIdentityRsServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsIdentityRsServerConfig.IdentityRSServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// rsClient implements the client subcommand.
// CLI wrapper for client operations.
func rsClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, RSUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the Resource Server service")

	return 1
}

// rsServiceInit implements the init subcommand.
// Generates PKI certificates for identity-rs TLS endpoints via the framework PKI init.
func rsServiceInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, RSUsageInit)

		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentityRSServiceID, args, stdout, stderr)
}
