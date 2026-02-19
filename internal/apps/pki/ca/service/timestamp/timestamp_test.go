// Copyright (c) 2025 Justin Cranford

package timestamp

import (
	"crypto"
	sha256 "crypto/sha256"
	"encoding/asn1"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
)

func TestPKIStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status PKIStatus
		want   string
	}{
		{"granted", PKIStatusGranted, "granted"},
		{"grantedWithMods", PKIStatusGrantedWithMods, "grantedWithMods"},
		{"rejection", PKIStatusRejection, "rejection"},
		{"waiting", PKIStatusWaiting, "waiting"},
		{"revocationWarning", PKIStatusRevocationWarning, "revocationWarning"},
		{"revocationNotification", PKIStatusRevocationNotification, "revocationNotification"},
		{"unknown", PKIStatus(100), "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.status.String()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestPKIFailureInfo_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		failInfo PKIFailureInfo
		want     string
	}{
		{"badAlg", PKIFailureBadAlg, "badAlg"},
		{"badRequest", PKIFailureBadRequest, "badRequest"},
		{"badDataFormat", PKIFailureBadDataFormat, "badDataFormat"},
		{"timeNotAvailable", PKIFailureTimeNotAvailable, "timeNotAvailable"},
		{"unacceptedPolicy", PKIFailureUnacceptedPolicy, "unacceptedPolicy"},
		{"unacceptedExtension", PKIFailureUnacceptedExtension, "unacceptedExtension"},
		{"addInfoNotAvailable", PKIFailureAddInfoNotAvailable, "addInfoNotAvailable"},
		{"systemFailure", PKIFailureSystemFailure, "systemFailure"},
		{"unknown", PKIFailureInfo(100), "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.failInfo.String()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestHashAlgorithm_OID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		alg  HashAlgorithm
		want asn1.ObjectIdentifier
	}{
		{"SHA-256", HashAlgorithmSHA256, asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}},
		{"SHA-384", HashAlgorithmSHA384, asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 2}},
		{"SHA-512", HashAlgorithmSHA512, asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 3}},
		{"unknown", HashAlgorithm("unknown"), nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.alg.OID()
			require.Equal(t, tc.want, got)
		})
	}
}

func TestHashAlgorithm_CryptoHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		alg    HashAlgorithm
		want   int
		isZero bool
	}{
		{"SHA-256", HashAlgorithmSHA256, 32, false},
		{"SHA-384", HashAlgorithmSHA384, 48, false},
		{"SHA-512", HashAlgorithmSHA512, 64, false},
		{"unknown", HashAlgorithm("unknown"), 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hash := tc.alg.CryptoHash()
			if tc.isZero {
				require.Equal(t, crypto.Hash(0), hash)
			} else {
				require.Equal(t, tc.want, hash.Size())
			}
		})
	}
}

func TestNewTSAService(t *testing.T) {
	t.Parallel()

	cert, key := createTSACert(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()
	policy := asn1.ObjectIdentifier{1, 2, 3, 4, 5}

	tests := []struct {
		name    string
		config  *TSAConfig
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
			name: "nil-certificate",
			config: &TSAConfig{
				Certificate: nil,
				PrivateKey:  key,
				Provider:    provider,
				Policy:      policy,
			},
			wantErr: true,
			errMsg:  "certificate is required",
		},
		{
			name: "nil-private-key",
			config: &TSAConfig{
				Certificate: cert,
				PrivateKey:  nil,
				Provider:    provider,
				Policy:      policy,
			},
			wantErr: true,
			errMsg:  "private key is required",
		},
		{
			name: "nil-provider",
			config: &TSAConfig{
				Certificate: cert,
				PrivateKey:  key,
				Provider:    nil,
				Policy:      policy,
			},
			wantErr: true,
			errMsg:  "crypto provider is required",
		},
		{
			name: "nil-policy",
			config: &TSAConfig{
				Certificate: cert,
				PrivateKey:  key,
				Provider:    provider,
				Policy:      nil,
			},
			wantErr: true,
			errMsg:  "policy OID is required",
		},
		{
			name: "valid-config",
			config: &TSAConfig{
				Certificate: cert,
				PrivateKey:  key,
				Provider:    provider,
				Policy:      policy,
			},
			wantErr: false,
		},
		{
			name: "valid-config-with-algorithms",
			config: &TSAConfig{
				Certificate:        cert,
				PrivateKey:         key,
				Provider:           provider,
				Policy:             policy,
				AcceptedAlgorithms: []HashAlgorithm{HashAlgorithmSHA256, HashAlgorithmSHA384},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, err := NewTSAService(tc.config)

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

func TestTSAService_CreateTimestamp(t *testing.T) {
	t.Parallel()

	cert, key := createTSACert(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()
	policy := asn1.ObjectIdentifier{1, 2, 3, 4, 5}

	config := &TSAConfig{
		Certificate:        cert,
		PrivateKey:         key,
		Provider:           provider,
		Policy:             policy,
		AcceptedAlgorithms: []HashAlgorithm{HashAlgorithmSHA256, HashAlgorithmSHA384},
		Accuracy: &Accuracy{
			Seconds: 1,
			Millis:  0,
			Micros:  0,
		},
	}

	svc, err := NewTSAService(config)
	require.NoError(t, err)

	// Create a hash of some data.
	data := []byte("test data to timestamp")
	hash := sha256.Sum256(data)

	tests := []struct {
		name       string
		request    *TimestampRequest
		wantStatus PKIStatus
	}{
		{
			name:       "nil-request",
			request:    nil,
			wantStatus: PKIStatusRejection,
		},
		{
			name: "empty-hash",
			request: &TimestampRequest{
				Version: 1,
				MessageImprint: MessageImprint{
					HashAlgorithm: HashAlgorithmSHA256,
					HashedMessage: nil,
				},
			},
			wantStatus: PKIStatusRejection,
		},
		{
			name: "unsupported-algorithm",
			request: &TimestampRequest{
				Version: 1,
				MessageImprint: MessageImprint{
					HashAlgorithm: HashAlgorithm("MD5"),
					HashedMessage: hash[:],
				},
			},
			wantStatus: PKIStatusRejection,
		},
		{
			name: "wrong-hash-length",
			request: &TimestampRequest{
				Version: 1,
				MessageImprint: MessageImprint{
					HashAlgorithm: HashAlgorithmSHA256,
					HashedMessage: []byte("short"),
				},
			},
			wantStatus: PKIStatusRejection,
		},
		{
			name: "valid-request",
			request: &TimestampRequest{
				Version: 1,
				MessageImprint: MessageImprint{
					HashAlgorithm: HashAlgorithmSHA256,
					HashedMessage: hash[:],
				},
				CertReq: true,
			},
			wantStatus: PKIStatusGranted,
		},
		{
			name: "valid-request-with-nonce",
			request: &TimestampRequest{
				Version: 1,
				MessageImprint: MessageImprint{
					HashAlgorithm: HashAlgorithmSHA256,
					HashedMessage: hash[:],
				},
				Nonce:   big.NewInt(12345),
				CertReq: true,
			},
			wantStatus: PKIStatusGranted,
		},
		{
			name: "valid-request-with-policy",
			request: &TimestampRequest{
				Version: 1,
				MessageImprint: MessageImprint{
					HashAlgorithm: HashAlgorithmSHA256,
					HashedMessage: hash[:],
				},
				ReqPolicy: policy,
				CertReq:   true,
			},
			wantStatus: PKIStatusGranted,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp, err := svc.CreateTimestamp(tc.request)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, tc.wantStatus, resp.Status.Status)

			if tc.wantStatus == PKIStatusGranted {
				require.NotNil(t, resp.TimeStampToken)
				require.NotNil(t, resp.TimeStampToken.TSTInfo.SerialNumber)
				require.False(t, resp.TimeStampToken.TSTInfo.GenTime.IsZero())
			}
		})
	}
}
