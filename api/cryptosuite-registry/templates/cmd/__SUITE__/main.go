//go:build ignore

// Copyright (c) 2025 Justin Cranford

// Package main provides the __SUITE__ suite entry point.
//
// Structural invariants (enforced by lint-fitness cmd-suite-template):
//   - Exactly one Go source file: main.go
//   - Package declaration: package main
//   - Exactly two imports: "os" and "cryptoutil/internal/apps/__SUITE__"
//   - Import alias follows cryptoutilApps<PascalCaseSuite> convention
//   - Single func main() calling os.Exit(<alias>.Suite(os.Args, os.Stdin, os.Stdout, os.Stderr))
//   - CRITICAL: Suite uses os.Args (full args including program name), NOT os.Args[1:]
//     Suite router needs argv[0] to display meaningful usage text for nested product routing.
//
// Example expansion for cryptoutil (__SUITE__=cryptoutil):
//
//	import cryptoutilAppsCryptoutil "cryptoutil/internal/apps/cryptoutil"
//	func main() { os.Exit(cryptoutilAppsCryptoutil.Suite(os.Args, os.Stdin, os.Stdout, os.Stderr)) }
package main

import (
	"os"

	cryptoutilApps__SUITE__ "cryptoutil/internal/apps/__SUITE__"
)

func main() {
	os.Exit(cryptoutilApps__SUITE__.Suite(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
