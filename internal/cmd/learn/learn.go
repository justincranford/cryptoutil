// Copyright (c) 2025 Justin Cranford
//
//

// Package learn implements the learn product command router.
package learn

import (
	"fmt"
	"os"

	"cryptoutil/internal/cmd/learn/im"
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

// Learn implements the learn product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil learn im server
// - Product: learn im server
// - Product-Service: learn-im server (via main.go delegation).
func Learn(args []string) int {
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

// printUsage prints the learn product usage information.
func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: learn <service> <subcommand> [options]

Available services:
  im          Instant messaging service

Use "learn <service> help" for service-specific help.
Use "learn version" for version information.`)
}

// printVersion prints the learn product version information.
func printVersion() {
	// Version information should be injected from the calling binary.
	fmt.Println("learn product (cryptoutil)")
}
