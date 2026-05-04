// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package rs provides the identity-rs service entry point.
package rs

import (
	"context"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/identity-rs/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/identity-rs/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Rs implements the identity-rs service subcommand handler.
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
		rsInit,
	)
}

// rsServerStart implements the server subcommand.
func rsServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.IdentityRSServerSettings]{
			UsageServer:  RSUsageServer,
			ServiceLabel: cryptoutilSharedMagic.IdentityRSServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.IdentityRSServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.IdentityRSServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.IdentityRSServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// rsClient implements the client subcommand.
// CLI wrapper for client operations.
func rsClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.IdentityRSServiceID, UsageText: RSUsageClient}) {
		return 0
	}

	return 1
}

// rsInit implements the init subcommand.
// Generates PKI certificates for identity-rs TLS endpoints via the framework PKI init.
func rsInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: RSUsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.IdentityRSServiceID, args, stdout, stderr)
}
