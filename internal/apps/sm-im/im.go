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
	return cryptoutilTemplateCli.RouteServiceFromIdentity(
		cryptoutilTemplateCli.NewServiceIdentity(
			cryptoutilSharedMagic.IMServiceID,
			cryptoutilSharedMagic.IMProductName,
			cryptoutilSharedMagic.IMServiceName,
			cryptoutilSharedMagic.IMDisplayName,
			uint16(cryptoutilSharedMagic.IMServicePort),
			cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			cryptoutilAppsServiceServer.NewIMServerFromConfig,
		),
		args, stdout, stderr,
	)
}
