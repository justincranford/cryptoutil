// Copyright (c) 2025-2026 Justin Cranford.

package client

import (
	http "net/http"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testAuthLoginPath    = "/login"
	testAuthBaseURL      = "https://localhost"
	testAuthMarshalError = "failed to marshal request"
)

func TestUserAuthenticationHelpers_InjectedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		run     func() error
		wantErr string
	}{
		{
			name: "registration request marshal failure",
			run: func() error {
				_, err := registerUser(&http.Client{}, testAuthBaseURL, cryptoutilSharedMagic.PathRegistration, testAuthLoginPath, testUsername, testUsername, func(any) ([]byte, error) {
					return nil, errTestGenerateFailure
				})

				return err
			},
			wantErr: testAuthMarshalError,
		},
		{
			name: "login request marshal failure",
			run: func() error {
				_, err := loginUser(&http.Client{}, testAuthBaseURL, testAuthLoginPath, testUsername, testUsername, func(any) ([]byte, error) {
					return nil, errTestGenerateFailure
				})

				return err
			},
			wantErr: testAuthMarshalError,
		},
		{
			name: "test username generation failure",
			run: func() error {
				_, err := registerTestUser(&http.Client{}, testAuthBaseURL, cryptoutilSharedMagic.PathRegistration, testAuthLoginPath, func() (string, error) {
					return "", errTestGenerateFailure
				}, func() (string, error) {
					return testUsername, nil
				})

				return err
			},
			wantErr: "failed to generate username",
		},
		{
			name: "test password generation failure",
			run: func() error {
				_, err := registerTestUser(&http.Client{}, testAuthBaseURL, cryptoutilSharedMagic.PathRegistration, testAuthLoginPath, func() (string, error) {
					return testUsername, nil
				}, func() (string, error) {
					return "", errTestGenerateFailure
				})

				return err
			},
			wantErr: "failed to generate password",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorContains(t, testCase.run(), testCase.wantErr)
		})
	}
}
