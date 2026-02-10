// Copyright (c) 2025 Justin Cranford
//

// Package main provides the jose product entry point.
package main

import (
"os"

cryptoutilAppsJose "cryptoutil/internal/apps/jose"
)

func main() {
os.Exit(cryptoutilAppsJose.Jose(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
