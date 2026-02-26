// Copyright (c) 2025 Justin Cranford
//
//

package digests

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestCaseHKDFHappyPath struct {
	name              string
	digestName        string
	secret            []byte
	salt              []byte
	info              []byte
	outputBytesLength int
}

type TestCaseHKDFSadPath struct {
	name              string
	digestName        string
	secret            []byte
	salt              []byte
	info              []byte
	outputBytesLength int
	expectedError     error
}

func TestHKDFHappyPath(t *testing.T) {
	t.Parallel()

	happyPathTests := []TestCaseHKDFHappyPath{
		{"Valid SHA512", cryptoutilSharedMagic.SHA512, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.MinSerialNumberBits},
		{"Valid SHA384", cryptoutilSharedMagic.SHA384, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HMACSHA384KeySize},
		{"Valid SHA256", cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes},
		{"Valid SHA224", cryptoutilSharedMagic.SHA224, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFSHA224OutputLength},
		{"Max Output Length SHA256", cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFMaxMultiplier * cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes},
		{"Max Output Length SHA512", cryptoutilSharedMagic.SHA512, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFMaxMultiplier * cryptoutilSharedMagic.MinSerialNumberBits},
	}

	for _, tt := range happyPathTests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := HKDF(tt.digestName, tt.secret, tt.salt, tt.info, tt.outputBytesLength)
			require.NoError(t, err, "HKDF should not fail with valid input")

			require.Len(t, output, tt.outputBytesLength, "HKDF should return output of correct length")
		})
	}

	t.Run("Unique Output for Different Salts", func(t *testing.T) {
		output1, err := HKDF(cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt1"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
		require.NoError(t, err, "HKDF should not fail")

		output2, err := HKDF(cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt2"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
		require.NoError(t, err, "HKDF should not fail")

		require.NotEqual(t, output1, output2, "HKDF output should be unique for different salts")
	})
}

func TestHKDFHappyPathDifferentDigest(t *testing.T) {
	t.Parallel()
	// NOTE: SHA224 uses SHA-256 internally for FIPS 140-2/140-3 compliance (see hkdf.go).
	// Therefore SHA224 and SHA256 produce the same output. Only test SHA256, SHA384, SHA512.
	output1, err := HKDF(cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	require.NoError(t, err, "HKDF SHA256 should not fail")

	output2, err := HKDF(cryptoutilSharedMagic.SHA384, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	require.NoError(t, err, "HKDF SHA384 should not fail")

	output3, err := HKDF(cryptoutilSharedMagic.SHA512, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	require.NoError(t, err, "HKDF SHA512 should not fail")

	require.NotEqual(t, output1, output2, "HKDF output should be unique for different digests")
	require.NotEqual(t, output1, output3, "HKDF output should be unique for different digests")
	require.NotEqual(t, output2, output3, "HKDF output should be unique for different digests")
}

func TestHKDFHappyPathDifferentSecret(t *testing.T) {
	t.Parallel()

	output1, err := HKDF(cryptoutilSharedMagic.SHA256, []byte("secret1"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	require.NoError(t, err, "HKDF with secret1 should not fail")

	output2, err := HKDF(cryptoutilSharedMagic.SHA256, []byte("secret2"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	require.NoError(t, err, "HKDF with secret2 should not fail")

	require.NotEqual(t, output1, output2, "HKDF output should be unique for different secrets")
}

func TestHKDFHappyPathDifferentSalt(t *testing.T) {
	t.Parallel()

	output1, err := HKDF(cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt1"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	require.NoError(t, err, "HKDF with salt1 should not fail")

	output2, err := HKDF(cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt2"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	require.NoError(t, err, "HKDF with salt2 should not fail")

	require.NotEqual(t, output1, output2, "HKDF output should be unique for different salts")
}

func TestHKDFHappyPathDifferentInfo(t *testing.T) {
	t.Parallel()

	output1, err := HKDF(cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt"), []byte("info1"), cryptoutilSharedMagic.HKDFSHA224OutputLength)
	require.NoError(t, err, "HKDF with info1 should not fail")

	output2, err := HKDF(cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt"), []byte("info2"), cryptoutilSharedMagic.HKDFSHA224OutputLength)
	require.NoError(t, err, "HKDF with info2 should not fail")

	require.NotEqual(t, output1, output2, "HKDF output should be unique for different info")
}

func TestHKDFSadPath(t *testing.T) {
	t.Parallel()

	sadPathTests := []TestCaseHKDFSadPath{
		{"Invalid Digest Name", "InvalidDigest", []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes, ErrInvalidNilDigestFunction},
		{"Nil Secret", cryptoutilSharedMagic.SHA256, nil, []byte("salt"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes, ErrInvalidNilSecret},
		{"Empty Secret", cryptoutilSharedMagic.SHA256, []byte{}, []byte("salt"), []byte("info"), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes, ErrInvalidEmptySecret},
		// {"Nil Salt", "SHA256", []byte("secret"), nil, []byte("info"), 32, ErrInvalidNilSalt},
		// {"Empty Salt", "SHA256", []byte("secret"), []byte{}, []byte("info"), 32, ErrInvalidEmptySalt},
		// {"Nil Info", "SHA256", []byte("secret"), []byte("salt"), nil, 32, ErrInvalidNilInfo},
		// {"Empty Info", "SHA256", []byte("secret"), []byte("salt"), []byte{}, 32, ErrInvalidEmptyInfo},
		{"Negative Output Length", cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt"), []byte("info"), -1, ErrInvalidOutputBytesLengthNegative},
		{"Zero Output Length", cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt"), []byte("info"), 0, ErrInvalidOutputBytesLengthZero},
		// NOTE: SHA224 uses SHA-256 internally for FIPS compliance, so its max output is 255*32=8160, same as SHA256.
		{"Excessive Output Length SHA224", cryptoutilSharedMagic.SHA224, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFMaxMultiplier*cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes + 1, ErrInvalidOutputBytesLengthTooBig},
		{"Excessive Output Length SHA256", cryptoutilSharedMagic.SHA256, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFMaxMultiplier*cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes + 1, ErrInvalidOutputBytesLengthTooBig},
		{"Excessive Output Length SHA384", cryptoutilSharedMagic.SHA384, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFMaxMultiplier*cryptoutilSharedMagic.HMACSHA384KeySize + 1, ErrInvalidOutputBytesLengthTooBig},
		{"Excessive Output Length SHA512", cryptoutilSharedMagic.SHA512, []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFMaxMultiplier*cryptoutilSharedMagic.MinSerialNumberBits + 1, ErrInvalidOutputBytesLengthTooBig},
	}

	for _, tt := range sadPathTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HKDF(tt.digestName, tt.secret, tt.salt, tt.info, tt.outputBytesLength)
			require.Error(t, err, "HKDF should fail with invalid input")
			require.ErrorIs(t, err, tt.expectedError, "HKDF should return expected error")
		})
	}
}
