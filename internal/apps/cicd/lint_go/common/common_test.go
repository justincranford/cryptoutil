// Copyright (c) 2025 Justin Cranford

package common

import (
"go/ast"
"go/token"
"os"
"path/filepath"
"testing"

"github.com/stretchr/testify/require"
)

func TestPrintCryptoViolations(t *testing.T) {
t.Parallel()

violations := []CryptoViolation{
{File: "test.go", Line: 10, Issue: "weak crypto", Content: "md5.New()"},
}

// Should not panic.
PrintCryptoViolations("test-category", violations)
}

func TestParseMagicDir_ValidDir(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()

// Create a simple Go file with constants.
content := `package magic

const (
TestPort    = 8080
TestProtocol = "https"
)

var NotAConst = "skip"
`
require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "magic_test_constants.go"), []byte(content), 0o600))

inv, err := ParseMagicDir(tmpDir)
require.NoError(t, err)
require.NotNil(t, inv)
require.Len(t, inv.Constants, 2)
require.Contains(t, inv.ByName, "TestPort")
require.Contains(t, inv.ByName, "TestProtocol")
}

func TestParseMagicDir_InvalidDir(t *testing.T) {
t.Parallel()

_, err := ParseMagicDir("/nonexistent/dir")
require.Error(t, err)
}

func TestParseMagicDir_SkipTestFiles(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()

// Create a test file that should be filtered out.
content := `package magic

const TestOnly = "should-be-skipped"
`
require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "magic_test.go"), []byte(content), 0o600))

// Create a non-test file.
content2 := `package magic

const ProductionConst = "kept"
`
require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "magic_prod.go"), []byte(content2), 0o600))

inv, err := ParseMagicDir(tmpDir)
require.NoError(t, err)
require.Len(t, inv.Constants, 1)
require.Equal(t, "ProductionConst", inv.Constants[0].Name)
}

func TestParseMagicDir_DerivedConstant(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()

// Create a file with a derived constant (value is another ident, not a literal).
content := `package magic

const Base = "base"
const Derived = Base
`
require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "magic_derived.go"), []byte(content), 0o600))

inv, err := ParseMagicDir(tmpDir)
require.NoError(t, err)

// Only Base should be in inventory (Derived is skipped as non-BasicLit).
require.Len(t, inv.Constants, 1)
require.Equal(t, "Base", inv.Constants[0].Name)
}

func TestParseMagicDir_TestingFile(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()

// Create magic_testing.go which should mark constants as test constants.
content := `package magic

const TestingConst = "test-value"
`
require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "magic_testing.go"), []byte(content), 0o600))

inv, err := ParseMagicDir(tmpDir)
require.NoError(t, err)
require.Len(t, inv.Constants, 1)
require.True(t, inv.Constants[0].IsTestConst)
}

func TestParseMagicDir_ConstGroupNoValues(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()

// Create a file with iota constants (no explicit values for some).
content := `package magic

const (
A = iota
B
C
)

const Named = "literal"
`
require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "magic_iota.go"), []byte(content), 0o600))

inv, err := ParseMagicDir(tmpDir)
require.NoError(t, err)

// Only "Named" should be in the inventory (iota values are CallExpr, not BasicLit).
require.Len(t, inv.Constants, 1)
require.Equal(t, "Named", inv.Constants[0].Name)
}

func TestIsMagicTrivialLiteral_String(t *testing.T) {
t.Parallel()

tests := []struct {
name     string
value    string
expected bool
}{
{name: "empty string", value: `""`, expected: true},
{name: "single char", value: `"a"`, expected: true},
{name: "two chars", value: `"ab"`, expected: true},
{name: "three chars", value: `"abc"`, expected: false},
{name: "long string", value: `"https"`, expected: false},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

lit := &ast.BasicLit{Kind: token.STRING, Value: tc.value}
result := IsMagicTrivialLiteral(lit)
require.Equal(t, tc.expected, result)
})
}
}

func TestIsMagicTrivialLiteral_Int(t *testing.T) {
t.Parallel()

tests := []struct {
name     string
value    string
expected bool
}{
{name: "zero", value: "0", expected: true},
{name: "one", value: "1", expected: true},
{name: "negative one", value: "-1", expected: true},
{name: "large number", value: "8080", expected: false},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

lit := &ast.BasicLit{Kind: token.INT, Value: tc.value}
result := IsMagicTrivialLiteral(lit)
require.Equal(t, tc.expected, result)
})
}
}

func TestParseMagicDir_MultipleFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create constants in two different files to exercise cross-file sort.
	content1 := "package magic\n\nconst FileOneConst = \"one\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "magic_a.go"), []byte(content1), 0o600))

	content2 := "package magic\n\nconst FileTwoConst = \"two\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "magic_b.go"), []byte(content2), 0o600))

	inv, err := ParseMagicDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, inv.Constants, 2)
	require.Equal(t, "FileOneConst", inv.Constants[0].Name)
	require.Equal(t, "FileTwoConst", inv.Constants[1].Name)
}

func TestIsMagicTrivialLiteral_Float(t *testing.T) {
t.Parallel()

lit := &ast.BasicLit{Kind: token.FLOAT, Value: "3.14"}
result := IsMagicTrivialLiteral(lit)
require.False(t, result, "Float should not be trivial")
}

func TestMagicShouldSkipPath(t *testing.T) {
t.Parallel()

tests := []struct {
name     string
path     string
expected bool
}{
{name: "vendor dir", path: "vendor/pkg/file.go", expected: true},
{name: "test-output dir", path: "test-output/report.txt", expected: true},
{name: "workflow-reports dir", path: "workflow-reports/report.json", expected: true},
{name: "api client dir", path: "api/client/client.gen.go", expected: true},
{name: "api model dir", path: "api/model/models.gen.go", expected: true},
{name: "api server dir", path: "api/server/server.gen.go", expected: true},
{name: "normal path", path: "internal/server/handler.go", expected: false},
{name: "api non-generated", path: "api/custom/handler.go", expected: false},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

result := MagicShouldSkipPath(tc.path)
require.Equal(t, tc.expected, result)
})
}
}

func TestIsMagicGeneratedFile(t *testing.T) {
t.Parallel()

tests := []struct {
name     string
filename string
expected bool
}{
{name: "gen.go suffix", filename: "server.gen.go", expected: true},
{name: "_gen_ contains", filename: "types_gen_models.go", expected: true},
{name: "normal file", filename: "handler.go", expected: false},
{name: "test file", filename: "handler_test.go", expected: false},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

result := IsMagicGeneratedFile(tc.filename)
require.Equal(t, tc.expected, result)
})
}
}
