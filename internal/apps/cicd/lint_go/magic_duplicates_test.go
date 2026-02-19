// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

// writeMagicFile creates a file inside dir with the given content.
func writeMagicFile(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600)
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_NoDuplicates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeMagicFile(t, dir, "magic_strings.go", `package magic

const (
ProtocolHTTPS = "https"
SchemeHTTP    = "http"
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := checkMagicDuplicatesInDir(logger, dir)
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_WithDuplicates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeMagicFile(t, dir, "magic_net.go", `package magic

const (
ProtocolHTTPS = "https"
SchemeHTTPS   = "https"
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := checkMagicDuplicatesInDir(logger, dir)
	// magic-duplicates is informational: violations are logged but do not return an error.
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_TrivialInts_NotDuplicate(t *testing.T) {
	t.Parallel()

	// Trivial integers (0,1,2,3,4,-1) are still flagged as duplicates if they
	// share the same value — isMagicTrivialLiteral only suppresses usage scanning,
	// not the duplicate-definition check.
	dir := t.TempDir()
	writeMagicFile(t, dir, "magic_sizes.go", `package magic

const (
DefaultSizeA = 5
DefaultSizeB = 5
)
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := checkMagicDuplicatesInDir(logger, dir)
	// magic-duplicates is informational: violations are logged but do not return an error.
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_InvalidDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := checkMagicDuplicatesInDir(logger, "/nonexistent/path/that/does/not/exist")
	require.Error(t, err)
}

func TestCheckMagicDuplicatesInDir_SingleConstant(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeMagicFile(t, dir, "magic_only.go", `package magic

const Alone = "solo"
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := checkMagicDuplicatesInDir(logger, dir)
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_EmptyPackage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeMagicFile(t, dir, "magic.go", `package magic
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := checkMagicDuplicatesInDir(logger, dir)
	require.NoError(t, err)
}

func TestCheckMagicDuplicatesInDir_MultiFile_Duplicates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeMagicFile(t, dir, "magic_a.go", `package magic

const AlgoRSA = "RSA"
`)
	writeMagicFile(t, dir, "magic_b.go", `package magic

const AlgorithmRSA = "RSA"
`)

	logger := cryptoutilCmdCicdCommon.NewLogger("magic-dup-test")
	err := checkMagicDuplicatesInDir(logger, dir)
	// magic-duplicates is informational: violations are logged but do not return an error.
	require.NoError(t, err)
}
