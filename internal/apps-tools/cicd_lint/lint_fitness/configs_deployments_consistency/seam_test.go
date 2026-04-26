// Copyright (c) 2025 Justin Cranford

package configs_deployments_consistency

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestFindViolationsInDir_StatError(t *testing.T) {
	t.Parallel()

	violations, err := FindViolationsInDir(".", func(_ string) (os.FileInfo, error) {
		return nil, errors.New("injected stat error")
	}, os.ReadDir)
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "deployments/ directory not found")
}

func TestFindViolationsInDir_ReadDirError(t *testing.T) {
	t.Parallel()

	violations, err := FindViolationsInDir(".", func(_ string) (os.FileInfo, error) {
		return nil, nil
	}, func(_ string) ([]fs.DirEntry, error) {
		return nil, errors.New("injected readdir error")
	})
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "failed to read deployments/ directory")
}

func TestFindViolationsInDir_ConfigsStatError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(root+"/deployments/sm-kms", cryptoutilSharedMagic.CICDTempDirPermissions))

	callCount := 0
	violations, err := FindViolationsInDir(root, func(path string) (os.FileInfo, error) {
		callCount++
		if callCount == 1 {
			return os.Stat(path) // Let deployments/ stat succeed.
		}

		return nil, errors.New("injected configs stat error")
	}, os.ReadDir)
	require.NoError(t, err)
	require.NotEmpty(t, violations)
}
