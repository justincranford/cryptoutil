// Copyright (c) 2025 Justin Cranford

// Package client provides error path tests using injectable fns.
package client

import (
"errors"
"testing"

"github.com/stretchr/testify/require"

cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

// errTestGenerateFailure is used to inject into the generate functions.
var errTestGenerateFailure = errors.New("test: generate failure")

const testUsername = "user"

// TestGenerateCredentials_UsernameError verifies username generation failure is propagated.
// Sequential: calls generateCredentials with injected stub (no package-level state mutation).
func TestGenerateCredentials_UsernameError(t *testing.T) {
t.Parallel()

_, _, err := generateCredentials(
func() (string, error) { return "", errTestGenerateFailure },
cryptoutilSharedUtilRandom.GeneratePasswordSimple,
)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to generate username")
}

// TestGenerateCredentials_PasswordError verifies password generation failure is propagated.
func TestGenerateCredentials_PasswordError(t *testing.T) {
t.Parallel()

_, _, err := generateCredentials(
func() (string, error) { return testUsername, nil },
func() (string, error) { return "", errTestGenerateFailure },
)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to generate password")
}
