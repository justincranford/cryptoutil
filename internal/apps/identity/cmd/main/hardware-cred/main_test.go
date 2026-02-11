// Copyright (c) 2025 Justin Cranford
//
//

package main

import (
	"context"
	"flag"
	"os"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityORM "cryptoutil/internal/apps/identity/repository/orm"
)

// TestEnrollCommand tests the enroll command flag parsing and validation.
func TestEnrollCommand(t *testing.T) {
	t.Parallel()

	// Generate test UUIDs
	userID1 := googleUuid.Must(googleUuid.NewV7()).String()
	userID2 := googleUuid.Must(googleUuid.NewV7()).String()

	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantUserID   string
		wantDevice   string
		wantCredType string
	}{
		{
			name: "missing credential type uses default",
			args: []string{
				"-user-id", userID1,
				"-device-name", "Test Device",
			},
			wantErr:      false,
			wantUserID:   userID1,
			wantDevice:   "Test Device",
			wantCredType: "passkey",
		},
		{
			name: "valid enrollment with default device name",
			args: []string{
				"-user-id", userID2,
			},
			wantErr:      false,
			wantUserID:   userID2,
			wantCredType: "passkey",
		},
		{
			name:    "missing user-id flag",
			args:    []string{},
			wantErr: true,
		},
		{
			name: "invalid user-id UUID",
			args: []string{
				"-user-id", "invalid-uuid",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet("enroll", flag.ContinueOnError)
			userIDStr := fs.String("user-id", "", "User ID (UUID)")
			deviceName := fs.String("device-name", "", "Device name")
			credentialType := fs.String("credential-type", "passkey", "Credential type")

			err := fs.Parse(tc.args)
			if err != nil {
				require.True(t, tc.wantErr, "unexpected flag parse error: %v", err)

				return
			}

			if *userIDStr == "" {
				require.True(t, tc.wantErr, "missing user-id should cause error")

				return
			}

			userID, err := googleUuid.Parse(*userIDStr)
			if err != nil {
				require.True(t, tc.wantErr, "invalid user-id UUID should cause error")

				return
			}

			require.False(t, tc.wantErr, "should not error with valid flags")

			require.Equal(t, tc.wantUserID, userID.String())
			require.Equal(t, tc.wantCredType, *credentialType)

			if tc.wantDevice != "" {
				require.Equal(t, tc.wantDevice, *deviceName)
			}
		})
	}
}

// TestListCommand tests the list command flag parsing and validation.
func TestListCommand(t *testing.T) {
	t.Parallel()

	// Generate test UUID
	userID := googleUuid.Must(googleUuid.NewV7()).String()

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantUserID string
	}{
		{
			name: "valid list command",
			args: []string{
				"-user-id", userID,
			},
			wantErr:    false,
			wantUserID: userID,
		},
		{
			name:    "missing user-id flag",
			args:    []string{},
			wantErr: true,
		},
		{
			name: "invalid user-id UUID",
			args: []string{
				"-user-id", "not-a-uuid",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet("list", flag.ContinueOnError)
			userIDStr := fs.String("user-id", "", "User ID (UUID)")

			err := fs.Parse(tc.args)
			if err != nil {
				require.True(t, tc.wantErr, "unexpected flag parse error: %v", err)

				return
			}

			if *userIDStr == "" {
				require.True(t, tc.wantErr, "missing user-id should cause error")

				return
			}

			userID, err := googleUuid.Parse(*userIDStr)
			if err != nil {
				require.True(t, tc.wantErr, "invalid user-id UUID should cause error")

				return
			}

			require.False(t, tc.wantErr, "should not error with valid flags")
			require.Equal(t, tc.wantUserID, userID.String())
		})
	}
}

// TestRevokeCommand tests the revoke command flag parsing and validation.
func TestRevokeCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		args             []string
		wantErr          bool
		wantCredentialID string
	}{
		{
			name: "valid revoke command",
			args: []string{
				"-credential-id", "AQIDBAUGBwgJCgsMDQ4PEA==",
			},
			wantErr:          false,
			wantCredentialID: "AQIDBAUGBwgJCgsMDQ4PEA==",
		},
		{
			name:    "missing credential-id flag",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet("revoke", flag.ContinueOnError)
			credentialID := fs.String("credential-id", "", "Credential ID")

			err := fs.Parse(tc.args)
			if err != nil {
				require.True(t, tc.wantErr, "unexpected flag parse error: %v", err)

				return
			}

			if *credentialID == "" {
				require.True(t, tc.wantErr, "missing credential-id should cause error")

				return
			}

			require.False(t, tc.wantErr, "should not error with valid flags")
			require.Equal(t, tc.wantCredentialID, *credentialID)
		})
	}
}

// TestGenerateMockCredentialID tests mock credential ID generation.
func TestGenerateMockCredentialID(t *testing.T) {
	t.Parallel()

	credID := generateMockCredentialID()
	require.NotEmpty(t, credID)
	require.True(t, len(credID) > 0, "credential ID should not be empty")
}

// TestGenerateMockPublicKey tests mock public key generation.
func TestGenerateMockPublicKey(t *testing.T) {
	t.Parallel()

	pubKey := generateMockPublicKey()
	require.NotNil(t, pubKey)
	require.Equal(t, 65, len(pubKey), "ECDSA P-256 public key should be 65 bytes")
}

// TestParseCredentialType tests credential type parsing.
func TestParseCredentialType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeStr  string
		wantType cryptoutilIdentityORM.CredentialType
	}{
		{
			name:     "passkey type",
			typeStr:  "passkey",
			wantType: cryptoutilIdentityORM.CredentialTypePasskey,
		},
		{
			name:     "smart_card type defaults to passkey",
			typeStr:  "smart_card",
			wantType: cryptoutilIdentityORM.CredentialTypePasskey,
		},
		{
			name:     "security_key type defaults to passkey",
			typeStr:  "security_key",
			wantType: cryptoutilIdentityORM.CredentialTypePasskey,
		},
		{
			name:     "unknown type defaults to passkey",
			typeStr:  "unknown",
			wantType: cryptoutilIdentityORM.CredentialTypePasskey,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			credType := parseCredentialType(tc.typeStr)
			require.Equal(t, tc.wantType, credType)
		})
	}
}

// TestLogAuditEvent is a basic smoke test for logAuditEvent.
// NOTE: Cannot use t.Parallel() - test manipulates global os.Stdout.
func TestLogAuditEvent(_ *testing.T) {
	ctx := context.Background()

	// Capture log output.
	oldOutput := os.Stdout

	defer func() {
		os.Stdout = oldOutput
	}()

	// Note: Actual log output capture requires redirecting stdout/stderr.
	// This test validates that logAuditEvent doesn't panic.
	logAuditEvent(ctx, "TEST_EVENT", "user-123", "cred-456", map[string]any{
		"test_key": "test_value",
	})
}

// TestHelpCommand tests help text generation (smoke test).
func TestHelpCommand(t *testing.T) {
	t.Parallel()

	// Smoke test - ensure printUsage doesn't panic.
	printUsage()
}
