// Copyright (c) 2025 Justin Cranford

// Package main is the entrypoint for identity-demo demonstration service.
package main

import (
	"os"

	cryptoutilAppsIdentityDemo "cryptoutil/internal/apps/identity/demo"
)

func main() {
	os.Exit(cryptoutilAppsIdentityDemo.Demo(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
