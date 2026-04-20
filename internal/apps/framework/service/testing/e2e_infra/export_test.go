// Copyright (c) 2025 Justin Cranford
//
//

package e2e_infra

import "time"

// DefaultHealthPollInterval exposes the package-level polling cadence for test overrides.
// Sequential tests may set this to a small value (e.g. 10ms) and restore it after.
var DefaultHealthPollInterval = &defaultHealthPollInterval

// SetDefaultHealthPollInterval overrides the poll interval and returns a restore function.
// Callers MUST use Sequential tests (no t.Parallel()) when using this.
func SetDefaultHealthPollInterval(d time.Duration) func() {
	old := defaultHealthPollInterval
	defaultHealthPollInterval = d

	return func() { defaultHealthPollInterval = old }
}
