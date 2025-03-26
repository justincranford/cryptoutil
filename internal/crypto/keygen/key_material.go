package keygen

import (
	"fmt"
)

func GenerateKeyMaterial(algorithm string) ([]byte, error) {
	var key Key
	var err error
	switch string(algorithm) {
	case "AES-256", "AES256":
		key, err = GenerateAESKey(256)
	case "AES-192", "AES192":
		key, err = GenerateAESKey(192)
	case "AES-128", "AES128":
		key, err = GenerateAESKey(128)
	default:
		return nil, fmt.Errorf("unsuppported algorithm")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate key material: %w", err)
	}
	return key.Private.([]byte), nil
}
