// Copyright (c) 2025 Justin Cranford
//

// Package sm implements the sm (Secrets Manager) product command router.
package sm

import (
	"fmt"
	"io"

	cryptoutilAppsSmKms "cryptoutil/internal/apps/sm/kms"
)

const (
	helpCommand      = "help"
	helpFlag         = "--help"
	helpShortFlag    = "-h"
	versionCommand   = "version"
	versionFlag      = "--version"
	versionShortFlag = "-v"
)

// Sm implements the sm (Secrets Manager) product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil sm kms server
// - Product: sm kms server
// - Product-Service: sm-kms server (via main.go delegation).
func Sm(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
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
	case "kms":
		return cryptoutilAppsSmKms.Kms(args[1:], stdin, stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown service: %s\n\n", args[0])
		printUsage(stderr)

		return 1
	}
}

// printUsage prints the sm product usage information.
func printUsage(stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, `Usage: sm <service> <subcommand> [options]

Available services:
  kms         Key Management Service

Use "sm <service> help" for service-specific help.
Use "sm version" for version information.`)
}

// printVersion prints the sm product version information.
func printVersion(stdout io.Writer) {
	// Version information should be injected from the calling binary.
	_, _ = fmt.Fprintln(stdout, "sm product (cryptoutil)")
}
