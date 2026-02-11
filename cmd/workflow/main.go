// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides the workflow command-line entry point.
package main

import (
	"os"

	cryptoutilAppsWorkflow "cryptoutil/internal/apps/workflow"
)

func main() {
	os.Exit(cryptoutilAppsWorkflow.Workflow(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
