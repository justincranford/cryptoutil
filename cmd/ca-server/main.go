// Copyright (c) 2025 Justin Cranford

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cryptoutilCACmd "cryptoutil/internal/ca/server/cmd"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ca-server",
		Short: "CA Server",
		Long: `CA Server - Certificate Authority REST API.

Provides certificate issuance, revocation, and status services.

API Endpoints:
  /api/v1/ca/ca                          - List CAs
  /api/v1/ca/ca/{caId}                   - Get CA details
  /api/v1/ca/ca/{caId}/crl               - Download CRL
  /api/v1/ca/enroll                      - Submit enrollment request
  /api/v1/ca/certificates                - List certificates
  /api/v1/ca/certificates/{serialNumber} - Get/revoke certificate
  /api/v1/ca/profiles                    - List certificate profiles
  /api/v1/ca/ocsp                        - OCSP responder`,
	}

	rootCmd.AddCommand(cryptoutilCACmd.NewStartCommand())
	rootCmd.AddCommand(cryptoutilCACmd.NewHealthCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
