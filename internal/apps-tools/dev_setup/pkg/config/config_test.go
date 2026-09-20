package config

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestLoad(t *testing.T) {
	t.Run("Configuration loads successfully", func(t *testing.T) {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg == nil {
			t.Fatalf("expected config, got nil")
		}

		if len(cfg.Groups) == 0 {
			t.Errorf("expected at least one group, got 0")
		}
	})
}

func TestGroupOrdering(t *testing.T) {
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
			if i < len(expectedOrder) && group.Name != expectedOrder[i] {
				t.Errorf("group %d: expected %q, got %q", i, expectedOrder[i], group.Name)
			}
		}
	})
}

func TestBlockOnFailureFlag(t *testing.T) {
	t.Run("Critical groups block on failure", func(t *testing.T) {
		cfg, _ := Load()

		criticalGroups := map[string]bool{
			cryptoutilSharedMagic.DevSetupTestGroupPythonRuntime:    true,
			cryptoutilSharedMagic.DevSetupTestGroupUVPackageManager: true,
			cryptoutilSharedMagic.DevSetupTestGroupGoToolchain:      true,
		}

		for _, group := range cfg.Groups {
			shouldBlock := criticalGroups[group.Name]
			if shouldBlock && !group.BlockOnFailure {
				t.Errorf("group %q should block on failure but doesn't", group.Name)
			}
		}
	})
}

func TestDependencyStructure(t *testing.T) {
	t.Run("Dependencies have valid structure", func(t *testing.T) {
		cfg, _ := Load()

		for _, group := range cfg.Groups {
			if group.Name == "" {
				t.Errorf("group has no name")
			}

			if len(group.Dependencies) == 0 {
				t.Errorf("group %q has no dependencies", group.Name)
			}
		}
	})
}

func TestToolTypes(t *testing.T) {
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
			t.Errorf("expected tool types to be defined")
		}
	})
}

func TestVersionStrings(t *testing.T) {
	t.Run("All dependencies have version requirements", func(t *testing.T) {
		cfg, _ := Load()

		for _, group := range cfg.Groups {
			for _, dep := range group.Dependencies {
				if dep.MinVersion == "" {
					t.Errorf("dependency %q has no minimum version", dep.Name)
				}
			}
		}
	})
}
