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

// Sequential: modifies package-level configsDeploymentsStatFn seam.
func TestFindViolationsInDir_StatError(t *testing.T) {
	orig := configsDeploymentsStatFn

	defer func() { configsDeploymentsStatFn = orig }()

	configsDeploymentsStatFn = func(_ string) (os.FileInfo, error) {
		return nil, errors.New("injected stat error")
	}

	violations, err := FindViolationsInDir(".")
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "deployments/ directory not found")
}

// Sequential: modifies package-level configsDeploymentsReadDirFn seam.
func TestFindViolationsInDir_ReadDirError(t *testing.T) {
	orig := configsDeploymentsReadDirFn
	origStat := configsDeploymentsStatFn

	defer func() {
		configsDeploymentsReadDirFn = orig
		configsDeploymentsStatFn = origStat
	}()

	configsDeploymentsStatFn = func(_ string) (os.FileInfo, error) {
		return nil, nil
	}
	configsDeploymentsReadDirFn = func(_ string) ([]fs.DirEntry, error) {
		return nil, errors.New("injected readdir error")
	}

	violations, err := FindViolationsInDir(".")
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "failed to read deployments/ directory")
}

// Sequential: modifies package-level configsDeploymentsStatFn seam.
func TestFindViolationsInDir_ConfigsStatError(t *testing.T) {
	origStat := configsDeploymentsStatFn

	defer func() {
		configsDeploymentsStatFn = origStat
	}()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(root+"/deployments/sm-kms", cryptoutilSharedMagic.CICDTempDirPermissions))

	callCount := 0
	configsDeploymentsStatFn = func(path string) (os.FileInfo, error) {
		callCount++
		if callCount == 1 {
			return os.Stat(path) // Let deployments/ stat succeed
		}

		return nil, errors.New("injected configs stat error")
	}

	violations, err := FindViolationsInDir(root)
	require.NoError(t, err)
	require.NotEmpty(t, violations)
}
