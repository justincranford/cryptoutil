// Copyright (c) 2025 Justin Cranford

// Package no_local_ps_id_server_stubs detects struct types in PS-ID test files that
// re-implement IPublicServer or IAdminServer locally instead of using the shared stubs
// in internal/apps/framework/service/testing/stubs.
//
// A struct is flagged when its file contains all methods of either interface:
//
//	IPublicServer: Start, Shutdown, ActualPort, PublicBaseURL
//	IAdminServer:  Start, Shutdown, ActualPort, SetReady, AdminBaseURL, AdminTLSRootCAPool
//
// Only files under internal/apps/{ps-id}/ are scanned (framework and tools dirs are excluded).
// A stub belongs in the shared package; duplicating it in each PS-ID creates drift and bugs.
package no_local_ps_id_server_stubs

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// publicServerMethods is the complete method set of IPublicServer.
var publicServerMethods = []string{"Start", "Shutdown", "ActualPort", "PublicBaseURL"}

// adminServerMethods is the complete method set of IAdminServer.
var adminServerMethods = []string{"Start", "Shutdown", "ActualPort", "SetReady", "AdminBaseURL", "AdminTLSRootCAPool"}

// Violation records one detected local stub type.
type Violation struct {
	File       string
	Line       int
	StructName string
	Interface  string
}

// Check runs the linter from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir scans PS-ID directories under rootDir for local stub types.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking for local PS-ID server stubs that should use the shared stubs package...")

	violations, err := FindViolations(rootDir)
	if err != nil {
		return fmt.Errorf("no-local-ps-id-server-stubs: failed to walk directory: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "%s:%d: struct %q re-implements %s locally — use cryptoutil/internal/apps/framework/service/testing/stubs instead\n",
				v.File, v.Line, v.StructName, v.Interface)
		}

		return fmt.Errorf("found %d local PS-ID server stub(s): move to shared stubs package", len(violations))
	}

	logger.LogWithPrefix("no-local-ps-id-server-stubs", "✅ No local PS-ID server stubs found")

	return nil
}

// FindViolations returns all violations found under rootDir.
func FindViolations(rootDir string) ([]Violation, error) {
	var violations []Violation

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			switch d.Name() {
			case cryptoutilSharedMagic.CICDExcludeDirVendor, cryptoutilSharedMagic.CICDExcludeDirGit:
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		normalized := filepath.ToSlash(path)

		if !isPSIDTestFile(normalized) {
			return nil
		}

		vs, err := checkFile(path)
		if err != nil {
			return err
		}

		violations = append(violations, vs...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", rootDir, err)
	}

	return violations, nil
}

// isPSIDTestFile returns true for test files under internal/apps/{ps-id}/ but not under
// framework/, tools/, or template/ directories.
func isPSIDTestFile(normalized string) bool {
	const prefix = "internal/apps/"

	idx := strings.Index(normalized, prefix)
	if idx < 0 {
		return false
	}

	rest := normalized[idx+len(prefix):]
	slash := strings.Index(rest, "/")

	if slash < 0 {
		return false
	}

	psID := rest[:slash]

	// Exclude shared infrastructure directories.
	switch psID {
	case "framework", "tools", "template":
		return false
	}

	return true
}

// checkFile parses one test file and returns stub violations.
func checkFile(filePath string) ([]Violation, error) {
	fset := token.NewFileSet()

	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	f, err := parser.ParseFile(fset, filePath, src, 0)
	if err != nil {
		return nil, nil //nolint:nilerr // skip files that don't parse (build-tag-excluded, etc.)
	}

	// Collect all method names defined on each receiver type in this file.
	// receiverMethods maps type name → set of method names.
	receiverMethods := collectReceiverMethods(f)

	var violations []Violation

	for typeName, methods := range receiverMethods {
		iface, matched := matchesInterface(methods)
		if !matched {
			continue
		}

		// Find the line where the struct type is declared.
		line := findTypeLine(fset, f, typeName)

		violations = append(violations, Violation{
			File:       filePath,
			Line:       line,
			StructName: typeName,
			Interface:  iface,
		})
	}

	return violations, nil
}

// collectReceiverMethods returns a map of receiver type name → set of method names
// for all methods declared in the file.
func collectReceiverMethods(f *ast.File) map[string]map[string]bool {
	result := make(map[string]map[string]bool)

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}

		typeName := receiverTypeName(fn.Recv.List[0].Type)
		if typeName == "" {
			continue
		}

		if result[typeName] == nil {
			result[typeName] = make(map[string]bool)
		}

		result[typeName][fn.Name.Name] = true
	}

	return result
}

// receiverTypeName extracts the base type name from a receiver expression.
func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	}

	return ""
}

// matchesInterface returns the interface name and true when the method set satisfies
// IPublicServer or IAdminServer completely.
func matchesInterface(methods map[string]bool) (string, bool) {
	if hasAll(methods, publicServerMethods) {
		return "IPublicServer", true
	}

	if hasAll(methods, adminServerMethods) {
		return "IAdminServer", true
	}

	return "", false
}

// hasAll returns true when every required name is present in methods.
func hasAll(methods map[string]bool, required []string) bool {
	for _, m := range required {
		if !methods[m] {
			return false
		}
	}

	return true
}

// findTypeLine returns the source line of the struct type declaration for typeName, or 0.
func findTypeLine(fset *token.FileSet, f *ast.File, typeName string) int {
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if ts.Name.Name == typeName {
				return fset.Position(ts.Pos()).Line
			}
		}
	}

	return 0
}
