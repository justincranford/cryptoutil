// Copyright (c) 2025 Justin Cranford

package security

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
"context"
ecdsa "crypto/ecdsa"
"crypto/ed25519"
"crypto/elliptic"
crand "crypto/rand"
rsa "crypto/rsa"
"crypto/x509"
"crypto/x509/pkix"
"math/big"
"testing"
"time"

"github.com/stretchr/testify/require"
)

// ─── ValidateCSR ─────────────────────────────────────────────────────────────

func TestValidator_ValidateCSR_RSABelowMinimum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := &Config{
		MinRSAKeySize:              cryptoutilSharedMagic.RSA4096KeySize,
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.SHA256WithRSA},
	}
	validator := NewValidator(config)

	key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	csr := createCSRWithRSAKey(t, key, nil)

	result, err := validator.ValidateCSR(ctx, csr)
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

	require.True(t, found, "expected below-minimum CSR error, got: %v", result.Errors)
}

func TestValidator_ValidateCSR_ECDSABelowMinimum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := &Config{
		MinECKeySize:               cryptoutilSharedMagic.SymmetricKeySize384,
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA256},
	}
	validator := NewValidator(config)

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	csr := createTestCSR(t, key, nil)

	result, err := validator.ValidateCSR(ctx, csr)
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

	require.True(t, found, "expected below-minimum CSR error, got: %v", result.Errors)
}

func TestValidator_ValidateCSR_Ed25519(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := &Config{
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.PureEd25519},
	}
	validator := NewValidator(config)

	pub, priv, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	_ = pub

	csr := createCSRWithEd25519Key(t, priv, []string{"example.com"})

	result, err := validator.ValidateCSR(ctx, csr)
	require.NoError(t, err)
	require.True(t, result.Valid)
}

func TestValidator_ValidateCSR_DisallowedSignatureAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := &Config{
		MinECKeySize:               cryptoutilSharedMagic.MaxUnsealSharedSecrets,
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA384},
	}
	validator := NewValidator(config)

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	// ECDSAWithSHA256 is NOT in the allowed list.
	csr := createTestCSR(t, key, []string{"example.com"})

	result, err := validator.ValidateCSR(ctx, csr)
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)

	found := false

	for _, e := range result.Errors {
		if containsStr(e, "not in allowed list") {
			found = true

			break
		}
	}

	require.True(t, found, "expected not-in-allowed-list error, got: %v", result.Errors)
}

func TestValidator_ValidateCSR_RequireSAN_NoSAN(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := &Config{
		MinECKeySize:               cryptoutilSharedMagic.MaxUnsealSharedSecrets,
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA256},
		RequireSAN:                 true,
	}
	validator := NewValidator(config)

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	csr := createTestCSR(t, key, nil) // No SANs.

	result, err := validator.ValidateCSR(ctx, csr)
	require.NoError(t, err)
	require.True(t, result.Valid) // Warning only.

	found := false

	for _, w := range result.Warnings {
		if containsStr(w, "Subject Alternative Name") {
			found = true

			break
		}
	}

	require.True(t, found, "expected SAN warning, got: %v", result.Warnings)
}

// ─── ScanCertificateChain: error from ValidateCertificate ───────────────────

func TestScanner_ScanCertificateChain_InvalidCert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use a config that rejects the cert (empty allowed algorithms list → all certs are invalid for signing alg).
	// But we need ValidateCertificate to return an error (not just invalid result).
	// ValidateCertificate returns error only when cert.PublicKey type causes validateKeySize to error.
	// Actually validateKeySize only returns nil error. Let me use an expired cert in a chain scan.
	// ScanCertificateChain returns error from ValidateCertificate only if ValidateCertificate returns error.
	// ValidateCertificate returns error only for nil cert.
	// So to exercise the error branch, we'd need to pass a nil in the chain.
	// Let's test with an invalid (expired) cert to ensure the chain returns a non-valid result.
	scanner := NewScanner(nil)

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	// Expired cert.
	expiredCert := createTestCert(t, key, false,
		time.Now().UTC().Add(-730*cryptoutilSharedMagic.HoursPerDay*time.Hour),
		time.Now().UTC().Add(-cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour))

	result, err := scanner.ScanCertificateChain(ctx, []*x509.Certificate{expiredCert})
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)
}

func TestScanner_ScanCertificateChain_ValidatesChainLinkage_IssuerSubjectMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	scanner := NewScanner(nil)

	// Two CA certs where child says it was issued by "Fake Issuer" but parent is a different CA.
	key1, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	key2, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	// Create cert1 signed by key2 but with Issuer != key2's cert's Subject.
	rootCert := createTestCACert(t, key2, nil, nil, "Root CA")

	// Create an intermediate signed by rootCert (properly linked).
	intCert := createTestCACert(t, key1, rootCert, key2, "Intermediate CA")

	// Create another root to form a broken chain.
	key3, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	otherRoot := createTestCACert(t, key3, nil, nil, "Other Root CA")

	// intCert is signed by rootCert but we present otherRoot as the parent.
	result, err := scanner.ScanCertificateChain(ctx, []*x509.Certificate{intCert, otherRoot})
	require.NoError(t, err)
	require.False(t, result.Valid) // Signature check fails.
	_ = result
}

// ─── GenerateReport: mitigated threats, Low, Info severity ───────────────────

func TestGenerateReport_MitigatedThreats(t *testing.T) {
	t.Parallel()

	threatModel := &ThreatModel{
		Name:    "Test",
		Version: cryptoutilSharedMagic.ServiceVersion,
		Threats: []Threat{
			{ID: "T-001", Status: "mitigated"},
			{ID: "T-002", Status: "mitigated"},
			{ID: "T-003", Status: "open"},
		},
	}

	report := GenerateReport(threatModel, nil)
	require.NotNil(t, report)
	require.Equal(t, 3, report.Summary.TotalThreats)
	require.Equal(t, 2, report.Summary.MitigatedThreats)
	require.Equal(t, 1, report.Summary.OpenThreats)
}

func TestGenerateReport_LowAndInfoSeverity(t *testing.T) {
	t.Parallel()

	validations := []ValidationResult{
		{
			Valid:    true,
			Errors:   []string{},
			Warnings: []string{},
			Vulnerabilities: []Vulnerability{
				{ID: "V-001", Severity: SeverityLow},
				{ID: "V-002", Severity: SeverityInfo},
			},
		},
	}

	report := GenerateReport(nil, validations)
	require.NotNil(t, report)
	require.Equal(t, 2, report.Summary.TotalVulnerabilities)
	require.Equal(t, 1, report.Summary.LowCount)
	require.Equal(t, 1, report.Summary.InfoCount)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func mustGenerateECKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	return key
}

func mustGenerateRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	return key
}

func createCSRWithRSAKey(t *testing.T, key *rsa.PrivateKey, dnsNames []string) *x509.CertificateRequest {
	t.Helper()

	template := &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: "Test CSR RSA"},
		DNSNames: dnsNames,
	}

	csrBytes, err := x509.CreateCertificateRequest(crand.Reader, template, key)
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrBytes)
	require.NoError(t, err)

	return csr
}

func createCSRWithEd25519Key(t *testing.T, key ed25519.PrivateKey, dnsNames []string) *x509.CertificateRequest {
	t.Helper()

	template := &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: "Test CSR Ed25519"},
		DNSNames: dnsNames,
	}

	csrBytes, err := x509.CreateCertificateRequest(crand.Reader, template, key)
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrBytes)
	require.NoError(t, err)

	return csr
}

// containsStr is a simple substring check to avoid importing strings in test code.
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}

			return false
		}())
}

func TestNewScanner(t *testing.T) {
	t.Parallel()

	scanner := NewScanner(nil)
	require.NotNil(t, scanner)

	scanner2 := NewScanner(DefaultConfig())
	require.NotNil(t, scanner2)
}

// ─── validateKeySize: unknown public key type in cert ────────────────────────

func TestValidator_ValidateCertificate_UnknownPublicKeyType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := &Config{
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA256},
		MaxCertValidityDays:        cryptoutilSharedMagic.TLSMaxValidityEndEntityDays,
	}
	validator := NewValidator(config)

	// Build a cert struct directly with an unsupported public key type.
	cert := &x509.Certificate{
		IsCA:               false,
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		NotBefore:          time.Now().UTC(),
		NotAfter:           time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:           x509.KeyUsageDigitalSignature,
		PublicKey:          "unsupported-key-type", // triggers default branch.
		DNSNames:           []string{"example.com"},
	}

	result, err := validator.ValidateCertificate(ctx, cert)
	require.NoError(t, err)
	require.True(t, result.Valid) // Unknown key type is a warning, not error.

	found := false

	for _, w := range result.Warnings {
		if containsStr(w, "unknown public key type") {
			found = true

			break
		}
	}

	require.True(t, found, "expected unknown public key warning, got: %v", result.Warnings)
}

// ─── ScanCertificateChain: nil cert in chain triggers error ──────────────────

func TestScanner_ScanCertificateChain_NilCertInChain(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	scanner := NewScanner(nil)

	// Nil cert in chain causes ValidateCertificate to return error.
	chain := []*x509.Certificate{nil}

	result, err := scanner.ScanCertificateChain(ctx, chain)
	require.Error(t, err)
	require.Nil(t, result)
}

// Ensure big.NewInt is used so the import is used in this file.
var _ = big.NewInt
