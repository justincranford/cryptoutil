// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides the CICD utilities entry point.
package main

import (
	"os"

	cryptoutilAppsCicd "cryptoutil/internal/apps/cicd"
)

func main() {
	os.Exit(cryptoutilAppsCicd.Cicd(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
