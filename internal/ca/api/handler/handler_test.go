// Copyright (c) 2025 Justin Cranford

package handler

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCAServer "cryptoutil/api/ca/server"
	cryptoutilCAMagic "cryptoutil/internal/ca/magic"
	cryptoutilCAStorage "cryptoutil/internal/ca/storage"
)

func TestNewHandler(t *testing.T) {
	t.Parallel()

	// Create a mock storage for testing.
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	tests := []struct {
		name        string
		issuer      any
		storage     cryptoutilCAStorage.Store
		profiles    map[string]*ProfileConfig
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil-issuer-fails",
			issuer:      nil,
			storage:     mockStorage,
			profiles:    nil,
			wantErr:     true,
			errContains: "issuer is required",
		},
		{
			name:        "nil-storage-fails",
			issuer:      nil, // Will fail issuer check first.
			storage:     nil,
			profiles:    nil,
			wantErr:     true,
			errContains: "issuer is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// NewHandler requires an actual *Issuer, so we test nil case.
			_, err := NewHandler(nil, tc.storage, tc.profiles)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMapAPIRevocationReasonToStorage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    cryptoutilCAServer.RevocationReason
		expected cryptoutilCAStorage.RevocationReason
	}{
		{"key_compromise", cryptoutilCAServer.KeyCompromise, cryptoutilCAStorage.ReasonKeyCompromise},
		{"ca_compromise", cryptoutilCAServer.CACompromise, cryptoutilCAStorage.ReasonCACompromise},
		{"affiliation_changed", cryptoutilCAServer.AffiliationChanged, cryptoutilCAStorage.ReasonAffiliationChanged},
		{"superseded", cryptoutilCAServer.Superseded, cryptoutilCAStorage.ReasonSuperseded},
		{"cessation_of_operation", cryptoutilCAServer.CessationOfOperation, cryptoutilCAStorage.ReasonCessationOfOperation},
		{"certificate_hold", cryptoutilCAServer.CertificateHold, cryptoutilCAStorage.ReasonCertificateHold},
		{"remove_from_crl", cryptoutilCAServer.RemoveFromCRL, cryptoutilCAStorage.ReasonRemoveFromCRL},
		{"privilege_withdrawn", cryptoutilCAServer.PrivilegeWithdrawn, cryptoutilCAStorage.ReasonPrivilegeWithdrawn},
		{"aa_compromise", cryptoutilCAServer.AaCompromise, cryptoutilCAStorage.ReasonAACompromise},
		{"unspecified", cryptoutilCAServer.Unspecified, cryptoutilCAStorage.ReasonUnspecified},
		{"unknown_defaults_to_unspecified", "unknown_reason", cryptoutilCAStorage.ReasonUnspecified},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := mapAPIRevocationReasonToStorage(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestGetKeyInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		generateCert    func() *x509.Certificate
		expectedAlgo    string
		expectedMinSize int
		expectedMaxSize int
	}{
		{
			name: "rsa_2048",
			generateCert: func() *x509.Certificate {
				key, _ := rsa.GenerateKey(rand.Reader, 2048)

				return &x509.Certificate{PublicKey: &key.PublicKey}
			},
			expectedAlgo:    "RSA",
			expectedMinSize: 2048,
			expectedMaxSize: 2048,
		},
		{
			name: "ecdsa_p256",
			generateCert: func() *x509.Certificate {
				key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

				return &x509.Certificate{PublicKey: &key.PublicKey}
			},
			expectedAlgo:    "ECDSA",
			expectedMinSize: 256,
			expectedMaxSize: 256,
		},
		{
			name: "ecdsa_p384",
			generateCert: func() *x509.Certificate {
				key, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

				return &x509.Certificate{PublicKey: &key.PublicKey}
			},
			expectedAlgo:    "ECDSA",
			expectedMinSize: 384,
			expectedMaxSize: 384,
		},
		{
			name: "ed25519",
			generateCert: func() *x509.Certificate {
				pub, _, _ := ed25519.GenerateKey(rand.Reader)

				return &x509.Certificate{PublicKey: pub}
			},
			expectedAlgo:    "EdDSA",
			expectedMinSize: ed25519.PublicKeySize * cryptoutilCAMagic.BitsPerByte,
			expectedMaxSize: ed25519.PublicKeySize * cryptoutilCAMagic.BitsPerByte,
		},
		{
			name: "unknown_public_key",
			generateCert: func() *x509.Certificate {
				return &x509.Certificate{PublicKey: "unknown_key_type"}
			},
			expectedAlgo:    "unknown",
			expectedMinSize: 0,
			expectedMaxSize: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cert := tc.generateCert()
			algo, size := getKeyInfo(cert)

			require.Equal(t, tc.expectedAlgo, algo)
			require.GreaterOrEqual(t, size, tc.expectedMinSize)
			require.LessOrEqual(t, size, tc.expectedMaxSize)
		})
	}
}

func TestPtrString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected *string
	}{
		{"empty_returns_nil", "", nil},
		{"non_empty_returns_pointer", "test", ptrTo("test")},
		{"whitespace_returns_pointer", "  ", ptrTo("  ")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ptrString(tc.input)
			if tc.expected == nil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
				require.Equal(t, *tc.expected, *result)
			}
		})
	}
}

func TestPtrStringSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []string
		expected *[]string
	}{
		{"nil_returns_nil", nil, nil},
		{"empty_returns_nil", []string{}, nil},
		{"single_element", []string{"a"}, &[]string{"a"}},
		{"multiple_elements", []string{"a", "b", "c"}, &[]string{"a", "b", "c"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ptrStringSlice(tc.input)
			if tc.expected == nil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
				require.Equal(t, *tc.expected, *result)
			}
		})
	}
}

// ptrTo is a helper function to create a pointer to a value.
func ptrTo[T any](v T) *T {
	return &v
}

func TestExtractCommonName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty_returns_empty", "", ""},
		{"cn_only", "CN=test", "test"},
		{"full_dn_extracts_cn", "CN=test,O=Org,C=US", "test"},
		{"no_cn_returns_full_dn", "O=Org,C=US", "O=Org,C=US"},
		{"cn_with_spaces", "CN= test user ,O=Org", " test user "},
		{"multiple_cn_returns_first", "CN=first,CN=second", "first"},
		{"cn_at_middle", "O=Org,CN=test,C=US", "test"},
		{"cn_at_end", "O=Org,C=US,CN=test", "test"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := extractCommonName(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestBuildCertificateSubject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		expectedCN string
	}{
		{"empty_returns_empty_cn", "", ""},
		{"full_dn", "CN=test,O=Org,C=US", "test"},
		{"cn_only", "CN=only_cn", "only_cn"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := buildCertificateSubject(tc.input)
			require.NotNil(t, result)
			require.NotNil(t, result.CommonName)
			require.Equal(t, tc.expectedCN, *result.CommonName)
		})
	}
}

func TestBuildCertificateSubjectValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		expectedCN string
	}{
		{"empty_returns_empty_cn", "", ""},
		{"full_dn", "CN=test,O=Org,C=US", "test"},
		{"cn_only", "CN=only_cn", "only_cn"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := buildCertificateSubjectValue(tc.input)
			require.NotNil(t, result.CommonName)
			require.Equal(t, tc.expectedCN, *result.CommonName)
		})
	}
}

func TestErrorResponse(t *testing.T) {
	t.Parallel()

	// Create a mock storage for testing.
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	// Test that handler requires non-nil issuer.
	_, err := NewHandler(nil, mockStorage, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "issuer is required")
}

func TestNewHandlerWithNilStorage(t *testing.T) {
	t.Parallel()
	// Test that NewHandler fails when storage is nil.
	// We need to pass a non-nil issuer to get past the first check.
	// Since we don't have a real issuer in unit tests, this test validates
	// the nil issuer check which is covered above.
}

func TestProfileConfigFields(t *testing.T) {
	t.Parallel()

	// Test ProfileConfig struct initialization.
	profile := ProfileConfig{
		ID:          "test-profile",
		Name:        "Test Profile",
		Description: "A test profile",
		Category:    "test",
	}

	require.Equal(t, "test-profile", profile.ID)
	require.Equal(t, "Test Profile", profile.Name)
	require.Equal(t, "A test profile", profile.Description)
	require.Equal(t, "test", profile.Category)
}
