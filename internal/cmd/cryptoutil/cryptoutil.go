// Copyright (c) 2025 Justin Cranford
//
//

// Package cmd provides command-line interface for cryptoutil operations.
package cmd

import (
	"fmt"
	"os"

	cryptoutilCmdCipher "cryptoutil/internal/cmd/cipher"
	cryptoutilCmdCryptoutilAuthz "cryptoutil/internal/cmd/cryptoutil/authz"
	cryptoutilCmdCryptoutilCa "cryptoutil/internal/cmd/cryptoutil/ca"
	cryptoutilCmdCryptoutilIdentity "cryptoutil/internal/cmd/cryptoutil/identity"
	cryptoutilCmdCryptoutilIdp "cryptoutil/internal/cmd/cryptoutil/idp"
	cryptoutilCmdCryptoutilJose "cryptoutil/internal/cmd/cryptoutil/jose"
	cryptoutilCmdCryptoutilRp "cryptoutil/internal/cmd/cryptoutil/rp"
	cryptoutilCmdCryptoutilRs "cryptoutil/internal/cmd/cryptoutil/rs"
	cryptoutilCmdCryptoutilSpa "cryptoutil/internal/cmd/cryptoutil/spa"
	cryptoutilKmsCmd "cryptoutil/internal/apps/sm/kms/cmd"
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
		cryptoutilCmdCryptoutilIdentity.Execute(parameters)
	case "identity-authz":
		cryptoutilCmdCryptoutilAuthz.Execute(parameters)
	case "identity-idp":
		cryptoutilCmdCryptoutilIdp.Execute(parameters)
	case "identity-rs":
		cryptoutilCmdCryptoutilRs.Execute(parameters)
	case "identity-rp":
		cryptoutilCmdCryptoutilRp.Execute(parameters)
	case "identity-spa":
		cryptoutilCmdCryptoutilSpa.Execute(parameters)
	case "jose":
		cryptoutilCmdCryptoutilJose.Execute(parameters)
	case "ca":
		cryptoutilCmdCryptoutilCa.Execute(parameters)
	case "cipher":
		exitCode := cryptoutilCmdCipher.Cipher(parameters)
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
	fmt.Println("  kms           - Key Management Service")
	fmt.Println("  identity      - Identity Services (legacy unified)")
	fmt.Println("  identity-authz - Identity Authorization Server")
	fmt.Println("  identity-idp   - Identity Provider")
	fmt.Println("  identity-rs    - Identity Resource Server")
	fmt.Println("  identity-rp    - Identity Relying Party (BFF reference implementation)")
	fmt.Println("  identity-spa   - Identity SPA (Single Page Application reference implementation)")
	fmt.Println("  jose          - JOSE Authority")
	fmt.Println("  ca            - Certificate Authority")
	fmt.Println("  cipher        - Cipher services (educational)")
	fmt.Println("  help          - Show this help message")
}
