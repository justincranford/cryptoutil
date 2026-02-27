// Copyright (c) 2025 Justin Cranford
//
//

// Package apperr provides application-level error types and HTTP error utilities.
package apperr

import "errors"

// Application error variables for common error conditions.
var (
	// ErrCantBeNil indicates that a value cannot be nil.
	ErrCantBeNil = errors.New("jwks can't be nil")
	// ErrCantBeEmpty indicates that a value cannot be empty.
	ErrCantBeEmpty = errors.New("jwks can't be empty")
	// ErrUUIDCantBeNil indicates that a UUID cannot be nil.
	ErrUUIDCantBeNil = errors.New("UUID can't be nil")
	// ErrUUIDCantBeZero indicates that a UUID cannot be zero.
	ErrUUIDCantBeZero = errors.New("UUID can't be zero UUID")
	// ErrUUIDCantBeMax indicates that a UUID cannot be max value.
	ErrUUIDCantBeMax = errors.New("UUID can't be max UUID")
	// ErrUUIDsCantBeNil indicates that UUIDs slice cannot be nil.
	ErrUUIDsCantBeNil = errors.New("UUIDs can't be nil")
	// ErrUUIDsCantBeEmpty indicates that UUIDs slice cannot be empty.
	ErrUUIDsCantBeEmpty = errors.New("UUIDs can't be empty")
	// ErrJWKMustBeEncryptJWK indicates that a JWK must be an encrypt JWK.
	ErrJWKMustBeEncryptJWK = errors.New("JWK must be an encrypt JWK")
	// ErrJWKMustBeDecryptJWK indicates that a JWK must be a decrypt JWK.
	ErrJWKMustBeDecryptJWK = errors.New("JWK must be a decrypt JWK")
	// ErrJWKMustBeSignJWK indicates that a JWK must be a sign JWK.
	ErrJWKMustBeSignJWK = errors.New("JWK must be a sign JWK")
	// ErrJWKMustBeVerifyJWK indicates that a JWK must be a verify JWK.
	ErrJWKMustBeVerifyJWK = errors.New("JWK must be a verify JWK")

	// Errs is a collection of all application errors for validation.
	Errs = []error{
		ErrCantBeNil,
		ErrCantBeEmpty,
		ErrUUIDCantBeNil,
		ErrUUIDCantBeZero,
		ErrUUIDCantBeMax,
		ErrUUIDsCantBeNil,
		ErrUUIDsCantBeEmpty,
	}
)

// IsAppErr checks if the target error is an application error.
func IsAppErr(target error) bool {
	return ContainsError(Errs, target)
}

// ContainsError checks if the target error exists in the errs slice.
func ContainsError(errs []error, target error) bool {
	for _, err := range errs {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}
