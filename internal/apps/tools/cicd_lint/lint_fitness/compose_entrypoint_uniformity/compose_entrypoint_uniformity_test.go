// Copyright (c) 2025 Justin Cranford

package compose_entrypoint_uniformity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func newLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

// findProjectRoot walks up directories until it finds go.mod.
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
			t.Skip("skipping: cannot find project root (no go.mod)")
		}

		dir = parent
	}
}

// buildComposeBlock returns a YAML services block with the 4 canonical app
// service entries for the given PS-ID, all using the expected command arrays.
func buildComposeBlock(psID string) string {
	var sb strings.Builder

	for _, v := range orderedVariants {
		svcName := lintFitnessRegistry.ComposeServiceName(psID, v)
		cmd := expectedCommand(psID, v)

		// Build compact JSON-style array literal for the YAML value.
		quoted := make([]string, len(cmd))
		for i, arg := range cmd {
			quoted[i] = fmt.Sprintf("%q", arg)
		}

		sb.WriteString(fmt.Sprintf("  %s:\n    command: [%s]\n", svcName, strings.Join(quoted, ", ")))
	}

	return sb.String()
}

// writeComposeYML creates deployments/{psID}/compose.yml under tmpDir with the given content.
func writeComposeYML(t *testing.T, tmpDir, psID, content string) {
	t.Helper()

	dir := filepath.Join(tmpDir, "deployments", psID)
	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte("services:\n"+content), cryptoutilSharedMagic.FilePermissions))
}

// setupAllComposeFiles creates correct compose files for all 10 PS-IDs in tmpDir.
func setupAllComposeFiles(t *testing.T, tmpDir string) {
	t.Helper()

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		writeComposeYML(t, tmpDir, ps.PSID, buildComposeBlock(ps.PSID))
	}
}

// TestCheck_RealWorkspace verifies the linter passes on the actual project workspace.
func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	err := CheckInDir(newLogger(), root)
	require.NoError(t, err)
}

// Sequential: changes process working directory (os.Chdir is global process state).
func TestCheck_DelegatesToCheckInDir(t *testing.T) {
	root := findProjectRoot(t)

	orig, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() { _ = os.Chdir(orig) }()

	err = Check(newLogger())
	require.NoError(t, err, "Check() should delegate to CheckInDir and pass")
}

func TestCheckInDir_AllCorrect(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupAllComposeFiles(t, tmpDir)

	err := CheckInDir(newLogger(), tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingComposeFile(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSMIM

	tmpDir := t.TempDir()
	setupAllComposeFiles(t, tmpDir)

	require.NoError(t, os.Remove(filepath.Join(tmpDir, "deployments", psID, "compose.yml")))

	err := CheckInDir(newLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), psID)
	assert.Contains(t, err.Error(), "cannot read")
}

func TestCheckInDir_InvalidYAML(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSMIM

	tmpDir := t.TempDir()
	setupAllComposeFiles(t, tmpDir)

	dir := filepath.Join(tmpDir, "deployments", psID)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte("services: [\ninvalid yaml"), cryptoutilSharedMagic.FilePermissions))

	err := CheckInDir(newLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), psID)
	assert.Contains(t, err.Error(), "cannot parse")
}

func TestCheckInDir_MissingService(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSMIM
	missingVariant := lintFitnessRegistry.ComposeVariantSQLite1
	missingSvc := lintFitnessRegistry.ComposeServiceName(psID, missingVariant)

	tmpDir := t.TempDir()
	setupAllComposeFiles(t, tmpDir)

	// Rebuild compose block without sqlite-1 entry.
	var sb strings.Builder

	for _, v := range orderedVariants {
		if v == missingVariant {
			continue
		}

		svcName := lintFitnessRegistry.ComposeServiceName(psID, v)
		cmd := expectedCommand(psID, v)
		quoted := make([]string, len(cmd))

		for i, arg := range cmd {
			quoted[i] = fmt.Sprintf("%q", arg)
		}

		sb.WriteString(fmt.Sprintf("  %s:\n    command: [%s]\n", svcName, strings.Join(quoted, ", ")))
	}

	writeComposeYML(t, tmpDir, psID, sb.String())

	err := CheckInDir(newLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), missingSvc)
	assert.Contains(t, err.Error(), "not found")
}

func TestCheckInDir_CommandMismatch(t *testing.T) {
	t.Parallel()

	psID := cryptoutilSharedMagic.OTLPServiceSMIM
	variant := lintFitnessRegistry.ComposeVariantSQLite1
	svcName := lintFitnessRegistry.ComposeServiceName(psID, variant)

	tests := []struct {
		name    string
		command []string
	}{
		{
			// Wrong first element (wrong subcommand).
			name:    "wrong subcommand",
			command: append([]string{cryptoutilSharedMagic.KMSServiceName}, expectedCommand(psID, variant)[1:]...),
		},
		{
			// Wrong DSN for sqlite variant.
			name:    "wrong DSN",
			command: append(expectedCommand(psID, variant)[:len(expectedCommand(psID, variant))-1], dsnPostgres),
		},
		{
			// Extra arg at the end (length mismatch).
			name:    "extra arg",
			command: append(expectedCommand(psID, variant), "--extra"),
		},
		{
			// Too few args (length mismatch).
			name:    "too few args",
			command: expectedCommand(psID, variant)[:4],
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			setupAllComposeFiles(t, tmpDir)

			// Overwrite sm-im/sqlite-1 with the wrong command.
			svcBlock := buildComposeBlock(psID)
			// Replace the sqlite-1 entry with a malformed command.
			quoted := make([]string, len(tc.command))
			for i, arg := range tc.command {
				quoted[i] = fmt.Sprintf("%q", arg)
			}

			badEntry := fmt.Sprintf("  %s:\n    command: [%s]\n", svcName, strings.Join(quoted, ", "))
			goodEntry := fmt.Sprintf("  %s:\n    command: [%s]", svcName, strings.Join(func() []string {
				want := expectedCommand(psID, variant)

				q := make([]string, len(want))
				for i, a := range want {
					q[i] = fmt.Sprintf("%q", a)
				}

				return q
			}(), ", "))

			svcBlock = strings.Replace(svcBlock, goodEntry, badEntry, 1)
			writeComposeYML(t, tmpDir, psID, svcBlock)

			err := CheckInDir(newLogger(), tmpDir)
			require.Error(t, err)
			assert.Contains(t, err.Error(), psID)
		})
	}
}

func TestCheckInDir_ReadFileFnError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	stubReadFileFn := func(_ string) ([]byte, error) {
		return nil, os.ErrPermission
	}

	err := checkInDir(newLogger(), tmpDir, stubReadFileFn)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read")
}
