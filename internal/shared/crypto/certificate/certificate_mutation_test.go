// Copyright (c) 2025 Justin Cranford
//
//

package certificate

import (
	"crypto/x509"
	"io"
	"math/big"
	"net"
	"testing"
	"time"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestRandomizedNotBeforeNotAfterCA_ExactMaxCACertDurationSucceeds kills the CONDITIONALS_BOUNDARY
// mutant at serial_number.go:35 which changes `>` to `>=`. Uses the exact TLSMaxCACertDuration
// boundary value (25 years), which should succeed with `>` but fail with `>=`.
func TestRandomizedNotBeforeNotAfterCA_ExactMaxCACertDurationSucceeds(t *testing.T) {
	t.Parallel()

	notBefore, notAfter, err := randomizedNotBeforeNotAfterCA(
		time.Now().UTC(),
		cryptoutilSharedMagic.TLSMaxCACertDuration, // exactly 25 years (max allowed).
		time.Minute,
		cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes*time.Minute,
	)
	require.NoError(t, err)
	require.True(t, notBefore.Before(notAfter))
}

// TestRandomizedNotBeforeNotAfterCA_BoundaryDurationSucceeds kills the CONDITIONALS_BOUNDARY
// mutant at serial_number.go:42 which changes `>` to `>=` in the post-check. Injects randIntFn
// to force the random offset to exactly maxRangeOffset, creating a boundary value where
// actualDuration == requestedDuration+maxSubtract (passes with `>`, fails with `>=`).
func TestRandomizedNotBeforeNotAfterCA_BoundaryDurationSucceeds(t *testing.T) {
	t.Parallel()

	minSub := 5 * time.Minute
	maxSub := 10 * time.Minute
	maxRangeOffset := int64(maxSub - minSub)

	original := randIntFn

	// The only randIntFn call inside generateNotBeforeNotAfter is the random offset.
	// Force it to exactly maxRangeOffset to hit the boundary.
	randIntFn = func(_ io.Reader, _ *big.Int) (*big.Int, error) {
		return big.NewInt(maxRangeOffset), nil // Force exact boundary offset.
	}

	defer func() { randIntFn = original }()

	_, _, err := randomizedNotBeforeNotAfterCA(time.Now().UTC(), 24*time.Hour, minSub, maxSub)
	require.NoError(t, err) // Original `>`: no error. Mutant `>=`: would error.
}

// TestRandomizedNotBeforeNotAfterEndEntity_BoundaryDurationSucceeds kills the CONDITIONALS_BOUNDARY
// mutant at serial_number.go:57 which changes `>` to `>=` in the post-check. Same approach as
// the CA variant above.
func TestRandomizedNotBeforeNotAfterEndEntity_BoundaryDurationSucceeds(t *testing.T) {
	t.Parallel()

	minSub := 5 * time.Minute
	maxSub := 10 * time.Minute
	maxRangeOffset := int64(maxSub - minSub)

	original := randIntFn

	// Force exact boundary offset (same as CA variant).
	randIntFn = func(_ io.Reader, _ *big.Int) (*big.Int, error) {
		return big.NewInt(maxRangeOffset), nil
	}

	defer func() { randIntFn = original }()

	_, _, err := randomizedNotBeforeNotAfterEndEntity(time.Now().UTC(), 24*time.Hour, minSub, maxSub)
	require.NoError(t, err) // Original `>`: no error. Mutant `>=`: would error.
}

// TestCreateCASubjects_SubjectNaming kills ARITHMETIC_BASE mutants at certificates.go:70:73 and :75.
// These mutants change `len(keyPairs)-1-i` to `len(keyPairs)+1-i` or `len(keyPairs)-1+i`.
// Verifies exact subject naming pattern: subjects[0] gets highest number, subjects[last] gets 0.
func TestCreateCASubjects_SubjectNaming(t *testing.T) {
	t.Parallel()

	keyPairs := testKeyGenPool.GetMany(3)

	subjects, err := CreateCASubjects(keyPairs, "Test CA", testCACertValidity10Years)
	require.NoError(t, err)
	require.Len(t, subjects, 3)

	// Loop iterates i=2,1,0. SubjectName = fmt.Sprintf("%s %d", prefix, len(keyPairs)-1-i).
	// i=2: "Test CA 0" (root, subjects[2])
	// i=1: "Test CA 1" (intermediate, subjects[1])
	// i=0: "Test CA 2" (issuing, subjects[0]).
	require.Equal(t, "Test CA 2", subjects[0].SubjectName)
	require.Equal(t, "Test CA 1", subjects[1].SubjectName)
	require.Equal(t, "Test CA 0", subjects[2].SubjectName)
}

// TestCreateCASubjects_ErrorMessageNumber kills ARITHMETIC_BASE mutants at certificates.go:81:78
// and :80. These mutants change the number in the error message. Verifies the exact subject
// number appears in the error message when creation fails.
func TestCreateCASubjects_ErrorMessageNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		keyPairs       []*cryptoutilSharedCryptoKeygen.KeyPair
		wantErrContain string
	}{
		{
			name: "single key pair with nil public key",
			keyPairs: func() []*cryptoutilSharedCryptoKeygen.KeyPair {
				kp := testKeyGenPool.GetMany(1)[0]

				return []*cryptoutilSharedCryptoKeygen.KeyPair{
					{Public: nil, Private: kp.Private},
				}
			}(),
			wantErrContain: "failed to create CA subject 0:", // len(1)-1-0 = 0.
		},
		{
			name: "second key pair with nil public key",
			keyPairs: func() []*cryptoutilSharedCryptoKeygen.KeyPair {
				good := testKeyGenPool.GetMany(1)[0]
				bad := testKeyGenPool.GetMany(1)[0]

				return []*cryptoutilSharedCryptoKeygen.KeyPair{
					good,
					{Public: nil, Private: bad.Private}, // Root (processed first at i=1).
				}
			}(),
			wantErrContain: "failed to create CA subject 0:", // len(2)-1-1 = 0.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := CreateCASubjects(tc.keyPairs, "Error Test CA", testCACertValidity10Years)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErrContain)
		})
	}
}

// TestBuildTLSCertificate_IntermediatePoolContents kills the CONDITIONALS_BOUNDARY mutant at
// certificates.go:200:16 which changes `j < len(chain)-1` to `j <= len(chain)-1`, incorrectly
// adding the root cert to the intermediate pool. Verifies intermediate pool does NOT contain root.
func TestBuildTLSCertificate_IntermediatePoolContents(t *testing.T) {
	t.Parallel()

	// 4 key pairs: end entity + 2 intermediates + root = 4 certs in chain.
	keyPairs := testKeyGenPool.GetMany(4)

	caSubjects, err := CreateCASubjects(keyPairs[1:], "Pool Test CA", testCACertValidity10Years)
	require.NoError(t, err)

	endEntity, err := CreateEndEntitySubject(
		caSubjects[0], keyPairs[0], "Pool Test EE",
		testEndEntityCertValidity396Days,
		[]string{"localhost"},
		[]net.IP{net.ParseIP("127.0.0.1")},
		nil, nil,
		x509.KeyUsageDigitalSignature,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	)
	require.NoError(t, err)

	_, rootPool, intermediatePool, err := BuildTLSCertificate(endEntity)
	require.NoError(t, err)

	chain := endEntity.KeyMaterial.CertificateChain
	require.GreaterOrEqual(t, len(chain), 3, "need at least 3 certs for meaningful test")

	// Build expected pools.
	expectedRootPool := x509.NewCertPool()
	expectedRootPool.AddCert(chain[len(chain)-1]) // Only root cert.

	expectedIntermediatePool := x509.NewCertPool()
	for j := 1; j < len(chain)-1; j++ {
		expectedIntermediatePool.AddCert(chain[j]) // Only intermediates, NOT root.
	}

	require.True(t, rootPool.Equal(expectedRootPool), "root pool should contain only root cert")
	require.True(t, intermediatePool.Equal(expectedIntermediatePool), "intermediate pool should not contain root cert")
}
