// Copyright (c) 2025-2026 Justin Cranford.
//

package handler

import (
	json "encoding/json"
	http "net/http"
	"testing"

	cryptoutilAppsFrameworkTestDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilAppsSmKmsRepository "cryptoutil/internal/apps/sm-kms/server/repository"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newMessageHandlerForTest(t *testing.T, db *gorm.DB) *MessageHandler {
	t.Helper()

	return NewMessageHandler(
		cryptoutilAppsSmKmsRepository.NewMessageRepository(db),
		cryptoutilAppsSmKmsRepository.NewMessageRecipientJWKRepository(db, nil),
	)
}

func migrateMessageTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	require.NoError(t, db.AutoMigrate(&cryptoutilAppsSmKmsModel.Message{}, &cryptoutilAppsSmKmsModel.MessageRecipientJWK{}))
}

func TestMessageHandler_ListMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		invalidID  bool
		migrate    bool
		seed       bool
		wantStatus int
		wantCount  int
	}{
		{name: "invalid recipient id", invalidID: true, wantStatus: fiber.StatusBadRequest},
		{name: "repository query failure", wantStatus: fiber.StatusInternalServerError},
		{name: "empty result", migrate: true, wantStatus: fiber.StatusOK},
		{name: "matching message", migrate: true, seed: true, wantStatus: fiber.StatusOK, wantCount: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
			if tc.migrate {
				migrateMessageTables(t, db)
			}

			recipientID := newCompatTestUUID(t)
			if tc.seed {
				messageID := newCompatTestUUID(t)
				require.NoError(t, db.Create(&cryptoutilAppsSmKmsModel.Message{ID: messageID, SenderID: newCompatTestUUID(t), JWE: testCompatMessageJWE}).Error)
				require.NoError(t, db.Create(&cryptoutilAppsSmKmsModel.MessageRecipientJWK{ID: newCompatTestUUID(t), RecipientID: recipientID, MessageID: messageID, EncryptedJWK: testCompatMessageJWE}).Error)
			}

			handler := newMessageHandlerForTest(t, db)

			query := recipientID.String()
			if tc.invalidID {
				query = testCompatInvalidUUID
			}

			path := testCompatMessagePath + "?" + testCompatRecipientQuery + "=" + query
			status, body := runCompatFiberRequest(t, func(app *fiber.App) {
				app.Get(testCompatMessagePath, handler.HandleListMessages())
			}, http.MethodGet, path)

			require.Equal(t, tc.wantStatus, status)

			if tc.wantStatus == fiber.StatusOK {
				var messages []cryptoutilAppsSmKmsModel.Message
				require.NoError(t, json.Unmarshal(body, &messages))
				require.Len(t, messages, tc.wantCount)
			}
		})
	}
}

func TestMessageHandler_GetMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		invalidID  bool
		seed       bool
		wantStatus int
	}{
		{name: testCompatInvalidMessageIDCase, invalidID: true, wantStatus: fiber.StatusBadRequest},
		{name: "message missing", wantStatus: fiber.StatusNotFound},
		{name: "message found", seed: true, wantStatus: fiber.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
			require.NoError(t, db.AutoMigrate(&cryptoutilAppsSmKmsModel.Message{}))

			id := newCompatTestUUID(t)
			if tc.seed {
				require.NoError(t, db.Create(&cryptoutilAppsSmKmsModel.Message{ID: id, SenderID: newCompatTestUUID(t), JWE: testCompatMessageJWE}).Error)
			}

			pathID := id.String()
			if tc.invalidID {
				pathID = testCompatInvalidUUID
			}

			handler := newMessageHandlerForTest(t, db)
			path := testCompatMessagePath + "/" + pathID
			status, body := runCompatFiberRequest(t, func(app *fiber.App) {
				app.Get(testCompatMessagePath+"/"+testCompatMessageIDParam, handler.HandleGetMessage())
			}, http.MethodGet, path)

			require.Equal(t, tc.wantStatus, status)

			if tc.wantStatus == fiber.StatusOK {
				require.Contains(t, string(body), id.String())
			}
		})
	}
}

func TestMessageHandler_DeleteMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		invalidID            bool
		migrateMessages      bool
		migrateRecipientJWKs bool
		dropMessagesTable    bool
		seed                 bool
		wantStatus           int
	}{
		{name: testCompatInvalidMessageIDCase, invalidID: true, wantStatus: fiber.StatusBadRequest},
		{name: "recipient key deletion failure", migrateMessages: true, wantStatus: fiber.StatusInternalServerError},
		{name: "message deletion failure", migrateMessages: true, migrateRecipientJWKs: true, dropMessagesTable: true, wantStatus: fiber.StatusInternalServerError},
		{name: "message and recipient keys deleted", migrateMessages: true, migrateRecipientJWKs: true, seed: true, wantStatus: fiber.StatusNoContent},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
			if tc.migrateMessages {
				require.NoError(t, db.AutoMigrate(&cryptoutilAppsSmKmsModel.Message{}))
			}

			if tc.migrateRecipientJWKs {
				require.NoError(t, db.AutoMigrate(&cryptoutilAppsSmKmsModel.MessageRecipientJWK{}))
			}

			id := newCompatTestUUID(t)
			if tc.seed {
				recipientID := newCompatTestUUID(t)
				require.NoError(t, db.Create(&cryptoutilAppsSmKmsModel.Message{ID: id, SenderID: newCompatTestUUID(t), JWE: testCompatMessageJWE}).Error)
				require.NoError(t, db.Create(&cryptoutilAppsSmKmsModel.MessageRecipientJWK{ID: newCompatTestUUID(t), RecipientID: recipientID, MessageID: id, EncryptedJWK: testCompatMessageJWE}).Error)
			}

			if tc.dropMessagesTable {
				require.NoError(t, db.Migrator().DropTable(&cryptoutilAppsSmKmsModel.Message{}))
			}

			pathID := id.String()
			if tc.invalidID {
				pathID = testCompatInvalidUUID
			}

			handler := newMessageHandlerForTest(t, db)
			path := testCompatMessagePath + "/" + pathID
			status, _ := runCompatFiberRequest(t, func(app *fiber.App) {
				app.Delete(testCompatMessagePath+"/"+testCompatMessageIDParam, handler.HandleDeleteMessage())
			}, http.MethodDelete, path)

			require.Equal(t, tc.wantStatus, status)

			if tc.wantStatus == fiber.StatusNoContent {
				var (
					messageCount   int64
					recipientCount int64
				)

				require.NoError(t, db.Model(&cryptoutilAppsSmKmsModel.Message{}).Where("id = ?", id).Count(&messageCount).Error)
				require.NoError(t, db.Model(&cryptoutilAppsSmKmsModel.MessageRecipientJWK{}).Where("message_id = ?", id).Count(&recipientCount).Error)
				require.Zero(t, messageCount)
				require.Zero(t, recipientCount)
			}
		})
	}
}

func TestMessageHandler_NotImplementedRoutes(t *testing.T) {
	t.Parallel()

	handler := NewMessageHandler(nil, nil)
	tests := []struct {
		name   string
		method string
		path   string
		add    func(*fiber.App, *MessageHandler)
	}{
		{name: "send message", method: http.MethodPost, path: testCompatMessageSendPath, add: func(app *fiber.App, handler *MessageHandler) {
			app.Post(testCompatMessageSendPath, handler.HandleSendMessage())
		}},
		{name: "receive messages", method: http.MethodGet, path: testCompatMessageReceivePath, add: func(app *fiber.App, handler *MessageHandler) {
			app.Get(testCompatMessageReceivePath, handler.HandleReceiveMessages())
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status, _ := runCompatFiberRequest(t, func(app *fiber.App) { tc.add(app, handler) }, tc.method, tc.path)
			require.Equal(t, fiber.StatusNotImplemented, status)
		})
	}
}
