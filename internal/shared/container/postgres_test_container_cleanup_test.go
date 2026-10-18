package container

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type partialPostgresContainer struct {
	testcontainers.Container
	terminateFn func(context.Context, ...testcontainers.TerminateOption) error
}

func (c *partialPostgresContainer) Terminate(ctx context.Context, options ...testcontainers.TerminateOption) error {
	return c.terminateFn(ctx, options...)
}

func TestNewPostgresTestContainer_CleansUpPartialStartup(t *testing.T) {
	t.Parallel()

	startupErr := errors.New("injected startup failure")
	cleanupErr := errors.New("injected cleanup failure")
	tests := []struct {
		name                string
		cancelCallerContext bool
		cleanupErr          error
		wantCleanupError    bool
	}{
		{name: containerTestCleanupSurvivesCallerCancellation, cancelCallerContext: true},
		{name: "cleanup error is preserved", cleanupErr: cleanupErr, wantCleanupError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			terminateCalls := 0
			partial := &postgres.PostgresContainer{
				Container: &partialPostgresContainer{
					terminateFn: func(ctx context.Context, _ ...testcontainers.TerminateOption) error {
						terminateCalls++

						require.NoError(t, ctx.Err())
						_, hasDeadline := ctx.Deadline()
						require.True(t, hasDeadline)

						return tc.cleanupErr
					},
				},
			}

			ctx := context.Background()
			if tc.cancelCallerContext {
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()

				ctx = canceledCtx
			}

			container, err := newPostgresTestContainer(ctx, func(context.Context, string, ...testcontainers.ContainerCustomizer) (*postgres.PostgresContainer, error) {
				return partial, startupErr
			})

			require.ErrorIs(t, err, startupErr)
			require.Equal(t, 1, terminateCalls)
			require.Nil(t, container)

			if tc.wantCleanupError {
				require.ErrorIs(t, err, cleanupErr)
				require.ErrorContains(t, err, "failed to clean up container")
			}
		})
	}
}
