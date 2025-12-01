// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides demo data seeding functionality for the KMS server.
package demo

import (
	"context"
	"fmt"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilBusinessLogic "cryptoutil/internal/server/businesslogic"
)

// DemoKeyConfig defines a demo key to be seeded.
type DemoKeyConfig struct {
	Name        string
	Description string
	Algorithm   cryptoutilOpenapiModel.ElasticKeyAlgorithm
}

// DefaultDemoKeys returns the default set of demo keys to seed.
func DefaultDemoKeys() []DemoKeyConfig {
	return []DemoKeyConfig{
		{
			Name:        "demo-encryption-aes256",
			Description: "Demo AES-256-GCM encryption key",
			Algorithm:   cryptoutilOpenapiModel.A256GCMDir,
		},
		{
			Name:        "demo-signing-rsa2048",
			Description: "Demo RSA-2048 signing key",
			Algorithm:   cryptoutilOpenapiModel.RS256,
		},
		{
			Name:        "demo-signing-ec256",
			Description: "Demo EC P-256 signing key",
			Algorithm:   cryptoutilOpenapiModel.ES256,
		},
		{
			Name:        "demo-wrapping-aes256kw",
			Description: "Demo AES-256 key wrapping key",
			Algorithm:   cryptoutilOpenapiModel.A256GCMA256KW,
		},
	}
}

// SeedDemoData creates demo keys in the KMS database.
// This function is idempotent - it skips keys that already exist by name.
func SeedDemoData(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, businessLogicService *cryptoutilBusinessLogic.BusinessLogicService) error {
	telemetryService.Slogger.Info("Starting demo data seeding")

	keys := DefaultDemoKeys()
	seededCount := 0
	skippedCount := 0

	for _, keyConfig := range keys {
		// Check if key already exists by listing keys and checking names.
		existingKeys, err := businessLogicService.GetElasticKeys(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to check existing keys: %w", err)
		}

		keyExists := false

		for _, existingKey := range existingKeys {
			if existingKey.Name != nil && *existingKey.Name == keyConfig.Name {
				keyExists = true

				telemetryService.Slogger.Debug("Demo key already exists, skipping", "name", keyConfig.Name)

				skippedCount++

				break
			}
		}

		if keyExists {
			continue
		}

		// Create the demo key.
		keyCreate := &cryptoutilOpenapiModel.ElasticKeyCreate{
			Name:        keyConfig.Name,
			Description: keyConfig.Description,
			Algorithm:   &keyConfig.Algorithm,
		}

		_, err = businessLogicService.AddElasticKey(ctx, keyCreate)
		if err != nil {
			telemetryService.Slogger.Error("Failed to create demo key", "name", keyConfig.Name, "error", err)

			return fmt.Errorf("failed to create demo key %s: %w", keyConfig.Name, err)
		}

		telemetryService.Slogger.Info("Created demo key", "name", keyConfig.Name, "algorithm", keyConfig.Algorithm)

		seededCount++
	}

	telemetryService.Slogger.Info("Demo data seeding complete", "seeded", seededCount, "skipped", skippedCount, "total", len(keys))

	return nil
}
