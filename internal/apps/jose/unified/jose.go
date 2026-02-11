// Copyright (c) 2025 Justin Cranford
//
//

// Package jose provides the unified command interface for JOSE Authority service.
package jose

import (
	"os"

	cryptoutilAppsJoseJa "cryptoutil/internal/apps/jose/ja"
)

// Execute handles JOSE service commands by delegating to ja.Ja.
// This provides a unified command interface for the cryptoutil tool.
func Execute(parameters []string) {
	// ja.Ja expects args like: ["jose-ja", "start", ...]
	// We receive: ["start", ...] from parameters
	args := append([]string{"jose-ja"}, parameters...)
	exitCode := cryptoutilAppsJoseJa.Ja(args, os.Stdin, os.Stdout, os.Stderr)
	os.Exit(exitCode)
}
