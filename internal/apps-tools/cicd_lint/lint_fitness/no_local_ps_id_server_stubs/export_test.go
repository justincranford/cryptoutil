// Copyright (c) 2025-2026 Justin Cranford.
package no_local_ps_id_server_stubs

import (
	"go/ast"
	"go/token"
	"io/fs"
)

func ExportedFindViolationsWithWalkDir(rootDir string, walkDirFn func(string, fs.WalkDirFunc) error) ([]Violation, error) {
	return findViolations(rootDir, walkDirFn)
}

func ExportedFindViolationsWithFunctions(
	rootDir string,
	walkDirFn func(string, fs.WalkDirFunc) error,
	checkFileFn func(string) ([]Violation, error),
) ([]Violation, error) {
	return findViolationsWithFunctions(rootDir, walkDirFn, checkFileFn)
}

func ExportedIsPSIDTestFile(path string) bool {
	return isPSIDTestFile(path)
}

func ExportedReceiverTypeName(expr ast.Expr) string {
	return receiverTypeName(expr)
}

func ExportedFindTypeLine(fset *token.FileSet, file *ast.File, typeName string) int {
	return findTypeLine(fset, file, typeName)
}

func ExportedCheckFile(path string) ([]Violation, error) {
	return checkFile(path)
}

func ExportedCollectReceiverMethods(file *ast.File) map[string]map[string]bool {
	return collectReceiverMethods(file)
}
