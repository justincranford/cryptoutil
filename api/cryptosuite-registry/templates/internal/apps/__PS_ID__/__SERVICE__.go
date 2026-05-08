//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package __SERVICE__ provides the __PS_ID__ service entry point.
package __SERVICE__

import (
	"context"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
)

// __ENTRY_FUNC__ implements the __PS_ID__ service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func __ENTRY_FUNC__(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	id := cryptoutilTemplateCli.ServiceIdentity{
		ServiceID:   cryptoutilSharedMagic.__SERVICE_ID_CONST__,
		ProductName: cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		ServiceName: cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
		DisplayName: cryptoutilSharedMagic.__SERVICE_DISPLAY_NAME_CONST__,
		ServicePort: uint16(cryptoutilSharedMagic.__SERVICE_PORT_CONST__),
	}

	return cryptoutilTemplateCli.RouteServiceFromIdentity(
		id,
		args, stdout, stderr,
		func(serverArgs []string, serverStdout, serverStderr io.Writer) int {
			return cryptoutilTemplateCli.StartServiceServer(
				serverArgs,
				serverStdout,
				serverStderr,
				cryptoutilTemplateCli.ServerStartOptions[*__SERVICE_CONFIG_ALIAS__.__SERVER_SETTINGS_TYPE__]{
					UsageServer:  cryptoutilTemplateCli.BuildServerUsage(id),
					ServiceLabel: cryptoutilSharedMagic.__SERVICE_ID_CONST__,
					FlagSetName:  cryptoutilTemplateCli.ServerFlagSetName(cryptoutilSharedMagic.__SERVICE_ID_CONST__),
					ParseConfig:  __SERVICE_CONFIG_ALIAS__.__PARSE_CONFIG_FUNC__,
					NewServer: func(ctx context.Context, settings *__SERVICE_CONFIG_ALIAS__.__SERVER_SETTINGS_TYPE__) (cryptoutilTemplateCli.ReadyStarter, error) {
						return __SERVER_ALIAS__.__NEW_SERVER_CONSTRUCTOR__(ctx, settings)
					},
					BindAddresses: func(settings *__SERVICE_CONFIG_ALIAS__.__SERVER_SETTINGS_TYPE__) (string, uint16, string, uint16) {
						return settings.BindPublicAddress, settings.BindPublicPort, settings.BindPrivateAddress, settings.BindPrivatePort
					},
				},
			)
		},
	)
}
