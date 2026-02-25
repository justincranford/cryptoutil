// Copyright (c) 2025 Justin Cranford
//

// Package main provides the identity-rp service entry point.
package main

import (
"os"

cryptoutilAppsIdentityRp "cryptoutil/internal/apps/identity/rp"
)

func main() {
os.Exit(cryptoutilAppsIdentityRp.Rp(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
