// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides the unified identity command-line entry point.
package main

import (
	"fmt"
	"os"

	cryptoutilIdentityCmd "cryptoutil/internal/cmd/cryptoutil/identity"
)

func main() {
	args := os.Args[1:] // Skip program name
	if len(args) == 0 {
		fmt.Println("Usage: identity <subcommand> [flags]")
		fmt.Println("Subcommands: start, stop, status, health")
		os.Exit(1)
	}

	cryptoutilIdentityCmd.Execute(args)
}
