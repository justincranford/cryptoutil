package lint_go

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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

	violations, err := checkFileForMathRand(cleanFile)
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

	violations, err := checkFileForMathRand(badFile)
	require.NoError(t, err)
	require.Len(t, violations, 2)
	require.Contains(t, violations[0].Issue, "imports math/rand")
	require.Contains(t, violations[1].Issue, "uses math/rand function")
}

func TestCheckCryptoRand_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := checkFileForMathRand("/nonexistent/path.go")
	require.Error(t, err)
}

func TestCheckInsecureSkipVerify_Clean(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cleanFile := filepath.Join(tempDir, "clean.go")
	content := `package main

import (
	"crypto/tls"
)

func main() {
	config := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	println(config)
}
`

	err := os.WriteFile(cleanFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkFileForInsecureSkipVerify(cleanFile)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckInsecureSkipVerify_Violation(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	badFile := filepath.Join(tempDir, "bad.go")
	// Use concatenation to avoid triggering the linter on this test file.
	content := "package main\n\nimport (\n\t\"crypto/tls\"\n)\n\nfunc main() {\n\tconfig := &tls.Config{\n\t\t" + "Insecure" + "SkipVerify: true,\n\t}\n\tprintln(config)\n}\n"

	err := os.WriteFile(badFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkFileForInsecureSkipVerify(badFile)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0].Issue, "disables TLS certificate verification")
}

func TestCheckInsecureSkipVerify_WithNolint(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	cleanFile := filepath.Join(tempDir, "clean.go")
	// Use concatenation to avoid triggering the linter on this test file.
	content := "package main\n\nimport (\n\t\"crypto/tls\"\n)\n\nfunc main() {\n\tconfig := &tls.Config{\n\t\t" + "Insecure" + "SkipVerify: true, //nolint:all\n\t}\n\tprintln(config)\n}\n"

	err := os.WriteFile(cleanFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkFileForInsecureSkipVerify(cleanFile)
	require.NoError(t, err)
	require.Empty(t, violations) // Should be skipped due to nolint.
}

func TestCheckInsecureSkipVerify_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := checkFileForInsecureSkipVerify("/nonexistent/path.go")
	require.Error(t, err)
}

func TestPrintMathRandViolations(t *testing.T) {
	t.Parallel()

	violations := []cryptoViolation{
		{File: "file1.go", Line: 10, Content: "import math/rand", Issue: "imports math/rand instead of crypto/rand"},
		{File: "file1.go", Line: 20, Content: "rand.Float64()", Issue: "uses math/rand function"},
	}

	// Just verify the print function does not panic.
	printCryptoViolations("math/rand", violations)
}

func TestPrintInsecureSkipVerifyViolations(t *testing.T) {
	t.Parallel()

	violations := []cryptoViolation{
		{File: "file2.go", Line: 5, Content: "TLS config", Issue: "disables TLS certificate verification"},
	}

	// Just verify the print function does not panic.
	printCryptoViolations("InsecureSkipVerify", violations)
}
