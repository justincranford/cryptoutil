package service

import (
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	"fmt"
)

func generateKeyMaterial(algorithm string) ([]byte, error) {
	var key cryptoutilKeygen.Key
	var err error
	switch string(algorithm) {
	case "AES-256", "AES256":
		key, err = cryptoutilKeygen.GenerateAESKey(256)
	case "AES-192", "AES192":
		key, err = cryptoutilKeygen.GenerateAESKey(192)
	case "AES-128", "AES128":
		key, err = cryptoutilKeygen.GenerateAESKey(128)
	default:
		return nil, fmt.Errorf("unsuppported algorithm")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate key material: %w", err)
	}
	return key.Private.([]byte), nil
}
