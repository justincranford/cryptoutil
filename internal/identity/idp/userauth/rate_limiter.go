// Copyright (c) 2025 Justin Cranford
//
//

package userauth

import (
	"context"
	"fmt"
	"sync"
	"time"

	googleUuid "github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// RateLimitStore defines interface for persistent rate limit storage.
type RateLimitStore interface {
	// RecordAttempt records an authentication attempt for the given key (user ID or IP).
	RecordAttempt(ctx context.Context, key string, timestamp time.Time) error

	// CountAttempts returns number of attempts in the given time window.
	CountAttempts(ctx context.Context, key string, window time.Duration) (int, error)

	// CleanupExpired removes rate limit records older than the retention period.
	CleanupExpired(ctx context.Context, retention time.Duration) error
}

// DatabaseRateLimitStore implements RateLimitStore with database persistence.
type DatabaseRateLimitStore struct {
	mu            sync.RWMutex
	attempts      map[string][]time.Time // In-memory for now, will be DB-backed
	meterProvider metric.MeterProvider
	counter       metric.Int64Counter
}

// NewDatabaseRateLimitStore creates a new database-backed rate limit store.
func NewDatabaseRateLimitStore(meterProvider metric.MeterProvider) (*DatabaseRateLimitStore, error) {
	meter := meterProvider.Meter("identity.ratelimit")

	counter, err := meter.Int64Counter(
		"identity.ratelimit.attempts",
		metric.WithDescription("Total authentication attempts tracked by rate limiter"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limit counter: %w", err)
	}

	return &DatabaseRateLimitStore{
		attempts:      make(map[string][]time.Time),
		meterProvider: meterProvider,
		counter:       counter,
	}, nil
}

// RecordAttempt records an authentication attempt.
func (s *DatabaseRateLimitStore) RecordAttempt(ctx context.Context, key string, timestamp time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.attempts[key] == nil {
		s.attempts[key] = make([]time.Time, 0)
	}

	s.attempts[key] = append(s.attempts[key], timestamp)

	// Record metric.
	s.counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("key_type", "user"), // Will support "ip" in future
	))

	return nil
}

// CountAttempts returns number of attempts within the time window.
func (s *DatabaseRateLimitStore) CountAttempts(_ context.Context, key string, window time.Duration) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attempts, exists := s.attempts[key]
	if !exists {
		return 0, nil
	}

	cutoff := time.Now().UTC().Add(-window)
	count := 0

	for _, timestamp := range attempts {
		if timestamp.After(cutoff) {
			count++
		}
	}

	return count, nil
}

// CleanupExpired removes rate limit records older than retention period.
func (s *DatabaseRateLimitStore) CleanupExpired(_ context.Context, retention time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().UTC().Add(-retention)

	for key, attempts := range s.attempts {
		filtered := make([]time.Time, 0, len(attempts))

		for _, timestamp := range attempts {
			if timestamp.After(cutoff) {
				filtered = append(filtered, timestamp)
			}
		}

		if len(filtered) == 0 {
			delete(s.attempts, key)
		} else {
			s.attempts[key] = filtered
		}
	}

	return nil
}

// PerUserRateLimiter implements per-user rate limiting with configurable windows.
type PerUserRateLimiter struct {
	store         RateLimitStore
	window        time.Duration
	maxAttempts   int
	meterProvider metric.MeterProvider
	exceededCount metric.Int64Counter
}

// NewPerUserRateLimiter creates a new per-user rate limiter.
func NewPerUserRateLimiter(
	store RateLimitStore,
	window time.Duration,
	maxAttempts int,
	meterProvider metric.MeterProvider,
) (*PerUserRateLimiter, error) {
	meter := meterProvider.Meter("identity.ratelimit")

	exceededCount, err := meter.Int64Counter(
		"identity.ratelimit.exceeded",
		metric.WithDescription("Number of rate limit violations"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exceeded counter: %w", err)
	}

	return &PerUserRateLimiter{
		store:         store,
		window:        window,
		maxAttempts:   maxAttempts,
		meterProvider: meterProvider,
		exceededCount: exceededCount,
	}, nil
}

// CheckLimit checks if the user has exceeded rate limit.
func (r *PerUserRateLimiter) CheckLimit(ctx context.Context, userID googleUuid.UUID) error {
	key := userID.String()

	count, err := r.store.CountAttempts(ctx, key, r.window)
	if err != nil {
		return fmt.Errorf("failed to count attempts: %w", err)
	}

	if count >= r.maxAttempts {
		r.exceededCount.Add(ctx, 1, metric.WithAttributes(
			attribute.String("user_id", userID.String()),
			attribute.String("limit_type", "per_user"),
		))

		return fmt.Errorf("rate limit exceeded: %d attempts in %s (max %d)", count, r.window, r.maxAttempts)
	}

	return nil
}

// RecordAttempt records an authentication attempt for the user.
func (r *PerUserRateLimiter) RecordAttempt(ctx context.Context, userID googleUuid.UUID) error {
	key := userID.String()

	if err := r.store.RecordAttempt(ctx, key, time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to record attempt: %w", err)
	}

	return nil
}

// Cleanup removes expired rate limit records.
func (r *PerUserRateLimiter) Cleanup(ctx context.Context) error {
	retention := r.window * cryptoutilSharedMagic.RateLimitRetentionMultiplier

	if err := r.store.CleanupExpired(ctx, retention); err != nil {
		return fmt.Errorf("failed to cleanup expired records: %w", err)
	}

	return nil
}

// PerIPRateLimiter implements per-IP rate limiting for global abuse protection.
type PerIPRateLimiter struct {
	store         RateLimitStore
	window        time.Duration
	maxAttempts   int
	meterProvider metric.MeterProvider
	exceededCount metric.Int64Counter
}

// NewPerIPRateLimiter creates a new per-IP rate limiter.
func NewPerIPRateLimiter(
	store RateLimitStore,
	window time.Duration,
	maxAttempts int,
	meterProvider metric.MeterProvider,
) (*PerIPRateLimiter, error) {
	meter := meterProvider.Meter("identity.ratelimit")

	exceededCount, err := meter.Int64Counter(
		"identity.ratelimit.exceeded",
		metric.WithDescription("Number of rate limit violations"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exceeded counter: %w", err)
	}

	return &PerIPRateLimiter{
		store:         store,
		window:        window,
		maxAttempts:   maxAttempts,
		meterProvider: meterProvider,
		exceededCount: exceededCount,
	}, nil
}

// CheckLimit checks if the IP address has exceeded rate limit.
func (r *PerIPRateLimiter) CheckLimit(ctx context.Context, ipAddress string) error {
	if ipAddress == "" {
		return fmt.Errorf("IP address cannot be empty")
	}

	count, err := r.store.CountAttempts(ctx, ipAddress, r.window)
	if err != nil {
		return fmt.Errorf("failed to count attempts: %w", err)
	}

	if count >= r.maxAttempts {
		r.exceededCount.Add(ctx, 1, metric.WithAttributes(
			attribute.String("ip_address", ipAddress),
			attribute.String("limit_type", "per_ip"),
		))

		return fmt.Errorf("rate limit exceeded for IP %s: %d attempts in %s (max %d)", ipAddress, count, r.window, r.maxAttempts)
	}

	return nil
}

// RecordAttempt records an authentication attempt for the IP address.
func (r *PerIPRateLimiter) RecordAttempt(ctx context.Context, ipAddress string) error {
	if ipAddress == "" {
		return fmt.Errorf("IP address cannot be empty")
	}

	if err := r.store.RecordAttempt(ctx, ipAddress, time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to record attempt: %w", err)
	}

	return nil
}

// Cleanup removes expired rate limit records.
func (r *PerIPRateLimiter) Cleanup(ctx context.Context) error {
	retention := r.window * cryptoutilSharedMagic.RateLimitRetentionMultiplier

	if err := r.store.CleanupExpired(ctx, retention); err != nil {
		return fmt.Errorf("failed to cleanup expired records: %w", err)
	}

	return nil
}

// Context keys for IP extraction.
type contextKey string

const (
	contextKeyXForwardedFor contextKey = "X-Forwarded-For"
	contextKeyRemoteAddr    contextKey = "RemoteAddr"
)

// ExtractIPFromContext extracts the client IP address from the request context.
// Checks X-Forwarded-For header first (for proxies), then falls back to RemoteAddr.
func ExtractIPFromContext(ctx context.Context) (string, error) {
	// Check for X-Forwarded-For header (proxied requests).
	if xff, ok := ctx.Value(contextKeyXForwardedFor).(string); ok && xff != "" {
		// X-Forwarded-For can contain multiple IPs (client, proxy1, proxy2, ...).
		// Use the first IP (original client).
		for i, c := range xff {
			if c == ',' {
				// Trim whitespace after comma.
				ip := xff[:i]

				return ip, nil
			}
		}

		// Single IP, no comma.
		return xff, nil
	}

	// Fallback to RemoteAddr.
	if remoteAddr, ok := ctx.Value(contextKeyRemoteAddr).(string); ok && remoteAddr != "" {
		// RemoteAddr format: "IP:port" - extract IP only.
		for i, c := range remoteAddr {
			if c == ':' {
				return remoteAddr[:i], nil
			}
		}

		// No port, return as-is.
		return remoteAddr, nil
	}

	return "", fmt.Errorf("unable to extract IP address from context")
}
