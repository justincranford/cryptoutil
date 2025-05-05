package keygen

import (
	"crypto/rand"
	"fmt"
)

func GenerateSharedSecrets(secretBytesCount, secretBytesLength int) ([][]byte, error) {
	if secretBytesCount == 0 {
		return nil, fmt.Errorf("secretBytes count can't be zero")
	} else if secretBytesCount < 0 {
		return nil, fmt.Errorf("secretBytes count can't be negative")
	} else if secretBytesCount >= 256 {
		return nil, fmt.Errorf("secretBytes count can't be greater than 256")
	} else if secretBytesLength < 32 {
		return nil, fmt.Errorf("secretBytes length can't be greater than 32")
	} else if secretBytesLength > 64 {
		return nil, fmt.Errorf("secretBytes length can't be greater than 64")
	}

	concatSharedSecrets := make([]byte, secretBytesCount*secretBytesLength) // max 255 * 64
	if _, err := rand.Read(concatSharedSecrets); err != nil {
		return nil, fmt.Errorf("failed to generate concatenated shared secrets: %w", err)
	}

	sharedSecrets := make([][]byte, secretBytesCount)
	for i := range secretBytesCount {
		startOffset := i * secretBytesLength
		sharedSecrets[i] = concatSharedSecrets[startOffset : startOffset+secretBytesLength]
	}
	return sharedSecrets, nil
}
