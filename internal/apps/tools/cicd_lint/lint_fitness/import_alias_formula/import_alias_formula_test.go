// Copyright (c) 2025 Justin Cranford

package import_alias_formula

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// buildAliasRoot creates a temp root dir containing a minimal alias_map.yaml.
func buildAliasRoot(t *testing.T) string {
	t.Helper()

	rootDir := t.TempDir()
	yamlDir := filepath.Dir(filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile)))
	require.NoError(t, os.MkdirAll(yamlDir, 0o700))

	yamlContent := "external_aliases:\n  - import_path: \"encoding/json\"\n    alias: encodingJson\ninternal_aliases:\n  - import_path: \"example.com/myproject/mypackage\"\n    alias: myprojectMypackage\n"
	destPath := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile))
	require.NoError(t, os.WriteFile(destPath, []byte(yamlContent), cryptoutilSharedMagic.FilePermissionsDefault))

	return rootDir
}

// buildEmptyAliasRoot creates a temp root dir containing an empty alias_map.yaml.
func buildEmptyAliasRoot(t *testing.T) string {
	t.Helper()

	rootDir := t.TempDir()
	yamlDir := filepath.Dir(filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile)))
	require.NoError(t, os.MkdirAll(yamlDir, 0o700))

	destPath := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile))
	require.NoError(t, os.WriteFile(destPath, []byte("external_aliases: []\ninternal_aliases: []\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	return rootDir
}

// writeGoFile writes a .go source file in dir.
func writeGoFile(t *testing.T, dir, name, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))
}

func TestLoadAliasMap_HappyPath(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	m, err := LoadAliasMap(rootDir, os.ReadFile)

	require.NoError(t, err)
	require.NotNil(t, m)
	require.Len(t, m.ExternalAliases, 1)
	require.Equal(t, "encoding/json", m.ExternalAliases[0].ImportPath)
	require.Equal(t, "encodingJson", m.ExternalAliases[0].Alias)
	require.Len(t, m.InternalAliases, 1)
}

func TestLoadAliasMap_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rootDir    string
		readFileFn func(string) ([]byte, error)
		wantErr    string
	}{
		{name: "file not found", rootDir: t.TempDir(), readFileFn: os.ReadFile, wantErr: "failed to read"},
		{name: "read file error", rootDir: "dummy", readFileFn: func(_ string) ([]byte, error) { return nil, errors.New("read error") }, wantErr: "failed to read"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m, err := LoadAliasMap(tc.rootDir, tc.readFileFn)
			require.Error(t, err)
			require.Nil(t, m)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestLoadAliasMap_InvalidYAML(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	yamlDir := filepath.Dir(filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile)))
	require.NoError(t, os.MkdirAll(yamlDir, 0o700))

	destPath := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile))
	require.NoError(t, os.WriteFile(destPath, []byte("!!! not: valid: yaml: ["), cryptoutilSharedMagic.FilePermissionsDefault))

	m, err := LoadAliasMap(rootDir, os.ReadFile)
	require.Error(t, err)
	require.Nil(t, m)
	require.Contains(t, err.Error(), "failed to parse")
}

func TestAllEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		aliasMap *AliasMap
		wantLen  int
	}{
		{
			name: "combines lists",
			aliasMap: &AliasMap{
				ExternalAliases: []AliasEntry{{ImportPath: "a", Alias: "aa"}},
				InternalAliases: []AliasEntry{{ImportPath: "b", Alias: "bb"}, {ImportPath: "c", Alias: "cc"}},
			},
			wantLen: 3,
		},
		{name: "empty map", aliasMap: &AliasMap{}, wantLen: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Len(t, AllEntries(tc.aliasMap), tc.wantLen)
		})
	}
}

func TestCheckInDir_HappyPath_NoViolations(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	writeGoFile(t, filepath.Join(rootDir, "mypkg"), "correct.go", "package mypkg\n\nimport (\n\tencodingJson \"encoding/json\"\n\t\"fmt\"\n)\n\nvar _ = encodingJson.Marshal\nvar _ = fmt.Println\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	require.NoError(t, CheckInDir(logger, rootDir, os.ReadFile, filepath.WalkDir))
}

func TestCheckInDir_FileHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr string
	}{
		{
			name: "violation wrong alias",
			setup: func(t *testing.T) string {
				t.Helper()
				rootDir := buildAliasRoot(t)
				writeGoFile(t, filepath.Join(rootDir, "mypkg"), "wrong.go", "package mypkg\n\nimport (\n\tjson \"encoding/json\"\n)\n\nvar _ = json.Marshal\n")

				return rootDir
			},
			wantErr: "violation(s)",
		},
		{
			name: "violation no alias",
			setup: func(t *testing.T) string {
				t.Helper()
				rootDir := buildAliasRoot(t)
				writeGoFile(t, filepath.Join(rootDir, "mypkg"), "noalias.go", "package mypkg\n\nimport (\n\t\"encoding/json\"\n)\n\nvar _ = json.Marshal\n")

				return rootDir
			},
			wantErr: "violation(s)",
		},
		{
			name: "blank import allowed",
			setup: func(t *testing.T) string {
				t.Helper()
				rootDir := buildAliasRoot(t)
				writeGoFile(t, filepath.Join(rootDir, "mypkg"), "blank.go", "package mypkg\n\nimport (\n\t_ \"encoding/json\"\n)\n")

				return rootDir
			},
		},
		{
			name: "dot import allowed",
			setup: func(t *testing.T) string {
				t.Helper()
				rootDir := buildAliasRoot(t)
				writeGoFile(t, filepath.Join(rootDir, "mypkg"), "dot.go", "package mypkg\n\nimport (\n\t. \"encoding/json\"\n)\n\nvar _ = Marshal\n")

				return rootDir
			},
		},
		{
			name: "unparsable file skipped",
			setup: func(t *testing.T) string {
				t.Helper()
				rootDir := buildAliasRoot(t)
				writeGoFile(t, filepath.Join(rootDir, "mypkg"), "broken.go", "this is not valid go source code !!!")

				return rootDir
			},
		},
		{
			name: "generated file skipped",
			setup: func(t *testing.T) string {
				t.Helper()
				rootDir := buildAliasRoot(t)
				writeGoFile(t, filepath.Join(rootDir, "mypkg"), "generated.go", "// Code generated by some-tool/v2; DO NOT EDIT.\npackage mypkg\n\nimport (\n\t\"encoding/json\"\n)\n\nvar _ = json.Marshal\n")

				return rootDir
			},
		},
		{
			name: "empty alias map skips",
			setup: func(t *testing.T) string {
				t.Helper()

				return buildEmptyAliasRoot(t)
			},
		},
		{
			name: "excluded vendor dir",
			setup: func(t *testing.T) string {
				t.Helper()
				rootDir := buildAliasRoot(t)
				writeGoFile(t, filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirVendor, "somepkg"), "violation.go", "package somepkg\n\nimport (\n\t\"encoding/json\"\n)\n\nvar _ = json.Marshal\n")

				return rootDir
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := tc.setup(t)
			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err := CheckInDir(logger, rootDir, os.ReadFile, filepath.WalkDir)

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCheckInDir_LoadAliasMapError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, t.TempDir(), os.ReadFile, filepath.WalkDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read")
}

func TestCheckInDir_WalkCallbackError(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir, os.ReadFile, func(_ string, fn fs.WalkDirFunc) error {
		return fn("somepath", nil, errors.New("permission denied"))
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk")
}

func TestCheckInDir_CheckFileReadError(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	writeGoFile(t, filepath.Join(rootDir, "mypkg"), "ok.go", "package mypkg\n")

	callCount := 0
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir, func(path string) ([]byte, error) {
		callCount++
		if callCount == 1 {
			return os.ReadFile(path)
		}

		return nil, errors.New("disk error")
	}, filepath.WalkDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "disk error")
}

// Sequential: mutates findImportAliasProjectRootFn package-level state.
func TestCheck_ProjectRootNotFound(t *testing.T) {
	orig := findImportAliasProjectRootFn
	findImportAliasProjectRootFn = func() (string, error) { return "", errors.New("go.mod not found") }

	defer func() { findImportAliasProjectRootFn = orig }()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "go.mod not found")
}

// Sequential: mutates findImportAliasProjectRootFn package-level state.
func TestCheck_HappyPath(t *testing.T) {
	rootDir := buildEmptyAliasRoot(t)

	orig := findImportAliasProjectRootFn
	findImportAliasProjectRootFn = func() (string, error) { return rootDir, nil }

	defer func() { findImportAliasProjectRootFn = orig }()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	require.NoError(t, Check(logger))
}

func TestFindProjectRoot_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		getwdFn func() (string, error)
		wantErr string
	}{
		{name: "getwd error", getwdFn: func() (string, error) { return "", errors.New("getwd failed") }, wantErr: "failed to get working directory"},
		{name: "go.mod not found", getwdFn: func() (string, error) { return t.TempDir(), nil }, wantErr: "go.mod not found"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := findImportAliasProjectRoot(tc.getwdFn)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestFindProjectRoot_HappyPath(t *testing.T) {
	t.Parallel()

	root, err := findImportAliasProjectRoot(os.Getwd)
	require.NoError(t, err)
	require.NotEmpty(t, root)

	_, statErr := os.Stat(filepath.Join(root, "go.mod"))
	require.NoError(t, statErr, "returned root should contain go.mod")
}

func TestIsGeneratedGoFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantGen bool
	}{
		{
			name: "generated marker present",
			setup: func(t *testing.T) string {
				t.Helper()
				p := filepath.Join(t.TempDir(), "gen.go")
				require.NoError(t, os.WriteFile(p, []byte("// Code generated by mytool; DO NOT EDIT.\npackage x\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return p
			},
			wantGen: true,
		},
		{
			name: "no generated marker",
			setup: func(t *testing.T) string {
				t.Helper()
				p := filepath.Join(t.TempDir(), "normal.go")
				require.NoError(t, os.WriteFile(p, []byte("// Copyright (c) 2025\npackage x\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return p
			},
			wantGen: false,
		},
		{name: "read error returns false", setup: func(_ *testing.T) string { return "/nonexistent/path/gen.go" }, wantGen: false},
		{
			name: "large file with marker in first 512 bytes",
			setup: func(t *testing.T) string {
				t.Helper()
				p := filepath.Join(t.TempDir(), "big_gen.go")
				header := "// Code generated by mytool; DO NOT EDIT.\npackage x\n"
				require.NoError(t, os.WriteFile(p, append([]byte(header), make([]byte, 600)...), cryptoutilSharedMagic.CacheFilePermissions))

				return p
			},
			wantGen: true,
		},
		{
			name: "large file without marker",
			setup: func(t *testing.T) string {
				t.Helper()
				p := filepath.Join(t.TempDir(), "big_normal.go")
				header := "// Copyright (c) 2025 example\npackage x\n"
				require.NoError(t, os.WriteFile(p, append([]byte(header), make([]byte, 600)...), cryptoutilSharedMagic.CacheFilePermissions))

				return p
			},
			wantGen: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.wantGen, isGeneratedGoFile(tc.setup(t), os.ReadFile))
		})
	}
}

func TestCheckInDir_RawStringLiteralNoFalsePositive(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	writeGoFile(t, filepath.Join(rootDir, "mypkg"), "rawstr.go", "package mypkg\n\nfunc example() string {\n\treturn `\nimport (\n\t\"encoding/json\"\n)\n`\n}\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	require.NoError(t, CheckInDir(logger, rootDir, os.ReadFile, filepath.WalkDir))
}
