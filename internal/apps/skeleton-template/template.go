// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package template provides the skeleton-template service entry point.
package template

import (
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsServiceServer "cryptoutil/internal/apps/skeleton-template/server"
	cryptoutilAppsServiceServerConfig "cryptoutil/internal/apps/skeleton-template/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Template implements the skeleton-template service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Template(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteServiceFromIdentity(
		cryptoutilTemplateCli.NewServiceIdentity(
			cryptoutilSharedMagic.SkeletonTemplateServiceID,
			cryptoutilSharedMagic.SkeletonProductName,
			cryptoutilSharedMagic.SkeletonTemplateServiceName,
			cryptoutilSharedMagic.TemplateDisplayName,
			uint16(cryptoutilSharedMagic.SkeletonTemplateServicePort),
			cryptoutilAppsServiceServerConfig.ParseWithFlagSet,
			cryptoutilAppsServiceServer.NewFromConfig,
		),
		args, stdout, stderr,
	)
}
