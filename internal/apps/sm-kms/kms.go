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
	return cryptoutilTemplateCli.RouteServiceFromIdentity(
		cryptoutilTemplateCli.NewServiceIdentity(
			cryptoutilSharedMagic.KMSServiceID,
			cryptoutilSharedMagic.SMProductName,
			cryptoutilSharedMagic.KMSServiceName,
			cryptoutilSharedMagic.KMSDisplayName,
			uint16(cryptoutilSharedMagic.KMSServicePort),
			cryptoutilAppsFrameworkServiceConfig.ParseWithFlagSet,
			cryptoutilAppsServiceServer.NewKMSServerFromConfig,
		),
		args, stdout, stderr,
	)
}
