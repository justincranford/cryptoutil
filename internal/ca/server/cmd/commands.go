// Copyright (c) 2025 Justin Cranford

// Package cmd provides CLI commands for the CA Server.
package cmd

import (
	"context"
	"fmt"
	http "net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilCAServer "cryptoutil/internal/ca/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// NewStartCommand creates the start command for the CA server.
func NewStartCommand() *cobra.Command {
	var (
		configFiles []string
		bindAddr    string
		bindPort    uint16
		devMode     bool
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the CA Server",
		Long: `Start the CA Server with the specified configuration.

Examples:
  # Start with default settings (dev mode)
  pki-ca start --dev

  # Start with custom config file
  pki-ca start --config pki-ca.yml

  # Start with multiple config files (merged in order)
  pki-ca start --config ca-common.yml --config ca-instance.yml --config ca-otel.yml

  # Start with custom bind address and port
  pki-ca start --bind 0.0.0.0 --port 8091`,
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

			// Load configuration.
			var settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings

			var err error

			if len(configFiles) > 0 {
				// Build args with multiple --config flags for Parse()
				parseArgs := []string{"start"} // Subcommand required by Parse()
				for _, cf := range configFiles {
					parseArgs = append(parseArgs, "--config", cf)
				}

				settings, err = cryptoutilAppsTemplateServiceConfig.Parse(parseArgs, false)
				if err != nil {
					return fmt.Errorf("failed to load config files: %w", err)
				}
			} else {
				// Use default settings for CA server.
				settings = cryptoutilAppsTemplateServiceConfig.NewForCAServer(bindAddr, bindPort, devMode)
			}

			// Create and start the server.
			server, err := cryptoutilCAServer.NewServer(ctx, settings)
			if err != nil {
				return fmt.Errorf("failed to create CA server: %w", err)
			}

			defer func() {
				if shutdownErr := server.Shutdown(); shutdownErr != nil {
					fmt.Printf("Server shutdown error: %v\n", shutdownErr)
				}
			}()

			fmt.Printf("CA Server starting on %s:%d\n", settings.BindPublicAddress, settings.BindPublicPort)

			// Start the server (blocks until context is cancelled).
			if err := server.Start(ctx); err != nil {
				return fmt.Errorf("server error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&configFiles, "config", "c", nil, "Path to configuration file (can be specified multiple times)")
	cmd.Flags().StringVarP(&bindAddr, "bind", "b", cryptoutilSharedMagic.IPv4Loopback, "Bind address")
	cmd.Flags().Uint16VarP(&bindPort, "port", "p", cryptoutilSharedMagic.DefaultPublicPortCAServer, "Bind port")
	cmd.Flags().BoolVar(&devMode, "dev", false, "Enable development mode (relaxed security)")

	return cmd
}

// NewHealthCommand creates the health check command.
func NewHealthCommand() *cobra.Command {
	var serverURL string

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check CA server health",
		Long:  "Send a health check request to the CA Server.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			// Check liveness probe.
			livezURL := serverURL + "/admin/api/v1/livez"

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, livezURL, nil)
			if err != nil {
				return fmt.Errorf("failed to create liveness request: %w", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to check liveness: %w", err)
			}

			_ = resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("liveness check failed with status %d", resp.StatusCode)
			}

			// Check readiness probe.
			readyzURL := serverURL + "/admin/api/v1/readyz"

			req, err = http.NewRequestWithContext(ctx, http.MethodGet, readyzURL, nil)
			if err != nil {
				return fmt.Errorf("failed to create readiness request: %w", err)
			}

			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to check readiness: %w", err)
			}

			_ = resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("readiness check failed with status %d", resp.StatusCode)
			}

			fmt.Printf("CA server at %s is healthy\n", serverURL)

			return nil
		},
	}

	cmd.Flags().StringVarP(&serverURL, "url", "u", "https://127.0.0.1:8091", "CA server URL")

	return cmd
}
