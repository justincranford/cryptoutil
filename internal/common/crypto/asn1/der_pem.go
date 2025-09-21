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
	PemTypePkcs8PrivateKey = "PRIVATE KEY"
	PemTypePkixPublicKey   = "PUBLIC KEY"
	PemTypeRsaPrivateKey   = "RSA PRIVATE KEY"
	PemTypeRsaPublicKey    = "RSA PUBLIC KEY"
	PemTypeEcPrivateKey    = "EC PRIVATE KEY"
	PemTypeCertificate     = "CERTIFICATE"
	PemTypeCsr             = "CERTIFICATE REQUEST"
	PemTypeSecretKey       = "SECRET KEY"
)

var PemTypes = []string{
	PemTypePkcs8PrivateKey, PemTypePkixPublicKey, PemTypeRsaPrivateKey, PemTypeRsaPublicKey, PemTypeEcPrivateKey, PemTypeCertificate, PemTypeCsr, PemTypeSecretKey,
}

func PemEncodes(keys any) ([][]byte, error) {
	switch expression := keys.(type) {
	case []*x509.Certificate:
		var pemBytesList [][]byte
		for _, k := range expression {
			pemBytes, err := PemEncode(k)
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

func DerEncodes(key any) ([][]byte, error) {
	var derBytesList [][]byte
	switch expression := key.(type) {
	case []*x509.Certificate:
		for _, k := range expression {
			derBytes, _, err := DerEncode(k)
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

func PemEncode(key any) ([]byte, error) {
	derBytes, pemType, err := DerEncode(key)
	if err != nil {
		return nil, fmt.Errorf("encode failed: %w", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Bytes: derBytes, Type: pemType})
	return pemBytes, nil
}

func DerEncode(key any) ([]byte, string, error) {
	switch x509Type := key.(type) {
	case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey, *ecdh.PrivateKey:
		privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(x509Type)
		if err != nil {
			return nil, "", fmt.Errorf("encode failed: %w", err)
		}
		return privateKeyBytes, PemTypePkcs8PrivateKey, nil
	case *rsa.PublicKey, *ecdsa.PublicKey, ed25519.PublicKey, *ecdh.PublicKey:
		publicKeyBytes, err := x509.MarshalPKIXPublicKey(x509Type)
		if err != nil {
			return nil, "", fmt.Errorf("encode failed: %w", err)
		}
		return publicKeyBytes, PemTypePkixPublicKey, nil
	case *x509.Certificate:
		return x509Type.Raw, PemTypeCertificate, nil
	case *x509.CertificateRequest:
		return x509Type.Raw, PemTypeCsr, nil
	case []byte:
		byteKey, ok := key.([]byte)
		if !ok {
			return nil, "", fmt.Errorf("type assertion to []byte failed")
		}
		return byteKey, PemTypeSecretKey, nil
	default:
		return nil, "", fmt.Errorf("not supported [%T]", x509Type)
	}
}

func DerDecode(bytes []byte, x509Type string) (any, error) {
	var key any
	var err error
	switch x509Type {
	case PemTypePkcs8PrivateKey:
		key, err = x509.ParsePKCS8PrivateKey(bytes) // Generic: RSA, EC, ED
	case PemTypePkixPublicKey:
		key, err = x509.ParsePKIXPublicKey(bytes) // Generic: RSA, EC, ED
	case PemTypeRsaPrivateKey:
		key, err = x509.ParsePKCS1PrivateKey(bytes) // RSA PrivateKey
	case PemTypeRsaPublicKey:
		key, err = x509.ParsePKCS1PublicKey(bytes) // RSA PublicKey
	case PemTypeEcPrivateKey:
		key, err = x509.ParseECPrivateKey(bytes) // EC, ED PrivateKey
	case PemTypeCertificate:
		key, err = x509.ParseCertificate(bytes)
	case PemTypeCsr:
		key, err = x509.ParseCertificateRequest(bytes)
	case PemTypeSecretKey:
		key, err = bytes, nil // AES, HMAC, AES-HMAC
	default:
		return nil, fmt.Errorf("type not supported: %s", x509Type)
	}
	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}
	return key, nil
}

func DerDecodes(bytes []byte) (any, string, error) {
	for _, derType := range PemTypes {
		key, err := DerDecode(bytes, derType)
		if err == nil {
			return key, derType, nil
		}
	}
	return nil, "", fmt.Errorf("decode failed")
}

func PemDecode(bytes []byte) (any, error) {
	block, rest := pem.Decode(bytes)
	_ = rest // Intentionally ignore remaining bytes after PEM block
	if block == nil {
		return nil, fmt.Errorf("parse PEM failed")
	}
	return DerDecode(block.Bytes, block.Type)
}

func PemRead(filename string) (any, error) {
	pemBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return PemDecode(pemBytes)
}

func DerRead(filename string) (any, string, error) {
	derBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, "", fmt.Errorf("read failed: %w", err)
	}

	key, derType, err := DerDecodes(derBytes)
	if err != nil {
		return nil, "", fmt.Errorf("decode failed: %w", err)
	}
	return key, derType, nil
}

func PemWrite(key any, filename string) error {
	pemBytes, err := PemEncode(key)
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

func DerWrite(key any, filename string) error {
	derBytes, _, err := DerEncode(key)
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
