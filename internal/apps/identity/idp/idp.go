// Copyright (c) 2025 Justin Cranford
//

// Package idp provides the Identity Provider service entry point.
package idp

import (
	"io"

	cryptoutilAppsIdentityIdpUnified "cryptoutil/internal/apps/identity/idp/unified"
)

// Idp is the entry point for the identity-idp command.
func Idp(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// Delegate to unified implementation
	// TODO: Update unified to return exit code instead of calling os.Exit
	cryptoutilAppsIdentityIdpUnified.Execute(args)

	return 0
}
