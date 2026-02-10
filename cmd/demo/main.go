// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides a unified demo CLI for cryptoutil products.
// This single binary supports subcommands for KMS, Identity, and full integration demos.
package main

import (
	"os"

	cryptoutilAppsDemo "cryptoutil/internal/apps/demo"
)

func main() {
	os.Exit(cryptoutilAppsDemo.Demo(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
