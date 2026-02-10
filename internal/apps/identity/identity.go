// Copyright (c) 2025 Justin Cranford
//

// Package identity implements the identity product command router.
package identity

import (
	"fmt"
	"io"

	cryptoutilAppsIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilAppsIdentityIdp "cryptoutil/internal/apps/identity/idp"
	cryptoutilAppsIdentityRp "cryptoutil/internal/apps/identity/rp"
	cryptoutilAppsIdentityRs "cryptoutil/internal/apps/identity/rs"
	cryptoutilAppsIdentitySpa "cryptoutil/internal/apps/identity/spa"
)

const (
	helpCommand      = "help"
	helpFlag         = "--help"
	helpShortFlag    = "-h"
	versionCommand   = "version"
	versionFlag      = "--version"
	versionShortFlag = "-v"
)

// Identity implements the identity product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil identity authz server
// - Product: identity authz server
// - Product-Service: identity-authz server (via main.go delegation).
func Identity(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
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
	case "authz":
		return cryptoutilAppsIdentityAuthz.Authz(args[1:], stdin, stdout, stderr)
	case "idp":
		return cryptoutilAppsIdentityIdp.Idp(args[1:], stdin, stdout, stderr)
	case "rp":
		return cryptoutilAppsIdentityRp.Rp(args[1:], stdin, stdout, stderr)
	case "rs":
		return cryptoutilAppsIdentityRs.Rs(args[1:], stdin, stdout, stderr)
	case "spa":
		return cryptoutilAppsIdentitySpa.Spa(args[1:], stdin, stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown service: %s\n\n", args[0])
		printUsage(stderr)

		return 1
	}
}

// printUsage prints the identity product usage information.
func printUsage(stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, `Usage: identity <service> <subcommand> [options]

Available services:
  authz       OAuth 2.1 Authorization Server
  idp         OIDC 1.0 Identity Provider
  rp          OAuth 2.1 Relying Party (BFF reference implementation)
  rs          OAuth 2.1 Resource Server
  spa         Single Page Application (SPA reference implementation)

Use "identity <service> help" for service-specific help.
Use "identity version" for version information.`)
}

// printVersion prints the identity product version information.
func printVersion(stdout io.Writer) {
	// Version information should be injected from the calling binary.
	_, _ = fmt.Fprintln(stdout, "identity product (cryptoutil)")
}
