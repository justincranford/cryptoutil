// Copyright 2025 Marlon Almeida. All rights reserved.

package repository

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

func TestToAppErr_NilError(t *testing.T) {
	t.Parallel()

	result := toAppErr(nil)
	require.NoError(t, result)
}

func TestToAppErr_RecordNotFound(t *testing.T) {
	t.Parallel()

	err := toAppErr(gorm.ErrRecordNotFound)
	require.Error(t, err)

	appErr := &cryptoutilSharedApperr.Error{}
	ok := errors.As(err, &appErr)
	require.True(t, ok)
	require.Equal(t, 404, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestToAppErr_DuplicatedKey(t *testing.T) {
	t.Parallel()

	err := toAppErr(gorm.ErrDuplicatedKey)
	require.Error(t, err)

	appErr := &cryptoutilSharedApperr.Error{}
	ok := errors.As(err, &appErr)
	require.True(t, ok)
	require.Equal(t, 409, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestToAppErr_SQLiteUniqueConstraint(t *testing.T) {
	t.Parallel()

	err := toAppErr(errors.New("constraint failed: UNIQUE constraint failed: users.email (2067)"))
	require.Error(t, err)

	appErr := &cryptoutilSharedApperr.Error{}
	ok := errors.As(err, &appErr)
	require.True(t, ok)
	require.Equal(t, 409, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestToAppErr_SQLiteUniqueConstraintErrorCode(t *testing.T) {
	t.Parallel()

	err := toAppErr(errors.New("some error (2067)"))
	require.Error(t, err)

	appErr := &cryptoutilSharedApperr.Error{}
	ok := errors.As(err, &appErr)
	require.True(t, ok)
	require.Equal(t, 409, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestToAppErr_GenericDatabaseError(t *testing.T) {
	t.Parallel()

	err := toAppErr(errors.New("database connection failed"))
	require.Error(t, err)

	appErr := &cryptoutilSharedApperr.Error{}
	ok := errors.As(err, &appErr)
	require.True(t, ok)
	require.Equal(t, cryptoutilSharedMagic.TestDefaultRateLimitServiceIP, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestToAppErr_WrappedRecordNotFound(t *testing.T) {
	t.Parallel()

	wrappedErr := errors.Join(gorm.ErrRecordNotFound, errors.New("additional context"))
	err := toAppErr(wrappedErr)
	require.Error(t, err)

	appErr := &cryptoutilSharedApperr.Error{}
	ok := errors.As(err, &appErr)
	require.True(t, ok)
	require.Equal(t, 404, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestToAppErr_WrappedDuplicatedKey(t *testing.T) {
	t.Parallel()

	wrappedErr := errors.Join(gorm.ErrDuplicatedKey, errors.New("additional context"))
	err := toAppErr(wrappedErr)
	require.Error(t, err)

	appErr := &cryptoutilSharedApperr.Error{}
	ok := errors.As(err, &appErr)
	require.True(t, ok)
	require.Equal(t, 409, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}
