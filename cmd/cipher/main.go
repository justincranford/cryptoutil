// Copyright (c) 2025 Justin Cranford
//

// Package main is the entry point for the Cipher application.
package main

import (
	"os"

	cryptoutilAppsCipher "cryptoutil/internal/apps/cipher"
)

func main() {
	os.Exit(cryptoutilAppsCipher.Cipher(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
