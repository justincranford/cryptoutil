// Copyright (c) 2025 Justin Cranford
//
//

package util

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

func GenerateString(length int) (string, error) {
	bytesNeeded := (length + 1) / 2

	randomBytes, err := GenerateBytes(bytesNeeded)
	if err != nil {
		return "", fmt.Errorf("failed to generate %d random bytes for string of length %d: %w", bytesNeeded, length, err)
	}

	return hex.EncodeToString(randomBytes)[:length], nil
}

func GenerateBytes(lengthBytes int) ([]byte, error) {
	bytes := make([]byte, lengthBytes)

	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate %d bytes: %w", lengthBytes, err)
	}

	return bytes, nil
}

func GenerateMultipleBytes(count, lengthBytes int) ([][]byte, error) {
	if count < 1 {
		return nil, fmt.Errorf("count can't be less than 1")
	} else if lengthBytes < 1 {
		return nil, fmt.Errorf("length can't be less than 1")
	}

	concatSharedSecrets := make([]byte, count*lengthBytes) // max 255 * 64

	_, err := rand.Read(concatSharedSecrets)
	if err != nil {
		return nil, fmt.Errorf("failed to generate consecutive byte slices: %w", err)
	}

	nBytes := make([][]byte, count)

	for i := range count {
		startOffset := i * lengthBytes
		nBytes[i] = concatSharedSecrets[startOffset : startOffset+lengthBytes]
	}

	return nBytes, nil
}

func ConcatBytes(list [][]byte) []byte {
	var combined []byte
	for _, b := range list {
		combined = append(combined, b...)
	}

	return combined
}

func StringsToBytes(values ...string) [][]byte {
	result := make([][]byte, 0, len(values))
	for _, s := range values {
		result = append(result, []byte(s))
	}

	return result
}

func StringPointersToBytes(values ...*string) [][]byte {
	var result [][]byte

	for _, s := range values {
		if s != nil {
			result = append(result, []byte(*s))
		}
	}

	return result
}

func Uint64ToBytes(val uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, val)

	return bytes
}

func Uint32ToBytes(val uint32) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, val)

	return bytes
}

func Uint16ToBytes(val uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, val)

	return bytes
}

// safeIntToUint64 safely converts int64 to uint64 preserving bit pattern.
// This conversion is always safe for same-width signed/unsigned types.
func safeIntToUint64(val int64) uint64 {
	// gosec G115: This conversion is safe - same bit width preserves all values
	// Negative values are preserved in two's complement representation
	return uint64(val) // #nosec G115
}

// safeIntToUint32 safely converts int32 to uint32 preserving bit pattern.
// This conversion is always safe for same-width signed/unsigned types.
func safeIntToUint32(val int32) uint32 {
	// gosec G115: This conversion is safe - same bit width preserves all values
	// Negative values are preserved in two's complement representation
	return uint32(val) // #nosec G115
}

// safeIntToUint16 safely converts int16 to uint16 preserving bit pattern.
// This conversion is always safe for same-width signed/unsigned types.
func safeIntToUint16(val int16) uint16 {
	// gosec G115: This conversion is safe - same bit width preserves all values
	// Negative values are preserved in two's complement representation
	return uint16(val) // #nosec G115
}

func Int64ToBytes(val int64) []byte {
	// Safe conversion: int64 to uint64 preserves bit pattern for signed/unsigned of same width
	// This conversion is always safe as both types use the same bit representation
	return Uint64ToBytes(safeIntToUint64(val))
}

func Int32ToBytes(val int32) []byte {
	// Safe conversion: int32 to uint32 preserves bit pattern for signed/unsigned of same width
	// This conversion is always safe as both types use the same bit representation
	return Uint32ToBytes(safeIntToUint32(val))
}

func Int16ToBytes(val int16) []byte {
	// Safe conversion: int16 to uint16 preserves bit pattern for signed/unsigned of same width
	// This conversion is always safe as both types use the same bit representation
	return Uint16ToBytes(safeIntToUint16(val))
}
