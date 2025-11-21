package go_fix_staticcheck_error_strings

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

// commonAcronyms is a map of uppercase acronyms that should be preserved at the start of error strings.
var commonAcronyms = map[string]bool{
	"HTTP":    true,
	"HTTPS":   true,
	"URL":     true,
	"URI":     true,
	"JSON":    true,
	"XML":     true,
	"API":     true,
	"ID":      true,
	"UUID":    true,
	"SQL":     true,
	"DB":      true,
	"TLS":     true,
	"SSL":     true,
	"TCP":     true,
	"UDP":     true,
	"IP":      true,
	"DNS":     true,
	"HTML":    true,
	"CSS":     true,
	"JS":      true,
	"TS":      true,
	"CSV":     true,
	"YAML":    true,
	"TOML":    true,
	"JWT":     true,
	"OAuth":   true,
	"OIDC":    true,
	"SAML":    true,
	"LDAP":    true,
	"SMTP":    true,
	"IMAP":    true,
	"POP3":    true,
	"FTP":     true,
	"SFTP":    true,
	"SSH":     true,
	"RSA":     true,
	"AES":     true,
	"HMAC":    true,
	"SHA":     true,
	"MD5":     true,
	"ECDSA":   true,
	"ECDH":    true,
	"JWK":     true,
	"JWS":     true,
	"JWE":     true,
	"JWA":     true,
	"PEM":     true,
	"DER":     true,
	"ASN":     true,
	"PKCS":    true,
	"PKIX":    true,
	"X509":    true,
	"CSR":     true,
	"CRL":     true,
	"OCSP":    true,
	"CA":      true,
	"CN":      true,
	"OU":      true,
	"O":       true,
	"L":       true,
	"ST":      true,
	"C":       true,
	"EC":      true,
	"ED25519": true,
	"ED448":   true,
}

// errorStringPattern matches error string declarations that should be checked.
var errorStringPattern = regexp.MustCompile(`(?m)^\s*(?:var|const)\s+\w+\s*=\s*(?:errors\.New|fmt\.Errorf)\s*\(\s*"([A-Z][^"]*)"`)

// Fix analyzes Go files and fixes error strings that start with uppercase letters (except acronyms).
// Returns the number of files processed, modified, and issues fixed.
func Fix(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) (int, int, int, error) {
	var processed, modified, issuesFixed int

	if err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, non-Go files, and test files.
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		changed, fixes, err := fixStaticcheckInFile(logger, path)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", path, err)
		}

		processed++
		if changed {
			modified++
			issuesFixed += fixes
			logger.Log(fmt.Sprintf("Fixed %d error strings in: %s", fixes, path))
		}

		return nil
	}); err != nil {
		return processed, modified, issuesFixed, fmt.Errorf("failed to walk directory: %w", err)
	}

	return processed, modified, issuesFixed, nil
}

// fixStaticcheckInFile processes a single file and fixes error strings that start with uppercase letters.
// Returns whether the file was changed and the number of fixes applied.
func fixStaticcheckInFile(logger *cryptoutilCmdCicdCommon.Logger, filePath string) (bool, int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read file: %w", err)
	}

	original := string(content)
	fixed := original
	changed := false
	fixCount := 0

	// Find all error string declarations.
	matches := errorStringPattern.FindAllStringSubmatch(original, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		fullMatch := match[0]
		errorMsg := match[1]

		// Extract the first word to check if it's an acronym.
		firstWord := strings.Fields(errorMsg)[0]
		firstWord = strings.TrimRight(firstWord, ":")

		// If it's a known acronym, skip it.
		if commonAcronyms[firstWord] {
			continue
		}

		// Lowercase the first letter.
		fixedMsg := strings.ToLower(errorMsg[:1]) + errorMsg[1:]
		fixedMatch := strings.Replace(fullMatch, `"`+errorMsg+`"`, `"`+fixedMsg+`"`, 1)

		// Replace in the file content.
		fixed = strings.Replace(fixed, fullMatch, fixedMatch, 1)
		changed = true
		fixCount++
	}

	if changed {
		const filePermissions = 0o600
		if err := os.WriteFile(filePath, []byte(fixed), filePermissions); err != nil {
			return false, fixCount, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return changed, fixCount, nil
}
