// Copyright (c) 2025 Justin Cranford

package thelper

import (
	"errors"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

// TestFix_PrinterFprintError covers the printer.Fprint error path in fixTHelperInFileWithPrinter.
func TestFix_PrinterFprintError(t *testing.T) {
	t.Parallel()

	stubFprintFn := func(_ io.Writer, _ *token.FileSet, _ any) error {
		return errors.New("injected fprint error")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "setup_test.go"),
		[]byte(testContentSetupMissingHelper), cryptoutilSharedMagic.CacheFilePermissions))

	_, _, err := fixTHelperInFileWithPrinter(logger, filepath.Join(tmpDir, "setup_test.go"), stubFprintFn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write file")
}
