// Copyright (c) 2025-2026 Justin Cranford.

package repository

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	cryptoutilAppsFrameworkServiceTestHelpDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const testOperationSign = "sign"

func newRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := cryptoutilAppsFrameworkServiceTestHelpDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(
		&cryptoutilAppsSmKmsModel.AuditConfig{},
		&cryptoutilAppsSmKmsModel.AuditLogEntry{},
		&cryptoutilAppsSmKmsModel.ElasticJWK{},
		&cryptoutilAppsSmKmsModel.MaterialJWK{},
		&cryptoutilAppsSmKmsModel.Message{},
		&cryptoutilAppsSmKmsModel.MessageRecipientJWK{},
	))

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	return db
}

func newRepositoryTestElasticJWK(tenantID googleUuid.UUID) *cryptoutilAppsSmKmsModel.ElasticJWK {
	id := googleUuid.Must(googleUuid.NewV7())

	return &cryptoutilAppsSmKmsModel.ElasticJWK{
		ID:                   id,
		TenantID:             tenantID,
		KID:                  id.String(),
		KeyType:              cryptoutilAppsSmKmsModel.KeyTypeRSA,
		Algorithm:            cryptoutilSharedMagic.JoseAlgRS256,
		Use:                  cryptoutilAppsSmKmsModel.KeyUseSig,
		MaxMaterials:         10,
		CurrentMaterialCount: 1,
		CreatedAt:            time.Now().UTC(),
	}
}

func newRepositoryTestMaterial(elasticJWKID googleUuid.UUID, active bool) *cryptoutilAppsSmKmsModel.MaterialJWK {
	id := googleUuid.Must(googleUuid.NewV7())

	return &cryptoutilAppsSmKmsModel.MaterialJWK{
		ID:             id,
		ElasticJWKID:   elasticJWKID,
		MaterialKID:    id.String(),
		PrivateJWKJWE:  base64.StdEncoding.EncodeToString([]byte("private")),
		PublicJWKJWE:   base64.StdEncoding.EncodeToString([]byte(cryptoutilSharedMagic.PublicLabel)),
		Active:         active,
		CreatedAt:      time.Now().UTC(),
		BarrierVersion: 1,
	}
}

func TestRepository_GormCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := newRepositoryTestDB(t)
	tenantID := googleUuid.Must(googleUuid.NewV7())
	elasticJWK := newRepositoryTestElasticJWK(tenantID)
	require.NoError(t, db.Create(elasticJWK).Error)

	t.Run("audit configuration", func(t *testing.T) {
		t.Parallel()

		repo := NewAuditConfigRepository(db)
		config := &cryptoutilAppsSmKmsModel.AuditConfig{
			TenantID: tenantID, Operation: cryptoutilAppsSmKmsModel.OperationSign, Enabled: true, SamplingRate: 1,
		}
		require.NoError(t, repo.Upsert(ctx, config))

		got, err := repo.Get(ctx, tenantID, config.Operation)
		require.NoError(t, err)
		require.Equal(t, config.SamplingRate, got.SamplingRate)

		configs, err := repo.GetAllForTenant(ctx, tenantID)
		require.NoError(t, err)
		require.Len(t, configs, 1)

		shouldAudit, err := repo.ShouldAudit(ctx, tenantID, config.Operation)
		require.NoError(t, err)
		require.True(t, shouldAudit)

		config.Enabled = false
		require.NoError(t, repo.Upsert(ctx, config))
		shouldAudit, err = repo.ShouldAudit(ctx, tenantID, config.Operation)
		require.NoError(t, err)
		require.False(t, shouldAudit)

		require.NoError(t, repo.Delete(ctx, tenantID, config.Operation))
		_, err = repo.Get(ctx, tenantID, config.Operation)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("audit log", func(t *testing.T) {
		t.Parallel()

		repo := NewAuditLogRepository(db)
		elasticID := elasticJWK.ID
		entry := &cryptoutilAppsSmKmsModel.AuditLogEntry{
			ID:           googleUuid.Must(googleUuid.NewV7()),
			TenantID:     tenantID,
			ElasticJWKID: &elasticID,
			Operation:    cryptoutilAppsSmKmsModel.OperationSign,
			Success:      true,
			RequestID:    googleUuid.Must(googleUuid.NewV7()).String(),
			CreatedAt:    time.Now().UTC(),
		}
		require.NoError(t, repo.Create(ctx, entry))

		entries, total, err := repo.List(ctx, tenantID, 0, 10)
		require.NoError(t, err)
		require.EqualValues(t, 1, total)
		require.Len(t, entries, 1)

		entries, total, err = repo.ListByElasticJWK(ctx, elasticJWK.ID, 0, 10)
		require.NoError(t, err)
		require.EqualValues(t, 1, total)
		require.Len(t, entries, 1)

		entries, total, err = repo.ListByOperation(ctx, tenantID, entry.Operation, 0, 10)
		require.NoError(t, err)
		require.EqualValues(t, 1, total)
		require.Len(t, entries, 1)

		got, err := repo.GetByRequestID(ctx, entry.RequestID)
		require.NoError(t, err)
		require.Equal(t, entry.ID, got.ID)

		oldEntry := &cryptoutilAppsSmKmsModel.AuditLogEntry{
			ID:        googleUuid.Must(googleUuid.NewV7()),
			TenantID:  tenantID,
			Operation: cryptoutilAppsSmKmsModel.OperationGenerate,
			Success:   false,
			RequestID: googleUuid.Must(googleUuid.NewV7()).String(),
			CreatedAt: time.Now().UTC().AddDate(-2, 0, 0),
		}
		require.NoError(t, repo.Create(ctx, oldEntry))

		deleted, err := repo.DeleteOlderThan(ctx, tenantID, 365)
		require.NoError(t, err)
		require.EqualValues(t, 1, deleted)
	})

	t.Run("elastic JWK", func(t *testing.T) {
		t.Parallel()

		repo := NewElasticJWKRepository(db)
		item := newRepositoryTestElasticJWK(tenantID)
		require.NoError(t, repo.Create(ctx, item))

		got, err := repo.Get(ctx, tenantID, item.KID)
		require.NoError(t, err)
		require.Equal(t, item.ID, got.ID)

		got, err = repo.GetByID(ctx, item.ID)
		require.NoError(t, err)
		require.Equal(t, item.KID, got.KID)

		items, total, err := repo.List(ctx, tenantID, 0, 10)
		require.NoError(t, err)
		require.EqualValues(t, 2, total)
		require.Len(t, items, 2)

		item.MaxMaterials = 20
		require.NoError(t, repo.Update(ctx, item))
		require.NoError(t, repo.IncrementMaterialCount(ctx, item.ID))
		require.NoError(t, repo.DecrementMaterialCount(ctx, item.ID))
		require.NoError(t, repo.Delete(ctx, item.ID))
		_, err = repo.GetByID(ctx, item.ID)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("material JWK", func(t *testing.T) {
		t.Parallel()

		repo := NewMaterialJWKRepository(db)
		active := newRepositoryTestMaterial(elasticJWK.ID, true)
		require.NoError(t, repo.Create(ctx, active))

		got, err := repo.GetByMaterialKID(ctx, active.MaterialKID)
		require.NoError(t, err)
		require.Equal(t, active.ID, got.ID)

		got, err = repo.GetByID(ctx, active.ID)
		require.NoError(t, err)
		require.Equal(t, active.MaterialKID, got.MaterialKID)

		got, err = repo.GetActiveMaterial(ctx, elasticJWK.ID)
		require.NoError(t, err)
		require.Equal(t, active.ID, got.ID)

		items, total, err := repo.ListByElasticJWK(ctx, elasticJWK.ID, 0, 10)
		require.NoError(t, err)
		require.EqualValues(t, 1, total)
		require.Len(t, items, 1)

		next := newRepositoryTestMaterial(elasticJWK.ID, false)
		require.NoError(t, repo.RotateMaterial(ctx, elasticJWK.ID, next))
		got, err = repo.GetActiveMaterial(ctx, elasticJWK.ID)
		require.NoError(t, err)
		require.Equal(t, next.ID, got.ID)

		retired, err := repo.GetByID(ctx, active.ID)
		require.NoError(t, err)
		require.NotNil(t, retired.RetiredAt)

		count, err := repo.CountMaterials(ctx, elasticJWK.ID)
		require.NoError(t, err)
		require.EqualValues(t, 2, count)

		require.NoError(t, repo.RetireMaterial(ctx, next.ID))
		require.NoError(t, repo.Delete(ctx, active.ID))
	})

	t.Run("messages", func(t *testing.T) {
		t.Parallel()

		messageRepo := NewMessageRepository(db)
		recipientRepo := NewMessageRecipientJWKRepository(db, nil)
		message := &cryptoutilAppsSmKmsModel.Message{
			ID:       googleUuid.Must(googleUuid.NewV7()),
			SenderID: googleUuid.Must(googleUuid.NewV7()),
			JWE:      "encrypted-message",
		}
		require.NoError(t, messageRepo.Create(ctx, message))

		recipient := &cryptoutilAppsSmKmsModel.MessageRecipientJWK{
			ID:           googleUuid.Must(googleUuid.NewV7()),
			RecipientID:  googleUuid.Must(googleUuid.NewV7()),
			MessageID:    message.ID,
			EncryptedJWK: "encrypted-recipient-key",
		}
		require.NoError(t, recipientRepo.Create(ctx, recipient))

		gotMessage, err := messageRepo.FindByID(ctx, message.ID)
		require.NoError(t, err)
		require.Equal(t, message.ID, gotMessage.ID)

		messages, err := messageRepo.FindByRecipientID(ctx, recipient.RecipientID)
		require.NoError(t, err)
		require.Len(t, messages, 1)

		gotRecipient, err := recipientRepo.FindByRecipientAndMessage(ctx, recipient.RecipientID, message.ID)
		require.NoError(t, err)
		require.Equal(t, recipient.ID, gotRecipient.ID)

		recipients, err := recipientRepo.FindByMessageID(ctx, message.ID)
		require.NoError(t, err)
		require.Len(t, recipients, 1)

		require.NoError(t, recipientRepo.DeleteByMessageID(ctx, message.ID))
		require.NoError(t, recipientRepo.Delete(ctx, recipient.ID))
		require.NoError(t, messageRepo.MarkAsRead(ctx, message.ID))
		require.NoError(t, messageRepo.Delete(ctx, message.ID))
	})
}

func TestRepository_GormErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := cryptoutilAppsFrameworkServiceTestHelpDB.NewInMemorySQLiteDB(t)
	tenantID := googleUuid.Must(googleUuid.NewV7())
	elasticID := googleUuid.Must(googleUuid.NewV7())
	itemID := googleUuid.Must(googleUuid.NewV7())
	requestID := googleUuid.Must(googleUuid.NewV7()).String()
	auditLog := &cryptoutilAppsSmKmsModel.AuditLogEntry{ID: itemID, TenantID: tenantID, Operation: testOperationSign, RequestID: requestID}
	message := &cryptoutilAppsSmKmsModel.Message{ID: itemID}
	recipient := &cryptoutilAppsSmKmsModel.MessageRecipientJWK{ID: itemID}

	tests := []struct {
		name string
		call func() error
	}{
		{name: "audit config get", call: func() error {
			_, err := NewAuditConfigRepository(db).Get(ctx, tenantID, testOperationSign)

			return err
		}},
		{name: "audit config list", call: func() error {
			_, err := NewAuditConfigRepository(db).GetAllForTenant(ctx, tenantID)

			return err
		}},
		{name: "audit config upsert", call: func() error {
			config := &cryptoutilAppsSmKmsModel.AuditConfig{TenantID: tenantID, Operation: testOperationSign}

			return NewAuditConfigRepository(db).Upsert(ctx, config)
		}},
		{name: "audit config delete", call: func() error { return NewAuditConfigRepository(db).Delete(ctx, tenantID, testOperationSign) }},
		{name: "audit config should audit", call: func() error {
			_, err := NewAuditConfigRepository(db).ShouldAudit(ctx, tenantID, testOperationSign)

			return err
		}},
		{name: "audit log create", call: func() error { return NewAuditLogRepository(db).Create(ctx, auditLog) }},
		{name: "audit log list", call: func() error {
			_, _, err := NewAuditLogRepository(db).List(ctx, tenantID, 0, 10)

			return err
		}},
		{name: "audit log list by elastic", call: func() error {
			_, _, err := NewAuditLogRepository(db).ListByElasticJWK(ctx, elasticID, 0, 10)

			return err
		}},
		{name: "audit log list by operation", call: func() error {
			_, _, err := NewAuditLogRepository(db).ListByOperation(ctx, tenantID, testOperationSign, 0, 10)

			return err
		}},
		{name: "audit log by request ID", call: func() error {
			_, err := NewAuditLogRepository(db).GetByRequestID(ctx, requestID)

			return err
		}},
		{name: "audit log delete old", call: func() error {
			_, err := NewAuditLogRepository(db).DeleteOlderThan(ctx, tenantID, 30)

			return err
		}},
		{name: "elastic create", call: func() error {
			return NewElasticJWKRepository(db).Create(ctx, newRepositoryTestElasticJWK(tenantID))
		}},
		{name: "elastic get", call: func() error {
			_, err := NewElasticJWKRepository(db).Get(ctx, tenantID, "kid")

			return err
		}},
		{name: "elastic get by ID", call: func() error {
			_, err := NewElasticJWKRepository(db).GetByID(ctx, elasticID)

			return err
		}},
		{name: "elastic list", call: func() error {
			_, _, err := NewElasticJWKRepository(db).List(ctx, tenantID, 0, 10)

			return err
		}},
		{name: "elastic update", call: func() error {
			return NewElasticJWKRepository(db).Update(ctx, newRepositoryTestElasticJWK(tenantID))
		}},
		{name: "elastic delete", call: func() error { return NewElasticJWKRepository(db).Delete(ctx, elasticID) }},
		{name: "elastic increment", call: func() error { return NewElasticJWKRepository(db).IncrementMaterialCount(ctx, elasticID) }},
		{name: "elastic decrement", call: func() error { return NewElasticJWKRepository(db).DecrementMaterialCount(ctx, elasticID) }},
		{name: "material create", call: func() error {
			return NewMaterialJWKRepository(db).Create(ctx, newRepositoryTestMaterial(elasticID, true))
		}},
		{name: "material get by kid", call: func() error {
			_, err := NewMaterialJWKRepository(db).GetByMaterialKID(ctx, "kid")

			return err
		}},
		{name: "material get by ID", call: func() error {
			_, err := NewMaterialJWKRepository(db).GetByID(ctx, itemID)

			return err
		}},
		{name: "material get active", call: func() error {
			_, err := NewMaterialJWKRepository(db).GetActiveMaterial(ctx, elasticID)

			return err
		}},
		{name: "material list", call: func() error {
			_, _, err := NewMaterialJWKRepository(db).ListByElasticJWK(ctx, elasticID, 0, 10)

			return err
		}},
		{name: "material rotate", call: func() error {
			return NewMaterialJWKRepository(db).RotateMaterial(ctx, elasticID, newRepositoryTestMaterial(elasticID, true))
		}},
		{name: "material retire", call: func() error { return NewMaterialJWKRepository(db).RetireMaterial(ctx, itemID) }},
		{name: "material delete", call: func() error { return NewMaterialJWKRepository(db).Delete(ctx, itemID) }},
		{name: "material count", call: func() error {
			_, err := NewMaterialJWKRepository(db).CountMaterials(ctx, elasticID)

			return err
		}},
		{name: "message create", call: func() error { return NewMessageRepository(db).Create(ctx, message) }},
		{name: "message get", call: func() error {
			_, err := NewMessageRepository(db).FindByID(ctx, itemID)

			return err
		}},
		{name: "message list by recipient", call: func() error {
			_, err := NewMessageRepository(db).FindByRecipientID(ctx, itemID)

			return err
		}},
		{name: "message mark read", call: func() error { return NewMessageRepository(db).MarkAsRead(ctx, itemID) }},
		{name: "message delete", call: func() error { return NewMessageRepository(db).Delete(ctx, itemID) }},
		{name: "recipient create", call: func() error { return NewMessageRecipientJWKRepository(db, nil).Create(ctx, recipient) }},
		{name: "recipient find pair", call: func() error {
			_, err := NewMessageRecipientJWKRepository(db, nil).FindByRecipientAndMessage(ctx, itemID, itemID)

			return err
		}},
		{name: "recipient list", call: func() error {
			_, err := NewMessageRecipientJWKRepository(db, nil).FindByMessageID(ctx, itemID)

			return err
		}},
		{name: "recipient delete", call: func() error { return NewMessageRecipientJWKRepository(db, nil).Delete(ctx, itemID) }},
		{name: "recipient delete by message", call: func() error { return NewMessageRecipientJWKRepository(db, nil).DeleteByMessageID(ctx, itemID) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Error(t, tc.call())
		})
	}
}
