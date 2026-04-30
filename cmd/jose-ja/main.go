// Copyright (c) 2025-2026 Justin Cranford.
//

// Package main provides the jose-ja service entry point.
package main

import (
	"os"

	cryptoutilAppsJoseJa "cryptoutil/internal/apps/jose-ja"
)

func main() {
	os.Exit(cryptoutilAppsJoseJa.Ja(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
