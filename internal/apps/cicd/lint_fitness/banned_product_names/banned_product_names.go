// Copyright (c) 2025 Justin Cranford

// Package banned_product_names enforces that banned legacy product/service names
// do not appear in any tracked file. This prevents naming regression drift — for
// example, old "Cipher IM" product names re-appearing after the SM rename.
package banned_product_names

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// bannedPhrases is the canonical list of banned legacy product/service name phrases.
// These are exact-match strings. The substring "cipher" alone is NOT banned — only
// these exact phrases (e.g., "cipher.Block" and "ciphertext" remain permitted).
var bannedPhrases = []string{
	"Cipher IM",
	"cipher-im",
	"cipher_im",
	"CipherIM",
	"cryptoutilCmdCipher",
}

// scannedExtensions lists the file extensions checked for banned phrases.
var scannedExtensions = []string{".go", ".yml", ".yaml", ".sql", ".md"}

// excludedDirs is the list of directory names that are skipped during scanning.
// Planning docs (docs/) may reference banned phrases when documenting the migration;
// this directory is excluded to avoid false positives.
// The banned_product_names/ directory itself defines the banned phrase constants
// and is excluded to prevent self-referential false positives.
// The test-output/ directory contains historical session artifacts and is excluded.
var excludedDirs = []string{
	cryptoutilSharedMagic.CICDExcludeDirGit,
	cryptoutilSharedMagic.CICDExcludeDirVendor,
	cryptoutilSharedMagic.CICDExcludeDirDocs,
	cryptoutilSharedMagic.CICDExcludeDirTestOutput,
	cryptoutilSharedMagic.CICDExcludeDirBannedProductNamesCheck,
}

// Violation holds information about a banned phrase found in a file.
type Violation struct {
	File    string
	Line    int
	Phrase  string
	Content string
}

// Check runs the banned-product-names check from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir scans rootDir for banned product/service name phrases.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking for banned product/service name phrases...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to scan for banned phrases: %w", err)
	}

	if len(violations) > 0 {
		printViolations(violations)

		return fmt.Errorf("found %d banned product/service name violations", len(violations))
	}

	logger.Log("banned-product-names: no banned phrases found")

	return nil
}

// FindViolationsInDir walks rootDir and collects all banned phrase occurrences.
func FindViolationsInDir(rootDir string) ([]Violation, error) {
	var violations []Violation

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			for _, excluded := range excludedDirs {
				if d.Name() == excluded {
					return filepath.SkipDir
				}
			}

			return nil
		}

		// Skip Go test files: they may reference banned phrases as negative test data
		// (e.g., testing that a validator rejects the old name), not as production drift.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		if !isScannedExtension(path) {
			return nil
		}

		fileViolations, err := FindViolationsInFile(path)
		if err != nil {
			return err
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", rootDir, err)
	}

	return violations, nil
}

// FindViolationsInFile scans a single file for banned product/service phrases.
func FindViolationsInFile(filePath string) ([]Violation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	defer func() { _ = file.Close() }()

	var violations []Violation

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, phrase := range bannedPhrases {
			if strings.Contains(line, phrase) {
				violations = append(violations, Violation{
					File:    filePath,
					Line:    lineNum,
					Phrase:  phrase,
					Content: strings.TrimSpace(line),
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return violations, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return violations, nil
}

// isScannedExtension reports whether path ends with one of the scanned file extensions.
func isScannedExtension(path string) bool {
	for _, ext := range scannedExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}

// printViolations prints all banned phrase violations to stderr.
func printViolations(violations []Violation) {
	fmt.Fprintf(os.Stderr, "\n❌ Found %d banned product/service name violations:\n", len(violations))

	for _, v := range violations {
		fmt.Fprintf(os.Stderr, "  %s:%d: banned phrase %q\n", v.File, v.Line, v.Phrase)
		fmt.Fprintf(os.Stderr, "    > %s\n", v.Content)
	}

	fmt.Fprintln(os.Stderr)
}
