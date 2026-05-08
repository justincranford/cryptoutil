// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package spa provides the identity-spa service entry point.
package spa

import (
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/identity-spa/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/identity-spa/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Spa implements the identity-spa service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Spa(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteServiceFromIdentity(
		cryptoutilTemplateCli.NewServiceIdentity(
			cryptoutilSharedMagic.IdentitySPAServiceID,
			cryptoutilSharedMagic.IdentityProductName,
			cryptoutilSharedMagic.SPAServiceName,
			cryptoutilSharedMagic.SPADisplayName,
			uint16(cryptoutilSharedMagic.IdentitySPAServicePort),
			cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			cryptoutilAppsServiceServer.NewFromConfig,
		),
		args, stdout, stderr,
	)
}
