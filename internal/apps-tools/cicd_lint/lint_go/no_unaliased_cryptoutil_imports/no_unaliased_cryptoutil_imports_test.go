// Copyright (c) 2025-2026 Justin Cranford.
package no_unaliased_cryptoutil_imports

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const testDirectoryTreeWalkErrorMessage = "error walking directory tree"

func TestFindUnaliasedCryptoutilImports_ErrorPaths(t *testing.T) {
	t.Parallel()

	tempFile := filepath.Join(t.TempDir(), "existing.go")
	require.NoError(t, os.WriteFile(tempFile, []byte("package test\n"), 0o600))
	fileInfo, err := os.Stat(tempFile)
	require.NoError(t, err)

	tests := []struct {
		name    string
		walkFn  func(string, filepath.WalkFunc) error
		wantErr string
	}{
		{
			name: "walk error",
			walkFn: func(string, filepath.WalkFunc) error {
				return fmt.Errorf("injected walk error")
			},
			wantErr: testDirectoryTreeWalkErrorMessage,
		},
		{
			name: "walk callback error",
			walkFn: func(root string, walkFn filepath.WalkFunc) error {
				return walkFn(filepath.Join(root, "callback.go"), nil, fmt.Errorf("injected callback error"))
			},
			wantErr: testDirectoryTreeWalkErrorMessage,
		},
		{
			name: "Go file inspection error",
			walkFn: func(root string, walkFn filepath.WalkFunc) error {
				return walkFn(filepath.Join(root, "missing.go"), fileInfo, nil)
			},
			wantErr: "error checking",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := FindUnaliasedCryptoutilImports(tc.walkFn)

			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
