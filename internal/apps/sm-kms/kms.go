// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package kms provides the sm-kms service entry point.
package kms

import (
	"context"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/sm-kms/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Kms implements the sm-kms service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Kms(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.KMSServiceID,
			ProductName:       cryptoutilSharedMagic.SMProductName,
			ServiceName:       cryptoutilSharedMagic.KMSServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.KMSServicePort),
			UsageMain:         KMSUsageMain,
			UsageServer:       KMSUsageServer,
			UsageClient:       KMSUsageClient,
			UsageInit:         KMSUsageInit,
			UsageHealth:       KMSUsageHealth,
			UsageLivez:        KMSUsageLivez,
			UsageReadyz:       KMSUsageReadyz,
			UsageShutdown:     KMSUsageShutdown,
		},
		args, stdout, stderr,
		kmsServerStart,
		kmsClient,
		kmsInit,
	)
}

// kmsServerStart implements the server subcommand.
func kmsServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings]{
			UsageServer:  KMSUsageServer,
			ServiceLabel: cryptoutilSharedMagic.KMSServiceID,
			FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.KMSServiceID),
			ParseConfig:  cryptoutilAppsFrameworkServiceConfig.ParseWithFlagSet,
			NewServer: func(ctx context.Context, settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
				return cryptoutilAppsServiceServer.NewKMSServerFromConfig(ctx, settings)
			},
			BindAddresses: func(settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (string, uint16, string, uint16) {
				return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
			},
		},
	)
}

// kmsClient implements the client subcommand.
// CLI wrapper for client operations.
func kmsClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.KMSServiceID, UsageText: KMSUsageClient}) {
		return 0
	}

	return 1
}

// kmsInit implements the init subcommand.
// Generates PKI certificates for sm-kms TLS endpoints via the framework PKI init.
func kmsInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: KMSUsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.KMSServiceID, args, stdout, stderr)
}
