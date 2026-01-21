// Copyright (c) 2025 Justin Cranford
//
//

package digests

import (
	"crypto/sha256"
	"crypto/sha512"
)

// SHA512 computes the SHA-512 hash of the input bytes.
func SHA512(bytes []byte) []byte {
	digest := sha512.Sum512(bytes)

	return digest[:]
}

// SHA384 computes the SHA-384 hash of the input bytes.
func SHA384(bytes []byte) []byte {
	digest := sha512.Sum384(bytes)

	return digest[:]
}

// SHA256 computes the SHA-256 hash of the input bytes.
func SHA256(bytes []byte) []byte {
	digest := sha256.Sum256(bytes)

	return digest[:]
}

// SHA224 computes the SHA-224 hash of the input bytes.
func SHA224(bytes []byte) []byte {
	digest := sha256.Sum224(bytes)

	return digest[:]
}
