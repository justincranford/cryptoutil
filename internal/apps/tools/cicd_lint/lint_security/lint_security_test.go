// Copyright (c) 2025 Justin Cranford

package lint_security

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestLint_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger, map[string][]string{})

	require.NoError(t, err, "Lint should succeed with no files")
}

func TestLint_CleanGoFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	goFile := filepath.Join(tmpDir, "clean.go")
	content := `package main

import (
	"crypto/rand"
	"fmt"
)

func main() {
	fmt.Println(rand.Reader)
}
`

	err := os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Lint(logger, map[string][]string{"go": {goFile}})

	require.NoError(t, err, "Lint should succeed with clean files")
}

func TestLint_BannedImportDetected(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	goFile := filepath.Join(tmpDir, "bad.go")
	content := `package main

import "math/rand"

func main() {
	_ = rand.Intn(10)
}
`

	err := os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Lint(logger, map[string][]string{"go": {goFile}})

	require.Error(t, err, "Lint should fail when banned imports found")
	require.Contains(t, err.Error(), "lint-security failed")
}
