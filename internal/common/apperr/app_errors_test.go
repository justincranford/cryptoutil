// Copyright (c) 2025 Justin Cranford

package apperr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
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
			target:   cryptoutilAppErr.ErrCantBeNil,
			expected: true,
		},
		{
			name:     "is-apperr-cant-be-empty",
			target:   cryptoutilAppErr.ErrCantBeEmpty,
			expected: true,
		},
		{
			name:     "is-apperr-uuid-cant-be-nil",
			target:   cryptoutilAppErr.ErrUUIDCantBeNil,
			expected: true,
		},
		{
			name:     "is-apperr-uuid-cant-be-zero",
			target:   cryptoutilAppErr.ErrUUIDCantBeZero,
			expected: true,
		},
		{
			name:     "is-apperr-uuid-cant-be-max",
			target:   cryptoutilAppErr.ErrUUIDCantBeMax,
			expected: true,
		},
		{
			name:     "is-apperr-uuids-cant-be-nil",
			target:   cryptoutilAppErr.ErrUUIDsCantBeNil,
			expected: true,
		},
		{
			name:     "is-apperr-uuids-cant-be-empty",
			target:   cryptoutilAppErr.ErrUUIDsCantBeEmpty,
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

			result := cryptoutilAppErr.IsAppErr(tc.target)
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

			result := cryptoutilAppErr.ContainsError(tc.errs, tc.target)
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
			err:  cryptoutilAppErr.ErrJWKMustBeEncryptJWK,
			msg:  "JWK must be an encrypt JWK",
		},
		{
			name: "jwk-must-be-decrypt-jwk",
			err:  cryptoutilAppErr.ErrJWKMustBeDecryptJWK,
			msg:  "JWK must be a decrypt JWK",
		},
		{
			name: "jwk-must-be-sign-jwk",
			err:  cryptoutilAppErr.ErrJWKMustBeSignJWK,
			msg:  "JWK must be a sign JWK",
		},
		{
			name: "jwk-must-be-verify-jwk",
			err:  cryptoutilAppErr.ErrJWKMustBeVerifyJWK,
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
		cryptoutilAppErr.ErrCantBeNil,
		cryptoutilAppErr.ErrCantBeEmpty,
		cryptoutilAppErr.ErrUUIDCantBeNil,
		cryptoutilAppErr.ErrUUIDCantBeZero,
		cryptoutilAppErr.ErrUUIDCantBeMax,
		cryptoutilAppErr.ErrUUIDsCantBeNil,
		cryptoutilAppErr.ErrUUIDsCantBeEmpty,
	}

	require.Len(t, cryptoutilAppErr.Errs, len(expectedErrs))

	for _, expectedErr := range expectedErrs {
		found := false

		for _, actualErr := range cryptoutilAppErr.Errs {
			if errors.Is(actualErr, expectedErr) {
				found = true

				break
			}
		}

		require.True(t, found, "Expected error %v to be in Errs slice", expectedErr)
	}
}
