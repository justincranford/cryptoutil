// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"crypto"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestValidateStoredCertificateIssuer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		legacy        bool
		nilRecord     bool
		mutateRecord  func(*cryptoutilCAStorage.StoredCertificate)
		mutateIssuer  func(*testIssuerSetup)
		wantErrorText string
	}{
		{
			name: "current record matches issuer",
		},
		{
			name:   "legacy record verified against issuer",
			legacy: true,
		},
		{
			name: "authority ID mismatch",
			mutateRecord: func(record *cryptoutilCAStorage.StoredCertificate) {
				record.IssuerAuthorityID = testDifferentAuthorityID
			},
			wantErrorText: "authority ID does not match",
		},
		{
			name: "issuer DN mismatch",
			mutateRecord: func(record *cryptoutilCAStorage.StoredCertificate) {
				record.IssuerDN = "CN=untrusted issuer"
			},
			wantErrorText: "issuer DN does not match",
		},
		{
			name: "authority key ID mismatch",
			mutateRecord: func(record *cryptoutilCAStorage.StoredCertificate) {
				record.AuthorityKeyID = "deadbeef"
			},
			wantErrorText: testAuthorityKeyIDMismatch,
		},
		{
			name: "configured authority identity missing",
			mutateIssuer: func(setup *testIssuerSetup) {
				setup.Issuer.GetCAConfig().AuthorityID = testWhitespace
			},
			wantErrorText: "authority ID does not match configured issuer",
		},
		{
			name: "configured authority certificate missing",
			mutateIssuer: func(setup *testIssuerSetup) {
				setup.Issuer.GetCAConfig().Certificate = nil
			},
			wantErrorText: testCertificateIdentityNotConfigured,
		},
		{
			name: "configured authority subject key ID missing",
			mutateIssuer: func(setup *testIssuerSetup) {
				setup.Issuer.GetCAConfig().Certificate.SubjectKeyId = nil
			},
			wantErrorText: testCertificateIdentityNotConfigured,
		},
		{
			name:          "nil stored certificate",
			nilRecord:     true,
			wantErrorText: "stored certificate is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			setup := createTestIssuer(t)
			certificatePEM := createOCSPTestCertificatePEM(
				t,
				setup.Issuer.GetCAConfig().Certificate,
				testIssuerSigner(t, setup),
				big.NewInt(42),
			)

			record := storedCertificateForIssuer(setup, certificatePEM)
			if tc.legacy {
				record.IssuerAuthorityID = ""
			}

			if tc.mutateRecord != nil {
				tc.mutateRecord(record)
			}

			if tc.mutateIssuer != nil {
				tc.mutateIssuer(setup)
			}

			if tc.nilRecord {
				record = nil
			}

			err := validateStoredCertificateIssuer(record, setup.Issuer)
			if tc.wantErrorText != "" {
				require.ErrorContains(t, err, tc.wantErrorText)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidateLegacyStoredCertificateIssuer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		certificatePEM func(*testing.T, *testIssuerSetup) string
		serialNumber   string
		wantErrorText  string
	}{
		{
			name: "valid certificate",
		},
		{
			name:           "missing PEM block",
			certificatePEM: func(*testing.T, *testIssuerSetup) string { return "not PEM" },
			wantErrorText:  testStoredCertificatePEMInvalid,
		},
		{
			name: "wrong PEM block type",
			certificatePEM: func(t *testing.T, setup *testIssuerSetup) string {
				t.Helper()

				return string(pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCSR, Bytes: []byte("unused")}))
			},
			wantErrorText: testStoredCertificatePEMInvalid,
		},
		{
			name: "invalid certificate DER",
			certificatePEM: func(t *testing.T, _ *testIssuerSetup) string {
				t.Helper()

				return string(pem.EncodeToMemory(&pem.Block{
					Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
					Bytes: []byte("invalid DER"),
				}))
			},
			wantErrorText: testFailedToParseStoredCertificate,
		},
		{
			name:          "serial mismatch",
			serialNumber:  "2b",
			wantErrorText: testSerialNumberMismatch,
		},
		{
			name: "issuer mismatch",
			certificatePEM: func(t *testing.T, setup *testIssuerSetup) string {
				t.Helper()
				other := createTestIssuer(t)

				return createOCSPTestCertificatePEM(
					t,
					other.Issuer.GetCAConfig().Certificate,
					testIssuerSigner(t, other),
					big.NewInt(42),
				)
			},
			wantErrorText: "authority key ID does not match",
		},
		{
			name: "signature mismatch",
			certificatePEM: func(t *testing.T, setup *testIssuerSetup) string {
				t.Helper()
				certificatePEM := createOCSPTestCertificatePEM(
					t,
					setup.Issuer.GetCAConfig().Certificate,
					testIssuerSigner(t, setup),
					big.NewInt(42),
				)
				block, _ := pem.Decode([]byte(certificatePEM))
				require.NotNil(t, block)
				block.Bytes[len(block.Bytes)-1] ^= 0xff

				return string(pem.EncodeToMemory(block))
			},
			wantErrorText: testSignatureMismatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			setup := createTestIssuer(t)

			certificatePEM := ""
			if tc.certificatePEM != nil {
				certificatePEM = tc.certificatePEM(t, setup)
			} else {
				certificatePEM = createOCSPTestCertificatePEM(
					t,
					setup.Issuer.GetCAConfig().Certificate,
					testIssuerSigner(t, setup),
					big.NewInt(42),
				)
			}

			serialNumber := tc.serialNumber
			if serialNumber == "" {
				serialNumber = testSerialNumberHex
			}

			err := validateLegacyStoredCertificateIssuer(
				&cryptoutilCAStorage.StoredCertificate{
					SerialNumber:   serialNumber,
					CertificatePEM: certificatePEM,
				},
				setup.Issuer.GetCAConfig().Certificate,
			)
			if tc.wantErrorText != "" {
				require.ErrorContains(t, err, tc.wantErrorText)

				return
			}

			require.NoError(t, err)
		})
	}
}

func testIssuerSigner(t *testing.T, setup *testIssuerSetup) crypto.Signer {
	t.Helper()

	signer, ok := setup.Issuer.GetCAConfig().PrivateKey.(crypto.Signer)
	require.True(t, ok)

	return signer
}

func storedCertificateForIssuer(setup *testIssuerSetup, certificatePEM string) *cryptoutilCAStorage.StoredCertificate {
	caConfig := setup.Issuer.GetCAConfig()

	return &cryptoutilCAStorage.StoredCertificate{
		SerialNumber:      testSerialNumberHex,
		IssuerDN:          caConfig.Certificate.Subject.String(),
		IssuerAuthorityID: caConfig.AuthorityID,
		CertificatePEM:    certificatePEM,
		AuthorityKeyID:    fmtIssuerKeyID(caConfig.Certificate),
		NotBefore:         time.Now().UTC(),
		NotAfter:          time.Now().UTC().Add(time.Hour),
	}
}
