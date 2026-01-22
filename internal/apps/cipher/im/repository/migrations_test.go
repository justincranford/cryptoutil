// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMergedFS_ReadFile tests the ReadFile method of the merged filesystem.
func TestMergedFS_ReadFile(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Type assert to ReadFileFS to test ReadFile method.
	readFileFS, ok := mergedFS.(fs.ReadFileFS)
	require.True(t, ok, "mergedFS should implement fs.ReadFileFS")

	tests := []struct {
		name        string
		filename    string
		wantErr     bool
		wantContent bool
	}{
		{
			name:        "read cipher-im migration",
			filename:    "migrations/2001_init.up.sql",
			wantErr:     false,
			wantContent: true,
		},
		{
			name:        "read template migration",
			filename:    "migrations/1001_session_management.up.sql",
			wantErr:     false,
			wantContent: true,
		},
		{
			name:        "read non-existent file",
			filename:    "migrations/9999_nonexistent.sql",
			wantErr:     true,
			wantContent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := readFileFS.ReadFile(tt.filename)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, data)
			} else {
				require.NoError(t, err)

				if tt.wantContent {
					require.NotEmpty(t, data)
				}
			}
		})
	}
}

// TestMergedFS_Stat tests the Stat method of the merged filesystem.
func TestMergedFS_Stat(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Type assert to StatFS to test Stat method.
	statFS, ok := mergedFS.(fs.StatFS)
	require.True(t, ok, "mergedFS should implement fs.StatFS")

	tests := []struct {
		name     string
		filename string
		wantErr  bool
		wantDir  bool
	}{
		{
			name:     "stat cipher-im migration",
			filename: "migrations/2001_init.up.sql",
			wantErr:  false,
			wantDir:  false,
		},
		{
			name:     "stat template migration",
			filename: "migrations/1001_session_management.up.sql",
			wantErr:  false,
			wantDir:  false,
		},
		{
			name:     "stat non-existent file",
			filename: "migrations/9999_nonexistent.sql",
			wantErr:  true,
			wantDir:  false,
		},
		{
			name:     "stat root directory",
			filename: ".",
			wantErr:  false,
			wantDir:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info, err := statFS.Stat(tt.filename)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, info)
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)

				if tt.wantDir {
					require.True(t, info.IsDir())
				} else {
					require.False(t, info.IsDir())
					require.Greater(t, info.Size(), int64(0))
				}
			}
		})
	}
}

// TestMergedFS_Open tests the Open method for file reading.
func TestMergedFS_Open(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "open cipher-im migration",
			filename: "migrations/2001_init.up.sql",
			wantErr:  false,
		},
		{
			name:     "open template migration",
			filename: "migrations/1001_session_management.up.sql",
			wantErr:  false,
		},
		{
			name:     "open non-existent file",
			filename: "migrations/9999_nonexistent.sql",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			file, err := mergedFS.Open(tt.filename)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, file)
			} else {
				require.NoError(t, err)
				require.NotNil(t, file)

				// Read some content to verify file is readable.
				data, readErr := io.ReadAll(file)
				require.NoError(t, readErr)
				require.NotEmpty(t, data)

				// Close the file.
				require.NoError(t, file.Close())
			}
		})
	}
}

// TestMergedFS_ReadDir tests the ReadDir method.
func TestMergedFS_ReadDir(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Type assert to ReadDirFS to test ReadDir method.
	readDirFS, ok := mergedFS.(fs.ReadDirFS)
	require.True(t, ok, "mergedFS should implement fs.ReadDirFS")

	entries, err := readDirFS.ReadDir("migrations")
	require.NoError(t, err)

	// Should have at least template migrations (1001-1004) plus cipher-im migrations (2001+).
	require.GreaterOrEqual(t, len(entries), 5, "should have at least 5 migration files")

	// Verify entries contain expected migrations.
	fileNames := make(map[string]bool)
	for _, entry := range entries {
		fileNames[entry.Name()] = true
	}

	// Check for template migrations.
	require.True(t, fileNames["1001_session_management.up.sql"], "should contain 1001_session_management.up.sql")

	// Check for cipher-im migrations.
	require.True(t, fileNames["2001_init.up.sql"], "should contain 2001_init.up.sql")
}
