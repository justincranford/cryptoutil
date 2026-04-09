// Copyright (c) 2025 Justin Cranford

package pki_ca_profile_schema

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------
// validateProfile — direct unit tests
// -----------------------------------------------------------------------

func TestValidateProfile(t *testing.T) {
	t.Parallel()

	baseProfile := func(mutate func(*ProfileSpec)) ProfileFile {
		p := &ProfileSpec{
			Name:        "test",
			Description: "desc",
			Validity:    &ValiditySpec{MaxDays: cryptoutilSharedMagic.DefaultCertificateMaxAgeDays, MinDays: 1, DefaultDays: cryptoutilSharedMagic.StrictCertificateMaxAgeDays},
			Key: &KeySpec{
				AllowedAlgorithms:  []AlgorithmEntry{{Algorithm: cryptoutilSharedMagic.KeyTypeRSA}},
				DefaultAlgorithm:   cryptoutilSharedMagic.KeyTypeRSA,
				DefaultCurveOrSize: cryptoutilSharedMagic.RSAKeySize2048,
			},
			KeyUsage:         []string{"digitalSignature"},
			ExtendedKeyUsage: &ExtendedKeyUsage{Required: []string{}, Optional: []string{}},
		}
		if mutate != nil {
			mutate(p)
		}

		return ProfileFile{Profile: p}
	}

	tests := []struct {
		name      string
		pf        ProfileFile
		wantErr   string
		wantEmpty bool
	}{
		{
			name:    "missing profile field",
			pf:      ProfileFile{},
			wantErr: "missing required top-level field 'profile'",
		},
		{
			name: "missing description",
			pf: baseProfile(func(p *ProfileSpec) {
				p.Description = "  "
			}),
			wantErr: "profile.description",
		},
		{
			name: "missing validity",
			pf: baseProfile(func(p *ProfileSpec) {
				p.Validity = nil
			}),
			wantErr: "profile.validity is required",
		},
		{
			name: "missing key",
			pf: baseProfile(func(p *ProfileSpec) {
				p.Key = nil
			}),
			wantErr: "profile.key is required",
		},
		{
			name: "algorithm missing name",
			pf: baseProfile(func(p *ProfileSpec) {
				p.Key.AllowedAlgorithms = []AlgorithmEntry{{Algorithm: ""}}
			}),
			wantErr: "algorithm",
		},
		{
			name: "unknown default algorithm",
			pf: baseProfile(func(p *ProfileSpec) {
				p.Key.DefaultAlgorithm = "QuantumAlgo"
			}),
			wantErr: "known algorithm",
		},
		{
			name: "nil default curve or size for RSA",
			pf: baseProfile(func(p *ProfileSpec) {
				p.Key.DefaultCurveOrSize = nil
			}),
			wantErr: "default_curve_or_size",
		},
		{
			name: "nil default curve or size for Ed25519 is allowed",
			pf: baseProfile(func(p *ProfileSpec) {
				p.Key.AllowedAlgorithms = []AlgorithmEntry{{Algorithm: cryptoutilSharedMagic.EdCurveEd25519}}
				p.Key.DefaultAlgorithm = cryptoutilSharedMagic.EdCurveEd25519
				p.Key.DefaultCurveOrSize = nil
			}),
			wantEmpty: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errs := validateProfile(tc.pf, "test.yaml")

			if tc.wantEmpty {
				for _, e := range errs {
					require.NotContains(t, e, "default_curve_or_size")
				}

				return
			}

			require.NotEmpty(t, errs)
			requireAnyContains(t, errs, tc.wantErr)
		})
	}
}

// -----------------------------------------------------------------------
// Boundary tests — kill CONDITIONALS_BOUNDARY / CONDITIONALS_NEGATION mutations
// -----------------------------------------------------------------------

func TestValidateValidity(t *testing.T) {
	t.Parallel()

	days30 := cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days
	days365 := cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year

	tests := []struct {
		name      string
		spec      *ValiditySpec
		wantEmpty bool
		wantErr   string
	}{
		{
			name:      "min days one is valid",
			spec:      &ValiditySpec{MinDays: 1, MaxDays: 1, DefaultDays: 1},
			wantEmpty: true,
		},
		{
			name:    "min days zero is error",
			spec:    &ValiditySpec{MinDays: 0, MaxDays: days30, DefaultDays: 1},
			wantErr: "min_days must be >= 1",
		},
		{
			name:    "min days negative is error",
			spec:    &ValiditySpec{MinDays: -1, MaxDays: days30, DefaultDays: 1},
			wantErr: "min_days must be >= 1",
		},
		{
			name:      "max days equal min days",
			spec:      &ValiditySpec{MinDays: days30, MaxDays: days30, DefaultDays: days30},
			wantEmpty: true,
		},
		{
			name:      "max days exactly cap",
			spec:      &ValiditySpec{MinDays: 1, MaxDays: maxValidityDaysAbsoluteCap, DefaultDays: cryptoutilSharedMagic.StrictCertificateMaxAgeDays},
			wantEmpty: true,
		},
		{
			name:      "default days equal min days",
			spec:      &ValiditySpec{MinDays: days30, MaxDays: days365, DefaultDays: days30},
			wantEmpty: true,
		},
		{
			name:      "default days equal max days",
			spec:      &ValiditySpec{MinDays: 1, MaxDays: days365, DefaultDays: days365},
			wantEmpty: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errs := validateValidity(tc.spec, "boundary.yaml")

			if tc.wantEmpty {
				require.Empty(t, errs)
			} else {
				require.NotEmpty(t, errs)
				require.Contains(t, errs[0], tc.wantErr)
			}
		})
	}
}

// -----------------------------------------------------------------------
// validateSAN — all fields populated
// -----------------------------------------------------------------------

func TestValidateSAN_AllFieldsPopulated(t *testing.T) {
	t.Parallel()

	boolTrue := true
	boolFalse := false
	zeroEntries := 0

	san := &SANSpec{
		AllowDNSNames:       &boolTrue,
		AllowIPAddresses:    &boolFalse,
		AllowEmailAddresses: &boolFalse,
		AllowURIs:           &boolFalse,
		RequireAtLeastOne:   &boolTrue,
		MaxEntries:          &zeroEntries,
	}

	errs := validateSAN(san)

	require.Empty(t, errs, "SAN with all fields set (including max_entries=0) should be valid")
}

// -----------------------------------------------------------------------
// Check — project root seam tests
// -----------------------------------------------------------------------

// Sequential: mutates findPKIProfileRootFn package-level state.
func TestCheck_ProjectRootNotFound(t *testing.T) {
	original := findPKIProfileRootFn

	findPKIProfileRootFn = func() (string, error) {
		return "", errors.New("injected root error")
	}

	t.Cleanup(func() { findPKIProfileRootFn = original })

	logger := newTestLogger(t)

	err := Check(logger)

	require.Error(t, err)
	require.Contains(t, err.Error(), "injected root error")
}

// Sequential: mutates findPKIProfileRootFn package-level state.
func TestCheck_HappyPath(t *testing.T) {
	original := findPKIProfileRootFn

	findPKIProfileRootFn = func() (string, error) {
		return buildProfileRoot(t), nil
	}

	t.Cleanup(func() { findPKIProfileRootFn = original })

	logger := newTestLogger(t)

	err := Check(logger)

	require.NoError(t, err)
}

// -----------------------------------------------------------------------
// findPKIProfileProjectRoot seam tests
// -----------------------------------------------------------------------

func TestFindProjectRoot_GetwdError(t *testing.T) {
	t.Parallel()

	_, err := findPKIProfileProjectRoot(func() (string, error) {
		return "", errors.New("injected getwd error")
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get working directory")
}

func TestFindProjectRoot_GoModNotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := findPKIProfileProjectRoot(func() (string, error) {
		return tmpDir, nil
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "go.mod not found")
}

func TestFindProjectRoot_HappyPath(t *testing.T) {
	t.Parallel()

	// Real cwd is within the cryptoutil project so go.mod WILL be found.
	dir, err := findPKIProfileProjectRoot(os.Getwd)

	require.NoError(t, err)
	require.DirExists(t, dir)
	require.FileExists(t, filepath.Join(dir, "go.mod"))
}

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

// requireAnyContains asserts at least one string in the slice contains substr.
func requireAnyContains(t *testing.T, strs []string, substr string) {
	t.Helper()

	for _, s := range strs {
		if len(s) >= len(substr) && containsSubstr(s, substr) {
			return
		}
	}

	t.Fatalf("expected at least one string in %v to contain %q", strs, substr)
}

// containsSubstr is a simple substring search.
func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}

	return false
}
