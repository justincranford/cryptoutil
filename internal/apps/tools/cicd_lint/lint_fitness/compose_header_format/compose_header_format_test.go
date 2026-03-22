// Copyright (c) 2025 Justin Cranford

package compose_header_format_test

import (
"fmt"
"os"
"path/filepath"
"strings"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
lintFitnessComposeHeaderFormat "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_header_format"
lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
return cryptoutilCmdCicdCommon.NewLogger("test")
}

func findProjectRoot(t *testing.T) string {
t.Helper()

dir, err := os.Getwd()
require.NoError(t, err, "failed to get working directory")

for {
if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
return dir
}

parent := filepath.Dir(dir)
if parent == dir {
t.Skip("skipping integration test: cannot find project root (no go.mod)")
}

dir = parent
}
}

// createComposeFile writes a minimal compose.yml with the given header lines.
func createComposeFile(t *testing.T, tmpDir, psID, line3, line5 string) {
t.Helper()

deployDir := filepath.Join(tmpDir, "deployments", psID)
require.NoError(t, os.MkdirAll(deployDir, cryptoutilSharedMagic.CICDTempDirPermissions))

content := fmt.Sprintf("# $schema: ...\n#\n%s\n#\n%s\n", line3, line5)
require.NoError(t, os.WriteFile(filepath.Join(deployDir, "compose.yml"), []byte(content), cryptoutilSharedMagic.FilePermissions))
}

// setupAllComposeFiles creates complete minimal compose.yml files for all PS.
func setupAllComposeFiles(t *testing.T, tmpDir string) {
t.Helper()

for _, ps := range lintFitnessRegistry.AllProductServices() {
line3 := "# " + strings.ToUpper(ps.PSID) + " Docker Compose Configuration"
line5 := "# SERVICE-level deployment for " + ps.DisplayName + "."
createComposeFile(t, tmpDir, ps.PSID, line3, line5)
}
}

func TestCheck_RealWorkspace(t *testing.T) {
t.Parallel()

root := findProjectRoot(t)

err := lintFitnessComposeHeaderFormat.CheckInDir(newTestLogger(), root)
require.NoError(t, err)
}

func TestCheckInDir_AllCorrect(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
setupAllComposeFiles(t, tmpDir)

err := lintFitnessComposeHeaderFormat.CheckInDir(newTestLogger(), tmpDir)
require.NoError(t, err)
}

func TestCheckInDir_MissingFile(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
setupAllComposeFiles(t, tmpDir)

// Remove one compose file.
require.NoError(t, os.Remove(filepath.Join(tmpDir, "deployments", cryptoutilSharedMagic.OTLPServiceSMIM, "compose.yml")))

err := lintFitnessComposeHeaderFormat.CheckInDir(newTestLogger(), tmpDir)
require.Error(t, err)
assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
}

func TestCheckInDir_WrongLine3_Lowercase(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
setupAllComposeFiles(t, tmpDir)

// Overwrite sm-im compose with lowercase line 3.
createComposeFile(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMIM,
"# "+cryptoutilSharedMagic.OTLPServiceSMIM+" Docker Compose Configuration",
"# SERVICE-level deployment for Secrets Manager Instant Messenger.",
)

err := lintFitnessComposeHeaderFormat.CheckInDir(newTestLogger(), tmpDir)
require.Error(t, err)
assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMIM)
assert.Contains(t, err.Error(), "line 3")
}

func TestCheckInDir_WrongLine5_WrongDisplayName(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
setupAllComposeFiles(t, tmpDir)

// Overwrite sm-kms compose with wrong display name on line 5.
createComposeFile(t, tmpDir, cryptoutilSharedMagic.OTLPServiceSMKMS,
"# SM-KMS Docker Compose Configuration",
"# SERVICE-level deployment for SM Key Management Service.",
)

err := lintFitnessComposeHeaderFormat.CheckInDir(newTestLogger(), tmpDir)
require.Error(t, err)
assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceSMKMS)
assert.Contains(t, err.Error(), "line 5")
}

func TestCheckInDir_TooFewLines(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
setupAllComposeFiles(t, tmpDir)

// Write a compose file with only 3 lines for jose-ja.
deployDir := filepath.Join(tmpDir, "deployments", cryptoutilSharedMagic.OTLPServiceJoseJA)
require.NoError(t, os.MkdirAll(deployDir, cryptoutilSharedMagic.CICDTempDirPermissions))
require.NoError(t, os.WriteFile(
filepath.Join(deployDir, "compose.yml"),
[]byte("line1\nline2\nline3\n"),
cryptoutilSharedMagic.FilePermissions,
))

err := lintFitnessComposeHeaderFormat.CheckInDir(newTestLogger(), tmpDir)
require.Error(t, err)
assert.Contains(t, err.Error(), cryptoutilSharedMagic.OTLPServiceJoseJA)
}
