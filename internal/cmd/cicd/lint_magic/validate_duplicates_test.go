// Copyright (c) 2025 Justin Cranford

package lint_magic

import (
"os"
"path/filepath"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

// writeMagicFixture writes a Go source file into dir and returns its path.
func writeMagicFixture(t *testing.T, dir, filename, src string) {
t.Helper()

path := filepath.Join(dir, filename)
require.NoError(t, os.WriteFile(path, []byte(src), 0o600))
}

func TestValidateDuplicates_NoDuplicates(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeMagicFixture(t, dir, "magic_test_fixture.go", `package magic

const (
ProtocolHTTPS = "https"
ProtocolHTTP  = "http"
PortHTTPS     = 443
)
`)

result, err := ValidateDuplicates(dir)
require.NoError(t, err)

assert.True(t, result.Valid)
assert.Empty(t, result.Duplicates)
assert.Empty(t, result.Errors)
}

func TestValidateDuplicates_WithDuplicates(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeMagicFixture(t, dir, "magic_test_fixture.go", `package magic

const (
FilePermDefault    = 0o600
FilePermOwnerRW    = 0o600
ProtocolHTTPS      = "https"
ProtocolHTTPSAlias = "https"
)
`)

result, err := ValidateDuplicates(dir)
require.NoError(t, err)

assert.False(t, result.Valid)
assert.Len(t, result.Duplicates, 2)
assert.Empty(t, result.Errors)
}

func TestValidateDuplicates_SkipsDerivedConstants(t *testing.T) {
t.Parallel()

dir := t.TempDir()
// DefaultProfile = EmptyString is a derived constant (references another ident).
// It should not appear in the inventory and therefore not cause a false duplicate.
writeMagicFixture(t, dir, "magic_test_fixture.go", `package magic

const (
EmptyString    = ""
DefaultProfile = EmptyString
ProtocolHTTPS  = "https"
)
`)

result, err := ValidateDuplicates(dir)
require.NoError(t, err)

assert.True(t, result.Valid)
assert.Empty(t, result.Duplicates)
}

func TestValidateDuplicates_MultipleFiles(t *testing.T) {
t.Parallel()

dir := t.TempDir()
writeMagicFixture(t, dir, "magic_network.go", `package magic

const ProtocolHTTPS = "https"
`)
writeMagicFixture(t, dir, "magic_api.go", `package magic

const APIProtocol = "https"
`)

result, err := ValidateDuplicates(dir)
require.NoError(t, err)

assert.False(t, result.Valid)
require.Len(t, result.Duplicates, 1)
assert.Equal(t, `"https"`, result.Duplicates[0].Value)
assert.Len(t, result.Duplicates[0].Constants, 2)
}

func TestValidateDuplicates_NonexistentDir(t *testing.T) {
t.Parallel()

result, err := ValidateDuplicates("/tmp/this-dir-does-not-exist-lint-magic")
require.NoError(t, err)

assert.False(t, result.Valid)
assert.NotEmpty(t, result.Errors)
}

func TestFormatDuplicatesResult_OKOutput(t *testing.T) {
t.Parallel()

result := &DuplicatesResult{Valid: true}
output := FormatDuplicatesResult(result)

assert.Contains(t, output, "OK")
assert.NotContains(t, output, "FAIL")
}

func TestFormatDuplicatesResult_FailOutput(t *testing.T) {
t.Parallel()

result := &DuplicatesResult{
Valid: false,
Duplicates: []DuplicateGroup{
{
Value: `"https"`,
Constants: []MagicConstant{
{Name: "ProtocolHTTPS", Value: `"https"`, File: "magic_network.go", Line: 5},
{Name: "APIProtocol", Value: `"https"`, File: "magic_api.go", Line: 3},
},
},
},
}

output := FormatDuplicatesResult(result)

assert.Contains(t, output, "FAIL")
assert.Contains(t, output, `"https"`)
assert.Contains(t, output, "ProtocolHTTPS")
assert.Contains(t, output, "APIProtocol")
}

func TestSortDuplicateGroups_Deterministic(t *testing.T) {
t.Parallel()

groups := []DuplicateGroup{
{Value: `"zzz"`, Constants: []MagicConstant{{Name: "Z"}}},
{Value: `"aaa"`, Constants: []MagicConstant{{Name: "A"}}},
{Value: `"mmm"`, Constants: []MagicConstant{{Name: "M"}}},
}

sortDuplicateGroups(groups)

assert.Equal(t, `"aaa"`, groups[0].Value)
assert.Equal(t, `"mmm"`, groups[1].Value)
assert.Equal(t, `"zzz"`, groups[2].Value)
}
