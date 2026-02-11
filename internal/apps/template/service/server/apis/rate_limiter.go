// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package apis provides HTTP handlers and routing for template service APIs.
package apis

import (
	"sync"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// RateLimiter provides in-memory rate limiting per IP address.
// Uses token bucket algorithm for request throttling.
type RateLimiter struct {
	mu             sync.RWMutex
	buckets        map[string]*tokenBucket
	requestsPerMin int
	burstSize      int
	cleanupTicker  *time.Ticker
	stopCleanup    chan struct{}
}

// tokenBucket represents a token bucket for a single IP address.
type tokenBucket struct {
	tokens         int
	lastRefillTime time.Time
}

// NewRateLimiter creates a new rate limiter.
// requestsPerMin: Maximum requests per minute per IP.
// burstSize: Maximum burst requests (tokens available at start).
func NewRateLimiter(requestsPerMin, burstSize int) *RateLimiter {
	rl := &RateLimiter{
		buckets:        make(map[string]*tokenBucket),
		requestsPerMin: requestsPerMin,
		burstSize:      burstSize,
		cleanupTicker:  time.NewTicker(cryptoutilSharedMagic.RateLimitCleanupIntervalMinutes * time.Minute),
		stopCleanup:    make(chan struct{}),
	}

	// Start cleanup goroutine to remove stale buckets.
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request from the given IP should be allowed.
// Returns true if allowed, false if rate limit exceeded.
func (rl *RateLimiter) Allow(ipAddress string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[ipAddress]
	if !exists {
		// First request from this IP - create bucket with full tokens.
		bucket = &tokenBucket{
			tokens:         rl.burstSize,
			lastRefillTime: time.Now().UTC(),
		}
		rl.buckets[ipAddress] = bucket
	}

	// Refill tokens based on time elapsed.
	now := time.Now().UTC()
	elapsed := now.Sub(bucket.lastRefillTime)
	tokensToAdd := int(elapsed.Seconds() * float64(rl.requestsPerMin) / cryptoutilSharedMagic.RateLimitSecondsPerMinute)

	if tokensToAdd > 0 {
		bucket.tokens += tokensToAdd
		if bucket.tokens > rl.burstSize {
			bucket.tokens = rl.burstSize
		}

		bucket.lastRefillTime = now
	}

	// Check if tokens available.
	if bucket.tokens > 0 {
		bucket.tokens--

		return true
	}

	return false
}

// cleanupLoop removes stale buckets that haven't been used in 10 minutes.
func (rl *RateLimiter) cleanupLoop() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

// cleanup removes buckets that haven't been used in 10 minutes.
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().UTC()
	threshold := cryptoutilSharedMagic.RateLimitStaleThresholdMinutes * time.Minute

	for ip, bucket := range rl.buckets {
		if now.Sub(bucket.lastRefillTime) > threshold {
			delete(rl.buckets, ip)
		}
	}
}

// Stop stops the cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
	rl.cleanupTicker.Stop()
}
