// Copyright (c) 2025 Justin Cranford

package cicd

import (
	"fmt"
	"io"

	cryptoutilAppsCicd "cryptoutil/internal/apps/cicd"
	cryptoutilCleanupGitHub "cryptoutil/internal/apps/cicd/cleanup_github"
	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
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

	switch command {
	case "lint-deployments":
		return cryptoutilLintDeployments.Main(args[2:])
	case "generate-listings":
		return cryptoutilLintDeployments.Main([]string{"generate-listings"})
	case "validate-mirror":
		return cryptoutilLintDeployments.Main([]string{"validate-mirror"})
	case "validate-compose":
		return cryptoutilLintDeployments.Main(append([]string{"validate-compose"}, args[2:]...))
	case "validate-config":
		return cryptoutilLintDeployments.Main(append([]string{"validate-config"}, args[2:]...))
	case "validate-all":
		return cryptoutilLintDeployments.Main(append([]string{"validate-all"}, args[2:]...))
	case "check-chunk-verification":
		return CheckChunkVerification(stdout, stderr)
	case "validate-propagation":
		return ValidatePropagationCommand(stdout, stderr)
	case "validate-chunks":
		return ValidateChunksCommand(stdout, stderr)
	case "cleanup-runs":
		return runCleanup(args[2:], stderr, cryptoutilCleanupGitHub.CleanupRuns)
	case "cleanup-artifacts":
		return runCleanup(args[2:], stderr, cryptoutilCleanupGitHub.CleanupArtifacts)
	case "cleanup-caches":
		return runCleanup(args[2:], stderr, cryptoutilCleanupGitHub.CleanupCaches)
	case "cleanup-all":
		return runCleanup(args[2:], stderr, cryptoutilCleanupGitHub.CleanupAll)
	case "lint-text", "lint-go", "lint-go-test", "lint-compose", "lint-ports",
		"lint-workflow", "lint-go-mod", "lint-golangci", "format-go", "format-go-test":
		return cryptoutilAppsCicd.Cicd(args, stdin, stdout, stderr)
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

Cleanup Commands (GitHub Actions storage):
  cleanup-runs              Delete old workflow runs (default: >7 days)
  cleanup-artifacts         Delete old artifacts (default: >7 days)
  cleanup-caches            Delete stale caches (default: not accessed in 7 days)
  cleanup-all               Run all cleanup operations

  Cleanup flags:
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
  cicd cleanup-all                          # Dry-run preview
  cicd cleanup-all --confirm                # Execute deletions
  cicd cleanup-runs --max-age-days=7        # Preview runs older than 7 days

See: docs/ARCHITECTURE-TODO.md for architectural documentation (pending).
`)
}

// runCleanup initializes config from args and runs the specified cleanup function.
func runCleanup(args []string, stderr io.Writer, cleanupFn func(*cryptoutilCleanupGitHub.CleanupConfig) error) int {
	logger := cryptoutilCmdCicdCommon.NewLogger("cleanup")
	cfg := cryptoutilCleanupGitHub.NewDefaultConfig(logger)

	if err := cryptoutilCleanupGitHub.ParseArgs(args, cfg); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)

		return 1
	}

	if err := cleanupFn(cfg); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)

		return 1
	}

	return 0
}
