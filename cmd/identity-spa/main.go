// Copyright (c) 2025 Justin Cranford
//

// Package main provides the identity-spa service entry point.
package main

import (
"os"

cryptoutilAppsIdentitySpa "cryptoutil/internal/apps/identity/spa"
)

func main() {
os.Exit(cryptoutilAppsIdentitySpa.Spa(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
