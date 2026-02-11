// Copyright (c) 2025 Justin Cranford

package timestamp

import (
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	sha256 "crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"testing"
	"time"

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

func TestTSAService_UniqueSerialNumbers(t *testing.T) {
	t.Parallel()

	cert, key := createTSACert(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	config := &TSAConfig{
		Certificate:        cert,
		PrivateKey:         key,
		Provider:           provider,
		Policy:             asn1.ObjectIdentifier{1, 2, 3, 4, 5},
		AcceptedAlgorithms: []HashAlgorithm{HashAlgorithmSHA256},
	}

	svc, err := NewTSAService(config)
	require.NoError(t, err)

	// Generate multiple timestamps and verify serial numbers are unique.
	hash := sha256.Sum256([]byte("test"))
	serials := make(map[string]bool)

	for i := 0; i < 100; i++ {
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
	defaultPolicy := asn1.ObjectIdentifier{1, 2, 3, 4, 5}
	otherPolicy := asn1.ObjectIdentifier{1, 2, 3, 4, 6}

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
		Policy:       asn1.ObjectIdentifier{1, 2, 3, 4, 5},
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
	require.Equal(t, "SHA-256", entry.HashAlgorithm)
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
						Policy:       asn1.ObjectIdentifier{1, 2, 3, 4, 5},
						SerialNumber: big.NewInt(12345),
						GenTime:      time.Now().UTC(),
						MessageImprint: MessageImprint{
							HashAlgorithm: HashAlgorithmSHA256,
							HashedMessage: make([]byte, 32),
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
			oid:  asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1},
			want: HashAlgorithmSHA256,
		},
		{
			name: "sha384",
			oid:  asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 2},
			want: HashAlgorithmSHA384,
		},
		{
			name: "sha512",
			oid:  asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 3},
			want: HashAlgorithmSHA512,
		},
		{
			name:    "unknown-oid",
			oid:     asn1.ObjectIdentifier{1, 2, 3, 4, 5},
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
		NotAfter:              time.Now().UTC().Add(365 * 24 * time.Hour),
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
