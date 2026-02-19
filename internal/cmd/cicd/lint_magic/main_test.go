// Copyright (c) 2025 Justin Cranford

package lint_magic

import (
"bytes"
"testing"

"github.com/stretchr/testify/assert"
)

func TestMainWithWriters_NoArgs(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

code := mainWithWriters(nil, &stdout, &stderr)

assert.Equal(t, 1, code)
assert.Contains(t, stdout.String(), "Usage")
}

func TestMainWithWriters_Help(t *testing.T) {
t.Parallel()

tests := []struct {
name string
args []string
}{
{name: "help", args: []string{"help"}},
{name: "--help", args: []string{"--help"}},
{name: "-h", args: []string{"-h"}},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

code := mainWithWriters(tc.args, &stdout, &stderr)

assert.Equal(t, 0, code)
assert.Contains(t, stdout.String(), "validate-duplicates")
assert.Contains(t, stdout.String(), "validate-usage")
assert.Contains(t, stdout.String(), "validate-all")
assert.Empty(t, stderr.String())
})
}
}

func TestMainWithWriters_UnknownCommand(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

code := mainWithWriters([]string{"unknown-command"}, &stdout, &stderr)

assert.Equal(t, 1, code)
assert.Contains(t, stderr.String(), "Unknown command")
}

func TestMainWithWriters_ValidateDuplicatesClean(t *testing.T) {
t.Parallel()

magicDir, _ := setupUsageFixture(t,
`package magic

const ProtocolHTTPS = "https"
`,
"",
)

var stdout, stderr bytes.Buffer

code := mainWithWriters([]string{"validate-duplicates", magicDir}, &stdout, &stderr)

assert.Equal(t, 0, code)
assert.Contains(t, stdout.String(), "OK")
assert.Empty(t, stderr.String())
}

func TestMainWithWriters_ValidateUsageClean(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const ProtocolHTTPS = "https"
`,
`package service

func f() string { return "grpc" }
`,
)

var stdout, stderr bytes.Buffer

code := mainWithWriters([]string{"validate-usage", magicDir, rootDir}, &stdout, &stderr)

assert.Equal(t, 0, code)
assert.Contains(t, stdout.String(), "OK")
}

func TestMainWithWriters_ValidateAllFail(t *testing.T) {
t.Parallel()

magicDir, rootDir := setupUsageFixture(t,
`package magic

const ProtocolHTTPS = "https"
`,
`package service

func f() string { return "https" }
`,
)

var stdout, stderr bytes.Buffer

code := mainWithWriters([]string{"validate-all", magicDir, rootDir}, &stdout, &stderr)

assert.Equal(t, 1, code)
assert.Contains(t, stdout.String(), "FAIL")
}

func TestMainWithWriters_ValidateDuplicatesNonexistentDir(t *testing.T) {
t.Parallel()

var stdout, stderr bytes.Buffer

code := mainWithWriters([]string{"validate-duplicates", "/nonexistent/path/lint/magic"}, &stdout, &stderr)

assert.Equal(t, 1, code)
}
