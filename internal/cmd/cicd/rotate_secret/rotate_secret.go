// Copyright (c) 2025 Justin Cranford

package rotate_secret

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
	cryptoutilIdentityRotation "cryptoutil/internal/identity/rotation"
)

// Config holds the CLI configuration for secret rotation.
type Config struct {
	ClientID     string
	GracePeriod  time.Duration
	Reason       string
	OutputFormat string
}

// Execute runs the rotate-secret command.
func Execute(logger *cryptoutilCmdCicdCommon.Logger, args []string) error {
	cfg, err := parseFlags(args)
	if err != nil {
		return fmt.Errorf("flag parsing error: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("configuration validation error: %w", err)
	}

	// Create database connection.
	db, err := setupDatabase()
	if err != nil {
		return fmt.Errorf("database setup error: %w", err)
	}

	sqlDB, _ := db.DB()

	defer func() { _ = sqlDB.Close() }()

	// Create rotation service.
	service := cryptoutilIdentityRotation.NewSecretRotationService(db)

	// Parse client ID.
	clientID, err := googleUuid.Parse(cfg.ClientID)
	if err != nil {
		return fmt.Errorf("invalid client ID format: %w", err)
	}

	// Perform rotation.
	ctx := context.Background()

	const defaultInitiator = "cli-tool"

	result, err := service.RotateClientSecret(ctx, clientID, cfg.GracePeriod, defaultInitiator, cfg.Reason)
	if err != nil {
		return fmt.Errorf("rotation failed: %w", err)
	}

	// Output result.
	if err := outputResult(result, cfg.OutputFormat); err != nil {
		return fmt.Errorf("output error: %w", err)
	}

	logger.Log("Secret rotation completed successfully")

	return nil
}

func parseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("rotate-secret", flag.ContinueOnError)

	cfg := &Config{}

	const (
		defaultGracePeriod  = 24 * time.Hour
		defaultOutputFormat = "text"
	)

	fs.StringVar(&cfg.ClientID, "client-id", "", "Client UUID (required)")
	fs.DurationVar(&cfg.GracePeriod, "grace-period", defaultGracePeriod, "Grace period for old secret (default: 24h)")
	fs.StringVar(&cfg.Reason, "reason", "", "Rotation reason for audit trail")
	fs.StringVar(&cfg.OutputFormat, "output", defaultOutputFormat, "Output format: text or json (default: text)")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.ClientID == "" {
		return fmt.Errorf("--client-id is required")
	}

	if _, err := googleUuid.Parse(cfg.ClientID); err != nil {
		return fmt.Errorf("invalid client ID format (expected UUID): %w", err)
	}

	if cfg.GracePeriod < 0 {
		return fmt.Errorf("grace period must be non-negative")
	}

	const (
		outputFormatText = "text"
		outputFormatJSON = "json"
	)
	if cfg.OutputFormat != outputFormatText && cfg.OutputFormat != outputFormatJSON {
		return fmt.Errorf("output format must be 'text' or 'json'")
	}

	return nil
}

func outputResult(result *cryptoutilIdentityRotation.RotateClientSecretResult, format string) error {
	switch format {
	case "json":
		return outputJSON(result)
	case "text":
		return outputText(result)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func outputJSON(result *cryptoutilIdentityRotation.RotateClientSecretResult) error {
	output := map[string]any{
		"old_version":      result.OldVersion,
		"new_version":      result.NewVersion,
		"new_secret":       result.NewSecretPlaintext,
		"grace_period_end": result.GracePeriodEnd.Format(time.RFC3339),
		"event_id":         result.EventID.String(),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	if err := enc.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON output: %w", err)
	}

	return nil
}

func outputText(result *cryptoutilIdentityRotation.RotateClientSecretResult) error {
	fmt.Println("Client Secret Rotation Complete")
	fmt.Printf("Old Version: %d\n", result.OldVersion)
	fmt.Printf("New Version: %d\n", result.NewVersion)
	fmt.Printf("New Secret: %s\n", result.NewSecretPlaintext)
	fmt.Printf("Grace Period Ends: %s\n", result.GracePeriodEnd.Format(time.RFC3339))
	fmt.Printf("Event ID: %s\n", result.EventID.String())

	return nil
}
