// Copyright (c) 2025 Justin Cranford

package legacy_ports

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestCheck_EmptyLegacyPorts(t *testing.T) {
	// Cannot be parallel: modifies package-level injectable var.
	originalFn := legacyPortsAllFn

	defer func() { legacyPortsAllFn = originalFn }()

	legacyPortsAllFn = func() []uint16 {
		return nil
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, map[string][]string{})
	require.NoError(t, err)
}

func TestCheckFile_ShortRegexMatch(t *testing.T) {
	// Cannot be parallel: modifies package-level injectable var.
	originalFn := legacyPortsFindAllFn

	defer func() { legacyPortsFindAllFn = originalFn }()

	// Return matches with only 1 element (no capture group) to trigger the len(match) < 2 guard.
	legacyPortsFindAllFn = func(_ string, _ int) [][]string {
		return [][]string{{"8080"}} // Missing capture group element.
	}

	// Create a test file with content that would normally match.
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("port: 8080\n"), 0o600))

	violations := CheckFile(testFile, []uint16{8080})
	require.Empty(t, violations)
}
