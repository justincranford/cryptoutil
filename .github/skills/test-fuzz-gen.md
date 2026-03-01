# test-fuzz-gen

Generate `_fuzz_test.go` fuzz tests conforming to cryptoutil project standards.

## Purpose

Use when creating fuzz tests for functions that parse or process external input.
Fuzz tests go in a separate `_fuzz_test.go` file (ONLY fuzz functions).

## Key Rules

- File suffix: `_fuzz_test.go` (ONLY fuzz functions, never mixed with unit tests)
- Minimum fuzz time: `15s` per test
- Build tag: `//go:build fuzz` (optional — run with `-tags fuzz` or `-fuzz=FuzzXxx`)
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

See [ARCHITECTURE.md Section 10.7 Fuzz Testing Strategy](../../docs/ARCHITECTURE.md#107-fuzz-testing-strategy) for fuzz testing requirements.
