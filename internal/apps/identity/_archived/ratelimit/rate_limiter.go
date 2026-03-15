// Copyright (c) 2025 Justin Cranford

// Package ratelimit provides rate limiting functionality for identity services.
package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter tracks request counts per key with time windows.
type RateLimiter struct {
	mu         sync.RWMutex
	requests   map[string][]time.Time // key -> timestamps of requests.
	maxCount   int                    // maximum requests per window.
	windowSize time.Duration          // time window duration.
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(maxCount int, windowSize time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:   make(map[string][]time.Time),
		maxCount:   maxCount,
		windowSize: windowSize,
	}
}

// Allow checks if a request is allowed for the given key.
func (rl *RateLimiter) Allow(key string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().UTC()
	windowStart := now.Add(-rl.windowSize)

	// Clean up old requests outside the window.
	if timestamps, exists := rl.requests[key]; exists {
		validTimestamps := []time.Time{}

		for _, ts := range timestamps {
			if ts.After(windowStart) {
				validTimestamps = append(validTimestamps, ts)
			}
		}

		rl.requests[key] = validTimestamps
	}

	// Check if limit exceeded.
	if len(rl.requests[key]) >= rl.maxCount {
		return fmt.Errorf("rate limit exceeded: %d requests allowed per %v", rl.maxCount, rl.windowSize)
	}

	// Record this request.
	rl.requests[key] = append(rl.requests[key], now)

	return nil
}

// Reset clears all rate limit data for a given key.
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.requests, key)
}

// GetCount returns the current request count for a key within the window.
func (rl *RateLimiter) GetCount(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now().UTC()
	windowStart := now.Add(-rl.windowSize)

	count := 0

	if timestamps, exists := rl.requests[key]; exists {
		for _, ts := range timestamps {
			if ts.After(windowStart) {
				count++
			}
		}
	}

	return count
}
