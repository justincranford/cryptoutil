// Copyright (c) 2025 Justin Cranford
//

// Package pki implements the pki product command router.
package pki

import (
	"fmt"
	"io"

	cryptoutilAppsPkiCa "cryptoutil/internal/apps/pki/ca"
)

const (
	helpCommand      = "help"
	helpFlag         = "--help"
	helpShortFlag    = "-h"
	versionCommand   = "version"
	versionFlag      = "--version"
	versionShortFlag = "-v"
)

// Pki implements the pki product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil pki ca server
// - Product: pki ca server
// - Product-Service: pki-ca server (via main.go delegation).
func Pki(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
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
	case "ca":
		return cryptoutilAppsPkiCa.Ca(args[1:], stdin, stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown service: %s\n\n", args[0])
		printUsage(stderr)

		return 1
	}
}

// printUsage prints the pki product usage information.
func printUsage(stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, `Usage: pki <service> <subcommand> [options]

Available services:
  ca          Certificate Authority service

Use "pki <service> help" for service-specific help.
Use "pki version" for version information.`)
}

// printVersion prints the pki product version information.
func printVersion(stdout io.Writer) {
	// Version information should be injected from the calling binary.
	_, _ = fmt.Fprintln(stdout, "pki product (cryptoutil)")
}
