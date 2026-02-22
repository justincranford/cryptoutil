// Copyright (c) 2025 Justin Cranford

package apperr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequireNoError_NilError(t *testing.T) {
	t.Parallel()

	// Should not panic when error is nil.
	require.NotPanics(t, func() {
		RequireNoError(nil, "should not panic")
	})
}

func TestRequireNoError_NonNilError(t *testing.T) {
	t.Parallel()

	// Should panic when error is non-nil.
	require.PanicsWithValue(t, "test message: test error", func() {
		RequireNoError(errors.New("test error"), "test message")
	})
}
