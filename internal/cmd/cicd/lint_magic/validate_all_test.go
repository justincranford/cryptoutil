// Copyright (c) 2025 Justin Cranford

package lint_magic

import (
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

func TestValidateAll_Clean(t *testing.T) {
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

result, err := ValidateAll(magicDir, rootDir)
require.NoError(t, err)

assert.True(t, result.Valid)
assert.NotNil(t, result.Duplicates)
assert.NotNil(t, result.Usage)
assert.Empty(t, result.Errors)
}

func TestValidateAll_BothValidatorsFail(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const (
ProtocolHTTPS      = "https"
ProtocolHTTPSAlias = "https"
)
`,
`package service

func f() string { return "https" }
`,
)

result, err := ValidateAll(magicDir, rootDir)
require.NoError(t, err)

assert.False(t, result.Valid)
assert.NotNil(t, result.Duplicates)
assert.False(t, result.Duplicates.Valid)
assert.NotNil(t, result.Usage)
assert.False(t, result.Usage.Valid)
}

func TestValidateAll_OnlyDuplicatesFail(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const (
ProtocolHTTPS      = "https"
ProtocolHTTPSAlias = "https"
)
`,
"",
)

result, err := ValidateAll(magicDir, rootDir)
require.NoError(t, err)

assert.False(t, result.Valid)
assert.False(t, result.Duplicates.Valid)
assert.True(t, result.Usage.Valid)
}

func TestValidateAll_OnlyUsageFails(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const ProtocolHTTPS = "https"
`,
`package service

func f() string { return "https" }
`,
)

result, err := ValidateAll(magicDir, rootDir)
require.NoError(t, err)

assert.False(t, result.Valid)
assert.True(t, result.Duplicates.Valid)
assert.False(t, result.Usage.Valid)
}

func TestFormatAllResult_OKOutput(t *testing.T) {
t.Parallel()

result := &AllResult{
Valid:      true,
Duplicates: &DuplicatesResult{Valid: true},
Usage:      &UsageResult{Valid: true},
}

output := FormatAllResult(result)

assert.Contains(t, output, "validate-all: OK")
assert.NotContains(t, output, "validate-all: FAIL")
}

func TestFormatAllResult_FailOutput(t *testing.T) {
t.Parallel()

result := &AllResult{
Valid:      false,
Duplicates: &DuplicatesResult{Valid: false, Duplicates: []DuplicateGroup{{Value: `"x"`, Constants: []MagicConstant{{Name: "A"}, {Name: "B"}}}}},
Usage:      &UsageResult{Valid: true},
}

output := FormatAllResult(result)

assert.Contains(t, output, "validate-all: FAIL")
}
