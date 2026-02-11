// Copyright (c) 2025 Justin Cranford
//

// Package rs provides the Resource Server service entry point.
package rs

import (
	"io"

	cryptoutilAppsIdentityRsUnified "cryptoutil/internal/apps/identity/rs/unified"
)

// Rs is the entry point for the identity-rs command.
func Rs(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// Delegate to unified implementation
	// TODO: Update unified to return exit code instead of calling os.Exit
	cryptoutilAppsIdentityRsUnified.Execute(args)

	return 0
}
