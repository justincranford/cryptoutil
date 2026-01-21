// Copyright (c) 2025 Justin Cranford
//
//

package cmd

import (
	"fmt"
	"os"

	cryptoutilCipherCmd "cryptoutil/internal/cmd/cipher"
	cryptoutilCACmd "cryptoutil/internal/cmd/cryptoutil/ca"
	cryptoutilIdentityCmd "cryptoutil/internal/cmd/cryptoutil/identity"
	cryptoutilJoseCmd "cryptoutil/internal/cmd/cryptoutil/jose"
	cryptoutilKmsCmd "cryptoutil/internal/kms/cmd"
)

// Execute runs the cryptoutil command-line interface.
func Execute() {
	executable := os.Args[0] // Example executable: ./cryptoutil
	if len(os.Args) < 2 {
		printUsage(executable)
		os.Exit(1)
	}

	// TODO product := os.Args[1] // Example products: sm, identity, jose, pki, cipher
	service := os.Args[1]     // Example services: kms, ca, ja, im, authz, idp, rs, rp, spa
	parameters := os.Args[2:] // Example parameters: --config-file, --port, --host, etc.

	// TODO fix the switch values to use product values, and call corresponding PRODUCT.go
	// current implementation is a mismatch of product and service names to be fixed later

	switch service {
	case "kms":
		cryptoutilKmsCmd.Server(parameters)
	case "identity":
		cryptoutilIdentityCmd.Execute(parameters)
	case "jose":
		cryptoutilJoseCmd.Execute(parameters)
	case "ca":
		cryptoutilCACmd.Execute(parameters)
	case "cipher":
		exitCode := cryptoutilCipherCmd.Cipher(parameters)
		os.Exit(exitCode)
	case "help":
		printUsage(executable)
	default:
		printUsage(executable)
		fmt.Printf("Unknown command: %s %s\n", executable, service)
		os.Exit(1)
	}
}

func printUsage(executable string) {
	fmt.Printf("Usage: %s <product> [options]\n", executable)
	fmt.Println("  kms      - Key Management Service")
	fmt.Println("  identity - Identity Services (authz, idp, rs)")
	fmt.Println("  jose     - JOSE Authority")
	fmt.Println("  ca       - Certificate Authority")
	fmt.Println("  learn    - Educational and demonstration services")
	fmt.Println("  help     - Show this help message")
}
