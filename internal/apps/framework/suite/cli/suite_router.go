// Copyright (c) 2025 Justin Cranford
//

// Package cli provides the suite-level command router.
// The suite CLI entrypoint uses RouteSuite from this package.
package cli

import (
	"fmt"
	"io"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// SuiteConfig holds configuration for a suite CLI entrypoint.
type SuiteConfig struct {
	// UsageText is the complete usage message displayed for --help or unknown products.
	UsageText string
}

// ProductEntry represents a single product within the suite.
type ProductEntry struct {
	// Name is the product identifier (e.g., "sm", "jose", "pki", "identity", "skeleton", "pki-init").
	Name string
	// Handler is the function to call when this product is selected.
	Handler func(args []string, stdin io.Reader, stdout, stderr io.Writer) int
}

// RouteSuite implements the standard suite command router.
// It handles help flags and routes to the appropriate product handler.
//
// Call pattern:
//
// cryptoutil <product> <service> <subcommand>.
func RouteSuite(cfg SuiteConfig, args []string, stdin io.Reader, stdout, stderr io.Writer, products []ProductEntry) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(stderr, cfg.UsageText)

		return 1
	}

	// Check for help flags.
	if args[0] == cryptoutilSharedMagic.CLIHelpCommand || args[0] == cryptoutilSharedMagic.CLIHelpFlag || args[0] == cryptoutilSharedMagic.CLIHelpShortFlag {
		_, _ = fmt.Fprintln(stderr, cfg.UsageText)

		return 0
	}

	// Route to product command.
	for _, p := range products {
		if args[0] == p.Name {
			return p.Handler(args[1:], stdin, stdout, stderr)
		}
	}

	_, _ = fmt.Fprintf(stderr, "Unknown product: %s\n\n", args[0])
	_, _ = fmt.Fprintln(stderr, cfg.UsageText)

	return 1
}
