// Copyright (c) 2025 Justin Cranford
//
//

package asn1

import (
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	rsa "crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// PEMTypes lists all supported PEM type identifiers for DER/PEM encoding.
var PEMTypes = []string{
	cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, cryptoutilSharedMagic.StringPEMTypePKIXPublicKey, cryptoutilSharedMagic.StringPEMTypeRSAPrivateKey, cryptoutilSharedMagic.StringPEMTypeRSAPublicKey, cryptoutilSharedMagic.StringPEMTypeECPrivateKey, cryptoutilSharedMagic.StringPEMTypeCertificate, cryptoutilSharedMagic.StringPEMTypeCSR, cryptoutilSharedMagic.StringPEMTypeSecretKey,
}

// Injectable vars for testing - allows error path coverage without modifying public API.
var (
	x509MarshalPKCS8PrivateKeyFn = x509.MarshalPKCS8PrivateKey
	x509MarshalPKIXPublicKeyFn   = x509.MarshalPKIXPublicKey
	derDecodesPEMTypes           = PEMTypes
	pemEncodeInternalFn          = PEMEncode
	derEncodeInternalFn          = DEREncode
)

// PEMEncodes encodes multiple keys (e.g., certificate chains) to PEM format.
func PEMEncodes(keys any) ([][]byte, error) {
	switch expression := keys.(type) {
	case []*x509.Certificate:
		var pemBytesList [][]byte

		for _, k := range expression {
			pemBytes, err := pemEncodeInternalFn(k)
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

// DEREncodes encodes multiple keys (e.g., certificate chains) to DER format.
func DEREncodes(key any) ([][]byte, error) {
	var derBytesList [][]byte

	switch expression := key.(type) {
	case []*x509.Certificate:
		for _, k := range expression {
			derBytes, _, err := derEncodeInternalFn(k)
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

// PEMEncode encodes a single key to PEM format.
func PEMEncode(key any) ([]byte, error) {
	derBytes, pemType, err := DEREncode(key)
	if err != nil {
		return nil, fmt.Errorf("encode failed: %w", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Bytes: derBytes, Type: pemType})

	return pemBytes, nil
}

// DEREncode encodes a single key to DER format and returns the PEM type identifier.
func DEREncode(key any) ([]byte, string, error) {
	switch x509Type := key.(type) {
	case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey, *ecdh.PrivateKey:
		privateKeyBytes, err := x509MarshalPKCS8PrivateKeyFn(x509Type)
		if err != nil {
			return nil, "", fmt.Errorf("encode failed: %w", err)
		}

		return privateKeyBytes, cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, nil
	case *rsa.PublicKey, *ecdsa.PublicKey, ed25519.PublicKey, *ecdh.PublicKey:
		publicKeyBytes, err := x509MarshalPKIXPublicKeyFn(x509Type)
		if err != nil {
			return nil, "", fmt.Errorf("encode failed: %w", err)
		}

		return publicKeyBytes, cryptoutilSharedMagic.StringPEMTypePKIXPublicKey, nil
	case *x509.Certificate:
		return x509Type.Raw, cryptoutilSharedMagic.StringPEMTypeCertificate, nil
	case *x509.CertificateRequest:
		return x509Type.Raw, cryptoutilSharedMagic.StringPEMTypeCSR, nil
	case []byte:
		return x509Type, cryptoutilSharedMagic.StringPEMTypeSecretKey, nil
	default:
		return nil, "", fmt.Errorf("not supported [%T]", x509Type)
	}
}

// DERDecode decodes DER-encoded bytes using the specified PEM type identifier.
func DERDecode(bytes []byte, x509Type string) (any, error) {
	var key any

	var err error

	switch x509Type {
	case cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey:
		key, err = x509.ParsePKCS8PrivateKey(bytes) // Generic: RSA, EC, ED
	case cryptoutilSharedMagic.StringPEMTypePKIXPublicKey:
		key, err = x509.ParsePKIXPublicKey(bytes) // Generic: RSA, EC, ED
	case cryptoutilSharedMagic.StringPEMTypeRSAPrivateKey:
		key, err = x509.ParsePKCS1PrivateKey(bytes) // RSA PrivateKey
	case cryptoutilSharedMagic.StringPEMTypeRSAPublicKey:
		key, err = x509.ParsePKCS1PublicKey(bytes) // RSA PublicKey
	case cryptoutilSharedMagic.StringPEMTypeECPrivateKey:
		key, err = x509.ParseECPrivateKey(bytes) // EC, ED PrivateKey
	case cryptoutilSharedMagic.StringPEMTypeCertificate:
		key, err = x509.ParseCertificate(bytes)
	case cryptoutilSharedMagic.StringPEMTypeCSR:
		key, err = x509.ParseCertificateRequest(bytes)
	case cryptoutilSharedMagic.StringPEMTypeSecretKey:
		key, err = bytes, nil // AES, HMAC, AES-HMAC
	default:
		return nil, fmt.Errorf("type not supported: %s", x509Type)
	}

	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return key, nil
}

// DERDecodes attempts to decode DER-encoded bytes by trying all known PEM types.
func DERDecodes(bytes []byte) (any, string, error) {
	for _, derType := range derDecodesPEMTypes {
		key, err := DERDecode(bytes, derType)
		if err == nil {
			return key, derType, nil
		}
	}

	return nil, "", fmt.Errorf("decode failed")
}

// PEMDecode decodes PEM-encoded bytes to the appropriate key type.
func PEMDecode(bytes []byte) (any, error) {
	block, rest := pem.Decode(bytes)
	_ = rest // Intentionally ignore remaining bytes after PEM block

	if block == nil {
		return nil, fmt.Errorf("parse PEM failed")
	}

	return DERDecode(block.Bytes, block.Type)
}

// PEMRead reads and decodes a PEM-encoded file.
func PEMRead(filename string) (any, error) {
	pemBytes, err := os.ReadFile(filename) // #nosec G304 -- Legitimate file reading for crypto operations
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return PEMDecode(pemBytes)
}

// DERRead reads and decodes a DER-encoded file.
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

// PEMWrite encodes and writes a key to a PEM-encoded file.
func PEMWrite(key any, filename string) error {
	pemBytes, err := PEMEncode(key)
	if err != nil {
		return fmt.Errorf("encode failed: %w", err)
	}

	dir := filepath.Dir(filename)

	err = os.MkdirAll(dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute)
	if err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	err = os.WriteFile(filename, pemBytes, cryptoutilSharedMagic.FilePermOwnerReadWriteOnly)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}

// DERWrite encodes and writes a key to a DER-encoded file.
func DERWrite(key any, filename string) error {
	derBytes, _, err := DEREncode(key)
	if err != nil {
		return fmt.Errorf("encode failed: %w", err)
	}

	dir := filepath.Dir(filename)

	err = os.MkdirAll(dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute)
	if err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	err = os.WriteFile(filename, derBytes, cryptoutilSharedMagic.FilePermOwnerReadWriteOnly)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}
