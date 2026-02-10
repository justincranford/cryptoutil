// Copyright (c) 2025 Justin Cranford
//

// Package main provides the cryptoutil suite entry point.
package main

import (
	"os"

	cryptoutilAppsCryptoutil "cryptoutil/internal/apps/cryptoutil"
)

func main() {
	os.Exit(cryptoutilAppsCryptoutil.Suite(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
