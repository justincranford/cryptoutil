// Copyright (c) 2025 Justin Cranford

package cmd

import (
	"fmt"
	"io"

	cryptoutilAppsCicd "cryptoutil/internal/apps-tools/cicd_lint"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Cicd is the thin entry point for the cicd-lint command.
// It validates the command and delegates to internal/apps/cicd for all processing.
// Flags (e.g. -q, --summary) are allowed before or after command names.
// Returns exit code: 0 for success, 1 for failure.
func Cicd(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		printUsage(stdout)

		return 1
	}

	// Check for help flags anywhere in the argument list.
	if hasHelpFlag(args[1:]) {
		printUsage(stdout)

		return 0
	}

	// Find the first non-flag argument to determine the command for routing.
	firstCommand := firstNonFlag(args[1:])

	switch {
	case firstCommand == "":
		// Only flags provided (non-help) — delegate to apps/cicd which will return usage error.
		return cryptoutilAppsCicd.Cicd(args, stdin, stdout, stderr)
	case cryptoutilSharedMagic.ValidCommands[firstCommand]:
		// ALL valid commands route to apps/cicd for processing.
		return cryptoutilAppsCicd.Cicd(args, stdin, stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown command: %s\n\n", firstCommand)
		printUsage(stderr)

		return 1
	}
}

// hasHelpFlag returns true if any help flag (-h, --help, help) is present in args.
func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == cryptoutilSharedMagic.CLIHelpCommand || arg == cryptoutilSharedMagic.CLIHelpFlag || arg == "-h" {
			return true
		}
	}

	return false
}

// firstNonFlag returns the first argument that does not begin with '-'.
// Returns empty string if no non-flag argument is found.
func firstNonFlag(args []string) string {
	for _, arg := range args {
		if arg != "" && arg[0] != '-' {
			return arg
		}
	}

	return ""
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprint(w, `cicd-lint - Cryptoutil CI/CD linter and formatter tools

Usage:
	cicd-lint <command> [args]

Lint Commands:
	lint-fitness              Architecture fitness functions (cross-service isolation, file limits, etc.)
	lint-text                 Enforce UTF-8 file encoding (no BOM)
	lint-go                   Go package linters (circular deps, CGO-free SQLite)
	lint-go-test              Go test file linters (test patterns)
	lint-go-mod               Go module linters (dependency updates)
	lint-golangci             golangci-lint config validation (v2 compatibility)
	lint-compose              Docker Compose file linters (admin port exposure)
	lint-ports                Port assignment validation (standardized ports)
	lint-workflow             Workflow file linters (GitHub Actions)
	lint-deployments          Deployment structure and config file validation
	lint-docs                 Documentation chunk verification and propagation validation

Format Commands:
	format-go                 Go file formatters (any, copyloopvar)
	format-go-test            Go test file formatters (t.Helper)

Script Commands:
	github-cleanup            GitHub Actions storage cleanup (runs, artifacts, caches)

	GitHub Cleanup Flags:
	--confirm               Execute deletions (default: dry-run preview only)
	--max-age-days=N        Age threshold in days (default: 7)
	--keep-min-runs=N       Min successful runs to keep per workflow (default: 3)
	--repo=owner/repo       Target repo (default: auto-detect from cwd)

	help, --help, -h          Show this help message

Examples:
	cicd-lint lint-text
	cicd-lint lint-go lint-go-test format-go
	cicd-lint lint-deployments
	cicd-lint lint-docs
	cicd-lint github-cleanup                   # Dry-run preview
	cicd-lint github-cleanup --confirm         # Execute deletions
	cicd-lint github-cleanup --max-age-days=7  # Preview runs older than 7 days
`)
}
