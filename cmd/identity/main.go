// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "identity",
		Short: "Unified identity services CLI",
		Long:  "Manage OAuth 2.1 Authorization Server, OIDC Identity Provider, and Resource Server",
	}

	rootCmd.AddCommand(newStartCommand())
	rootCmd.AddCommand(newStopCommand())
	rootCmd.AddCommand(newStatusCommand())
	rootCmd.AddCommand(newHealthCommand())
	rootCmd.AddCommand(newTestCommand())
	rootCmd.AddCommand(newLogsCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
