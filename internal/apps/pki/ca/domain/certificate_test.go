// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"testing"

	testify "github.com/stretchr/testify/require"
)

func TestCAType_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		caType   CAType
		expected string
	}{
		{name: "root", caType: CATypeRoot, expected: "root"},
		{name: "intermediate", caType: CATypeIntermediate, expected: "intermediate"},
		{name: "issuing", caType: CATypeIssuing, expected: "issuing"},
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
		{name: "active", status: CAStatusActive, expected: "active"},
		{name: "pending", status: CAStatusPending, expected: "pending"},
		{name: "suspended", status: CAStatusSuspended, expected: "suspended"},
		{name: "revoked", status: CAStatusRevoked, expected: "revoked"},
		{name: "expired", status: CAStatusExpired, expected: "expired"},
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
		Name:      "test-ca",
		Type:      CATypeRoot,
		Status:    CAStatusPending,
		SubjectCN: "Test CA",
		IssuerCN:  "Test CA",
	}
	testify.Equal(t, "test-ca", ca.Name)
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
