// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentNativeProtectionRejectsAbsentRoot(t *testing.T) {
	t.Parallel()

	require.Error(t, secureDevelopmentDirectory(filepath.Join(t.TempDir(), "absent-owned-directory")))
}
