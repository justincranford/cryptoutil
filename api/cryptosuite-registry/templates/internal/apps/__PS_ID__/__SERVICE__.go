//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package __SERVICE__ provides the __PS_ID__ service entry point.
package __SERVICE__

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	__FIRST_APP_IMPORT_ALIAS__ "__FIRST_APP_IMPORT_PATH__"
	__SECOND_APP_IMPORT_ALIAS__ "__SECOND_APP_IMPORT_PATH__"
	__THIRD_APP_IMPORT_ALIAS__ "__THIRD_APP_IMPORT_PATH__"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// __USAGE_PREFIX__UsageMain is the main usage message for the __PS_ID__ command.
	__USAGE_PREFIX__UsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_DISPLAY_NAME_CONST__,
	)

	// __USAGE_PREFIX__UsageServer is the usage message for the server subcommand.
	__USAGE_PREFIX__UsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_DISPLAY_NAME_CONST__,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.__SERVICE_ID_CONST__, cryptoutilSharedMagic.__SERVICE_ID_CONST__),
	)

	// __USAGE_PREFIX__UsageClient is the usage message for the client subcommand.
	__USAGE_PREFIX__UsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_DISPLAY_NAME_CONST__,
	)

	// __USAGE_PREFIX__UsageInit is the usage message for the init subcommand.
	__USAGE_PREFIX__UsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_DISPLAY_NAME_CONST__,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.__SERVICE_ID_CONST__, cryptoutilSharedMagic.__SERVICE_ID_CONST__),
	)

	// __USAGE_PREFIX__UsageHealth is the usage message for the health subcommand.
	__USAGE_PREFIX__UsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
		fmt.Sprintf("%d", cryptoutilSharedMagic.__SERVICE_PORT_CONST__),
	)

	// __USAGE_PREFIX__UsageLivez is the usage message for the livez subcommand.
	__USAGE_PREFIX__UsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
	)

	// __USAGE_PREFIX__UsageReadyz is the usage message for the readyz subcommand.
	__USAGE_PREFIX__UsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
	)

	// __USAGE_PREFIX__UsageShutdown is the usage message for the shutdown subcommand.
	__USAGE_PREFIX__UsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
	)
)

// __ENTRY_FUNC__ implements the __PS_ID__ service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func __ENTRY_FUNC__(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.__SERVICE_ID_CONST__,
			ProductName:       cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
			ServiceName:       cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.__SERVICE_PORT_CONST__),
			UsageMain:         __USAGE_PREFIX__UsageMain,
			UsageServer:       __USAGE_PREFIX__UsageServer,
			UsageClient:       __USAGE_PREFIX__UsageClient,
			UsageInit:         __USAGE_PREFIX__UsageInit,
			UsageHealth:       __USAGE_PREFIX__UsageHealth,
			UsageLivez:        __USAGE_PREFIX__UsageLivez,
			UsageReadyz:       __USAGE_PREFIX__UsageReadyz,
			UsageShutdown:     __USAGE_PREFIX__UsageShutdown,
		},
		args, stdout, stderr,
		__SERVICE__ServerStart,
		__SERVICE__Client,
		__SERVICE__Init,
	)
}

// __SERVICE__ServerStart implements the server subcommand.
func __SERVICE__ServerStart(args []string, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.StartServiceServer(
		args,
		stdout,
		stderr,
		cryptoutilTemplateCli.ServerStartOptions[*__SERVICE_CONFIG_ALIAS__.__SERVER_SETTINGS_TYPE__]{
			UsageServer:  __USAGE_PREFIX__UsageServer,
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
}

// __SERVICE__Client implements the client subcommand.
// CLI wrapper for client operations.
func __SERVICE__Client(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, ServiceID: cryptoutilSharedMagic.__SERVICE_ID_CONST__, UsageText: __USAGE_PREFIX__UsageClient}) {
		return 0
	}

	return 1
}

// __SERVICE__Init implements the init subcommand.
// Generates PKI certificates for __PS_ID__ TLS endpoints via the framework PKI init.
func __SERVICE__Init(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args, cryptoutilTemplateCli.ClientNotImplementedMessageConfig{Stderr: stderr, UsageText: __USAGE_PREFIX__UsageInit}) {
		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.__SERVICE_ID_CONST__, args, stdout, stderr)
}
