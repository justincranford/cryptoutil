package util

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

func GenerateBytes(lengthBytes int) ([]byte, error) {
	bytes := make([]byte, lengthBytes)
	if _, err := rand.Read(bytes); err != nil {
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
	if _, err := rand.Read(concatSharedSecrets); err != nil {
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
	var result [][]byte
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

func Int64ToBytes(val int64) []byte {
	return Uint64ToBytes(uint64(val))
}

func Int32ToBytes(val int32) []byte {
	return Uint32ToBytes(uint32(val))
}

func Int16ToBytes(val int16) []byte {
	return Uint16ToBytes(uint16(val))
}
