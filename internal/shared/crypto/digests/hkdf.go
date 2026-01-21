// Copyright (c) 2025 Justin Cranford
//
//

package digests

import (
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"

	"golang.org/x/crypto/hkdf"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// Error variables for HKDF validation.
var (
	// ErrInvalidNilDigestFunction indicates that a nil digest function was provided.
	ErrInvalidNilDigestFunction         = errors.New("digest function can't be nil; supported options are SHA512, SHA384, SHA256, SHA224")
	ErrInvalidNilSecret                 = errors.New("secret can't be nil; generate a random value, and protect it")
	ErrInvalidEmptySecret               = errors.New("secret can't be empty; generate a random value, and protect it")
	ErrInvalidOutputBytesLengthNegative = errors.New("outputBytesLength can't be negative; minimum should be 1 * digest block size, but can be truncated for some use cases")
	ErrInvalidOutputBytesLengthZero     = errors.New("outputBytesLength can't be zero; minimum should be 1 * digest block size, but can be truncated for some use cases")
	ErrInvalidOutputBytesLengthTooBig   = errors.New("outputBytesLength too big; maximum is 255 * digest block size")
)

// Digest name constants for HKDF operations.
const (
	// DigestSHA512 is the name constant for SHA-512 digest.
	DigestSHA512 = cryptoutilMagic.SHA512
	DigestSHA384 = cryptoutilMagic.SHA384
	DigestSHA256 = cryptoutilMagic.SHA256
	DigestSHA224 = cryptoutilMagic.SHA224
)

// HKDFwithSHA512 performs HKDF key derivation using SHA-512 digest.
func HKDFwithSHA512(secret, salt, info []byte, outputBytesLength int) ([]byte, error) {
	return HKDF("SHA512", secret, salt, info, outputBytesLength)
}

// HKDFwithSHA384 performs HKDF key derivation using SHA-384 digest.
func HKDFwithSHA384(secret, salt, info []byte, outputBytesLength int) ([]byte, error) {
	return HKDF("SHA384", secret, salt, info, outputBytesLength)
}

// HKDFwithSHA256 performs HKDF key derivation using SHA-256 digest.
func HKDFwithSHA256(secret, salt, info []byte, outputBytesLength int) ([]byte, error) {
	return HKDF("SHA256", secret, salt, info, outputBytesLength)
}

// HKDFwithSHA224 performs HKDF key derivation using SHA-224 digest.
func HKDFwithSHA224(secret, salt, info []byte, outputBytesLength int) ([]byte, error) {
	return HKDF("SHA224", secret, salt, info, outputBytesLength)
}

// HKDF Supported digestNames: "SHA512", "SHA384", "SHA256", "SHA224".
func HKDF(digestName string, secretBytes, saltBytes, infoBytes []byte, outputBytesLength int) ([]byte, error) {
	var digestFunction func() hash.Hash

	var digestLength int

	switch digestName {
	case DigestSHA512:
		digestFunction = sha512.New
		digestLength = 64
	case DigestSHA384:
		digestFunction = sha512.New384
		digestLength = 48
	case DigestSHA256:
		digestFunction = sha256.New
		digestLength = 32
	case DigestSHA224:
		// FIPS 140-2/140-3 compliance: Use full SHA-256 instead of SHA-224
		digestFunction = sha256.New
		digestLength = 32
	default:
		return nil, fmt.Errorf("invalid digest name: %s. %w", digestName, ErrInvalidNilDigestFunction)
	}

	var errs []error
	if secretBytes == nil { // pragma: allowlist secret
		errs = append(errs, ErrInvalidNilSecret)
	} else if len(secretBytes) == 0 {
		errs = append(errs, ErrInvalidEmptySecret)
	}

	if outputBytesLength < 0 {
		errs = append(errs, ErrInvalidOutputBytesLengthNegative)
	} else if outputBytesLength == 0 {
		errs = append(errs, ErrInvalidOutputBytesLengthZero)
	} else if outputBytesLength > 255*digestLength {
		errs = append(errs, ErrInvalidOutputBytesLengthTooBig)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid parameters for HKDF: %w", errors.Join(errs...))
	}

	hkdfAlgorithm := hkdf.New(digestFunction, secretBytes, saltBytes, infoBytes)
	hkdfOutputBytes := make([]byte, outputBytesLength)

	_, err := hkdfAlgorithm.Read(hkdfOutputBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compute HKDF: %w", err)
	}

	return hkdfOutputBytes, nil
}
