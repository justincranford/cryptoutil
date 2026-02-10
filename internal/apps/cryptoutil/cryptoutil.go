// Copyright (c) 2025 Justin Cranford
//
//

// Package cryptoutil provides command-line interface for cryptoutil suite operations.
package cryptoutil

import (
	"fmt"
	"io"

	cryptoutilAppsCipher "cryptoutil/internal/apps/cipher"
	cryptoutilAppsIdentity "cryptoutil/internal/apps/identity"
	cryptoutilAppsJose "cryptoutil/internal/apps/jose"
	cryptoutilAppsPki "cryptoutil/internal/apps/pki"
	cryptoutilAppsSm "cryptoutil/internal/apps/sm"
)

// Suite runs the cryptoutil suite command-line interface.
// This is the entry point for the suite-level CLI.
func Suite(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		printUsage(stderr)

		return 1
	}

	product := args[1]     // Example products: sm, identity, jose, pki, cipher
	parameters := args[2:] // Example parameters: service names, --config-file, --port, --host, etc.

	// Route to product command.
	switch product {
	case "cipher":
		return cryptoutilAppsCipher.Cipher(parameters, stdin, stdout, stderr)
	case "identity":
		return cryptoutilAppsIdentity.Identity(parameters, stdin, stdout, stderr)
	case "jose":
		return cryptoutilAppsJose.Jose(parameters, stdin, stdout, stderr)
	case "pki":
		return cryptoutilAppsPki.Pki(parameters, stdin, stdout, stderr)
	case "sm":
		return cryptoutilAppsSm.Sm(parameters, stdin, stdout, stderr)
	case "help", "--help", "-h":
		printUsage(stderr)

		return 0
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown product: %s\n\n", product)
		printUsage(stderr)

		return 1
	}
}

// printUsage prints the cryptoutil suite usage information.
func printUsage(stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, `Usage: cryptoutil <product> [service] [options]

Available products:
  cipher      Cipher product (encrypted messaging)
  identity    Identity product (OAuth 2.1, OIDC 1.0)
  jose        JOSE product (JWK/JWS/JWE/JWT operations)
  pki         PKI product (X.509 certificates, CA)
  sm          Secrets Manager product (KMS)

Use "cryptoutil <product> help" for product-specific help.`)
}
