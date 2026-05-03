// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package kms provides the Key Management Service entry point.
package kms

import (
	"context"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilTemplateCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceLifecycle "cryptoutil/internal/apps-framework/service/lifecycle"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilAppsSmKmsServer "cryptoutil/internal/apps/sm-kms/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
)

// Kms implements the Key Management Service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Kms(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:         cryptoutilSharedMagic.KMSServiceID,
			ProductName:       cryptoutilSharedMagic.SMProductName,
			ServiceName:       cryptoutilSharedMagic.KMSServiceName,
			DefaultPublicPort: uint16(cryptoutilSharedMagic.KMSServicePort),
			UsageMain:         KMSUsageMain,
			UsageServer:       KMSUsageServer,
			UsageClient:       KMSUsageClient,
			UsageInit:         KMSUsageInit,
			UsageHealth:       KMSUsageHealth,
			UsageLivez:        KMSUsageLivez,
			UsageReadyz:       KMSUsageReadyz,
			UsageShutdown:     KMSUsageShutdown,
		},
		args, stdout, stderr,
		kmsServerStart,
		kmsClient,
		kmsInit,
	)
}

// kmsServerStart implements the server subcommand.
func kmsServerStart(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, KMSUsageServer)

		return 0
	}

	ctx := context.Background()

	// Parse configuration using config.Parse() which leverages viper+pflag.
	// Note: We prepend "start" as the subcommand for Parse() to validate.
	argsWithSubcommand := append([]string{"start"}, args...)

	fs := pflag.NewFlagSet("sm-kms-server", pflag.ContinueOnError)

	settings, err := cryptoutilAppsFrameworkServiceConfig.ParseWithFlagSet(fs, argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilAppsSmKmsServer.NewKMSServerFromConfig(ctx, settings)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	// Mark server as ready so /admin/api/v1/readyz return 200 OK instead of 503 Service Unavailable.
	srv.SetReady(true)

	_, _ = fmt.Fprintf(stdout, "🚀 Starting sm-kms service...\n")
	_, _ = fmt.Fprintf(stdout, "   Public Server: https://%s:%d\n", settings.BindPublicAddress, settings.BindPublicPort)
	_, _ = fmt.Fprintf(stdout, "   Admin Server:  https://%s:%d\n", settings.BindPrivateAddress, settings.BindPrivatePort)

	exitCode := cryptoutilAppsFrameworkServiceLifecycle.RunService(ctx, stdout, stderr, srv)

	_, _ = fmt.Fprintln(stdout, "✅ sm-kms service stopped")

	return exitCode
}

// kmsClient implements the client subcommand.
// CLI wrapper for client operations.
func kmsClient(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, KMSUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the KMS service")

	return 1
}

// kmsInit implements the init subcommand.
// Generates PKI certificates for sm-kms TLS endpoints via the framework PKI init.
func kmsInit(args []string, stdout, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, KMSUsageInit)

		return 0
	}

	return cryptoutilAppsFrameworkTls.InitForService(cryptoutilSharedMagic.KMSServiceID, args, stdout, stderr)
}
