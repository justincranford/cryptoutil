package main

import (
"fmt"
"os"

cryptoutilLintDeployments "cryptoutil/internal/cmd/cicd/lint_deployments"
)

func main() {
if len(os.Args) < 2 {
printUsage()
os.Exit(1)
}

command := os.Args[1]

switch command {
case "lint-deployments":
os.Exit(cryptoutilLintDeployments.Main(os.Args[2:]))
case "help", "--help", "-h":
printUsage()
os.Exit(0)
default:
fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
printUsage()
os.Exit(1)
}
}

func printUsage() {
fmt.Println(`cicd - Cryptoutil CI/CD linter and formatter tools

Usage:
  cicd <command> [args]

Commands:
  lint-deployments [dir]    Validate deployment directory structures
                            Default dir: deployments/

  help, --help, -h          Show this help message

Examples:
  cicd lint-deployments
  cicd lint-deployments /tmp/test-deployments

See: docs/ARCHITECTURE-TODO.md for architectural documentation (pending).
`)
}
