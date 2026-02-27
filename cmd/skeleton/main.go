// Copyright (c) 2025 Justin Cranford
//

// Package main provides the skeleton product entry point.
package main

import (
	"os"

	cryptoutilAppsSkeleton "cryptoutil/internal/apps/skeleton"
)

func main() {
	os.Exit(cryptoutilAppsSkeleton.Skeleton(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
