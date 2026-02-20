// Copyright (c) 2025 Justin Cranford

package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindComposeFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		files    map[string][]string
		expected int
	}{
		{
			name:     "no files",
			files:    map[string][]string{},
			expected: 0,
		},
		{
			name: "compose.yml",
			files: map[string][]string{
				"yml": {"compose.yml"},
			},
			expected: 1,
		},
		{
			name: "docker-compose.yml",
			files: map[string][]string{
				"yml": {"docker-compose.yml"},
			},
			expected: 1,
		},
		{
			name: "compose.yaml",
			files: map[string][]string{
				"yaml": {"compose.yaml"},
			},
			expected: 1,
		},
		{
			name: "multiple compose files",
			files: map[string][]string{
				"yml":  {"compose.yml", "docker-compose.yml", "compose.demo.yml"},
				"yaml": {"compose.yaml"},
			},
			expected: 4,
		},
		{
			name: "non-compose yml files",
			files: map[string][]string{
				"yml": {"config.yml", "settings.yml"},
			},
			expected: 0,
		},
		{
			name: "mixed files",
			files: map[string][]string{
				"yml": {"compose.yml", "config.yml"},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FindComposeFiles(tt.files)
			require.Len(t, result, tt.expected)
		})
	}
}
