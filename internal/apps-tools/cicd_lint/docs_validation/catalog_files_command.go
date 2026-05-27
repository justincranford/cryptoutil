// Copyright (c) 2025-2026 Justin Cranford.
// This file contains the format helpers and command entry points for the catalog linters.
// Core extraction and check logic lives in catalog_files.go.
package docs_validation

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// formatCatalogFilesResults formats the catalog-files check result as human-readable text.
func formatCatalogFilesResults(result *CatalogFilesResult) string {
	var sb strings.Builder

	sb.WriteString("=== ENG-HANDBOOK.md Catalog Files Check ===\n\n")

	if len(result.Violations) == 0 {
		fmt.Fprintf(&sb, "All %d catalog entries match their files on disk.\n", result.Checked)

		return sb.String()
	}

	sort.Slice(result.Violations, func(i, j int) bool {
		if result.Violations[i].File != result.Violations[j].File {
			return result.Violations[i].File < result.Violations[j].File
		}

		return result.Violations[i].Field < result.Violations[j].Field
	})

	fmt.Fprintf(&sb, "VIOLATIONS (%d):\n", len(result.Violations))

	for _, v := range result.Violations {
		fmt.Fprintf(&sb, "  [%s] %s: %s\n", v.Field, v.File, v.Detail)
	}

	return sb.String()
}

// formatCatalogPropagationResults formats the catalog-propagation check result as human-readable text.
func formatCatalogPropagationResults(result *CatalogPropagationResult) string {
	var sb strings.Builder

	sb.WriteString("=== ENG-HANDBOOK.md Catalog Propagation Check ===\n\n")

	if len(result.Violations) == 0 {
		fmt.Fprintf(&sb, "All %d catalogued chunk(s) found with matching content.\n", result.Checked)

		return sb.String()
	}

	sort.Slice(result.Violations, func(i, j int) bool {
		if result.Violations[i].File != result.Violations[j].File {
			return result.Violations[i].File < result.Violations[j].File
		}

		return result.Violations[i].Field < result.Violations[j].Field
	})

	fmt.Fprintf(&sb, "VIOLATIONS (%d):\n", len(result.Violations))

	for _, v := range result.Violations {
		fmt.Fprintf(&sb, "  [%s] %s: %s\n", v.Field, v.File, v.Detail)
	}

	return sb.String()
}

// CatalogFilesCommand is the entry point for the lint-catalog-files linter.
// It verifies that every @file-catalog / @file-catalog-pair entry in ENG-HANDBOOK.md
// matches the corresponding file(s) on disk.
// Returns 0 on success, 1 on violations.
func CatalogFilesCommand(stdout, stderr io.Writer) int {
	return catalogFilesCommand(stdout, stderr, findProjectRoot)
}

func catalogFilesCommand(stdout, stderr io.Writer, rootFn func() (string, error)) int {
	rootDir, err := rootFn()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "catalog-files: error: %v\n", err)

		return 1
	}

	result, err := CheckCatalogFiles(rootDir, rootedReadFile(rootDir))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "catalog-files: error: %v\n", err)

		return 1
	}

	_, _ = fmt.Fprint(stdout, formatCatalogFilesResults(result))

	if len(result.Violations) > 0 {
		_, _ = fmt.Fprintf(stderr, "catalog-files: %d violation(s) found\n", len(result.Violations))

		return 1
	}

	return 0
}

// CatalogPropagationCommand is the entry point for the lint-catalog-propagation linter.
// It verifies that every @appendix-propagate chunk targeting a catalogued file has a
// matching @source block inside that catalog entry's body.
// Returns 0 on success, 1 on violations.
func CatalogPropagationCommand(stdout, stderr io.Writer) int {
	return catalogPropagationCommand(stdout, stderr, findProjectRoot)
}

func catalogPropagationCommand(stdout, stderr io.Writer, rootFn func() (string, error)) int {
	rootDir, err := rootFn()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "catalog-propagation: error: %v\n", err)

		return 1
	}

	result, err := CheckCatalogPropagation(rootDir, rootedReadFile(rootDir))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "catalog-propagation: error: %v\n", err)

		return 1
	}

	_, _ = fmt.Fprint(stdout, formatCatalogPropagationResults(result))

	if len(result.Violations) > 0 {
		_, _ = fmt.Fprintf(stderr, "catalog-propagation: %d violation(s) found\n", len(result.Violations))

		return 1
	}

	return 0
}
