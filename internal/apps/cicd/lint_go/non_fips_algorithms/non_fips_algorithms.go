// Copyright (c) 2025 Justin Cranford

// Package non_fips_algorithms detects banned non-FIPS algorithms in Go code.
package non_fips_algorithms

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// Banned non-FIPS algorithms and their FIPS-approved alternatives.
var bannedAlgorithms = map[string]string{
	// Password hashing.
	"bcrypt":   "PBKDF2-HMAC-SHA256 (600k iterations per OWASP 2025)",
	"scrypt":   "PBKDF2-HMAC-SHA256 (600k iterations per OWASP 2025)",
	"argon2":   "PBKDF2-HMAC-SHA256 (600k iterations per OWASP 2025)",
	"Argon2":   "PBKDF2-HMAC-SHA256 (600k iterations per OWASP 2025)",
	"Argon2i":  "PBKDF2-HMAC-SHA256 (600k iterations per OWASP 2025)",
	"Argon2id": "PBKDF2-HMAC-SHA256 (600k iterations per OWASP 2025)",

	// Weak digests.
	"md5.New":       "SHA-256/384/512",
	"md5.Sum":       "SHA-256/384/512",
	"sha1.New":      "SHA-256/384/512",
	"sha1.Sum":      "SHA-256/384/512",
	"crypto.MD5":    "SHA-256/384/512",
	"crypto.SHA1":   "SHA-256/384/512",
	"crypto.SHA224": "SHA-256/384/512 (SHA-224 is weak)",
	"sha256.New224": "SHA-256 (SHA-224 is weak)",

	// Weak symmetric ciphers.
	"des.NewCipher":    "AES-GCM (128/192/256 bits)",
	"des.NewTripleDES": "AES-GCM (128/192/256 bits)",
	"rc4.NewCipher":    "AES-GCM (128/192/256 bits)",
	"rc2.NewCipher":    "AES-GCM (128/192/256 bits)",
	"crypto.DES":       "AES-GCM (128/192/256 bits)",
	"crypto.3DES":      "AES-GCM (128/192/256 bits)",
	"crypto.RC4":       "AES-GCM (128/192/256 bits)",

	// Weak asymmetric algorithms.
	"dsa.GenerateKey":        "RSA ≥2048 or ECDSA P-256/384/521",
	"dsa.GenerateParameters": "RSA ≥2048 or ECDSA P-256/384/521",
	"dsa.Sign":               "RSA ≥2048 or ECDSA P-256/384/521",
	"crypto.DSA":             "RSA ≥2048 or ECDSA P-256/384/521",

	// Weak elliptic curves.
	"elliptic.P224": "P-256/P-384/P-521 (P-224 is weak)",
	"secp224r1":     "P-256/P-384/P-521 (P-224 is weak)",

	// Imports to detect.
	`"crypto/md5"`:                 "crypto/sha256 or crypto/sha512",
	`"crypto/sha1"`:                "crypto/sha256 or crypto/sha512",
	`"crypto/des"`:                 "crypto/aes",
	`"crypto/rc4"`:                 "crypto/aes",
	`"crypto/dsa"`:                 "crypto/rsa or crypto/ecdsa or crypto/ed25519",
	`"golang.org/x/crypto/bcrypt"`: "golang.org/x/crypto/pbkdf2",
	`"golang.org/x/crypto/scrypt"`: "golang.org/x/crypto/pbkdf2",
	`"golang.org/x/crypto/argon2"`: "golang.org/x/crypto/pbkdf2",
}

// Check detects banned non-FIPS algorithms in Go code.
// Returns error if violations found.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Checking for non-FIPS algorithms...")

	// Find all .go files (exclude vendor, test-output).
	goFiles, err := FindGoFiles()
	if err != nil {
		return fmt.Errorf("failed to find Go files: %w", err)
	}

	violations := make(map[string][]string) // file -> list of issues

	for _, filePath := range goFiles {
		fileViolations := CheckFileForNonFIPS(filePath)
		if len(fileViolations) > 0 {
			violations[filePath] = fileViolations
		}
	}

	if len(violations) > 0 {
		PrintNonFIPSViolations(violations)

		return fmt.Errorf("found non-FIPS algorithm violations in %d files", len(violations))
	}

	logger.Log("✅ Non-FIPS algorithm check passed")

	return nil
}

// FindGoFiles finds all .go files in the project (exclude vendor, test-output, .git, test files, nonfips.go).
func FindGoFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Skip directories to exclude.
		if info.IsDir() {
			switch info.Name() {
			case "vendor", "test-output", ".git", "node_modules":
				return filepath.SkipDir
			}

			return nil
		}

		// Exclude test files (intentionally test banned algorithms).
		// Exclude nonfips.go (contains bannedAlgorithms map with all keywords).
		// Exclude password and pbkdf2 packages (contain bcrypt for backward compatibility).
		if filepath.Ext(path) == ".go" &&
			!strings.HasSuffix(path, "_test.go") &&
			!strings.HasSuffix(path, "nonfips.go") &&
			!strings.Contains(path, filepath.Join("internal", "shared", "crypto", "password")) &&
			!strings.Contains(path, filepath.Join("internal", "shared", "crypto", "pbkdf2")) {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	return files, nil
}

// CheckFileForNonFIPS checks a single Go file for non-FIPS algorithm usage.
// Returns list of violations (empty if clean).
func CheckFileForNonFIPS(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Check for each banned algorithm.
	for banned, alternative := range bannedAlgorithms {
		// Use case-sensitive exact match for function calls and imports.
		pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(banned) + `\b`)

		if pattern.MatchString(contentStr) {
			// Find line numbers.
			for i, line := range lines {
				if pattern.MatchString(line) {
					// Skip if line contains nolint comment (gosec handles some of these).
					if strings.Contains(line, "nolint") {
						continue
					}

					// Check if the line is covered by a nolint comment on the import line.
					// For imports, check if there's a nolint comment anywhere mentioning this algorithm.
					skipViolation := false

					for _, checkLine := range lines {
						// If there's a nolint comment that mentions the banned algorithm or gosec.
						if strings.Contains(checkLine, "nolint") &&
							(strings.Contains(strings.ToLower(checkLine), strings.ToLower(banned)) ||
								strings.Contains(checkLine, "gosec")) {
							skipViolation = true

							break
						}
					}

					if skipViolation {
						continue
					}

					issues = append(issues, fmt.Sprintf(
						"Line %d: Found '%s' (non-FIPS) - use %s instead",
						i+1, banned, alternative,
					))
				}
			}
		}
	}

	return issues
}

// PrintNonFIPSViolations prints formatted non-FIPS violations to stderr.
func PrintNonFIPSViolations(violations map[string][]string) {
	totalIssues := 0
	for _, issues := range violations {
		totalIssues += len(issues)
	}

	fmt.Fprintf(os.Stderr, "\n❌ Found %d non-FIPS algorithm violations:\n\n", totalIssues)

	for filePath, issues := range violations {
		fmt.Fprintf(os.Stderr, "%s:\n", filePath)

		for _, issue := range issues {
			fmt.Fprintf(os.Stderr, "  - %s\n", issue)
		}

		fmt.Fprintln(os.Stderr)
	}

	fmt.Fprintln(os.Stderr, "Why this matters:")
	fmt.Fprintln(os.Stderr, "  - FIPS 140-3 compliance is MANDATORY for cryptoutil")
	fmt.Fprintln(os.Stderr, "  - Non-FIPS algorithms are BANNED (no exceptions)")
	fmt.Fprintln(os.Stderr, "  - See .github/instructions/02-07.cryptography.instructions.md for approved algorithms")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "FIPS-Approved Alternatives:")
	fmt.Fprintln(os.Stderr, "  - Password hashing: PBKDF2-HMAC-SHA256 (600k iterations per OWASP 2025)")
	fmt.Fprintln(os.Stderr, "  - Digests: SHA-256/384/512, HMAC-SHA256/384/512")
	fmt.Fprintln(os.Stderr, "  - Symmetric: AES ≥128 (GCM, CBC+HMAC)")
	fmt.Fprintln(os.Stderr, "  - Asymmetric: RSA ≥2048, ECDSA P-256/384/521, EdDSA Ed25519/448")
	fmt.Fprintln(os.Stderr, "  - KDF: PBKDF2-HMAC-SHA256/384/512, HKDF-SHA256/384/512")
	fmt.Fprintln(os.Stderr, "\nPlease replace banned algorithms with FIPS-approved alternatives.")
}
