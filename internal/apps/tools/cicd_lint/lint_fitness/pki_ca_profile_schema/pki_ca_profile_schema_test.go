// Copyright (c) 2025 Justin Cranford

package pki_ca_profile_schema

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------
// Test helpers
// -----------------------------------------------------------------------

// buildProfileRoot creates a temp root dir with a pkica profiles directory.
func buildProfileRoot(t *testing.T) string {
	t.Helper()

	rootDir := t.TempDir()
	profilesDir := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDPKICAProfilesDir))
	require.NoError(t, os.MkdirAll(profilesDir, cryptoutilSharedMagic.CacheFilePermissions))

	return rootDir
}

// writeProfileYAML writes a YAML file in the profiles directory.
func writeProfileYAML(t *testing.T, rootDir, name, content string) string {
	t.Helper()

	profilesDir := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDPKICAProfilesDir))
	path := filepath.Join(profilesDir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	return path
}

// minValidProfile returns a minimal valid profile YAML.
func minValidProfile(name string) string {
	return fmt.Sprintf(`profile:
  name: %q
  description: "Test profile for %s"
  validity:
    max_days: 365
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms:
      - algorithm: "RSA"
        min_size: 2048
        max_size: 4096
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required:
      - "serverAuth"
    optional: []
`, name, name)
}

func newTestLogger(t *testing.T) *cryptoutilCmdCicdCommon.Logger {
	t.Helper()

	return cryptoutilCmdCicdCommon.NewLogger("test-pki-ca-profile-schema")
}

// -----------------------------------------------------------------------
// CheckInDir — happy paths
// -----------------------------------------------------------------------

func TestCheckInDir_HappyPath_NoProfiles(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_HappyPath_ValidProfile(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "tls-server.yaml", minValidProfile("tls-server"))
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_HappyPath_Ed25519NullCurve(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "ssh-host.yaml", `profile:
  name: "ssh-host"
  description: "SSH Host Certificate"
  validity:
    max_days: 365
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms:
      - algorithm: "Ed25519"
    default_algorithm: "Ed25519"
    default_curve_or_size: null
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required: []
    optional: []
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_HappyPath_ZeroMinDays(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "k8s-workload.yaml", `profile:
  name: "k8s-workload"
  description: "Short-lived Kubernetes Workload Certificate"
  validity:
    max_days: 1
    min_days: 0
    default_days: 1
  key:
    allowed_algorithms:
      - algorithm: "ECDSA"
        allowed_curves:
          - "P-256"
    default_algorithm: "ECDSA"
    default_curve_or_size: "P-256"
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required: []
    optional: []
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

func TestCheckInDir_NonYAMLFileSkipped(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	profilesDir := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDPKICAProfilesDir))
	// Write a JSON schema file — it should be skipped (not .yaml extension)
	require.NoError(t, os.WriteFile(filepath.Join(profilesDir, "profile-schema.json"), []byte("{}"), cryptoutilSharedMagic.FilePermissionsDefault))
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.NoError(t, err)
}

// -----------------------------------------------------------------------
// CheckInDir — violations
// -----------------------------------------------------------------------

func TestCheckInDir_MissingProfileField(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "bad.yaml", "validity:\n  max_days: 365\n") // no 'profile:' key
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "violation")
}

func TestCheckInDir_EmptyProfileName(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "nameless.yaml", `profile:
  name: ""
  description: "Test"
  validity:
    max_days: 365
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms:
      - algorithm: "RSA"
        min_size: 2048
        max_size: 4096
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required: []
    optional: []
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "profile.name")
}

func TestCheckInDir_InvalidMaxDays(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "bad-validity.yaml", `profile:
  name: "test"
  description: "Test"
  validity:
    max_days: 5
    min_days: 10
    default_days: 7
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required: []
    optional: []
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "max_days")
}

func TestCheckInDir_DefaultDaysOutOfRange(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "bad-default.yaml", `profile:
  name: "test"
  description: "Test"
  validity:
    max_days: 365
    min_days: 1
    default_days: 500
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required: []
    optional: []
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "default_days")
}

func TestCheckInDir_MaxDaysExceedsCap(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "too-long.yaml", `profile:
  name: "too-long"
  description: "Test"
  validity:
    max_days: 99999
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required: []
    optional: []
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "absolute cap")
}

func TestCheckInDir_EmptyAllowedAlgorithms(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "no-alg.yaml", `profile:
  name: "test"
  description: "Test"
  validity:
    max_days: 365
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms: []
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required: []
    optional: []
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "at least one")
}

func TestCheckInDir_UnknownAlgorithm(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "bad-alg.yaml", `profile:
  name: "test"
  description: "Test"
  validity:
    max_days: 365
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms:
      - algorithm: "DSA"
    default_algorithm: "DSA"
    default_curve_or_size: 1024
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required: []
    optional: []
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "known algorithm")
}

func TestCheckInDir_EmptyKeyUsage(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "no-ku.yaml", `profile:
  name: "test"
  description: "Test"
  validity:
    max_days: 365
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage: []
  extended_key_usage:
    required: []
    optional: []
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "key_usage")
}

func TestCheckInDir_UnknownKeyUsage(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "unknown-ku.yaml", `profile:
  name: "test"
  description: "Test"
  validity:
    max_days: 365
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage:
    - "superPower"
  extended_key_usage:
    required: []
    optional: []
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown value")
}

func TestCheckInDir_MissingExtendedKeyUsage(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "no-eku.yaml", `profile:
  name: "test"
  description: "Test"
  validity:
    max_days: 365
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage:
    - "digitalSignature"
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "extended_key_usage")
}

func TestCheckInDir_SANNegativeMaxEntries(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "bad-san.yaml", `profile:
  name: "test"
  description: "Test"
  validity:
    max_days: 365
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required: []
    optional: []
  san:
    allow_dns_names: true
    allow_ip_addresses: false
    allow_email_addresses: false
    allow_uris: false
    require_at_least_one: true
    max_entries: -1
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "max_entries")
}

func TestCheckInDir_SANMissingRequiredFields(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "partial-san.yaml", `profile:
  name: "test"
  description: "Test"
  validity:
    max_days: 365
    min_days: 1
    default_days: 90
  key:
    allowed_algorithms:
      - algorithm: "RSA"
    default_algorithm: "RSA"
    default_curve_or_size: 2048
  key_usage:
    - "digitalSignature"
  extended_key_usage:
    required: []
    optional: []
  san:
    allow_dns_names: true
`)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "san.")
}

// -----------------------------------------------------------------------
// CheckInDir — seam injection tests
// -----------------------------------------------------------------------

func TestCheckInDir_WalkError(t *testing.T) {
	original := pkiProfileWalkDirFn

	pkiProfileWalkDirFn = func(_ string, _ fs.WalkDirFunc) error {
		return errors.New("injected walk error")
	}

	t.Cleanup(func() { pkiProfileWalkDirFn = original })

	rootDir := buildProfileRoot(t)
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "injected walk error")
}

func TestCheckInDir_ReadFileError(t *testing.T) {
	originalWalk := pkiProfileWalkDirFn
	originalRead := pkiProfileReadFileFn

	// Use real walk to find the file, then inject read error.
	pkiProfileReadFileFn = func(_ string) ([]byte, error) {
		return nil, errors.New("injected read error")
	}

	t.Cleanup(func() {
		pkiProfileWalkDirFn = originalWalk
		pkiProfileReadFileFn = originalRead
	})

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "any.yaml", minValidProfile("any"))
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "injected read error")
}

func TestCheckInDir_InvalidYAMLSkipsWithError(t *testing.T) {
	t.Parallel()

	rootDir := buildProfileRoot(t)
	writeProfileYAML(t, rootDir, "bad-syntax.yaml", "!!! not: valid: yaml: [unparseable")
	logger := newTestLogger(t)

	err := CheckInDir(logger, rootDir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse")
}

// -----------------------------------------------------------------------
// Check — project root seam tests
// -----------------------------------------------------------------------

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
	original := pkiProfileGetwdFn

	pkiProfileGetwdFn = func() (string, error) {
		return "", errors.New("injected getwd error")
	}

	t.Cleanup(func() { pkiProfileGetwdFn = original })

	_, err := findPKIProfileProjectRoot()

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get working directory")
}

func TestFindProjectRoot_GoModNotFound(t *testing.T) {
	original := pkiProfileGetwdFn

	pkiProfileGetwdFn = func() (string, error) {
		return t.TempDir(), nil
	}

	t.Cleanup(func() { pkiProfileGetwdFn = original })

	_, err := findPKIProfileProjectRoot()

	require.Error(t, err)
	require.Contains(t, err.Error(), "go.mod not found")
}

func TestFindProjectRoot_HappyPath(t *testing.T) {
	t.Parallel()

	// Real cwd is within the cryptoutil project so go.mod WILL be found.
	dir, err := findPKIProfileProjectRoot()

	require.NoError(t, err)
	require.DirExists(t, dir)
	require.FileExists(t, filepath.Join(dir, "go.mod"))
}

// -----------------------------------------------------------------------
// validateProfile — direct unit tests
// -----------------------------------------------------------------------

func TestValidateProfile_MissingProfileField(t *testing.T) {
	t.Parallel()

	errs := validateProfile(ProfileFile{}, "test.yaml")

	require.Len(t, errs, 1)
	require.Contains(t, errs[0], "missing required top-level field 'profile'")
}

func TestValidateProfile_MissingDescription(t *testing.T) {
	t.Parallel()

	pf := ProfileFile{
		Profile: &ProfileSpec{
			Name:        "test",
			Description: "  ", // whitespace only
			Validity:    &ValiditySpec{MaxDays: cryptoutilSharedMagic.DefaultCertificateMaxAgeDays, MinDays: 1, DefaultDays: cryptoutilSharedMagic.StrictCertificateMaxAgeDays},
			Key: &KeySpec{
				AllowedAlgorithms:  []AlgorithmEntry{{Algorithm: cryptoutilSharedMagic.KeyTypeRSA}},
				DefaultAlgorithm:   cryptoutilSharedMagic.KeyTypeRSA,
				DefaultCurveOrSize: cryptoutilSharedMagic.RSAKeySize2048,
			},
			KeyUsage:         []string{"digitalSignature"},
			ExtendedKeyUsage: &ExtendedKeyUsage{Required: []string{}, Optional: []string{}},
		},
	}

	errs := validateProfile(pf, "test.yaml")

	require.Len(t, errs, 1)
	require.Contains(t, errs[0], "profile.description")
}

func TestValidateProfile_MissingValidity(t *testing.T) {
	t.Parallel()

	pf := ProfileFile{
		Profile: &ProfileSpec{
			Name:        "test",
			Description: "desc",
			Key: &KeySpec{
				AllowedAlgorithms:  []AlgorithmEntry{{Algorithm: cryptoutilSharedMagic.KeyTypeRSA}},
				DefaultAlgorithm:   cryptoutilSharedMagic.KeyTypeRSA,
				DefaultCurveOrSize: cryptoutilSharedMagic.RSAKeySize2048,
			},
			KeyUsage:         []string{"digitalSignature"},
			ExtendedKeyUsage: &ExtendedKeyUsage{},
		},
	}

	errs := validateProfile(pf, "test.yaml")

	require.NotEmpty(t, errs)
	require.Contains(t, errs[0], "profile.validity is required")
}

func TestValidateProfile_MissingKey(t *testing.T) {
	t.Parallel()

	pf := ProfileFile{
		Profile: &ProfileSpec{
			Name:             "test",
			Description:      "desc",
			Validity:         &ValiditySpec{MaxDays: cryptoutilSharedMagic.DefaultCertificateMaxAgeDays, MinDays: 1, DefaultDays: cryptoutilSharedMagic.StrictCertificateMaxAgeDays},
			KeyUsage:         []string{"digitalSignature"},
			ExtendedKeyUsage: &ExtendedKeyUsage{},
		},
	}

	errs := validateProfile(pf, "test.yaml")

	require.NotEmpty(t, errs)
	require.Contains(t, errs[0], "profile.key is required")
}

func TestValidateProfile_AlgorithmMissingName(t *testing.T) {
	t.Parallel()

	pf := ProfileFile{
		Profile: &ProfileSpec{
			Name:             "test",
			Description:      "desc",
			Validity:         &ValiditySpec{MaxDays: cryptoutilSharedMagic.DefaultCertificateMaxAgeDays, MinDays: 1, DefaultDays: cryptoutilSharedMagic.StrictCertificateMaxAgeDays},
			Key:              &KeySpec{AllowedAlgorithms: []AlgorithmEntry{{Algorithm: ""}}, DefaultAlgorithm: cryptoutilSharedMagic.KeyTypeRSA, DefaultCurveOrSize: cryptoutilSharedMagic.RSAKeySize2048},
			KeyUsage:         []string{"digitalSignature"},
			ExtendedKeyUsage: &ExtendedKeyUsage{},
		},
	}

	errs := validateProfile(pf, "test.yaml")

	require.NotEmpty(t, errs)

	hasAlgErr := false

	for _, e := range errs {
		if contains(e, "algorithm") {
			hasAlgErr = true
		}
	}

	require.True(t, hasAlgErr)
}

func TestValidateProfile_UnknownDefaultAlgorithm(t *testing.T) {
	t.Parallel()

	pf := ProfileFile{
		Profile: &ProfileSpec{
			Name:             "test",
			Description:      "desc",
			Validity:         &ValiditySpec{MaxDays: cryptoutilSharedMagic.DefaultCertificateMaxAgeDays, MinDays: 1, DefaultDays: cryptoutilSharedMagic.StrictCertificateMaxAgeDays},
			Key:              &KeySpec{AllowedAlgorithms: []AlgorithmEntry{{Algorithm: cryptoutilSharedMagic.KeyTypeRSA}}, DefaultAlgorithm: "QuantumAlgo", DefaultCurveOrSize: cryptoutilSharedMagic.RSAKeySize2048},
			KeyUsage:         []string{"digitalSignature"},
			ExtendedKeyUsage: &ExtendedKeyUsage{},
		},
	}

	errs := validateProfile(pf, "test.yaml")

	require.NotEmpty(t, errs)

	hasErr := false

	for _, e := range errs {
		if contains(e, "known algorithm") {
			hasErr = true
		}
	}

	require.True(t, hasErr)
}

func TestValidateProfile_NilDefaultCurveOrSizeForRSA(t *testing.T) {
	t.Parallel()

	pf := ProfileFile{
		Profile: &ProfileSpec{
			Name:             "test",
			Description:      "desc",
			Validity:         &ValiditySpec{MaxDays: cryptoutilSharedMagic.DefaultCertificateMaxAgeDays, MinDays: 1, DefaultDays: cryptoutilSharedMagic.StrictCertificateMaxAgeDays},
			Key:              &KeySpec{AllowedAlgorithms: []AlgorithmEntry{{Algorithm: cryptoutilSharedMagic.KeyTypeRSA}}, DefaultAlgorithm: cryptoutilSharedMagic.KeyTypeRSA, DefaultCurveOrSize: nil},
			KeyUsage:         []string{"digitalSignature"},
			ExtendedKeyUsage: &ExtendedKeyUsage{},
		},
	}

	errs := validateProfile(pf, "test.yaml")

	hasErr := false

	for _, e := range errs {
		if contains(e, "default_curve_or_size") {
			hasErr = true
		}
	}

	require.True(t, hasErr)
}

func TestValidateProfile_NilDefaultCurveOrSizeForEd25519_Allowed(t *testing.T) {
	t.Parallel()

	pf := ProfileFile{
		Profile: &ProfileSpec{
			Name:             "test",
			Description:      "desc",
			Validity:         &ValiditySpec{MaxDays: cryptoutilSharedMagic.DefaultCertificateMaxAgeDays, MinDays: 1, DefaultDays: cryptoutilSharedMagic.StrictCertificateMaxAgeDays},
			Key:              &KeySpec{AllowedAlgorithms: []AlgorithmEntry{{Algorithm: cryptoutilSharedMagic.EdCurveEd25519}}, DefaultAlgorithm: cryptoutilSharedMagic.EdCurveEd25519, DefaultCurveOrSize: nil},
			KeyUsage:         []string{"digitalSignature"},
			ExtendedKeyUsage: &ExtendedKeyUsage{},
		},
	}

	errs := validateProfile(pf, "test.yaml")

	for _, e := range errs {
		require.NotContains(t, e, "default_curve_or_size")
	}
}

// -----------------------------------------------------------------------
// Boundary tests — kill CONDITIONALS_BOUNDARY / CONDITIONALS_NEGATION mutations
// -----------------------------------------------------------------------

func TestValidateValidity_MaxDaysEqualMinDays(t *testing.T) {
	t.Parallel()

	days := cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days

	errs := validateValidity(&ValiditySpec{MinDays: days, MaxDays: days, DefaultDays: days}, "boundary.yaml")

	require.Empty(t, errs, "max_days == min_days should be valid")
}

func TestValidateValidity_MaxDaysExactlyCap(t *testing.T) {
	t.Parallel()

	errs := validateValidity(&ValiditySpec{MinDays: 1, MaxDays: maxValidityDaysAbsoluteCap, DefaultDays: cryptoutilSharedMagic.StrictCertificateMaxAgeDays}, "cap.yaml")

	require.Empty(t, errs, "max_days == absolute cap should be valid")
}

func TestValidateValidity_DefaultDaysEqualMinDays(t *testing.T) {
	t.Parallel()

	days30 := cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days
	days365 := cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year

	errs := validateValidity(&ValiditySpec{MinDays: days30, MaxDays: days365, DefaultDays: days30}, "boundary.yaml")

	require.Empty(t, errs, "default_days == min_days should be valid")
}

func TestValidateValidity_DefaultDaysEqualMaxDays(t *testing.T) {
	t.Parallel()

	days365 := cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year

	errs := validateValidity(&ValiditySpec{MinDays: 1, MaxDays: days365, DefaultDays: days365}, "boundary.yaml")

	require.Empty(t, errs, "default_days == max_days should be valid")
}

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
// Helpers
// -----------------------------------------------------------------------

// contains is a nil-safe substring check for test assertions.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || stringContains(s, substr))
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}

	return false
}
