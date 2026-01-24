// Copyright (c) 2025 Justin Cranford
//
//

// Package main is the entry point for the Cipher application.
package main

import (
	"os"

	cryptoutilAppsCipher "cryptoutil/internal/apps/cipher"
)

func main() {
	exitCode := cryptoutilAppsCipher.Cipher(os.Args[1:])
	os.Exit(exitCode)
}
