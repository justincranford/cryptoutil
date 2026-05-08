//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
//
//

// Package __SERVICE__ provides the __PS_ID__ service entry point.
package __SERVICE__

import (
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	__FIRST_APP_IMPORT_ALIAS__ "__FIRST_APP_IMPORT_PATH__"
	__SECOND_APP_IMPORT_ALIAS__ "__SECOND_APP_IMPORT_PATH__"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// __ENTRY_FUNC__ implements the __PS_ID__ service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func __ENTRY_FUNC__(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteServiceFromIdentity(
		cryptoutilTemplateCli.NewServiceIdentity(
			cryptoutilSharedMagic.__SERVICE_ID_CONST__,
			cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
			cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
			cryptoutilSharedMagic.__SERVICE_DISPLAY_NAME_CONST__,
			uint16(cryptoutilSharedMagic.__SERVICE_PORT_CONST__),
			__SERVICE_CONFIG_ALIAS__.__PARSE_CONFIG_FUNC__,
			__SERVER_ALIAS__.__NEW_SERVER_CONSTRUCTOR__,
		),
		args, stdout, stderr,
	)
}
