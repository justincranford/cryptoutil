// Copyright (c) 2025 Justin Cranford
//
//

// Package cipher implements the cipher product command router.
package cipher

import (
	"fmt"
	"io"

	cryptoutilAppsCipherIm "cryptoutil/internal/apps/cipher/im"
)

const (
	helpCommand      = "help"
	helpFlag         = "--help"
	helpShortFlag    = "-h"
	versionCommand   = "version"
	versionFlag      = "--version"
	versionShortFlag = "-v"
)

// Cipher implements the cipher product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil cipher im server
// - Product: cipher im server
// - Product-Service: cipher-im server (via main.go delegation).
func Cipher(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
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
	case "im":
		return cryptoutilAppsCipherIm.Im(args[1:], stdin, stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown service: %s\n\n", args[0])
		printUsage(stderr)

		return 1
	}
}

// printUsage prints the cipher product usage information.
func printUsage(stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, `Usage: cipher <service> <subcommand> [options]

Available services:
  im          Instant messaging service

Use "cipher <service> help" for service-specific help.
Use "cipher version" for version information.`)
}

// printVersion prints the cipher product version information.
func printVersion(stdout io.Writer) {
	// Version information should be injected from the calling binary.
	_, _ = fmt.Fprintln(stdout, "cipher product (cryptoutil)")
}
