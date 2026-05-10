// Copyright (c) 2025-2026 Justin Cranford.
// Package testmain_orchestration_policy enforces that every server and client TestMain file
// in internal/apps/{PS-ID}/ imports the canonical test_orch_integration package from the
// apps-framework service. This prevents ad-hoc server startup logic from re-appearing in
// PS-ID packages after the framework-v21 migration.
//
// Files checked:
//   - internal/apps/{PS-ID}/server/testmain_test.go (all 10 PS-IDs)
//   - internal/apps/{PS-ID}/client/testmain_test.go (where client/ exists)
//
// The linter reads only the import section of each file; it does not parse ASTs.
// The required import substring is "service/test_orch_integration".
package testmain_orchestration_policy

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	// requiredImportSubstring is the import path substring that must appear in every
	// server and client testmain_test.go to prove test_orch_integration is used.
	requiredImportSubstring = "service/test_orch_integration"

	// testmainFileName is the required TestMain filename.
	testmainFileName = "testmain_test.go"
)

// Violation records one detected policy breach.
type Violation struct {
	File   string
	Reason string
}

// Check runs the linter from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir scans PS-ID directories under rootDir for orchestration policy violations.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking that server/client TestMain files import test_orch_integration...")

	violations, err := FindViolations(rootDir)
	if err != nil {
		return fmt.Errorf("testmain-orchestration-policy: directory walk failed: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "%s: %s \u2014 must import %q\n",
				v.File, v.Reason, requiredImportSubstring)
		}

		return fmt.Errorf("found %d testmain orchestration policy violation(s)", len(violations))
	}

	logger.LogWithPrefix("testmain-orchestration-policy", "\u2705 All server/client TestMain files use test_orch_integration")

	return nil
}

// FindViolations returns all policy violations found under rootDir.
func FindViolations(rootDir string) ([]Violation, error) {
	var violations []Violation

	appsDir := filepath.Join(rootDir, "internal", "apps")

	if _, err := os.Stat(appsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(appsDir, ps.PSID)

		// Check server/testmain_test.go -- required for every PS-ID.
		serverTestMain := filepath.Join(psDir, "server", testmainFileName)

		serverViolations, err := checkTestMainFile(serverTestMain, ps.PSID, "server")
		if err != nil {
			return nil, err
		}

		violations = append(violations, serverViolations...)

		// Check client/testmain_test.go -- only if client/ directory exists.
		clientTestMain := filepath.Join(psDir, "client", testmainFileName)
		if _, err := os.Stat(clientTestMain); err == nil {
			clientViolations, err := checkTestMainFile(clientTestMain, ps.PSID, "client")
			if err != nil {
				return nil, err
			}

			violations = append(violations, clientViolations...)
		}
	}

	return violations, nil
}

// checkTestMainFile verifies that a specific testmain_test.go file exists and imports
// the required test_orch_integration package. Returns empty slice if the file passes.
func checkTestMainFile(filePath, psID, subPkg string) ([]Violation, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Violation{{
				File:   filePath,
				Reason: fmt.Sprintf("PS-ID %s %s/testmain_test.go is missing", psID, subPkg),
			}}, nil
		}

		return nil, fmt.Errorf("stat %s: %w", filePath, err)
	}

	if info.IsDir() {
		return []Violation{{
			File:   filePath,
			Reason: fmt.Sprintf("PS-ID %s %s/testmain_test.go is a directory, not a file", psID, subPkg),
		}}, nil
	}

	hasImport, err := fileContainsLine(filePath, requiredImportSubstring)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filePath, err)
	}

	if !hasImport {
		return []Violation{{
			File:   filePath,
			Reason: fmt.Sprintf("PS-ID %s %s/testmain_test.go does not import test_orch_integration", psID, subPkg),
		}}, nil
	}

	return nil, nil
}

// fileContainsLine returns true if any line in the file contains the given substring.
func fileContainsLine(filePath, substring string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("open %s: %w", filePath, err)
	}

	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), substring) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("scan %s: %w", filePath, err)
	}

	return false, nil
}

// WalkTestMainFiles walks rootDir and returns all testmain_test.go paths under internal/apps/{PS-ID}/server/
// and internal/apps/{PS-ID}/client/. Used by tests to validate file discovery.
func WalkTestMainFiles(rootDir string) ([]string, error) {
	var result []string

	appsDir := filepath.Join(rootDir, "internal", "apps")

	err := filepath.WalkDir(appsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		switch {
		case d.IsDir() && (d.Name() == cryptoutilSharedMagic.CICDExcludeDirGit || d.Name() == cryptoutilSharedMagic.CICDExcludeDirVendor):
			return filepath.SkipDir
		case !d.IsDir() && d.Name() == testmainFileName:
			relDir := filepath.Base(filepath.Dir(path))
			if relDir == "server" || relDir == "client" {
				result = append(result, path)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", appsDir, err)
	}

	return result, nil
}
