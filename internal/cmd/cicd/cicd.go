// Copyright (c) 2025 Justin Cranford

package cicd

import (
	"fmt"
	"io"

	cryptoutilAppsCicd "cryptoutil/internal/apps/cicd"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Cicd is the thin entry point for the cicd command.
// It validates the command and delegates to internal/apps/cicd for all processing.
// Returns exit code: 0 for success, 1 for failure.
func Cicd(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		printUsage(stdout)

		return 1
	}

	command := args[1]

	switch {
	case cryptoutilSharedMagic.ValidCommands[command]:
		// ALL valid commands route to apps/cicd for processing.
		return cryptoutilAppsCicd.Cicd(args, stdin, stdout, stderr)
	case command == "help" || command == "--help" || command == "-h":
		printUsage(stdout)

		return 0
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown command: %s\n\n", command)
		printUsage(stderr)

		return 1
	}
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprint(w, `cicd - Cryptoutil CI/CD linter and formatter tools

Usage:
  cicd <command> [args]

Lint Commands:
  lint-text                 Enforce UTF-8 file encoding (no BOM)
  lint-go                   Go package linters (circular deps, CGO-free SQLite)
  lint-go-test              Go test file linters (test patterns)
  lint-go-mod               Go module linters (dependency updates)
  lint-golangci             golangci-lint config validation (v2 compatibility)
  lint-compose              Docker Compose file linters (admin port exposure)
  lint-ports                Port assignment validation (standardized ports)
  lint-workflow             Workflow file linters (GitHub Actions)

Format Commands:
  format-go                 Go file formatters (any, copyloopvar)
  format-go-test            Go test file formatters (t.Helper)

Deployment Commands:
  lint-deployments [dir]    Validate deployment directory structures
                            Default dir: deployments/
  validate-all [dirs...]    Run all 8 deployment validators sequentially
                            Defaults: deployments/ configs/

Documentation Commands:
  check-chunk-verification  Verify ARCHITECTURE.md chunks propagated to instruction files
  validate-propagation      Validate all ARCHITECTURE.md section references
  validate-chunks           Compare @propagate/@source marker content for staleness

GitHub Cleanup Commands:
  github-cleanup-runs       Delete old workflow runs (default: >7 days)
  github-cleanup-artifacts  Delete old artifacts (default: >7 days)
  github-cleanup-caches     Delete stale caches (default: not accessed in 7 days)
  github-cleanup-all        Run all cleanup operations

  GitHub Cleanup Flags:
    --confirm               Execute deletions (default: dry-run preview only)
    --max-age-days=N        Age threshold in days (default: 7)
    --keep-min-runs=N       Min successful runs to keep per workflow (default: 3)
    --repo=owner/repo       Target repo (default: auto-detect from cwd)

  help, --help, -h          Show this help message

Examples:
  cicd lint-text
  cicd lint-go lint-go-test format-go
  cicd lint-deployments
  cicd validate-all deployments configs
  cicd github-cleanup-all                   # Dry-run preview
  cicd github-cleanup-all --confirm         # Execute deletions
  cicd github-cleanup-runs --max-age-days=7 # Preview runs older than 7 days

See: docs/ARCHITECTURE-TODO.md for architectural documentation (pending).
`)
}
