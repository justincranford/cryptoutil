package codec

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func PublicKeyToPEMBlock(key interface{}) (*pem.Block, error) {
	var pemBlock *pem.Block

	switch k := key.(type) {
	case *rsa.PublicKey:
		bytes, err := x509.MarshalPKIXPublicKey(k)
		if err != nil {
			return nil, err
		}
		pemBlock = &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: bytes,
		}
	case *ecdsa.PublicKey:
		bytes, err := x509.MarshalPKIXPublicKey(k)
		if err != nil {
			return nil, err
		}
		pemBlock = &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: bytes,
		}
	case ed25519.PublicKey:
		bytes, err := x509.MarshalPKIXPublicKey(k)
		if err != nil {
			return nil, err
		}
		pemBlock = &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: bytes,
		}
	default:
		return nil, fmt.Errorf("unsupported public key type")
	}

	return pemBlock, nil
}

func PrivateKeyToPEMBlock(key interface{}) (*pem.Block, error) {
	var pemBlock *pem.Block

	switch k := key.(type) {
	case *rsa.PrivateKey:
		pemBlock = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		}
	case *ecdsa.PrivateKey:
		bytes, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}
		pemBlock = &pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: bytes,
		}
	case ed25519.PrivateKey:
		bytes, err := x509.MarshalPKCS8PrivateKey(k)
		if err != nil {
			return nil, err
		}
		pemBlock = &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: bytes,
		}
	default:
		return nil, fmt.Errorf("unsupported private key type")
	}

	return pemBlock, nil
}

func KeyToPEMString(key interface{}, isPub bool) (string, error) {
	var pemBlock *pem.Block
	var err error

	if isPub {
		pemBlock, err = PublicKeyToPEMBlock(key)
	} else {
		pemBlock, err = PrivateKeyToPEMBlock(key)
	}

	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(pemBlock)), nil
}

func KeyToDERBytes(key interface{}, isPub bool) ([]byte, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		if isPub {
			return x509.MarshalPKIXPublicKey(&k.PublicKey)
		}
		return x509.MarshalPKCS1PrivateKey(k), nil
	case *ecdsa.PrivateKey:
		if isPub {
			return x509.MarshalPKIXPublicKey(&k.PublicKey)
		}
		return x509.MarshalECPrivateKey(k)
	case ed25519.PrivateKey:
		if isPub {
			return x509.MarshalPKIXPublicKey(k.Public())
		}
		return x509.MarshalPKCS8PrivateKey(k)
	case *rsa.PublicKey, *ecdsa.PublicKey, ed25519.PublicKey:
		return x509.MarshalPKIXPublicKey(k)
	default:
		return nil, fmt.Errorf("unsupported key type")
	}
}

func KeyFromPEMString(pemString string) (interface{}, error) {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	case "PUBLIC KEY":
		return x509.ParsePKIXPublicKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}
}

func KeyFromDERBytes(derBytes []byte) (interface{}, error) {
	if pub, err := x509.ParsePKIXPublicKey(derBytes); err == nil {
		return pub, nil
	}

	if priv, err := x509.ParsePKCS8PrivateKey(derBytes); err == nil {
		return priv, nil
	}

	if priv, err := x509.ParsePKCS1PrivateKey(derBytes); err == nil {
		return priv, nil
	}

	if priv, err := x509.ParseECPrivateKey(derBytes); err == nil {
		return priv, nil
	}

	return nil, fmt.Errorf("failed to parse DER bytes")
}

func WriteKeyToPEMFile(key interface{}, filename string, isPub bool) error {
	pemString, err := KeyToPEMString(key, isPub)
	if err != nil {
		return err
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(filename, []byte(pemString), 0600)
}

func WriteKeyToDERFile(key interface{}, filename string, isPub bool) error {
	derBytes, err := KeyToDERBytes(key, isPub)
	if err != nil {
		return err
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(filename, derBytes, 0600)
}

func ReadKeyFromPEMFile(filename string) (interface{}, error) {
	pemBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return KeyFromPEMString(string(pemBytes))
}

func ReadKeyFromDERFile(filename string) (interface{}, error) {
	derBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return KeyFromDERBytes(derBytes)
}
