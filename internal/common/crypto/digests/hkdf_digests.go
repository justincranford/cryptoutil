package digests

import (
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"

	"golang.org/x/crypto/hkdf"
)

var (
	ErrInvalidNilDigestFunction         = errors.New("digest function can't be nil; supported options are SHA512, SHA384, SHA256, SHA224")
	ErrInvalidNilSecret                 = errors.New("secret can't be nil; generate a random value, and protect it")
	ErrInvalidEmptySecret               = errors.New("secret can't be empty; generate a random value, and protect it")
	ErrInvalidOutputBytesLengthNegative = errors.New("outputBytesLength can't be negative; minimum should be 1 * digest block size, but can be truncated for some use cases")
	ErrInvalidOutputBytesLengthZero     = errors.New("outputBytesLength can't be zero; minimum should be 1 * digest block size, but can be truncated for some use cases")
	ErrInvalidOutputBytesLengthTooBig   = errors.New("outputBytesLength too big; maximum is 255 * digest block size")
)

const (
	DigestSHA512 = "SHA512" // pragma: allowlist secret
	DigestSHA384 = "SHA384" // pragma: allowlist secret
	DigestSHA256 = "SHA256" // pragma: allowlist secret
	DigestSHA224 = "SHA224" // pragma: allowlist secret
)

func HKDFwithSHA512(secret, salt, info []byte, outputBytesLength int) ([]byte, error) {
	return HKDF("SHA512", secret, salt, info, outputBytesLength)
}

func HKDFwithSHA384(secret, salt, info []byte, outputBytesLength int) ([]byte, error) {
	return HKDF("SHA384", secret, salt, info, outputBytesLength)
}

func HKDFwithSHA256(secret, salt, info []byte, outputBytesLength int) ([]byte, error) {
	return HKDF("SHA256", secret, salt, info, outputBytesLength)
}

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
		digestFunction = sha256.New224
		digestLength = 28
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
