// Copyright (c) 2025 Justin Cranford

package format_go

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

const (
	// utcSuffixLength is the length of the string ".UTC()" that we append.
	utcSuffixLength = 6

	// timePkgName is the package name for time package.
	timePkgName = "time"

	// nowFuncName is the function name for time.Now().
	nowFuncName = "Now"
)

// enforceTimeNowUTC enforces time.Now().UTC() standardization across all Go files.
// Replaces all time.Now() calls with time.Now().UTC() to prevent SQLite/GORM timezone issues.
//
// CRITICAL SELF-MODIFICATION PREVENTION:
// This file and its tests MUST use exclusion patterns to avoid self-modification.
// The exclusion pattern "format_go" in GetGoFiles() prevents this file from being processed.
//
// Context: LLM agents repeatedly use time.Now() without .UTC() despite copilot instructions.
// Non-UTC timestamps cause test failures in non-UTC timezones (PST/EST).
// Automated enforcement prevents this recurring mistake.
//
// Returns an error if any files were modified (to indicate changes were made).
func enforceTimeNowUTC(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Enforcing time.Now().UTC() standardization in Go files...")

	// Get only Go files from the map.
	goFiles := filterGoFiles(filesByExtension)

	if len(goFiles) == 0 {
		logger.Log("time.Now().UTC() enforcement completed (no Go files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d Go files to process", len(goFiles)))

	// Process each file.
	filesModified := 0
	totalReplacements := 0

	for _, filePath := range goFiles {
		replacements, err := processGoFileForTimeNowUTC(filePath)
		if err != nil {
			logger.Log(fmt.Sprintf("Error processing %s: %v", filePath, err))

			continue
		}

		if replacements > 0 {
			filesModified++
			totalReplacements += replacements
			logger.Log(fmt.Sprintf("Modified %s: %d replacements", filePath, replacements))
		}
	}

	// Summary.
	fmt.Fprintf(os.Stderr, "\n=== TIME.NOW().UTC() ENFORCEMENT SUMMARY ===\n")
	fmt.Fprintf(os.Stderr, "Files processed: %d\n", len(goFiles))
	fmt.Fprintf(os.Stderr, "Files modified: %d\n", filesModified)
	fmt.Fprintf(os.Stderr, "Total replacements: %d\n", totalReplacements)

	if filesModified > 0 {
		fmt.Fprintln(os.Stderr, "\n Successfully applied time.Now().UTC() fixes")
		fmt.Fprintln(os.Stderr, "Please review and commit the changes")

		return fmt.Errorf("modified %d files with %d total replacements", filesModified, totalReplacements)
	}

	fmt.Fprintln(os.Stderr, "\n All Go files already use time.Now().UTC()")

	logger.Log("time.Now().UTC() enforcement completed")

	return nil
}

// processGoFileForTimeNowUTC applies time.Now().UTC() fixes to a single file using AST traversal.
// Returns the number of replacements made and any error encountered.
func processGoFileForTimeNowUTC(filePath string) (int, error) {
	// DEFENSIVE CHECK: Never process format_go package source files
	absPath, pathErr := filepath.Abs(filePath)
	if pathErr == nil {
		if strings.Contains(absPath, filepath.Join("internal", "cmd", "cicd", "format_go")) &&
			!strings.Contains(absPath, filepath.Join("R:", "temp")) && // Not tmpDir
			!strings.Contains(absPath, filepath.Join("C:", "temp")) { // Not tmpDir
			return 0, nil // Skip self-modification silently
		}
	}

	// Read the file.
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)

	// Parse the Go source file.
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filePath, originalContent, parser.ParseComments)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file: %w", err)
	}

	// Track position adjustments for multiple replacements.
	var adjustments []struct {
		pos    int
		oldLen int
		newLen int
	}

	// Track time.Now() calls that are already wrapped in .UTC().
	alreadyWrapped := make(map[*ast.CallExpr]bool)

	// First pass: Find all time.Now().UTC() patterns.
	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Check for .UTC() selector.
		if selExpr.Sel.Name != "UTC" {
			return true
		}

		// Check if X is a CallExpr for time.Now().
		innerCall, ok := selExpr.X.(*ast.CallExpr)
		if !ok {
			return true
		}

		innerSel, ok := innerCall.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if innerSel.Sel.Name != nowFuncName {
			return true
		}

		ident, ok := innerSel.X.(*ast.Ident)
		if !ok || ident.Name != timePkgName {
			return true
		}

		// Mark this time.Now() call as already wrapped.
		alreadyWrapped[innerCall] = true

		return true
	})

	// Second pass: Find time.Now() calls that need .UTC() added.
	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if this is a selector expression (package.Function).
		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Check if selector is "Now".
		if selExpr.Sel.Name != nowFuncName {
			return true
		}

		// Check if package is "time".
		ident, ok := selExpr.X.(*ast.Ident)
		if !ok || ident.Name != timePkgName {
			return true
		}

		// Skip if already wrapped in .UTC().
		if alreadyWrapped[callExpr] {
			return true
		}

		// Record the replacement position.
		callStart := fset.Position(callExpr.Pos()).Offset
		callLength := fset.Position(callExpr.End()).Offset - callStart

		adjustments = append(adjustments, struct {
			pos    int
			oldLen int
			newLen int
		}{
			pos:    callStart,
			oldLen: callLength,
			newLen: callLength + utcSuffixLength,
		})

		return true
	})

	replacements := len(adjustments)

	// Apply replacements in reverse order to maintain position accuracy.
	if replacements > 0 {
		// Build new content with replacements.
		result := originalContent

		// Apply from end to beginning to avoid position shifts.
		for i := len(adjustments) - 1; i >= 0; i-- {
			adj := adjustments[i]
			before := result[:adj.pos+adj.oldLen]
			after := result[adj.pos+adj.oldLen:]
			result = before + ".UTC()" + after
		}

		err = cryptoutilSharedUtilFiles.WriteFile(filePath, result, cryptoutilSharedMagic.FilePermissionsDefault)
		if err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return replacements, nil
}
