//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package __SERVICE__ provides the __PS_ID__ service entry point.
package __SERVICE__

import (
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
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
