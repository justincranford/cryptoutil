// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestKeyRotationEvent_BeforeCreate(t *testing.T) {
	t.Parallel()

	event := &KeyRotationEvent{
		EventType: "rotation",
		KeyType:   cryptoutilSharedMagic.ParamClientSecret,
		KeyID:     cryptoutilSharedMagic.TestClientID,
		Initiator: "test-user",
	}

	err := event.BeforeCreate(nil)
	require.NoError(t, err, "BeforeCreate should not return error")
	require.NotEqual(t, googleUuid.Nil, event.ID, "ID should be generated")
	require.False(t, event.Timestamp.IsZero(), "Timestamp should be set")
}

func TestKeyRotationEvent_TableName(t *testing.T) {
	t.Parallel()

	event := &KeyRotationEvent{}

	tableName := event.TableName()
	require.Equal(t, "key_rotation_events", tableName, "TableName should match expected")
}

func TestKeyRotationEvent_Constants(t *testing.T) {
	t.Parallel()

	// Event type constants.
	require.Equal(t, "rotation", EventTypeRotation)
	require.Equal(t, "revocation", EventTypeRevocation)
	require.Equal(t, "expiration", EventTypeExpiration)

	// Key type constants.
	require.Equal(t, cryptoutilSharedMagic.ParamClientSecret, KeyTypeClientSecret)
	require.Equal(t, "jwk", KeyTypeJWK)
	require.Equal(t, "api_key", KeyTypeAPIKey)
}

func TestKeyRotationEvent_FieldValidation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	oldVersion := 1
	newVersion := 2
	gracePeriod := "24h"

	event := &KeyRotationEvent{
		ID:            googleUuid.Must(googleUuid.NewV7()),
		EventType:     "rotation",
		KeyType:       cryptoutilSharedMagic.ParamClientSecret,
		KeyID:         cryptoutilSharedMagic.TestClientID,
		Timestamp:     now,
		Initiator:     "test-user",
		OldKeyVersion: &oldVersion,
		NewKeyVersion: &newVersion,
		GracePeriod:   &gracePeriod,
		Reason:        "scheduled",
		Metadata:      `{"test": "data"}`,
		Success:       boolPtr(true),
		ErrorMessage:  "",
	}

	require.Equal(t, "rotation", event.EventType)
	require.Equal(t, cryptoutilSharedMagic.ParamClientSecret, event.KeyType)
	require.Equal(t, cryptoutilSharedMagic.TestClientID, event.KeyID)
	require.Equal(t, now, event.Timestamp)
	require.Equal(t, "test-user", event.Initiator)
	require.NotNil(t, event.OldKeyVersion)
	require.Equal(t, 1, *event.OldKeyVersion)
	require.NotNil(t, event.NewKeyVersion)
	require.Equal(t, 2, *event.NewKeyVersion)
	require.NotNil(t, event.GracePeriod)
	require.Equal(t, "24h", *event.GracePeriod)
	require.Equal(t, "scheduled", event.Reason)
	require.Equal(t, `{"test": "data"}`, event.Metadata)
	require.NotNil(t, event.Success)
	require.True(t, *event.Success)
	require.Empty(t, event.ErrorMessage)
}
