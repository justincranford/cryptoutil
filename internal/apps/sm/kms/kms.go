// Copyright (c) 2025 Justin Cranford
//
//

// Package kms provides the Key Management Service entry point.
package kms

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilKMSServer "cryptoutil/internal/apps/sm/kms/server"
	cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
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

	// Parse base service template configuration.
	// Note: We prepend "start" as the subcommand for Parse() to validate.
	argsWithSubcommand := append([]string{"start"}, args...)

	fs := pflag.NewFlagSet("sm-kms-server", pflag.ContinueOnError)

	settings, err := cryptoutilAppsTemplateServiceConfig.ParseWithFlagSet(fs, argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilKMSServer.NewKMSServer(ctx, settings)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	// Start server with graceful shutdown.
	errChan := make(chan error, 1)

	go func() {
		_, _ = fmt.Fprintf(stdout, "🚀 Starting sm-kms service...\n")
		_, _ = fmt.Fprintf(stdout, "   Public Server: https://%s:%d\n", settings.BindPublicAddress, settings.BindPublicPort)
		_, _ = fmt.Fprintf(stdout, "   Admin Server:  https://%s:%d\n", settings.BindPrivateAddress, settings.BindPrivatePort)

		errChan <- srv.Start()
	}()

	// Wait for interrupt signal.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "❌ Server error: %v\n", err)

			return 1
		}
	case sig := <-sigChan:
		fmt.Printf("\n⏹️  Received signal %v, shutting down gracefully...\n", sig)
		srv.Shutdown()
	}

	fmt.Println("✅ sm-kms service stopped")

	return 0
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
// CLI wrapper for database and configuration initialization.
func kmsInit(args []string, _, stderr io.Writer) int {
	if cryptoutilTemplateCli.IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, KMSUsageInit)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "❌ Init subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will initialize database schema and configuration")

	return 1
}
