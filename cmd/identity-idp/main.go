// Copyright (c) 2025 Justin Cranford
//

// Package main provides the identity-idp service entry point.
package main

import (
	"os"

	cryptoutilAppsIdentityIdp "cryptoutil/internal/apps/identity/idp"
)

func main() {
	os.Exit(cryptoutilAppsIdentityIdp.Idp(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
