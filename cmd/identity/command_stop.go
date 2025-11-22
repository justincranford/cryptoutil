// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStopCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "stop [services...]",
		Short: "Stop identity services",
		Long: `Stop one or more identity services gracefully.

Examples:
  # Stop all services
  identity stop

  # Stop specific services
  identity stop authz

  # Force stop (no graceful shutdown)
  identity stop --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			services := args
			if len(services) == 0 {
				services = []string{"authz", "idp", "rs"}
			}

			fmt.Printf("Stopping services: %v\n", services)
			fmt.Printf("Force: %v\n", force)

			// TODO: Implement service shutdown logic
			// 1. If Docker mode active: docker compose down <services>
			// 2. If local processes: Send SIGTERM to PIDs
			// 3. Wait for graceful shutdown (default 10s timeout)
			// 4. If --force: Send SIGKILL

			return fmt.Errorf("stop command not yet implemented")
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force stop without graceful shutdown")

	return cmd
}
