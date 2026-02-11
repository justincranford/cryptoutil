// Copyright (c) 2025 Justin Cranford
//

// Package rp provides the Relying Party service entry point.
package rp

import (
	"io"

	cryptoutilAppsIdentityRpUnified "cryptoutil/internal/apps/identity/rp/unified"
)

// Rp is the entry point for the identity-rp command.
func Rp(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// Delegate to unified implementation
	// TODO: Update unified to return exit code instead of calling os.Exit
	cryptoutilAppsIdentityRpUnified.Execute(args)

	return 0
}
