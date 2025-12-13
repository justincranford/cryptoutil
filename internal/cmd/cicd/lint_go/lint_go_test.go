// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

func TestCheckDependencies_NoCycles(t *testing.T) {
	t.Parallel()

	// Simulate go list -json output with no cycles.
	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/c", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err, "Should not detect cycles in acyclic graph")
}

func TestCheckDependencies_WithCycle(t *testing.T) {
	t.Parallel()

	// Simulate go list -json output with a cycle: a -> b -> c -> a.
	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/a"]}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err, "Should detect cycle")
	require.Contains(t, err.Error(), "circular dependency", "Error should mention circular dependency")
}

func TestCheckDependencies_EmptyOutput(t *testing.T) {
	t.Parallel()

	err := CheckDependencies("")
	require.NoError(t, err, "Empty output should not cause error")
}

func TestCheckDependencies_SinglePackage(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err, "Single package with no imports should not cause error")
}

func TestCheckDependencies_ExternalDepsIgnored(t *testing.T) {
	t.Parallel()

	// External dependencies should be ignored.
	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["github.com/external/pkg", "example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["fmt", "github.com/another/pkg"]}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err, "External dependencies should be ignored")
}

func TestGetModulePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		packages map[string][]string
		expected string
	}{
		{
			name:     "empty packages",
			packages: map[string][]string{},
			expected: "",
		},
		{
			name: "single package",
			packages: map[string][]string{
				"example.com/pkg/a": {},
			},
			expected: "example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := getModulePath(tc.packages)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestLint(t *testing.T) {
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger)

	require.Error(t, err, "Lint fails when go.mod not in current directory")
	require.Contains(t, err.Error(), "lint-go failed")
}

func TestCheckDependencies_MalformedJSON(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
invalid json line
{"ImportPath": "example.com/pkg/b", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode package info")
}

func TestCheckDependencies_ComplexCycle(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b", "example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/d", "Imports": ["example.com/pkg/a"]}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependency")
}

func TestCheckDependencies_SelfReference(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/a"]}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependency")
}

func TestCheckDependencies_MultipleDisconnectedGraphs(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": []}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/d", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err)
}

func TestGetModulePath_MultiplePackages(t *testing.T) {
	t.Parallel()

	packages := map[string][]string{
		"example.com/pkg/a": {},
		"example.com/pkg/b": {},
		"example.com/pkg/c": {},
	}

	result := getModulePath(packages)
	require.Equal(t, "example.com", result)
}

func TestGetModulePath_DifferentPrefixes(t *testing.T) {
	t.Parallel()

	packages := map[string][]string{
		"github.com/user/repo/pkg/a": {},
		"github.com/user/repo/pkg/b": {},
	}

	result := getModulePath(packages)
	require.Equal(t, "github.com", result)
}
