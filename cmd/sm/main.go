// Copyright (c) 2025 Justin Cranford
//

// Package main provides the sm product entry point.
package main

import (
"os"

cryptoutilAppsSm "cryptoutil/internal/apps/sm"
)

func main() {
os.Exit(cryptoutilAppsSm.Sm(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
