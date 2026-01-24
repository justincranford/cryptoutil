// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides the workflow command-line entry point.
package main

import (
	"os"

	cryptoutilWorkflow "cryptoutil/internal/cmd/workflow"
)

func main() {
	os.Exit(cryptoutilWorkflow.Run(os.Args[1:]))
}
