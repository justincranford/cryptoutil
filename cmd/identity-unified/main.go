// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides the unified identity command-line entry point.
package main

import (
	"os"

	cryptoutilAppsIdentityUnified "cryptoutil/internal/apps/identity/unified"
)

func main() {
	os.Exit(cryptoutilAppsIdentityUnified.Unified(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
