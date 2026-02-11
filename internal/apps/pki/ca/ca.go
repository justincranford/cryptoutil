// Copyright (c) 2025 Justin Cranford
//
//

package ca

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	cryptoutilAppsPkiCaServerCmd "cryptoutil/internal/apps/pki/ca/server/cmd"
)

// Ca is the entry point for the pki-ca service.
func Ca(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// Create root command.
	rootCmd := &cobra.Command{
		Use:   "pki-ca",
		Short: "CA (Certificate Authority) service for PKI product",
		Long:  "Provides certificate issuance, management, and validation services",
	}

	// Set IO streams.
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	// Add subcommands.
	rootCmd.AddCommand(cryptoutilAppsPkiCaServerCmd.NewStartCommand())
	rootCmd.AddCommand(cryptoutilAppsPkiCaServerCmd.NewHealthCommand())

	// Set args (skip program name if provided).
	if len(args) > 0 && args[0] == "pki-ca" {
		args = args[1:]
	}

	rootCmd.SetArgs(args)

	// Execute command.
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
	}

	return 0
}
