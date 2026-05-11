// Copyright (c) 2025-2026 Justin Cranford.

package testmain_integration_tag_policy

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckFile_ReadError(t *testing.T) {
	t.Parallel()

	violations, err := checkFile("ignored", func(string) ([]byte, error) {
		return nil, errors.New("injected read failure")
	})
	require.Nil(t, violations)
	require.Error(t, err)
	require.ErrorContains(t, err, "injected read failure")
}

func TestFindViolationsWithReader_ReadError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	serverDir := filepath.Join(tempDir, "internal", "apps", "sm-kms", "server")
	require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, testmainFileName), []byte("package server\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := findViolationsWithReader(tempDir, func(string) ([]byte, error) {
		return nil, errors.New("injected read failure")
	})
	require.Nil(t, violations)
	require.Error(t, err)
	require.ErrorContains(t, err, "injected read failure")
}

func TestCheckFile_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		wantCount int
		wantTag   string
	}{
		{name: "no tags", content: "package server\n", wantCount: 0},
		{name: "integration go build tag", content: "//go:build integration\npackage server\n", wantCount: 1, wantTag: "//go:build integration"},
		{name: "legacy build tag", content: "// +build integration\npackage server\n", wantCount: 1, wantTag: "// +build integration"},
		{name: "e2e go build tag", content: "//go:build e2e\npackage server\n", wantCount: 1, wantTag: "//go:build e2e"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations, err := checkFile("testmain_test.go", func(string) ([]byte, error) {
				return []byte(tc.content), nil
			})
			require.NoError(t, err)
			require.Len(t, violations, tc.wantCount)

			if tc.wantTag != "" {
				require.Equal(t, tc.wantTag, violations[0].Tag)
			}
		})
	}
}

func TestFindViolationsWithReader_InvalidRootPath(t *testing.T) {
	t.Parallel()

	violations, err := findViolationsWithReader("bad\x00root", os.ReadFile)
	require.Nil(t, violations)
	require.Error(t, err)
	require.ErrorContains(t, err, "walk")
}
