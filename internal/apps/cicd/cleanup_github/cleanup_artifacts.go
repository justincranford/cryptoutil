// Copyright (c) 2025 Justin Cranford

package cleanup_github

import (
json "encoding/json"
"fmt"
"strconv"
"time"
)

// CleanupArtifacts deletes artifacts older than MaxAgeDays.
func CleanupArtifacts(cfg *CleanupConfig) error {
cfg.Logger.Log(fmt.Sprintf("Fetching artifacts older than %d days...", cfg.MaxAgeDays))

artifacts, totalSize, err := listArtifacts(cfg)
if err != nil {
return fmt.Errorf("failed to list artifacts: %w", err)
}

cfg.Logger.Log(fmt.Sprintf("Total artifacts: %d (%.2f MB)", len(artifacts), float64(totalSize)/bytesPerMB))

cutoff := time.Now().AddDate(0, 0, -cfg.MaxAgeDays)

var toDelete []artifact

var totalSizeToFree int64

for _, a := range artifacts {
createdAt, parseErr := time.Parse(time.RFC3339, a.CreatedAt)
if parseErr != nil {
cfg.Logger.Log(fmt.Sprintf("WARNING: Cannot parse date for artifact %d: %v", a.ID, parseErr))

continue
}

if createdAt.After(cutoff) {
continue
}

toDelete = append(toDelete, a)

totalSizeToFree += a.SizeBytes
}

cfg.Logger.Log(fmt.Sprintf("Artifacts eligible for deletion: %d (%.2f MB)", len(toDelete), float64(totalSizeToFree)/bytesPerMB))

if len(toDelete) == 0 {
cfg.Logger.Log("No artifacts to delete.")

return nil
}

if !cfg.Confirm {
cfg.Logger.Log(fmt.Sprintf("DRY-RUN: Would delete %d artifacts (%.2f MB). Pass --confirm to execute.", len(toDelete), float64(totalSizeToFree)/bytesPerMB))

return nil
}

deleted := 0
errCount := 0

for _, a := range toDelete {
if delErr := deleteArtifact(cfg, a.ID); delErr != nil {
cfg.Logger.Log(fmt.Sprintf("ERROR deleting artifact %d: %v", a.ID, delErr))

errCount++
} else {
deleted++
}
}

cfg.Logger.Log(fmt.Sprintf("Deleted %d artifacts (%d errors)", deleted, errCount))

if errCount > 0 {
return fmt.Errorf("failed to delete %d artifacts", errCount)
}

return nil
}

// listArtifacts fetches all artifacts using the REST API with pagination.
func listArtifacts(cfg *CleanupConfig) ([]artifact, int64, error) {
var allArtifacts []artifact

var totalSize int64

for page := 1; page <= maxPages; page++ {
args := append([]string{"api"}, repoArgs(cfg)...)
args = append(args,
"-X", "GET",
fmt.Sprintf("/repos/{owner}/{repo}/actions/artifacts?per_page=%d&page=%d", maxPerPage, page),
)

output, err := ghExec(args...)
if err != nil {
return nil, 0, fmt.Errorf("failed to fetch artifacts page %d: %w", page, err)
}

var resp artifactsResponse
if err := json.Unmarshal(output, &resp); err != nil {
return nil, 0, fmt.Errorf("failed to parse artifacts page %d: %w", page, err)
}

for _, a := range resp.Artifacts {
allArtifacts = append(allArtifacts, a)

totalSize += a.SizeBytes
}

if int64(page*maxPerPage) >= resp.TotalCount {
break
}
}

return allArtifacts, totalSize, nil
}

// deleteArtifact deletes a single artifact by ID.
func deleteArtifact(cfg *CleanupConfig, artifactID int64) error {
args := append([]string{"api"}, repoArgs(cfg)...)
args = append(args,
"-X", "DELETE",
"/repos/{owner}/{repo}/actions/artifacts/"+strconv.FormatInt(artifactID, 10),
)

if _, err := ghExec(args...); err != nil {
return fmt.Errorf("failed to delete artifact %d: %w", artifactID, err)
}

return nil
}
