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

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilKMSServer "cryptoutil/internal/apps/sm/kms/server"
	cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	helpCommand   = "help"
	helpFlag      = "--help"
	helpShortFlag = "-h"
)

// Kms implements the Key Management Service subcommand handler.
// Handles subcommands: server, client, init, health, livez, readyz, shutdown.
func Kms(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteService(
		cryptoutilTemplateCli.ServiceConfig{
			ServiceID:        "sm-kms",
			ProductName:      "sm",
			ServiceName:      "kms",
			DefaultPublicPort: uint16(cryptoutilSharedMagic.KMSServicePort),
			UsageMain:        KMSUsageMain,
			UsageServer:      KMSUsageServer,
			UsageClient:      KMSUsageClient,
			UsageInit:        KMSUsageInit,
			UsageHealth:      KMSUsageHealth,
			UsageLivez:       KMSUsageLivez,
			UsageReadyz:      KMSUsageReadyz,
			UsageShutdown:    KMSUsageShutdown,
		},
		args, stdout, stderr,
		kmsServerStart,
		kmsClient,
		kmsInit,
	)
}

// kmsServerStart implements the server subcommand.
func kmsServerStart(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, KMSUsageServer)

		return 0
	}

	ctx := context.Background()

	// Parse base service template configuration.
	// Note: We prepend "start" as the subcommand for Parse() to validate.
	argsWithSubcommand := append([]string{"start"}, args...)

	settings, err := cryptoutilAppsTemplateServiceConfig.Parse(argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "\u274c Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := cryptoutilKMSServer.NewKMSServer(ctx, settings)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "\u274c Failed to create server: %v\n", err)

		return 1
	}

	// Start server with graceful shutdown.
	errChan := make(chan error, 1)

	go func() {
		_, _ = fmt.Fprintf(stdout, "\U0001f680 Starting sm-kms service...\n")
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
			_, _ = fmt.Fprintf(stderr, "\u274c Server error: %v\n", err)

			return 1
		}
	case sig := <-sigChan:
		fmt.Printf("\n\u23f9\ufe0f  Received signal %v, shutting down gracefully...\n", sig)
		srv.Shutdown()
	}

	fmt.Println("\u2705 sm-kms service stopped")

	return 0
}

// kmsClient implements the client subcommand.
// CLI wrapper for client operations.
func kmsClient(args []string, _, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, KMSUsageClient)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "\u274c Client subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will provide CLI tools for interacting with the KMS service")

	return 1
}

// kmsInit implements the init subcommand.
// CLI wrapper for database and configuration initialization.
func kmsInit(args []string, _, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, KMSUsageInit)

		return 0
	}

	_, _ = fmt.Fprintln(stderr, "\u274c Init subcommand not yet implemented")
	_, _ = fmt.Fprintln(stderr, "   This will initialize database schema and configuration")

	return 1
}
