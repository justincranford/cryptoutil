// Copyright (c) 2025 Justin Cranford

// Package banned_imports checks Go files for banned cryptographic imports.
package banned_imports

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"regexp"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

// bannedImportPatterns maps banned import paths to the reason they are banned.
var bannedImportPatterns = []struct {
	pattern *regexp.Regexp
	reason  string
}{
	{regexp.MustCompile(`"golang\.org/x/crypto/argon2"`), "Argon2 is not FIPS 140-3 approved. Use PBKDF2 or HKDF."},
	{regexp.MustCompile(`"golang\.org/x/crypto/bcrypt"`), "bcrypt is not FIPS 140-3 approved. Use PBKDF2 or HKDF."},
	{regexp.MustCompile(`"golang\.org/x/crypto/scrypt"`), "scrypt is not FIPS 140-3 approved. Use PBKDF2 or HKDF."},
	{regexp.MustCompile(`"crypto/des"`), "DES/3DES is not FIPS 140-3 approved. Use AES."},
	{regexp.MustCompile(`"crypto/md5"`), "MD5 is not FIPS 140-3 approved. Use SHA-256 or SHA-512."},
	{regexp.MustCompile(`"crypto/rc4"`), "RC4 is not FIPS 140-3 approved. Use AES."},
	{regexp.MustCompile(`"math/rand"`), "math/rand is not cryptographically secure. Use crypto/rand."},
}

// Check scans Go files for banned cryptographic imports.
func Check(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Checking for banned cryptographic imports...")

	goFiles := filesByExtension["go"]
	if len(goFiles) == 0 {
		logger.Log("No Go files to check")

		return nil
	}

	// Filter out test files and self-exclusion files.
	filtered := cryptoutilCmdCicdCommon.FilterFilesForCommand(goFiles, "lint-security")

	var violations []string

	for _, filePath := range filtered {
		fileViolations, err := checkFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to check %s: %w", filePath, err)
		}

		violations = append(violations, fileViolations...)
	}

	if len(violations) > 0 {
		fmt.Fprintln(os.Stderr, "\n❌ Found banned cryptographic imports:")

		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  - %s\n", v)
		}

		return fmt.Errorf("banned-imports: found %d violations", len(violations))
	}

	logger.Log("No banned cryptographic imports found")

	return nil
}

// checkFile scans a single Go file for banned imports using the Go AST parser.
// This correctly ignores import paths that appear inside string literals or comments.
func checkFile(filePath string) ([]string, error) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var violations []string

	for _, importSpec := range f.Imports {
		importPath := importSpec.Path.Value // quoted string, e.g. `"golang.org/x/crypto/bcrypt"`
		lineNum := fset.Position(importSpec.Path.Pos()).Line

		for _, banned := range bannedImportPatterns {
			if banned.pattern.MatchString(importPath) {
				violations = append(violations, fmt.Sprintf("%s:%d: %s", filePath, lineNum, banned.reason))
			}
		}
	}

	return violations, nil
}
