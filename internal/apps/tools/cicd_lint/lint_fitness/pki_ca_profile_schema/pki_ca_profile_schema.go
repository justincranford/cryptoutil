// Copyright (c) 2025 Justin Cranford

// Package pki_ca_profile_schema validates the structural correctness of all
// PKI-CA certificate profile YAML files in configs/pki-ca/profiles/.
//
// Validation enforces:
//   - Required top-level field: profile
//   - Required profile fields: name, description, validity, key, key_usage, extended_key_usage
//   - validity: max_days >= min_days >= 1; default_days in [min_days, max_days]
//   - validity: max_days <= 10_950 (30 years absolute cap)
//   - key: at least one allowed_algorithms entry; default_algorithm is one of RSA/ECDSA/Ed25519
//   - key: each algorithm entry has required algorithm field
//   - key_usage: non-empty list; each value is a known RFC 5280 key usage name
//   - extended_key_usage: must have both required and optional list fields
//   - san (if present): required fields present; max_entries >= 0
package pki_ca_profile_schema

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

// known key usage values from RFC 5280.
var knownKeyUsages = map[string]bool{
	"digitalSignature": true,
	"nonRepudiation":   true,
	"keyEncipherment":  true,
	"dataEncipherment": true,
	"keyAgreement":     true,
	"keyCertSign":      true,
	"cRLSign":          true,
	"encipherOnly":     true,
	"decipherOnly":     true,
}

// known key algorithms.
var knownAlgorithms = map[string]bool{
	cryptoutilSharedMagic.KeyTypeRSA:     true,
	"ECDSA":                              true,
	cryptoutilSharedMagic.EdCurveEd25519: true,
}

const (
	maxValidityDaysAbsoluteCap = 10_950 // 30 years
)

// ProfileFile is the top-level structure of a PKI-CA profile YAML file.
type ProfileFile struct {
	Profile *ProfileSpec `yaml:"profile"`
}

// ProfileSpec contains all profile fields.
type ProfileSpec struct {
	Name             string            `yaml:"name"`
	Description      string            `yaml:"description"`
	Validity         *ValiditySpec     `yaml:"validity"`
	Key              *KeySpec          `yaml:"key"`
	KeyUsage         []string          `yaml:"key_usage"`
	ExtendedKeyUsage *ExtendedKeyUsage `yaml:"extended_key_usage"`
	Subject          map[string]any    `yaml:"subject"`
	SAN              *SANSpec          `yaml:"san"`
	Extensions       *ExtensionsSpec   `yaml:"extensions"`
}

// ValiditySpec holds validity period constraints.
type ValiditySpec struct {
	MaxDays     int `yaml:"max_days"`
	MinDays     int `yaml:"min_days"`
	DefaultDays int `yaml:"default_days"`
}

// KeySpec holds key algorithm constraints.
type KeySpec struct {
	AllowedAlgorithms  []AlgorithmEntry `yaml:"allowed_algorithms"`
	DefaultAlgorithm   string           `yaml:"default_algorithm"`
	DefaultCurveOrSize any              `yaml:"default_curve_or_size"`
}

// AlgorithmEntry represents one allowed key algorithm.
type AlgorithmEntry struct {
	Algorithm     string   `yaml:"algorithm"`
	MinSize       int      `yaml:"min_size"`
	MaxSize       int      `yaml:"max_size"`
	AllowedCurves []string `yaml:"allowed_curves"`
}

// ExtendedKeyUsage holds required and optional EKU values.
type ExtendedKeyUsage struct {
	Required []string `yaml:"required"`
	Optional []string `yaml:"optional"`
}

// SANSpec holds Subject Alternative Name constraints.
type SANSpec struct {
	AllowDNSNames       *bool `yaml:"allow_dns_names"`
	AllowIPAddresses    *bool `yaml:"allow_ip_addresses"`
	AllowEmailAddresses *bool `yaml:"allow_email_addresses"`
	AllowURIs           *bool `yaml:"allow_uris"`
	RequireAtLeastOne   *bool `yaml:"require_at_least_one"`
	MaxEntries          *int  `yaml:"max_entries"`
}

// ExtensionsSpec holds required and optional extension names.
type ExtensionsSpec struct {
	Required []string `yaml:"required"`
	Optional []string `yaml:"optional"`
}

// test seams.
var (
	findPKIProfileRootFn = func() (string, error) { return findPKIProfileProjectRoot(os.Getwd) }
)

// Check validates all PKI-CA profile YAML files from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	rootDir, err := findPKIProfileRootFn()
	if err != nil {
		return err
	}

	return CheckInDir(logger, rootDir, os.ReadFile, filepath.WalkDir)
}

// CheckInDir validates all PKI-CA profile YAML files under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, readFileFn func(string) ([]byte, error), walkDirFn func(string, fs.WalkDirFunc) error) error {
	profilesDir := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDPKICAProfilesDir))

	var violations []string

	err := walkDirFn(profilesDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("failed to walk %s: %w", profilesDir, walkErr)
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		// skip profile-schema.json (not .yaml so already skipped) and any schema files
		if strings.HasSuffix(filepath.Base(path), "-schema.yaml") {
			return nil
		}

		fileViolations, err := checkProfileFile(path, readFileFn)
		if err != nil {
			return err
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if err != nil {
		return err
	}

	if len(violations) > 0 {
		return fmt.Errorf("pki-ca-profile-schema: %d violation(s):\n%s", len(violations), strings.Join(violations, "\n"))
	}

	logger.Log(fmt.Sprintf("pki-ca-profile-schema: all %d profile files pass schema validation", countProfileFiles(profilesDir, walkDirFn)))

	return nil
}

// countProfileFiles counts YAML files in the profiles directory (non-schema).
func countProfileFiles(profilesDir string, walkDirFn func(string, fs.WalkDirFunc) error) int {
	count := 0

	_ = walkDirFn(profilesDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(filepath.Base(path), "-schema.yaml") {
			count++
		}

		return nil
	})

	return count
}

// checkProfileFile validates one profile YAML file and returns a list of violations.
func checkProfileFile(path string, readFileFn func(string) ([]byte, error)) ([]string, error) {
	data, err := readFileFn(path) //nolint:gosec // path is controlled by the tooling framework
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var pf ProfileFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	base := filepath.Base(path)

	var violations []string

	errs := validateProfile(pf, base)
	for _, e := range errs {
		violations = append(violations, fmt.Sprintf("  %s: %s", base, e))
	}

	return violations, nil
}

// validateProfile validates a parsed ProfileFile and returns error strings.
func validateProfile(pf ProfileFile, filename string) []string {
	var errs []string

	if pf.Profile == nil {
		return []string{"missing required top-level field 'profile'"}
	}

	p := pf.Profile

	if strings.TrimSpace(p.Name) == "" {
		errs = append(errs, "profile.name is required and must be non-empty")
	}

	if strings.TrimSpace(p.Description) == "" {
		errs = append(errs, "profile.description is required and must be non-empty")
	}

	errs = append(errs, validateValidity(p.Validity, filename)...)
	errs = append(errs, validateKey(p.Key)...)
	errs = append(errs, validateKeyUsage(p.KeyUsage)...)
	errs = append(errs, validateExtendedKeyUsage(p.ExtendedKeyUsage)...)

	if p.SAN != nil {
		errs = append(errs, validateSAN(p.SAN)...)
	}

	return errs
}

func validateValidity(v *ValiditySpec, filename string) []string {
	_ = filename

	var errs []string

	if v == nil {
		return []string{"profile.validity is required"}
	}

	if v.MinDays < 1 {
		errs = append(errs, fmt.Sprintf("profile.validity.min_days must be >= 1, got %d", v.MinDays))
	}

	if v.MaxDays < v.MinDays {
		errs = append(errs, fmt.Sprintf("profile.validity.max_days (%d) must be >= min_days (%d)", v.MaxDays, v.MinDays))
	}

	if v.MaxDays > maxValidityDaysAbsoluteCap {
		errs = append(errs, fmt.Sprintf("profile.validity.max_days (%d) exceeds absolute cap of %d (30 years)", v.MaxDays, maxValidityDaysAbsoluteCap))
	}

	if v.DefaultDays < v.MinDays || v.DefaultDays > v.MaxDays {
		errs = append(errs, fmt.Sprintf("profile.validity.default_days (%d) must be in [%d, %d]", v.DefaultDays, v.MinDays, v.MaxDays))
	}

	return errs
}

func validateKey(k *KeySpec) []string {
	var errs []string

	if k == nil {
		return []string{"profile.key is required"}
	}

	if len(k.AllowedAlgorithms) == 0 {
		errs = append(errs, "profile.key.allowed_algorithms must have at least one entry")
	}

	for i, alg := range k.AllowedAlgorithms {
		if strings.TrimSpace(alg.Algorithm) == "" {
			errs = append(errs, fmt.Sprintf("profile.key.allowed_algorithms[%d].algorithm is required", i))
		} else if !knownAlgorithms[alg.Algorithm] {
			errs = append(errs, fmt.Sprintf("profile.key.allowed_algorithms[%d].algorithm '%s' is not a known algorithm (RSA, ECDSA, Ed25519)", i, alg.Algorithm))
		}
	}

	if strings.TrimSpace(k.DefaultAlgorithm) == "" {
		errs = append(errs, "profile.key.default_algorithm is required")
	} else if !knownAlgorithms[k.DefaultAlgorithm] {
		errs = append(errs, fmt.Sprintf("profile.key.default_algorithm '%s' is not a known algorithm (RSA, ECDSA, Ed25519)", k.DefaultAlgorithm))
	}

	// default_curve_or_size may be null for Ed25519 (no curve or size parameter).
	if k.DefaultCurveOrSize == nil && k.DefaultAlgorithm != cryptoutilSharedMagic.EdCurveEd25519 {
		errs = append(errs, "profile.key.default_curve_or_size is required for RSA and ECDSA")
	}

	return errs
}

func validateKeyUsage(ku []string) []string {
	var errs []string

	if len(ku) == 0 {
		return []string{"profile.key_usage must have at least one entry"}
	}

	for _, usage := range ku {
		if !knownKeyUsages[usage] {
			errs = append(errs, fmt.Sprintf("profile.key_usage contains unknown value '%s'", usage))
		}
	}

	return errs
}

func validateExtendedKeyUsage(eku *ExtendedKeyUsage) []string {
	if eku == nil {
		return []string{"profile.extended_key_usage is required (must have required and optional lists)"}
	}

	// required and optional are allowed to be empty slices, that's fine.
	return nil
}

func validateSAN(san *SANSpec) []string {
	var errs []string

	if san.AllowDNSNames == nil {
		errs = append(errs, "profile.san.allow_dns_names is required")
	}

	if san.AllowIPAddresses == nil {
		errs = append(errs, "profile.san.allow_ip_addresses is required")
	}

	if san.AllowEmailAddresses == nil {
		errs = append(errs, "profile.san.allow_email_addresses is required")
	}

	if san.AllowURIs == nil {
		errs = append(errs, "profile.san.allow_uris is required")
	}

	if san.RequireAtLeastOne == nil {
		errs = append(errs, "profile.san.require_at_least_one is required")
	}

	if san.MaxEntries == nil {
		errs = append(errs, "profile.san.max_entries is required")
	} else if *san.MaxEntries < 0 {
		errs = append(errs, fmt.Sprintf("profile.san.max_entries must be >= 0, got %d", *san.MaxEntries))
	}

	return errs
}

// findPKIProfileProjectRoot walks up from the current working directory to find go.mod.
func findPKIProfileProjectRoot(getwdFn func() (string, error)) (string, error) {
	cwd, err := getwdFn()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	dir := cwd

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	return "", fmt.Errorf("go.mod not found: walked up from %s", cwd)
}
