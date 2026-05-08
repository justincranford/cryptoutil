// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package authz provides the identity-authz service entry point.
package authz

import (
	"context"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/identity-authz/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Authz implements the identity-authz service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Authz(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	id := cryptoutilTemplateCli.ServiceIdentity{
		ServiceID:   cryptoutilSharedMagic.IdentityAuthzServiceID,
		ProductName: cryptoutilSharedMagic.IdentityProductName,
		ServiceName: cryptoutilSharedMagic.AuthzServiceName,
		DisplayName: cryptoutilSharedMagic.AuthzDisplayName,
		ServicePort: uint16(cryptoutilSharedMagic.IdentityAuthzServicePort),
	}

	return cryptoutilTemplateCli.RouteServiceFromIdentity(
		id,
		args, stdout, stderr,
		func(serverArgs []string, serverStdout, serverStderr io.Writer) int {
			return cryptoutilTemplateCli.StartServiceServer(
				serverArgs,
				serverStdout,
				serverStderr,
				cryptoutilTemplateCli.ServerStartOptions[*cryptoutilAppsServiceServerConfig.IdentityAuthzServerSettings]{
					UsageServer:  cryptoutilTemplateCli.BuildServerUsage(id),
					ServiceLabel: cryptoutilSharedMagic.IdentityAuthzServiceID,
					FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.IdentityAuthzServiceID),
					ParseConfig:  cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
					NewServer: func(ctx context.Context, settings *cryptoutilAppsServiceServerConfig.IdentityAuthzServerSettings) (cryptoutilTemplateCli.ReadyStarter, error) {
						return cryptoutilAppsServiceServer.NewFromConfig(ctx, settings)
					},
					BindAddresses: func(settings *cryptoutilAppsServiceServerConfig.IdentityAuthzServerSettings) (string, uint16, string, uint16) {
						return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
					},
				},
			)
		},
	)
}
