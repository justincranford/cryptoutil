// Copyright (c) 2025 Justin Cranford

package timestamp

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	sha256 "crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/asn1"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
)

func TestTSAService_UniqueSerialNumbers(t *testing.T) {
	t.Parallel()

	cert, key := createTSACert(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	config := &TSAConfig{
		Certificate:        cert,
		PrivateKey:         key,
		Provider:           provider,
		Policy:             asn1.ObjectIdentifier{1, 2, 3, 4, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
		AcceptedAlgorithms: []HashAlgorithm{HashAlgorithmSHA256},
	}

	svc, err := NewTSAService(config)
	require.NoError(t, err)

	// Generate multiple timestamps and verify serial numbers are unique.
	hash := sha256.Sum256([]byte("test"))
	serials := make(map[string]bool)

	for i := 0; i < cryptoutilSharedMagic.JoseJAMaxMaterials; i++ {
		req := &TimestampRequest{
			Version: 1,
			MessageImprint: MessageImprint{
				HashAlgorithm: HashAlgorithmSHA256,
				HashedMessage: hash[:],
			},
		}

		resp, err := svc.CreateTimestamp(req)
		require.NoError(t, err)
		require.Equal(t, PKIStatusGranted, resp.Status.Status)

		serial := resp.TimeStampToken.TSTInfo.SerialNumber.String()
		require.False(t, serials[serial], "duplicate serial number: %s", serial)
		serials[serial] = true
	}
}

func TestTSAService_AcceptedPolicies(t *testing.T) {
	t.Parallel()

	cert, key := createTSACert(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()
	defaultPolicy := asn1.ObjectIdentifier{1, 2, 3, 4, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries}
	otherPolicy := asn1.ObjectIdentifier{1, 2, 3, 4, cryptoutilSharedMagic.DefaultEmailOTPLength}

	config := &TSAConfig{
		Certificate:        cert,
		PrivateKey:         key,
		Provider:           provider,
		Policy:             defaultPolicy,
		AcceptedPolicies:   []asn1.ObjectIdentifier{defaultPolicy},
		AcceptedAlgorithms: []HashAlgorithm{HashAlgorithmSHA256},
	}

	svc, err := NewTSAService(config)
	require.NoError(t, err)

	hash := sha256.Sum256([]byte("test"))

	// Request with accepted policy should work.
	reqGood := &TimestampRequest{
		Version: 1,
		MessageImprint: MessageImprint{
			HashAlgorithm: HashAlgorithmSHA256,
			HashedMessage: hash[:],
		},
		ReqPolicy: defaultPolicy,
	}

	resp, err := svc.CreateTimestamp(reqGood)
	require.NoError(t, err)
	require.Equal(t, PKIStatusGranted, resp.Status.Status)

	// Request with unaccepted policy should be rejected.
	reqBad := &TimestampRequest{
		Version: 1,
		MessageImprint: MessageImprint{
			HashAlgorithm: HashAlgorithmSHA256,
			HashedMessage: hash[:],
		},
		ReqPolicy: otherPolicy,
	}

	resp, err = svc.CreateTimestamp(reqBad)
	require.NoError(t, err)
	require.Equal(t, PKIStatusRejection, resp.Status.Status)
}

func TestTSTInfo_ToEntry(t *testing.T) {
	t.Parallel()

	hash := sha256.Sum256([]byte("test"))
	genTime := time.Now().UTC()

	tstInfo := &TSTInfo{
		Version:      1,
		Policy:       asn1.ObjectIdentifier{1, 2, 3, 4, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
		SerialNumber: big.NewInt(12345),
		GenTime:      genTime,
		MessageImprint: MessageImprint{
			HashAlgorithm: HashAlgorithmSHA256,
			HashedMessage: hash[:],
		},
		Nonce: big.NewInt(67890),
	}

	entry := tstInfo.ToEntry("CN=Test TSA")
	require.Equal(t, "3039", entry.SerialNumber) // 12345 in hex.
	require.Equal(t, genTime, entry.GenTime)
	require.Equal(t, "1.2.3.4.5", entry.Policy)
	require.Equal(t, cryptoutilSharedMagic.PBKDF2DefaultAlgorithm, entry.HashAlgorithm)
	require.Equal(t, "CN=Test TSA", entry.TSACertificate)
	require.NotEmpty(t, entry.Nonce)
}

func TestParseTimestampRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty-input",
			input:   []byte{},
			wantErr: true,
			errMsg:  "empty timestamp request",
		},
		{
			name:    "invalid-asn1",
			input:   []byte{0x01, 0x02, 0x03},
			wantErr: true,
			errMsg:  "failed to parse timestamp request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := ParseTimestampRequest(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				require.Nil(t, req)
			} else {
				require.NoError(t, err)
				require.NotNil(t, req)
			}
		})
	}
}

func TestSerializeTimestampResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		resp    *TimestampResponse
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil-response",
			resp:    nil,
			wantErr: true,
			errMsg:  "response is nil",
		},
		{
			name: "rejection-response",
			resp: &TimestampResponse{
				Status: PKIStatusInfo{
					Status: PKIStatusRejection,
					// Note: StatusString omitted as ASN.1 encoder has issues with optional utf8 strings.
				},
			},
			wantErr: false,
		},
		{
			name: "granted-response",
			resp: &TimestampResponse{
				Status: PKIStatusInfo{
					Status: PKIStatusGranted,
				},
				TimeStampToken: &TimeStampToken{
					TSTInfo: TSTInfo{
						Version:      1,
						Policy:       asn1.ObjectIdentifier{1, 2, 3, 4, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
						SerialNumber: big.NewInt(12345),
						GenTime:      time.Now().UTC(),
						MessageImprint: MessageImprint{
							HashAlgorithm: HashAlgorithmSHA256,
							HashedMessage: make([]byte, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes),
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			der, err := SerializeTimestampResponse(tc.resp)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				require.Nil(t, der)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, der)
			}
		})
	}
}

func TestOidToHashAlgorithm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		oid     asn1.ObjectIdentifier
		want    HashAlgorithm
		wantErr bool
	}{
		{
			name: "sha256",
			oid:  asn1.ObjectIdentifier{2, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 840, 1, 101, 3, 4, 2, 1},
			want: HashAlgorithmSHA256,
		},
		{
			name: "sha384",
			oid:  asn1.ObjectIdentifier{2, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 840, 1, 101, 3, 4, 2, 2},
			want: HashAlgorithmSHA384,
		},
		{
			name: "sha512",
			oid:  asn1.ObjectIdentifier{2, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 840, 1, 101, 3, 4, 2, 3},
			want: HashAlgorithmSHA512,
		},
		{
			name:    "unknown-oid",
			oid:     asn1.ObjectIdentifier{1, 2, 3, 4, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			alg, err := oidToHashAlgorithm(tc.oid)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, alg)
			}
		})
	}
}

// Helper functions.

func createTSACert(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test TSA",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert, privateKey
}
