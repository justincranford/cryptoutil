// Copyright (c) 2025 Justin Cranford

package function_var_redeclaration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintGoFunctionVarRedeclaration "cryptoutil/internal/apps/tools/cicd_lint/lint_go/function_var_redeclaration"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// newLogger creates a test logger for the test.
func newLogger(t *testing.T) *cryptoutilCmdCicdCommon.Logger {
	t.Helper()

	return cryptoutilCmdCicdCommon.NewLogger(t.Name())
}

// writeFile creates a file at path with the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
}

func TestCheckInDir_NoViolations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "clean.go"), `package mypackage

import "fmt"

var localVar = "hello"
var number = 42
var computed = fmt.Sprintf("%d", number)
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_DetectsViolation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "seam.go"), `package mypackage

import "path/filepath"

var walkFn = filepath.Walk
var absFn = filepath.Abs
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.Error(t, err)
	require.Contains(t, err.Error(), "function-var-redeclaration")
	require.Contains(t, err.Error(), "2 violation(s)")
}

func TestCheckInDir_SkipsTestFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// _test.go files are allowed to have seam vars.
	writeFile(t, filepath.Join(dir, "seams_test.go"), `package mypackage_test

import "path/filepath"

var walkFn = filepath.Walk
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_SkipsExportTestGo(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// export_test.go is a seam exposure file — allowed to reference unexported names.
	writeFile(t, filepath.Join(dir, "export_test.go"), `package mypackage

import "path/filepath"

var ExportedWalkFn = filepath.Walk
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_SkipsCallExpressions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// pkg.Func() is a call expression — not a seam var.
	writeFile(t, filepath.Join(dir, "constructor.go"), `package mypackage

import "sync"

var mu = sync.NewMutex()
var once = sync.Once{}
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_SkipsTypedVars(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Typed var with no initialiser — not a seam var.
	writeFile(t, filepath.Join(dir, "typed.go"), `package mypackage

import "sync"

var mu sync.Mutex
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_WalkError(t *testing.T) {
	t.Parallel()

	stubWalkFn := func(_ string, _ filepath.WalkFunc) error {
		return fmt.Errorf("injected walk error")
	}

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, ".", stubWalkFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "directory walk failed")
}

func TestCheckInDir_SkipsNonGoFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Non-.go files should not cause violations.
	writeFile(t, filepath.Join(dir, "data.txt"), `var walkFn = filepath.Walk`)
	writeFile(t, filepath.Join(dir, "Makefile"), `var walkFn = filepath.Walk`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_SkipsVendorDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	vendorDir := filepath.Join(dir, cryptoutilSharedMagic.CICDExcludeDirVendor, "somepkg")
	require.NoError(t, os.MkdirAll(vendorDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	writeFile(t, filepath.Join(vendorDir, "seam.go"), `package somepkg

import "path/filepath"

var walkFn = filepath.Walk
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_SkipsDotDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	hidden := filepath.Join(dir, ".hidden", "pkg")
	require.NoError(t, os.MkdirAll(hidden, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	writeFile(t, filepath.Join(hidden, "seam.go"), `package pkg

import "path/filepath"

var walkFn = filepath.Walk
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_SkipsUnderscoreDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	underscore := filepath.Join(dir, "_internal", "pkg")
	require.NoError(t, os.MkdirAll(underscore, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	writeFile(t, filepath.Join(underscore, "seam.go"), `package pkg

import "path/filepath"

var walkFn = filepath.Walk
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheck_NoViolationsOnCurrentCodebase(t *testing.T) {
	t.Parallel()

	// After completing the pre-work refactorings (Task 3.4 pre-work), the codebase
	// must have zero function-var redeclarations in production code.
	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.Check(logger)

	require.NoError(t, err)
}

func TestCheckInDir_MultipleFilesMultipleViolations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"), `package p

import "path/filepath"

var absFn = filepath.Abs
`)
	writeFile(t, filepath.Join(dir, "b.go"), `package p

import "path/filepath"

var walkFn = filepath.Walk
var joinFn = filepath.Join
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.Error(t, err)
	require.Contains(t, err.Error(), "3 violation(s)")
}

func TestCheckInDir_VarGroupMixedValues(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// var block: one is selector-expr (violation), one is basic lit (clean).
	writeFile(t, filepath.Join(dir, "mixed.go"), `package p

import "path/filepath"

var (
	walkFn    = filepath.Walk
	localName = "hello"
)
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.Error(t, err)
	require.Contains(t, err.Error(), "1 violation(s)")
}

func TestCheckInDir_ParseError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Invalid Go syntax — parser returns error; checkFile silently returns nil.
	writeFile(t, filepath.Join(dir, "broken.go"), `package p THIS IS NOT VALID GO {{{`)

	logger := newLogger(t)
	// A parse error causes checkFile to return nil (silent skip), so no violations.
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_NestedSelector(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// var fn = pkg.sub.Method — sel.X is a SelectorExpr not an Ident; must NOT be flagged.
	writeFile(t, filepath.Join(dir, "nested.go"), `package p

import "net/http"

var tlsConn = http.DefaultClient.Transport
`)

	logger := newLogger(t)
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}

func TestCheckInDir_SkipsNonFnNamedVars(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// var defaultCertPEM = magic.SomeConst — selector expr but name does NOT end in "Fn".
	// This is a legitimate pattern for default values from the magic package.
	writeFile(t, filepath.Join(dir, "defaults.go"), `package p

import "path/filepath"

var defaultBase = filepath.Separator
var configPath = filepath.Join
`)

	logger := newLogger(t)
	// configPath references filepath.Join but name doesn't end in "Fn" — not flagged.
	err := lintGoFunctionVarRedeclaration.CheckInDir(logger, dir, filepath.Walk)

	require.NoError(t, err)
}
