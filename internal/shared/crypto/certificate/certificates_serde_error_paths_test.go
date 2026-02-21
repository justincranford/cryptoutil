// Copyright (c) 2025 Justin Cranford
//
//

package certificate

import (
	"crypto/x509"
	json "encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// serdeTestFakePrivKey is a private key type unknown to x509.MarshalPKCS8PrivateKey.
// Used to trigger the MarshalPKCS8PrivateKey error path in serializeKeyMaterial.
type serdeTestFakePrivKey struct{}

// serdeTestFakePublicKey is a public key type unknown to x509.MarshalPKIXPublicKey.
// Used to trigger the MarshalPKIXPublicKey error path in serializeKeyMaterial.
type serdeTestFakePublicKey struct{}

// TestSerializeSubjects_AdditionalErrors covers missing error paths in SerializeSubjects.
func TestSerializeSubjects_AdditionalErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		subjects       []*Subject
		wantErrContain string
	}{
		{
			name: "empty IssuerName",
			subjects: []*Subject{{
				SubjectName: "Test CA",
				IssuerName:  "", // Empty IssuerName should fail.
				Duration:    testHourDuration,
			}},
			wantErrContain: "has empty IssuerName",
		},
		{
			name: "zero Duration",
			subjects: []*Subject{{
				SubjectName: "Test CA",
				IssuerName:  "Root CA",
				Duration:    0, // Zero Duration should fail.
			}},
			wantErrContain: "has zero or negative Duration",
		},
		{
			name: "serializeKeyMaterial fails with empty chain",
			subjects: []*Subject{{
				SubjectName: "Test Subject",
				IssuerName:  "Root CA",
				Duration:    testHourDuration,
				IsCA:        false,
				MaxPathLen:  0,
				KeyMaterial: KeyMaterial{
					CertificateChain: nil, // Empty chain triggers serializeKeyMaterial error.
					PublicKey:        nil,
				},
			}},
			wantErrContain: "failed to convert KeyMaterial to JSON format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := SerializeSubjects(tt.subjects, false)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErrContain)
		})
	}
}

// TestDeserializeSubjects_InvalidJSON tests that DeserializeSubjects errors
// when the input bytes are not valid JSON.
func TestDeserializeSubjects_InvalidJSON(t *testing.T) {
	t.Parallel()

	invalidJSON := [][]byte{{0x01, 0x02, 0x03}} // Not valid JSON.

	_, err := DeserializeSubjects(invalidJSON)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to deserialize KeyMaterialEncoded")
}

// TestDeserializeSubjects_DeserializeKeyMaterialError tests that DeserializeSubjects
// errors when the deserialized KeyMaterialEncoded fails validation.
func TestDeserializeSubjects_DeserializeKeyMaterialError(t *testing.T) {
	t.Parallel()

	// Marshal a KeyMaterialEncoded with an empty DER certificate chain.
	kme := KeyMaterialEncoded{
		DERCertificateChain: [][]byte{}, // Empty triggers deserializeKeyMaterial failure.
		DERPublicKey:        []byte("fake"),
	}

	jsonBytes, err := json.Marshal(kme)
	require.NoError(t, err)

	_, err = DeserializeSubjects([][]byte{jsonBytes})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to convert KeyMaterialEncoded to KeyMaterial")
}

// TestDeserializeKeyMaterial_AllValidationErrors covers all validation error paths
// in the private deserializeKeyMaterial function.
func TestDeserializeKeyMaterial_AllValidationErrors(t *testing.T) {
	t.Parallel()

	// Build a real valid cert to get real DER bytes.
	keyPair := testKeyGenPool.GetMany(1)[0]

	certTemplate, err := CertificateTemplateCA("Issuer", "Test CA", testCACertValidity10Years, 0)
	require.NoError(t, err)

	cert, _, _, err := SignCertificate(nil, keyPair.Private, certTemplate, keyPair.Public, x509.ECDSAWithSHA256)
	require.NoError(t, err)

	validDERCert := cert.Raw

	validDERPublicKey, err := x509.MarshalPKIXPublicKey(keyPair.Public)
	require.NoError(t, err)

	tests := []struct {
		name           string
		input          *KeyMaterialEncoded
		wantErrContain string
	}{
		{
			name:           "nil input",
			input:          nil,
			wantErrContain: "keyMaterialEncoded cannot be nil",
		},
		{
			name: "empty DER certificate chain",
			input: &KeyMaterialEncoded{
				DERCertificateChain: [][]byte{},
				DERPublicKey:        validDERPublicKey,
			},
			wantErrContain: "DER certificate chain cannot be empty",
		},
		{
			name: "empty DER public key",
			input: &KeyMaterialEncoded{
				DERCertificateChain: [][]byte{validDERCert},
				DERPublicKey:        []byte{},
			},
			wantErrContain: "DER public key cannot be empty",
		},
		{
			name: "empty DER bytes in chain",
			input: &KeyMaterialEncoded{
				DERCertificateChain: [][]byte{{}}, // Contains an empty entry.
				DERPublicKey:        validDERPublicKey,
			},
			wantErrContain: "DER certificate at index 0 in chain cannot be empty",
		},
		{
			name: "invalid DER certificate bytes",
			input: &KeyMaterialEncoded{
				DERCertificateChain: [][]byte{[]byte("invalid-der")},
				DERPublicKey:        validDERPublicKey,
			},
			wantErrContain: "failed to parse certificate 0 from DER",
		},
		{
			name: "invalid DER public key bytes",
			input: &KeyMaterialEncoded{
				DERCertificateChain: [][]byte{validDERCert},
				DERPublicKey:        []byte("invalid-public-key-der"),
			},
			wantErrContain: "failed to parse public key from DER",
		},
		{
			name: "invalid DER private key bytes",
			input: &KeyMaterialEncoded{
				DERCertificateChain: [][]byte{validDERCert},
				DERPublicKey:        validDERPublicKey,
				DERPrivateKey:       []byte("invalid-private-key-der"),
			},
			wantErrContain: "failed to parse private key from DER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := deserializeKeyMaterial(tt.input)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErrContain)
		})
	}
}

// TestGenerateNotBeforeNotAfter_InvalidParams tests edge cases in the private
// generateNotBeforeNotAfter function.
func TestGenerateNotBeforeNotAfter_InvalidParams(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name           string
		requestedStart time.Time
		requestedDur   time.Duration
		minSubtract    time.Duration
		maxSubtract    time.Duration
		wantErrContain string
	}{
		{
			name:           "minSubtract is zero",
			requestedStart: now,
			requestedDur:   24 * time.Hour,
			minSubtract:    0,
			maxSubtract:    time.Hour,
			wantErrContain: "minSubtract must be positive",
		},
		{
			name:           "maxSubtract is zero",
			requestedStart: now,
			requestedDur:   24 * time.Hour,
			minSubtract:    time.Minute,
			maxSubtract:    0,
			wantErrContain: "maxSubtract must be positive",
		},
		{
			name:           "maxRangeOffset is zero (maxSubtract equals minSubtract)",
			requestedStart: now,
			requestedDur:   24 * time.Hour,
			minSubtract:    time.Hour,
			maxSubtract:    time.Hour, // maxRangeOffset = maxSubtract - minSubtract = 0.
			wantErrContain: "maxRangeOffset must be positive",
		},
		{
			name:           "maxRangeOffset is negative (maxSubtract less than minSubtract)",
			requestedStart: now,
			requestedDur:   24 * time.Hour,
			minSubtract:    2 * time.Hour,
			maxSubtract:    time.Hour, // maxRangeOffset = 1h - 2h = -1h.
			wantErrContain: "maxRangeOffset must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := generateNotBeforeNotAfter(tt.requestedStart, tt.requestedDur, tt.minSubtract, tt.maxSubtract)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErrContain)
		})
	}
}

// TestRandomizedNotBeforeNotAfterCA_ExcessiveDuration tests that
// randomizedNotBeforeNotAfterCA returns an error for duration exceeding 25 years.
func TestRandomizedNotBeforeNotAfterCA_ExcessiveDuration(t *testing.T) {
	t.Parallel()

	// 26 years exceeds TLSDefaultMaxCACertDuration (25 years).
	excessiveDuration := time.Duration(26 * 365 * 24 * time.Hour)

	_, _, err := randomizedNotBeforeNotAfterCA(
		time.Now().UTC(),
		excessiveDuration,
		time.Minute,
		120*time.Minute,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requestedDuration exceeds maxCACertDuration")
}

// TestRandomizedNotBeforeNotAfterEndEntity_ExcessiveDuration tests that
// randomizedNotBeforeNotAfterEndEntity returns an error for duration exceeding 397 days.
func TestRandomizedNotBeforeNotAfterEndEntity_ExcessiveDuration(t *testing.T) {
	t.Parallel()

	// 398 days exceeds TLSDefaultSubscriberCertDuration (397 days).
	excessiveDuration := 398 * 24 * time.Hour

	_, _, err := randomizedNotBeforeNotAfterEndEntity(
		time.Now().UTC(),
		excessiveDuration,
		time.Minute,
		120*time.Minute,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requestedDuration exceeds maxSubscriberCertDuration")
}

// TestRandomizedNotBeforeNotAfterEndEntity_ZeroDuration tests that
// randomizedNotBeforeNotAfterEndEntity propagates an error from generateNotBeforeNotAfter
// when requestedDuration is zero (passes the maxSubscriberCert check but fails validation).
func TestRandomizedNotBeforeNotAfterEndEntity_ZeroDuration(t *testing.T) {
	t.Parallel()

	_, _, err := randomizedNotBeforeNotAfterEndEntity(
		time.Now().UTC(),
		0, // Zero duration passes the > maxSubscriberCert check but fails in generateNotBeforeNotAfter.
		time.Minute,
		2*time.Minute,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate notBefore/notAfter")
}

// TestSerializeKeyMaterial_FakeKeys covers the MarshalPKCS8PrivateKey and
// MarshalPKIXPublicKey error paths in serializeKeyMaterial (lines 405 and 413).
func TestSerializeKeyMaterial_FakeKeys(t *testing.T) {
	t.Parallel()

	// Build a real cert chain required by serializeKeyMaterial validation.
	keyPair := testKeyGenPool.GetMany(1)[0]

	certTemplate, err := CertificateTemplateCA("Test Issuer", "Test CA", testCACertValidity10Years, 0)
	require.NoError(t, err)

	cert, _, _, err := SignCertificate(nil, keyPair.Private, certTemplate, keyPair.Public, x509.ECDSAWithSHA256)
	require.NoError(t, err)

	validChain := []*x509.Certificate{cert}

	t.Run("fake private key triggers MarshalPKCS8PrivateKey error", func(t *testing.T) {
		t.Parallel()

		// serdeTestFakePrivKey is unknown to x509.MarshalPKCS8PrivateKey.
		material := &KeyMaterial{
			CertificateChain: validChain,
			PublicKey:        keyPair.Public,         // Valid public key passes nil check.
			PrivateKey:       serdeTestFakePrivKey{}, // Unknown type → MarshalPKCS8PrivateKey fails.
		}

		_, err := serializeKeyMaterial(material, true) // includePrivateKey=true enters the branch.
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to marshal private key to DER")
	})

	t.Run("fake public key triggers MarshalPKIXPublicKey error", func(t *testing.T) {
		t.Parallel()

		// serdeTestFakePublicKey is unknown to x509.MarshalPKIXPublicKey.
		material := &KeyMaterial{
			CertificateChain: validChain,
			PublicKey:        serdeTestFakePublicKey{}, // Unknown type → MarshalPKIXPublicKey fails.
			PrivateKey:       nil,
		}

		_, err := serializeKeyMaterial(material, false) // includePrivateKey=false skips MarshalPKCS8.
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to marshal public key to DER")
	})
}
