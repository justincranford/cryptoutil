package fileutil

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
)

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(filename, data, perm)
}

func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
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
