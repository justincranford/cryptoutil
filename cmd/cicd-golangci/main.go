// Copyright (c) 2025-2026 Justin Cranford.
// Package main provides the repository-safe golangci-lint wrapper.
package main

import (
	"os"

	cryptoutilAppsToolsCicdGolangci "cryptoutil/internal/apps-tools/cicd_golangci"
)

func main() {
	os.Exit(cryptoutilAppsToolsCicdGolangci.InternalMain(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
