// Copyright (c) 2025 Justin Cranford

package apperr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

func TestIsAppErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		target   error
		expected bool
	}{
		{
			name:     "is-apperr-cant-be-nil",
			target:   cryptoutilSharedApperr.ErrCantBeNil,
			expected: true,
		},
		{
			name:     "is-apperr-cant-be-empty",
			target:   cryptoutilSharedApperr.ErrCantBeEmpty,
			expected: true,
		},
		{
			name:     "is-apperr-uuid-cant-be-nil",
			target:   cryptoutilSharedApperr.ErrUUIDCantBeNil,
			expected: true,
		},
		{
			name:     "is-apperr-uuid-cant-be-zero",
			target:   cryptoutilSharedApperr.ErrUUIDCantBeZero,
			expected: true,
		},
		{
			name:     "is-apperr-uuid-cant-be-max",
			target:   cryptoutilSharedApperr.ErrUUIDCantBeMax,
			expected: true,
		},
		{
			name:     "is-apperr-uuids-cant-be-nil",
			target:   cryptoutilSharedApperr.ErrUUIDsCantBeNil,
			expected: true,
		},
		{
			name:     "is-apperr-uuids-cant-be-empty",
			target:   cryptoutilSharedApperr.ErrUUIDsCantBeEmpty,
			expected: true,
		},
		{
			name:     "is-not-apperr-random-error",
			target:   errors.New("random error"),
			expected: false,
		},
		{
			name:     "is-not-apperr-nil",
			target:   nil,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilSharedApperr.IsAppErr(tc.target)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestContainsError(t *testing.T) {
	t.Parallel()

	errOne := errors.New("error one")
	errTwo := errors.New("error two")
	errThree := errors.New("error three")
	errFour := errors.New("error four")

	errs := []error{errOne, errTwo, errThree}

	tests := []struct {
		name     string
		errs     []error
		target   error
		expected bool
	}{
		{
			name:     "contains-first-error",
			errs:     errs,
			target:   errOne,
			expected: true,
		},
		{
			name:     "contains-middle-error",
			errs:     errs,
			target:   errTwo,
			expected: true,
		},
		{
			name:     "contains-last-error",
			errs:     errs,
			target:   errThree,
			expected: true,
		},
		{
			name:     "does-not-contain-error",
			errs:     errs,
			target:   errFour,
			expected: false,
		},
		{
			name:     "empty-slice-no-match",
			errs:     []error{},
			target:   errOne,
			expected: false,
		},
		{
			name:     "nil-slice-no-match",
			errs:     nil,
			target:   errOne,
			expected: false,
		},
		{
			name:     "target-is-nil",
			errs:     errs,
			target:   nil,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilSharedApperr.ContainsError(tc.errs, tc.target)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestJWKErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{
			name: "jwk-must-be-encrypt-jwk",
			err:  cryptoutilSharedApperr.ErrJWKMustBeEncryptJWK,
			msg:  "JWK must be an encrypt JWK",
		},
		{
			name: "jwk-must-be-decrypt-jwk",
			err:  cryptoutilSharedApperr.ErrJWKMustBeDecryptJWK,
			msg:  "JWK must be a decrypt JWK",
		},
		{
			name: "jwk-must-be-sign-jwk",
			err:  cryptoutilSharedApperr.ErrJWKMustBeSignJWK,
			msg:  "JWK must be a sign JWK",
		},
		{
			name: "jwk-must-be-verify-jwk",
			err:  cryptoutilSharedApperr.ErrJWKMustBeVerifyJWK,
			msg:  "JWK must be a verify JWK",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, tc.err)
			require.Equal(t, tc.msg, tc.err.Error())
		})
	}
}

func TestErrsSliceContainsAllExpectedErrors(t *testing.T) {
	t.Parallel()

	// Verify that Errs slice contains exactly the expected errors.
	expectedErrs := []error{
		cryptoutilSharedApperr.ErrCantBeNil,
		cryptoutilSharedApperr.ErrCantBeEmpty,
		cryptoutilSharedApperr.ErrUUIDCantBeNil,
		cryptoutilSharedApperr.ErrUUIDCantBeZero,
		cryptoutilSharedApperr.ErrUUIDCantBeMax,
		cryptoutilSharedApperr.ErrUUIDsCantBeNil,
		cryptoutilSharedApperr.ErrUUIDsCantBeEmpty,
	}

	require.Len(t, cryptoutilSharedApperr.Errs, len(expectedErrs))

	for _, expectedErr := range expectedErrs {
		found := false

		for _, actualErr := range cryptoutilSharedApperr.Errs {
			if errors.Is(actualErr, expectedErr) {
				found = true

				break
			}
		}

		require.True(t, found, "Expected error %v to be in Errs slice", expectedErr)
	}
}
