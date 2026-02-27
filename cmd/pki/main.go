// Copyright (c) 2025 Justin Cranford
//

// Package main provides the pki product entry point.
package main

import (
	"os"

	cryptoutilAppsPki "cryptoutil/internal/apps/pki"
)

func main() {
	os.Exit(cryptoutilAppsPki.Pki(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
