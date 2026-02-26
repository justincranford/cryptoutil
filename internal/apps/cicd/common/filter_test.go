// Copyright (c) 2025 Justin Cranford

package common

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

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
			commandName: "lint-text",
			expected:    []string{},
		},
		{
			name:        "filters self-exclusion for lint-text",
			files:       []string{"internal/apps/cicd/lint_text/utf8.go", "internal/common/magic/magic.go"},
			commandName: "lint-text",
			expected:    []string{"internal/common/magic/magic.go"},
		},
		{
			name:        "filters self-exclusion for format-go",
			files:       []string{"internal/apps/cicd/format_go/enforce_any.go", "internal/common/magic/magic.go"},
			commandName: "format-go",
			expected:    []string{"internal/common/magic/magic.go"},
		},
		{
			name:        "filters generated files",
			files:       []string{"api/server/generated_gen.go", "internal/common/magic/magic.go", "proto/service.pb.go"},
			commandName: "lint-text",
			expected:    []string{"internal/common/magic/magic.go"},
		},
		{
			name:        "no exclusions for unknown command",
			files:       []string{"internal/apps/cicd/lint_text/utf8.go", "internal/common/magic/magic.go"},
			commandName: "unknown-command",
			expected:    []string{"internal/apps/cicd/lint_text/utf8.go", "internal/common/magic/magic.go"},
		},
		{
			name:        "keeps non-matching files",
			files:       []string{"internal/server/server.go", "cmd/app/main.go"},
			commandName: "lint-go",
			expected:    []string{"internal/server/server.go", "cmd/app/main.go"},
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
			name:     "empty map",
			fileMap:  map[string][]string{},
			expected: 0,
		},
		{
			name: "single extension",
			fileMap: map[string][]string{
				"go": {"file1.go", "file2.go"},
			},
			expected: 2,
		},
		{
			name: "multiple extensions",
			fileMap: map[string][]string{
				"go":   {"file1.go", "file2.go"},
				"yaml": {"config.yaml"},
				"md":   {"README.md", "CHANGELOG.md"},
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
			name:        "empty map",
			fileMap:     map[string][]string{},
			commandName: "format-go",
			expected:    nil,
		},
		{
			name: "extracts go files",
			fileMap: map[string][]string{
				"go":   {"file1.go", "file2.go"},
				"yaml": {"config.yaml"},
			},
			commandName: "lint-go",
			expected:    []string{"file1.go", "file2.go"},
		},
		{
			name: "filters self-exclusion",
			fileMap: map[string][]string{
				"go": {"internal/apps/cicd/format_go/enforce_any.go", "internal/common/magic/magic.go"},
			},
			commandName: "format-go",
			expected:    []string{"internal/common/magic/magic.go"},
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
