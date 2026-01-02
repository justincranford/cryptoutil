// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"os"

	cryptoutilCipherCmd "cryptoutil/internal/cmd/cipher"
)

func main() {
	exitCode := cryptoutilCipherCmd.Cipher(os.Args[1:])
	os.Exit(exitCode)
}
