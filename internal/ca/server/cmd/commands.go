// Copyright (c) 2025 Justin Cranford

// Package cmd provides CLI commands for the CA Server.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	cryptoutilCAServer "cryptoutil/internal/ca/server"
	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// NewStartCommand creates the start command for the CA server.
func NewStartCommand() *cobra.Command {
	var (
		configFile string
		bindAddr   string
		bindPort   uint16
		devMode    bool
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the CA Server",
		Long: `Start the CA Server with the specified configuration.

Examples:
  # Start with default settings (dev mode)
  ca-server start --dev

  # Start with custom config file
  ca-server start --config ca-server.yml

  # Start with custom bind address and port
  ca-server start --bind 0.0.0.0 --port 8091`,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			var settings *cryptoutilConfig.Settings

			var err error

			if configFile != "" {
				settings, err = cryptoutilConfig.NewFromFile(configFile)
				if err != nil {
					return fmt.Errorf("failed to load config file %s: %w", configFile, err)
				}
			} else {
				// Use default settings for CA server.
				settings = cryptoutilConfig.NewForCAServer(bindAddr, bindPort, devMode)
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

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	cmd.Flags().StringVarP(&bindAddr, "bind", "b", cryptoutilMagic.IPv4Loopback, "Bind address")
	cmd.Flags().Uint16VarP(&bindPort, "port", "p", cryptoutilMagic.DefaultPublicPortCAServer, "Bind port")
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement health check client.
			fmt.Printf("Checking health of CA server at %s\n", serverURL)

			return nil
		},
	}

	cmd.Flags().StringVarP(&serverURL, "url", "u", "https://127.0.0.1:8091", "CA server URL")

	return cmd
}
