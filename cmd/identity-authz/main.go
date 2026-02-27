// Copyright (c) 2025 Justin Cranford
//

// Package main provides the identity-authz service entry point.
package main

import (
	"os"

	cryptoutilAppsIdentityAuthz "cryptoutil/internal/apps/identity/authz"
)

func main() {
	os.Exit(cryptoutilAppsIdentityAuthz.Authz(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
