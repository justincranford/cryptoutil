// Copyright (c) 2025 Justin Cranford

package cleanup_github

import (
	json "encoding/json"
	"fmt"
	"strconv"
	"time"
)

// apiRun is the REST API response format for workflow runs.
type apiRun struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	CreatedAt  string `json:"created_at"`
	Name       string `json:"name"`
	WorkflowID int64  `json:"workflow_id"`
}

// apiRunsResponse is the REST API response for listing workflow runs.
type apiRunsResponse struct {
	TotalCount   int64    `json:"total_count"`
	WorkflowRuns []apiRun `json:"workflow_runs"`
}

// CleanupRuns deletes workflow runs older than MaxAgeDays.
// Keeps at least KeepMinRuns successful runs per workflow.
func CleanupRuns(cfg *CleanupConfig) error {
	cfg.Logger.Log(fmt.Sprintf("Fetching workflow runs older than %d days...", cfg.MaxAgeDays))

	runs, err := listWorkflowRuns(cfg)
	if err != nil {
		return fmt.Errorf("failed to list workflow runs: %w", err)
	}

	cfg.Logger.Log(fmt.Sprintf("Total workflow runs fetched: %d", len(runs)))

	cutoff := time.Now().AddDate(0, 0, -cfg.MaxAgeDays)

	// Count successful runs per workflow for protection.
	successCountByWorkflow := make(map[string]int)

	for _, run := range runs {
		if run.Conclusion == conclusionSuccess {
			successCountByWorkflow[run.Name]++
		}
	}

	// Determine which runs to delete.
	var toDelete []workflowRun

	successKept := make(map[string]int)

	for _, run := range runs {
		createdAt, parseErr := time.Parse(time.RFC3339, run.CreatedAt)
		if parseErr != nil {
			cfg.Logger.Log(fmt.Sprintf("WARNING: Cannot parse date for run %d: %v", run.DatabaseID, parseErr))

			continue
		}

		if createdAt.After(cutoff) {
			continue
		}

		// Protect minimum successful runs.
		if run.Conclusion == conclusionSuccess && successKept[run.Name] < cfg.KeepMinRuns {
			successKept[run.Name]++

			continue
		}

		toDelete = append(toDelete, run)
	}

	cfg.Logger.Log(fmt.Sprintf("Workflow runs eligible for deletion: %d", len(toDelete)))

	if len(toDelete) == 0 {
		cfg.Logger.Log("No workflow runs to delete.")

		return nil
	}

	if !cfg.Confirm {
		cfg.Logger.Log(fmt.Sprintf("DRY-RUN: Would delete %d workflow runs. Pass --confirm to execute.", len(toDelete)))

		for _, run := range toDelete {
			cfg.Logger.Log(fmt.Sprintf("  [DRY-RUN] Run #%d (%s) from %s - %s", run.DatabaseID, run.Name, run.CreatedAt, run.Conclusion))
		}

		return nil
	}

	deleted := 0
	errCount := 0

	for _, run := range toDelete {
		if delErr := deleteWorkflowRun(cfg, run.DatabaseID); delErr != nil {
			cfg.Logger.Log(fmt.Sprintf("ERROR deleting run %d: %v", run.DatabaseID, delErr))

			errCount++
		} else {
			deleted++
		}
	}

	cfg.Logger.Log(fmt.Sprintf("Deleted %d workflow runs (%d errors)", deleted, errCount))

	if errCount > 0 {
		return fmt.Errorf("failed to delete %d workflow runs", errCount)
	}

	return nil
}

// listWorkflowRuns fetches all workflow runs using the REST API.
func listWorkflowRuns(cfg *CleanupConfig) ([]workflowRun, error) {
	runs, err := listWorkflowRunsAPI(cfg)
	if err != nil {
		return nil, fmt.Errorf("REST API listing failed: %w", err)
	}

	return runs, nil
}

// listWorkflowRunsAPI fetches workflow runs using the REST API with pagination.
func listWorkflowRunsAPI(cfg *CleanupConfig) ([]workflowRun, error) {
	var allRuns []workflowRun

	for page := 1; page <= maxPages; page++ {
		args := append([]string{"api"}, repoArgs(cfg)...)
		args = append(args,
			"-X", "GET",
			fmt.Sprintf("/repos/{owner}/{repo}/actions/runs?per_page=%d&page=%d", maxPerPage, page),
		)

		output, err := ghExec(args...)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
		}

		var resp apiRunsResponse
		if err := json.Unmarshal(output, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse page %d: %w", page, err)
		}

		for _, r := range resp.WorkflowRuns {
			allRuns = append(allRuns, workflowRun{
				DatabaseID: r.ID,
				Status:     r.Status,
				Conclusion: r.Conclusion,
				CreatedAt:  r.CreatedAt,
				Name:       r.Name,
				WorkflowID: r.WorkflowID,
			})
		}

		if int64(page*maxPerPage) >= resp.TotalCount {
			break
		}
	}

	return allRuns, nil
}

// deleteWorkflowRun deletes a single workflow run by ID.
func deleteWorkflowRun(cfg *CleanupConfig, runID int64) error {
	args := append([]string{"api"}, repoArgs(cfg)...)
	args = append(args,
		"-X", "DELETE",
		"/repos/{owner}/{repo}/actions/runs/"+strconv.FormatInt(runID, 10),
	)

	if _, err := ghExec(args...); err != nil {
		return fmt.Errorf("failed to delete run %d: %w", runID, err)
	}

	return nil
}
