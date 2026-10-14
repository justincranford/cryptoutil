// Copyright (c) 2025-2026 Justin Cranford.
package revocation

import (
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
)

func TestRevocationReason_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		reason RevocationReason
		want   string
	}{
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameUnspecified, ReasonUnspecified, cryptoutilSharedMagic.RFC5280RevocationReasonNameUnspecified},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameKeyCompromise, ReasonKeyCompromise, cryptoutilSharedMagic.RFC5280RevocationReasonNameKeyCompromise},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameCACompromise, ReasonCACompromise, cryptoutilSharedMagic.RFC5280RevocationReasonNameCACompromise},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameAffiliationChanged, ReasonAffiliationChanged, cryptoutilSharedMagic.RFC5280RevocationReasonNameAffiliationChanged},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameSuperseded, ReasonSuperseded, cryptoutilSharedMagic.RFC5280RevocationReasonNameSuperseded},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameCessationOfOperation, ReasonCessationOfOperation, cryptoutilSharedMagic.RFC5280RevocationReasonNameCessationOfOperation},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameCertificateHold, ReasonCertificateHold, cryptoutilSharedMagic.RFC5280RevocationReasonNameCertificateHold},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameRemoveFromCRL, ReasonRemoveFromCRL, cryptoutilSharedMagic.RFC5280RevocationReasonNameRemoveFromCRL},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNamePrivilegeWithdrawn, ReasonPrivilegeWithdrawn, cryptoutilSharedMagic.RFC5280RevocationReasonNamePrivilegeWithdrawn},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameAACompromise, ReasonAACompromise, cryptoutilSharedMagic.RFC5280RevocationReasonNameAACompromise},
		{cryptoutilSharedMagic.UNKNOWN, RevocationReason(cryptoutilSharedMagic.JoseJAMaxMaterials), cryptoutilSharedMagic.UNKNOWN},
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
			name:    testLiteralNilConfig,
			config:  nil,
			wantErr: true,
			errMsg:  cryptoutilSharedMagic.PKIErrorConfigRequired,
		},
		{
			name: testLiteralNilIssuer,
			config: &CRLConfig{
				Issuer:     nil,
				PrivateKey: caKey,
				Provider:   provider,
			},
			wantErr: true,
			errMsg:  cryptoutilSharedMagic.PKIErrorIssuerCertificateRequired,
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
			name: testLiteralNilProvider,
			config: &CRLConfig{
				Issuer:     caCert,
				PrivateKey: caKey,
				Provider:   nil,
			},
			wantErr: true,
			errMsg:  cryptoutilSharedMagic.PKIErrorCryptoProviderRequired,
		},
		{
			name: testLiteralValidConfig,
			config: &CRLConfig{
				Issuer:     caCert,
				PrivateKey: caKey,
				Provider:   provider,
				Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
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
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
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
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
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
		Validity:         cryptoutilSharedMagic.HoursPerDay * time.Hour,
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
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
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
