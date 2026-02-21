// Copyright (c) 2025 Justin Cranford
//
//

package random

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	googleUuid "github.com/google/uuid"
)

// uuidNewV7 is the UUID v7 generator — injectable for testing error paths.
var uuidNewV7 = googleUuid.NewV7

// globalRandReader is the random reader — injectable for testing error paths.
var globalRandReader io.Reader = crand.Reader

// GenerateUsernameSimple generates a random username with "user_" prefix and full UUID suffix.
// Uses UUIDv7 for time-ordered uniqueness and concurrency safety.
// Returns full UUID (36 chars) to prevent collisions in parallel test execution.
// For test scenarios requiring specific lengths, use GenerateUsername(t, length) instead.
func GenerateUsernameSimple() (string, error) {
	id, err := uuidNewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID for username: %w", err)
	}

	return "user_" + id.String(), nil
}

// GeneratePasswordSimple generates a random password with "pass_" prefix and full UUID suffix.
// Provides sufficient entropy for test passwords while maintaining readability.
// For test scenarios requiring specific lengths, use GeneratePassword(t, length) instead.
func GeneratePasswordSimple() (string, error) {
	id, err := uuidNewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID for password: %w", err)
	}

	return "pass_" + id.String(), nil
}

// GenerateString generates a random hexadecimal string of the specified length.
func GenerateString(length int) (string, error) {
	bytesNeeded := (length + 1) / 2

	randomBytes, err := GenerateBytes(bytesNeeded)
	if err != nil {
		return "", fmt.Errorf("failed to generate %d random bytes for string of length %d: %w", bytesNeeded, length, err)
	}

	return hex.EncodeToString(randomBytes)[:length], nil
}

// GenerateBytes generates a slice of random bytes of the specified length.
func GenerateBytes(lengthBytes int) ([]byte, error) {
	bytes := make([]byte, lengthBytes)

	_, err := io.ReadFull(globalRandReader, bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate %d bytes: %w", lengthBytes, err)
	}

	return bytes, nil
}

// GenerateMultipleBytes generates multiple slices of random bytes.
func GenerateMultipleBytes(count, lengthBytes int) ([][]byte, error) {
	if count < 1 {
		return nil, fmt.Errorf("count can't be less than 1")
	} else if lengthBytes < 1 {
		return nil, fmt.Errorf("length can't be less than 1")
	}

	concatSharedSecrets := make([]byte, count*lengthBytes) // max 255 * 64

	_, err := io.ReadFull(globalRandReader, concatSharedSecrets)
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

// ConcatBytes concatenates multiple byte slices into a single slice.
func ConcatBytes(list [][]byte) []byte {
	var combined []byte
	for _, b := range list {
		combined = append(combined, b...)
	}

	return combined
}

// StringsToBytes converts multiple strings to byte slices.
func StringsToBytes(values ...string) [][]byte {
	result := make([][]byte, 0, len(values))
	for _, s := range values {
		result = append(result, []byte(s))
	}

	return result
}

// StringPointersToBytes converts string pointers to byte slices, skipping nil pointers.
func StringPointersToBytes(values ...*string) [][]byte {
	var result [][]byte

	for _, s := range values {
		if s != nil {
			result = append(result, []byte(*s))
		}
	}

	return result
}

// Uint64ToBytes converts a uint64 value to a big-endian byte slice.
func Uint64ToBytes(val uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, val)

	return bytes
}

// Uint32ToBytes converts a uint32 value to a big-endian byte slice.
func Uint32ToBytes(val uint32) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, val)

	return bytes
}

// Uint16ToBytes converts a uint16 value to a big-endian byte slice.
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

// Int64ToBytes converts an int64 value to a big-endian byte slice.
func Int64ToBytes(val int64) []byte {
	// Safe conversion: int64 to uint64 preserves bit pattern for signed/unsigned of same width
	// This conversion is always safe as both types use the same bit representation
	return Uint64ToBytes(safeIntToUint64(val))
}

// Int32ToBytes converts an int32 value to a big-endian byte slice.
func Int32ToBytes(val int32) []byte {
	// Safe conversion: int32 to uint32 preserves bit pattern for signed/unsigned of same width
	// This conversion is always safe as both types use the same bit representation
	return Uint32ToBytes(safeIntToUint32(val))
}

// Int16ToBytes converts an int16 value to a big-endian byte slice.
func Int16ToBytes(val int16) []byte {
	// Safe conversion: int16 to uint16 preserves bit pattern for signed/unsigned of same width
	// This conversion is always safe as both types use the same bit representation
	return Uint16ToBytes(safeIntToUint16(val))
}
