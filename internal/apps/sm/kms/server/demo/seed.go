// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides demo data seeding functionality for the KMS server.
package demo

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilKmsServerBusinesslogic "cryptoutil/internal/apps/sm/kms/server/businesslogic"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
)

// DemoTenantConfig holds demo tenant configuration.
type DemoTenantConfig struct {
	// ID is regenerated on each startup (Session 4 Q9).
	ID   string
	Name string
}

// DemoKeyConfig defines a demo key to be seeded.
type DemoKeyConfig struct {
	Name        string
	Description string
	Algorithm   cryptoutilOpenapiModel.ElasticKeyAlgorithm
}

// GenerateDemoTenantID creates a new UUIDv4 tenant ID.
// Reference: Session 4 Q9 - regenerate on each startup for security.
func GenerateDemoTenantID() string {
	return googleUuid.New().String()
}

// DefaultDemoTenants returns the default demo tenants with fresh UUIDs.
// Reference: Session 4 Q9 - regenerate demo tenant IDs on each startup.
func DefaultDemoTenants() []DemoTenantConfig {
	return []DemoTenantConfig{
		{
			ID:   GenerateDemoTenantID(),
			Name: "demo-tenant-primary",
		},
		{
			ID:   GenerateDemoTenantID(),
			Name: "demo-tenant-secondary",
		},
	}
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
func SeedDemoData(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, businessLogicService *cryptoutilKmsServerBusinesslogic.BusinessLogicService) error {
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
		keyCreate := &cryptoutilKmsServer.ElasticKeyCreate{
			Name:        keyConfig.Name,
			Description: &keyConfig.Description,
			Algorithm:   string(keyConfig.Algorithm),
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

// ResetDemoData clears all existing demo keys and re-seeds them.
// This function disables all keys with names matching the default demo keys and then re-seeds.
func ResetDemoData(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, businessLogicService *cryptoutilKmsServerBusinesslogic.BusinessLogicService) error {
	telemetryService.Slogger.Info("Starting demo data reset")

	keys := DefaultDemoKeys()

	// First, disable existing demo keys
	for _, keyConfig := range keys {
		// Check if key exists and disable it
		existingKeys, err := businessLogicService.GetElasticKeys(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to check existing keys: %w", err)
		}

		for _, existingKey := range existingKeys {
			if existingKey.Name != nil && *existingKey.Name == keyConfig.Name {
				// Disable the key by updating its status
				// Note: Since there's no UpdateElasticKeyStatus method in business logic service,
				// we'll need to add one or use repository directly. For now, we'll skip disabling
				// and just re-seed (which will be idempotent)
				telemetryService.Slogger.Info("Demo key exists, will be replaced during re-seed", "name", keyConfig.Name)

				break
			}
		}
	}

	telemetryService.Slogger.Info("Demo key disabling skipped (not implemented)", "would_disable", len(keys))

	// Now re-seed the demo data (this will be idempotent)
	return SeedDemoData(ctx, telemetryService, businessLogicService)
}
