// Copyright (c) 2025 Justin Cranford

// Package lint_skill_command_drift validates that every Copilot skill in
// .github/skills/NAME/ has a matching Claude Code command at
// .claude/commands/NAME.md and that each Claude command file references its
// corresponding Copilot skill file.
package lint_skill_command_drift

import (
	"bytes"
	"fmt"
	"io"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps/tools/cicd_lint/docs_validation"
)

// Check validates that all Copilot skills have matching Claude Code commands and
// that each Claude command references its Copilot skill source.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return check(logger, func(stdout, stderr io.Writer) int {
		return cryptoutilDocsValidation.SkillCommandDriftCommand(stdout, stderr)
	})
}

func check(logger *cryptoutilCmdCicdCommon.Logger, fn func(io.Writer, io.Writer) int) error {
	var stdout, stderr bytes.Buffer

	exitCode := fn(&stdout, &stderr)

	if stdout.Len() > 0 {
		logger.Log(stdout.String())
	}

	if exitCode != 0 {
		if stderr.Len() > 0 {
			return fmt.Errorf("lint-skill-command-drift failed: %s", stderr.String())
		}

		return fmt.Errorf("lint-skill-command-drift failed: skill/command drift violations found")
	}

	return nil
}
