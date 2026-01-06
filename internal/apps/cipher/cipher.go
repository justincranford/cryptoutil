// Copyright (c) 2025 Justin Cranford
//
//

// Package cipher implements the cipher product command router.
package cipher

import (
	"fmt"
	"os"

	"cryptoutil/internal/apps/cipher/im"
)

const (
	helpCommand      = "help"
	helpFlag         = "--help"
	helpShortFlag    = "-h"
	versionCommand   = "version"
	versionFlag      = "--version"
	versionShortFlag = "-v"
	urlFlag          = "--url"
)

// Cipher implements the cipher product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil cipher im server
// - Product: cipher im server
// - Product-Service: cipher-im server (via main.go delegation).
func Cipher(args []string) int {
	if len(args) == 0 {
		printUsage()

		return 1
	}

	// Check for help flags.
	if args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag {
		printUsage()

		return 0
	}

	// Check for version flags.
	if args[0] == versionCommand || args[0] == versionFlag || args[0] == versionShortFlag {
		printVersion()

		return 0
	}

	// Route to service command.
	switch args[0] {
	case "im":
		return im.IM(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown service: %s\n\n", args[0])
		printUsage()

		return 1
	}
}

// printUsage prints the cipher product usage information.
func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: cipher <service> <subcommand> [options]

Available services:
  im          Instant messaging service

Use "cipher <service> help" for service-specific help.
Use "cipher version" for version information.`)
}

// printVersion prints the cipher product version information.
func printVersion() {
	// Version information should be injected from the calling binary.
	fmt.Println("cipher product (cryptoutil)")
}
