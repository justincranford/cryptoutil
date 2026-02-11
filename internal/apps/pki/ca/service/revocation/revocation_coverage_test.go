// Copyright (c) 2025 Justin Cranford

package revocation

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
)

// TestCRLService_GetRevokedCertificates tests retrieving revoked certificates.
func TestCRLService_GetRevokedCertificates(t *testing.T) {
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

	// Initially empty.
	revoked := svc.GetRevokedCertificates()
	require.Len(t, revoked, 0)

	// Revoke a certificate.
	serial1 := big.NewInt(12345)
	err = svc.Revoke(serial1, ReasonKeyCompromise)
	require.NoError(t, err)

	// Should have 1 revoked certificate.
	revoked = svc.GetRevokedCertificates()
	require.Len(t, revoked, 1)
	require.Equal(t, serial1, revoked[0].SerialNumber)
	require.Equal(t, ReasonKeyCompromise, revoked[0].Reason)

	// Revoke another certificate.
	serial2 := big.NewInt(67890)
	err = svc.Revoke(serial2, ReasonSuperseded)
	require.NoError(t, err)

	// Should have 2 revoked certificates.
	revoked = svc.GetRevokedCertificates()
	require.Len(t, revoked, 2)
}
