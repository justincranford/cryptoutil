// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"errors"
	"strings"
)

// JobArguments binds the existing graph Job identity to the supported suite
// operation and a manifest inside its already owned external Secret.
func JobArguments(name string) ([]string, string, error) {
	operation, secret := "", ""

	switch {
	case name == "db-init":
		operation = Initialize
		secret = "postgres-bootstrap"
	case name == "replication":
		operation = Replicate
		secret = "postgres-follower-bootstrap"
	case strings.HasSuffix(name, "-migrate"):
		psid := strings.TrimSuffix(name, "-migrate")
		if _, err := Migrations(psid); err != nil {
			return nil, "", err
		}

		operation = Migrate
		secret = psid + "-migration-db"
	default:
		return nil, "", errors.New("unsupported database Job")
	}

	return []string{Command, "--operation=" + operation, "--input=" + SecretDirectory + secret + "/" + InputName}, secret, nil
}
