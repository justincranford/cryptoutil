// Copyright (c) 2025-2026 Justin Cranford.
package gen_config_initialisms

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermOwnerReadWriteOnly)
	require.NoError(t, err)

	return path
}

func makeTestAPIDir(t *testing.T) (rootDir, apiDir string) {
	t.Helper()
	rootDir = t.TempDir()
	apiDir = filepath.Join(rootDir, "api")
	require.NoError(t, os.MkdirAll(apiDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	return rootDir, apiDir
}

const fullBaseList = "output-options:\n  name-normalizer: ToCamelCaseWithInitialisms\n  additional-initialisms:\n    - IDS\n    - JWT\n    - JWK\n    - JWE\n    - JWS\n    - OIDC\n    - SAML\n    - AES\n    - GCM\n    - CBC\n    - RSA\n    - EC\n    - HMAC\n    - SHA\n    - TLS\n    - IP\n    - AI\n    - ML\n    - KEM\n    - PEM\n    - DER\n    - DSA\n    - IKM\n"

func TestCheckMissingInitialisms_AllPresent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, genConfigServerFileName, fullBaseList)
	missing, err := checkMissingInitialisms(path)
	require.NoError(t, err)
	require.Empty(t, missing)
}

func TestCheckMissingInitialisms_MissingOne(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// All base initialisms except DSA.
	noDSA := "output-options:\n  additional-initialisms:\n    - IDS\n    - JWT\n    - JWK\n    - JWE\n    - JWS\n    - OIDC\n    - SAML\n    - AES\n    - GCM\n    - CBC\n    - RSA\n    - EC\n    - HMAC\n    - SHA\n    - TLS\n    - IP\n    - AI\n    - ML\n    - KEM\n    - PEM\n    - DER\n    - IKM\n"
	path := writeTestFile(t, dir, genConfigServerFileName, noDSA)
	missing, err := checkMissingInitialisms(path)
	require.NoError(t, err)
	require.Equal(t, []string{"DSA"}, missing)
}

func TestCheckMissingInitialisms_EmptyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, genConfigServerFileName, "")
	missing, err := checkMissingInitialisms(path)
	require.NoError(t, err)
	require.Len(t, missing, len(baseInitialisms))
}

func TestCheckMissingInitialisms_WithDomainExtras(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	content := fullBaseList + "    - CSR\n    - CA\n"
	path := writeTestFile(t, dir, genConfigServerFileName, content)
	missing, err := checkMissingInitialisms(path)
	require.NoError(t, err)
	require.Empty(t, missing)
}

func TestCheckMissingInitialisms_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := checkMissingInitialisms("/nonexistent/path/openapi-gen_config_server.yaml")
	require.Error(t, err)
}

// TestCheckMissingInitialisms_ScannerError verifies that a line exceeding bufio.MaxScanTokenSize triggers scanner.Err().
func TestCheckMissingInitialisms_ScannerError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, genConfigServerFileName)
	// A single line exceeding 64KB triggers bufio.ErrTooLong from scanner.Err().
	hugeLine := "- IDS " + strings.Repeat("x", 70000) + "\n"
	require.NoError(t, os.WriteFile(path, []byte(hugeLine), cryptoutilSharedMagic.FilePermOwnerReadWriteOnly))

	_, err := checkMissingInitialisms(path)
	require.Error(t, err)
}

func TestCheckInDir_Clean(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	writeTestFile(t, apiDir, genConfigServerFileName, fullBaseList)

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.NoError(t, err)
}

func TestCheckInDir_Violation(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	writeTestFile(t, apiDir, genConfigServerFileName, "output-options:\n  additional-initialisms:\n    - JWT\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "gen config initialisms violation")
}

func TestCheckInDir_NonServerConfigs(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	writeTestFile(t, apiDir, "openapi-gen_config_client.yaml", "output-options:\n  name-normalizer: ToCamelCaseWithInitialisms\n")
	writeTestFile(t, apiDir, "openapi-gen_config_models.yaml", "output-options:\n  skip-prune: true\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.NoError(t, err)
}

func TestCheckInDir_SkipsGitAndVendor(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	gitDir := filepath.Join(apiDir, cryptoutilSharedMagic.CICDExcludeDirGit)
	require.NoError(t, os.MkdirAll(gitDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	writeTestFile(t, gitDir, genConfigServerFileName, "")

	vendorDir := filepath.Join(apiDir, cryptoutilSharedMagic.CICDExcludeDirVendor)
	require.NoError(t, os.MkdirAll(vendorDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	writeTestFile(t, vendorDir, genConfigServerFileName, "")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.NoError(t, err)
}

func TestCheckInDir_MultipleServices(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	svc1Dir := filepath.Join(apiDir, cryptoutilSharedMagic.OTLPServiceSkeletonTemplate)
	require.NoError(t, os.MkdirAll(svc1Dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	writeTestFile(t, svc1Dir, genConfigServerFileName, fullBaseList+"    - JWKS\n")

	svc2Dir := filepath.Join(apiDir, cryptoutilSharedMagic.OTLPServiceSMKMS)
	require.NoError(t, os.MkdirAll(svc2Dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	writeTestFile(t, svc2Dir, genConfigServerFileName, "output-options:\n  additional-initialisms:\n    - JWT\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.Error(t, err)
}

func TestCheckInDir_NoAPIDir(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "api directory not found")
}

func TestCheckInDir_NoServerConfigs(t *testing.T) {
	t.Parallel()
	rootDir, apiDir := makeTestAPIDir(t)
	writeTestFile(t, apiDir, "openapi-gen_config_client.yaml", "# client config\n")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms")
	err := CheckInDir(logger, rootDir)
	require.NoError(t, err)
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	projectRoot, findErr := findProjectRoot()
	if findErr != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectRoot))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms-integration")
	err = Check(logger)
	require.NoError(t, err, "All server gen configs should contain the base initialisms list")
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}

		dir = parent
	}
}

func TestCheckInDir_OperationPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*testing.T, *genConfigInitialismsOperations, string, string)
		wantErr string
	}{
		{
			name: "empty api directory succeeds",
		},
		{
			name: "stat error is reported",
			setup: func(_ *testing.T, operations *genConfigInitialismsOperations, _, _ string) {
				operations.statFn = func(string) (os.FileInfo, error) {
					return nil, fmt.Errorf("injected api stat error")
				}
			},
			wantErr: "failed to stat api directory",
		},
		{
			name: "walk error is reported",
			setup: func(_ *testing.T, operations *genConfigInitialismsOperations, _, _ string) {
				operations.walkDirFn = func(string, fs.WalkDirFunc) error {
					return fmt.Errorf("injected walk error")
				}
			},
			wantErr: genConfigInitialismsWalkErrorPrefix,
		},
		{
			name: "walk callback entry error is reported",
			setup: func(_ *testing.T, operations *genConfigInitialismsOperations, apiDir, _ string) {
				operations.walkDirFn = func(_ string, walkFn fs.WalkDirFunc) error {
					return walkFn(apiDir, nil, fmt.Errorf("injected entry error"))
				}
			},
			wantErr: genConfigInitialismsWalkErrorPrefix,
		},
		{
			name: "file validation error is reported",
			setup: func(t *testing.T, operations *genConfigInitialismsOperations, _, apiDir string) {
				writeTestFile(t, apiDir, genConfigServerFileName, fullBaseList)

				operations.checkMissingInitialisms = func(string) ([]string, error) {
					return nil, fmt.Errorf("injected file validation error")
				}
			},
			wantErr: "failed to check",
		},
		{
			name: "relative path error is reported",
			setup: func(t *testing.T, operations *genConfigInitialismsOperations, _, apiDir string) {
				writeTestFile(t, apiDir, genConfigServerFileName, fullBaseList)

				operations.checkMissingInitialisms = func(string) ([]string, error) {
					return []string{"JWK"}, nil
				}
				operations.relFn = func(string, string) (string, error) {
					return "", fmt.Errorf("injected relative path error")
				}
			},
			wantErr: "failed to get path relative to workspace root",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir, apiDir := makeTestAPIDir(t)

			operations := productionGenConfigInitialismsOperations()
			if tc.setup != nil {
				tc.setup(t, &operations, apiDir, apiDir)
			}

			err := checkInDirWithOperations(cryptoutilCmdCicdCommon.NewLogger("test-gen-config-initialisms"), rootDir, operations)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}
