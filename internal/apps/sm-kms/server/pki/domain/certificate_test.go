// Copyright (c) 2025-2026 Justin Cranford.
//
//

package domain

import (
	"testing"

	testify "github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCAType_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		caType   CAType
		expected string
	}{
		{name: cryptoutilSharedMagic.ROOT, caType: CATypeRoot, expected: cryptoutilSharedMagic.ROOT},
		{name: cryptoutilSharedMagic.INTERMEDIATE, caType: CATypeIntermediate, expected: cryptoutilSharedMagic.INTERMEDIATE},
		{name: cryptoutilSharedMagic.ISSUING, caType: CATypeIssuing, expected: cryptoutilSharedMagic.ISSUING},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testify.Equal(t, tc.expected, string(tc.caType))
		})
	}
}

func TestCAStatus_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   CAStatus
		expected string
	}{
		{name: cryptoutilSharedMagic.ACTIVE, status: CAStatusActive, expected: cryptoutilSharedMagic.ACTIVE},
		{name: cryptoutilSharedMagic.PENDING, status: CAStatusPending, expected: cryptoutilSharedMagic.PENDING},
		{name: cryptoutilSharedMagic.SUSPENDED, status: CAStatusSuspended, expected: cryptoutilSharedMagic.SUSPENDED},
		{name: cryptoutilSharedMagic.REVOKED, status: CAStatusRevoked, expected: cryptoutilSharedMagic.REVOKED},
		{name: cryptoutilSharedMagic.EXPIRED, status: CAStatusExpired, expected: cryptoutilSharedMagic.EXPIRED},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testify.Equal(t, tc.expected, string(tc.status))
		})
	}
}

func TestCertificateAuthority_Struct(t *testing.T) {
	t.Parallel()

	ca := CertificateAuthority{
		Name:      cryptoutilSharedMagic.TEST_CA_NAME,
		Type:      CATypeRoot,
		Status:    CAStatusPending,
		SubjectCN: cryptoutilSharedMagic.TEST_CA_CN,
		IssuerCN:  cryptoutilSharedMagic.TEST_CA_CN,
	}
	testify.Equal(t, cryptoutilSharedMagic.TEST_CA_NAME, ca.Name)
	testify.Equal(t, CATypeRoot, ca.Type)
	testify.Equal(t, CAStatusPending, ca.Status)
	testify.Nil(t, ca.ParentID)
}

func TestKeyMaterial_Struct(t *testing.T) {
	t.Parallel()

	km := KeyMaterial{}
	testify.Nil(t, km.PublicKey)
	testify.Nil(t, km.PrivateKey)
	testify.Nil(t, km.CertificateChain)
}
