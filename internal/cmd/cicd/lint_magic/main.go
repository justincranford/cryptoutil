// Copyright (c) 2025 Justin Cranford

package lint_magic

import (
"fmt"
"io"
"os"
)

// Main is the CLI entry point for the magic-constants linter.
// Returns exit code: 0 for success, 1 for failure.
func Main(args []string) int {
return mainWithWriters(args, os.Stdout, os.Stderr)
}

// mainWithWriters is the testable entry point that accepts explicit writers.
func mainWithWriters(args []string, stdout, stderr io.Writer) int {
if len(args) == 0 {
printUsage(stdout)

return 1
}

cmd := args[0]

switch cmd {
case "validate-duplicates":
return runValidateDuplicates(args[1:], stdout, stderr)

case "validate-usage":
return runValidateUsage(args[1:], stdout, stderr)

case "validate-all":
return runValidateAll(args[1:], stdout, stderr)

case "help", "--help", "-h":
printUsage(stdout)

return 0

default:
_, _ = fmt.Fprintf(stderr, "Unknown command: %s\n\n", cmd)
printUsage(stderr)

return 1
}
}

func runValidateDuplicates(args []string, stdout, stderr io.Writer) int {
magicDir := defaultMagicDir
if len(args) > 0 {
magicDir = args[0]
}

result, err := ValidateDuplicates(magicDir)
if err != nil {
_, _ = fmt.Fprintf(stderr, "ERROR: %v\n", err)

return 1
}

	_, _ = fmt.Fprint(stdout, FormatDuplicatesResult(result))

if !result.Valid {
return 1
}

return 0
}

func runValidateUsage(args []string, stdout, stderr io.Writer) int {
magicDir := defaultMagicDir
rootDir := defaultRootDir

if len(args) > 0 {
magicDir = args[0]
}

if len(args) > 1 {
rootDir = args[1]
}

result, err := ValidateUsage(magicDir, rootDir)
if err != nil {
_, _ = fmt.Fprintf(stderr, "ERROR: %v\n", err)

return 1
}

	_, _ = fmt.Fprint(stdout, FormatUsageResult(result))

if !result.Valid {
return 1
}

return 0
}

func runValidateAll(args []string, stdout, stderr io.Writer) int {
magicDir := defaultMagicDir
rootDir := defaultRootDir

if len(args) > 0 {
magicDir = args[0]
}

if len(args) > 1 {
rootDir = args[1]
}

result, err := ValidateAll(magicDir, rootDir)
if err != nil {
_, _ = fmt.Fprintf(stderr, "ERROR: %v\n", err)

return 1
}

	_, _ = fmt.Fprint(stdout, FormatAllResult(result))

if !result.Valid {
return 1
}

return 0
}

func printUsage(w io.Writer) {
_, _ = fmt.Fprint(w, `lint-magic - Magic constant validators for internal/shared/magic

Usage:
  lint-magic <command> [args]

Commands:
  validate-duplicates [magic-dir]
        Scan magic-dir for constants that share an identical literal value.
        Two constants with the same value indicate possible duplication.
        Default magic-dir: internal/shared/magic

  validate-usage [magic-dir [root-dir]]
        Build an inventory of all magic constants, then scan root-dir for
        Go files that use those values as bare literals or redefine them as
        local constants outside the magic package.  These violations fall
        through the built-in mnd/goconst linters because mnd only handles
        numbers and goconst requires >=2 occurrences within a single file.
        Default magic-dir: internal/shared/magic
        Default root-dir:  .

  validate-all [magic-dir [root-dir]]
        Run both validators and aggregate the results.  Both validators
        always run to completion so all problems are visible in one pass.

  help, --help, -h
        Show this help message.

Examples:
  lint-magic validate-duplicates
  lint-magic validate-usage
  lint-magic validate-all
  lint-magic validate-all internal/shared/magic .
`)
}
