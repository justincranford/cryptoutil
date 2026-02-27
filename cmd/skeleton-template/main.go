// Copyright (c) 2025 Justin Cranford
//

// Package main provides the skeleton-template service entry point.
package main

import (
	"os"

	cryptoutilAppsSkeletonTemplate "cryptoutil/internal/apps/skeleton/template"
)

func main() {
	os.Exit(cryptoutilAppsSkeletonTemplate.Template(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
