// Copyright (c) 2025-2026 Justin Cranford.

package container

import (
	"context"
	"errors"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	containerTestCleanupSurvivesCallerCancellation = "cleanup survives caller cancellation"
	containerTestConnectionStringFailure           = "connection string failure"
)

func newContainerTestPassword(t *testing.T) string {
	t.Helper()

	return googleUuid.Must(googleUuid.NewV7()).String()
}

func TestStartContainer_CleansPartialContainerOnStartupError(t *testing.T) {
	t.Parallel()

	startupErr := errors.New("injected generic container startup error")
	cleanupErr := errors.New("injected generic container cleanup error")

	tests := []struct {
		name             string
		cancelContext    bool
		cleanupError     error
		wantCleanupError bool
	}{
		{name: containerTestCleanupSurvivesCallerCancellation, cancelContext: true},
		{name: "cleanup error is retained", cleanupError: cleanupErr, wantCleanupError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			telemetryService := setupTestTelemetry(t)
			terminateCalls := 0
			partialContainer := &mockContainer{
				terminateFn: func(ctx context.Context, _ ...testcontainers.TerminateOption) error {
					terminateCalls++

					require.NoError(t, ctx.Err())
					_, hasDeadline := ctx.Deadline()
					require.True(t, hasDeadline)

					return tc.cleanupError
				},
			}

			ctx := context.Background()
			if tc.cancelContext {
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()

				ctx = canceledCtx
			}

			container, cleanup, err := startContainer(
				ctx,
				telemetryService,
				testcontainers.ContainerRequest{},
				func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
					return partialContainer, startupErr
				},
			)

			require.Nil(t, container)
			require.Nil(t, cleanup)
			require.ErrorIs(t, err, startupErr)
			require.Equal(t, 1, terminateCalls)

			if tc.wantCleanupError {
				require.ErrorIs(t, err, cleanupErr)
				require.ErrorContains(t, err, "failed to clean up container")
			}
		})
	}
}

func TestStartContainer_CleanupSurvivesCallerCancellation(t *testing.T) {
	t.Parallel()

	telemetryService := setupTestTelemetry(t)
	terminateCalls := 0
	startedContainer := &mockContainer{
		terminateFn: func(ctx context.Context, _ ...testcontainers.TerminateOption) error {
			terminateCalls++

			require.NoError(t, ctx.Err())
			_, hasDeadline := ctx.Deadline()
			require.True(t, hasDeadline)

			return nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())

	container, cleanup, err := startContainer(
		ctx,
		telemetryService,
		testcontainers.ContainerRequest{},
		func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
			return startedContainer, nil
		},
	)
	require.NoError(t, err)
	require.Same(t, startedContainer, container)
	require.NotNil(t, cleanup)

	cancel()
	cleanup()

	require.Equal(t, 1, terminateCalls)
}

func TestStartContainer_RejectsNilContainer(t *testing.T) {
	t.Parallel()

	telemetryService := setupTestTelemetry(t)

	var nilPointerContainer *mockContainer

	tests := []struct {
		name      string
		container testcontainers.Container
	}{
		{name: "nil interface"},
		{name: "typed nil pointer", container: nilPointerContainer},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			container, cleanup, err := startContainer(
				context.Background(),
				telemetryService,
				testcontainers.ContainerRequest{},
				func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
					return tc.container, nil
				},
			)

			require.Nil(t, container)
			require.Nil(t, cleanup)
			require.ErrorContains(t, err, "container startup returned nil container")
		})
	}
}

func TestSetupSharedPostgresContainer_CleansContainerAfterErrors(t *testing.T) {
	t.Parallel()

	startupErr := errors.New("injected PostgreSQL startup error")
	connectionStringErr := errors.New("injected connection string error")
	cleanupErr := errors.New("injected PostgreSQL cleanup error")

	tests := []struct {
		name                  string
		runErr                error
		connectionStringErr   error
		cleanupError          error
		wantCleanupError      bool
		wantConnectionStrings int
	}{
		{name: "partial container returned with startup error", runErr: startupErr},
		{
			name:                  containerTestConnectionStringFailure,
			connectionStringErr:   connectionStringErr,
			wantConnectionStrings: 1,
		},
		{
			name:             "startup and cleanup errors are retained",
			runErr:           startupErr,
			cleanupError:     cleanupErr,
			wantCleanupError: true,
		},
		{
			name:                  "connection string and cleanup errors are retained",
			connectionStringErr:   connectionStringErr,
			cleanupError:          cleanupErr,
			wantCleanupError:      true,
			wantConnectionStrings: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			terminateCalls := 0
			partialContainer := &postgres.PostgresContainer{
				Container: &mockContainer{
					terminateFn: func(ctx context.Context, _ ...testcontainers.TerminateOption) error {
						terminateCalls++

						require.NoError(t, ctx.Err())

						_, hasDeadline := ctx.Deadline()
						require.True(t, hasDeadline)

						return tc.cleanupError
					},
				},
			}
			connectionStringCalls := 0

			container, connection, err := setupSharedPostgresContainer(
				context.Background(),
				func() (string, error) { return newContainerTestPassword(t), nil },
				func(context.Context, string, ...testcontainers.ContainerCustomizer) (*postgres.PostgresContainer, error) {
					return partialContainer, tc.runErr
				},
				func(context.Context, *postgres.PostgresContainer) (string, error) {
					connectionStringCalls++

					return "", tc.connectionStringErr
				},
			)

			require.Nil(t, container)
			require.Empty(t, connection)
			require.Error(t, err)
			require.Equal(t, tc.wantConnectionStrings, connectionStringCalls)
			require.Equal(t, 1, terminateCalls)

			switch {
			case tc.runErr != nil:
				require.ErrorIs(t, err, startupErr)
			case tc.connectionStringErr != nil:
				require.ErrorIs(t, err, connectionStringErr)
			}

			if tc.wantCleanupError {
				require.ErrorIs(t, err, cleanupErr)
				require.ErrorContains(t, err, "failed to clean up PostgreSQL container")
			}
		})
	}
}
