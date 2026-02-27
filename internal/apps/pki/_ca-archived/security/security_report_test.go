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
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
		NotAfter:              time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
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
		NotAfter:              time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
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
