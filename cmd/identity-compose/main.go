// Copyright (c) 2025 Justin Cranford

// Package main is the entrypoint for identity-compose orchestration service.
package main

import (
	"os"

	cryptoutilAppsIdentityCompose "cryptoutil/internal/apps/identity/compose"
)

func main() {
	os.Exit(cryptoutilAppsIdentityCompose.Compose(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
