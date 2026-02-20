// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoCryptoRand "cryptoutil/internal/apps/cicd/lint_go/crypto_rand"
	lintGoInsecureSkipVerify "cryptoutil/internal/apps/cicd/lint_go/insecure_skip_verify"
)

// TestCheckCryptoRandInDir_WalkError verifies that lintGoCryptoRand.CheckInDir
// returns error when lintGoCryptoRand.FindMathRandViolationsInDir returns a walk error.
func TestCheckCryptoRandInDir_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create an inaccessible subdirectory to trigger walk error.
	badDir := filepath.Join(tmpDir, "baddir")
	err := os.MkdirAll(badDir, 0o000)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chmod(badDir, 0o700) })

	err = lintGoCryptoRand.CheckInDir(logger, tmpDir)
	require.Error(t, err, "Should return error when walk fails")
	require.Contains(t, err.Error(), "failed to check math/rand usage")
}

// TestCheckInsecureSkipVerifyInDir_WalkError verifies that lintGoInsecureSkipVerify.CheckInDir
// returns error when lintGoInsecureSkipVerify.FindInsecureSkipVerifyViolationsInDir returns a walk error.
func TestCheckInsecureSkipVerifyInDir_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create an inaccessible subdirectory to trigger walk error.
	badDir := filepath.Join(tmpDir, "baddir")
	err := os.MkdirAll(badDir, 0o000)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chmod(badDir, 0o700) })

	err = lintGoInsecureSkipVerify.CheckInDir(logger, tmpDir)
	require.Error(t, err, "Should return error when walk fails")
	require.Contains(t, err.Error(), "failed to check InsecureSkipVerify")
}
