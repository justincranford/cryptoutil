// Copyright (c) 2025 Justin Cranford
//

// Package jose implements the jose product command router.
package jose

import (
	"fmt"
	"io"

	cryptoutilAppsJoseJa "cryptoutil/internal/apps/jose/ja"
)

const (
	helpCommand      = "help"
	helpFlag         = "--help"
	helpShortFlag    = "-h"
	versionCommand   = "version"
	versionFlag      = "--version"
	versionShortFlag = "-v"
)

// Jose implements the jose product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil jose ja server
// - Product: jose ja server
// - Product-Service: jose-ja server (via main.go delegation).
func Jose(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)

		return 1
	}

	// Check for help flags.
	if args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag {
		printUsage(stderr)

		return 0
	}

	// Check for version flags.
	if args[0] == versionCommand || args[0] == versionFlag || args[0] == versionShortFlag {
		printVersion(stdout)

		return 0
	}

	// Route to service command.
	switch args[0] {
	case "ja":
		return cryptoutilAppsJoseJa.Ja(args[1:], stdin, stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown service: %s\n\n", args[0])
		printUsage(stderr)

		return 1
	}
}

// printUsage prints the jose product usage information.
func printUsage(stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, `Usage: jose <service> <subcommand> [options]

Available services:
  ja          JWK Authority service

Use "jose <service> help" for service-specific help.
Use "jose version" for version information.`)
}

// printVersion prints the jose product version information.
func printVersion(stdout io.Writer) {
	// Version information should be injected from the calling binary.
	_, _ = fmt.Fprintln(stdout, "jose product (cryptoutil)")
}
