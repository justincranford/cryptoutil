// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdp_MainHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Idp([]string{"--help"}, nil, &stdout, &stderr)

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String()+stderr.String(), "identity idp")
}
