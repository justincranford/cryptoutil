// Copyright (c) 2025 Justin Cranford

package cicd

import (
	"fmt"
	"io"

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

Commands:
  lint-deployments [dir]    Validate deployment directory structures
                            Default dir: deployments/

  validate-all [deployments-dir configs-dir]
                            Run all 8 deployment validators sequentially with aggregated reporting
                            Defaults: deployments/ configs/

  check-chunk-verification  Verify ARCHITECTURE.md chunks propagated to instruction files

  help, --help, -h          Show this help message

Examples:
  cicd lint-deployments
  cicd lint-deployments /tmp/test-deployments
  cicd validate-all
  cicd validate-all deployments configs

See: docs/ARCHITECTURE-TODO.md for architectural documentation (pending).
`)
}
