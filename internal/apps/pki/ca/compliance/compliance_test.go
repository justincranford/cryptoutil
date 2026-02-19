// Copyright (c) 2025 Justin Cranford

package compliance

import (
	"bytes"
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
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
				return createTestCert(t, key, false, time.Now().UTC(), time.Now().UTC().Add(365*24*time.Hour), []string{"example.com"})
			},
			wantErr: false,
		},
		{
			name:      "rfc5280 valid cert",
			framework: FrameworkRFC5280,
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now().UTC(), time.Now().UTC().Add(365*24*time.Hour), []string{"example.com"})
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
				return createTestCert(t, key, false, time.Now().UTC(), time.Now().UTC().Add(365*24*time.Hour), []string{"example.com"})
			},
			expectCompliant: true,
		},
		{
			name: "certificate exceeds validity period",
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now().UTC(), time.Now().UTC().Add(500*24*time.Hour), []string{"example.com"})
			},
			expectCompliant: false,
		},
		{
			name: "certificate without SAN",
			certFunc: func() *x509.Certificate {
				return createTestCert(t, key, false, time.Now().UTC(), time.Now().UTC().Add(365*24*time.Hour), nil)
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

	cert := createTestCert(t, key, false, time.Now().UTC(), time.Now().UTC().Add(365*24*time.Hour), []string{"example.com"})

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
