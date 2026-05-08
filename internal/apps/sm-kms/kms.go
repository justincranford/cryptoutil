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
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/sm-kms/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Kms implements the sm-kms service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Kms(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	id := cryptoutilTemplateCli.ServiceIdentity{
		ServiceID:   cryptoutilSharedMagic.KMSServiceID,
		ProductName: cryptoutilSharedMagic.SMProductName,
		ServiceName: cryptoutilSharedMagic.KMSServiceName,
		DisplayName: cryptoutilSharedMagic.KMSDisplayName,
		ServicePort: uint16(cryptoutilSharedMagic.KMSServicePort),
	}

	return cryptoutilTemplateCli.RouteServiceFromIdentity(
		id,
		args, stdout, stderr,
		func(serverArgs []string, serverStdout, serverStderr io.Writer) int {
			return cryptoutilTemplateCli.StartServiceServer(
				serverArgs,
				serverStdout,
				serverStderr,
				cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings]{
					UsageServer:  cryptoutilTemplateCli.BuildServerUsage(id),
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
		},
	)
}
