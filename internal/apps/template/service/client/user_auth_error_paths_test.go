// Copyright (c) 2025 Justin Cranford

// Package client provides error path tests using injectable vars.
package client

import (
"errors"
"testing"

"github.com/stretchr/testify/require"
)

// errTestGenerateFailure is used to inject into the generate functions.
var errTestGenerateFailure = errors.New("test: generate failure")

const testUsername = "user"

// TestRegisterTestUserService_UsernameError tests RegisterTestUserService
// when username generation fails.
// NOTE: Must NOT use t.Parallel() - modifies package-level var.
func TestRegisterTestUserService_UsernameError(t *testing.T) {
orig := templateClientGenerateUsernameSimpleFn
templateClientGenerateUsernameSimpleFn = func() (string, error) {
return "", errTestGenerateFailure
}

defer func() { templateClientGenerateUsernameSimpleFn = orig }()

_, err := RegisterTestUserService(nil, "https://localhost")
require.Error(t, err)
require.Contains(t, err.Error(), "failed to generate username")
}

// TestRegisterTestUserService_PasswordError tests RegisterTestUserService
// when password generation fails.
// NOTE: Must NOT use t.Parallel() - modifies package-level var.
func TestRegisterTestUserService_PasswordError(t *testing.T) {
origUsername := templateClientGenerateUsernameSimpleFn
origPassword := templateClientGeneratePasswordSimpleFn

templateClientGenerateUsernameSimpleFn = func() (string, error) { return testUsername, nil }
templateClientGeneratePasswordSimpleFn = func() (string, error) { return "", errTestGenerateFailure }

defer func() {
templateClientGenerateUsernameSimpleFn = origUsername
templateClientGeneratePasswordSimpleFn = origPassword
}()

_, err := RegisterTestUserService(nil, "https://localhost")
require.Error(t, err)
require.Contains(t, err.Error(), "failed to generate password")
}

// TestRegisterTestUserBrowser_UsernameError tests RegisterTestUserBrowser
// when username generation fails.
// NOTE: Must NOT use t.Parallel() - modifies package-level var.
func TestRegisterTestUserBrowser_UsernameError(t *testing.T) {
orig := templateClientGenerateUsernameSimpleFn
templateClientGenerateUsernameSimpleFn = func() (string, error) {
return "", errTestGenerateFailure
}

defer func() { templateClientGenerateUsernameSimpleFn = orig }()

_, err := RegisterTestUserBrowser(nil, "https://localhost")
require.Error(t, err)
require.Contains(t, err.Error(), "failed to generate username")
}

// TestRegisterTestUserBrowser_PasswordError tests RegisterTestUserBrowser
// when password generation fails.
// NOTE: Must NOT use t.Parallel() - modifies package-level var.
func TestRegisterTestUserBrowser_PasswordError(t *testing.T) {
origUsername := templateClientGenerateUsernameSimpleFn
origPassword := templateClientGeneratePasswordSimpleFn

templateClientGenerateUsernameSimpleFn = func() (string, error) { return testUsername, nil }
templateClientGeneratePasswordSimpleFn = func() (string, error) { return "", errTestGenerateFailure }

defer func() {
templateClientGenerateUsernameSimpleFn = origUsername
templateClientGeneratePasswordSimpleFn = origPassword
}()

_, err := RegisterTestUserBrowser(nil, "https://localhost")
require.Error(t, err)
require.Contains(t, err.Error(), "failed to generate password")
}
