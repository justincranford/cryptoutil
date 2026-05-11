// Copyright (c) 2025-2026 Justin Cranford.

package testmain_e2e_policy

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const compliantTestMain = `//go:build e2e

package e2e_test

import (
	"os"
	"testing"

	cryptoutilTestOrchE2e "cryptoutil/internal/apps-framework/service/test_orch_e2e"
)

func TestMain(m *testing.M) {
	os.Exit(cryptoutilTestOrchE2e.SetupE2ETestMain(m, cryptoutilTestOrchE2e.E2ETestConfig{}, nil))
}
`

const legacyTestMain = `//go:build e2e

package e2e_test

import (
	"os"
	"testing"

	cryptoutilAppsFrameworkTestingE2eInfra "cryptoutil/internal/apps-framework/service/testing/e2e_infra"
)

func TestMain(m *testing.M) {
	os.Exit(cryptoutilAppsFrameworkTestingE2eInfra.SetupE2ETestMain(m, cryptoutilAppsFrameworkTestingE2eInfra.E2ETestConfig{}, nil))
}
`

func writePolicyFile(t *testing.T, root, content string) {
	t.Helper()

	targetDir := filepath.Join(root, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, "e2e")
	requireNoError(t, os.MkdirAll(targetDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	requireNoError(t, os.WriteFile(filepath.Join(targetDir, e2eTestMainFileName), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))
}

func TestFindViolationsWithReader_ReadError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	writePolicyFile(t, tempDir, compliantTestMain)

	violations, err := findViolationsWithReader(tempDir, func(string) ([]byte, error) {
		return nil, errors.New("injected read failure")
	})
	if violations != nil {
		t.Fatalf("violations must be nil on read failure")
	}

	if err == nil {
		t.Fatal("expected read failure")
	}
}

func TestFindViolationsWithReader_InvalidRootPath(t *testing.T) {
	t.Parallel()

	violations, err := findViolationsWithReader("bad\x00root", os.ReadFile)
	if violations != nil {
		t.Fatalf("violations must be nil on walk failure")
	}

	if err == nil {
		t.Fatal("expected walk failure")
	}
}

func TestFindViolationsWithDeps_WalkFnError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	requireNoError(t, os.MkdirAll(filepath.Join(tempDir, "internal", "apps"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	violations, err := findViolationsWithDeps(tempDir, os.ReadFile, func(string, fs.WalkDirFunc) error {
		return errors.New("injected walk failure")
	})
	if violations != nil {
		t.Fatal("violations must be nil on walk function failure")
	}

	if err == nil {
		t.Fatal("expected walk function failure")
	}
}

func TestFindViolationsWithDeps_WalkCallbackError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	requireNoError(t, os.MkdirAll(filepath.Join(tempDir, "internal", "apps"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	violations, err := findViolationsWithDeps(tempDir, os.ReadFile, func(_ string, walkFn fs.WalkDirFunc) error {
		return walkFn(filepath.Join(tempDir, "internal", "apps", "node"), nil, errors.New("injected callback failure"))
	})
	if violations != nil {
		t.Fatal("violations must be nil on walk callback failure")
	}

	if err == nil {
		t.Fatal("expected walk callback failure")
	}
}

func TestCheckInDirWithReader_Compliant_NoError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	writePolicyFile(t, tempDir, compliantTestMain)

	logger := cryptoutilCmdCicdCommon.NewLogger("testmain-e2e-policy-internal")

	err := checkInDirWithReader(logger, tempDir, os.ReadFile)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCheckInDirWithReader_Violations_ReturnError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	writePolicyFile(t, tempDir, legacyTestMain)

	logger := cryptoutilCmdCicdCommon.NewLogger("testmain-e2e-policy-internal")

	err := checkInDirWithReader(logger, tempDir, os.ReadFile)
	if err == nil {
		t.Fatal("expected violation error")
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}
