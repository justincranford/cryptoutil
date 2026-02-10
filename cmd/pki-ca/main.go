// Copyright (c) 2025 Justin Cranford
//

// Package main provides the pki-ca service entry point.
package main

import (
	"os"

	cryptoutilAppsPkiCa "cryptoutil/internal/apps/pki/ca"
)

func main() {
	os.Exit(cryptoutilAppsPkiCa.Ca(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
