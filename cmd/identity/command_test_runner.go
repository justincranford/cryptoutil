// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTestCommand() *cobra.Command {
	var suite string
	var pkg string

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run test suites",
		Long: `Run unit, integration, or e2e tests.

Examples:
  # Run all tests
  identity test

  # Run specific suite
  identity test --suite unit
  identity test --suite integration
  identity test --suite e2e

  # Run specific packages
  identity test --package ./internal/identity/authz/...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Suite: %s\n", suite)
			fmt.Printf("Package: %s\n", pkg)

			// TODO: Implement test execution logic
			// 1. Execute go test with appropriate flags
			// 2. For e2e tests: Start services if not running, run tests, stop services
			// 3. Stream output to stdout
			// 4. Return go test exit code

			return fmt.Errorf("test command not yet implemented")
		},
	}

	cmd.Flags().StringVar(&suite, "suite", "", "Test suite (unit, integration, e2e)")
	cmd.Flags().StringVar(&pkg, "package", "", "Specific package to test")

	return cmd
}
