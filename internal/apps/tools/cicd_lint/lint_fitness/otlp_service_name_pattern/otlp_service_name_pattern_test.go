// Copyright (c) 2025 Justin Cranford

package otlp_service_name_pattern

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheck_PassesOnProjectRoot(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRoot(t)

	logger := cryptoutilCmdCicdCommon.NewLogger("test-otlp-service-name-pattern")

	err := CheckInDir(logger, projectRoot)
	require.NoError(t, err, "Project should pass OTLP service name pattern check")
}

func TestCheckInDir_NoConfigsDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-otlp-no-configs-dir")

	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err, "Missing configs dir should be skipped, not fail")
}

func TestCheckOTLPServiceValue_CorrectNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		psID           string
		expectedSuffix string
		otlpService    string
	}{
		{
			name:           "sqlite config correct",
			psID:           cryptoutilSharedMagic.OTLPServiceSMIM,
			expectedSuffix: "-sqlite-1",
			otlpService:    cryptoutilSharedMagic.OTLPServiceSMIM + "-sqlite-1",
		},
		{
			name:           "postgres-1 config correct",
			psID:           cryptoutilSharedMagic.OTLPServiceSMKMS,
			expectedSuffix: "-postgres-1",
			otlpService:    cryptoutilSharedMagic.OTLPServiceSMKMS + "-postgres-1",
		},
		{
			name:           "postgres-2 config correct",
			psID:           cryptoutilSharedMagic.OTLPServiceJoseJA,
			expectedSuffix: "-postgres-2",
			otlpService:    cryptoutilSharedMagic.OTLPServiceJoseJA + "-postgres-2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config-test.yml")

			content := "otlp-service: \"" + tc.otlpService + "\"\n"
			require.NoError(t, os.WriteFile(configPath, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

			violations := checkOTLPServiceValue(configPath, tc.psID, tc.expectedSuffix, tmpDir)
			require.Empty(t, violations, "Expected no violations for correct OTLP service name")
		})
	}
}

func TestCheckOTLPServiceValue_IncorrectNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		psID           string
		expectedSuffix string
		otlpService    string
		wantErrContain string
	}{
		{
			name:           "legacy pg abbreviation",
			psID:           cryptoutilSharedMagic.OTLPServiceSMKMS,
			expectedSuffix: "-postgres-1",
			otlpService:    cryptoutilSharedMagic.OTLPServiceSMKMS + "-pg-1",
			wantErrContain: `got "sm-kms-pg-1", want "sm-kms-postgres-1"`,
		},
		{
			name:           "missing trailing number",
			psID:           cryptoutilSharedMagic.OTLPServiceSMIM,
			expectedSuffix: "-sqlite-1",
			otlpService:    cryptoutilSharedMagic.OTLPServiceSMIM + "-sqlite",
			wantErrContain: `got "sm-im-sqlite", want "sm-im-sqlite-1"`,
		},
		{
			name:           "wrong ps-id",
			psID:           cryptoutilSharedMagic.OTLPServiceSMIM,
			expectedSuffix: "-postgres-1",
			otlpService:    "cipher-im-postgres-1",
			wantErrContain: `got "cipher-im-postgres-1", want "sm-im-postgres-1"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config-test.yml")

			content := "otlp-service: \"" + tc.otlpService + "\"\n"
			require.NoError(t, os.WriteFile(configPath, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

			violations := checkOTLPServiceValue(configPath, tc.psID, tc.expectedSuffix, tmpDir)
			require.NotEmpty(t, violations, "Expected violation for incorrect OTLP service name")
			require.Contains(t, violations[0], tc.wantErrContain)
		})
	}
}

func TestCheckInDir_NoOTLPServiceKey(t *testing.T) {
	t.Parallel()

	// A config file without an otlp-service key should not be flagged.
	tmpDir := t.TempDir()
	productDir := filepath.Join(tmpDir, "configs", cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.IMServiceName)
	require.NoError(t, os.MkdirAll(productDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	configContent := "bind-public-port: 8080\n"
	require.NoError(t, os.WriteFile(filepath.Join(productDir, "config-sqlite.yml"), []byte(configContent), cryptoutilSharedMagic.FilePermissionsDefault))

	logger := cryptoutilCmdCicdCommon.NewLogger("test-otlp-no-key")

	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err, "Config without otlp-service key should not fail")
}

func TestCheckInDir_ViolationReported(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	productDir := filepath.Join(tmpDir, "configs", cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.KMSServiceName)
	require.NoError(t, os.MkdirAll(productDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Use old legacy naming.
	configContent := "otlp-service: \"" + cryptoutilSharedMagic.OTLPServiceSMKMS + "-pg-1\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(productDir, "config-pg-1.yml"), []byte(configContent), cryptoutilSharedMagic.FilePermissionsDefault))

	logger := cryptoutilCmdCicdCommon.NewLogger("test-otlp-violation")

	err := CheckInDir(logger, tmpDir)
	require.Error(t, err, "Legacy pg naming should be reported as violation")
	require.Contains(t, err.Error(), "OTLP service name violations")
}

// findProjectRoot locates the project root by finding go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("Skipping — cannot find project root (no go.mod)")
		}

		dir = parent
	}
}
