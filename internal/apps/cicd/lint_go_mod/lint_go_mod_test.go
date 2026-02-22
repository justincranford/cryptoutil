// Copyright (c) 2025 Justin Cranford

package lint_go_mod

import (
"os"
"testing"

"github.com/stretchr/testify/require"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

func TestLint_NoGoMod(t *testing.T) {
// NOTE: Cannot use t.Parallel() - test changes working directory.
origDir, err := os.Getwd()
require.NoError(t, err)

defer func() { require.NoError(t, os.Chdir(origDir)) }()

tmpDir := t.TempDir()
require.NoError(t, os.Chdir(tmpDir))

logger := cryptoutilCmdCicdCommon.NewLogger("test")

err = Lint(logger)
require.Error(t, err, "Lint should fail when no go.mod is present")
}

func TestLint_UpToDateGoMod(t *testing.T) {
// NOTE: Cannot use t.Parallel() - test changes working directory.
origDir, err := os.Getwd()
require.NoError(t, err)

defer func() { require.NoError(t, os.Chdir(origDir)) }()

tmpDir := t.TempDir()
require.NoError(t, os.Chdir(tmpDir))

// Create minimal go.mod with no external dependencies (always up-to-date).
goModContent := "module testmod\n\ngo 1.21\n"
require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), 0o600))

// Create empty go.sum (required by outdated_deps checker).
require.NoError(t, os.WriteFile("go.sum", []byte(""), 0o600))

logger := cryptoutilCmdCicdCommon.NewLogger("test")

err = Lint(logger)
require.NoError(t, err, "Lint should succeed with up-to-date go.mod")
}
