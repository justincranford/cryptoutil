// Copyright (c) 2025 Justin Cranford
//
//

// Package main is the entrypoint for learn-im encrypted instant messaging service.
package main

import (
	"os"

	"cryptoutil/internal/cmd/learn/im"
)

// Version information (injected during build).
// Kept for future use when version flag is implemented.
var (
	_ = "dev"     // version
	_ = "unknown" // buildDate
	_ = "unknown" // gitCommit
)

func main() {
	os.Exit(internalMain(os.Args))
}

// internalMain implements main logic with testable dependencies.
// Delegates to internal/cmd/learn/im.IM() for all functionality.
func internalMain(args []string) int {
	// For Product-Service pattern, args[0] is the executable name
	// Pass remaining args to IM() which will route to subcommands
	// Default behavior: if no args, IM() defaults to "server" subcommand
	return im.IM(args[1:])
}
