# test-benchmark-gen

Generate `_bench_test.go` benchmark tests — mandatory for crypto operations.

## Purpose

Use when benchmarking performance-sensitive code, especially crypto operations.
Benchmarks go in a separate `_bench_test.go` file.

## Key Rules

- File suffix: `_bench_test.go` (ONLY benchmark functions)
- **MANDATORY** for: RSA/ECDSA/AES/HMAC operations, key generation, hashing
- `b.ResetTimer()` AFTER setup, BEFORE the benchmarked loop
- `b.ReportAllocs()` for allocation-sensitive code
- Use `UUIDv7` for unique test identifiers per iteration

## Template

```go
package mypkg_test

import (
"testing"

googleUuid "github.com/google/uuid"
mypkg "cryptoutil/internal/path/to/mypkg"
)

func BenchmarkOperationName(b *testing.B) {
// Setup (not timed)
ctx := context.Background()
svc := mypkg.NewService()
b.ReportAllocs()
b.ResetTimer()

for i := 0; i < b.N; i++ {
id := googleUuid.NewV7().String() // unique per iteration
_, err := svc.DoOperation(ctx, id)
if err != nil {
b.Fatal(err)
}
}
}

// Benchmark table for multiple sizes/algorithms
func BenchmarkKeyGen(b *testing.B) {
cases := []struct{ name string; bits int }{
{"RSA-2048", 2048},
{"RSA-4096", 4096},
}
for _, tc := range cases {
b.Run(tc.name, func(b *testing.B) {
b.ResetTimer()
for i := 0; i < b.N; i++ {
_ = generateKey(tc.bits)
}
})
}
}
```

## References

See [ARCHITECTURE.md Section 10.8 Benchmark Testing Strategy](../../docs/ARCHITECTURE.md#108-benchmark-testing-strategy) for benchmarking requirements.
