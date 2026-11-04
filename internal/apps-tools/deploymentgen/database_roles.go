// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"errors"
	"fmt"
	"strings"

	db "cryptoutil/internal/apps-tools/deploymentrun/database"
)

// BuildRoleProvisionInput binds explicit protected transport references to the
// independent canonical graph. It generates no database credential or authority.
func BuildRoleProvisionInput(model Model, releaseKey string, connections db.Input) (db.RoleProvisionInput, error) {
	inventory, err := Build(model, releaseKey)
	if err != nil || len(inventory.Releases) != 1 || len(inventory.Releases[0].Databases) == 0 {
		return db.RoleProvisionInput{}, errors.New("one canonical PostgreSQL release required for ordinary role provisioning")
	}
	// Build accepted exactly one selected canonical key; its validated tier,
	// scope and mode identifiers cannot contain the ':' delimiter.
	parts := strings.Split(inventory.Releases[0].Key, ":")

	input := db.RoleProvisionInput{Run: model.Run, Tier: parts[0], Scope: parts[1], Variant: parts[2], Database: connections}
	if err := input.Validate(); err != nil {
		return db.RoleProvisionInput{}, fmt.Errorf("ordinary role protected graph contract: %w", err)
	}

	release := inventory.Releases[0]
	if len(input.Database.Targets) != len(release.Databases) {
		return db.RoleProvisionInput{}, errors.New("ordinary role graph catalog incomplete")
	}

	expected := map[string]bool{}
	for _, database := range release.Databases {
		expected[database.PSID] = true
	}

	for _, target := range input.Database.Targets {
		if !expected[target.PSID] {
			return db.RoleProvisionInput{}, errors.New("ordinary role service outside owning graph")
		}
	}

	return input, nil
}
