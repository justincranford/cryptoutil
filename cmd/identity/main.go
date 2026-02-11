// Copyright (c) 2025 Justin Cranford
//

// Package main provides the identity product entry point.
package main

import (
"os"

cryptoutilAppsIdentity "cryptoutil/internal/apps/identity"
)

func main() {
os.Exit(cryptoutilAppsIdentity.Identity(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
