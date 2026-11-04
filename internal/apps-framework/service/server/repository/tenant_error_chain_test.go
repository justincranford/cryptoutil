// Copyright (c) 2025-2026 Justin Cranford.

package repository

import (
	"errors"
	"testing"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRepositoryErrorWrappingPreservesCauseAndHTTPIdentity(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		cause error
		code  int
	}{
		{"not found", gorm.ErrRecordNotFound, 404},
		{"duplicate", gorm.ErrDuplicatedKey, 409},
		{"SQLite constraint", errors.New("UNIQUE constraint"), 409},
		{"SQLite code", errors.New("(2067)"), 409},
		{"generic", errors.New("bounded synthetic database error"), 500},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mapped := toAppErr(test.cause)
			require.ErrorIs(t, mapped, test.cause)

			var typed *cryptoutilSharedApperr.Error
			require.ErrorAs(t, mapped, &typed)
			require.Equal(t, test.code, int(typed.HTTPStatusLineAndCode.StatusLine.StatusCode))
			require.Same(t, test.cause, typed.Err)
			require.ErrorIs(t, mapped, typed, "the typed application error must remain in the chain")
		})
	}
}
