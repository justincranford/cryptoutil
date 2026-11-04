// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateDirectoryChainPreservesAbsolutePathError(t *testing.T) {
	t.Parallel()

	pathError := errors.New("injected absolute path resolution failure")
	err := validateDirectoryChainWithAbs(
		operatingSystemFileSystem{},
		"/unused",
		func(string) (string, error) {
			return "", pathError
		},
	)
	require.ErrorContains(t, err, "resolve materializer output parent")
	require.ErrorIs(t, err, pathError)
}
