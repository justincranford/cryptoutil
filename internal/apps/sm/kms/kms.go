// Copyright (c) 2025 Justin Cranford
//

// Package kms provides the KMS service entry point.
package kms

import (
	"io"

	cryptoutilKmsCmd "cryptoutil/internal/apps/sm/kms/cmd"
)

// Main is the entry point for the sm-kms command.
func Main(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// TODO: Add proper command handling
	cryptoutilKmsCmd.Server(args)

	return 0
}
