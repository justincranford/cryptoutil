// Copyright (c) 2025 Justin Cranford
//
//

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	cryptoutilIdentityBootstrap "cryptoutil/internal/identity/bootstrap"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityHealthcheck "cryptoutil/internal/identity/healthcheck"
	cryptoutilIdentityProcess "cryptoutil/internal/identity/process"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

const (
	healthCheckMaxRetries = 10
)

// NewStartCommand creates a new cobra command for starting identity services.
func NewStartCommand() *cobra.Command {
	var (
		profile    string
		useDocker  bool
		useLocal   bool
		configFile string
		background bool
		wait       bool
		timeout    string
		resetDemo  bool
	)

	cmd := &cobra.Command{
		Use:   "start [services...]",
		Short: "Start identity services",
		Long: `Start one or more identity services (authz, idp, rs).
Defaults to starting all services with demo profile.

Examples:
  # Start all services with demo profile
  identity start --profile demo

  # Start specific services
  identity start authz idp --profile ci

  # Start with custom config
  identity start --config custom.yml

  # Start in Docker
  identity start --profile full-stack --docker

  # Start in background (local processes)
  identity start --background`,
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := context.Background()

			// services := args  // Not used - would filter servicesToStart if implemented
			// if len(services) == 0 {
			//	services = []string{"authz", "idp", "rs"}
			// }

			// Load profile configuration
			var (
				profileCfg *cryptoutilIdentityConfig.ProfileConfig
				err        error
			)

			if configFile != "" {
				// Load from custom config file
				profileCfg, err = cryptoutilIdentityConfig.LoadProfileFromFile(configFile)
				if err != nil {
					return fmt.Errorf("failed to load config file %s: %w", configFile, err)
				}
			} else {
				// Load from profile name
				profileCfg, err = cryptoutilIdentityConfig.LoadProfile(profile)
				if err != nil {
					return fmt.Errorf("failed to load profile %s: %w", profile, err)
				}
			}

			// Validate profile configuration
			if err := profileCfg.Validate(); err != nil {
				return fmt.Errorf("invalid profile configuration: %w", err)
			}

			// Reset demo data if requested
			if resetDemo {
				fmt.Println("Resetting demo data...")

				// Create repository factory to reset data
				dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
					Type: "sqlite",
					DSN:  ":memory:", // TODO: Use actual database config from profile
				}

				repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
				if err != nil {
					return fmt.Errorf("failed to create repository factory for reset: %w", err)
				}

				defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Cleanup on exit

				// Run migrations
				if err := repoFactory.AutoMigrate(ctx); err != nil {
					return fmt.Errorf("failed to auto-migrate for reset: %w", err)
				}

				// Reset and reseed demo data
				if err := cryptoutilIdentityBootstrap.ResetAndReseedDemo(ctx, repoFactory); err != nil {
					return fmt.Errorf("failed to reset demo data: %w", err)
				}

				fmt.Println("Demo data reset successfully!")
			}

			// Parse timeout duration
			timeoutDuration, err := time.ParseDuration(timeout)
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

			// Start services based on profile
			servicesToStart := []struct {
				name       string
				enabled    bool
				binary     string
				configFile string
			}{
				{"authz", profileCfg.Services.AuthZ.Enabled, "bin/authz", "configs/identity/authz.yml"},
				{"idp", profileCfg.Services.IDP.Enabled, "bin/idp", "configs/identity/idp.yml"},
				{"rs", profileCfg.Services.RS.Enabled, "bin/rs", "configs/identity/rs.yml"},
			}

			for _, svc := range servicesToStart {
				if !svc.enabled {
					fmt.Printf("Skipping %s (disabled in profile)\n", svc.name)

					continue
				}

				fmt.Printf("Starting %s...\n", svc.name)

				args := []string{"--config", svc.configFile}
				if err := procManager.Start(ctx, svc.name, svc.binary, args); err != nil {
					return fmt.Errorf("failed to start %s: %w", svc.name, err)
				}
			}

			// Wait for health checks if requested
			if wait {
				// In development/local environments, skip TLS verification for self-signed certs.
				// TODO: Make this configurable via config file for production deployments.
				skipTLSVerify := true
				poller := cryptoutilIdentityHealthcheck.NewPoller(timeoutDuration, healthCheckMaxRetries, skipTLSVerify)
				healthURLs := []struct {
					name string
					url  string
				}{
					{"authz", "https://" + profileCfg.Services.AuthZ.BindAddress + "/health"},
					{"idp", "https://" + profileCfg.Services.IDP.BindAddress + "/health"},
					{"rs", "https://" + profileCfg.Services.RS.BindAddress + "/health"},
				}

				for _, health := range healthURLs {
					if !isServiceEnabled(health.name, servicesToStart) {
						continue
					}

					fmt.Printf("Waiting for %s health check...\n", health.name)

					ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
					defer cancel()

					resp, pollErr := poller.Poll(ctx, health.url)
					if pollErr != nil {
						return fmt.Errorf("%s health check failed: %w", health.name, pollErr)
					}

					fmt.Printf("  %s: %s\n", health.name, resp.Status)
				}
			}

			fmt.Println("All services started successfully!")

			return nil
		},
	}

	cmd.Flags().StringVarP(&profile, "profile", "p", "demo", "Configuration profile (demo, authz-only, authz-idp, full-stack, ci)")
	cmd.Flags().BoolVar(&useDocker, "docker", false, "Use Docker Compose orchestration")
	cmd.Flags().BoolVar(&useLocal, "local", true, "Run services as local processes (default)")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Override with custom config file")
	cmd.Flags().BoolVarP(&background, "background", "d", false, "Detach services to background")
	cmd.Flags().BoolVar(&wait, "wait", true, "Wait for health checks before returning")
	cmd.Flags().StringVar(&timeout, "timeout", "30s", "Health check timeout")
	cmd.Flags().BoolVar(&resetDemo, "reset-demo", false, "Reset demo data before starting services")

	return cmd
}

// isServiceEnabled checks if a service is enabled in the services-to-start list.
func isServiceEnabled(serviceName string, services []struct {
	name       string
	enabled    bool
	binary     string
	configFile string
},
) bool {
	for _, svc := range services {
		if svc.name == serviceName {
			return svc.enabled
		}
	}

	return false
}
