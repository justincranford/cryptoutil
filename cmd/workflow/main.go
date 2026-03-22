// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides the workflow command-line entry point.
package main

import (
	"os"

	cryptoutilAppsToolsWorkflow "cryptoutil/internal/apps/tools/workflow"
)

func main() {
	os.Exit(cryptoutilAppsToolsWorkflow.Workflow(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
