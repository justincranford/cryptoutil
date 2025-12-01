// Copyright (c) 2025 Justin Cranford
//
//

// Package hsm provides placeholder interfaces for hardware security module (HSM) integration.
// This package is designed to be extensible for future PKCS#11 and YubiKey support.
//
// Current Status: PLACEHOLDER
// This package defines interfaces that will be implemented in future versions
// to support hardware-backed key storage and cryptographic operations.
//
// Future Implementation Plans:
//   - PKCS#11 support for HSM devices (Thales Luna, AWS CloudHSM, etc.)
//   - YubiKey PIV support for smart card operations
//   - Azure Key Vault and AWS KMS integration
//
// Reference: Session 5 Q2 - Design extensible API with placeholder package.
package hsm

import (
	"context"
	"crypto"
	"crypto/x509"
)

// Provider represents a hardware security module provider.
// Implementations of this interface will provide access to HSM devices.
type Provider interface {
	// Open initializes the connection to the HSM.
	Open(ctx context.Context, config *ProviderConfig) error

	// Close terminates the connection to the HSM.
	Close() error

	// IsAvailable returns true if the HSM is available and ready.
	IsAvailable() bool

	// GetSlots returns available slots on the HSM.
	GetSlots() ([]Slot, error)
}

// ProviderConfig holds configuration for an HSM provider.
type ProviderConfig struct {
	// Type is the HSM provider type (e.g., "pkcs11", "yubikey").
	Type string

	// LibraryPath is the path to the PKCS#11 library (for PKCS#11 providers).
	LibraryPath string

	// PIN is the user PIN for authentication (stored securely, not logged).
	PIN string

	// SlotID is the slot identifier to use.
	SlotID uint

	// TokenLabel is the optional token label for identification.
	TokenLabel string
}

// Slot represents a slot on an HSM device.
type Slot interface {
	// ID returns the slot identifier.
	ID() uint

	// Label returns the slot label.
	Label() string

	// ListKeys returns all keys available in this slot.
	ListKeys() ([]KeyInfo, error)

	// GetKey retrieves a key by ID or label.
	GetKey(ctx context.Context, identifier string) (Key, error)

	// GenerateKey generates a new key pair in the slot.
	GenerateKey(ctx context.Context, spec *KeySpec) (Key, error)

	// ImportCertificate imports a certificate into the slot.
	ImportCertificate(ctx context.Context, cert *x509.Certificate) error
}

// KeyInfo provides metadata about a key stored in the HSM.
type KeyInfo struct {
	// ID is the unique identifier for the key.
	ID string

	// Label is the human-readable label for the key.
	Label string

	// Type is the key type (e.g., "RSA", "EC", "AES").
	Type string

	// Size is the key size in bits.
	Size int

	// CanSign indicates if the key can be used for signing.
	CanSign bool

	// CanEncrypt indicates if the key can be used for encryption.
	CanEncrypt bool

	// CanDecrypt indicates if the key can be used for decryption.
	CanDecrypt bool

	// CanWrap indicates if the key can be used for key wrapping.
	CanWrap bool

	// CanUnwrap indicates if the key can be used for key unwrapping.
	CanUnwrap bool
}

// Key represents a key stored in the HSM.
// Keys in HSMs are typically non-exportable; operations are performed on-device.
type Key interface {
	// Info returns metadata about the key.
	Info() KeyInfo

	// Signer returns a crypto.Signer interface for signing operations.
	// Returns nil if the key cannot sign.
	Signer() crypto.Signer

	// Decrypter returns a crypto.Decrypter interface for decryption operations.
	// Returns nil if the key cannot decrypt.
	Decrypter() crypto.Decrypter

	// Delete removes the key from the HSM.
	Delete(ctx context.Context) error
}

// KeySpec defines the specification for generating a new key.
type KeySpec struct {
	// Label is the human-readable label for the key.
	Label string

	// Type is the key type to generate (e.g., "RSA", "EC").
	Type string

	// Size is the key size in bits (e.g., 2048, 4096 for RSA; 256, 384 for EC).
	Size int

	// Curve is the elliptic curve name for EC keys (e.g., "P-256", "P-384").
	Curve string

	// Extractable indicates if the key can be exported (usually false for HSM keys).
	Extractable bool

	// Usage specifies allowed key operations.
	Usage KeyUsage
}

// KeyUsage specifies allowed operations for a key.
type KeyUsage struct {
	Sign    bool
	Verify  bool
	Encrypt bool
	Decrypt bool
	Wrap    bool
	Unwrap  bool
}

// ErrNotImplemented is returned when an HSM operation is not yet implemented.
type ErrNotImplemented struct {
	Operation string
}

func (e *ErrNotImplemented) Error() string {
	return "HSM operation not implemented: " + e.Operation
}

// ErrHSMNotAvailable is returned when the HSM is not available.
type ErrHSMNotAvailable struct {
	Reason string
}

func (e *ErrHSMNotAvailable) Error() string {
	return "HSM not available: " + e.Reason
}

// ErrKeyNotFound is returned when a key cannot be found in the HSM.
type ErrKeyNotFound struct {
	Identifier string
}

func (e *ErrKeyNotFound) Error() string {
	return "HSM key not found: " + e.Identifier
}
