package crypto_rand

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	lintGoCommon "cryptoutil/internal/apps/cicd/lint_go/common"
)

func TestCheckCryptoRand_Clean(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cleanFile := filepath.Join(tempDir, "clean.go")
	content := `package main

import (
	"crypto/rand"
)

func main() {
	buf := make([]byte, 32)
	rand.Read(buf)
}
`

	err := os.WriteFile(cleanFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckFileForMathRand(cleanFile)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckCryptoRand_MathRandImport(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	badFile := filepath.Join(tempDir, "bad.go")
	content := `package main

import (
	"math/rand"
)

func main() {
	x := rand.Intn(100)
	println(x)
}
`

	err := os.WriteFile(badFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckFileForMathRand(badFile)
	require.NoError(t, err)
	require.Len(t, violations, 2)
	require.Contains(t, violations[0].Issue, "imports math/rand")
	require.Contains(t, violations[1].Issue, "uses math/rand function")
}

func TestCheckCryptoRand_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := CheckFileForMathRand("/nonexistent/path.go")
	require.Error(t, err)
}

func TestPrintMathRandViolations(t *testing.T) {
	t.Parallel()

	violations := []lintGoCommon.CryptoViolation{
		{File: "file1.go", Line: 10, Content: "import math/rand", Issue: "imports math/rand instead of crypto/rand"},
		{File: "file1.go", Line: 20, Content: "rand.Float64()", Issue: "uses math/rand function"},
	}

	// Just verify the print function does not panic.
	lintGoCommon.PrintCryptoViolations("math/rand", violations)
}
