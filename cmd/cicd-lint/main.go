// Copyright (c) 2025 Justin Cranford

// Package main provides the cicd-lint command entry point.
package main

import (
	"os"

	cryptoutilCmdCicd "cryptoutil/internal/cmd/cicd_lint"
)

func main() {
	os.Exit(cryptoutilCmdCicd.Cicd(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
