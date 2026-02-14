// Package ja provides the JWK Authority service entry point.
package ja

import (
	"context"
	"io"
	"log"

	cryptoutilAppsJoseJaServer "cryptoutil/internal/apps/jose/ja/server"
	cryptoutilAppsJoseJaServerConfig "cryptoutil/internal/apps/jose/ja/server/config"
)

// Ja is the entry point for the jose-ja service.
// It follows the standard pattern for service-level CLI entry points.
func Ja(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		log.Println("Usage: jose-ja <subcommand> [flags]")

		return 1
	}

	subcommand := args[0]
	switch subcommand {
	case "start":
		return startServer(args[1:])
	default:
		log.Printf("Unknown subcommand: %s\n", subcommand)

		return 1
	}
}

// startServer starts the jose-ja server.
func startServer(args []string) int {
	ctx := context.Background()

	// Parse configuration using jose-ja config package.
	// exitIfHelp=true to match standard CLI behavior.
	settings, err := cryptoutilAppsJoseJaServerConfig.Parse(args, true)
	if err != nil {
		log.Printf("Failed to parse configuration: %v\n", err)

		return 1
	}

	// Create server using NewFromConfig pattern.
	server, err := cryptoutilAppsJoseJaServer.NewFromConfig(ctx, settings)
	if err != nil {
		log.Printf("Failed to create server: %v\n", err)

		return 1
	}

	// Start server (blocks until shutdown).
	if err := server.Start(ctx); err != nil {
		log.Printf("Server error: %v\n", err)

		return 1
	}

	return 0
}
