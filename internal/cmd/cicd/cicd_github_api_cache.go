// Package cicd provides GitHub API caching functionality for CI/CD operations.
//
// This file contains the caching layer for GitHub API responses to improve
// performance of workflow linting operations by reducing external API calls.
package cicd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// GitHubRelease represents a GitHub release API response.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// GitHubAPICacheEntry represents a cached API response with expiration.
type GitHubAPICacheEntry struct {
	Value     string
	ExpiresAt time.Time
}

// GitHubAPICache provides thread-safe caching of GitHub API responses.
type GitHubAPICache struct {
	mu    sync.RWMutex
	cache map[string]GitHubAPICacheEntry
}

// NewGitHubAPICache creates a new GitHub API cache instance.
func NewGitHubAPICache() *GitHubAPICache {
	return &GitHubAPICache{
		cache: make(map[string]GitHubAPICacheEntry),
	}
}

// Get retrieves a value from the cache if it exists and hasn't expired.
func (c *GitHubAPICache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return "", false
	}

	if time.Now().After(entry.ExpiresAt) {
		// Entry has expired, remove it
		delete(c.cache, key)

		return "", false
	}

	return entry.Value, true
}

// Set stores a value in the cache with the configured TTL.
func (c *GitHubAPICache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = GitHubAPICacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(cryptoutilMagic.TimeoutGitHubAPICacheTTL),
	}
}

// Global GitHub API cache instance.
var githubAPICache = NewGitHubAPICache()

// getLatestVersion retrieves the latest version of a GitHub action.
// It uses caching to avoid repeated API calls for the same action.
func getLatestVersion(logger *LogUtil, actionName string) (string, error) {
	// Check cache first
	cacheKey := "release:" + actionName
	if cachedVersion, found := githubAPICache.Get(cacheKey); found {
		logger.Log(fmt.Sprintf("Cache hit for %s: %s", actionName, cachedVersion))

		return cachedVersion, nil
	}

	logger.Log(fmt.Sprintf("Cache miss for %s, making API call", actionName))

	// GitHub API has rate limits, so add a delay
	time.Sleep(cryptoutilMagic.TimeoutGitHubAPIDelay)

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", actionName)

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilMagic.TimeoutGitHubAPITimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Use GitHub token if available to increase rate limit
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Set User-Agent as recommended by GitHub API
	req.Header.Set("User-Agent", "check-script")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make HTTP request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close HTTP response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		// Some actions might not have releases, try tags
		version, err := getLatestTag(logger, actionName)
		if err != nil {
			return "", err
		}
		// Cache the result
		githubAPICache.Set(cacheKey, version)

		return version, nil
	} else if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("GitHub API rate limit exceeded (403). Set GITHUB_TOKEN environment variable to increase limit")
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to unmarshal GitHub release JSON: %w", err)
	}

	// Cache the result
	githubAPICache.Set(cacheKey, release.TagName)

	return release.TagName, nil
}

// getLatestTag retrieves the latest tag for a GitHub repository.
// Used as a fallback when releases are not available.
func getLatestTag(logger *LogUtil, actionName string) (string, error) {
	// Check cache first
	cacheKey := "tags:" + actionName
	if cachedVersion, found := githubAPICache.Get(cacheKey); found {
		logger.Log(fmt.Sprintf("Cache hit for %s tags: %s", actionName, cachedVersion))

		return cachedVersion, nil
	}

	logger.Log(fmt.Sprintf("Cache miss for %s tags, making API call", actionName))

	url := fmt.Sprintf("https://api.github.com/repos/%s/tags", actionName)

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilMagic.TimeoutGitHubAPITimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request for tags: %w", err)
	}

	// Use GitHub token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	req.Header.Set("User-Agent", "check-script")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make HTTP request for tags: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close HTTP response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("GitHub API rate limit exceeded (403). Set GITHUB_TOKEN environment variable to increase limit")
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read tags response body: %w", err)
	}

	var tags []struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(body, &tags); err != nil {
		return "", fmt.Errorf("failed to unmarshal GitHub tags JSON: %w", err)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found")
	}

	// Return the first tag (should be the latest)
	latestTag := tags[0].Name

	// Cache the result
	githubAPICache.Set(cacheKey, latestTag)

	return latestTag, nil
}
