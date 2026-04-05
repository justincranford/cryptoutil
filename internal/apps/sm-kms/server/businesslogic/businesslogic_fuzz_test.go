// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"testing"

	googleUuid "github.com/google/uuid"
)

// IMPORTANT: All Fuzz* test function names MUST be unique and MUST NOT be substrings of any other fuzz test names.
// This ensures cross-platform compatibility with the -fuzz parameter (no quotes or regex needed).

// FuzzPostDecryptByElasticKeyIDBytes tests PostDecryptByElasticKeyID with arbitrary byte inputs.
// Verifies that the JWE parser and input-handling code never panics.
// The stack is created once per fuzz function (using f, not per-seed t) to avoid
// SQLite URI corruption caused by the '#' in seed test names like "seed#0".
func FuzzPostDecryptByElasticKeyIDBytes(f *testing.F) {
	stack := setupTestStack(f)

	// Seed with invalid inputs to help the fuzzer explore parser error paths.
	f.Add([]byte("not-a-jwe"))
	f.Add([]byte{})
	f.Add([]byte("invalid.compact.jwe.format.here"))

	f.Fuzz(func(t *testing.T, data []byte) {
		anyEKID := googleUuid.New()

		// Must not panic; errors are expected for arbitrary input.
		_, _ = stack.service.PostDecryptByElasticKeyID(stack.ctx, &anyEKID, data)
	})
}

// FuzzPostVerifyByElasticKeyIDBytes tests PostVerifyByElasticKeyID with arbitrary byte inputs.
// Verifies that the JWS parser and input-handling code never panics.
// The stack is created once per fuzz function (using f, not per-seed t) to avoid
// SQLite URI corruption caused by the '#' in seed test names like "seed#0".
func FuzzPostVerifyByElasticKeyIDBytes(f *testing.F) {
	stack := setupTestStack(f)

	// Seed with invalid inputs to help the fuzzer explore parser error paths.
	f.Add([]byte("not-a-jws"))
	f.Add([]byte{})
	f.Add([]byte("invalid.jws.compact.format"))

	f.Fuzz(func(t *testing.T, data []byte) {
		anyEKID := googleUuid.New()

		// Must not panic; errors are expected for arbitrary input.
		_, _ = stack.service.PostVerifyByElasticKeyID(stack.ctx, &anyEKID, data)
	})
}
