// Copyright (c) 2025 Justin Cranford
//
//

package cmd

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	cryptoutilIdentityProcess "cryptoutil/internal/apps/identity/process"
)

// NewStatusCommand creates a status command for identity services.
func NewStatusCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   cryptoutilSharedMagic.StringStatus,
		Short: "Show status of identity services",
		Long: `Show running status, PID, uptime, and health of all identity services.

Examples:
  # Show status table
  identity status

  # Show status as JSON
  identity status --json`,
		RunE: func(_ *cobra.Command, _ []string) error {
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

			// Check status of each service
			services := []string{cryptoutilSharedMagic.AuthzServiceName, cryptoutilSharedMagic.IDPServiceName, "rs"}

			type ServiceStatus struct {
				Name    string `json:"name"`
				Running bool   `json:"running"`
				PID     int    `json:"pid,omitempty"`
			}

			statuses := make([]ServiceStatus, 0, len(services))
			for _, svc := range services {
				status := ServiceStatus{Name: svc}
				if procManager.IsRunning(svc) {
					status.Running = true

					pid, pidErr := procManager.GetPID(svc)
					if pidErr == nil {
						status.PID = pid
					}
				}

				statuses = append(statuses, status)
			}

			// Output results
			if jsonOutput {
				jsonBytes, err := json.MarshalIndent(statuses, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}

				fmt.Println(string(jsonBytes))
			} else {
				fmt.Println("SERVICE   STATUS      PID")

				for _, s := range statuses {
					statusStr := "stopped"
					pidStr := "-"

					if s.Running {
						statusStr = cryptoutilSharedMagic.DockerServiceStateRunning

						if s.PID > 0 {
							pidStr = fmt.Sprintf("%d", s.PID)
						}
					}

					fmt.Printf("%-9s %-11s %s\n", s.Name, statusStr, pidStr)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output status as JSON")

	return cmd
}
