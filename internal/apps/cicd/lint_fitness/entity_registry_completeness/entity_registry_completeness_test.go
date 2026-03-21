// Copyright (c) 2025 Justin Cranford

package entity_registry_completeness_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintFitnessRegistry "cryptoutil/internal/apps/cicd/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	lintFitnessEntityRegistryCompleteness "cryptoutil/internal/apps/cicd/lint_fitness/entity_registry_completeness"
)

func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("skipping integration test: cannot find project root (no go.mod)")
		}

		dir = parent
	}
}

func setupAllComponents(t *testing.T, tmpDir string) {
	t.Helper()

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		deploymentsDir := filepath.Join(tmpDir, "deployments", ps.PSID)
		require.NoError(t, os.MkdirAll(deploymentsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		configsDir := filepath.Join(tmpDir, "configs", ps.Product, ps.Service)
		require.NoError(t, os.MkdirAll(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		magicDir := filepath.Join(tmpDir, "internal", "shared", "magic")
		require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		magicFile := filepath.Join(magicDir, ps.MagicFile)
		if _, err := os.Stat(magicFile); os.IsNotExist(err) {
			require.NoError(t, os.WriteFile(magicFile, []byte("package magic\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := lintFitnessEntityRegistryCompleteness.CheckInDir(logger, root)
	require.NoError(t, err)
}

func TestCheckInDir_AllComponentsPresent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllComponents(t, tmpDir)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := lintFitnessEntityRegistryCompleteness.CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingDeploymentDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		psID string
	}{
		{name: cryptoutilSharedMagic.OTLPServiceSMIM, psID: cryptoutilSharedMagic.OTLPServiceSMIM},
		{name: cryptoutilSharedMagic.OTLPServiceSMKMS, psID: cryptoutilSharedMagic.OTLPServiceSMKMS},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			setupAllComponents(t, tmpDir)

			deploymentsDir := filepath.Join(tmpDir, "deployments", tc.psID)
			require.NoError(t, os.RemoveAll(deploymentsDir))

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err := lintFitnessEntityRegistryCompleteness.CheckInDir(logger, tmpDir)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.psID)
			require.Contains(t, err.Error(), "missing deployments/")
		})
	}
}

func TestCheckInDir_MissingConfigsDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		product string
		service string
	}{
		{name: cryptoutilSharedMagic.OTLPServiceSMIM, product: cryptoutilSharedMagic.SMProductName, service: cryptoutilSharedMagic.IMServiceName},
		{name: cryptoutilSharedMagic.OTLPServiceJoseJA, product: cryptoutilSharedMagic.JoseProductName, service: cryptoutilSharedMagic.JoseJAServiceName},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			setupAllComponents(t, tmpDir)

			configsDir := filepath.Join(tmpDir, "configs", tc.product, tc.service)
			require.NoError(t, os.RemoveAll(configsDir))

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err := lintFitnessEntityRegistryCompleteness.CheckInDir(logger, tmpDir)
			require.Error(t, err)
			require.Contains(t, err.Error(), "missing configs/")
		})
	}
}

func TestCheckInDir_MissingMagicFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		magicFile string
	}{
		{name: cryptoutilSharedMagic.SMProductName, magicFile: "magic_sm.go"},
		{name: cryptoutilSharedMagic.PKIProductName, magicFile: "magic_pki.go"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			setupAllComponents(t, tmpDir)

			magicFilePath := filepath.Join(tmpDir, "internal", "shared", "magic", tc.magicFile)
			require.NoError(t, os.Remove(magicFilePath))

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err := lintFitnessEntityRegistryCompleteness.CheckInDir(logger, tmpDir)
			require.Error(t, err)
			require.Contains(t, err.Error(), "missing internal/shared/magic/")
		})
	}
}
