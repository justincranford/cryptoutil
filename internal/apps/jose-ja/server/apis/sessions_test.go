// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSessionHandler_NilManager(t *testing.T) {
	t.Parallel()

	// NewSessionHandler with nil manager returns a handler (pass-through to template).
	handler := NewSessionHandler(nil)
	require.NotNil(t, handler)
}
