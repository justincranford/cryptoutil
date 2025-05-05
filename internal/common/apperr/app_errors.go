package apperr

import "errors"

var (
	ErrCantBeNil        = errors.New("jwks can't be nil")
	ErrCantBeEmpty      = errors.New("jwks can't be empty")
	ErrUUIDCantBeNil    = errors.New("UUID can't be nil")
	ErrUUIDCantBeZero   = errors.New("UUID can't be zero UUID")
	ErrUUIDCantBeMax    = errors.New("UUID can't be max UUID")
	ErrUUIDsCantBeNil   = errors.New("UUIDs can't be nil")
	ErrUUIDsCantBeEmpty = errors.New("UUIDs can't be empty")

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

func IsAppErr(target error) bool {
	return ContainsError(Errs, target)
}

func ContainsError(errs []error, target error) bool {
	for _, err := range errs {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}
