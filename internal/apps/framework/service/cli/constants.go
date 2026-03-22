// Copyright (c) 2025 Justin Cranford
//

// Package cli provides common CLI utilities for service entrypoints.
// All SERVICE CLI entrypoints use RouteService.
package cli

const (
	helpCommand      = "help"
	helpFlag         = "--help"
	helpShortFlag    = "-h"
	urlFlag          = "--url"
	cacertFlag       = "--cacert"
	versionCommand   = "version"
	versionFlag      = "--version"
	versionShortFlag = "-v"
)

// IsHelpRequest returns true when args begins with a help flag or subcommand.
func IsHelpRequest(args []string) bool {
	return len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag)
}
