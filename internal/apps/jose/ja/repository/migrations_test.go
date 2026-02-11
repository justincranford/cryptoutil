// Copyright (c) 2025 Justin Cranford

package repository

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetMergedMigrationsFS(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()
	require.NotNil(t, mergedFS)
}

func TestMergedFS_Open_JoseJAFile(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Open a jose-ja migration file (2001+).
	file, err := mergedFS.Open("migrations/2001_elastic_jwks.up.sql")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NoError(t, file.Close())
}

func TestMergedFS_Open_TemplateFile(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Open a template migration file (1001-1004).
	file, err := mergedFS.Open("migrations/1001_session_management.up.sql")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NoError(t, file.Close())
}

func TestMergedFS_Open_NonExistent(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Opening non-existent file should fail.
	file, err := mergedFS.Open("migrations/9999_nonexistent.up.sql")
	require.Error(t, err)
	require.Nil(t, file)
}

func TestMergedFS_ReadDir(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Cast to fs.ReadDirFS to use ReadDir.
	readDirFS, ok := mergedFS.(fs.ReadDirFS)
	require.True(t, ok, "mergedFS should implement fs.ReadDirFS")

	entries, err := readDirFS.ReadDir("migrations")
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	// Should contain both template migrations (1001+) and jose-ja migrations (2001+).
	hasTemplate := false
	hasJoseJA := false

	for _, entry := range entries {
		name := entry.Name()

		if name == "1001_session_management.up.sql" {
			hasTemplate = true
		}

		if name == "2001_elastic_jwks.up.sql" {
			hasJoseJA = true
		}
	}

	require.True(t, hasTemplate, "Should contain template migration 1001")
	require.True(t, hasJoseJA, "Should contain jose-ja migration 2001")
}

func TestMergedFS_ReadFile(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Cast to fs.ReadFileFS to use ReadFile.
	readFileFS, ok := mergedFS.(fs.ReadFileFS)
	require.True(t, ok, "mergedFS should implement fs.ReadFileFS")

	// Read a jose-ja migration file.
	data, err := readFileFS.ReadFile("migrations/2001_elastic_jwks.up.sql")
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.Contains(t, string(data), "CREATE TABLE")
}

func TestMergedFS_ReadFile_TemplateFile(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Cast to fs.ReadFileFS to use ReadFile.
	readFileFS, ok := mergedFS.(fs.ReadFileFS)
	require.True(t, ok, "mergedFS should implement fs.ReadFileFS")

	// Read a template migration file.
	data, err := readFileFS.ReadFile("migrations/1001_session_management.up.sql")
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.Contains(t, string(data), "CREATE TABLE")
}

func TestMergedFS_ReadFile_NonExistent(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Cast to fs.ReadFileFS to use ReadFile.
	readFileFS, ok := mergedFS.(fs.ReadFileFS)
	require.True(t, ok, "mergedFS should implement fs.ReadFileFS")

	// Reading non-existent file should fail.
	data, err := readFileFS.ReadFile("migrations/9999_nonexistent.up.sql")
	require.Error(t, err)
	require.Nil(t, data)
}

func TestMergedFS_Stat_JoseJAFile(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Cast to fs.StatFS to use Stat.
	statFS, ok := mergedFS.(fs.StatFS)
	require.True(t, ok, "mergedFS should implement fs.StatFS")

	// Stat a jose-ja migration file.
	info, err := statFS.Stat("migrations/2001_elastic_jwks.up.sql")
	require.NoError(t, err)
	require.NotNil(t, info)
	require.False(t, info.IsDir())
	require.True(t, info.Size() > 0)
}

func TestMergedFS_Stat_TemplateFile(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Cast to fs.StatFS to use Stat.
	statFS, ok := mergedFS.(fs.StatFS)
	require.True(t, ok, "mergedFS should implement fs.StatFS")

	// Stat a template migration file.
	info, err := statFS.Stat("migrations/1001_session_management.up.sql")
	require.NoError(t, err)
	require.NotNil(t, info)
	require.False(t, info.IsDir())
	require.True(t, info.Size() > 0)
}

func TestMergedFS_Stat_NonExistent(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Cast to fs.StatFS to use Stat.
	statFS, ok := mergedFS.(fs.StatFS)
	require.True(t, ok, "mergedFS should implement fs.StatFS")

	// Stating non-existent file should fail.
	info, err := statFS.Stat("migrations/9999_nonexistent.up.sql")
	require.Error(t, err)
	require.Nil(t, info)
}

func TestMergedFS_Stat_Directory(t *testing.T) {
	t.Parallel()

	mergedFS := GetMergedMigrationsFS()

	// Cast to fs.StatFS to use Stat.
	statFS, ok := mergedFS.(fs.StatFS)
	require.True(t, ok, "mergedFS should implement fs.StatFS")

	// Stat the migrations directory.
	info, err := statFS.Stat("migrations")
	require.NoError(t, err)
	require.NotNil(t, info)
	require.True(t, info.IsDir())
}
