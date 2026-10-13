// Copyright (c) 2025-2026 Justin Cranford.
// Package gen_config_initialisms validates that all openapi-gen_config_server.yaml files
// contain the full canonical base initialisms list defined in ENG-HANDBOOK.md Section 8.1.4.
package gen_config_initialisms

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// baseInitialisms is the canonical list from ENG-HANDBOOK.md Section 8.1.4.
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
const genConfigInitialismsWalkErrorPrefix = "failed to walk api directory"

const genConfigServerFileName = "openapi-gen_config_server.yaml"

type genConfigInitialismsOperations struct {
	statFn                  func(string) (os.FileInfo, error)
	walkDirFn               func(string, fs.WalkDirFunc) error
	checkMissingInitialisms func(string) ([]string, error)
	relFn                   func(string, string) (string, error)
}

func productionGenConfigInitialismsOperations() genConfigInitialismsOperations {
	return genConfigInitialismsOperations{
		statFn:                  os.Stat,
		walkDirFn:               filepath.WalkDir,
		checkMissingInitialisms: checkMissingInitialisms,
		relFn:                   filepath.Rel,
	}
}

// Check validates gen config initialisms from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates gen config server files under rootDir/api/.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithOperations(logger, rootDir, productionGenConfigInitialismsOperations())
}

func checkInDirWithOperations(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, operations genConfigInitialismsOperations) error {
	apiDir := filepath.Join(rootDir, "api")

	if _, statErr := operations.statFn(apiDir); statErr != nil {
		if os.IsNotExist(statErr) {
			return fmt.Errorf("api directory not found at %s", apiDir)
		}

		return fmt.Errorf("failed to stat api directory %s: %w", apiDir, statErr)
	}

	var violations []string

	walkErr := operations.walkDirFn(apiDir, func(path string, d fs.DirEntry, err error) error {
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

		missing, checkErr := operations.checkMissingInitialisms(path)
		if checkErr != nil {
			return fmt.Errorf("failed to check %s: %w", path, checkErr)
		}

		for _, item := range missing {
			rel, relErr := operations.relFn(rootDir, path)
			if relErr != nil {
				return fmt.Errorf("failed to get path relative to workspace root: %w", relErr)
			}

			violations = append(violations, fmt.Sprintf("%s: missing base initialism %q (see ENG-HANDBOOK.md Section 8.1.4)", rel, item))
		}

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("%s: %w", genConfigInitialismsWalkErrorPrefix, walkErr)
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
