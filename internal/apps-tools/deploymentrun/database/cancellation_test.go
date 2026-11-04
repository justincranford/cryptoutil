// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecutePreservesCancellationAfterOpening(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := Execute(ctx, Migrate, inputFixture(Migrate), func(context.Context, Connection, string) (Client, error) {
		cancel()

		return nil, errors.New("DO-NOT-LOG")
	})
	require.ErrorIs(t, err, context.Canceled)
	require.NotContains(t, err.Error(), "DO-NOT-LOG")
}
