// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// TestCheckCryptoRandInDir_WalkError verifies that checkCryptoRandInDir
// returns error when findMathRandViolationsInDir returns a walk error.
func TestCheckCryptoRandInDir_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create an inaccessible subdirectory to trigger walk error.
	badDir := filepath.Join(tmpDir, "baddir")
	err := os.MkdirAll(badDir, 0o000)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chmod(badDir, 0o700) })

	err = checkCryptoRandInDir(logger, tmpDir)
	require.Error(t, err, "Should return error when walk fails")
	require.Contains(t, err.Error(), "failed to check math/rand usage")
}

// TestCheckInsecureSkipVerifyInDir_WalkError verifies that checkInsecureSkipVerifyInDir
// returns error when findInsecureSkipVerifyViolationsInDir returns a walk error.
func TestCheckInsecureSkipVerifyInDir_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create an inaccessible subdirectory to trigger walk error.
	badDir := filepath.Join(tmpDir, "baddir")
	err := os.MkdirAll(badDir, 0o000)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chmod(badDir, 0o700) })

	err = checkInsecureSkipVerifyInDir(logger, tmpDir)
	require.Error(t, err, "Should return error when walk fails")
	require.Contains(t, err.Error(), "failed to check InsecureSkipVerify")
}
