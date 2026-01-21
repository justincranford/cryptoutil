// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides the entry point for the jose-server command.
package main

import (
	"fmt"
	"os"

	cryptoutilJoseCmd "cryptoutil/internal/cmd/cryptoutil/jose"
)

func main() {
	args := os.Args[1:] // Skip program name
	if len(args) == 0 {
		fmt.Println("Usage: jose-server <subcommand> [flags]")
		fmt.Println("Subcommands: start, stop, status, health")
		os.Exit(1)
	}

	cryptoutilJoseCmd.Execute(args)
}
