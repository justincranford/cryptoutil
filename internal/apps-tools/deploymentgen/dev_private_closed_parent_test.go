// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentPrivateClosedRootRejectsNestedParentCreation(t *testing.T) {
	t.Parallel()
	directory, err := CreateDevelopmentPrivateDirectory(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, directory.Close())
	require.ErrorContains(t, directory.Write("closed-child/public-input", []byte("public test payload")), "parent creation")
}
