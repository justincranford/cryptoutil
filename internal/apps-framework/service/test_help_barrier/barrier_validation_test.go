// Copyright (c) 2025-2026 Justin Cranford.

package test_help_barrier

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTestBarrierService_NilDatabase(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		NewTestBarrierService(t, nil)
	})
}
