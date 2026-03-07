---
name: test-fuzz-gen
description: "Generate _fuzz_test.go fuzz tests conforming to cryptoutil project standards. Use when adding fuzz coverage for parsers, decoders, or crypto input handling to ensure correct build tags, 15s minimum fuzz time, seed corpus, and safe assertion patterns."
argument-hint: "[package or function name]"
---

Generate `_fuzz_test.go` fuzz tests conforming to cryptoutil project standards.

## Purpose

Use when creating fuzz tests for functions that parse or process external input.
Fuzz tests go in a separate `_fuzz_test.go` file (ONLY fuzz functions).

## Key Rules

- File suffix: `_fuzz_test.go` (ONLY fuzz functions, never mixed with unit tests)
- Minimum fuzz time: `15s` per test
- **CRITICAL: Function names MUST NOT be substrings of other fuzz function names** — e.g. use `FuzzHKDFAllVariants`, NEVER `FuzzHKDF` if `FuzzHKDFAllVariants` exists in the same package
- Build tag: `//go:build fuzz` (optional marker — run with `-tags fuzz` or `-fuzz=FuzzXxx` without tag)
- Property tests that MUST NOT run during fuzzing: add `//go:build !fuzz` at top of `_property_test.go` file
- Corpus: provide seed entries covering edge cases (empty, nil, boundary values)
- Run from project root: `go test -fuzz=FuzzXxx -fuzztime=15s ./path/to/pkg`

## Template

```go
//go:build fuzz

package mypkg_test

import (
"testing"
)

func FuzzParseInput(f *testing.F) {
// Seed corpus — cover edge cases
f.Add([]byte(""))
f.Add([]byte("valid-input"))
f.Add([]byte("{invalid json}"))
f.Add([]byte("\x00\xff"))

f.Fuzz(func(t *testing.T, data []byte) {
// Must not panic
result, _ := ParseInput(data)
if result != nil {
// Validate invariants
_ = result
}
})
}
```

## References

Read [ARCHITECTURE.md Section 10.7 Fuzz Testing Strategy](../../../docs/ARCHITECTURE.md#107-fuzz-testing-strategy) for fuzz testing requirements — apply the 15s minimum fuzz time, `_fuzz_test.go` file suffix, unique function name rule, and seed corpus requirements from this section.

Read [ARCHITECTURE.md Section 10.1 Testing Strategy Overview](../../../docs/ARCHITECTURE.md#101-testing-strategy-overview) for test file type suffixes — ensure `_fuzz_test.go` files contain ONLY fuzz functions and cross-check that `_property_test.go` files use `//go:build !fuzz` if they must not run during fuzz corpus execution.
