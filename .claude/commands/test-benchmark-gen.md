Generate a `_bench_test.go` file for crypto operations or performance-sensitive code.

**Full Copilot original**: [.github/skills/test-benchmark-gen/SKILL.md](.github/skills/test-benchmark-gen/SKILL.md)

## Rules

- `ResetTimer()` after any setup (MANDATORY for crypto benchmarks)
- `StopTimer()`/`StartTimer()` around per-iteration setup that shouldn't be measured
- `ReportAllocs()` for allocation-sensitive code
- `SetBytes(n)` for throughput measurement (AES, HMAC, hash operations)
- Table-driven benchmarks for multiple algorithm sizes/curves/key lengths
- `b.RunParallel()` for concurrent throughput benchmarks

## Template

```go
package packagename_test

import (
    "testing"
)

func BenchmarkFunctionName(b *testing.B) {
    // Setup (not measured)
    key := generateTestKey()

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        result, err := FunctionUnderTest(key, testData)
        if err != nil {
            b.Fatal(err)
        }
        _ = result
    }
}

// BenchmarkFunctionNameSizes tests across multiple input sizes.
func BenchmarkFunctionNameSizes(b *testing.B) {
    sizes := []int{64, 256, 1024, 4096, 16384}

    for _, size := range sizes {
        size := size
        b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
            data := make([]byte, size)
            _, _ = rand.Read(data)

            b.SetBytes(int64(size))
            b.ResetTimer()
            b.ReportAllocs()

            for i := 0; i < b.N; i++ {
                _, err := FunctionUnderTest(data)
                if err != nil {
                    b.Fatal(err)
                }
            }
        })
    }
}
```

## Crypto-Specific: Key Generation Benchmarks

```go
func BenchmarkKeyGenerationCurves(b *testing.B) {
    curves := []struct {
        name  string
        curve elliptic.Curve
    }{
        {"P-256", elliptic.P256()},
        {"P-384", elliptic.P384()},
        {"P-521", elliptic.P521()},
    }

    for _, tc := range curves {
        tc := tc
        b.Run(tc.name, func(b *testing.B) {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _, err := ecdsa.GenerateKey(tc.curve, rand.Reader)
                if err != nil {
                    b.Fatal(err)
                }
            }
        })
    }
}
```
