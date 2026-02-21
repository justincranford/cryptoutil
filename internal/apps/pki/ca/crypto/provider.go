// Copyright (c) 2025 Justin Cranford
//
//

// Package crypto provides cryptographic provider interfaces for the CA subsystem.
package crypto

import (
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	rsa "crypto/rsa"
	"crypto/x509"
	"fmt"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
)

// KeyType represents the type of cryptographic key.
type KeyType string

// Key type constants for supported algorithms.
const (
	// KeyTypeRSA represents RSA cryptographic keys.
	KeyTypeRSA KeyType = "RSA"
	// KeyTypeECDSA represents ECDSA cryptographic keys.
	KeyTypeECDSA KeyType = "ECDSA"
	// KeyTypeEdDSA represents EdDSA cryptographic keys.
	KeyTypeEdDSA KeyType = "EdDSA"
)

// Key size constants for FIPS 140-3 compliance.
const (
	MinRSAKeyBits    = 2048
	MediumRSAKeyBits = 3072
	LargeRSAKeyBits  = 4096
)

// KeySpec specifies the parameters for key generation.
type KeySpec struct {
	Type       KeyType
	RSABits    int    // For RSA keys: 2048, 3072, 4096
	ECDSACurve string // For ECDSA keys: P-256, P-384, P-521
	EdDSACurve string // For EdDSA keys: Ed25519, Ed448
}

// KeyPair holds a generated key pair.
type KeyPair struct {
	PublicKey  crypto.PublicKey
	PrivateKey crypto.PrivateKey
	Type       KeyType
	Algorithm  string // Human-readable algorithm description
}

// Provider defines the interface for cryptographic operations.
type Provider interface {
	// GenerateKeyPair generates a new key pair according to the spec.
	GenerateKeyPair(spec KeySpec) (*KeyPair, error)
	// Sign creates a signature over the digest using the private key.
	Sign(privateKey crypto.PrivateKey, digest []byte, opts crypto.SignerOpts) ([]byte, error)
	// Verify validates a signature using the public key.
	Verify(publicKey crypto.PublicKey, digest, signature []byte, opts crypto.SignerOpts) error
	// GetSignatureAlgorithm returns the appropriate x509.SignatureAlgorithm for the key.
	GetSignatureAlgorithm(publicKey crypto.PublicKey) (x509.SignatureAlgorithm, error)
}

// SoftwareProvider implements Provider using software-based cryptography.
type SoftwareProvider struct{}

// Injectable vars for testing - allows error path coverage without modifying public API.
var (
	pkiCryptoGenerateRSAKeyPairFn  func(int) (*cryptoutilSharedCryptoKeygen.KeyPair, error)                   = cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair
	pkiCryptoGenerateECDSAKeyPairFn func(elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) = cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair
)

// NewSoftwareProvider creates a new software-based crypto provider.
func NewSoftwareProvider() *SoftwareProvider {
	return &SoftwareProvider{}
}

// GenerateKeyPair generates a new key pair using the internal keygen package.
func (p *SoftwareProvider) GenerateKeyPair(spec KeySpec) (*KeyPair, error) {
	switch spec.Type {
	case KeyTypeRSA:
		return p.generateRSAKeyPair(spec.RSABits)
	case KeyTypeECDSA:
		return p.generateECDSAKeyPair(spec.ECDSACurve)
	case KeyTypeEdDSA:
		return p.generateEdDSAKeyPair(spec.EdDSACurve)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", spec.Type)
	}
}

func (p *SoftwareProvider) generateRSAKeyPair(bits int) (*KeyPair, error) {
	if bits < MinRSAKeyBits {
		return nil, fmt.Errorf("RSA key size must be at least %d bits, got %d", MinRSAKeyBits, bits)
	}

	keyPair, err := pkiCryptoGenerateRSAKeyPairFn(bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	return &KeyPair{
		PublicKey:  keyPair.Public,
		PrivateKey: keyPair.Private,
		Type:       KeyTypeRSA,
		Algorithm:  fmt.Sprintf("RSA-%d", bits),
	}, nil
}

func (p *SoftwareProvider) generateECDSAKeyPair(curve string) (*KeyPair, error) {
	var ecCurve elliptic.Curve

	switch curve {
	case "P-256":
		ecCurve = elliptic.P256()
	case "P-384":
		ecCurve = elliptic.P384()
	case "P-521":
		ecCurve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported ECDSA curve: %s", curve)
	}

	keyPair, err := pkiCryptoGenerateECDSAKeyPairFn(ecCurve)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA key pair: %w", err)
	}

	return &KeyPair{
		PublicKey:  keyPair.Public,
		PrivateKey: keyPair.Private,
		Type:       KeyTypeECDSA,
		Algorithm:  fmt.Sprintf("ECDSA-%s", curve),
	}, nil
}

func (p *SoftwareProvider) generateEdDSAKeyPair(curve string) (*KeyPair, error) {
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair(curve)
	if err != nil {
		return nil, fmt.Errorf("failed to generate EdDSA key pair: %w", err)
	}

	return &KeyPair{
		PublicKey:  keyPair.Public,
		PrivateKey: keyPair.Private,
		Type:       KeyTypeEdDSA,
		Algorithm:  curve,
	}, nil
}

// Sign creates a signature using the private key.
func (p *SoftwareProvider) Sign(privateKey crypto.PrivateKey, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("private key does not implement crypto.Signer")
	}

	signature, err := signer.Sign(nil, digest, opts)
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	return signature, nil
}

// Verify validates a signature using the public key.
func (p *SoftwareProvider) Verify(publicKey crypto.PublicKey, digest, signature []byte, opts crypto.SignerOpts) error {
	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		return p.verifyRSA(pub, digest, signature, opts)
	case *ecdsa.PublicKey:
		return p.verifyECDSA(pub, digest, signature)
	case ed25519.PublicKey:
		return p.verifyEdDSA(pub, digest, signature)
	default:
		return fmt.Errorf("unsupported public key type: %T", publicKey)
	}
}

func (p *SoftwareProvider) verifyRSA(pub *rsa.PublicKey, digest, signature []byte, opts crypto.SignerOpts) error {
	hash := opts.HashFunc()
	if hash == 0 {
		return fmt.Errorf("hash function required for RSA verification")
	}

	if err := rsa.VerifyPKCS1v15(pub, hash, digest, signature); err != nil {
		return fmt.Errorf("RSA signature verification failed: %w", err)
	}

	return nil
}

func (p *SoftwareProvider) verifyECDSA(pub *ecdsa.PublicKey, digest, signature []byte) error {
	if !ecdsa.VerifyASN1(pub, digest, signature) {
		return fmt.Errorf("ECDSA signature verification failed")
	}

	return nil
}

func (p *SoftwareProvider) verifyEdDSA(pub ed25519.PublicKey, digest, signature []byte) error {
	if !ed25519.Verify(pub, digest, signature) {
		return fmt.Errorf("EdDSA signature verification failed")
	}

	return nil
}

// GetSignatureAlgorithm returns the appropriate x509.SignatureAlgorithm for the key.
func (p *SoftwareProvider) GetSignatureAlgorithm(publicKey crypto.PublicKey) (x509.SignatureAlgorithm, error) {
	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		bits := pub.N.BitLen()
		if bits >= LargeRSAKeyBits {
			return x509.SHA512WithRSA, nil
		} else if bits >= MediumRSAKeyBits {
			return x509.SHA384WithRSA, nil
		}

		return x509.SHA256WithRSA, nil
	case *ecdsa.PublicKey:
		switch pub.Curve {
		case elliptic.P521():
			return x509.ECDSAWithSHA512, nil
		case elliptic.P384():
			return x509.ECDSAWithSHA384, nil
		default:
			return x509.ECDSAWithSHA256, nil
		}
	case ed25519.PublicKey:
		return x509.PureEd25519, nil
	default:
		return 0, fmt.Errorf("unsupported public key type: %T", publicKey)
	}
}

// ParseKeySpecFromConfig creates a KeySpec from configuration values.
func ParseKeySpecFromConfig(algorithm, curveOrSize string) (KeySpec, error) {
	spec := KeySpec{}

	switch algorithm {
	case "RSA":
		spec.Type = KeyTypeRSA

		var bits int
		if _, err := fmt.Sscanf(curveOrSize, "%d", &bits); err != nil {
			return spec, fmt.Errorf("invalid RSA key size: %s", curveOrSize)
		}

		spec.RSABits = bits
	case "ECDSA":
		spec.Type = KeyTypeECDSA
		spec.ECDSACurve = curveOrSize
	case "EdDSA":
		spec.Type = KeyTypeEdDSA
		spec.EdDSACurve = curveOrSize
	default:
		return spec, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	return spec, nil
}
