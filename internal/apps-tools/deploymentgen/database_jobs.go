// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"path"
	"slices"
	"strings"

	db "cryptoutil/internal/apps-tools/deploymentrun/database"
)

func validateDatabaseJobCommand(job LifecycleJobInput, sources []LifecycleSecretSource) error {
	args, secret, err := db.JobArguments(job.Name)
	if err != nil || !slices.Equal(args, job.Args) || !slices.Contains(job.SecretNames, secret) {
		return fmt.Errorf("database Job requires its exact supported suite operation and owned manifest")
	}

	for _, source := range sources {
		if source.Namespace == job.Namespace && source.Name == secret && source.Files[db.InputName] != "" {
			return nil
		}
	}

	return fmt.Errorf("database Job manifest file is absent from owned Secret source")
}

// ValidateDatabaseJobMaterial binds the protected decoded manifest and every
// referenced file to this independent graph and Job's restricted mounts.
func ValidateDatabaseJobMaterial(inventory Inventory, input LifecycleInput, secrets []LifecycleSecret) error {
	for _, job := range input.Jobs {
		if job.Name == lifecycleInfraPKIName || strings.HasPrefix(job.Name, lifecyclePKICommand) {
			continue
		}

		_, manifestSecret, err := db.JobArguments(job.Name)
		if err != nil {
			return fmt.Errorf("database job identity: %w", err)
		}

		var raw []byte

		for _, secret := range secrets {
			if secret.Metadata.Namespace == job.Namespace && secret.Metadata.Name == manifestSecret {
				raw = secret.Data[db.InputName]

				break
			}
		}

		manifest, err := db.Decode(raw)
		if err != nil {
			return fmt.Errorf("protected database manifest rejected")
		}

		operation := db.Replicate
		if job.Name == "db-init" {
			operation = db.Initialize
		} else if strings.HasSuffix(job.Name, "-migrate") {
			operation = db.Migrate
		}

		if manifest.Backend != db.Postgres {
			return fmt.Errorf("shared graph Jobs require PostgreSQL, not per-pod SQLite")
		}

		if err := manifest.Validate(operation); err != nil {
			return fmt.Errorf("database operation contract rejected")
		}

		expected := map[string]Database{}

		for _, release := range inventory.Releases {
			if operation == db.Migrate && release.Namespace != job.Namespace {
				continue
			}

			for _, database := range release.Databases {
				if operation != db.Migrate || database.PSID == strings.TrimSuffix(job.Name, "-migrate") {
					expected[database.Name] = database
				}
			}
		}

		if len(manifest.Targets) != len(expected) {
			return fmt.Errorf("database manifest must cover exact selected catalog")
		}

		for _, target := range manifest.Targets {
			database, exists := expected[target.Database]
			if !exists || target.PSID != database.PSID || target.Role != database.Role {
				return fmt.Errorf("database manifest graph identity mismatch")
			}

			if operation != db.Migrate && (target.Publication != database.Publication || target.Subscription != database.Subscription ||
				target.Slot != "slot_"+strings.TrimPrefix(database.Subscription, "sub_")) {
				return fmt.Errorf("replication manifest graph identity mismatch")
			}
		}

		for _, file := range manifest.Files() {
			secretName := path.Base(path.Dir(file))
			key := path.Base(file)

			if !slices.Contains(job.SecretNames, secretName) {
				return fmt.Errorf("database material path is outside restricted Job mounts")
			}

			found := false

			for _, secret := range secrets {
				if secret.Metadata.Namespace == job.Namespace && secret.Metadata.Name == secretName && len(secret.Data[key]) > 0 {
					found = true

					break
				}
			}

			if !found {
				return fmt.Errorf("database protected material key missing")
			}
		}
	}

	return nil
}
