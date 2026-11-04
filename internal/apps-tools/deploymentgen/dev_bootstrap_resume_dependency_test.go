// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testResumeNilContext       = "nil-context"
	testResumeNilRunner        = "nil-runner"
	testResumeUnboundedContext = "unbounded-context"
	testResumeCanceledContext  = "canceled-context"
	testResumeDifferentRun     = "different-run"
)

func TestDevelopmentRetainedVolumeCaptureDependencies(t *testing.T) {
	t.Parallel()

	for _, failure := range []string{testResumeNilContext, testResumeNilRunner, testResumeUnboundedContext, testResumeCanceledContext, testResumeDifferentRun, "empty-engine"} {
		t.Run(failure, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			plan := LocalK3sPlan{Run: DevelopmentSmokeRun}
			run := DockerCommandRunner(func(_ context.Context, _ string, _ ...string) ([]byte, error) { return nil, nil })

			switch failure {
			case testResumeNilContext:
				ctx = nil
			case testResumeNilRunner:
				run = nil
			case testResumeUnboundedContext:
				ctx = context.Background()
			case testResumeCanceledContext:
				cancel()
			case testResumeDifferentRun:
				plan.Run = "another-run"
			}

			snapshot, err := CaptureDevelopmentRetainedVolumes(ctx, plan, run)
			require.Error(t, err)
			require.Empty(t, snapshot.Volumes)
		})
	}
}
