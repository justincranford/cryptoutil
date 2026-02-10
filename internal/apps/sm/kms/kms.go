// Package kms provides the KMS service entry point.
package kms

import (
	"io"

	cryptoutilKmsCmd "cryptoutil/internal/apps/sm/kms/cmd"
)

// Kms is the entry point for the sm-kms command.
func Kms(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// TODO: Add proper command handling
	cryptoutilKmsCmd.Server(args)
	return 0
}
