// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cryptoutilJoseCmd "cryptoutil/internal/jose/server/cmd"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "jose-server",
		Short: "JOSE Authority Server",
		Long: `JOSE Authority Server - Standalone REST API for JOSE operations.

Provides JWK key management, JWS signing/verification, JWE encryption/decryption,
and JWT creation/verification operations.

API Endpoints:
  /jose/v1/jwk/generate    - Generate new JWK
  /jose/v1/jwk/{kid}       - Get/Delete JWK by Key ID
  /jose/v1/jwk             - List all JWKs
  /jose/v1/jwks            - Get JWKS (public keys)
  /jose/v1/jws/sign        - Sign payload
  /jose/v1/jws/verify      - Verify JWS signature
  /jose/v1/jwe/encrypt     - Encrypt payload
  /jose/v1/jwe/decrypt     - Decrypt JWE message
  /jose/v1/jwt/create      - Create JWT
  /jose/v1/jwt/verify      - Verify JWT`,
	}

	rootCmd.AddCommand(cryptoutilJoseCmd.NewStartCommand())
	rootCmd.AddCommand(cryptoutilJoseCmd.NewHealthCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
