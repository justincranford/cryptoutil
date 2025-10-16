package asn1

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

const (
	PEMTypePKCS8PrivateKey = "PRIVATE KEY"
	PEMTypePKIXPublicKey   = "PUBLIC KEY"
	PEMTypeRSAPrivateKey   = "RSA PRIVATE KEY"
	PEMTypeRSAPublicKey    = "RSA PUBLIC KEY"
	PEMTypeECPrivateKey    = "EC PRIVATE KEY"
	PEMTypeCertificate     = "CERTIFICATE"
	PEMTypeCSR             = "CERTIFICATE REQUEST"
	PEMTypeSecretKey       = "SECRET KEY"
)

var PEMTypes = []string{
	PEMTypePKCS8PrivateKey, PEMTypePKIXPublicKey, PEMTypeRSAPrivateKey, PEMTypeRSAPublicKey, PEMTypeECPrivateKey, PEMTypeCertificate, PEMTypeCSR, PEMTypeSecretKey,
}

func PEMEncodes(keys any) ([][]byte, error) {
	switch expression := keys.(type) {
	case []*x509.Certificate:
		var pemBytesList [][]byte

		for _, k := range expression {
			pemBytes, err := PEMEncode(k)
			if err != nil {
				return nil, fmt.Errorf("encode failed: %w", err)
			}

			pemBytesList = append(pemBytesList, pemBytes)
		}

		return pemBytesList, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", keys)
	}
}

func DEREncodes(key any) ([][]byte, error) {
	var derBytesList [][]byte

	switch expression := key.(type) {
	case []*x509.Certificate:
		for _, k := range expression {
			derBytes, _, err := DEREncode(k)
			if err != nil {
				return nil, fmt.Errorf("encode failed: %w", err)
			}

			derBytesList = append(derBytesList, derBytes)
		}

		return derBytesList, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", key)
	}
}

func PEMEncode(key any) ([]byte, error) {
	derBytes, pemType, err := DEREncode(key)
	if err != nil {
		return nil, fmt.Errorf("encode failed: %w", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Bytes: derBytes, Type: pemType})

	return pemBytes, nil
}

func DEREncode(key any) ([]byte, string, error) {
	switch x509Type := key.(type) {
	case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey, *ecdh.PrivateKey:
		privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(x509Type)
		if err != nil {
			return nil, "", fmt.Errorf("encode failed: %w", err)
		}

		return privateKeyBytes, PEMTypePKCS8PrivateKey, nil
	case *rsa.PublicKey, *ecdsa.PublicKey, ed25519.PublicKey, *ecdh.PublicKey:
		publicKeyBytes, err := x509.MarshalPKIXPublicKey(x509Type)
		if err != nil {
			return nil, "", fmt.Errorf("encode failed: %w", err)
		}

		return publicKeyBytes, PEMTypePKIXPublicKey, nil
	case *x509.Certificate:
		return x509Type.Raw, PEMTypeCertificate, nil
	case *x509.CertificateRequest:
		return x509Type.Raw, PEMTypeCSR, nil
	case []byte:
		byteKey, ok := key.([]byte)
		if !ok {
			return nil, "", fmt.Errorf("type assertion to []byte failed")
		}

		return byteKey, PEMTypeSecretKey, nil
	default:
		return nil, "", fmt.Errorf("not supported [%T]", x509Type)
	}
}

func DERDecode(bytes []byte, x509Type string) (any, error) {
	var key any

	var err error

	switch x509Type {
	case PEMTypePKCS8PrivateKey:
		key, err = x509.ParsePKCS8PrivateKey(bytes) // Generic: RSA, EC, ED
	case PEMTypePKIXPublicKey:
		key, err = x509.ParsePKIXPublicKey(bytes) // Generic: RSA, EC, ED
	case PEMTypeRSAPrivateKey:
		key, err = x509.ParsePKCS1PrivateKey(bytes) // RSA PrivateKey
	case PEMTypeRSAPublicKey:
		key, err = x509.ParsePKCS1PublicKey(bytes) // RSA PublicKey
	case PEMTypeECPrivateKey:
		key, err = x509.ParseECPrivateKey(bytes) // EC, ED PrivateKey
	case PEMTypeCertificate:
		key, err = x509.ParseCertificate(bytes)
	case PEMTypeCSR:
		key, err = x509.ParseCertificateRequest(bytes)
	case PEMTypeSecretKey:
		key, err = bytes, nil // AES, HMAC, AES-HMAC
	default:
		return nil, fmt.Errorf("type not supported: %s", x509Type)
	}

	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return key, nil
}

func DERDecodes(bytes []byte) (any, string, error) {
	for _, derType := range PEMTypes {
		key, err := DERDecode(bytes, derType)
		if err == nil {
			return key, derType, nil
		}
	}

	return nil, "", fmt.Errorf("decode failed")
}

func PEMDecode(bytes []byte) (any, error) {
	block, rest := pem.Decode(bytes)
	_ = rest // Intentionally ignore remaining bytes after PEM block

	if block == nil {
		return nil, fmt.Errorf("parse PEM failed")
	}

	return DERDecode(block.Bytes, block.Type)
}

func PEMRead(filename string) (any, error) {
	pemBytes, err := os.ReadFile(filename) // #nosec G304 -- Legitimate file reading for crypto operations
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return PEMDecode(pemBytes)
}

func DERRead(filename string) (any, string, error) {
	derBytes, err := os.ReadFile(filename) // #nosec G304 -- Legitimate file reading for crypto operations
	if err != nil {
		return nil, "", fmt.Errorf("read failed: %w", err)
	}

	key, derType, err := DERDecodes(derBytes)
	if err != nil {
		return nil, "", fmt.Errorf("decode failed: %w", err)
	}

	return key, derType, nil
}

func PEMWrite(key any, filename string) error {
	pemBytes, err := PEMEncode(key)
	if err != nil {
		return fmt.Errorf("encode failed: %w", err)
	}

	dir := filepath.Dir(filename)

	err = os.MkdirAll(dir, 0o750)
	if err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	err = os.WriteFile(filename, pemBytes, 0o600)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}

func DERWrite(key any, filename string) error {
	derBytes, _, err := DEREncode(key)
	if err != nil {
		return fmt.Errorf("encode failed: %w", err)
	}

	dir := filepath.Dir(filename)

	err = os.MkdirAll(dir, 0o750)
	if err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	err = os.WriteFile(filename, derBytes, 0o600)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}
