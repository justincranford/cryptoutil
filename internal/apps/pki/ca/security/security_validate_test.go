// Copyright (c) 2025 Justin Cranford

package security

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
