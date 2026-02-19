// Copyright (c) 2025 Justin Cranford

package lint_magic

import (
"os"
"path/filepath"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

// setupUsageFixture creates a temporary project layout with a magic package and
// a separate Go file that may contain violations.  Returns (magicDir, rootDir).
func setupUsageFixture(t *testing.T, magicSrc, violatingSrc string) (string, string) {
t.Helper()

rootDir := t.TempDir()
magicDir := filepath.Join(rootDir, "internal", "shared", "magic")
require.NoError(t, os.MkdirAll(magicDir, 0o700))

writeMagicFixture(t, magicDir, "magic_fixture.go", magicSrc)

if violatingSrc != "" {
serviceDir := filepath.Join(rootDir, "internal", "service")
require.NoError(t, os.MkdirAll(serviceDir, 0o700))
writeMagicFixture(t, serviceDir, "service.go", violatingSrc)
}

return magicDir, rootDir
}

func TestValidateUsage_NoViolations(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const ProtocolHTTPS = "https"
`,
`package service

import magic "cryptoutil/internal/shared/magic"

func bind() string { return magic.ProtocolHTTPS }
`,
)

result, err := ValidateUsage(magicDir, rootDir)
require.NoError(t, err)

assert.True(t, result.Valid)
assert.Empty(t, result.Violations)
assert.Empty(t, result.Errors)
}

func TestValidateUsage_LiteralViolation(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const ProtocolHTTPS = "https"
`,
`package service

func bind() string { return "https" }
`,
)

result, err := ValidateUsage(magicDir, rootDir)
require.NoError(t, err)

assert.False(t, result.Valid)
require.Len(t, result.Violations, 1)
assert.Equal(t, UsageKindLiteral, result.Violations[0].Kind)
assert.Equal(t, `"https"`, result.Violations[0].LiteralValue)
assert.Equal(t, "ProtocolHTTPS", result.Violations[0].MagicName)
}

func TestValidateUsage_ConstRedefineViolation(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const ProtocolHTTPS = "https"
`,
`package service

const localHTTPS = "https"
`,
)

result, err := ValidateUsage(magicDir, rootDir)
require.NoError(t, err)

assert.False(t, result.Valid)
require.Len(t, result.Violations, 1)
assert.Equal(t, UsageKindRedefine, result.Violations[0].Kind)
assert.Equal(t, `"https"`, result.Violations[0].LiteralValue)
}

func TestValidateUsage_TrivialStringsNotFlagged(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const ShortStr = "ab"
`,
`package service

func f() string { return "ab" }
`,
)

// "ab" has unquoted length 2 < minStringLen(3), so it should not be flagged.
result, err := ValidateUsage(magicDir, rootDir)
require.NoError(t, err)

assert.True(t, result.Valid)
assert.Empty(t, result.Violations)
}

func TestValidateUsage_TrivialIntsNotFlagged(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const One = 1
`,
`package service

func f() int { return 1 }
`,
)

// 1 is in trivialInts, so it should not be flagged.
result, err := ValidateUsage(magicDir, rootDir)
require.NoError(t, err)

assert.True(t, result.Valid)
assert.Empty(t, result.Violations)
}

func TestValidateUsage_NonTrivialIntFlagged(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const MaxRetry = 10
`,
`package service

func f() int { return 10 }
`,
)

result, err := ValidateUsage(magicDir, rootDir)
require.NoError(t, err)

assert.False(t, result.Valid)
require.Len(t, result.Violations, 1)
assert.Equal(t, "10", result.Violations[0].LiteralValue)
assert.Equal(t, "MaxRetry", result.Violations[0].MagicName)
}

func TestValidateUsage_SkipsMagicDirItself(t *testing.T) {
t.Parallel()

// Duplicate definitions inside the magic package must not self-flag.
magicDir, rootDir := setupUsageFixture(t,
`package magic

const (
ProtocolHTTPS      = "https"
ProtocolHTTPSAlias = "https"
)
`,
"",
)

result, err := ValidateUsage(magicDir, rootDir)
require.NoError(t, err)

assert.True(t, result.Valid)
assert.Empty(t, result.Violations)
}

func TestValidateUsage_SkipsGeneratedFiles(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const ProtocolHTTPS = "https"
`,
"",
)

// Write a generated file that uses the magic value.
genDir := filepath.Join(rootDir, "api", "model")
require.NoError(t, os.MkdirAll(genDir, 0o700))
writeMagicFixture(t, genDir, "models.gen.go", `package model

const scheme = "https"
`)

result, err := ValidateUsage(magicDir, rootDir)
require.NoError(t, err)

assert.True(t, result.Valid)
}

func TestValidateUsage_SkipsVendorDir(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const ProtocolHTTPS = "https"
`,
"",
)

vendorDir := filepath.Join(rootDir, "vendor", "somepkg")
require.NoError(t, os.MkdirAll(vendorDir, 0o700))
writeMagicFixture(t, vendorDir, "pkg.go", `package somepkg

const scheme = "https"
`)

result, err := ValidateUsage(magicDir, rootDir)
require.NoError(t, err)

assert.True(t, result.Valid)
}

func TestValidateUsage_NonexistentMagicDir(t *testing.T) {
t.Parallel()

result, err := ValidateUsage("/tmp/nonexistent-magic-lint-magic", ".")
require.NoError(t, err)

assert.False(t, result.Valid)
assert.NotEmpty(t, result.Errors)
}

func TestFormatUsageResult_OKOutput(t *testing.T) {
t.Parallel()

result := &UsageResult{Valid: true}
output := FormatUsageResult(result)

assert.Contains(t, output, "OK")
assert.NotContains(t, output, "FAIL")
}

func TestFormatUsageResult_FailOutput(t *testing.T) {
t.Parallel()

result := &UsageResult{
Valid: false,
Violations: []UsageViolation{
{
File:         "internal/service/service.go",
Line:         42,
Kind:         UsageKindLiteral,
LiteralValue: `"https"`,
MagicName:    "ProtocolHTTPS",
},
},
}

output := FormatUsageResult(result)

assert.Contains(t, output, "FAIL")
assert.Contains(t, output, "service.go")
assert.Contains(t, output, "literal-use")
assert.Contains(t, output, `"https"`)
assert.Contains(t, output, "ProtocolHTTPS")
}
