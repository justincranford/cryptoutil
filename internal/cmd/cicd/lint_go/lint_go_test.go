// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"testing"

	"github.com/stretchr/testify/require"
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
