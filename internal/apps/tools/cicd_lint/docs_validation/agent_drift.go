// Copyright (c) 2025 Justin Cranford

package docs_validation

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// AgentFrontmatter represents the parsed YAML frontmatter of an agent file.
// Only fields relevant to drift checking are included.
type AgentFrontmatter struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	ArgumentHint string `yaml:"argument-hint"`
}

// AgentDriftViolation describes a single agent drift error.
type AgentDriftViolation struct {
	CopilotFile string
	ClaudeFile  string
	Field       string
	Detail      string
}

// AgentDriftResult holds the result of the agent drift check.
type AgentDriftResult struct {
	Violations []AgentDriftViolation
	Checked    int
}

// splitMarkdownFrontmatter splits content into YAML frontmatter and body.
// Handles both LF and CRLF line endings.
func splitMarkdownFrontmatter(content string) (frontmatter string, body string, err error) {
	// Normalize CRLF to LF.
	content = strings.ReplaceAll(content, "\r\n", "\n")

	const delimiter = "---\n"

	if !strings.HasPrefix(content, delimiter) {
		return "", "", fmt.Errorf("file does not begin with YAML frontmatter delimiter '---'")
	}

	rest := content[len(delimiter):]

	// Find closing delimiter.
	const closingDelim = "\n---\n"

	idx := strings.Index(rest, closingDelim)
	if idx == -1 {
		// Also check for closing delimiter at end of file without trailing newline.
		const closingDelimEOF = "\n---"

		if strings.HasSuffix(rest, closingDelimEOF) {
			frontmatter = rest[:len(rest)-len(closingDelimEOF)]

			return frontmatter, "", nil
		}

		return "", "", fmt.Errorf("cannot find closing YAML frontmatter delimiter '---'")
	}

	frontmatter = rest[:idx]
	body = rest[idx+len(closingDelim):]

	return frontmatter, body, nil
}

// parseAgentFrontmatter extracts the YAML frontmatter struct from agent file content.
func parseAgentFrontmatter(content string) (AgentFrontmatter, string, error) {
	fm, body, err := splitMarkdownFrontmatter(content)
	if err != nil {
		return AgentFrontmatter{}, "", err
	}

	var result AgentFrontmatter
	if err = yaml.Unmarshal([]byte(fm), &result); err != nil {
		return AgentFrontmatter{}, "", fmt.Errorf("YAML parse error: %w", err)
	}

	return result, body, nil
}

// copilotPrefixStr is the required name prefix for Copilot agent files.
const copilotPrefixStr = "copilot-"

// claudePrefixStr is the required name prefix for Claude Code agent files.
const claudePrefixStr = "claude-"

// CheckAgentDrift validates that every Copilot agent file in .github/agents/
// has a matching Claude Code agent in .claude/agents/ with identical description,
// argument-hint, and body. Only the name prefix and Copilot-specific fields
// (tools:, handoffs:, skills:) may differ.
func CheckAgentDrift(rootDir string, readFileFn func(string) ([]byte, error)) (*AgentDriftResult, error) {
	result := &AgentDriftResult{}

	agentsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDGithubAgentsDir)

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", cryptoutilSharedMagic.CICDGithubAgentsDir, err)
	}

	var copilotFiles []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if matched, _ := filepath.Match(cryptoutilSharedMagic.CICDAgentsPattern, entry.Name()); matched {
			copilotFiles = append(copilotFiles, entry.Name())
		}
	}

	sort.Strings(copilotFiles)

	for _, fileName := range copilotFiles {
		copilotRelPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubAgentsDir, fileName))

		copilotContent, readErr := readFileFn(copilotRelPath)
		if readErr != nil {
			result.Violations = append(result.Violations, AgentDriftViolation{
				CopilotFile: copilotRelPath,
				Field:       "file",
				Detail:      fmt.Sprintf("cannot read Copilot agent file: %v", readErr),
			})

			continue
		}

		copilotFM, copilotBody, parseErr := parseAgentFrontmatter(string(copilotContent))
		if parseErr != nil {
			result.Violations = append(result.Violations, AgentDriftViolation{
				CopilotFile: copilotRelPath,
				Field:       "frontmatter",
				Detail:      fmt.Sprintf("Copilot agent parse error: %v", parseErr),
			})

			continue
		}

		// Derive expected Claude agent file path.
		// Copilot: .github/agents/beast-mode.agent.md (name: copilot-beast-mode)
		// Claude:  .claude/agents/beast-mode.md        (name: claude-beast-mode)
		baseName := strings.TrimSuffix(fileName, ".agent.md") // e.g., "beast-mode"
		claudeRelPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDClaudeAgentsDir, baseName+".md"))

		claudeContent, claudeReadErr := readFileFn(claudeRelPath)
		if claudeReadErr != nil {
			result.Violations = append(result.Violations, AgentDriftViolation{
				CopilotFile: copilotRelPath,
				ClaudeFile:  claudeRelPath,
				Field:       "missing",
				Detail:      fmt.Sprintf("Claude Code agent file not found: %s", claudeRelPath),
			})

			result.Checked++

			continue
		}

		claudeFM, claudeBody, claudeParseErr := parseAgentFrontmatter(string(claudeContent))
		if claudeParseErr != nil {
			result.Violations = append(result.Violations, AgentDriftViolation{
				CopilotFile: copilotRelPath,
				ClaudeFile:  claudeRelPath,
				Field:       "frontmatter",
				Detail:      fmt.Sprintf("Claude Code agent parse error: %v", claudeParseErr),
			})

			result.Checked++

			continue
		}

		result.Checked++

		expectedClaudeName := claudePrefixStr + baseName

		// Validate name: Copilot must have copilot- prefix; Claude must have claude- prefix.
		if !strings.HasPrefix(copilotFM.Name, copilotPrefixStr) {
			result.Violations = append(result.Violations, AgentDriftViolation{
				CopilotFile: copilotRelPath,
				ClaudeFile:  claudeRelPath,
				Field:       cryptoutilSharedMagic.CICDAgentFrontMatterNameField,
				Detail:      fmt.Sprintf("Copilot agent name %q must have prefix %q", copilotFM.Name, copilotPrefixStr),
			})
		}

		if claudeFM.Name != expectedClaudeName {
			result.Violations = append(result.Violations, AgentDriftViolation{
				CopilotFile: copilotRelPath,
				ClaudeFile:  claudeRelPath,
				Field:       cryptoutilSharedMagic.CICDAgentFrontMatterNameField,
				Detail:      fmt.Sprintf("Claude Code agent name %q does not match expected %q", claudeFM.Name, expectedClaudeName),
			})
		}

		// Validate description: must be verbatim identical.
		if copilotFM.Description != claudeFM.Description {
			result.Violations = append(result.Violations, AgentDriftViolation{
				CopilotFile: copilotRelPath,
				ClaudeFile:  claudeRelPath,
				Field:       "description",
				Detail: fmt.Sprintf("description mismatch:\n  copilot: %q\n  claude:  %q",
					copilotFM.Description, claudeFM.Description),
			})
		}

		// Validate argument-hint: must be verbatim identical when present.
		if copilotFM.ArgumentHint != claudeFM.ArgumentHint {
			result.Violations = append(result.Violations, AgentDriftViolation{
				CopilotFile: copilotRelPath,
				ClaudeFile:  claudeRelPath,
				Field:       "argument-hint",
				Detail: fmt.Sprintf("argument-hint mismatch:\n  copilot: %q\n  claude:  %q",
					copilotFM.ArgumentHint, claudeFM.ArgumentHint),
			})
		}

		// Validate body: must be verbatim identical.
		normalizedCopilotBody := strings.ReplaceAll(copilotBody, "\r\n", "\n")
		normalizedClaudeBody := strings.ReplaceAll(claudeBody, "\r\n", "\n")

		if normalizedCopilotBody != normalizedClaudeBody {
			result.Violations = append(result.Violations, AgentDriftViolation{
				CopilotFile: copilotRelPath,
				ClaudeFile:  claudeRelPath,
				Field:       "body",
				Detail:      "body content differs between Copilot and Claude Code agent files",
			})
		}
	}

	return result, nil
}

// formatAgentDriftResults formats the agent drift report for output.
func formatAgentDriftResults(result *AgentDriftResult) string {
	var sb strings.Builder

	sb.WriteString("=== Agent Drift Check ===\n\n")
	sb.WriteString(fmt.Sprintf("Checked %d Copilot/Claude Code agent pairs\n\n", result.Checked))

	if len(result.Violations) == 0 {
		sb.WriteString("✅ All agent pairs are in sync\n")

		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("❌ %d violation(s) found:\n\n", len(result.Violations)))

	for i, v := range result.Violations {
		sb.WriteString(fmt.Sprintf("[%d] field=%s\n", i+1, v.Field))

		if v.CopilotFile != "" {
			sb.WriteString(fmt.Sprintf("    copilot: %s\n", v.CopilotFile))
		}

		if v.ClaudeFile != "" {
			sb.WriteString(fmt.Sprintf("    claude:  %s\n", v.ClaudeFile))
		}

		sb.WriteString(fmt.Sprintf("    %s\n\n", v.Detail))
	}

	return sb.String()
}

// AgentDriftCommand runs the agent drift check and writes results to stdout/stderr.
// Returns 0 on success, 1 on violations.
func AgentDriftCommand(stdout, stderr io.Writer) int {
	rootDir, err := findProjectRootFn()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "agent-drift: cannot determine project root: %v\n", err)

		return 1
	}

	readFileFn := func(relPath string) ([]byte, error) {
		return os.ReadFile(filepath.Join(rootDir, filepath.FromSlash(relPath)))
	}

	result, err := CheckAgentDrift(rootDir, readFileFn)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "agent-drift: %v\n", err)

		return 1
	}

	_, _ = fmt.Fprint(stdout, formatAgentDriftResults(result))

	if len(result.Violations) > 0 {
		_, _ = fmt.Fprintf(stderr, "agent-drift: %d violation(s) found\n", len(result.Violations))

		return 1
	}

	return 0
}
