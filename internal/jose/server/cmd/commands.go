// Copyright (c) 2025 Justin Cranford
//
//

// Package cmd provides CLI commands for the JOSE Authority Server.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilJoseServer "cryptoutil/internal/jose/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// NewStartCommand creates the start command for the JOSE server.
func NewStartCommand() *cobra.Command {
	var (
		configFiles []string
		bindAddr    string
		bindPort    uint16
		devMode     bool
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the JOSE Authority Server",
		Long: `Start the JOSE Authority Server with the specified configuration.

Examples:
  # Start with default settings (dev mode)
  jose-server start --dev

  # Start with custom config file
  jose-server start --config jose-server.yml

  # Start with multiple config files (merged in order)
  jose-server start --config jose-common.yml --config jose-instance.yml --config jose-otel.yml

  # Start with custom bind address and port
  jose-server start --bind 0.0.0.0 --port 8090`,
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle interrupt signals.
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				fmt.Println("\nReceived interrupt signal, shutting down...")
				cancel()
			}()

			// Load configuration - avoid config.Parse() to prevent CLI flag conflicts.
			// JOSE uses simple settings without full config parsing.
			settings := cryptoutilAppsTemplateServiceConfig.NewForJOSEServer(bindAddr, bindPort, devMode)

			// Override with config files if provided.
			if len(configFiles) > 0 {
				// Build args with multiple --config flags for LoadConfig().
				parseArgs := []string{"start"} // Subcommand required.
				for _, cf := range configFiles {
					parseArgs = append(parseArgs, "--config", cf)
				}

				loadedSettings, err := cryptoutilAppsTemplateServiceConfig.Parse(parseArgs, false)
				if err != nil {
					return fmt.Errorf("failed to load config files: %w", err)
				}

				// Merge loaded settings into defaults.
				settings = loadedSettings
			}

			// Create TLS config for JOSE server.
			tlsCfg, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
				[]string{},
				[]string{"127.0.0.1", "::1"},
				cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
			)
			if err != nil {
				return fmt.Errorf("failed to generate TLS config: %w", err)
			}

			// Create and start the server.
			server, err := cryptoutilJoseServer.NewServer(ctx, settings, tlsCfg)
			if err != nil {
				return fmt.Errorf("failed to create JOSE server: %w", err)
			}

			defer func() {
				if shutdownErr := server.Shutdown(); shutdownErr != nil {
					fmt.Printf("Server shutdown error: %v\n", shutdownErr)
				}
			}()

			fmt.Printf("JOSE Authority Server starting on %s:%d\n", settings.BindPublicAddress, settings.BindPublicPort)

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&configFiles, "config", "c", nil, "Path to configuration file (can be specified multiple times)")
	cmd.Flags().StringVarP(&bindAddr, "bind", "b", cryptoutilSharedMagic.IPv4Loopback, "Bind address")
	cmd.Flags().Uint16VarP(&bindPort, "port", "p", cryptoutilSharedMagic.DefaultPublicPortJOSEServer, "Bind port")
	cmd.Flags().BoolVar(&devMode, "dev", false, "Enable development mode (relaxed security)")

	return cmd
}

// NewHealthCommand creates the health check command.
func NewHealthCommand() *cobra.Command {
	var serverURL string

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check JOSE server health",
		Long:  "Send a health check request to the JOSE Authority Server.",
		RunE: func(_ *cobra.Command, _ []string) error {
			// NOTE: Health check client will be implemented when remote health checking is needed.
			// For local health checks, use admin endpoints directly (e.g., curl https://127.0.0.1:9090/admin/api/v1/livez).
			fmt.Printf("Checking health of JOSE server at %s\n", serverURL)

			return nil
		},
	}

	cmd.Flags().StringVarP(&serverURL, "url", "u", "https://127.0.0.1:8090", "JOSE server URL")

	return cmd
}
