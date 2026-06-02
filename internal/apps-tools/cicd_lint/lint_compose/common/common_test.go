// Copyright (c) 2025-2026 Justin Cranford.
package common

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const testConfigYML = "config.yml"

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
			name: cryptoutilSharedMagic.COMPOSE_YML,
			files: map[string][]string{
				cryptoutilSharedMagic.YML: {cryptoutilSharedMagic.COMPOSE_YML},
			},
			expected: 1,
		},
		{
			name: cryptoutilSharedMagic.DOCKER_COMPOSE_YML,
			files: map[string][]string{
				cryptoutilSharedMagic.YML: {cryptoutilSharedMagic.DOCKER_COMPOSE_YML},
			},
			expected: 1,
		},
		{
			name: cryptoutilSharedMagic.COMPOSE_YAML,
			files: map[string][]string{
				cryptoutilSharedMagic.YAML: {cryptoutilSharedMagic.COMPOSE_YAML},
			},
			expected: 1,
		},
		{
			name: "multiple compose files",
			files: map[string][]string{
				cryptoutilSharedMagic.YML:  {cryptoutilSharedMagic.COMPOSE_YML, cryptoutilSharedMagic.DOCKER_COMPOSE_YML, "compose.demo.yml"},
				cryptoutilSharedMagic.YAML: {cryptoutilSharedMagic.COMPOSE_YAML},
			},
			expected: 4,
		},
		{
			name: "non-compose yml files",
			files: map[string][]string{
				cryptoutilSharedMagic.YML: {testConfigYML, "settings.yml"},
			},
			expected: 0,
		},
		{
			name: "mixed files",
			files: map[string][]string{
				cryptoutilSharedMagic.YML: {cryptoutilSharedMagic.COMPOSE_YML, testConfigYML},
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
