// Package cicd provides tests for dependency update checking functionality.
package cicd

import (
	"testing"

	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckDependencyUpdates(t *testing.T) {
	for _, tt := range checkDependencyUpdatesTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			actualOutdatedDeps, err := checkDependencyUpdates(tt.depCheckMode, tt.actualDeps, tt.directDeps)
			require.NoError(t, err)
			require.Len(t, actualOutdatedDeps, len(tt.expectedOutdatedDeps))

			for _, expectedOutdatedDep := range tt.expectedOutdatedDeps {
				require.Contains(t, actualOutdatedDeps, expectedOutdatedDep)
			}
		})
	}
}

type checkDependencyUpdatesTestCase struct {
	name                 string
	depCheckMode         cryptoutilMagic.DepCheckMode
	actualDeps           string
	expectedOutdatedDeps []string
	directDeps           map[string]bool
}

func checkDependencyUpdatesTestCases() []checkDependencyUpdatesTestCase {
	dep1 := "example.com/dep1"
	dep2 := "github.com/dep2"
	dep3 := "example.com/dep3"

	latestDep1 := dep1 + " v1.9.0"
	latestDep2 := dep2 + " v0.15.0"
	latestDep3 := dep3 + " v1.4.0"

	outdatedDep1 := dep1 + " v1.8.4 [v1.9.0]"
	outdatedDep2 := dep2 + " v0.14.0 [v0.15.0]"
	outdatedDep3 := dep3 + " v1.3.0 [v1.4.0]"

	tests := []checkDependencyUpdatesTestCase{
		{
			name:                 "Malformed Lines",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           outdatedDep1 + "\nmalformed\n" + outdatedDep2,
			expectedOutdatedDeps: []string{outdatedDep1, outdatedDep2},
			directDeps:           map[string]bool{dep1: true, dep2: true},
		},
		{
			name:                 "0 Deps, 0 Outdated, All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           "",
			expectedOutdatedDeps: []string{},
			directDeps:           map[string]bool{},
		},
		{
			name:                 "1 Direct Deps, 0 Outdated, All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           latestDep1,
			expectedOutdatedDeps: []string{},
			directDeps:           map[string]bool{dep1: true},
		},
		{
			name:                 "1 Direct Deps, 1 Outdated, All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           outdatedDep1,
			expectedOutdatedDeps: []string{outdatedDep1},
			directDeps:           map[string]bool{dep1: true},
		},
		{
			name:                 "2 Direct Deps, 0 Outdated, All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           latestDep1 + "\n" + latestDep2,
			expectedOutdatedDeps: []string{},
			directDeps:           map[string]bool{dep1: true, dep2: true},
		},
		{
			name:                 "2 Direct Deps, 1 Outdated (First), All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           outdatedDep1 + "\n" + latestDep2,
			expectedOutdatedDeps: []string{outdatedDep1},
			directDeps:           map[string]bool{dep1: true, dep2: true},
		},
		{
			name:                 "2 Direct Deps, 1 Outdated (Second), All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           latestDep1 + "\n" + outdatedDep2,
			expectedOutdatedDeps: []string{outdatedDep2},
			directDeps:           map[string]bool{dep1: true, dep2: true},
		},
		{
			name:                 "2 Direct Deps, 2 Outdated, All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           outdatedDep1 + "\n" + outdatedDep2,
			expectedOutdatedDeps: []string{outdatedDep1, outdatedDep2},
			directDeps:           map[string]bool{dep1: true, dep2: true},
		},
		{
			name:                 "3 Direct Deps, 0 Outdated, All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           latestDep1 + "\n" + latestDep2 + "\n" + latestDep3,
			expectedOutdatedDeps: []string{},
			directDeps:           map[string]bool{dep1: true, dep2: true, dep3: true},
		},
		{
			name:                 "3 Direct Deps, 1 Outdated (First), All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           outdatedDep1 + "\n" + latestDep2 + "\n" + latestDep3,
			expectedOutdatedDeps: []string{outdatedDep1},
			directDeps:           map[string]bool{dep1: true, dep2: true, dep3: true},
		},
		{
			name:                 "3 Direct Deps, 1 Outdated (Second), All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           latestDep1 + "\n" + outdatedDep2 + "\n" + latestDep3,
			expectedOutdatedDeps: []string{outdatedDep2},
			directDeps:           map[string]bool{dep1: true, dep2: true, dep3: true},
		},
		{
			name:                 "3 Direct Deps, 1 Outdated (Third), All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           latestDep1 + "\n" + latestDep2 + "\n" + outdatedDep3,
			expectedOutdatedDeps: []string{outdatedDep3},
			directDeps:           map[string]bool{dep1: true, dep2: true, dep3: true},
		},
	}

	return tests
}
