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
	return cryptoutilTemplateCli.RouteServiceFromIdentity(
		cryptoutilTemplateCli.NewServiceIdentity(
			cryptoutilSharedMagic.IdentityRSServiceID,
			cryptoutilSharedMagic.IdentityProductName,
			cryptoutilSharedMagic.RSServiceName,
			cryptoutilSharedMagic.RSDisplayName,
			uint16(cryptoutilSharedMagic.IdentityRSServicePort),
			cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			cryptoutilAppsServiceServer.NewFromConfig,
		),
		args, stdout, stderr,
	)
}
