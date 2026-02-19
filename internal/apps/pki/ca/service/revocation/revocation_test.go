// Copyright (c) 2025 Justin Cranford

package revocation

import (
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
)

func TestRevocationReason_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		reason RevocationReason
		want   string
	}{
		{"unspecified", ReasonUnspecified, "unspecified"},
		{"keyCompromise", ReasonKeyCompromise, "keyCompromise"},
		{"caCompromise", ReasonCACompromise, "caCompromise"},
		{"affiliationChanged", ReasonAffiliationChanged, "affiliationChanged"},
		{"superseded", ReasonSuperseded, "superseded"},
		{"cessationOfOperation", ReasonCessationOfOperation, "cessationOfOperation"},
		{"certificateHold", ReasonCertificateHold, "certificateHold"},
		{"removeFromCRL", ReasonRemoveFromCRL, "removeFromCRL"},
		{"privilegeWithdrawn", ReasonPrivilegeWithdrawn, "privilegeWithdrawn"},
		{"aaCompromise", ReasonAACompromise, "aaCompromise"},
		{"unknown", RevocationReason(100), "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.reason.String()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestNewCRLService(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	tests := []struct {
		name    string
		config  *CRLConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil-config",
			config:  nil,
			wantErr: true,
			errMsg:  "config is required",
		},
		{
			name: "nil-issuer",
			config: &CRLConfig{
				Issuer:     nil,
				PrivateKey: caKey,
				Provider:   provider,
			},
			wantErr: true,
			errMsg:  "issuer certificate is required",
		},
		{
			name: "nil-private-key",
			config: &CRLConfig{
				Issuer:     caCert,
				PrivateKey: nil,
				Provider:   provider,
			},
			wantErr: true,
			errMsg:  "private key is required",
		},
		{
			name: "nil-provider",
			config: &CRLConfig{
				Issuer:     caCert,
				PrivateKey: caKey,
				Provider:   nil,
			},
			wantErr: true,
			errMsg:  "crypto provider is required",
		},
		{
			name: "valid-config",
			config: &CRLConfig{
				Issuer:     caCert,
				PrivateKey: caKey,
				Provider:   provider,
				Validity:   24 * time.Hour,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, err := NewCRLService(tc.config)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				require.Nil(t, svc)
			} else {
				require.NoError(t, err)
				require.NotNil(t, svc)
			}
		})
	}
}

func TestCRLService_Revoke(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	config := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   24 * time.Hour,
	}

	svc, err := NewCRLService(config)
	require.NoError(t, err)

	// Test nil serial number.
	err = svc.Revoke(nil, ReasonUnspecified)
	require.Error(t, err)
	require.Contains(t, err.Error(), "serial number is required")

	// Test successful revocation.
	serial := big.NewInt(12345)
	err = svc.Revoke(serial, ReasonKeyCompromise)
	require.NoError(t, err)

	// Test duplicate revocation.
	err = svc.Revoke(serial, ReasonSuperseded)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already revoked")
}

func TestCRLService_IsRevoked(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	config := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   24 * time.Hour,
	}

	svc, err := NewCRLService(config)
	require.NoError(t, err)

	serial := big.NewInt(12345)

	// Not revoked initially.
	revoked, entry := svc.IsRevoked(serial)
	require.False(t, revoked)
	require.Nil(t, entry)

	// Revoke the certificate.
	err = svc.Revoke(serial, ReasonKeyCompromise)
	require.NoError(t, err)

	// Now it should be revoked.
	revoked, entry = svc.IsRevoked(serial)
	require.True(t, revoked)
	require.NotNil(t, entry)
	require.Equal(t, serial, entry.SerialNumber)
	require.Equal(t, ReasonKeyCompromise, entry.Reason)
}

func TestCRLService_GenerateCRL(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	config := &CRLConfig{
		Issuer:           caCert,
		PrivateKey:       caKey,
		Provider:         provider,
		Validity:         24 * time.Hour,
		NextUpdateBuffer: time.Hour,
	}

	svc, err := NewCRLService(config)
	require.NoError(t, err)

	// Add some revoked certs.
	err = svc.Revoke(big.NewInt(111), ReasonKeyCompromise)
	require.NoError(t, err)

	err = svc.Revoke(big.NewInt(222), ReasonSuperseded)
	require.NoError(t, err)

	// Generate CRL.
	crlDER, err := svc.GenerateCRL()
	require.NoError(t, err)
	require.NotEmpty(t, crlDER)

	// Parse and verify CRL.
	crl, err := x509.ParseRevocationList(crlDER)
	require.NoError(t, err)
	require.Equal(t, 2, len(crl.RevokedCertificateEntries))
}

func TestCRLService_GenerateCRLPEM(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	config := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   24 * time.Hour,
	}

	svc, err := NewCRLService(config)
	require.NoError(t, err)

	// Generate PEM CRL.
	crlPEM, err := svc.GenerateCRLPEM()
	require.NoError(t, err)
	require.NotEmpty(t, crlPEM)

	// Verify it's valid PEM.
	block, _ := pem.Decode(crlPEM)
	require.NotNil(t, block)
	require.Equal(t, "X509 CRL", block.Type)
}
