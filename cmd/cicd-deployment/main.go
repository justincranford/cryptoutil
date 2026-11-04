// Copyright (c) 2025-2026 Justin Cranford.
// Package main provides the cicd-deployment command entry point.
package main

import (
	"os"

	cryptoutilCICDDeployment "cryptoutil/internal/apps-tools/cicd_deployment"
)

func main() {
	os.Exit(cryptoutilCICDDeployment.Main(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
