// Copyright (c) 2025-2026 Justin Cranford.
package testutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeBufferWrite_ReportsWriterError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected buffer error")
	n, err := safeBufferWrite([]byte("buffer"), func([]byte) (int, error) {
		return 2, injectedErr
	})

	require.Equal(t, 2, n)
	require.ErrorIs(t, err, injectedErr)
	require.ErrorContains(t, err, "safe buffer write failed")
}
