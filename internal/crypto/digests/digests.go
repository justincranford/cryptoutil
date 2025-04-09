package digests

import (
	"crypto/sha256"
	"crypto/sha512"
)

func Sha512(bytes []byte) []byte {
	digest := sha512.Sum512(bytes)
	return digest[:]
}

func Sha384(bytes []byte) []byte {
	digest := sha512.Sum384(bytes)
	return digest[:]
}

func Sha256(bytes []byte) []byte {
	digest := sha256.Sum256(bytes)
	return digest[:]
}

func Sha224(bytes []byte) []byte {
	digest := sha256.Sum224(bytes)
	return digest[:]
}
