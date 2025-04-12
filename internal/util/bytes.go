package util

import "encoding/binary"

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
