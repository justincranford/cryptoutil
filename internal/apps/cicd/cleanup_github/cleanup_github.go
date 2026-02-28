// Copyright (c) 2025 Justin Cranford

// Package cleanup_github provides GitHub Actions storage cleanup automation.
// It uses the gh CLI to delete old workflow runs, artifacts, and caches.
//
// IMPORTANT: All destructive operations require --confirm flag.
// Default mode is dry-run (preview only).
package cleanup_github

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

const (
	// defaultMaxAgeDays is the default age threshold for cleanup operations.
	defaultMaxAgeDays = 30

	// defaultKeepMinRuns is the minimum number of successful runs to keep per workflow.
	defaultKeepMinRuns = 10

	// defaultRepo is the default repository (auto-detected by gh CLI).
	defaultRepo = ""

	// ghBinary is the gh CLI binary name.
	ghBinary = "gh"

	// maxPerPage is the maximum items per page for GitHub API pagination.
	maxPerPage = 100

	// maxPages is the maximum number of pages to fetch.
	maxPages = 50

	// conclusionSuccess is the GitHub Actions conclusion value for successful runs.
	conclusionSuccess = "success"

	// bytesPerMB is the number of bytes in a megabyte.
	bytesPerMB = 1024 * 1024

	// ghCommandTimeout is the maximum time to wait for a gh CLI command.
	ghCommandTimeout = 30 * time.Second
)

// CleanupConfig holds configuration for cleanup operations.
type CleanupConfig struct {
	// Repo is the GitHub repository in "owner/repo" format.
	// Empty string means auto-detect from current directory.
	Repo string

	// MaxAgeDays is the age threshold in days. Items older than this are eligible for deletion.
	MaxAgeDays int

	// KeepMinRuns is the minimum number of successful runs to keep per workflow.
	KeepMinRuns int

	// Confirm enables actual deletion. When false (default), operates in dry-run mode.
	Confirm bool

	// Logger is the CICD logger for structured output.
	Logger *cryptoutilCmdCicdCommon.Logger
}

// NewDefaultConfig creates a CleanupConfig with default values.
func NewDefaultConfig(logger *cryptoutilCmdCicdCommon.Logger) *CleanupConfig {
	return &CleanupConfig{
		Repo:        defaultRepo,
		MaxAgeDays:  defaultMaxAgeDays,
		KeepMinRuns: defaultKeepMinRuns,
		Confirm:     false,
		Logger:      logger,
	}
}

// workflowRun represents a GitHub Actions workflow run.
type workflowRun struct {
	DatabaseID int64  `json:"databaseId"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	CreatedAt  string `json:"createdAt"`
	Name       string `json:"name"`
	WorkflowID int64  `json:"workflowDatabaseId"`
}

// artifact represents a GitHub Actions artifact.
type artifact struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	SizeBytes int64  `json:"size_in_bytes"`
	CreatedAt string `json:"created_at"`
	Expired   bool   `json:"expired"`
}

// artifactsResponse is the API response for listing artifacts.
type artifactsResponse struct {
	TotalCount int64      `json:"total_count"`
	Artifacts  []artifact `json:"artifacts"`
}

// cache represents a GitHub Actions cache entry.
type cache struct {
	ID             int64  `json:"id"`
	Key            string `json:"key"`
	Ref            string `json:"ref"`
	SizeBytes      int64  `json:"size_in_bytes"`
	LastAccessedAt string `json:"last_accessed_at"`
	CreatedAt      string `json:"created_at"`
}

// cachesResponse is the API response for listing caches.
type cachesResponse struct {
	TotalCount    int64   `json:"total_count"`
	ActionsCaches []cache `json:"actions_caches"`
}

// CleanupAll runs all cleanup operations (runs + artifacts + caches).
func CleanupAll(cfg *CleanupConfig) error {
	cfg.Logger.Log("Starting full cleanup (runs + artifacts + caches)")

	var errs []string

	if err := CleanupRuns(cfg); err != nil {
		errs = append(errs, fmt.Sprintf("runs: %v", err))
	}

	if err := CleanupArtifacts(cfg); err != nil {
		errs = append(errs, fmt.Sprintf("artifacts: %v", err))
	}

	if err := CleanupCaches(cfg); err != nil {
		errs = append(errs, fmt.Sprintf("caches: %v", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errs, "; "))
	}

	cfg.Logger.Log("Full cleanup completed successfully")

	return nil
}

// ParseArgs parses command-line arguments for cleanup commands.
// Supported flags: --confirm, --max-age-days=N, --keep-min-runs=N, --repo=owner/repo.
func ParseArgs(args []string, cfg *CleanupConfig) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--confirm":
			cfg.Confirm = true
		case strings.HasPrefix(arg, "--max-age-days="):
			val := strings.TrimPrefix(arg, "--max-age-days=")

			days, err := strconv.Atoi(val)
			if err != nil || days < 1 {
				return fmt.Errorf("invalid --max-age-days value: %s (must be positive integer)", val)
			}

			cfg.MaxAgeDays = days
		case strings.HasPrefix(arg, "--keep-min-runs="):
			val := strings.TrimPrefix(arg, "--keep-min-runs=")

			count, err := strconv.Atoi(val)
			if err != nil || count < 0 {
				return fmt.Errorf("invalid --keep-min-runs value: %s (must be non-negative integer)", val)
			}

			cfg.KeepMinRuns = count
		case strings.HasPrefix(arg, "--repo="):
			cfg.Repo = strings.TrimPrefix(arg, "--repo=")
		default:
			return fmt.Errorf("unknown flag: %s", arg)
		}
	}

	return nil
}
