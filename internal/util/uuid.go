package util

import (
	"errors"

	googleUuid "github.com/google/uuid"
)

var (
	ZeroUUID       = googleUuid.UUID{}
	ErrNonZeroUUID = errors.New("UUID must not be 00000000-0000-0000-0000-000000000000")
)
