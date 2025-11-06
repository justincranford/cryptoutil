// Package cicd provides tests for dependency update checking functionality.
package cicd

import (
	"os"
	"testing"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

const (
	dep1         = "example.com/dep1"
	dep2         = "github.com/dep2"
	dep3         = "example.com/dep3"
	latestDep1   = dep1 + " v1.9.0"
	outdatedDep1 = dep1 + " v1.8.4 [v1.9.0]"
	latestDep2   = dep2 + " v0.15.0"
	outdatedDep2 = dep2 + " v0.14.0 [v0.15.0]"
	latestDep3   = dep3 + " v1.4.0"
)

func TestCheckDependencyUpdates(t *testing.T) {
	goModStat := &mockFileInfo{name: "go.mod", modTime: time.Now().UTC()}
	goSumStat := &mockFileInfo{name: "go.sum", modTime: time.Now().UTC()}

	tests := []struct {
		name                 string
		depCheckMode         cryptoutilMagic.DepCheckMode
		actualDeps           string
		expectedOutdatedDeps []string
		directDeps           map[string]bool
	}{
		{
			name:                 "EmptyOutput",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           "",
			expectedOutdatedDeps: []string{},
			directDeps:           nil,
		},
		{
			name:                 "NoOutdatedDeps",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           latestDep1 + "\n" + latestDep2,
			expectedOutdatedDeps: []string{},
			directDeps:           nil,
		},
		{
			name:                 "WithOutdatedDeps",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           outdatedDep1 + "\n" + outdatedDep2,
			expectedOutdatedDeps: []string{outdatedDep1, outdatedDep2},
			directDeps:           nil,
		},
		{
			name:                 "DirectMode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           outdatedDep1 + "\n" + latestDep2 + "\n" + latestDep3,
			expectedOutdatedDeps: []string{outdatedDep1},
			directDeps:           map[string]bool{dep1: true},
		},
		{
			name:                 "AllMode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           outdatedDep1 + "\n" + outdatedDep2,
			expectedOutdatedDeps: []string{outdatedDep1, outdatedDep2},
			directDeps:           nil,
		},
		{
			name:                 "MalformedLines",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           outdatedDep1 + "\nmalformed\n" + outdatedDep2,
			expectedOutdatedDeps: []string{outdatedDep1, outdatedDep2},
			directDeps:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualOutdatedDeps, err := checkDependencyUpdates(tt.depCheckMode, goModStat, goSumStat, tt.actualDeps, tt.directDeps)
			require.NoError(t, err)
			require.Len(t, actualOutdatedDeps, len(tt.expectedOutdatedDeps))

			for _, expectedOutdatedDep := range tt.expectedOutdatedDeps {
				require.Contains(t, actualOutdatedDeps, expectedOutdatedDep)
			}
		})
	}
}

// mockFileInfo implements os.FileInfo for testing.
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() any           { return nil }
