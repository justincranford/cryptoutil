//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.
// Package main provides the __PS_ID__ service entry point.
//
// Structural invariants (enforced by lint-fitness cmd-ps-id-template):
//   - Exactly one Go source file: main.go
//   - Package declaration: package main
//   - Exactly two imports: "os" and "cryptoutil/internal/apps/__PS_ID__"
//   - Import alias follows cryptoutilApps<PascalCasePSID> convention
//   - Single func main() calling os.Exit(<alias>.<PascalCaseService>(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
//
// Example expansion for sm-kms (__PS_ID__=sm-kms, __SERVICE__=kms):
//
//	import cryptoutilAppsSmKms "cryptoutil/internal/apps/sm-kms"
//	func main() { os.Exit(cryptoutilAppsSmKms.Kms(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
package main

import (
	"os"

	cryptoutilApps__PS_ID__ "cryptoutil/internal/apps/__PS_ID__"
)

func main() {
	os.Exit(cryptoutilApps__PS_ID__.__SERVICE__(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
