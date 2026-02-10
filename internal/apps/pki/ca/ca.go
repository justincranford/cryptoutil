// Copyright (c) 2025 Justin Cranford
//

// Package ca provides the CA service entry point.
package ca

import (
	"io"

	cryptoutilAppsPkiCaUnified "cryptoutil/internal/apps/pki/ca/unified"
)

// Ca is the entry point for the pki-ca command.
func Ca(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// Delegate to unified implementation
	// TODO: Update unified to return exit code instead of calling os.Exit
	cryptoutilAppsPkiCaUnified.Execute(args)

	return 0
}
