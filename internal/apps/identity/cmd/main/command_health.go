// Copyright (c) 2025 Justin Cranford
//
//

// Package cmd provides command-line interface for the identity service.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	cryptoutilIdentityHealthcheck "cryptoutil/internal/apps/identity/healthcheck"
)

const (
	defaultMaxRetries = 3
)

// NewHealthCommand creates a health check command for identity services.
func NewHealthCommand() *cobra.Command {
	var timeoutStr string

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check health of identity services",
		Long: `Poll /health endpoints and report readiness.
Exit 0 if all services healthy, exit 1 if any unhealthy.

Examples:
  # Check health of all services
  identity health`,
		RunE: func(_ *cobra.Command, _ []string) error {
			timeout, err := time.ParseDuration(timeoutStr)
			if err != nil {
				return fmt.Errorf("invalid timeout: %w", err)
			}

			// Health endpoints for all services
			healthURLs := []struct {
				name string
				url  string
			}{
				{"authz", "https://127.0.0.1:8080/health"},
				{"idp", "https://127.0.0.1:8081/health"},
				{"rs", "https://127.0.0.1:8082/health"},
			}

			// In development/local environments, skip TLS verification for self-signed certs.
			// TODO: Make this configurable via command flag for production deployments.
			skipTLSVerify := true
			poller := cryptoutilIdentityHealthcheck.NewPoller(timeout, defaultMaxRetries, skipTLSVerify)
			allHealthy := true

			for _, health := range healthURLs {
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				resp, pollErr := poller.Poll(ctx, health.url)
				if pollErr != nil {
					fmt.Printf("❌ %s: unhealthy (%v)\n", health.name, pollErr)

					allHealthy = false

					continue
				}

				if resp.Status == "healthy" {
					fmt.Printf("✅ %s: %s", health.name, resp.Status)

					if resp.Database != "" {
						fmt.Printf(" (database: %s)", resp.Database)
					}

					fmt.Println()
				} else {
					fmt.Printf("❌ %s: %s\n", health.name, resp.Status)

					allHealthy = false
				}
			}

			if !allHealthy {
				return fmt.Errorf("one or more services unhealthy")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&timeoutStr, "timeout", "5s", "Health check timeout per service")

	return cmd
}
