// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package ca provides the pki-ca service entry point.
package ca

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/pki-ca/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/pki-ca/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Ca implements the pki-ca service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Ca(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.PKICAServiceID,
			ProductName:       cryptoutilSharedMagic.PKIProductName,
			ServiceName:       cryptoutilSharedMagic.PKICAServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.PKICAServicePort),
			UsageMain:         CAUsageMain,
			UsageServer:       CAUsageServer,
			UsageClient:       CAUsageClient,
			UsageInit:         CAUsageInit,
			UsageHealth:       CAUsageHealth,
			UsageLivez:        CAUsageLivez,
			UsageReadyz:       CAUsageReadyz,
			UsageShutdown:     CAUsageShutdown,
		},
		args, stdout, stderr,
		caServerStart,
		caClient,
		caInit,
	)
}

// caServerStart implements the server subcommand.
func caServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.CAServerSettings]{
			UsageServer:  CAUsageServer,
			ServiceLabel: cryptoutilSharedMagic.PKICAServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.PKICAServiceID),
			ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.CAServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.CAServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// caClient implements the client subcommand.
// CLI wrapper for client operations.
func caClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.PKICAServiceID}) {
		_, _ = fmt.Fprintln(stderr, CAUsageClient)

		return 0
	}

	return 1
}

// caInit implements the init subcommand.
// Generates PKI certificates for pki-ca TLS endpoints via the framework PKI init.
func caInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, CAUsageInit)

		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.PKICAServiceID, args, stdout, stderr)
}
