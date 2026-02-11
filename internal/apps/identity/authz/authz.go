// Copyright (c) 2025 Justin Cranford
//

// Package authz provides the Authorization Server service entry point.
package authz

import (
	"io"

	cryptoutilAppsIdentityAuthzUnified "cryptoutil/internal/apps/identity/authz/unified"
)

// Authz is the entry point for the identity-authz command.
func Authz(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// Delegate to unified implementation
	// TODO: Update unified to return exit code instead of calling os.Exit
	cryptoutilAppsIdentityAuthzUnified.Execute(args)

	return 0
}
