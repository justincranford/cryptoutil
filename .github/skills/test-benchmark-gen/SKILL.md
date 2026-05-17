---
name: test-benchmark-gen
description: "Generate _bench_test.go benchmark tests conforming to cryptoutil standards. Use when adding performance benchmarks, especially for crypto operations where benchmarking is mandatory, to ensure correct ResetTimer/StopTimer patterns and sub-benchmark structure."
argument-hint: "[package or function name]"
---

Generate `_bench_test.go` benchmark tests — mandatory for crypto operations.

## Purpose

Use when benchmarking performance-sensitive code, especially crypto operations.
Benchmarks go in a separate `_bench_test.go` file.

## Key Rules

- File suffix: `_bench_test.go` (ONLY benchmark functions)
- **MANDATORY** for: RSA/ECDSA/AES/HMAC operations, key generation, hashing
- `b.ResetTimer()` AFTER setup, BEFORE the benchmarked loop
- `b.StopTimer()` / `b.StartTimer()` when per-iteration setup is needed inside the loop
- `b.ReportAllocs()` for allocation-sensitive code
- `b.SetBytes(n)` for throughput measurement on crypto operations (AES, HMAC, etc.)
- Benchmark only the code under test; keep fixture creation, UUID generation, TLS setup, and other harness work outside the timed region unless that work is part of the behavior being measured
- Run benchmarks: `go test -bench=. -benchmem ./pkg/crypto/...`
- Compare baseline versus current output using the same package path, benchmark filter, and `-benchmem` settings

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
_, err := svc.DoOperation(ctx, staticID)
if err != nil {
b.Fatal(err)
}
}
}

// Throughput benchmark for streaming crypto operations (AES, HMAC)
func BenchmarkAESEncrypt(b *testing.B) {
const msgSize = 1024
key := make([]byte, 32)
plaintext := make([]byte, msgSize)
b.SetBytes(msgSize) // enables MB/s reporting
b.ReportAllocs()
b.ResetTimer()

for i := 0; i < b.N; i++ {
_, _ = encrypt(key, plaintext)
}
}

// Benchmark with per-iteration setup (use StopTimer/StartTimer)
func BenchmarkWithSetup(b *testing.B) {
b.ResetTimer()
for i := 0; i < b.N; i++ {
b.StopTimer()
input := prepareInput() // per-iteration setup NOT measured
b.StartTimer()
_, _ = processInput(input)
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

## Reading Regressions

Use the same command before and after a change so the comparison is meaningful:

```bash
go test -run '^$' -bench BenchmarkOperationName -benchmem ./path/to/pkg
```

Treat these as common noise sources before concluding there is a real regression:

- TLS handshake or listener startup happening inside the timed loop
- Fixture generation or random identifier creation inside the timed loop
- Garbage-collection pressure caused by avoidable allocations in the benchmark harness
- Comparing runs with different package scopes, CPU load, or benchmark filters

## References

Read [ENG-HANDBOOK.md Section 10.8 Benchmark Testing Strategy](../../../docs/ENG-HANDBOOK.md#108-benchmark-testing-strategy) for benchmarking requirements — apply all benchmark standards including mandatory `_bench_test.go` suffix, `ResetTimer`/`StopTimer` patterns, and crypto operation requirements.
