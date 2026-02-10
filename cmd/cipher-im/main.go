// Copyright (c) 2025 Justin Cranford
//

// Package main is the entrypoint for cipher-im encrypted instant messaging service.
package main

import (
	"os"

	cryptoutilAppsCipherIm "cryptoutil/internal/apps/cipher/im"
)

func main() {
	os.Exit(cryptoutilAppsCipherIm.Im(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
