// Copyright (c) 2025 Justin Cranford
//
//

package magic

// ANSI color codes for console output.
const (
	// ColorReset - ANSI reset color code.
	ColorReset = "\033[0m"
	// ColorRed - ANSI red color code.
	ColorRed = "\033[31m"
	// ColorGreen - ANSI green color code.
	ColorGreen = "\033[32m"
	// ColorYellow - ANSI yellow color code.
	ColorYellow = "\033[33m"
	// ColorCyan - ANSI cyan color code.
	ColorCyan = "\033[36m"
)

// Layout constants for console output formatting.
const (
	// LineWidth - Standard line width for console reports.
	LineWidth = 80
	// MaxErrorDisplay - Maximum number of errors to display in reports.
	MaxErrorDisplay = 20
	// MaxWarningDisplay - Maximum number of warnings to display in reports.
	MaxWarningDisplay = 10
)
