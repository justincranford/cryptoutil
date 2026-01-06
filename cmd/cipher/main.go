// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"os"

	cryptoutilCipherApp "cryptoutil/internal/apps/cipher"
)

func main() {
	exitCode := cryptoutilCipherApp.Cipher(os.Args[1:])
	os.Exit(exitCode)
}
