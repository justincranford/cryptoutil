// Copyright (c) 2025 Justin Cranford
//
//

package domain_test

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"cryptoutil/internal/apps/cipher/im/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

func TestMessage_TableName(t *testing.T) {
	t.Parallel()

	m := domain.Message{}
	require.Equal(t, "messages", m.TableName())
}

func TestMessage_FieldTypes(t *testing.T) {
	t.Parallel()

	senderID := googleUuid.New()
	messageID := googleUuid.New()
	createdAt := time.Now()
	readAt := time.Now().Add(10 * time.Minute)
	jwe := `{"protected":"eyJhbGciOiJBMjU2R0NNS1ciLCJlbmMiOiJBMjU2R0NNIn0","iv":"test","ciphertext":"test","tag":"test"}`

	m := domain.Message{
		ID:        messageID,
		SenderID:  senderID,
		JWE:       jwe,
		CreatedAt: createdAt,
		ReadAt:    &readAt,
		Sender: cryptoutilAppsTemplateServiceServerRepository.User{
			ID:       senderID,
			Username: "testuser",
		},
	}

	require.Equal(t, messageID, m.ID)
	require.Equal(t, senderID, m.SenderID)
	require.Equal(t, jwe, m.JWE)
	require.Equal(t, createdAt.Unix(), m.CreatedAt.Unix())
	require.NotNil(t, m.ReadAt)
	require.Equal(t, readAt.Unix(), m.ReadAt.Unix())
	require.Equal(t, senderID, m.Sender.ID)
	require.Equal(t, "testuser", m.Sender.Username)
}

func TestMessage_NilReadAt(t *testing.T) {
	t.Parallel()

	m := domain.Message{
		ID:        googleUuid.New(),
		SenderID:  googleUuid.New(),
		JWE:       "test-jwe",
		CreatedAt: time.Now(),
		ReadAt:    nil, // Message not read yet
	}

	require.Nil(t, m.ReadAt)
}

func TestMessage_ZeroValue(t *testing.T) {
	t.Parallel()

	var m domain.Message

	require.Equal(t, googleUuid.Nil, m.ID)
	require.Equal(t, googleUuid.Nil, m.SenderID)
	require.Empty(t, m.JWE)
	require.True(t, m.CreatedAt.IsZero())
	require.Nil(t, m.ReadAt)
	require.Equal(t, googleUuid.Nil, m.Sender.ID)
}
