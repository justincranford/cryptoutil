// Copyright (c) 2025 Justin Cranford

package compliance

import (
	"bytes"
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewAuditLogger(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := NewAuditLogger(&buf)

	require.NotNil(t, logger)
}

func TestAuditLogger_Log(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		event   *AuditEvent
		wantErr bool
	}{
		{
			name: "valid event",
			event: &AuditEvent{
				ID:        "EVT-001",
				EventType: EventCertificateIssued,
				Actor:     "admin@example.com",
				Resource:  "cert-123",
				Action:    "issue",
				Outcome:   "success",
			},
			wantErr: false,
		},
		{
			name:    "nil event",
			event:   nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger := NewAuditLogger(&buf)

			err := logger.Log(ctx, tc.event)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, buf.String())
			}
		})
	}
}

func TestAuditLogger_AddWriter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var buf1, buf2 bytes.Buffer

	logger := NewAuditLogger(&buf1)
	logger.AddWriter(&buf2)

	event := &AuditEvent{
		ID:        "EVT-001",
		EventType: EventCertificateIssued,
		Actor:     "admin@example.com",
		Resource:  "cert-123",
		Action:    "issue",
		Outcome:   "success",
	}

	err := logger.Log(ctx, event)
	require.NoError(t, err)

	// Both writers should have the event.
	require.NotEmpty(t, buf1.String())
	require.NotEmpty(t, buf2.String())
	require.Equal(t, buf1.String(), buf2.String())
}

func TestNewChecker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		framework Framework
	}{
		{name: "cabf baseline", framework: FrameworkCABFBaseline},
		{name: "rfc5280", framework: FrameworkRFC5280},
		{name: "webtrust", framework: FrameworkWebTrust},
		{name: "nist", framework: FrameworkNIST80057},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			checker := NewChecker(tc.framework)
			require.NotNil(t, checker)
		})
	}
}

func TestChecker_CheckCertificate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	tests := []struct {
		name      string
		framework Framework
		certFunc  func() *x509.Certificate
		wantErr   bool
	}{
		{
			name:      "cabf valid cert",
			framework: FrameworkCABFBaseline,
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now(), time.Now().Add(365*24*time.Hour), []string{"example.com"})
			},
			wantErr: false,
		},
		{
			name:      "rfc5280 valid cert",
			framework: FrameworkRFC5280,
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now(), time.Now().Add(365*24*time.Hour), []string{"example.com"})
			},
			wantErr: false,
		},
		{
			name:      "nil certificate",
			framework: FrameworkCABFBaseline,
			certFunc:  func() *x509.Certificate { return nil },
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			checker := NewChecker(tc.framework)
			cert := tc.certFunc()

			requirements, err := checker.CheckCertificate(ctx, cert)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, requirements)
			}
		})
	}
}

func TestChecker_CABFRequirements(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	checker := NewChecker(FrameworkCABFBaseline)

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	tests := []struct {
		name            string
		certFunc        func() *x509.Certificate
		expectCompliant bool
	}{
		{
			name: "compliant certificate",
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now(), time.Now().Add(365*24*time.Hour), []string{"example.com"})
			},
			expectCompliant: true,
		},
		{
			name: "certificate exceeds validity period",
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now(), time.Now().Add(500*24*time.Hour), []string{"example.com"})
			},
			expectCompliant: false,
		},
		{
			name: "certificate without SAN",
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now(), time.Now().Add(365*24*time.Hour), nil)
			},
			expectCompliant: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cert := tc.certFunc()
			requirements, err := checker.CheckCertificate(ctx, cert)

			require.NoError(t, err)
			require.NotEmpty(t, requirements)

			// Check if any requirement is non-compliant.
			hasNonCompliant := false

			for _, req := range requirements {
				if req.Status == StatusNonCompliant {
					hasNonCompliant = true

					break
				}
			}

			if tc.expectCompliant {
				require.False(t, hasNonCompliant, "expected certificate to be compliant")
			} else {
				require.True(t, hasNonCompliant, "expected certificate to have compliance issues")
			}
		})
	}
}

func TestChecker_RFC5280Requirements(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	checker := NewChecker(FrameworkRFC5280)

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	cert := createTestCert(t, key, false, time.Now(), time.Now().Add(365*24*time.Hour), []string{"example.com"})

	requirements, err := checker.CheckCertificate(ctx, cert)
	require.NoError(t, err)
	require.NotEmpty(t, requirements)

	// Verify RFC 5280 specific requirements are present.
	hasSerialNumber := false
	hasValidity := false

	for _, req := range requirements {
		if req.Section == "4.1.2.2" {
			hasSerialNumber = true
		}

		if req.Section == "4.1.2.5" {
			hasValidity = true
		}
	}

	require.True(t, hasSerialNumber, "should check serial number")
	require.True(t, hasValidity, "should check validity")
}

func TestChecker_CACertificate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	checker := NewChecker(FrameworkCABFBaseline)

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	// Create CA certificate.
	caCert := createTestCACert(t, key)

	requirements, err := checker.CheckCertificate(ctx, caCert)
	require.NoError(t, err)

	// CA certificates should have Basic Constraints requirement.
	hasBasicConstraints := false

	for _, req := range requirements {
		if req.ID == "BR-7.1.2.4-BC" {
			hasBasicConstraints = true

			require.Equal(t, StatusCompliant, req.Status)
		}
	}

	require.True(t, hasBasicConstraints, "should check basic constraints for CA")
}

func TestGenerateReport(t *testing.T) {
	t.Parallel()

	requirements := []Requirement{
		{ID: "REQ-001", Status: StatusCompliant, Severity: SeverityMedium},
		{ID: "REQ-002", Status: StatusNonCompliant, Severity: SeverityCritical},
		{ID: "REQ-003", Status: StatusNonCompliant, Severity: SeverityHigh},
		{ID: "REQ-004", Status: StatusPartial, Severity: SeverityMedium},
		{ID: "REQ-005", Status: StatusNotApplicable, Severity: SeverityLow},
	}

	period := AuditPeriod{
		StartDate: time.Now().Add(-30 * 24 * time.Hour),
		EndDate:   time.Now(),
	}

	report := GenerateReport(FrameworkCABFBaseline, requirements, period, "test-auditor")

	require.NotNil(t, report)
	require.NotEmpty(t, report.ID)
	require.Equal(t, FrameworkCABFBaseline, report.Framework)
	require.Equal(t, "test-auditor", report.GeneratedBy)
	require.Len(t, report.Requirements, 5)

	// Verify summary.
	require.Equal(t, 5, report.Summary.TotalRequirements)
	require.Equal(t, 1, report.Summary.Compliant)
	require.Equal(t, 2, report.Summary.NonCompliant)
	require.Equal(t, 1, report.Summary.Partial)
	require.Equal(t, 1, report.Summary.NotApplicable)
	require.Equal(t, 1, report.Summary.CriticalFindings)
	require.Equal(t, 1, report.Summary.HighFindings)
}

func TestAuditTrailBuilder(t *testing.T) {
	t.Parallel()

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	builder := NewAuditTrailBuilder(startTime, endTime)
	require.NotNil(t, builder)

	// Add events.
	builder.AddEvent(&AuditEvent{
		ID:        "EVT-001",
		EventType: EventCertificateIssued,
		Actor:     "admin",
		Resource:  "cert-1",
		Action:    "issue",
		Outcome:   "success",
	})
	builder.AddEvent(&AuditEvent{
		ID:        "EVT-002",
		EventType: EventCertificateRevoked,
		Actor:     "admin",
		Resource:  "cert-2",
		Action:    "revoke",
		Outcome:   "success",
	})
	builder.AddEvent(&AuditEvent{
		ID:        "EVT-003",
		EventType: EventAuthenticationFailed,
		Actor:     "attacker",
		Resource:  "login",
		Action:    "authenticate",
		Outcome:   "failure",
	})
	builder.AddEvent(nil) // Should be ignored.

	trail := builder.Build()

	require.NotNil(t, trail)
	require.Equal(t, startTime, trail.StartTime)
	require.Equal(t, endTime, trail.EndTime)
	require.Len(t, trail.Events, 3)

	// Verify summary.
	require.Equal(t, 3, trail.Summary.TotalEvents)
	require.Equal(t, 1, trail.Summary.CertificatesIssued)
	require.Equal(t, 1, trail.Summary.CertificatesRevoked)
	require.Equal(t, 1, trail.Summary.FailedAuthentications)
}

func TestCreateCPSTemplate(t *testing.T) {
	t.Parallel()

	cps := CreateCPSTemplate()

	require.NotNil(t, cps)
	require.Equal(t, "CPS-001", cps.ID)
	require.Equal(t, "Certificate Practice Statement", cps.Title)
	require.NotEmpty(t, cps.Sections)
	require.Len(t, cps.Sections, 10)
}

func TestEvidenceCollector(t *testing.T) {
	t.Parallel()

	collector := NewEvidenceCollector()
	require.NotNil(t, collector)

	// Collect evidence.
	collector.Collect(Evidence{
		ID:          "EV-001",
		Type:        "audit_log",
		Description: "CA audit log for Q1",
		Source:      "siem",
	})
	collector.Collect(Evidence{
		ID:          "EV-002",
		Type:        "configuration",
		Description: "CA configuration snapshot",
		Source:      "config_repo",
	})
	collector.Collect(Evidence{
		ID:          "EV-003",
		Type:        "AUDIT_LOG", // Different case.
		Description: "CA audit log for Q2",
		Source:      "siem",
	})

	// Get all evidence.
	all := collector.GetEvidence()
	require.Len(t, all, 3)

	// Get evidence by type (case-insensitive).
	auditLogs := collector.GetEvidenceByType("audit_log")
	require.Len(t, auditLogs, 2)

	configs := collector.GetEvidenceByType("configuration")
	require.Len(t, configs, 1)
}

func TestFramework_Values(t *testing.T) {
	t.Parallel()

	require.Equal(t, Framework("cabf_baseline"), FrameworkCABFBaseline)
	require.Equal(t, Framework("webtrust"), FrameworkWebTrust)
	require.Equal(t, Framework("rfc5280"), FrameworkRFC5280)
	require.Equal(t, Framework("nist_sp_800_57"), FrameworkNIST80057)
}

func TestStatus_Values(t *testing.T) {
	t.Parallel()

	require.Equal(t, Status("compliant"), StatusCompliant)
	require.Equal(t, Status("non_compliant"), StatusNonCompliant)
	require.Equal(t, Status("partial"), StatusPartial)
	require.Equal(t, Status("not_applicable"), StatusNotApplicable)
}

func TestAuditEventType_Values(t *testing.T) {
	t.Parallel()

	require.Equal(t, AuditEventType("certificate_issued"), EventCertificateIssued)
	require.Equal(t, AuditEventType("certificate_revoked"), EventCertificateRevoked)
	require.Equal(t, AuditEventType("key_generated"), EventKeyGenerated)
	require.Equal(t, AuditEventType("csr_received"), EventCSRReceived)
	require.Equal(t, AuditEventType("crl_generated"), EventCRLGenerated)
}

func TestIsKnownExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		oid   string
		known bool
	}{
		{"2.5.29.14", true},  // Subject Key Identifier.
		{"2.5.29.15", true},  // Key Usage.
		{"2.5.29.17", true},  // Subject Alternative Name.
		{"2.5.29.19", true},  // Basic Constraints.
		{"2.5.29.31", true},  // CRL Distribution Points.
		{"1.2.3.4.5", false}, // Unknown OID.
	}

	for _, tc := range tests {
		t.Run(tc.oid, func(t *testing.T) {
			t.Parallel()

			result := isKnownExtension(tc.oid)
			require.Equal(t, tc.known, result)
		})
	}
}

// Helper functions.

func createTestCert(t *testing.T, key *ecdsa.PrivateKey, isCA bool, notBefore, notAfter time.Time, dnsNames []string) *x509.Certificate {
	t.Helper()

	// Use a large serial number with at least 64 bits of entropy for BR compliance.
	serialNumber, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "Test Certificate",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
		DNSNames:              dnsNames,
	}

	certBytes, err := x509.CreateCertificate(crand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	return cert
}

func createTestCACert(t *testing.T, key *ecdsa.PrivateKey) *x509.Certificate {
	t.Helper()

	// Use a large serial number with at least 64 bits of entropy for BR compliance.
	serialNumber, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	certBytes, err := x509.CreateCertificate(crand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	return cert
}
