// Copyright (c) 2025-2026 Justin Cranford.
// Package lint_agent_drift validates that every Copilot agent file in
// .github/agents/ has a matching Claude Code agent in .claude/agents/ with
// identical shared body content. Frontmatter metadata may differ by target,
// but name prefix conventions (copilot- vs claude-) remain enforced.
package lint_agent_drift

import (
	"bytes"
	"fmt"
	"io"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps-tools/cicd_lint/docs_validation"
)

// Check validates that all Copilot agents have matching Claude Code counterparts with
// identical shared body content.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return check(logger, func(stdout, stderr io.Writer) int {
		return cryptoutilDocsValidation.AgentDriftCommand(stdout, stderr)
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
			return fmt.Errorf("lint-agent-drift failed: %s", stderr.String())
		}

		return fmt.Errorf("lint-agent-drift failed: agent drift violations found")
	}

	return nil
}
