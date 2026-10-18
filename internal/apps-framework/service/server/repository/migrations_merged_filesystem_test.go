// Copyright (c) 2025-2026 Justin Cranford.

package repository

import (
	"embed"
	"errors"
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

const testMergedMigrationName = "migrations/1001_session_management.up.sql"

func TestMergedMigrationsFS_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		domainFS   embed.FS
		migration  string
		wantExists bool
	}{
		{name: "domain filesystem", domainFS: MigrationsFS, migration: testMergedMigrationName, wantExists: true},
		{name: "template fallback", domainFS: embed.FS{}, migration: testMergedMigrationName, wantExists: true},
		{name: "missing migration", domainFS: embed.FS{}, migration: "missing.sql"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			merged := NewMergedMigrationsFS(tc.domainFS)

			file, openErr := merged.Open(tc.migration)
			if !tc.wantExists {
				require.Nil(t, file)
				require.ErrorIs(t, openErr, fs.ErrNotExist)

				return
			}

			require.NoError(t, openErr)
			defer func() { require.NoError(t, file.Close()) }()

			contents, readErr := io.ReadAll(file)
			require.NoError(t, readErr)
			require.NotEmpty(t, contents)

			readFileContents, readFileErr := merged.ReadFile(tc.migration)
			require.NoError(t, readFileErr)
			require.Equal(t, contents, readFileContents)

			info, statErr := merged.Stat(tc.migration)
			require.NoError(t, statErr)
			require.Equal(t, "1001_session_management.up.sql", info.Name())
		})
	}
}

func TestMergedMigrationsFS_ReadFileMissingFromBothSources(t *testing.T) {
	t.Parallel()

	merged := NewMergedMigrationsFS(embed.FS{})
	data, err := merged.ReadFile("migrations/not-present.sql")
	require.Nil(t, data)
	require.ErrorContains(t, err, "failed to read from template")
}

func TestMergedMigrationsFS_ReadDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		domainFS embed.FS
		wantErr  bool
	}{
		{name: "merged directory entries", domainFS: MigrationsFS},
		{name: "template directory fallback", domainFS: embed.FS{}},
		{name: "missing directory", domainFS: embed.FS{}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			merged := NewMergedMigrationsFS(tc.domainFS)

			dir := "migrations"
			if tc.wantErr {
				dir = "missing"
			}

			entries, err := merged.ReadDir(dir)
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorContains(t, err, "directory not found")

				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, entries)
		})
	}
}

func TestMergedMigrationsFS_StatMissingFile(t *testing.T) {
	t.Parallel()

	merged := NewMergedMigrationsFS(embed.FS{})
	_, err := merged.Stat("missing.sql")

	require.Error(t, err)
	require.True(t, errors.Is(err, fs.ErrNotExist))
	require.ErrorContains(t, err, "failed to stat from template")
}
