// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package kms provides the sm-kms service entry point.
package kms

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/sm-kms/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// KMSUsageMain is the main usage message for the sm-kms command.
	KMSUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
		cryptoutilSharedMagic.KMSDisplayName,
	)

	// KMSUsageServer is the usage message for the server subcommand.
	KMSUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
		cryptoutilSharedMagic.KMSDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.KMSServiceID, cryptoutilSharedMagic.KMSServiceID),
	)

	// KMSUsageClient is the usage message for the client subcommand.
	KMSUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
		cryptoutilSharedMagic.KMSDisplayName,
	)

	// KMSUsageInit is the usage message for the init subcommand.
	KMSUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
		cryptoutilSharedMagic.KMSDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.KMSServiceID, cryptoutilSharedMagic.KMSServiceID),
	)

	// KMSUsageHealth is the usage message for the health subcommand.
	KMSUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.KMSServicePort),
	)

	// KMSUsageLivez is the usage message for the livez subcommand.
	KMSUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
	)

	// KMSUsageReadyz is the usage message for the readyz subcommand.
	KMSUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
	)

	// KMSUsageShutdown is the usage message for the shutdown subcommand.
	KMSUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
	)
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
