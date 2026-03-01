// Copyright (c) 2025 Justin Cranford

package cicd

import (
	"fmt"
	"io"

	cryptoutilAppsCicd "cryptoutil/internal/apps/cicd"
	cryptoutilGitHubCleanup "cryptoutil/internal/apps/cicd/github_cleanup"
	cryptoutilLintDeployments "cryptoutil/internal/cmd/cicd/lint_deployments"
)

// Cicd is the main entry point for the cicd command.
// It parses the command from args and delegates to the appropriate subcommand.
// Returns exit code: 0 for success, 1 for failure.
func Cicd(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		printUsage(stdout)

		return 1
	}

	command := args[1]
	extraArgs := getExtraArgs(args)

	switch command {
	// Lint/format commands routed to internal/apps/cicd batch processor.
	case "lint-text", "lint-go", "lint-go-test", "lint-compose", "lint-ports",
		"lint-workflow", "lint-go-mod", "lint-golangci", "format-go", "format-go-test":
		return cryptoutilAppsCicd.Cicd(args, stdin, stdout, stderr)

	// GitHub cleanup commands routed to internal/apps/cicd/github_cleanup.
	case "github-cleanup-runs", "github-cleanup-artifacts", "github-cleanup-caches", "github-cleanup-all":
		return cryptoutilGitHubCleanup.Main(command, extraArgs, stderr)

	// Deployment validators (TODO: migrate to internal/apps/cicd/lint_deployments).
	case "lint-deployments":
		return cryptoutilLintDeployments.Main(extraArgs)
	case "generate-listings":
		return cryptoutilLintDeployments.Main([]string{"generate-listings"})
	case "validate-mirror":
		return cryptoutilLintDeployments.Main([]string{"validate-mirror"})
	case "validate-compose":
		return cryptoutilLintDeployments.Main(append([]string{"validate-compose"}, extraArgs...))
	case "validate-config":
		return cryptoutilLintDeployments.Main(append([]string{"validate-config"}, extraArgs...))
	case "validate-all":
		return cryptoutilLintDeployments.Main(append([]string{"validate-all"}, extraArgs...))

	// Documentation validators (TODO: migrate to internal/apps/cicd/).
	case "check-chunk-verification":
		return CheckChunkVerification(stdout, stderr)
	case "validate-propagation":
		return ValidatePropagationCommand(stdout, stderr)
	case "validate-chunks":
		return ValidateChunksCommand(stdout, stderr)

	case "help", "--help", "-h":
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

// getExtraArgs safely extracts arguments after the command name.
// Returns empty slice if args has fewer than 3 elements (binary + command + extra).
func getExtraArgs(args []string) []string {
	const minArgsWithExtra = 3 // binary name + command + at least one extra arg

	if len(args) < minArgsWithExtra {
		return []string{}
	}

	return args[2:]
}
