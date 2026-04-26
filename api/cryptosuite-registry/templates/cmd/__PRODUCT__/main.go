//go:build ignore

// Copyright (c) 2025 Justin Cranford

// Package main provides the __PRODUCT__ product entry point.
//
// Structural invariants (enforced by lint-fitness cmd-product-template):
//   - Exactly one Go source file: main.go
//   - Package declaration: package main
//   - Exactly two imports: "os" and "cryptoutil/internal/apps/__PRODUCT__"
//   - Import alias follows cryptoutilApps<PascalCaseProduct> convention
//   - Single func main() calling os.Exit(<alias>.<PascalCaseProduct>(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
//
// Example expansion for sm (__PRODUCT__=sm):
//
//	import cryptoutilAppsSm "cryptoutil/internal/apps/sm"
//	func main() { os.Exit(cryptoutilAppsSm.Sm(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
package main

import (
	"os"

	cryptoutilApps__PRODUCT__ "cryptoutil/internal/apps/__PRODUCT__"
)

func main() {
	os.Exit(cryptoutilApps__PRODUCT__.__PRODUCT__(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
