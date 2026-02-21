// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestValidateRegistrationRequest_Table tests all validation branches in
// validateRegistrationRequest, covering lines 121-140.
func TestValidateRegistrationRequest_Table(t *testing.T) {
	t.Parallel()

	validUsername := strings.Repeat("a", cryptoutilSharedMagic.CipherMinUsernameLength)
	validEmail := "user@example.com"
	validPassword := strings.Repeat("p", cryptoutilSharedMagic.CipherMinPasswordLength)
	validTenantName := strings.Repeat("t", cryptoutilSharedMagic.CipherMinUsernameLength)

	tests := []struct {
		name       string
		body        RegisterUserRequest
		wantErrMsg string
	}{
		{
			name: "username too short",
			body: RegisterUserRequest{
				Username:   strings.Repeat("a", cryptoutilSharedMagic.CipherMinUsernameLength-1),
				Email:      validEmail,
				Password:   validPassword,
				TenantName: validTenantName,
			},
			wantErrMsg: "username must be at least",
		},
		{
			name: "username too long",
			body: RegisterUserRequest{
				Username:   strings.Repeat("a", cryptoutilSharedMagic.CipherMaxUsernameLength+1),
				Email:      validEmail,
				Password:   validPassword,
				TenantName: validTenantName,
			},
			wantErrMsg: "username must be at most",
		},
		{
			name: "invalid email",
			body: RegisterUserRequest{
				Username:   validUsername,
				Email:      "not-an-email",
				Password:   validPassword,
				TenantName: validTenantName,
			},
			wantErrMsg: "invalid email format",
		},
		{
			name: "password too short",
			body: RegisterUserRequest{
				Username:   validUsername,
				Email:      validEmail,
				Password:   strings.Repeat("p", cryptoutilSharedMagic.CipherMinPasswordLength-1),
				TenantName: validTenantName,
			},
			wantErrMsg: "password must be at least",
		},
		{
			name: "tenant name too short",
			body: RegisterUserRequest{
				Username:   validUsername,
				Email:      validEmail,
				Password:   validPassword,
				TenantName: strings.Repeat("t", cryptoutilSharedMagic.CipherMinUsernameLength-1),
			},
			wantErrMsg: "tenant name must be at least",
		},
		{
			name: "tenant name too long",
			body: RegisterUserRequest{
				Username:   validUsername,
				Email:      validEmail,
				Password:   validPassword,
				TenantName: strings.Repeat("t", cryptoutilSharedMagic.CipherMaxTenantNameLength+1),
			},
			wantErrMsg: "tenant name must be at most",
		},
		{
			name: "all valid",
			body: RegisterUserRequest{
				Username:   validUsername,
				Email:      validEmail,
				Password:   validPassword,
				TenantName: validTenantName,
			},
			wantErrMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body := tc.body
			err := validateRegistrationRequest(&body)

			if tc.wantErrMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrMsg)
			}
		})
	}
}
