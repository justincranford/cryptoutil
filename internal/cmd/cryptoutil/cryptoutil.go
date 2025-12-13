// Copyright (c) 2025 Justin Cranford
//
//

package cmd

import (
	"fmt"
	"os"

	cryptoutilIdentityCmd "cryptoutil/internal/cmd/cryptoutil/identity"
	cryptoutilJoseCmd "cryptoutil/internal/cmd/cryptoutil/jose"
	cryptoutilKmsCmd "cryptoutil/internal/kms/cmd"
)

func Execute() {
	executable := os.Args[0] // Example executable: ./cryptoutil
	if len(os.Args) < 2 {
		printUsage(executable)
		os.Exit(1)
	}

	product := os.Args[1]     // Example product: kms, identity, jose, ca
	parameters := os.Args[2:] // Example parameters: --config-file, --port, --host, etc.

	switch product {
	case "kms":
		cryptoutilKmsCmd.Server(parameters)
	case "identity":
		cryptoutilIdentityCmd.Execute(parameters)
	case "jose":
		cryptoutilJoseCmd.Execute(parameters)
	case "help":
		printUsage(executable)
	default:
		printUsage(executable)
		fmt.Printf("Unknown command: %s %s\n", executable, product)
		os.Exit(1)
	}
}

func printUsage(executable string) {
	fmt.Printf("Usage: %s <product> [options]\n", executable)
	fmt.Println("  kms      - Key Management Service")
	fmt.Println("  identity - Identity Services (authz, idp, rs)")
	fmt.Println("  jose     - JOSE Authority")
	fmt.Println("  ca       - Certificate Authority (planned)")
}
