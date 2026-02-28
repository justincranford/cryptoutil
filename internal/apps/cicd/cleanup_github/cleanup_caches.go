// Copyright (c) 2025 Justin Cranford

package cleanup_github

import (
json "encoding/json"
"fmt"
"strconv"
"time"
)

// CleanupCaches deletes caches not accessed within MaxAgeDays.
func CleanupCaches(cfg *CleanupConfig) error {
cfg.Logger.Log(fmt.Sprintf("Fetching caches not accessed in %d days...", cfg.MaxAgeDays))

caches, totalSize, err := listCaches(cfg)
if err != nil {
return fmt.Errorf("failed to list caches: %w", err)
}

cfg.Logger.Log(fmt.Sprintf("Total caches: %d (%.2f MB)", len(caches), float64(totalSize)/bytesPerMB))

cutoff := time.Now().AddDate(0, 0, -cfg.MaxAgeDays)

var toDelete []cache

var totalSizeToFree int64

for _, c := range caches {
lastAccessed, parseErr := time.Parse(time.RFC3339, c.LastAccessedAt)
if parseErr != nil {
cfg.Logger.Log(fmt.Sprintf("WARNING: Cannot parse date for cache %d: %v", c.ID, parseErr))

continue
}

if lastAccessed.After(cutoff) {
continue
}

toDelete = append(toDelete, c)

totalSizeToFree += c.SizeBytes
}

cfg.Logger.Log(fmt.Sprintf("Caches eligible for deletion: %d (%.2f MB)", len(toDelete), float64(totalSizeToFree)/bytesPerMB))

if len(toDelete) == 0 {
cfg.Logger.Log("No caches to delete.")

return nil
}

if !cfg.Confirm {
cfg.Logger.Log(fmt.Sprintf("DRY-RUN: Would delete %d caches (%.2f MB). Pass --confirm to execute.", len(toDelete), float64(totalSizeToFree)/bytesPerMB))

return nil
}

deleted := 0
errCount := 0

for _, c := range toDelete {
if delErr := deleteCache(cfg, c.ID); delErr != nil {
cfg.Logger.Log(fmt.Sprintf("ERROR deleting cache %d (%s): %v", c.ID, c.Key, delErr))

errCount++
} else {
deleted++
}
}

cfg.Logger.Log(fmt.Sprintf("Deleted %d caches (%d errors)", deleted, errCount))

if errCount > 0 {
return fmt.Errorf("failed to delete %d caches", errCount)
}

return nil
}

// listCaches fetches all caches using the REST API with pagination.
func listCaches(cfg *CleanupConfig) ([]cache, int64, error) {
var allCaches []cache

var totalSize int64

for page := 1; page <= maxPages; page++ {
args := append([]string{"api"}, repoArgs(cfg)...)
args = append(args,
"-X", "GET",
fmt.Sprintf("/repos/{owner}/{repo}/actions/cache/usage?per_page=%d&page=%d", maxPerPage, page),
)

output, err := ghExec(args...)
if err != nil {
return nil, 0, fmt.Errorf("failed to fetch caches page %d: %w", page, err)
}

var resp cachesResponse
if err := json.Unmarshal(output, &resp); err != nil {
return nil, 0, fmt.Errorf("failed to parse caches page %d: %w", page, err)
}

for _, c := range resp.ActionsCaches {
allCaches = append(allCaches, c)

totalSize += c.SizeBytes
}

if int64(page*maxPerPage) >= resp.TotalCount {
break
}
}

return allCaches, totalSize, nil
}

// deleteCache deletes a single cache by ID.
func deleteCache(cfg *CleanupConfig, cacheID int64) error {
args := append([]string{"api"}, repoArgs(cfg)...)
args = append(args,
"-X", "DELETE",
"/repos/{owner}/{repo}/actions/caches/"+strconv.FormatInt(cacheID, 10),
)

if _, err := ghExec(args...); err != nil {
return fmt.Errorf("failed to delete cache %d: %w", cacheID, err)
}

return nil
}
