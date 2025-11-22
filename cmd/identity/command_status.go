// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of identity services",
		Long: `Show running status, PID, uptime, and health of all identity services.

Examples:
  # Show status table
  identity status

  # Show status as JSON
  identity status --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("JSON output: %v\n", jsonOutput)

			// TODO: Implement status checking logic
			// 1. Check PID files or Docker container status
			// 2. Query /health endpoints
			// 3. Format output table (or JSON with --json flag)
			//
			// Example output:
			// SERVICE   STATUS    PID     UPTIME   HEALTH
			// authz     running   12345   1h23m    healthy
			// idp       running   12346   1h23m    healthy
			// rs        running   12347   1h23m    healthy

			return fmt.Errorf("status command not yet implemented")
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output status as JSON")

	return cmd
}
