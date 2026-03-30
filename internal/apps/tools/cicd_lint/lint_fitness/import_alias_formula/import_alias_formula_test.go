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

// -----------------------------------------------------------------------
// Test helpers
// -----------------------------------------------------------------------

// buildAliasRoot creates a temp root dir containing a minimal alias_map.yaml.
func buildAliasRoot(t *testing.T) string {
	t.Helper()

	rootDir := t.TempDir()
	yamlDir := filepath.Dir(filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile)))
	require.NoError(t, os.MkdirAll(yamlDir, 0o700))

	yamlContent := `external_aliases:
  - import_path: "encoding/json"
    alias: encodingJson
internal_aliases:
  - import_path: "example.com/myproject/mypackage"
    alias: myprojectMypackage
`
	destPath := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile))
	require.NoError(t, os.WriteFile(destPath, []byte(yamlContent), cryptoutilSharedMagic.FilePermissionsDefault))

	return rootDir
}

// writeGoFile writes a .go source file in dir/pkg/.
func writeGoFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0o700))
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	return path
}

// -----------------------------------------------------------------------
// LoadAliasMap
// -----------------------------------------------------------------------

func TestLoadAliasMap_HappyPath(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)

	m, err := LoadAliasMap(rootDir)

	require.NoError(t, err)
	require.NotNil(t, m)
	require.Len(t, m.ExternalAliases, 1)
	require.Equal(t, "encoding/json", m.ExternalAliases[0].ImportPath)
	require.Equal(t, "encodingJson", m.ExternalAliases[0].Alias)
	require.Len(t, m.InternalAliases, 1)
}

func TestLoadAliasMap_FileNotFound(t *testing.T) {
	t.Parallel()

	m, err := LoadAliasMap(t.TempDir())

	require.Error(t, err)
	require.Nil(t, m)
	require.Contains(t, err.Error(), "failed to read")
}

func TestLoadAliasMap_InvalidYAML(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	yamlDir := filepath.Dir(filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile)))
	require.NoError(t, os.MkdirAll(yamlDir, 0o700))

	destPath := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile))
	require.NoError(t, os.WriteFile(destPath, []byte("!!! not: valid: yaml: ["), cryptoutilSharedMagic.FilePermissionsDefault))

	m, err := LoadAliasMap(rootDir)

	require.Error(t, err)
	require.Nil(t, m)
	require.Contains(t, err.Error(), "failed to parse")
}

func TestLoadAliasMap_ReadFileError(t *testing.T) {
	orig := importAliasReadFileFn
	importAliasReadFileFn = func(_ string) ([]byte, error) { return nil, errors.New("read error") }

	defer func() { importAliasReadFileFn = orig }()

	m, err := LoadAliasMap("dummy")

	require.Error(t, err)
	require.Nil(t, m)
	require.Contains(t, err.Error(), "failed to read")
}

// -----------------------------------------------------------------------
// AllEntries
// -----------------------------------------------------------------------

func TestAllEntries_CombinesLists(t *testing.T) {
	t.Parallel()

	m := &AliasMap{
		ExternalAliases: []AliasEntry{{ImportPath: "a", Alias: "aa"}},
		InternalAliases: []AliasEntry{{ImportPath: "b", Alias: "bb"}, {ImportPath: "c", Alias: "cc"}},
	}

	entries := AllEntries(m)

	require.Len(t, entries, 3)
}

func TestAllEntries_EmptyMap(t *testing.T) {
	t.Parallel()

	entries := AllEntries(&AliasMap{})

	require.Empty(t, entries)
}

// -----------------------------------------------------------------------
// CheckInDir — happy path
// -----------------------------------------------------------------------

func TestCheckInDir_HappyPath_NoViolations(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	pkgDir := filepath.Join(rootDir, "mypkg")
	writeGoFile(t, pkgDir, "correct.go", `package mypkg

import (
	encodingJson "encoding/json"
	"fmt"
)

var _ = encodingJson.Marshal
var _ = fmt.Println
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_ViolationWrongAlias(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	pkgDir := filepath.Join(rootDir, "mypkg")
	writeGoFile(t, pkgDir, "wrong.go", `package mypkg

import (
	json "encoding/json"
)

var _ = json.Marshal
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation(s)")
}

func TestCheckInDir_ViolationNoAlias(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	pkgDir := filepath.Join(rootDir, "mypkg")
	writeGoFile(t, pkgDir, "noalias.go", `package mypkg

import (
	"encoding/json"
)

var _ = json.Marshal
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation(s)")
}

func TestCheckInDir_BlankImportAllowed(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	pkgDir := filepath.Join(rootDir, "mypkg")
	writeGoFile(t, pkgDir, "blank.go", `package mypkg

import (
	_ "encoding/json"
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_DotImportAllowed(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	pkgDir := filepath.Join(rootDir, "mypkg")
	// Dot imports expose all exported identifiers; project convention allows them
	// without requiring an explicit alias from the alias map.
	writeGoFile(t, pkgDir, "dot.go", `package mypkg

import (
	. "encoding/json"
)

var _ = Marshal
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_UnparsableFileSkipped(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	pkgDir := filepath.Join(rootDir, "mypkg")
	// This file has a syntax error — the AST parser will fail and skip it.
	writeGoFile(t, pkgDir, "broken.go", `this is not valid go source code !!!`)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_GeneratedFileSkipped(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	pkgDir := filepath.Join(rootDir, "mypkg")
	// Generated file uses wrong alias; should be skipped entirely.
	writeGoFile(t, pkgDir, "generated.go", `// Code generated by some-tool/v2; DO NOT EDIT.
package mypkg

import (
	"encoding/json"
)

var _ = json.Marshal
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_EmptyAliasMap_Skips(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	yamlDir := filepath.Dir(filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile)))
	require.NoError(t, os.MkdirAll(yamlDir, 0o700))

	destPath := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile))
	require.NoError(t, os.WriteFile(destPath, []byte("external_aliases: []\ninternal_aliases: []\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_ExcludedVendorDir(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	// Put a violating file inside vendor/ — it should be skipped.
	vendorPkgDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirVendor, "somepkg")
	writeGoFile(t, vendorPkgDir, "violation.go", `package somepkg

import (
	"encoding/json"
)

var _ = json.Marshal
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_LoadAliasMapError(t *testing.T) {
	// File not found → LoadAliasMap fails.
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, t.TempDir())

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read")
}

func TestCheckInDir_WalkCallbackError(t *testing.T) {
	orig := importAliasWalkDirFn
	importAliasWalkDirFn = func(_ string, fn fs.WalkDirFunc) error {
		// Simulate a walk-callback error.
		return fn("somepath", nil, errors.New("permission denied"))
	}

	defer func() { importAliasWalkDirFn = orig }()

	rootDir := buildAliasRoot(t)
	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk")
}

func TestCheckInDir_CheckFileReadError(t *testing.T) {
	// Set up a rootDir with a valid alias map and one .go file.
	rootDir := buildAliasRoot(t)
	pkgDir := filepath.Join(rootDir, "mypkg")
	writeGoFile(t, pkgDir, "ok.go", `package mypkg
`)

	// First call (read alias map) should succeed with the real function;
	// subsequent calls (for .go files) should fail.
	callCount := 0
	orig := importAliasReadFileFn
	origReal := os.ReadFile
	importAliasReadFileFn = func(path string) ([]byte, error) {
		callCount++
		// Allow the first read (alias map YAML) to succeed.
		if callCount == 1 {
			return origReal(path)
		}

		return nil, errors.New("disk error")
	}

	defer func() { importAliasReadFileFn = orig }()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "disk error")
}

// -----------------------------------------------------------------------
// Check (top-level, uses project root via seam)
// -----------------------------------------------------------------------

func TestCheck_ProjectRootNotFound(t *testing.T) {
	orig := findImportAliasProjectRootFn
	findImportAliasProjectRootFn = func() (string, error) {
		return "", errors.New("go.mod not found")
	}

	defer func() { findImportAliasProjectRootFn = orig }()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := Check(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "go.mod not found")
}

func TestCheck_HappyPath(t *testing.T) {
	// Use a temp dir as the "project root"; alias map returns no required aliases,
	// so CheckInDir will succeed without walking any Go files.
	rootDir := t.TempDir()
	yamlDir := filepath.Dir(filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile)))
	require.NoError(t, os.MkdirAll(yamlDir, 0o700))

	destPath := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile))
	require.NoError(t, os.WriteFile(destPath, []byte("external_aliases: []\ninternal_aliases: []\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	orig := findImportAliasProjectRootFn
	findImportAliasProjectRootFn = func() (string, error) { return rootDir, nil }

	defer func() { findImportAliasProjectRootFn = orig }()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := Check(logger)

	require.NoError(t, err)
}

// -----------------------------------------------------------------------
// findImportAliasProjectRoot (indirectly via seam)
// -----------------------------------------------------------------------

func TestFindProjectRoot_GetwdError(t *testing.T) {
	orig := importAliasGetwdFn
	importAliasGetwdFn = func() (string, error) { return "", errors.New("getwd failed") }

	defer func() { importAliasGetwdFn = orig }()

	_, err := findImportAliasProjectRoot()

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get working directory")
}

func TestFindProjectRoot_GoModNotFound(t *testing.T) {
	orig := importAliasGetwdFn
	// Point to a temp dir that has no go.mod ancestor.
	importAliasGetwdFn = func() (string, error) { return t.TempDir(), nil }

	defer func() { importAliasGetwdFn = orig }()

	_, err := findImportAliasProjectRoot()

	require.Error(t, err)
	require.Contains(t, err.Error(), "go.mod not found")
}

func TestFindProjectRoot_HappyPath(t *testing.T) {
	t.Parallel()

	// Real cwd is inside the project which has a go.mod — should succeed.
	root, err := findImportAliasProjectRoot()

	require.NoError(t, err)
	require.NotEmpty(t, root)

	_, statErr := os.Stat(filepath.Join(root, "go.mod"))
	require.NoError(t, statErr, "returned root should contain go.mod")
}

// -----------------------------------------------------------------------
// isGeneratedGoFile
// -----------------------------------------------------------------------

func TestIsGeneratedGoFile_True(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "gen.go")
	require.NoError(t, os.WriteFile(path, []byte("// Code generated by mytool; DO NOT EDIT.\npackage x\n"), cryptoutilSharedMagic.CacheFilePermissions))

	require.True(t, isGeneratedGoFile(path))
}

func TestIsGeneratedGoFile_False(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "normal.go")
	require.NoError(t, os.WriteFile(path, []byte("// Copyright (c) 2025\npackage x\n"), cryptoutilSharedMagic.CacheFilePermissions))

	require.False(t, isGeneratedGoFile(path))
}

func TestIsGeneratedGoFile_ReadError(t *testing.T) {
	// isGeneratedGoFile returns false on read error.
	require.False(t, isGeneratedGoFile("/nonexistent/path/gen.go"))
}

func TestIsGeneratedGoFile_LargeFile_MarkerInFirst512Bytes(t *testing.T) {
	t.Parallel()

	// File larger than 512 bytes with "Code generated" in the first 512 bytes.
	dir := t.TempDir()
	path := filepath.Join(dir, "big_gen.go")
	header := "// Code generated by mytool; DO NOT EDIT.\npackage x\n"
	padding := make([]byte, 600) // more than codeGeneratedCheckBytes
	require.NoError(t, os.WriteFile(path, append([]byte(header), padding...), cryptoutilSharedMagic.CacheFilePermissions))

	require.True(t, isGeneratedGoFile(path))
}

func TestIsGeneratedGoFile_LargeFile_NoMarker(t *testing.T) {
	t.Parallel()

	// File larger than 512 bytes with NO "Code generated" marker.
	dir := t.TempDir()
	path := filepath.Join(dir, "big_normal.go")
	header := "// Copyright (c) 2025 example\npackage x\n"
	padding := make([]byte, 600)
	require.NoError(t, os.WriteFile(path, append([]byte(header), padding...), cryptoutilSharedMagic.CacheFilePermissions))

	require.False(t, isGeneratedGoFile(path))
}

// -----------------------------------------------------------------------
// Raw string literal — no false positives from AST parser
// -----------------------------------------------------------------------

func TestCheckInDir_RawStringLiteralNoFalsePositive(t *testing.T) {
	t.Parallel()

	rootDir := buildAliasRoot(t)
	pkgDir := filepath.Join(rootDir, "mypkg")
	// The file contains a raw string with "encoding/json" inside a backtick block.
	// The AST parser should NOT report a violation for content inside raw strings.
	writeGoFile(t, pkgDir, "rawstr.go", `package mypkg

func example() string {
	return `+"`"+`
import (
	"encoding/json"
)
`+"`"+`
}
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}
