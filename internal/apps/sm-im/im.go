// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
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
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/sm-im/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/sm-im/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Im implements the sm-im service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Im(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	id := cryptoutilTemplateCli.ServiceIdentity{
		ServiceID:   cryptoutilSharedMagic.IMServiceID,
		ProductName: cryptoutilSharedMagic.IMProductName,
		ServiceName: cryptoutilSharedMagic.IMServiceName,
		DisplayName: cryptoutilSharedMagic.IMDisplayName,
		ServicePort: uint16(cryptoutilSharedMagic.IMServicePort),
	}

	return cryptoutilTemplateCli.RouteServiceFromIdentity(
		id,
		args, stdout, stderr,
		func(serverArgs []string, serverStdout, serverStderr io.Writer) int {
			return cryptoutilTemplateCli.StartServiceServer(
				serverArgs,
				serverStdout,
				serverStderr,
				cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.SmIMServerSettings]{
					UsageServer:  cryptoutilTemplateCli.BuildServerUsage(id),
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
		},
	)
}
