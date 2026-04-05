// Copyright (c) 2025 Justin Cranford

package legacy_ports

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func TestCheck_EmptyLegacyPorts(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckWithAllFn(logger, map[string][]string{}, func() []uint16 {
		return nil
	})
	require.NoError(t, err)
}

func TestCheckFile_ShortRegexMatch(t *testing.T) {
	t.Parallel()

	// Return matches with only 1 element (no capture group) to trigger the len(match) < 2 guard.
	stubFindAllFn := func(_ string, _ int) [][]string {
		return [][]string{{"8080"}} // Missing capture group element.
	}

	// Create a test file with content that would normally match.
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("port: 8080\n"), cryptoutilSharedMagic.CacheFilePermissions))

	violations := CheckFile(testFile, []uint16{cryptoutilSharedMagic.TestServerPort}, stubFindAllFn)
	require.Empty(t, violations)
}
