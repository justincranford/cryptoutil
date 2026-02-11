// Copyright (c) 2025 Justin Cranford
//

// Package spa provides the Single Page Application service entry point.
package spa

import (
	"io"

	cryptoutilAppsIdentitySpaUnified "cryptoutil/internal/apps/identity/spa/unified"
)

// Spa is the entry point for the identity-spa command.
func Spa(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// Delegate to unified implementation
	// TODO: Update unified to return exit code instead of calling os.Exit
	cryptoutilAppsIdentitySpaUnified.Execute(args)

	return 0
}
