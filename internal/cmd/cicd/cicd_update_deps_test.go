// Package cicd provides tests for dependency update checking functionality.
package cicd

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	mockGoListOutputWithOutdatedDeps = `github.com/stretchr/testify v1.8.4 [v1.9.0]
golang.org/x/crypto v0.14.0 [v0.15.0]
github.com/google/uuid v1.3.0
`
	mockGoListOutputSimpleOutdatedDeps = `github.com/stretchr/testify v1.8.4 [v1.9.0]
golang.org/x/crypto v0.14.0 [v0.15.0]
`
)

func TestCheckDependencyUpdates_NoOutdatedDeps(t *testing.T) {
	// Create mock file info
	goModStat := &mockFileInfo{name: "go.mod", modTime: time.Now()}
	goSumStat := &mockFileInfo{name: "go.sum", modTime: time.Now()}

	// Mock go list output with no outdated dependencies
	goListOutput := `github.com/stretchr/testify v1.8.4
golang.org/x/crypto v0.14.0
`

	outdated, err := checkDependencyUpdates(DepCheckAll, goModStat, goSumStat, goListOutput, nil)
	require.NoError(t, err)
	require.Empty(t, outdated)
}

func TestCheckDependencyUpdates_WithOutdatedDeps(t *testing.T) {
	// Create mock file info
	goModStat := &mockFileInfo{name: "go.mod", modTime: time.Now()}
	goSumStat := &mockFileInfo{name: "go.sum", modTime: time.Now()}

	// Mock go list output with outdated dependencies
	goListOutput := mockGoListOutputWithOutdatedDeps

	outdated, err := checkDependencyUpdates(DepCheckAll, goModStat, goSumStat, goListOutput, nil)
	require.NoError(t, err)
	require.Len(t, outdated, 2)
	require.Contains(t, outdated, "github.com/stretchr/testify v1.8.4 [v1.9.0]")
	require.Contains(t, outdated, "golang.org/x/crypto v0.14.0 [v0.15.0]")
}

func TestCheckDependencyUpdates_DirectMode(t *testing.T) {
	// Create mock file info
	goModStat := &mockFileInfo{name: "go.mod", modTime: time.Now()}
	goSumStat := &mockFileInfo{name: "go.sum", modTime: time.Now()}

	// Mock go list output with both direct and indirect outdated dependencies
	goListOutput := mockGoListOutputWithOutdatedDeps

	// Mock direct dependencies map
	directDeps := map[string]bool{
		"github.com/stretchr/testify": true,
	}

	outdated, err := checkDependencyUpdates(DepCheckDirect, goModStat, goSumStat, goListOutput, directDeps)
	require.NoError(t, err)
	require.Len(t, outdated, 1)
	require.Contains(t, outdated, "github.com/stretchr/testify v1.8.4 [v1.9.0]")
}

func TestCheckDependencyUpdates_AllMode(t *testing.T) {
	// Create mock file info
	goModStat := &mockFileInfo{name: "go.mod", modTime: time.Now()}
	goSumStat := &mockFileInfo{name: "go.sum", modTime: time.Now()}

	// Mock go list output with outdated dependencies
	goListOutput := mockGoListOutputSimpleOutdatedDeps

	outdated, err := checkDependencyUpdates(DepCheckAll, goModStat, goSumStat, goListOutput, nil)
	require.NoError(t, err)
	require.Len(t, outdated, 2)
	require.Contains(t, outdated, "github.com/stretchr/testify v1.8.4 [v1.9.0]")
	require.Contains(t, outdated, "golang.org/x/crypto v0.14.0 [v0.15.0]")
}

func TestCheckDependencyUpdates_EmptyOutput(t *testing.T) {
	// Create mock file info
	goModStat := &mockFileInfo{name: "go.mod", modTime: time.Now()}
	goSumStat := &mockFileInfo{name: "go.sum", modTime: time.Now()}

	// Empty go list output
	goListOutput := ""

	outdated, err := checkDependencyUpdates(DepCheckAll, goModStat, goSumStat, goListOutput, nil)
	require.NoError(t, err)
	require.Empty(t, outdated)
}

func TestCheckDependencyUpdates_MalformedLines(t *testing.T) {
	// Create mock file info
	goModStat := &mockFileInfo{name: "go.mod", modTime: time.Now()}
	goSumStat := &mockFileInfo{name: "go.sum", modTime: time.Now()}

	// Mock go list output with malformed lines (should be ignored)
	goListOutput := `github.com/stretchr/testify v1.8.4 [v1.9.0]
malformed line without brackets
another malformed line
golang.org/x/crypto v0.14.0 [v0.15.0]
`

	outdated, err := checkDependencyUpdates(DepCheckAll, goModStat, goSumStat, goListOutput, nil)
	require.NoError(t, err)
	require.Len(t, outdated, 2)
	require.Contains(t, outdated, "github.com/stretchr/testify v1.8.4 [v1.9.0]")
	require.Contains(t, outdated, "golang.org/x/crypto v0.14.0 [v0.15.0]")
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
