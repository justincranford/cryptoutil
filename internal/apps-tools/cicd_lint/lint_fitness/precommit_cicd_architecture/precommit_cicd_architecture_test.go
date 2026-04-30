// Copyright (c) 2025-2026 Justin Cranford.
package precommit_cicd_architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	. "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/precommit_cicd_architecture"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, rootDir, content string) {
	t.Helper()

	configPath := filepath.Join(rootDir, ".pre-commit-config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
}

func buildValidConfig() string {
	return strings.TrimSpace(`
repos:
  - repo: local
    hooks:
      - id: cicd-lint-pre-commit
        entry: go
        args: [run, cmd/cicd-lint/main.go, -q, lint-fitness, lint-text, lint-go, lint-go-test, lint-golangci, lint-compose, lint-openapi, lint-ports, lint-workflow, lint-deployments, lint-docs, lint-java-test, lint-python-test]
        require_serial: false
        stages: [pre-commit]

      - id: cicd-format-pre-commit
        entry: go
        args: [run, cmd/cicd-lint/main.go, -q, format-go, format-go-test]
        require_serial: true
        stages: [pre-commit]

      - id: cicd-lint-pre-push
        entry: go
        args: [run, cmd/cicd-lint/main.go, -q, lint-go-mod]
        require_serial: false
        stages: [pre-push]

      - id: cicd-format-pre-push
        entry: go
        args: [run, cmd/cicd-lint/main.go, -q, format-go, format-go-test]
        require_serial: true
        stages: [pre-push]
`)
}

func TestCheckInDir_ValidConfig(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeConfig(t, rootDir, buildValidConfig())

	logger := cryptoutilCmdCicdCommon.NewLogger("test-precommit-cicd-architecture")
	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_MixedHookRejected(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	content := strings.Replace(buildValidConfig(), "lint-go-mod]", "lint-go-mod, format-go]", 1)
	writeConfig(t, rootDir, content)

	logger := cryptoutilCmdCicdCommon.NewLogger("test-precommit-cicd-architecture")
	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "mixes lint-* and format-* commands")
}

func TestCheckInDir_MissingStageHookRejected(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	content := strings.Replace(
		buildValidConfig(),
		"- id: cicd-format-pre-push\n        entry: go\n        args: [run, cmd/cicd-lint/main.go, -q, format-go, format-go-test]\n        require_serial: true\n        stages: [pre-push]",
		"# removed pre-push format hook",
		1,
	)
	require.NotEqual(t, buildValidConfig(), content)
	writeConfig(t, rootDir, content)

	logger := cryptoutilCmdCicdCommon.NewLogger("test-precommit-cicd-architecture")
	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "expected exactly one format-only bulk cicd hook for stage \"pre-push\"")
}

func TestCheckInDir_RequireSerialValidation(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	content := strings.Replace(buildValidConfig(), "require_serial: false", "require_serial: true", 1)
	writeConfig(t, rootDir, content)

	logger := cryptoutilCmdCicdCommon.NewLogger("test-precommit-cicd-architecture")
	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "lint-only but require_serial=true")
}

func TestCheckInDir_MissingCommandCoverageRejected(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	content := strings.Replace(buildValidConfig(), "lint-java-test, ", "", 1)
	writeConfig(t, rootDir, content)

	logger := cryptoutilCmdCicdCommon.NewLogger("test-precommit-cicd-architecture")
	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "command \"lint-java-test\" is not present")
}
