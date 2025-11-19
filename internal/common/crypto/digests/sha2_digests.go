// Copyright (c) 2025 Justin Cranford
//
//

package digests

import (
	"crypto/sha256"
	"crypto/sha512"
)

func SHA512(bytes []byte) []byte {
	digest := sha512.Sum512(bytes)

	return digest[:]
}

func SHA384(bytes []byte) []byte {
	digest := sha512.Sum384(bytes)

	return digest[:]
}

func SHA256(bytes []byte) []byte {
	digest := sha256.Sum256(bytes)

	return digest[:]
}

func SHA224(bytes []byte) []byte {
	digest := sha256.Sum224(bytes)

	return digest[:]
}
