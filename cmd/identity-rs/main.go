// Copyright (c) 2025 Justin Cranford
//

// Package main provides the identity-rs service entry point.
package main

import (
"os"

cryptoutilAppsIdentityRs "cryptoutil/internal/apps/identity/rs"
)

func main() {
os.Exit(cryptoutilAppsIdentityRs.Rs(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
