// Copyright (c) 2025 Justin Cranford

package timestamp

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/asn1"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTimestampRequest_TrailingData(t *testing.T) {
	t.Parallel()

	// Build a valid timeStampReqASN1 with SHA-256 OID.
	req := timeStampReqASN1{
		Version: 1,
		MessageImprint: messageImprintASN1{
			HashAlgorithm: algorithmIdentifierASN1{
				Algorithm: asn1.ObjectIdentifier{2, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 840, 1, 101, 3, 4, 2, 1},
			},
			HashedMessage: make([]byte, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes),
		},
	}
	der, err := asn1.Marshal(req)
	require.NoError(t, err)

	// Append trailing byte.
	derWithTrailing := append(der, 0x00) //nolint:gocritic // false positive
	_, err = ParseTimestampRequest(derWithTrailing)
	require.Error(t, err)
	require.Contains(t, err.Error(), "trailing data")
}

func TestParseTimestampRequest_UnsupportedHashOID(t *testing.T) {
	t.Parallel()

	// Use an unsupported hash algorithm OID (MD5).
	req := timeStampReqASN1{
		Version: 1,
		MessageImprint: messageImprintASN1{
			HashAlgorithm: algorithmIdentifierASN1{
				Algorithm: asn1.ObjectIdentifier{1, 2, 840, 113549, 2, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries}, // MD5
			},
			HashedMessage: make([]byte, cryptoutilSharedMagic.RealmMinTokenLengthBytes),
		},
	}
	der, err := asn1.Marshal(req)
	require.NoError(t, err)

	_, err = ParseTimestampRequest(der)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported hash algorithm OID")
}

func TestParseTimestampRequest_SuccessWithExtensions(t *testing.T) {
	t.Parallel()

	// Build a valid request with nonce, policy, certReq, and extensions.
	req := timeStampReqASN1{
		Version: 1,
		MessageImprint: messageImprintASN1{
			HashAlgorithm: algorithmIdentifierASN1{
				Algorithm: asn1.ObjectIdentifier{2, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 840, 1, 101, 3, 4, 2, 1},
			},
			HashedMessage: make([]byte, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes),
		},
		ReqPolicy: asn1.ObjectIdentifier{1, 3, cryptoutilSharedMagic.DefaultEmailOTPLength, 1, 4, 1, 99999, 1},
		Nonce:     big.NewInt(cryptoutilSharedMagic.AnswerToLifeUniverseEverything),
		CertReq:   true,
		Extensions: []extensionASN1{
			{
				OID:      asn1.ObjectIdentifier{2, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, 29, 14},
				Critical: false,
				Value:    []byte("test-extension-value"),
			},
		},
	}
	der, err := asn1.Marshal(req)
	require.NoError(t, err)

	result, err := ParseTimestampRequest(der)
	require.NoError(t, err)
	require.Equal(t, 1, result.Version)
	require.Equal(t, HashAlgorithmSHA256, result.MessageImprint.HashAlgorithm)
	require.NotNil(t, result.Nonce)
	require.True(t, result.CertReq)
	require.Len(t, result.Extensions, 1)
}

func TestSerializeTimestampResponse_WithFailInfo(t *testing.T) {
	t.Parallel()

	failInfo := PKIFailureBadRequest
	resp := &TimestampResponse{
		Status: PKIStatusInfo{
			Status:   PKIStatusRejection,
			FailInfo: &failInfo,
		},
	}
	der, err := SerializeTimestampResponse(resp)
	require.NoError(t, err)
	require.NotEmpty(t, der)
}

func TestSerializeTimestampResponse_WithAccuracy(t *testing.T) {
	t.Parallel()

	resp := &TimestampResponse{
		Status: PKIStatusInfo{
			Status: PKIStatusGranted,
		},
		TimeStampToken: &TimeStampToken{
			TSTInfo: TSTInfo{
				Version: 1,
				Policy:  asn1.ObjectIdentifier{1, 3, cryptoutilSharedMagic.DefaultEmailOTPLength, 1, 4, 1, 99999, 1},
				MessageImprint: MessageImprint{
					HashAlgorithm: HashAlgorithmSHA256,
					HashedMessage: make([]byte, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes),
				},
				SerialNumber: big.NewInt(1),
				Accuracy: &Accuracy{
					Seconds: 1,
					Millis:  cryptoutilSharedMagic.TestDefaultRateLimitServiceIP,
					Micros:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				},
			},
		},
	}
	der, err := SerializeTimestampResponse(resp)
	require.NoError(t, err)
	require.NotEmpty(t, der)
}
