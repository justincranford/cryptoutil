// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides the cicd-workflow command-line entry point.
package main

import (
	"os"

	cryptoutilAppsToolsCicdWorkflow "cryptoutil/internal/apps/tools/cicd_workflow"
)

func main() {
	os.Exit(cryptoutilAppsToolsCicdWorkflow.Workflow(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
