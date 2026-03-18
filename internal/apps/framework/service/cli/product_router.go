// Copyright (c) 2025 Justin Cranford
//

package cli

import (
	"fmt"
	"io"
)

// ProductConfig holds configuration for a product CLI entrypoint.
type ProductConfig struct {
	// ProductName is the product name (e.g., "sm", "jose", "pki", "identity", "skeleton").
	ProductName string
	// UsageText is the complete usage message displayed for --help.
	UsageText string
	// VersionText is the version message displayed for --version.
	VersionText string
}

// ServiceEntry represents a single service within a product.
type ServiceEntry struct {
	// Name is the service subdirectory name (e.g., "im", "ja", "kms", "ca").
	Name string
	// Handler is the function to call when this service is selected.
	Handler func(args []string, stdin io.Reader, stdout, stderr io.Writer) int
}

// RouteProduct implements the standard product command router.
// It handles version/help flags and routes to the appropriate service handler.
//
// Call patterns:
//   - Suite:          cryptoutil <product> <service> <subcommand>
//   - Product:        <product> <service> <subcommand>
//   - Product-Service: <product>-<service> <subcommand> (via main.go delegation)
func RouteProduct(cfg ProductConfig, args []string, stdin io.Reader, stdout, stderr io.Writer, services []ServiceEntry) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(stderr, cfg.UsageText)

		return 1
	}

	// Check for help flags.
	if args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag {
		_, _ = fmt.Fprintln(stderr, cfg.UsageText)

		return 0
	}

	// Check for version flags.
	if args[0] == versionCommand || args[0] == versionFlag || args[0] == versionShortFlag {
		_, _ = fmt.Fprintln(stdout, cfg.VersionText)

		return 0
	}

	// Route to service command.
	for _, svc := range services {
		if args[0] == svc.Name {
			return svc.Handler(args[1:], stdin, stdout, stderr)
		}
	}

	_, _ = fmt.Fprintf(stderr, "Unknown service: %s\n\n", args[0])

	_, _ = fmt.Fprintln(stderr, cfg.UsageText)

	return 1
}
