// Copyright (c) 2025 Justin Cranford
//

// Package main is the entrypoint for sm-im encrypted instant messaging service.
package main

import (
	"os"

	cryptoutilAppsSmIm "cryptoutil/internal/apps/sm/im"
)

func main() {
	os.Exit(cryptoutilAppsSmIm.Im(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
