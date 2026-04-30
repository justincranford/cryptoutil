// Copyright (c) 2025-2026 Justin Cranford.
package lifecycle

import (
	"context"
	"io"
	"os"
)

// RunWithSignalChan exposes the internal runWithSignalChan function for testing.
// Tests inject a pre-populated signal channel to avoid sending real OS signals
// to the test process, which would terminate the entire test binary.
var RunWithSignalChan = func(
	ctx context.Context,
	stdout, stderr io.Writer,
	startFn func(ctx context.Context) error,
	shutdownFn func(ctx context.Context) error,
	sigChan <-chan os.Signal,
) int {
	return runWithSignalChan(ctx, stdout, stderr, startFn, shutdownFn, sigChan)
}
