// Copyright (c) 2025 Justin Cranford

package lint_javatest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilLintJavaTest "cryptoutil/internal/apps/tools/cicd_lint/lint_javatest"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestLint_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := cryptoutilLintJavaTest.Lint(logger, map[string][]string{})

	require.NoError(t, err, "Lint should succeed with no files")
}

func TestLint_NoJavaFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go":  {"main.go", "util.go"},
		"yml": {"config.yml"},
	}

	err := cryptoutilLintJavaTest.Lint(logger, filesByExtension)

	require.NoError(t, err, "Lint should succeed with no Java files")
}

func TestLint_ValidJavaFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "SecureSimulation.java")
	content := `package cryptoutil;

import java.security.SecureRandom;
import io.gatling.javaapi.core.*;

public class SecureSimulation extends Simulation {
    private static final SecureRandom SECURE_RANDOM = new SecureRandom();
}
`
	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"java": {testFile},
	}

	err = cryptoutilLintJavaTest.Lint(logger, filesByExtension)

	require.NoError(t, err, "Lint should succeed with SecureRandom usage")
}

func TestLint_InsecureRandom_NewRandom(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "InsecureSimulation.java")
	content := `package cryptoutil;

import java.util.Random;
import io.gatling.javaapi.core.*;

public class InsecureSimulation extends Simulation {
    private static final Random RANDOM = new Random();
}
`
	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"java": {testFile},
	}

	err = cryptoutilLintJavaTest.Lint(logger, filesByExtension)

	require.Error(t, err, "Lint should fail for new Random() usage")
	require.ErrorContains(t, err, "insecure RNG violations")
}

func TestLint_InsecureRandom_MathRandom(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "MathRandomSimulation.java")
	content := `package cryptoutil;

import io.gatling.javaapi.core.*;

public class MathRandomSimulation extends Simulation {
    private double roll() {
        return Math.random();
    }
}
`
	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"java": {testFile},
	}

	err = cryptoutilLintJavaTest.Lint(logger, filesByExtension)

	require.Error(t, err, "Lint should fail for Math.random() usage")
	require.ErrorContains(t, err, "insecure RNG violations")
}

func TestCheckInsecureRandom_FileNotFound(t *testing.T) {
	t.Parallel()

	issues := cryptoutilLintJavaTest.CheckInsecureRandom("/nonexistent/path/file.java")

	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "Error reading file")
}

func TestCheckInsecureRandom_MultipleViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "MultiViolation.java")
	content := `package cryptoutil;
import java.util.Random;

public class MultiViolation {
    private static final Random r1 = new Random();
    private static final Random r2 = new Random();
    private double roll() { return Math.random(); }
}
`
	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	issues := cryptoutilLintJavaTest.CheckInsecureRandom(testFile)

	require.Len(t, issues, 3, "Should detect two new Random() and one Math.random() violations")
}
