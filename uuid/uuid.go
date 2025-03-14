package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// generateUniqueID generates a UUIDv7 value based on the draft UUIDv7 specification.
func V7() string {
	uuid := make([]byte, 16)
	now := time.Now().UnixMilli()

	// Write the timestamp (48 bits) into the first 6 bytes
	binary.BigEndian.PutUint64(uuid[0:8], uint64(now))
	uuid[0] &= 0x0F // Clear the first 4 bits
	uuid[0] |= 0x70 // Set the version to 0111 (UUIDv7)

	// Fill the remaining 10 bytes with random data
	_, err := rand.Read(uuid[6:])
	if err != nil {
		panic(fmt.Sprintf("failed to generate random bytes: %v", err))
	}

	// Set the variant to RFC 4122
	uuid[8] &= 0x3F // Clear the first 2 bits
	uuid[8] |= 0x80 // Set the variant to 10

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(uuid[0:4]),
		binary.BigEndian.Uint16(uuid[4:6]),
		binary.BigEndian.Uint16(uuid[6:8]),
		binary.BigEndian.Uint16(uuid[8:10]),
		uuid[10:])
}
