// Copyright (c) 2025 Justin Cranford

// Package cicd_coverage validates that all cicd linter and formatter commands
// are covered in both the custom-cicd-lint composite action and the pre-commit
// configuration. This ensures no command is accidentally omitted from either
// local or CI enforcement.
//
// Checked files:
//   - .github/actions/custom-cicd-lint/action.yml
//   - .pre-commit-config.yaml
//   - .github/workflows/ci-cicd-lint.yml
package cicd_coverage

import (
"fmt"
"os"
"path/filepath"
"sort"
"strings"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// linterAndFormatterCmds returns sorted slices of all lint-* and format-* commands
// from the canonical ValidCommands registry.
func linterAndFormatterCmds() (linters, formatters []string) {
for cmd := range cryptoutilSharedMagic.ValidCommands {
switch {
case strings.HasPrefix(cmd, "lint-"):
linters = append(linters, cmd)
case strings.HasPrefix(cmd, "format-"):
formatters = append(formatters, cmd)
}
}

sort.Strings(linters)
sort.Strings(formatters)

return linters, formatters
}

// Check validates cicd command coverage using the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
return CheckInDir(logger, ".")
}

// CheckInDir validates cicd command coverage relative to rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
logger.Log("Checking cicd command coverage in action, pre-commit config, and CI workflow...")

linters, formatters := linterAndFormatterCmds()
allCmds := append(linters, formatters...) //nolint:gocritic // intentional append to new slice

var violations []string

// 1. Validate custom-cicd-lint/action.yml covers all linters and formatters.
actionPath := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "actions", "custom-cicd-lint", "action.yml")
violations = append(violations, checkFileCoverage(actionPath, "custom-cicd-lint/action.yml", allCmds)...)

// 2. Validate .pre-commit-config.yaml covers all linters and formatters.
preCommitPath := filepath.Join(rootDir, ".pre-commit-config.yaml")
violations = append(violations, checkFileCoverage(preCommitPath, ".pre-commit-config.yaml", allCmds)...)

// 3. Validate ci-cicd-lint.yml covers all linters and formatters.
workflowPath := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "workflows", "ci-cicd-lint.yml")
violations = append(violations, checkFileCoverage(workflowPath, "ci-cicd-lint.yml", allCmds)...)

if len(violations) > 0 {
return fmt.Errorf("cicd-coverage violations:\n%s", strings.Join(violations, "\n"))
}

logger.Log(fmt.Sprintf("cicd-coverage: all %d linters and %d formatters are covered in all 3 required files",
len(linters), len(formatters)))

return nil
}

// checkFileCoverage verifies that each command in cmds appears as a substring
// in the content of the file at path. Reports all missing commands.
func checkFileCoverage(path, label string, cmds []string) []string {
content, err := os.ReadFile(path)
if err != nil {
return []string{fmt.Sprintf("%s: cannot read file (%v) — ensure the file exists and is readable", label, err)}
}

text := string(content)

var violations []string

for _, cmd := range cmds {
if !strings.Contains(text, cmd) {
violations = append(violations, fmt.Sprintf("%s: missing command %q", label, cmd))
}
}

return violations
}
