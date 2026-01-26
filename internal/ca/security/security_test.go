// Copyright (c) 2025 Justin Cranford

package security

import (
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

func TestNewValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "with nil config uses defaults",
			config: nil,
		},
		{
			name:   "with custom config",
			config: &Config{MinRSAKeySize: 4096},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator := NewValidator(tc.config)
			require.NotNil(t, validator)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	require.Equal(t, 2048, config.MinRSAKeySize)
	require.Equal(t, 256, config.MinECKeySize)
	require.Equal(t, 398, config.MaxCertValidityDays)
	require.True(t, config.RequireKeyUsage)
	require.True(t, config.RequireBasicConstraints)
	require.True(t, config.RequireSAN)
	require.True(t, config.DisallowWeakAlgorithms)
	require.True(t, config.EnforcePathLengthConstraints)
	require.True(t, config.AuditLoggingEnabled)
	require.NotEmpty(t, config.AllowedSignatureAlgorithms)
}

func TestValidator_ValidateCertificate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewValidator(nil)

	// Generate test key.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	tests := []struct {
		name      string
		certFunc  func() *x509.Certificate
		wantValid bool
		wantErr   bool
	}{
		{
			name: "valid certificate",
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now().UTC(), time.Now().UTC().Add(365*24*time.Hour))
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "expired certificate",
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now().UTC().Add(-730*24*time.Hour), time.Now().UTC().Add(-365*24*time.Hour))
			},
			wantValid: false,
			wantErr:   false,
		},
		{
			name: "certificate with excessive validity",
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now().UTC(), time.Now().UTC().Add(500*24*time.Hour))
			},
			wantValid: false,
			wantErr:   false,
		},
		{
			name:      "nil certificate",
			certFunc:  func() *x509.Certificate { return nil },
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cert := tc.certFunc()
			result, err := validator.ValidateCertificate(ctx, cert)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.wantValid, result.Valid)
			}
		})
	}
}

func TestValidator_ValidateCertificate_KeySizes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		keyFunc   func() any
		config    *Config
		wantValid bool
	}{
		{
			name: "RSA 2048 with 2048 minimum",
			keyFunc: func() any {
				key, _ := rsa.GenerateKey(crand.Reader, 2048)

				return key
			},
			config:    &Config{MinRSAKeySize: 2048, AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.SHA256WithRSA}, MaxCertValidityDays: 398},
			wantValid: true,
		},
		{
			name: "RSA 2048 with 4096 minimum",
			keyFunc: func() any {
				key, _ := rsa.GenerateKey(crand.Reader, 2048)

				return key
			},
			config:    &Config{MinRSAKeySize: 4096, AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.SHA256WithRSA}, MaxCertValidityDays: 398},
			wantValid: false,
		},
		{
			name: "EC P-256 with 256 minimum",
			keyFunc: func() any {
				key, _ := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)

				return key
			},
			config:    &Config{MinECKeySize: 256, AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA256}, MaxCertValidityDays: 398},
			wantValid: true,
		},
		{
			name: "EC P-256 with 384 minimum",
			keyFunc: func() any {
				key, _ := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)

				return key
			},
			config:    &Config{MinECKeySize: 384, AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.ECDSAWithSHA256}, MaxCertValidityDays: 398},
			wantValid: false,
		},
		{
			name: "Ed25519 key",
			keyFunc: func() any {
				_, key, _ := ed25519.GenerateKey(crand.Reader)

				return key
			},
			config:    &Config{AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{x509.PureEd25519}, MaxCertValidityDays: 398},
			wantValid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator := NewValidator(tc.config)
			key := tc.keyFunc()

			cert := createTestCertWithKey(t, key, false, time.Now().UTC(), time.Now().UTC().Add(365*24*time.Hour))
			result, err := validator.ValidateCertificate(ctx, cert)

			require.NoError(t, err)
			require.Equal(t, tc.wantValid, result.Valid)
		})
	}
}

func TestValidator_ValidatePrivateKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewValidator(nil)

	tests := []struct {
		name      string
		keyFunc   func() any
		wantValid bool
		wantErr   bool
	}{
		{
			name: "valid RSA 2048 key",
			keyFunc: func() any {
				key, _ := rsa.GenerateKey(crand.Reader, 2048)

				return key
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "valid EC P-256 key",
			keyFunc: func() any {
				key, _ := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)

				return key
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "valid Ed25519 key",
			keyFunc: func() any {
				_, key, _ := ed25519.GenerateKey(crand.Reader)

				return key
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "nil key",
			keyFunc:   func() any { return nil },
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "unknown key type",
			keyFunc:   func() any { return "invalid key" },
			wantValid: true, // Unknown types generate warning but are valid.
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key := tc.keyFunc()
			result, err := validator.ValidatePrivateKey(ctx, key)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantValid, result.Valid)
			}
		})
	}
}

func TestValidator_ValidateCSR(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewValidator(nil)

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	tests := []struct {
		name      string
		csrFunc   func() *x509.CertificateRequest
		wantValid bool
		wantErr   bool
	}{
		{
			name: "valid CSR with SAN",
			csrFunc: func() *x509.CertificateRequest {
				return createTestCSR(t, key, []string{"example.com"})
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "CSR without SAN",
			csrFunc: func() *x509.CertificateRequest {
				return createTestCSR(t, key, nil)
			},
			wantValid: true, // Warning but still valid.
			wantErr:   false,
		},
		{
			name:      "nil CSR",
			csrFunc:   func() *x509.CertificateRequest { return nil },
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			csr := tc.csrFunc()
			result, err := validator.ValidateCSR(ctx, csr)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantValid, result.Valid)
			}
		})
	}
}

func TestValidator_WeakAlgorithms(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := DefaultConfig()
	config.DisallowWeakAlgorithms = true
	validator := NewValidator(config)

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	// Create certificate with weak algorithm indicator.
	cert := createTestCert(t, key, false, time.Now().UTC(), time.Now().UTC().Add(365*24*time.Hour))

	// The test certificate uses ECDSA with SHA256 which is not weak.
	result, err := validator.ValidateCertificate(ctx, cert)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Empty(t, result.Vulnerabilities)
}

func TestThreatModelBuilder(t *testing.T) {
	t.Parallel()

	builder := NewThreatModelBuilder("Test Model", "1.0.0")
	require.NotNil(t, builder)

	builder.WithDescription("Test description")
	builder.AddAsset(Asset{
		ID:          "ASSET-001",
		Name:        "Test Asset",
		Description: "A test asset",
		Type:        "test",
		Sensitivity: "high",
	})
	builder.AddThreat(Threat{
		ID:          "THREAT-001",
		Category:    ThreatSpoofing,
		Title:       "Test Threat",
		Description: "A test threat",
		Asset:       "ASSET-001",
		Severity:    SeverityHigh,
		Status:      "open",
	})
	builder.AddControl(Control{
		ID:          "CTRL-001",
		Name:        "Test Control",
		Description: "A test control",
		Type:        "technical",
		Mitigates:   []string{"THREAT-001"},
		Status:      "implemented",
	})

	model := builder.Build()

	require.Equal(t, "Test Model", model.Name)
	require.Equal(t, "1.0.0", model.Version)
	require.Equal(t, "Test description", model.Description)
	require.Len(t, model.Assets, 1)
	require.Len(t, model.Threats, 1)
	require.Len(t, model.Controls, 1)
}

func TestCAThreatModel(t *testing.T) {
	t.Parallel()

	model := CAThreatModel()

	require.NotNil(t, model)
	require.Equal(t, "CA Security Threat Model", model.Name)
	require.Equal(t, "1.0.0", model.Version)
	require.NotEmpty(t, model.Assets)
	require.NotEmpty(t, model.Threats)
	require.NotEmpty(t, model.Controls)

	// Verify STRIDE coverage.
	categories := make(map[ThreatCategory]bool)
	for _, threat := range model.Threats {
		categories[threat.Category] = true
	}

	require.True(t, categories[ThreatSpoofing])
	require.True(t, categories[ThreatTampering])
	require.True(t, categories[ThreatRepudiation])
	require.True(t, categories[ThreatInformationDisclose])
	require.True(t, categories[ThreatDenialOfService])
	require.True(t, categories[ThreatElevationPrivilege])
}

func TestScanner_ScanCertificateChain(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	scanner := NewScanner(nil)

	// Generate root CA.
	rootKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	rootCert := createTestCACert(t, rootKey, nil, nil, "Root CA")

	// Generate intermediate CA signed by root.
	intKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	intCert := createTestCACert(t, intKey, rootCert, rootKey, "Intermediate CA")

	// Generate leaf certificate signed by intermediate.
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	leafCert := createTestLeafCert(t, leafKey, intCert, intKey, "leaf.example.com")

	tests := []struct {
		name      string
		chain     []*x509.Certificate
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "valid chain",
			chain:     []*x509.Certificate{leafCert, intCert, rootCert},
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "single certificate",
			chain:     []*x509.Certificate{rootCert},
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "empty chain",
			chain:     []*x509.Certificate{},
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := scanner.ScanCertificateChain(ctx, tc.chain)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.wantValid, result.Valid)
			}
		})
	}
}

func TestScanner_InvalidChainLinkage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	scanner := NewScanner(nil)

	// Generate two unrelated CAs.
	key1, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	cert1 := createTestCACert(t, key1, nil, nil, "CA 1")

	key2, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	cert2 := createTestCACert(t, key2, nil, nil, "CA 2")

	// Create an invalid chain where certs are not actually linked.
	result, err := scanner.ScanCertificateChain(ctx, []*x509.Certificate{cert1, cert2})
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)
}

func TestGenerateReport(t *testing.T) {
	t.Parallel()

	threatModel := CAThreatModel()

	validations := []ValidationResult{
		{
			Valid:    true,
			Errors:   []string{},
			Warnings: []string{"test warning"},
			Vulnerabilities: []Vulnerability{
				{ID: "VULN-001", Severity: SeverityHigh},
				{ID: "VULN-002", Severity: SeverityMedium},
			},
		},
		{
			Valid:    false,
			Errors:   []string{"test error"},
			Warnings: []string{},
			Vulnerabilities: []Vulnerability{
				{ID: "VULN-003", Severity: SeverityCritical},
			},
		},
	}

	report := GenerateReport(threatModel, validations)

	require.NotNil(t, report)
	require.NotZero(t, report.GeneratedAt)
	require.Equal(t, threatModel, report.ThreatModel)
	require.Len(t, report.Validations, 2)

	// Verify summary.
	require.Equal(t, len(threatModel.Threats), report.Summary.TotalThreats)
	require.Equal(t, 3, report.Summary.TotalVulnerabilities)
	require.Equal(t, 1, report.Summary.CriticalCount)
	require.Equal(t, 1, report.Summary.HighCount)
	require.Equal(t, 1, report.Summary.MediumCount)
}

func TestGenerateReport_NilThreatModel(t *testing.T) {
	t.Parallel()

	report := GenerateReport(nil, nil)

	require.NotNil(t, report)
	require.Nil(t, report.ThreatModel)
	require.Equal(t, 0, report.Summary.TotalThreats)
}

func TestThreatCategory_Values(t *testing.T) {
	t.Parallel()

	require.Equal(t, ThreatCategory("spoofing"), ThreatSpoofing)
	require.Equal(t, ThreatCategory("tampering"), ThreatTampering)
	require.Equal(t, ThreatCategory("repudiation"), ThreatRepudiation)
	require.Equal(t, ThreatCategory("information_disclosure"), ThreatInformationDisclose)
	require.Equal(t, ThreatCategory("denial_of_service"), ThreatDenialOfService)
	require.Equal(t, ThreatCategory("elevation_of_privilege"), ThreatElevationPrivilege)
}

func TestSeverity_Values(t *testing.T) {
	t.Parallel()

	require.Equal(t, Severity("critical"), SeverityCritical)
	require.Equal(t, Severity("high"), SeverityHigh)
	require.Equal(t, Severity("medium"), SeverityMedium)
	require.Equal(t, Severity("low"), SeverityLow)
	require.Equal(t, Severity("info"), SeverityInfo)
}

// Helper functions.

func createTestCert(t *testing.T, key *ecdsa.PrivateKey, isCA bool, notBefore, notAfter time.Time) *x509.Certificate {
	t.Helper()

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test Certificate",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
		DNSNames:              []string{"test.example.com"},
	}

	certBytes, err := x509.CreateCertificate(crand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	return cert
}

func createTestCertWithKey(t *testing.T, key any, isCA bool, notBefore, notAfter time.Time) *x509.Certificate {
	t.Helper()

	var (
		pub    any
		sigAlg x509.SignatureAlgorithm
	)

	switch k := key.(type) {
	case *rsa.PrivateKey:
		pub = &k.PublicKey
		sigAlg = x509.SHA256WithRSA
	case *ecdsa.PrivateKey:
		pub = &k.PublicKey
		sigAlg = x509.ECDSAWithSHA256
	case ed25519.PrivateKey:
		pub = k.Public()
		sigAlg = x509.PureEd25519
	default:
		require.FailNow(t, "unsupported key type", "%T", key)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test Certificate",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		SignatureAlgorithm:    sigAlg,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
		DNSNames:              []string{"test.example.com"},
	}

	certBytes, err := x509.CreateCertificate(crand.Reader, template, template, pub, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	return cert
}

func createTestCSR(t *testing.T, key *ecdsa.PrivateKey, dnsNames []string) *x509.CertificateRequest {
	t.Helper()

	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "Test CSR",
		},
		DNSNames: dnsNames,
	}

	csrBytes, err := x509.CreateCertificateRequest(crand.Reader, template, key)
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrBytes)
	require.NoError(t, err)

	return csr
}

func createTestCACert(t *testing.T, key *ecdsa.PrivateKey, parent *x509.Certificate, parentKey *ecdsa.PrivateKey, cn string) *x509.Certificate {
	t.Helper()

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UTC().UnixNano()),
		Subject: pkix.Name{
			CommonName: cn,
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	if parent == nil {
		parent = template
		parentKey = key
	}

	certBytes, err := x509.CreateCertificate(crand.Reader, template, parent, &key.PublicKey, parentKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	return cert
}

func createTestLeafCert(t *testing.T, key *ecdsa.PrivateKey, parent *x509.Certificate, parentKey *ecdsa.PrivateKey, cn string) *x509.Certificate {
	t.Helper()

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UTC().UnixNano()),
		Subject: pkix.Name{
			CommonName: cn,
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              []string{cn},
	}

	certBytes, err := x509.CreateCertificate(crand.Reader, template, parent, &key.PublicKey, parentKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	return cert
}
