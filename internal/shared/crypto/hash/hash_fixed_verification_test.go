// Copyright (c) 2025-2026 Justin Cranford.
package hash

import (
	"errors"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testMalformedStoredHash       = "malformed"
	testInvalidStoredHashFormat   = "invalid hash format"
	testUnknownStoredHash         = "unknown$YWJj"
	testUnsupportedStoredHashAlg  = "unsupported algorithm"
	testStoredHashDecodeFailure   = "failed to decode stored derived key"
	testProvidedHashFailure       = "failed to hash provided secret"
	testInvalidProvidedHashFormat = "invalid provided hash format"
	testProvidedHashDecodeFailure = "failed to decode provided derived key"
)

func TestVerifyFixedHKDF_InvalidStoredHashes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		verify     func(string, string) (bool, error)
		storedHash string
		wantError  string
	}{
		{
			name:       "low entropy missing delimiter",
			verify:     VerifySecretHKDFFixed,
			storedHash: testMalformedStoredHash,
			wantError:  testInvalidStoredHashFormat,
		},
		{
			name:       "low entropy unsupported algorithm",
			verify:     VerifySecretHKDFFixed,
			storedHash: testUnknownStoredHash,
			wantError:  testUnsupportedStoredHashAlg,
		},
		{
			name:       "low entropy invalid base64",
			verify:     VerifySecretHKDFFixed,
			storedHash: cryptoutilSharedMagic.HKDFFixedLowHashName + "$!",
			wantError:  testStoredHashDecodeFailure,
		},
		{
			name:       "high entropy missing delimiter",
			verify:     VerifySecretHKDFFixedHigh,
			storedHash: testMalformedStoredHash,
			wantError:  testInvalidStoredHashFormat,
		},
		{
			name:       "high entropy unsupported algorithm",
			verify:     VerifySecretHKDFFixedHigh,
			storedHash: testUnknownStoredHash,
			wantError:  testUnsupportedStoredHashAlg,
		},
		{
			name:       "high entropy invalid base64",
			verify:     VerifySecretHKDFFixedHigh,
			storedHash: cryptoutilSharedMagic.HKDFFixedHighHashName + "$!",
			wantError:  testStoredHashDecodeFailure,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.verify(tc.storedHash, "provided-secret")
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestVerifyFixedHKDF_ProvidedHashFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		verify     func(string, string, func(string, []byte) (string, error)) (bool, error)
		storedHash string
		hashResult string
		hashErr    error
		wantError  string
	}{
		{
			name:       "low entropy hash function error",
			verify:     verifySecretHKDFFixedWithHashFn,
			storedHash: cryptoutilSharedMagic.HKDFFixedLowHashName + "$YWJj",
			hashErr:    errors.New("injected HKDF error"),
			wantError:  testProvidedHashFailure,
		},
		{
			name:       "low entropy malformed provided hash",
			verify:     verifySecretHKDFFixedWithHashFn,
			storedHash: cryptoutilSharedMagic.HKDFFixedLowHashName + "$YWJj",
			hashResult: testMalformedStoredHash,
			wantError:  testInvalidProvidedHashFormat,
		},
		{
			name:       "low entropy invalid provided key encoding",
			verify:     verifySecretHKDFFixedWithHashFn,
			storedHash: cryptoutilSharedMagic.HKDFFixedLowHashName + "$YWJj",
			hashResult: cryptoutilSharedMagic.HKDFFixedLowHashName + "$!",
			wantError:  testProvidedHashDecodeFailure,
		},
		{
			name:       "high entropy hash function error",
			verify:     verifySecretHKDFFixedHighWithHashFn,
			storedHash: cryptoutilSharedMagic.HKDFFixedHighHashName + "$YWJj",
			hashErr:    errors.New("injected HKDF error"),
			wantError:  testProvidedHashFailure,
		},
		{
			name:       "high entropy malformed provided hash",
			verify:     verifySecretHKDFFixedHighWithHashFn,
			storedHash: cryptoutilSharedMagic.HKDFFixedHighHashName + "$YWJj",
			hashResult: testMalformedStoredHash,
			wantError:  testInvalidProvidedHashFormat,
		},
		{
			name:       "high entropy invalid provided key encoding",
			verify:     verifySecretHKDFFixedHighWithHashFn,
			storedHash: cryptoutilSharedMagic.HKDFFixedHighHashName + "$YWJj",
			hashResult: cryptoutilSharedMagic.HKDFFixedHighHashName + "$!",
			wantError:  testProvidedHashDecodeFailure,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hashFn := func(string, []byte) (string, error) {
				return tc.hashResult, tc.hashErr
			}
			_, err := tc.verify(tc.storedHash, "provided-secret", hashFn)
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}
