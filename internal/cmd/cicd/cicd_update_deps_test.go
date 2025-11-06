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
	dep2         = "example.com/dep2"
	dep3         = "example.com/dep3"
	latestDep1   = dep1 + " v1.9.0"
	outdatedDep1 = dep1 + " v1.8.4 [v1.9.0]"
	latestDep2   = dep2 + " v0.15.0"
	outdatedDep2 = dep2 + " v0.14.0 [v0.15.0]"
	latestDep3   = dep3 + " v1.4.0"
)

var (
	goModStat = &mockFileInfo{name: "go.mod", modTime: time.Now().UTC()}
	goSumStat = &mockFileInfo{name: "go.sum", modTime: time.Now().UTC()}
)

func TestCheckDependencyUpdates_EmptyOutput(t *testing.T) {
	validateOutdatedDeps(t, cryptoutilMagic.DepCheckAll, "", []string{}, nil)
}

func TestCheckDependencyUpdates_NoOutdatedDeps(t *testing.T) {
	validateOutdatedDeps(t, cryptoutilMagic.DepCheckAll, latestDep1+"\n"+latestDep2, []string{}, nil)
}

func TestCheckDependencyUpdates_WithOutdatedDeps(t *testing.T) {
	validateOutdatedDeps(t, cryptoutilMagic.DepCheckAll, outdatedDep1+"\n"+outdatedDep2, []string{outdatedDep1, outdatedDep2}, nil)
}

func TestCheckDependencyUpdates_DirectMode(t *testing.T) {
	validateOutdatedDeps(t, cryptoutilMagic.DepCheckAll, outdatedDep1+"\n"+latestDep2+"\n"+latestDep3, []string{outdatedDep1}, map[string]bool{dep1: true})
}

func TestCheckDependencyUpdates_AllMode(t *testing.T) {
	validateOutdatedDeps(t, cryptoutilMagic.DepCheckAll, outdatedDep1+"\n"+outdatedDep2, []string{outdatedDep1, outdatedDep2}, nil)
}

func TestCheckDependencyUpdates_MalformedLines(t *testing.T) {
	validateOutdatedDeps(t, cryptoutilMagic.DepCheckAll, outdatedDep1+"\nmalformed\n"+outdatedDep2, []string{outdatedDep1, outdatedDep2}, nil)
}

func validateOutdatedDeps(t *testing.T, depCheckMode cryptoutilMagic.DepCheckMode, actualDeps string, expectedOutdatedDeps []string, directDeps map[string]bool) {
	t.Helper()

	actualOutdatedDeps, err := checkDependencyUpdates(depCheckMode, goModStat, goSumStat, actualDeps, directDeps)
	require.NoError(t, err)
	require.Len(t, actualOutdatedDeps, len(expectedOutdatedDeps))

	for _, expectedOutdatedDep := range expectedOutdatedDeps {
		require.Contains(t, actualOutdatedDeps, expectedOutdatedDep)
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
