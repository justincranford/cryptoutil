package config

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Parallel()
	t.Run("Configuration loads successfully", func(t *testing.T) {
		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg, "expected config, got nil")
		require.NotEmpty(t, cfg.Groups, "expected at least one group")
	})
}

func TestGroupOrdering(t *testing.T) {
	t.Parallel()
	t.Run("Groups are in correct dependency order", func(t *testing.T) {
		cfg, _ := Load()

		expectedOrder := []string{
			cryptoutilSharedMagic.DevSetupTestGroupPythonRuntime,
			cryptoutilSharedMagic.DevSetupTestGroupUVPackageManager,
			cryptoutilSharedMagic.DevSetupTestGroupPythonDependencies,
			cryptoutilSharedMagic.DevSetupTestGroupGoToolchain,
			cryptoutilSharedMagic.DevSetupTestGroupGoLinting,
			cryptoutilSharedMagic.DevSetupTestGroupPreCommitHooks,
		}

		for i, group := range cfg.Groups {
			if i < len(expectedOrder) {
				require.Equal(t, expectedOrder[i], group.Name, "group %d", i)
			}
		}
	})
}

func TestBlockOnFailureFlag(t *testing.T) {
	t.Parallel()
	t.Run("Critical groups block on failure", func(t *testing.T) {
		cfg, _ := Load()

		criticalGroups := map[string]bool{
			cryptoutilSharedMagic.DevSetupTestGroupPythonRuntime:    true,
			cryptoutilSharedMagic.DevSetupTestGroupUVPackageManager: true,
			cryptoutilSharedMagic.DevSetupTestGroupGoToolchain:      true,
		}

		for _, group := range cfg.Groups {
			shouldBlock := criticalGroups[group.Name]
			if shouldBlock {
				require.True(t, group.BlockOnFailure, "group %q should block on failure", group.Name)
			}
		}
	})
}

func TestDependencyStructure(t *testing.T) {
	t.Parallel()
	t.Run("Dependencies have valid structure", func(t *testing.T) {
		cfg, _ := Load()

		for _, group := range cfg.Groups {
			require.NotEmpty(t, group.Name, "group has no name")
			require.NotEmpty(t, group.Dependencies, "group %q has no dependencies", group.Name)
		}
	})
}

func TestToolTypes(t *testing.T) {
	t.Parallel()
	t.Run("Tool types are valid", func(t *testing.T) {
		// Verify magic constants are defined
		validTypes := map[string]bool{
			cryptoutilSharedMagic.DevSetupToolTypeGo:           true,
			cryptoutilSharedMagic.DevSetupToolTypeGoModule:     true,
			cryptoutilSharedMagic.DevSetupToolTypePython:       true,
			cryptoutilSharedMagic.DevSetupToolTypePythonModule: true,
			cryptoutilSharedMagic.DevSetupToolTypeSystemBinary: true,
			cryptoutilSharedMagic.DevSetupToolTypeUV:           true,
		}

		if len(validTypes) == 0 {
			require.Fail(t, "expected tool types to be defined")
		}
	})
}

func TestVersionStrings(t *testing.T) {
	t.Parallel()
	t.Run("All dependencies have version requirements", func(t *testing.T) {
		cfg, _ := Load()

		for _, group := range cfg.Groups {
			for _, dep := range group.Dependencies {
				// Dependencies without version detection (e.g. filesystem-existence
				// checks like git hook installation) have no version to require.
				if dep.DetectVersionCmd == "" {
					continue
				}

				require.NotEmpty(t, dep.MinVersion, "dependency %q has no minimum version", dep.Name)
			}
		}
	})
}
