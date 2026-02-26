// Copyright (c) 2025 Justin Cranford

package compliance

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const sectionBasicConstraints = "4.2.1.9"

// failingWriter always returns an error on Write.
type failingWriter struct{}

func (fw *failingWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("simulated write failure")
}

// mockRSAPublicKey implements the rsaPublicKey interface (Size() int) with a small key size.
type mockRSAPublicKey struct{}

func (m *mockRSAPublicKey) Size() int {
	return cryptoutilSharedMagic.MinSerialNumberBits // 512 bits, well below 2048 minimum.
}

func TestAuditLogger_Log_WriteError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := NewAuditLogger(&failingWriter{})

	event := &AuditEvent{
		ID:        "EVT-001",
		EventType: EventCertificateIssued,
		Actor:     "admin@example.com",
		Resource:  "cert-123",
		Action:    "issue",
		Outcome:   "success",
	}

	err := logger.Log(ctx, event)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write audit event")
}

func TestChecker_CheckCertificate_WebTrustFramework(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	checker := NewChecker(FrameworkWebTrust)

	cert := &x509.Certificate{
		SerialNumber: new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().UTC(),
		NotAfter:     time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{"example.com"},
	}

	requirements, err := checker.CheckCertificate(ctx, cert)
	require.NoError(t, err)
	require.NotEmpty(t, requirements)
}

func TestChecker_RFC5280_CACertificate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	checker := NewChecker(FrameworkRFC5280)

	// CA cert with valid BasicConstraints (covers evaluateBasicConstraints5280 compliant path).
	cert := &x509.Certificate{
		SerialNumber:          new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}

	requirements, err := checker.CheckCertificate(ctx, cert)
	require.NoError(t, err)

	// Should include Basic Constraints requirement.
	hasBC := false

	for _, req := range requirements {
		if req.Section == sectionBasicConstraints {
			hasBC = true

			require.Equal(t, StatusCompliant, req.Status)
		}
	}

	require.True(t, hasBC, "should include Basic Constraints check for CA cert")
}

func TestChecker_RFC5280_CACert_NonCompliant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	checker := NewChecker(FrameworkRFC5280)

	// CA cert WITHOUT valid BasicConstraints (covers evaluateBasicConstraints5280 non-compliant path).
	cert := &x509.Certificate{
		SerialNumber:          new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: false, // Non-compliant.
		KeyUsage:              x509.KeyUsageCertSign,
	}

	requirements, err := checker.CheckCertificate(ctx, cert)
	require.NoError(t, err)

	for _, req := range requirements {
		if req.Section == sectionBasicConstraints {
			require.Equal(t, StatusNonCompliant, req.Status)
		}
	}
}

func TestChecker_RFC5280_NonCompliantCert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	checker := NewChecker(FrameworkRFC5280)

	tests := []struct {
		name    string
		cert    *x509.Certificate
		section string
		status  Status
	}{
		{
			name: "negative serial number",
			cert: &x509.Certificate{
				SerialNumber: big.NewInt(-1),
				NotBefore:    time.Now().UTC(),
				NotAfter:     time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
			},
			section: "4.1.2.2",
			status:  StatusNonCompliant,
		},
		{
			name: "invalid validity period",
			cert: &x509.Certificate{
				SerialNumber: new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits),
				NotBefore:    time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				NotAfter:     time.Now().UTC(),
			},
			section: "4.1.2.5",
			status:  StatusNonCompliant,
		},
		{
			name: "unknown critical extension",
			cert: &x509.Certificate{
				SerialNumber: new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits),
				NotBefore:    time.Now().UTC(),
				NotAfter:     time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				Extensions: []pkix.Extension{
					{
						Id:       asn1.ObjectIdentifier{1, 2, 3, 4, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, cryptoutilSharedMagic.DefaultEmailOTPLength, cryptoutilSharedMagic.GitRecentActivityDays},
						Critical: true,
						Value:    []byte{0x01},
					},
				},
			},
			section: "4.2",
			status:  StatusNonCompliant,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			requirements, err := checker.CheckCertificate(ctx, tc.cert)
			require.NoError(t, err)

			for _, req := range requirements {
				if req.Section == tc.section {
					require.Equal(t, tc.status, req.Status,
						"section %s should be %s", tc.section, tc.status)
				}
			}
		})
	}
}

func TestChecker_CABF_NonCompliantCert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	checker := NewChecker(FrameworkCABFBaseline)

	tests := []struct {
		name    string
		cert    *x509.Certificate
		checkID string
		status  Status
	}{
		{
			name: "empty subject and no SAN",
			cert: &x509.Certificate{
				SerialNumber: new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits),
				Subject:      pkix.Name{},
				NotBefore:    time.Now().UTC(),
				NotAfter:     time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				KeyUsage:     x509.KeyUsageDigitalSignature,
			},
			checkID: "BR-7.1.2.1",
			status:  StatusNonCompliant,
		},
		{
			name: "small serial number",
			cert: &x509.Certificate{
				SerialNumber: big.NewInt(1),
				Subject:      pkix.Name{CommonName: "test"},
				NotBefore:    time.Now().UTC(),
				NotAfter:     time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				KeyUsage:     x509.KeyUsageDigitalSignature,
				DNSNames:     []string{"example.com"},
			},
			checkID: "BR-7.1.2.2",
			status:  StatusNonCompliant,
		},
		{
			name: "no key usage",
			cert: &x509.Certificate{
				SerialNumber: new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits),
				Subject:      pkix.Name{CommonName: "test"},
				NotBefore:    time.Now().UTC(),
				NotAfter:     time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				KeyUsage:     0,
				DNSNames:     []string{"example.com"},
			},
			checkID: "BR-7.1.2.4-KU",
			status:  StatusNonCompliant,
		},
		{
			name: "CA cert without BasicConstraintsValid",
			cert: &x509.Certificate{
				SerialNumber:          new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits),
				Subject:               pkix.Name{CommonName: "test CA"},
				NotBefore:             time.Now().UTC(),
				NotAfter:              time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				KeyUsage:              x509.KeyUsageCertSign,
				IsCA:                  true,
				BasicConstraintsValid: false,
			},
			checkID: "BR-7.1.2.4-BC",
			status:  StatusNonCompliant,
		},
		{
			name: "small RSA key",
			cert: &x509.Certificate{
				SerialNumber: new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits),
				Subject:      pkix.Name{CommonName: "test"},
				NotBefore:    time.Now().UTC(),
				NotAfter:     time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				KeyUsage:     x509.KeyUsageDigitalSignature,
				DNSNames:     []string{"example.com"},
				PublicKey:    &mockRSAPublicKey{},
			},
			checkID: "BR-6.1.5",
			status:  StatusNonCompliant,
		},
		{
			name: "weak signature algorithm",
			cert: &x509.Certificate{
				SerialNumber:       new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits),
				Subject:            pkix.Name{CommonName: "test"},
				NotBefore:          time.Now().UTC(),
				NotAfter:           time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				KeyUsage:           x509.KeyUsageDigitalSignature,
				DNSNames:           []string{"example.com"},
				SignatureAlgorithm: x509.SHA1WithRSA,
			},
			checkID: "BR-7.1.3",
			status:  StatusNonCompliant,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			requirements, err := checker.CheckCertificate(ctx, tc.cert)
			require.NoError(t, err)

			for _, req := range requirements {
				if req.ID == tc.checkID {
					require.Equal(t, tc.status, req.Status,
						"check %s should be %s", tc.checkID, tc.status)
				}
			}
		})
	}
}

func TestGenerateReport_MediumAndLowSeverity(t *testing.T) {
	t.Parallel()

	requirements := []Requirement{
		{ID: "REQ-001", Status: StatusNonCompliant, Severity: SeverityMedium},
		{ID: "REQ-002", Status: StatusNonCompliant, Severity: SeverityLow},
	}

	period := AuditPeriod{
		StartDate: time.Now().UTC().Add(-cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		EndDate:   time.Now().UTC(),
	}

	report := GenerateReport(FrameworkCABFBaseline, requirements, period, "test-auditor")
	require.NotNil(t, report)
	require.Equal(t, 2, report.Summary.TotalRequirements)
	require.Equal(t, 2, report.Summary.NonCompliant)
	require.Equal(t, 1, report.Summary.MediumFindings)
	require.Equal(t, 1, report.Summary.LowFindings)
}
