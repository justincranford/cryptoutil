// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"crypto"
	crand "crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTestDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestGetOrCreateCA_PersistsAuthorityAcrossRestart(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&cryptoutilCAStorage.StoredAuthority{}))

	authorityStore, err := cryptoutilCAStorage.NewGORMAuthorityStore(db)
	require.NoError(t, err)

	encrypt := func(_ context.Context, cleartext []byte) ([]byte, error) {
		return []byte(base64.StdEncoding.EncodeToString(cleartext)), nil
	}
	decrypt := func(_ context.Context, ciphertext []byte) ([]byte, error) {
		cleartext, err := base64.StdEncoding.DecodeString(string(ciphertext))
		if err != nil {
			return nil, err
		}

		return cleartext, nil
	}

	certificate, privateKey, err := getOrCreateCA(context.Background(), cryptoutilCACrypto.NewSoftwareProvider(), crand.Reader, authorityStore, encrypt, decrypt)
	require.NoError(t, err)
	require.NotNil(t, certificate)
	require.NotNil(t, privateKey)

	persisted, err := authorityStore.Get(context.Background(), cryptoutilSharedMagic.PKIDefaultAuthorityID)
	require.NoError(t, err)
	require.Equal(t, certificate.Raw, persisted.CertificateDER)
	require.NotEmpty(t, persisted.PrivateKeyEncrypted)

	restartedCertificate, restartedKey, err := getOrCreateCA(context.Background(), cryptoutilCACrypto.NewSoftwareProvider(), nil, authorityStore, encrypt, decrypt)
	require.NoError(t, err)
	require.Equal(t, certificate.Raw, restartedCertificate.Raw)

	certificatePublicKey, err := x509.MarshalPKIXPublicKey(certificate.PublicKey)
	require.NoError(t, err)

	signer, ok := restartedKey.(crypto.Signer)
	require.True(t, ok)

	restartedPublicKey, err := x509.MarshalPKIXPublicKey(signer.Public())
	require.NoError(t, err)
	require.Equal(t, certificatePublicKey, restartedPublicKey)
}

func TestGetOrCreateCA_EncryptionFailureDoesNotPersistKey(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&cryptoutilCAStorage.StoredAuthority{}))

	authorityStore, err := cryptoutilCAStorage.NewGORMAuthorityStore(db)
	require.NoError(t, err)

	encryptErr := errors.New("injected encryption failure")
	encrypt := func(context.Context, []byte) ([]byte, error) {
		return nil, encryptErr
	}
	decrypt := func(context.Context, []byte) ([]byte, error) {
		return nil, errors.New("unexpected decrypt")
	}

	_, _, err = getOrCreateCA(context.Background(), cryptoutilCACrypto.NewSoftwareProvider(), crand.Reader, authorityStore, encrypt, decrypt)
	require.ErrorIs(t, err, encryptErr)

	var count int64
	require.NoError(t, db.Model(&cryptoutilCAStorage.StoredAuthority{}).Count(&count).Error)
	require.Zero(t, count)
}
