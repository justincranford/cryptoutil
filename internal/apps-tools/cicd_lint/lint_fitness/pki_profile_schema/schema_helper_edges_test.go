// Copyright (c) 2025-2026 Justin Cranford.
package pki_profile_schema

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testEmptyAllowedAlgorithmsCaseName = "empty allowed algorithms"
	testProfileKeyRequiredError        = "profile.key is required"
)

func TestCheckFromWorkspaceRoot(t *testing.T) {
	t.Parallel()

	require.NoError(t, Check(newTestLogger(t)))
}

func TestCountProfileFilesIgnoresWalkErrors(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Empty(t, entries)

	filePath := filepath.Join(root, "profile.yaml")
	require.NoError(t, os.WriteFile(filePath, nil, cryptoutilSharedMagic.CacheFilePermissions))

	entries, err = os.ReadDir(root)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	count := countProfileFiles(root, func(path string, fn fs.WalkDirFunc) error {
		return fn(path, entries[0], errors.New("injected walk error"))
	})

	require.Zero(t, count)
}

func TestValidateKeyConditions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  *KeySpec
		want string
	}{
		{name: "nil key", want: testProfileKeyRequiredError},
		{name: testEmptyAllowedAlgorithmsCaseName, key: &KeySpec{DefaultAlgorithm: cryptoutilSharedMagic.KeyTypeRSA, DefaultCurveOrSize: 2048}, want: "must have at least one"},
		{name: "blank allowed algorithm", key: &KeySpec{AllowedAlgorithms: []AlgorithmEntry{{Algorithm: " "}}, DefaultAlgorithm: cryptoutilSharedMagic.KeyTypeRSA, DefaultCurveOrSize: 2048}, want: "algorithm is required"},
		{name: "blank default algorithm", key: &KeySpec{AllowedAlgorithms: []AlgorithmEntry{{Algorithm: cryptoutilSharedMagic.KeyTypeRSA}}}, want: "default_algorithm is required"},
		{name: "curve required for RSA", key: &KeySpec{AllowedAlgorithms: []AlgorithmEntry{{Algorithm: cryptoutilSharedMagic.KeyTypeRSA}}, DefaultAlgorithm: cryptoutilSharedMagic.KeyTypeRSA}, want: "default_curve_or_size is required"},
		{name: "Ed25519 curve is optional", key: &KeySpec{AllowedAlgorithms: []AlgorithmEntry{{Algorithm: cryptoutilSharedMagic.EdCurveEd25519}}, DefaultAlgorithm: cryptoutilSharedMagic.EdCurveEd25519}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errs := validateKey(tc.key)
			if tc.want == "" {
				require.Empty(t, errs)

				return
			}

			require.Contains(t, strings.Join(errs, "\n"), tc.want)
		})
	}
}

func TestValidateSANMaximumEntryBoundary(t *testing.T) {
	t.Parallel()

	allow := true
	maxEntries := 0
	san := &SANSpec{
		AllowDNSNames:       &allow,
		AllowIPAddresses:    &allow,
		AllowEmailAddresses: &allow,
		AllowURIs:           &allow,
		RequireAtLeastOne:   &allow,
		MaxEntries:          &maxEntries,
	}

	require.Empty(t, validateSAN(san))
}

func TestValidateSANMissingDNSSetting(t *testing.T) {
	t.Parallel()

	allow := true
	maxEntries := 1
	san := &SANSpec{
		AllowIPAddresses:    &allow,
		AllowEmailAddresses: &allow,
		AllowURIs:           &allow,
		RequireAtLeastOne:   &allow,
		MaxEntries:          &maxEntries,
	}

	require.Contains(t, strings.Join(validateSAN(san), "\n"), "profile.san.allow_dns_names is required")
}
