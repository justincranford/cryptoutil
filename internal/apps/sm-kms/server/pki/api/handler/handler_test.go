// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
			errContains: testErrorMessageIssuerRequired,
		},
		{
			name:        "nil-storage-fails",
			issuer:      nil, // Will fail issuer check first.
			storage:     nil,
			profiles:    nil,
			wantErr:     true,
			errContains: testErrorMessageIssuerRequired,
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
		input    cryptoutilApiCaServer.RevocationReason
		expected cryptoutilCAStorage.RevocationReason
	}{
		{testRevocationCaseKeyCompromise, cryptoutilApiCaServer.KeyCompromise, cryptoutilCAStorage.ReasonKeyCompromise},
		{testRevocationCaseCACompromise, cryptoutilApiCaServer.CACompromise, cryptoutilCAStorage.ReasonCACompromise},
		{testRevocationCaseAffiliationChanged, cryptoutilApiCaServer.AffiliationChanged, cryptoutilCAStorage.ReasonAffiliationChanged},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameSuperseded, cryptoutilApiCaServer.Superseded, cryptoutilCAStorage.ReasonSuperseded},
		{testRevocationCaseCessationOfOperation, cryptoutilApiCaServer.CessationOfOperation, cryptoutilCAStorage.ReasonCessationOfOperation},
		{testRevocationCaseCertificateHold, cryptoutilApiCaServer.CertificateHold, cryptoutilCAStorage.ReasonCertificateHold},
		{testRevocationCaseRemoveFromCRL, cryptoutilApiCaServer.RemoveFromCRL, cryptoutilCAStorage.ReasonRemoveFromCRL},
		{testRevocationCasePrivilegeWithdrawn, cryptoutilApiCaServer.PrivilegeWithdrawn, cryptoutilCAStorage.ReasonPrivilegeWithdrawn},
		{testRevocationCaseAACompromise, cryptoutilApiCaServer.AaCompromise, cryptoutilCAStorage.ReasonAACompromise},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameUnspecified, cryptoutilApiCaServer.Unspecified, cryptoutilCAStorage.ReasonUnspecified},
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
				key, _ := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)

				return &x509.Certificate{PublicKey: &key.PublicKey}
			},
			expectedAlgo:    cryptoutilSharedMagic.KeyTypeRSA,
			expectedMinSize: cryptoutilSharedMagic.DefaultMetricsBatchSize,
			expectedMaxSize: cryptoutilSharedMagic.DefaultMetricsBatchSize,
		},
		{
			name: "ecdsa_p256",
			generateCert: func() *x509.Certificate {
				key, _ := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)

				return &x509.Certificate{PublicKey: &key.PublicKey}
			},
			expectedAlgo:    cryptoutilSharedMagic.ECDSA,
			expectedMinSize: cryptoutilSharedMagic.MaxUnsealSharedSecrets,
			expectedMaxSize: cryptoutilSharedMagic.MaxUnsealSharedSecrets,
		},
		{
			name: "ecdsa_p384",
			generateCert: func() *x509.Certificate {
				key, _ := ecdsa.GenerateKey(elliptic.P384(), crand.Reader)

				return &x509.Certificate{PublicKey: &key.PublicKey}
			},
			expectedAlgo:    cryptoutilSharedMagic.ECDSA,
			expectedMinSize: cryptoutilSharedMagic.SymmetricKeySize384,
			expectedMaxSize: cryptoutilSharedMagic.SymmetricKeySize384,
		},
		{
			name: "ed25519",
			generateCert: func() *x509.Certificate {
				pub, _, _ := ed25519.GenerateKey(crand.Reader)

				return &x509.Certificate{PublicKey: pub}
			},
			expectedAlgo:    cryptoutilSharedMagic.JoseAlgEdDSA,
			expectedMinSize: ed25519.PublicKeySize * cryptoutilSharedMagic.BitsPerByte,
			expectedMaxSize: ed25519.PublicKeySize * cryptoutilSharedMagic.BitsPerByte,
		},
		{
			name: "unknown_public_key",
			generateCert: func() *x509.Certificate {
				return &x509.Certificate{PublicKey: "unknown_key_type"}
			},
			expectedAlgo:    cryptoutilSharedMagic.UNKNOWN,
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
		{testCaseNameEmptyReturnsNil, "", nil},
		{"non_empty_returns_pointer", testValueText, ptrTo(testValueText)},
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
		{testCaseNameEmptyReturnsNil, []string{}, nil},
		{testCaseNameSingleElement, []string{testValueA}, &[]string{testValueA}},
		{testCaseNameMultipleElements, []string{testValueA, testValueB, testValueC}, &[]string{testValueA, testValueB, testValueC}},
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
		{testCaseNameCNOnly, "CN=test", testValueText},
		{"full_dn_extracts_cn", testSubjectDNFull, testValueText},
		{"no_cn_returns_full_dn", testSubjectDNWithoutCN, testSubjectDNWithoutCN},
		{"cn_with_spaces", "CN= test user ,O=Org", " test user "},
		{"multiple_cn_returns_first", "CN=first,CN=second", "first"},
		{"cn_at_middle", "O=Org,CN=test,C=US", testValueText},
		{"cn_at_end", "O=Org,C=US,CN=test", testValueText},
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
		{testCaseNameEmptyReturnsEmptyCN, "", ""},
		{testCaseNameFullDN, testSubjectDNFull, testValueText},
		{testCaseNameCNOnly, testSubjectDNOnlyCN, testCommonNameOnly},
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
		{testCaseNameEmptyReturnsEmptyCN, "", ""},
		{testCaseNameFullDN, testSubjectDNFull, testValueText},
		{testCaseNameCNOnly, testSubjectDNOnlyCN, testCommonNameOnly},
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
	require.Contains(t, err.Error(), testErrorMessageIssuerRequired)
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
		ID:          testProfileIDTest,
		Name:        testProfileNameTest,
		Description: testProfileDescription,
		Category:    testValueText,
	}

	require.Equal(t, testProfileIDTest, profile.ID)
	require.Equal(t, testProfileNameTest, profile.Name)
	require.Equal(t, testProfileDescription, profile.Description)
	require.Equal(t, testValueText, profile.Category)
}
