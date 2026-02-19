// Copyright (c) 2025 Justin Cranford
//

// Package cli provides common CLI utilities for product and service entrypoints.
// All PRODUCT CLI entrypoints use RouteProduct; all SERVICE CLI entrypoints use RouteService.
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
