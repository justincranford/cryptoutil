// Copyright (c) 2025 Justin Cranford

package security

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ─── validateSignatureAlgorithm ──────────────────────────────────────────────

func TestValidator_ValidateCertificate_DisallowedSignatureAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Config that only allows ECDSAWithSHA384 — use a SHA256WithRSA cert.
	key, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	config := &Config{
		MinRSAKeySize:              2048,
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA384},
		MaxCertValidityDays:        398,
	}
	validator := NewValidator(config)

	cert := createTestCertWithKey(t, key, false,
		time.Now().UTC(), time.Now().UTC().Add(365*24*time.Hour))

	result, err := validator.ValidateCertificate(ctx, cert)
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)

	found := false

	for _, e := range result.Errors {
		if len(e) > 0 && containsStr(e, "not in allowed list") {
			found = true

			break
		}
	}

	require.True(t, found, "expected 'not in allowed list' error, got: %v", result.Errors)
}

// ─── validateValidityPeriod (not-yet-valid branch) ──────────────────────────

func TestValidator_ValidateCertificate_NotYetValid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	config := &Config{
		MinECKeySize:               256,
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA256},
		MaxCertValidityDays:        398,
	}
	validator := NewValidator(config)

	// Certificate that becomes valid in 30 days (not yet valid).
	cert := createTestCert(t, key, false,
		time.Now().UTC().Add(30*24*time.Hour),
		time.Now().UTC().Add(90*24*time.Hour))

	result, err := validator.ValidateCertificate(ctx, cert)
	require.NoError(t, err)
	require.True(t, result.Valid) // Not yet valid is a warning, not error.

	found := false

	for _, w := range result.Warnings {
		if containsStr(w, "not yet valid") {
			found = true

			break
		}
	}

	require.True(t, found, "expected 'not yet valid' warning, got: %v", result.Warnings)
}

// ─── validateExtensions ──────────────────────────────────────────────────────

func TestValidator_ValidateCertificate_CAMissingBasicConstraints(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := &Config{
		MinECKeySize:               256,
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA256},
		MaxCertValidityDays:        398,
		RequireBasicConstraints:    true,
	}
	validator := NewValidator(config)

	// Build a CA cert struct directly — BasicConstraintsValid=false.
	cert := &x509.Certificate{
		IsCA:                  true,
		BasicConstraintsValid: false,
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(365 * 24 * time.Hour),
		PublicKey:             mustGenerateECKey(t).Public(),
	}

	result, err := validator.ValidateCertificate(ctx, cert)
	require.NoError(t, err)
	require.False(t, result.Valid)

	found := false

	for _, e := range result.Errors {
		if containsStr(e, "CA certificate missing valid basic constraints") {
			found = true

			break
		}
	}

	require.True(t, found, "expected basic constraints error, got: %v", result.Errors)
}

func TestValidator_ValidateCertificate_NonCAMissingSAN(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := &Config{
		MinECKeySize:               256,
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA256},
		MaxCertValidityDays:        398,
		RequireSAN:                 true,
	}
	validator := NewValidator(config)

	// Non-CA cert with no SANs.
	cert := &x509.Certificate{
		IsCA:               false,
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		NotBefore:          time.Now().UTC(),
		NotAfter:           time.Now().UTC().Add(365 * 24 * time.Hour),
		KeyUsage:           x509.KeyUsageDigitalSignature,
		PublicKey:          mustGenerateECKey(t).Public(),
	}

	result, err := validator.ValidateCertificate(ctx, cert)
	require.NoError(t, err)
	require.True(t, result.Valid) // Missing SAN is a warning, not error.

	found := false

	for _, w := range result.Warnings {
		if containsStr(w, "Subject Alternative Name") {
			found = true

			break
		}
	}

	require.True(t, found, "expected SAN warning, got: %v", result.Warnings)
}

func TestValidator_ValidateCertificate_MissingKeyUsage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := &Config{
		MinECKeySize:               256,
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA256},
		MaxCertValidityDays:        398,
		RequireKeyUsage:            true,
	}
	validator := NewValidator(config)

	// Cert with no KeyUsage bits set.
	cert := &x509.Certificate{
		IsCA:               false,
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		NotBefore:          time.Now().UTC(),
		NotAfter:           time.Now().UTC().Add(365 * 24 * time.Hour),
		KeyUsage:           0,
		PublicKey:          mustGenerateECKey(t).Public(),
		DNSNames:           []string{"example.com"},
	}

	result, err := validator.ValidateCertificate(ctx, cert)
	require.NoError(t, err)
	require.True(t, result.Valid) // Missing key usage is a warning, not error.

	found := false

	for _, w := range result.Warnings {
		if containsStr(w, "key usage extension") {
			found = true

			break
		}
	}

	require.True(t, found, "expected key usage warning, got: %v", result.Warnings)
}

// ─── checkWeakAlgorithms ─────────────────────────────────────────────────────

func TestValidator_ValidateCertificate_WeakAlgorithm_SHA1WithRSA(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// DisallowWeakAlgorithms=true, SHA1WithRSA cert.
	config := &Config{
		MinRSAKeySize:              2048,
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.SHA1WithRSA},
		MaxCertValidityDays:        398,
		DisallowWeakAlgorithms:     true,
	}
	validator := NewValidator(config)

	cert := &x509.Certificate{
		SignatureAlgorithm: x509.SHA1WithRSA,
		NotBefore:          time.Now().UTC(),
		NotAfter:           time.Now().UTC().Add(365 * 24 * time.Hour),
		KeyUsage:           x509.KeyUsageDigitalSignature,
		DNSNames:           []string{"example.com"},
		PublicKey:          mustGenerateRSAKey(t).Public(),
	}

	result, err := validator.ValidateCertificate(ctx, cert)
	require.NoError(t, err)
	require.False(t, result.Valid)

	found := false

	for _, vuln := range result.Vulnerabilities {
		if vuln.ID == "WEAK-ALG-001" {
			found = true

			break
		}
	}

	require.True(t, found, "expected WEAK-ALG-001 vulnerability, got: %v", result.Vulnerabilities)
}

// ─── validatePathLength ───────────────────────────────────────────────────────

func TestValidator_ValidateCertificate_CANoPathLengthConstraint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := &Config{
		MinECKeySize:                 256,
		AllowedSignatureAlgorithms:   []x509.SignatureAlgorithm{x509.ECDSAWithSHA256},
		MaxCertValidityDays:          398,
		EnforcePathLengthConstraints: true,
	}
	validator := NewValidator(config)

	// CA cert with MaxPathLen=0 and MaxPathLenZero=false → no path length constraint.
	cert := &x509.Certificate{
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLen:            0,
		MaxPathLenZero:        false,
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		DNSNames:              []string{"ca.example.com"},
		PublicKey:             mustGenerateECKey(t).Public(),
	}

	result, err := validator.ValidateCertificate(ctx, cert)
	require.NoError(t, err)
	// Path length warning doesn't make cert invalid.
	found := false

	for _, w := range result.Warnings {
		if containsStr(w, "path length constraint") {
			found = true

			break
		}
	}

	require.True(t, found, "expected path length warning, got: %v", result.Warnings)
}

// ─── ValidatePrivateKey: below-minimum branches ──────────────────────────────

func TestValidator_ValidatePrivateKey_RSABelowMinimum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use a high minimum so 2048-bit key fails.
	config := &Config{MinRSAKeySize: 4096}
	validator := NewValidator(config)

	key, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	result, err := validator.ValidatePrivateKey(ctx, key)
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)

	found := false

	for _, e := range result.Errors {
		if containsStr(e, "below minimum") {
			found = true

			break
		}
	}

	require.True(t, found, "expected below-minimum error, got: %v", result.Errors)
}

func TestValidator_ValidatePrivateKey_ECDSABelowMinimum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use a high minimum so P-256 (256-bit) key fails.
	config := &Config{MinECKeySize: 384}
	validator := NewValidator(config)

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	result, err := validator.ValidatePrivateKey(ctx, key)
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)

	found := false

	for _, e := range result.Errors {
		if containsStr(e, "below minimum") {
			found = true

			break
		}
	}

	require.True(t, found, "expected below-minimum error, got: %v", result.Errors)
}
