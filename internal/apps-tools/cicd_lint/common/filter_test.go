// Copyright (c) 2025-2026 Justin Cranford.
package common

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestFilterFilesForCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		files       []string
		commandName string
		expected    []string
	}{
		{
			name:        "empty files",
			files:       []string{},
			commandName: testLintText,
			expected:    []string{},
		},
		{
			name:        "filters self-exclusion for lint-text",
			files:       []string{testInternalAppsToolsCicdLintLintTextUtf8Go, testInternalCommonMagicMagicGo},
			commandName: testLintText,
			expected:    []string{testInternalCommonMagicMagicGo},
		},
		{
			name:        "filters self-exclusion for format-go",
			files:       []string{testInternalAppsToolsCicdLintFormatGoEnforceAnyGo, testInternalCommonMagicMagicGo},
			commandName: valueFormatGo,
			expected:    []string{testInternalCommonMagicMagicGo},
		},
		{
			name:        "filters template go files for format-go",
			files:       []string{"api/cryptosuite-registry/templates/internal/apps/__PS_ID__/__SERVICE__.go", testInternalCommonMagicMagicGo},
			commandName: valueFormatGo,
			expected:    []string{testInternalCommonMagicMagicGo},
		},
		{
			name:        "filters generated files",
			files:       []string{"api/server/generated_gen.go", testInternalCommonMagicMagicGo, "proto/service.pb.go"},
			commandName: testLintText,
			expected:    []string{testInternalCommonMagicMagicGo},
		},
		{
			name:        "no exclusions for unknown command",
			files:       []string{testInternalAppsToolsCicdLintLintTextUtf8Go, testInternalCommonMagicMagicGo},
			commandName: "unknown-command",
			expected:    []string{testInternalAppsToolsCicdLintLintTextUtf8Go, testInternalCommonMagicMagicGo},
		},
		{
			name:        "keeps non-matching files",
			files:       []string{testInternalServerServerGo, testCmdAppMainGo},
			commandName: testLintGo,
			expected:    []string{testInternalServerServerGo, testCmdAppMainGo},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := FilterFilesForCommand(tc.files, tc.commandName)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestFlattenFileMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fileMap  map[string][]string
		expected int
	}{
		{
			name:     testEmptyMap,
			fileMap:  map[string][]string{},
			expected: 0,
		},
		{
			name: "single extension",
			fileMap: map[string][]string{
				valueGo: {testFile1Go, testFile2Go},
			},
			expected: 2,
		},
		{
			name: "multiple extensions",
			fileMap: map[string][]string{
				valueGo:  {testFile1Go, testFile2Go},
				testYAML: {testConfigYAML},
				"md":     {"README.md", "CHANGELOG.md"},
			},
			expected: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := FlattenFileMap(tc.fileMap)
			require.Len(t, result, tc.expected)
		})
	}
}

func TestGetGoFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileMap     map[string][]string
		commandName string
		expected    []string
	}{
		{
			name:        testEmptyMap,
			fileMap:     map[string][]string{},
			commandName: valueFormatGo,
			expected:    nil,
		},
		{
			name: "extracts go files",
			fileMap: map[string][]string{
				valueGo:  {testFile1Go, testFile2Go},
				testYAML: {testConfigYAML},
			},
			commandName: testLintGo,
			expected:    []string{testFile1Go, testFile2Go},
		},
		{
			name: "filters self-exclusion",
			fileMap: map[string][]string{
				valueGo: {testInternalAppsToolsCicdLintFormatGoEnforceAnyGo, testInternalCommonMagicMagicGo},
			},
			commandName: valueFormatGo,
			expected:    []string{testInternalCommonMagicMagicGo},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := GetGoFiles(tc.fileMap, tc.commandName)
			require.Equal(t, tc.expected, result)
		})
	}
}
