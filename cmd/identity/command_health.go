// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newHealthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check health of identity services",
		Long: `Poll /health endpoints and report readiness.
Exit 0 if all services healthy, exit 1 if any unhealthy.

Examples:
  # Check health of all services
  identity health`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement health checking logic
			// 1. HTTP GET to https://localhost:{port}/health for each service
			// 2. Parse JSON response: {"status": "healthy", "database": "ok"}
			// 3. Aggregate results
			// 4. Colorized output (green ✅ / red ❌)

			return fmt.Errorf("health command not yet implemented")
		},
	}

	return cmd
}
