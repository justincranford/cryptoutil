// Copyright (c) 2025 Justin Cranford
//
// TEMPLATE: Copy and rename 'skeleton' → your-service-name before use.

// Package main provides the skeleton-template service entry point.
package main

import (
	"os"

	cryptoutilAppsSkeletonTemplate "cryptoutil/internal/apps/skeleton/template"
)

func main() {
	os.Exit(cryptoutilAppsSkeletonTemplate.Template(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
