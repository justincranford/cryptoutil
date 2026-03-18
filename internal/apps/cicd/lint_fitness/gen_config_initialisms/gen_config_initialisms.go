// Copyright (c) 2025 Justin Cranford

// Package gen_config_initialisms validates that all openapi-gen_config_server.yaml files
// contain the full canonical base initialisms list defined in ARCHITECTURE.md Section 8.1.4.
package gen_config_initialisms

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// baseInitialisms is the canonical list from ARCHITECTURE.md Section 8.1.4.
// Every openapi-gen_config_server.yaml MUST include all of these.
var baseInitialisms = []string{
	"IDS",
	"JWT",
	"JWK",
	string(cryptoutilSharedMagic.SessionAlgorithmJWE),
	string(cryptoutilSharedMagic.SessionAlgorithmJWS),
	"OIDC",
	"SAML",
	"AES",
	"GCM",
	"CBC",
	cryptoutilSharedMagic.KeyTypeRSA,
	"EC",
	"HMAC",
	"SHA",
	"TLS",
	"IP",
	"AI",
	"ML",
	"KEM",
	"PEM",
	"DER",
	"DSA",
	"IKM",
}

// genConfigServerFileName is the target file name for server gen configs.
const genConfigServerFileName = "openapi-gen_config_server.yaml"

// Test seams: replaceable in tests for error path coverage.
var (
	osStatFunc                  = os.Stat
	filepathWalkDir             = filepath.WalkDir
	checkMissingInitialismsFunc = checkMissingInitialisms
)

// Check validates gen config initialisms from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates gen config server files under rootDir/api/.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	apiDir := filepath.Join(rootDir, "api")

	if _, statErr := osStatFunc(apiDir); os.IsNotExist(statErr) {
		return fmt.Errorf("api directory not found at %s", apiDir)
	}

	var violations []string

	walkErr := filepathWalkDir(apiDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			switch d.Name() {
			case cryptoutilSharedMagic.CICDExcludeDirGit, cryptoutilSharedMagic.CICDExcludeDirVendor:
				return filepath.SkipDir
			}

			return nil
		}

		if d.Name() != genConfigServerFileName {
			return nil
		}

		missing, checkErr := checkMissingInitialismsFunc(path)
		if checkErr != nil {
			return fmt.Errorf("failed to check %s: %w", path, checkErr)
		}

		for _, item := range missing {
			rel, _ := filepath.Rel(rootDir, path)
			violations = append(violations, fmt.Sprintf("%s: missing base initialism %q (see ARCHITECTURE.md Section 8.1.4)", rel, item))
		}

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("failed to walk api directory: %w", walkErr)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(os.Stderr, v)
		}

		return fmt.Errorf("found %d gen config initialisms violation(s)", len(violations))
	}

	logger.Log("gen-config-initialisms: all server gen configs contain the base initialisms list")

	return nil
}

// checkMissingInitialisms returns base initialisms absent from the given YAML file.
func checkMissingInitialisms(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filePath, err)
	}

	defer func() { _ = f.Close() }()

	present := make(map[string]bool, len(baseInitialisms))

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "- ") {
			continue
		}

		// Extract the initialism: "- IDS    # ..." -> "IDS"
		word := strings.Fields(strings.TrimPrefix(line, "- "))[0]
		present[word] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	var missing []string

	for _, initialism := range baseInitialisms {
		if !present[initialism] {
			missing = append(missing, initialism)
		}
	}

	return missing, nil
}
