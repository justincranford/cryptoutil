// Package cicd provides tests for circular dependency checking functionality.
package cicd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckCircularDependencies_NoPackages(t *testing.T) {
	// Test with empty JSON output
	err := checkCircularDependencies("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no packages found")
}

func TestCheckCircularDependencies_InvalidJSON(t *testing.T) {
	// Test with invalid JSON
	err := checkCircularDependencies(`{"invalid": json}`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse package info")
}

func TestCheckCircularDependencies_NoCircularDeps(t *testing.T) {
	// Test with valid packages but no circular dependencies
	jsonOutput := `{
		"ImportPath": "cryptoutil/internal/common/util",
		"Imports": ["fmt", "strings"]
	}{
		"ImportPath": "cryptoutil/internal/common/config",
		"Imports": ["cryptoutil/internal/common/util", "os"]
	}`

	err := checkCircularDependencies(jsonOutput)
	require.NoError(t, err)
}

func TestCheckCircularDependencies_WithCircularDeps(t *testing.T) {
	// Test with circular dependencies
	jsonOutput := `{
		"ImportPath": "cryptoutil/internal/common/util",
		"Imports": ["cryptoutil/internal/common/config"]
	}{
		"ImportPath": "cryptoutil/internal/common/config",
		"Imports": ["cryptoutil/internal/common/util"]
	}`

	err := checkCircularDependencies(jsonOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependencies detected")
	require.Contains(t, err.Error(), "Chain 1")
	require.Contains(t, err.Error(), "cryptoutil/internal/common/util")
	require.Contains(t, err.Error(), "cryptoutil/internal/common/config")
}

func TestCheckCircularDependencies_ComplexCircularDeps(t *testing.T) {
	// Test with more complex circular dependencies (A -> B -> C -> A)
	jsonOutput := `{
		"ImportPath": "cryptoutil/internal/a",
		"Imports": ["cryptoutil/internal/b"]
	}{
		"ImportPath": "cryptoutil/internal/b",
		"Imports": ["cryptoutil/internal/c"]
	}{
		"ImportPath": "cryptoutil/internal/c",
		"Imports": ["cryptoutil/internal/a"]
	}`

	err := checkCircularDependencies(jsonOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependencies detected")
	require.Contains(t, err.Error(), "Chain 1")
	require.Contains(t, err.Error(), "cryptoutil/internal/a")
	require.Contains(t, err.Error(), "cryptoutil/internal/b")
	require.Contains(t, err.Error(), "cryptoutil/internal/c")
}

func TestCheckCircularDependencies_IgnoresExternalDeps(t *testing.T) {
	// Test that external dependencies (not starting with cryptoutil/) are ignored
	jsonOutput := `{
		"ImportPath": "cryptoutil/internal/common/util",
		"Imports": ["fmt", "strings", "github.com/stretchr/testify/require"]
	}{
		"ImportPath": "github.com/stretchr/testify/require",
		"Imports": ["cryptoutil/internal/common/util"]
	}`

	// Should not detect circular dependency because external package importing internal is ignored
	err := checkCircularDependencies(jsonOutput)
	require.NoError(t, err)
}

func TestCheckCircularDependencies_MultipleChains(t *testing.T) {
	// Test with multiple separate circular dependency chains
	jsonOutput := `{
		"ImportPath": "cryptoutil/internal/a",
		"Imports": ["cryptoutil/internal/b"]
	}{
		"ImportPath": "cryptoutil/internal/b",
		"Imports": ["cryptoutil/internal/a"]
	}{
		"ImportPath": "cryptoutil/internal/x",
		"Imports": ["cryptoutil/internal/y"]
	}{
		"ImportPath": "cryptoutil/internal/y",
		"Imports": ["cryptoutil/internal/x"]
	}`

	err := checkCircularDependencies(jsonOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependencies detected")
	require.Contains(t, err.Error(), "2 circular dependency chain(s)")
}
