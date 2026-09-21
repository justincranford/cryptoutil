// Copyright (c) 2025-2026 Justin Cranford.
// Command cicd-dev-setup verifies and idempotently installs local developer tooling
// (Go, Python, golangci-lint, pre-commit, etc.) per configs/dev-setup.yml.
package main

import (
	"os"

	cryptoutilAppsToolsDevSetup "cryptoutil/internal/apps-tools/cicd_dev_setup"
)

func main() {
	os.Exit(cryptoutilAppsToolsDevSetup.Main(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
