// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cryptoutilIdentityCmd "cryptoutil/internal/identity/cmd/main"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "identity",
		Short: "Unified identity services CLI",
		Long:  "Manage OAuth 2.1 Authorization Server, OIDC Identity Provider, and Resource Server",
	}

	rootCmd.AddCommand(cryptoutilIdentityCmd.NewStartCommand())
	rootCmd.AddCommand(cryptoutilIdentityCmd.NewStopCommand())
	rootCmd.AddCommand(cryptoutilIdentityCmd.NewStatusCommand())
	rootCmd.AddCommand(cryptoutilIdentityCmd.NewHealthCommand())
	rootCmd.AddCommand(cryptoutilIdentityCmd.NewTestCommand())
	rootCmd.AddCommand(cryptoutilIdentityCmd.NewLogsCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
