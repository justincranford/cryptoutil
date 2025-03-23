package service

import (
	"cryptoutil/crypto/keygen"
	"fmt"
)

func generateKEKMaterial(gormKEKAlgorithm string) ([]byte, error) {
	var key keygen.Key
	var err error
	switch string(gormKEKAlgorithm) {
	case "AES-256", "AES256":
		key, err = keygen.GenerateAESKey(256)
	case "AES-192", "AES192":
		key, err = keygen.GenerateAESKey(192)
	case "AES-128", "AES128":
		key, err = keygen.GenerateAESKey(128)
	default:
		return nil, fmt.Errorf("invalid KEK Pool algorithm")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate KEK key material: %w", err)
	}
	return key.Private.([]byte), nil
}
