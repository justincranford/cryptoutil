// Copyright (c) 2025 Justin Cranford
//

// Package main provides the sm-kms service entry point.
package main

import (
	"os"

	cryptoutilAppsSmKms "cryptoutil/internal/apps/sm/kms"
)

func main() {
	os.Exit(cryptoutilAppsSmKms.Kms(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
