// Copyright (c) 2025 Justin Cranford
//
//

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	cryptoutilIdentityProcess "cryptoutil/internal/identity/process"
)

// NewStopCommand creates a stop command for identity services.
func NewStopCommand() *cobra.Command {
	var (
		force      bool
		timeoutStr string
	)

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
		RunE: func(_ *cobra.Command, args []string) error {
			services := args

			// Parse timeout
			timeout, err := time.ParseDuration(timeoutStr)
			if err != nil {
				return fmt.Errorf("invalid timeout: %w", err)
			}

			// Create process manager
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			pidDir := filepath.Join(homeDir, ".identity", "pids")

			procManager, err := cryptoutilIdentityProcess.NewManager(pidDir)
			if err != nil {
				return fmt.Errorf("failed to create process manager: %w", err)
			}

			// Stop all services or specific services
			if len(services) == 0 {
				fmt.Println("Stopping all services...")

				if err := procManager.StopAll(force, timeout); err != nil {
					return fmt.Errorf("failed to stop services: %w", err)
				}
			} else {
				for _, svc := range services {
					fmt.Printf("Stopping %s...\n", svc)

					if err := procManager.Stop(svc, force, timeout); err != nil {
						return fmt.Errorf("failed to stop %s: %w", svc, err)
					}
				}
			}

			fmt.Println("All services stopped successfully!")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force stop without graceful shutdown")
	cmd.Flags().StringVar(&timeoutStr, "timeout", "10s", "Graceful shutdown timeout")

	return cmd
}
