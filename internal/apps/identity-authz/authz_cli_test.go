// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"bytes"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestAuthz_MainHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Authz([]string{cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String()+stderr.String(), "identity authz")
}
