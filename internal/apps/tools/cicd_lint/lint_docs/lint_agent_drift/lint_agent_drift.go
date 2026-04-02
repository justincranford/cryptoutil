// Copyright (c) 2025 Justin Cranford

// Package lint_agent_drift validates that every Copilot agent file in
// .github/agents/ has a matching Claude Code agent in .claude/agents/ with
// identical description, argument-hint, and body. Only the name prefix
// (copilot- vs claude-) and Copilot-only fields (tools:, handoffs:, skills:)
// may differ.
package lint_agent_drift

import (
	"bytes"
	"fmt"
	"io"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps/tools/cicd_lint/docs_validation"
)

// agentDriftFn is the seam for testing, replacing AgentDriftCommand.
var agentDriftFn = func(stdout, stderr io.Writer) int {
	return cryptoutilDocsValidation.AgentDriftCommand(stdout, stderr)
}

// Check validates that all Copilot agents have matching Claude Code counterparts with
// identical description, argument-hint, and body content.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	var stdout, stderr bytes.Buffer

	exitCode := agentDriftFn(&stdout, &stderr)

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
