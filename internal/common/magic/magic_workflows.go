// Package magic provides commonly used magic numbers and values as named constants.
// This file contains workflow-related constants.
package magic

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
)

// Event types for GitHub Actions workflow dispatch.
const (
	// EventTypePush - Push event type for workflows.
	EventTypePush = "push"
	// EventTypeWorkflowDispatch - Workflow dispatch event type for workflows.
	EventTypeWorkflowDispatch = "workflow_dispatch"
)
