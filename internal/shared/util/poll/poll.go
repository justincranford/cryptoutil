// Copyright (c) 2025 Justin Cranford
//
//

// Package poll provides a parameterized polling loop for waiting on conditions with context support.
package poll

import (
	"context"
	"fmt"
	"time"
)

// ConditionFunc is a function that returns true when the polled condition is met.
// It may also return an error if the check itself fails fatally (non-retryable).
type ConditionFunc func(ctx context.Context) (done bool, err error)

// Until polls conditionFn at the given interval until it returns true, the context is canceled,
// or the timeout elapses. Returns nil when the condition is met, or an error otherwise.
func Until(ctx context.Context, timeout time.Duration, interval time.Duration, conditionFn ConditionFunc) error {
	deadline := time.Now().UTC().Add(timeout)

	for time.Now().UTC().Before(deadline) {
		done, err := conditionFn(ctx)
		if err != nil {
			return fmt.Errorf("poll condition failed: %w", err)
		}

		if done {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("poll canceled: %w", ctx.Err())
		case <-time.After(interval):
			// Continue polling.
		}
	}

	return fmt.Errorf("poll timed out after %v", timeout)
}
