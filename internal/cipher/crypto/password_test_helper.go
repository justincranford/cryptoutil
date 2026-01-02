// Copyright (c) 2025 Justin Cranford

package crypto

import (
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// HashPasswordForTest is a test helper that uses reduced iterations (1,000 instead of 600,000)
// for faster test execution. Production code MUST use HashPassword with 600,000 iterations.
//
// Performance: ~12ms (1,000 iterations) vs ~700ms (600,000 iterations) = 58× speedup.
//
// NOTE: This function is exported for testing purposes ONLY. Never use in production code.
func HashPasswordForTest(password string) ([]byte, error) {
	return hashPasswordWithIterations(password, cryptoutilMagic.PBKDF2V3Iterations)
}

// VerifyPasswordForTest is a test helper that uses reduced iterations (1,000 instead of 600,000)
// for faster test execution. Production code MUST use VerifyPassword with 600,000 iterations.
//
// Performance: ~12ms (1,000 iterations) vs ~700ms (600,000 iterations) = 58× speedup.
//
// NOTE: This function is exported for testing purposes ONLY. Never use in production code.
func VerifyPasswordForTest(password string, storedHash []byte) (bool, error) {
	return verifyPasswordWithIterations(password, storedHash, cryptoutilMagic.PBKDF2V3Iterations)
}
