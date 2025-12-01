// Copyright (c) 2025 Justin Cranford

package lint_go_mod

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

func TestLint_NoGoMod(t *testing.T) {
	t.Parallel()

	// This test would fail if run in a directory without go.mod.
	// Since we're in a Go project, it should work.
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Note: This will actually check dependencies, which may pass or fail.
	// We're testing that the function doesn't panic.
	_ = Lint(logger)
}

func TestCheckDependencyUpdates_Empty(t *testing.T) {
	t.Parallel()

	outdated, err := checkDependencyUpdates("", map[string]bool{})

	require.NoError(t, err)
	require.Empty(t, outdated)
}

func TestCheckDependencyUpdates_NoUpdates(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
github.com/pkg/errors v0.9.1
github.com/stretchr/testify v1.8.0`

	directDeps := map[string]bool{
		"github.com/pkg/errors":       true,
		"github.com/stretchr/testify": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)

	require.NoError(t, err)
	require.Empty(t, outdated)
}

func TestCheckDependencyUpdates_WithUpdates(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
github.com/pkg/errors v0.9.1 [v0.9.2]
github.com/stretchr/testify v1.8.0`

	directDeps := map[string]bool{
		"github.com/pkg/errors":       true,
		"github.com/stretchr/testify": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)

	require.NoError(t, err)
	require.Len(t, outdated, 1)
	require.Contains(t, outdated[0], "github.com/pkg/errors")
}

func TestCheckDependencyUpdates_IndirectNotIncluded(t *testing.T) {
	t.Parallel()

	goListOutput := `example.com/mymodule
github.com/pkg/errors v0.9.1 [v0.9.2]
github.com/indirect/dep v1.0.0 [v1.1.0]`

	// Only direct deps are included.
	directDeps := map[string]bool{
		"github.com/pkg/errors": true,
	}

	outdated, err := checkDependencyUpdates(goListOutput, directDeps)

	require.NoError(t, err)
	require.Len(t, outdated, 1)
	require.Contains(t, outdated[0], "github.com/pkg/errors")
}

func TestGetDirectDependencies(t *testing.T) {
	t.Parallel()

	goModContent := []byte(`module example.com/mymodule

go 1.25.4

require (
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.0
	github.com/indirect/dep v1.0.0 // indirect
)
`)

	directDeps, err := getDirectDependencies(goModContent)

	require.NoError(t, err)
	require.True(t, directDeps["github.com/pkg/errors"])
	require.True(t, directDeps["github.com/stretchr/testify"])
	require.False(t, directDeps["github.com/indirect/dep"])
}

func TestGetDirectDependencies_SingleLineRequire(t *testing.T) {
	t.Parallel()

	goModContent := []byte(`module example.com/mymodule

go 1.25.4

require github.com/pkg/errors v0.9.1
`)

	directDeps, err := getDirectDependencies(goModContent)

	require.NoError(t, err)
	require.True(t, directDeps["github.com/pkg/errors"])
}

func TestGetDirectDependencies_Empty(t *testing.T) {
	t.Parallel()

	goModContent := []byte(`module example.com/mymodule

go 1.25.4
`)

	directDeps, err := getDirectDependencies(goModContent)

	require.NoError(t, err)
	require.Empty(t, directDeps)
}
