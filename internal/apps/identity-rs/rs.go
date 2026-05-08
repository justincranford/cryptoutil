// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
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
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/identity-rs/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/identity-rs/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Rs implements the identity-rs service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Rs(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	id := cryptoutilTemplateCli.ServiceIdentity{
		ServiceID:   cryptoutilSharedMagic.IdentityRSServiceID,
		ProductName: cryptoutilSharedMagic.IdentityProductName,
		ServiceName: cryptoutilSharedMagic.RSServiceName,
		DisplayName: cryptoutilSharedMagic.RSDisplayName,
		ServicePort: uint16(cryptoutilSharedMagic.IdentityRSServicePort),
	}

	return cryptoutilTemplateCli.RouteServiceFromIdentity(
		id,
		args, stdout, stderr,
		func(serverArgs []string, serverStdout, serverStderr io.Writer) int {
			return cryptoutilTemplateCli.StartServiceServer(
				serverArgs,
				serverStdout,
				serverStderr,
				cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.IdentityRSServerSettings]{
					UsageServer:  cryptoutilTemplateCli.BuildServerUsage(id),
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
		},
	)
}
