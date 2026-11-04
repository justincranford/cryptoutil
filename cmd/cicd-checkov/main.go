// Copyright (c) 2025-2026 Justin Cranford.
// Package main provides the cicd-checkov command entry point.
package main

import (
	"os"

	cryptoutilCICDCheckov "cryptoutil/internal/apps-tools/cicd_checkov"
)

func main() {
	os.Exit(cryptoutilCICDCheckov.InternalMain(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
