// Copyright (c) 2025 Justin Cranford
//
//

package demo

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDemo_NotYetAvailable(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	exitCode := Demo(nil, nil, nil, &stderr)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "not yet available")
}
