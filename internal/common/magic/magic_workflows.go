// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "regexp"

// Status constants for workflow execution results.
const (
	// StatusSuccess - Success status string with emoji.
	StatusSuccess = "✅ SUCCESS"
	// StatusFailed - Failed status string with emoji.
	StatusFailed = "❌ FAILED"
	// TaskSuccess - Task success status.
	TaskSuccess = "SUCCESS"
	// TaskFailed - Task failed status.
	TaskFailed = "FAILED"
)

// Workflow names for GitHub Actions workflows.
const (
	// WorkflowNameDAST - DAST (Dynamic Application Security Testing) workflow name.
	WorkflowNameDAST = "dast"
	// WorkflowNameLoad - Load testing workflow name.
	WorkflowNameLoad = "load"
	// WorkflowsDir - Directory containing GitHub Actions workflow files.
	WorkflowsDir = ".github/workflows"
)

// Event types for GitHub Actions workflow dispatch.
const (
	// EventTypePush - Push event type for workflows.
	EventTypePush = "push"
	// EventTypeWorkflowDispatch - Workflow dispatch event type for workflows.
	EventTypeWorkflowDispatch = "workflow_dispatch"
)

// Regex patterns for workflow validation.
var (
	// RegexWorkflowActionUses - Regex to match "uses: owner/repo@version" patterns in GitHub Actions workflows.
	RegexWorkflowActionUses = regexp.MustCompile(`uses:\s*([^\s@]+)@([^\s]+)`)
	// RegexWorkflowName - Regex to match top-level "name:" field in workflow files.
	RegexWorkflowName = regexp.MustCompile(`(?m)^\s*name:\s*.+`)
)
