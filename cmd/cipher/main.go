// Copyright (c) 2025 Justin Cranford
//
//

// Package main is the entry point for the Cipher application.
package main

import (
	"os"

	cryptoutilCipherApp "cryptoutil/internal/apps/cipher"
)

func main() {
	exitCode := cryptoutilCipherApp.Cipher(os.Args[1:])
	os.Exit(exitCode)
}
