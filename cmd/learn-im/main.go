// Copyright (c) 2025 Justin Cranford
//
//

// Package main is the entrypoint for learn-im encrypted instant messaging service.
package main

import (
	"os"

	cryptoutilLearnCmd "cryptoutil/internal/cmd/learn"
)

// Version information (injected during build).
var (
	version   = "dev"
	buildDate = "unknown"
	gitCommit = "unknown"
)

func main() {
	os.Exit(internalMain(os.Args))
}

// internalMain implements main logic with testable dependencies.
// Delegates to internal/cmd/learn.IM() for all functionality.
func internalMain(args []string) int {
	// For Product-Service pattern, args[0] is the executable name
	// Pass remaining args to IM() which will route to subcommands
	// Default behavior: if no args, IM() defaults to "server" subcommand
	return cryptoutilLearnCmd.IM(args[1:])
}
