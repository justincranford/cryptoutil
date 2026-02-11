// Copyright (c) 2025 Justin Cranford
//
//

package domain_test

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsCipherImDomain "cryptoutil/internal/apps/cipher/im/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

func TestMessageRecipientJWK_TableName(t *testing.T) {
	t.Parallel()

	mrj := cryptoutilAppsCipherImDomain.MessageRecipientJWK{}
	require.Equal(t, "messages_recipient_jwks", mrj.TableName())
}

func TestMessageRecipientJWK_FieldTypes(t *testing.T) {
	t.Parallel()

	id := googleUuid.New()
	recipientID := googleUuid.New()
	messageID := googleUuid.New()
	encryptedJWK := `{"protected":"eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIn0","iv":"test","ciphertext":"test","tag":"test"}`
	createdAt := time.Now().UTC()

	mrj := cryptoutilAppsCipherImDomain.MessageRecipientJWK{
		ID:           id,
		RecipientID:  recipientID,
		MessageID:    messageID,
		EncryptedJWK: encryptedJWK,
		CreatedAt:    createdAt,
		Recipient: cryptoutilAppsTemplateServiceServerRepository.User{
			ID:       recipientID,
			Username: "recipient",
		},
		Message: cryptoutilAppsCipherImDomain.Message{
			ID:       messageID,
			SenderID: googleUuid.New(),
			JWE:      "test-message-jwe",
		},
	}

	require.Equal(t, id, mrj.ID)
	require.Equal(t, recipientID, mrj.RecipientID)
	require.Equal(t, messageID, mrj.MessageID)
	require.Equal(t, encryptedJWK, mrj.EncryptedJWK)
	require.Equal(t, createdAt.Unix(), mrj.CreatedAt.Unix())
	require.Equal(t, recipientID, mrj.Recipient.ID)
	require.Equal(t, "recipient", mrj.Recipient.Username)
	require.Equal(t, messageID, mrj.Message.ID)
	require.Equal(t, "test-message-jwe", mrj.Message.JWE)
}

func TestMessageRecipientJWK_ZeroValue(t *testing.T) {
	t.Parallel()

	var mrj cryptoutilAppsCipherImDomain.MessageRecipientJWK

	require.Equal(t, googleUuid.Nil, mrj.ID)
	require.Equal(t, googleUuid.Nil, mrj.RecipientID)
	require.Equal(t, googleUuid.Nil, mrj.MessageID)
	require.Empty(t, mrj.EncryptedJWK)
	require.True(t, mrj.CreatedAt.IsZero())
	require.Equal(t, googleUuid.Nil, mrj.Recipient.ID)
	require.Equal(t, googleUuid.Nil, mrj.Message.ID)
}

func TestMessageRecipientJWK_MultiRecipientScenario(t *testing.T) {
	t.Parallel()

	// Simulate 3 recipients for same message
	messageID := googleUuid.New()
	recipients := []googleUuid.UUID{
		googleUuid.New(),
		googleUuid.New(),
		googleUuid.New(),
	}

	jwks := make([]cryptoutilAppsCipherImDomain.MessageRecipientJWK, len(recipients))
	for i, recipID := range recipients {
		jwks[i] = cryptoutilAppsCipherImDomain.MessageRecipientJWK{
			ID:           googleUuid.New(),
			RecipientID:  recipID,
			MessageID:    messageID,
			EncryptedJWK: `{"protected":"test","iv":"test","ciphertext":"test","tag":"test"}`,
			CreatedAt:    time.Now().UTC(),
		}
	}

	// Verify all JWKs reference same message
	for _, jwk := range jwks {
		require.Equal(t, messageID, jwk.MessageID)
		require.NotEqual(t, googleUuid.Nil, jwk.RecipientID)
		require.NotEqual(t, googleUuid.Nil, jwk.ID)
	}

	// Verify unique recipient IDs
	for i := 0; i < len(jwks); i++ {
		for j := i + 1; j < len(jwks); j++ {
			require.NotEqual(t, jwks[i].RecipientID, jwks[j].RecipientID)
			require.NotEqual(t, jwks[i].ID, jwks[j].ID)
		}
	}
}
