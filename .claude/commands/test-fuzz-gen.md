---
name: test-fuzz-gen
description: "Generate _fuzz_test.go fuzz tests conforming to cryptoutil project standards. Use when adding fuzz coverage for parsers, decoders, or crypto input handling to ensure correct build tags, 15s minimum fuzz time, seed corpus, and safe assertion patterns."
argument-hint: "[package or function name]"
---

Generate a `_fuzz_test.go` file for the specified function or parser.

**Full Copilot original**: [.github/skills/test-fuzz-gen/SKILL.md](.github/skills/test-fuzz-gen/SKILL.md)

## Key Rules

- Build tag `//go:build fuzz` MANDATORY (top of file)
- Minimum 15 seconds fuzz time: `f.Fuzz(func(t *testing.T, ...) {...})`
- Function names MUST NOT be substrings of other fuzz function names (avoids go test -run collision)
- Seed corpus MUST include: empty input, nil/zero values, boundary values, valid examples
- Property-based equivalents go in `_property_test.go` with `//go:build !fuzz`
- `t.Parallel()` in fuzz functions

## Template

```go
//go:build fuzz

package packagename_test

import (
    "testing"
)

// FuzzFunctionName tests FunctionName with arbitrary inputs.
// CRITICAL: Name must not be a substring of any other Fuzz function name.
func FuzzFunctionNameInput(f *testing.F) {
    // Seed corpus
    f.Add([]byte{})           // empty
    f.Add([]byte("valid"))    // valid example
    f.Add([]byte{0xFF, 0x00}) // boundary bytes

    f.Fuzz(func(t *testing.T, data []byte) {
        t.Parallel()

        // Must not panic
        result, err := FunctionUnderTest(data)
        if err != nil {
            return // errors are acceptable
        }

        // Property: round-trip or invariant
        _ = result
    })
}
```

## Seed Corpus File

Create `testdata/fuzz/FuzzFunctionNameInput/` directory with seed files:
```
testdata/fuzz/FuzzFunctionNameInput/
├── empty
├── valid_example
└── boundary_values
```
