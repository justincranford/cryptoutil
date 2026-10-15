//go:build integration

// Copyright (c) 2025-2026 Justin Cranford.
package repository

import (
	"bytes"
	"context"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	http "net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ocsp"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilPKIServer "cryptoutil/internal/apps/sm-kms/server/pki/server"
	cryptoutilPKIRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilPKIStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedDatabase "cryptoutil/internal/shared/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

const testPKIPersistenceFQDN = "pki-persistence.example.test"

func TestPKIPersistence_ProductionMigrationsBarrierAndRestart(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "pki-production.sqlite")
	dsn := "file:" + filepath.ToSlash(databasePath) + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	authorityConfig := cryptoutilPKIServer.DefaultPKIAuthorityConfig()
	authorityConfig.ID = "production-chain-test-authority"

	telemetryConfig := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, telemetryConfig.ToTelemetrySettings())
	require.NoError(t, err)
	t.Cleanup(func() { telemetryService.Shutdown() })

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenService.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealKeysService.Shutdown() })

	sqlDB, gormDB := openMigratedProductionSQLite(t, dsn)
	barrierService, shutdownBarrier := newProductionTestBarrier(t, ctx, telemetryService, jwkGenService, unsealKeysService, gormDB)

	_, err = cryptoutilPKIServer.NewPKIComponentsWithAuthorityConfig(ctx, gormDB, barrierService, authorityConfig)
	require.NoError(t, err)

	authorityStore, err := cryptoutilPKIStorage.NewGORMAuthorityStore(gormDB)
	require.NoError(t, err)
	firstAuthority, err := authorityStore.Get(ctx, authorityConfig.ID)
	require.NoError(t, err)

	firstCertificateDER := append([]byte(nil), firstAuthority.CertificateDER...)
	firstEncryptedPrivateKey := append([]byte(nil), firstAuthority.PrivateKeyEncrypted...)

	certificate, err := x509.ParseCertificate(firstCertificateDER)
	require.NoError(t, err)
	encryptedKeyDER, err := barrierService.DecryptContentWithContext(ctx, firstEncryptedPrivateKey)
	require.NoError(t, err)
	privateKey, err := x509.ParsePKCS8PrivateKey(encryptedKeyDER)
	clear(encryptedKeyDER)
	require.NoError(t, err)

	caSigner, ok := privateKey.(crypto.Signer)
	require.True(t, ok)

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	now := time.Now().UTC()
	leafTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(0x1234),
		Subject:      pkix.Name{CommonName: testPKIPersistenceFQDN},
		DNSNames:     []string{testPKIPersistenceFQDN},
		NotBefore:    now.Add(-time.Minute),
		NotAfter:     now.Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	leafDER, err := x509.CreateCertificate(crand.Reader, leafTemplate, certificate, &leafKey.PublicKey, caSigner)
	require.NoError(t, err)
	leafCertificate, err := x509.ParseCertificate(leafDER)
	require.NoError(t, err)

	certificateStore, err := cryptoutilPKIStorage.NewGORMStore(gormDB)
	require.NoError(t, err)
	requestID := googleUUIDV7(t)
	tenantID := googleUUIDV7(t)
	tenantContext := cryptoutilSharedDatabase.WithTenantContext(ctx, &cryptoutilSharedDatabase.TenantContext{TenantID: tenantID})
	storedCertificate, err := cryptoutilPKIStorage.NewStoredCertificateFromX509(
		leafCertificate,
		cryptoutilSharedMagic.PKIProfileIDTLSServer,
		"production-chain-test",
	)
	require.NoError(t, err)

	storedCertificate.RequestID = requestID
	storedCertificate.IssuerAuthorityID = authorityConfig.ID
	storedCertificate.CertificatePEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: leafDER,
	}))
	require.NoError(t, certificateStore.Store(tenantContext, storedCertificate))

	duplicateRequest := *storedCertificate
	duplicateRequest.ID = googleUUIDV7(t)
	duplicateRequest.SerialNumber = "1235"
	require.Error(t, gormDB.Create(&duplicateRequest).Error, "production migration must enforce tenant request-ID uniqueness")

	require.NoError(t, certificateStore.Revoke(tenantContext, storedCertificate.ID, cryptoutilPKIStorage.ReasonKeyCompromise))

	issuerKeyID := fmt.Sprintf("%x", certificate.SubjectKeyId)
	revoked, err := certificateStore.GetRevokedByIssuerKeyID(ctx, issuerKeyID)
	require.NoError(t, err)
	require.Len(t, revoked, 1)
	require.Equal(t, storedCertificate.ID, revoked[0].ID)

	firstCRLNumber, err := certificateStore.AllocateCRLNumber(ctx, issuerKeyID)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(1), firstCRLNumber)

	shutdownBarrier()
	closeProductionSQLite(t, sqlDB)

	restartedSQLDB, restartedGORMDB := openMigratedProductionSQLite(t, dsn)

	restartedBarrierService, shutdownRestartedBarrier := newProductionTestBarrier(
		t, ctx, telemetryService, jwkGenService, unsealKeysService, restartedGORMDB,
	)
	defer shutdownRestartedBarrier()

	_, err = cryptoutilPKIServer.NewPKIComponentsWithAuthorityConfig(
		ctx, restartedGORMDB, restartedBarrierService, authorityConfig,
	)
	require.NoError(t, err)

	restartedAuthorityStore, err := cryptoutilPKIStorage.NewGORMAuthorityStore(restartedGORMDB)
	require.NoError(t, err)
	restartedAuthority, err := restartedAuthorityStore.Get(ctx, authorityConfig.ID)
	require.NoError(t, err)
	require.Equal(t, firstCertificateDER, restartedAuthority.CertificateDER)
	require.Equal(t, firstEncryptedPrivateKey, restartedAuthority.PrivateKeyEncrypted)

	restartedPrivateKeyDER, err := restartedBarrierService.DecryptContentWithContext(ctx, restartedAuthority.PrivateKeyEncrypted)
	require.NoError(t, err)
	restartedPrivateKey, err := x509.ParsePKCS8PrivateKey(restartedPrivateKeyDER)
	clear(restartedPrivateKeyDER)
	require.NoError(t, err)

	restartedCertificateStore, err := cryptoutilPKIStorage.NewGORMStore(restartedGORMDB)
	require.NoError(t, err)
	restartedCertificate, err := restartedCertificateStore.GetByRequestID(tenantContext, requestID)
	require.NoError(t, err)
	require.Equal(t, cryptoutilPKIStorage.StatusRevoked, restartedCertificate.Status)
	require.Equal(t, authorityConfig.ID, restartedCertificate.IssuerAuthorityID)
	require.NotNil(t, restartedCertificate.RevocationTime)
	require.NotNil(t, restartedCertificate.RevocationReason)

	restartedRevocations, err := restartedCertificateStore.GetRevokedByIssuerKeyID(ctx, issuerKeyID)
	require.NoError(t, err)
	require.Len(t, restartedRevocations, 1)
	require.Equal(t, storedCertificate.ID, restartedRevocations[0].ID)

	restartedSigner, ok := restartedPrivateKey.(crypto.Signer)
	require.True(t, ok)

	crlService, err := cryptoutilPKIRevocation.NewCRLService(&cryptoutilPKIRevocation.CRLConfig{
		Issuer:            certificate,
		PrivateKey:        restartedSigner,
		Provider:          cryptoutilCACrypto.NewSoftwareProvider(),
		Validity:          cryptoutilSharedMagic.Days7,
		NextUpdateBuffer:  cryptoutilSharedMagic.Days1,
		AllocateCRLNumber: restartedCertificateStore.AllocateCRLNumber,
	})
	require.NoError(t, err)

	revokedSerialNumber, ok := new(big.Int).SetString(restartedCertificate.SerialNumber, cryptoutilSharedMagic.HexBase)
	require.True(t, ok)

	crlDER, err := crlService.GenerateCRLWithRevokedCertificatesContext(ctx, []cryptoutilPKIRevocation.RevokedCertificate{{
		SerialNumber:   revokedSerialNumber,
		RevocationTime: *restartedCertificate.RevocationTime,
		Reason:         cryptoutilPKIRevocation.RevocationReason(*restartedCertificate.RevocationReason),
	}})
	require.NoError(t, err)
	parsedCRL, err := x509.ParseRevocationList(crlDER)
	require.NoError(t, err)
	require.NoError(t, parsedCRL.CheckSignatureFrom(certificate))
	require.Equal(t, big.NewInt(2), parsedCRL.Number)
	require.Len(t, parsedCRL.RevokedCertificateEntries, 1)
	require.Equal(t, leafCertificate.SerialNumber, parsedCRL.RevokedCertificateEntries[0].SerialNumber)

	ocspService, err := cryptoutilPKIRevocation.NewOCSPService(&cryptoutilPKIRevocation.OCSPConfig{
		Issuer:       certificate,
		Responder:    certificate,
		ResponderKey: restartedSigner,
		Provider:     cryptoutilCACrypto.NewSoftwareProvider(),
		Validity:     cryptoutilSharedMagic.Days1,
	}, crlService)
	require.NoError(t, err)
	ocspDER, err := ocspService.CreateResponseWithRevocation(
		leafCertificate,
		&cryptoutilPKIRevocation.RevokedCertificate{
			SerialNumber:   leafCertificate.SerialNumber,
			RevocationTime: *restartedCertificate.RevocationTime,
			Reason:         cryptoutilPKIRevocation.RevocationReason(*restartedCertificate.RevocationReason),
		},
	)
	require.NoError(t, err)
	parsedOCSP, err := ocsp.ParseResponse(ocspDER, certificate)
	require.NoError(t, err)
	require.Equal(t, ocsp.Revoked, parsedOCSP.Status)
	require.Equal(t, leafCertificate.SerialNumber, parsedOCSP.SerialNumber)

	_, err = restartedCertificateStore.GetByRequestID(
		cryptoutilSharedDatabase.WithTenantContext(ctx, &cryptoutilSharedDatabase.TenantContext{TenantID: googleUUIDV7(t)}),
		requestID,
	)
	require.ErrorIs(t, err, cryptoutilPKIStorage.ErrCertificateNotFound)

	secondCRLNumber, err := restartedCertificateStore.AllocateCRLNumber(ctx, issuerKeyID)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(3), secondCRLNumber)

	rotatedAuthorityConfig := authorityConfig
	rotatedAuthorityConfig.ID = "production-chain-rotated-authority"
	rotatedComponents, err := cryptoutilPKIServer.NewPKIComponentsWithAuthorityConfig(
		ctx, restartedGORMDB, restartedBarrierService, rotatedAuthorityConfig,
	)
	require.NoError(t, err)

	revocationApp := fiber.New(fiber.Config{DisableStartupMessage: true})
	require.NoError(t, rotatedComponents.RegisterPublicRevocationEndpoints(revocationApp))

	crlRequest := httptest.NewRequest(
		http.MethodGet,
		"/pki/cas/"+authorityConfig.ID+"/crl",
		nil,
	)
	crlResponse, err := revocationApp.Test(crlRequest, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, crlResponse.Body.Close()) }()

	require.Equal(t, http.StatusOK, crlResponse.StatusCode)
	publishedCRL, err := io.ReadAll(crlResponse.Body)
	require.NoError(t, err)
	publishedRevocationList, err := x509.ParseRevocationList(publishedCRL)
	require.NoError(t, err)
	require.NoError(t, publishedRevocationList.CheckSignatureFrom(certificate))
	require.Len(t, publishedRevocationList.RevokedCertificateEntries, 1)

	ocspRequest, err := ocsp.CreateRequest(leafCertificate, certificate, &ocsp.RequestOptions{Hash: crypto.SHA256})
	require.NoError(t, err)

	ocspHTTPRequest := httptest.NewRequest(
		http.MethodPost,
		"/pki/ocsp/"+authorityConfig.ID,
		bytes.NewReader(ocspRequest),
	)
	ocspHTTPResponse, err := revocationApp.Test(ocspHTTPRequest, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, ocspHTTPResponse.Body.Close()) }()

	require.Equal(t, http.StatusOK, ocspHTTPResponse.StatusCode)
	publishedOCSP, err := io.ReadAll(ocspHTTPResponse.Body)
	require.NoError(t, err)
	publishedOCSPResponse, err := ocsp.ParseResponse(publishedOCSP, certificate)
	require.NoError(t, err)
	require.Equal(t, ocsp.Revoked, publishedOCSPResponse.Status)
	require.Equal(t, leafCertificate.SerialNumber, publishedOCSPResponse.SerialNumber)

	closeProductionSQLite(t, restartedSQLDB)
}

func openMigratedProductionSQLite(t *testing.T, dsn string) (*sql.DB, *gorm.DB) {
	t.Helper()

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(16)

	gormDB, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	runner := cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(
		cryptoutilAppsFrameworkServiceServerRepository.NewMergedMigrationsFS(MigrationsFS),
		"migrations",
	)
	require.NoError(t, runner.Apply(sqlDB, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite))

	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	return sqlDB, gormDB
}

func newProductionTestBarrier(
	t *testing.T,
	ctx context.Context,
	telemetryService *cryptoutilSharedTelemetry.TelemetryService,
	jwkGenService *cryptoutilSharedCryptoJose.JWKGenService,
	unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService,
	db *gorm.DB,
) (*cryptoutilAppsFrameworkServiceServerBarrier.Service, func()) {
	t.Helper()

	barrierRepository, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)

	barrierService, err := cryptoutilAppsFrameworkServiceServerBarrier.NewService(
		ctx,
		telemetryService,
		jwkGenService,
		barrierRepository,
		unsealKeysService,
	)
	require.NoError(t, err)

	var shutdownOnce sync.Once

	shutdown := func() {
		shutdownOnce.Do(func() {
			barrierService.Shutdown()
			barrierRepository.Shutdown()
		})
	}
	t.Cleanup(shutdown)

	return barrierService, shutdown
}

func closeProductionSQLite(t *testing.T, db *sql.DB) {
	t.Helper()

	require.NoError(t, db.Close())
}

func googleUUIDV7(t *testing.T) googleUuid.UUID {
	t.Helper()

	id, err := googleUuid.NewV7()
	require.NoError(t, err)

	return id
}
