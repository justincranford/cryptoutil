// Copyright (c) 2025 Justin Cranford

package thelper

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

// TestFix_PrinterFprintError covers the printer.Fprint error path in fixTHelperInFile.
// NOT parallel — modifies package-level injectable var.
func TestFix_PrinterFprintError(t *testing.T) {
	original := printerFprintFn
	printerFprintFn = func(_ io.Writer, _ *token.FileSet, _ any) error {
		return errors.New("injected fprint error")
	}

	defer func() { printerFprintFn = original }()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "setup_test.go"),
		[]byte(testContentSetupMissingHelper), cryptoutilSharedMagic.CacheFilePermissions))

	_, _, _, err := Fix(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to process")
}
