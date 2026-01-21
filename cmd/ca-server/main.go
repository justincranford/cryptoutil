// Copyright (c) 2025 Justin Cranford

// Package main provides the CA server entry point.
package main

import (
	"os"

	cryptoutilCACmd "cryptoutil/internal/cmd/cryptoutil/ca"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"start"}
	}

	cryptoutilCACmd.Execute(args)
}
