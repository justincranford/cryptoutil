//go:build e2e

// Copyright (c) 2025-2026 Justin Cranford.
package e2e_test

import (
	"context"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	testcontainersPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilPKIServer "cryptoutil/internal/apps/sm-kms/server/pki/server"
	cryptoutilPKIRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilAppsSmKmsRepository "cryptoutil/internal/apps/sm-kms/server/repository"
	cryptoutilSharedContainer "cryptoutil/internal/shared/container"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedDatabase "cryptoutil/internal/shared/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

const (
	postgresContinuityAuthorityID = "postgres-e2e-authority"
	postgresContinuityRequesterID = "e2e-test"
)

func init() {
	postgresContinuityTestEnabled = true
	postgresContinuitySetup = setupPostgresContinuityContainer
}

func setupPostgresContinuityContainer(ctx context.Context) (*testcontainersPostgres.PostgresContainer, string, error) {
	container, err := cryptoutilSharedContainer.NewPostgresTestContainer(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start PostgreSQL continuity test container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		if terminateErr := container.Terminate(context.Background()); terminateErr != nil {
			return nil, "", fmt.Errorf(
				"failed to get PostgreSQL continuity test connection string: %w; failed to terminate container: %w",
				err,
				terminateErr,
			)
		}

		return nil, "", fmt.Errorf("failed to get PostgreSQL continuity test connection string: %w", err)
	}

	return container, dsn, nil
}

func TestPostgresStores_AllocateCRLNumbersAndPreserveRevocationAcrossConnections(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	firstInstanceDB, closeFirstInstanceDB := openPostgresConnection(t, postgresContinuityDSN)
	secondInstanceDB, closeSecondInstanceDB := openPostgresConnection(t, postgresContinuityDSN)
	firstStore, err := cryptoutilCAStorage.NewGORMStore(firstInstanceDB)
	require.NoError(t, err)
	secondStore, err := cryptoutilCAStorage.NewGORMStore(secondInstanceDB)
	require.NoError(t, err)

	stores := []*cryptoutilCAStorage.GORMStore{firstStore, secondStore}

	require.NoError(t, applyPostgresContinuityMigrations(t, firstInstanceDB))

	telemetryConfig := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(
		cryptoutilSharedMagic.IPv4Loopback,
		0,
		true,
	)
	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, telemetryConfig.ToTelemetrySettings())
	require.NoError(t, err)
	t.Cleanup(func() { telemetryService.Shutdown() })

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenService.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(
		&cryptoutilSharedCryptoJose.EncA256GCM,
		&cryptoutilSharedCryptoJose.AlgA256KW,
	)
	require.NoError(t, err)
	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealKeysService.Shutdown() })

	authorityConfig := cryptoutilPKIServer.DefaultPKIAuthorityConfig()
	authorityConfig.ID = postgresContinuityAuthorityID
	barrierService, shutdownBarrier := newPostgresContinuityBarrier(
		t,
		ctx,
		telemetryService,
		jwkGenService,
		unsealKeysService,
		firstInstanceDB,
	)
	_, err = cryptoutilPKIServer.NewPKIComponentsWithAuthorityConfig(ctx, firstInstanceDB, barrierService, authorityConfig)
	require.NoError(t, err)
	issuerCertificate, issuerKey, persistedAuthority := loadPostgresContinuityAuthority(
		t,
		ctx,
		firstInstanceDB,
		barrierService,
		authorityConfig.ID,
	)
	require.Equal(t, cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1, persistedAuthority.PrivateKeyEnvelopeVersion)

	issuerKeyID := fmt.Sprintf("%x", issuerCertificate.SubjectKeyId)

	requestID, err := googleUuid.NewV7()
	require.NoError(t, err)
	issuedCertificate := createPostgresContinuityIssuedCertificate(t, issuerCertificate, issuerKey, requestID)

	const allocationCount = 20

	type allocationResult struct {
		number int64
		err    error
	}

	results := make(chan allocationResult, allocationCount)

	var workers sync.WaitGroup

	for index := range allocationCount {
		store := stores[index%len(stores)]

		workers.Add(1)
		go func() {
			defer workers.Done()

			number, allocationErr := store.AllocateCRLNumber(ctx, issuerKeyID)
			if allocationErr != nil {
				results <- allocationResult{err: allocationErr}

				return
			}

			results <- allocationResult{number: number.Int64()}
		}()
	}

	workers.Wait()
	close(results)

	numbers := make([]int64, 0, allocationCount)

	for result := range results {
		require.NoError(t, result.err)
		numbers = append(numbers, result.number)
	}

	sort.Slice(numbers, func(left, right int) bool {
		return numbers[left] < numbers[right]
	})
	require.Len(t, numbers, allocationCount)

	for index, number := range numbers {
		require.Equal(t, int64(index+1), number)
	}

	tenantID, err := googleUuid.NewV7()
	require.NoError(t, err)

	storedCertificate, err := cryptoutilCAStorage.NewStoredCertificateFromX509(
		issuedCertificate,
		cryptoutilSharedMagic.PKIProfileIDTLSServer,
		postgresContinuityRequesterID,
	)
	require.NoError(t, err)

	tenantContext := cryptoutilSharedDatabase.WithTenantContext(ctx, &cryptoutilSharedDatabase.TenantContext{TenantID: tenantID})
	storedCertificate.RequestID = requestID
	storedCertificate.IssuerAuthorityID = postgresContinuityAuthorityID
	require.NoError(t, firstStore.Store(tenantContext, storedCertificate))
	require.NoError(t, secondStore.Revoke(tenantContext, storedCertificate.ID, cryptoutilCAStorage.ReasonKeyCompromise))

	revokedBeforeRestart, err := secondStore.GetRevokedByIssuerKeyID(ctx, issuerKeyID)
	require.NoError(t, err)
	require.Len(t, revokedBeforeRestart, 1)
	crlService := newPostgresContinuityCRLService(t, issuerCertificate, issuerKey, secondStore)
	crlDER, err := crlService.GenerateCRLWithRevokedCertificatesContext(
		ctx,
		postgreSQLContinuityRevocationEntries(t, revokedBeforeRestart),
	)
	require.NoError(t, err)
	crl, err := x509.ParseRevocationList(crlDER)
	require.NoError(t, err)
	require.NoError(t, crl.CheckSignatureFrom(issuerCertificate))
	require.Equal(t, big.NewInt(allocationCount+1), crl.Number)
	require.Len(t, crl.RevokedCertificateEntries, 1)
	require.Equal(t, issuedCertificate.SerialNumber, crl.RevokedCertificateEntries[0].SerialNumber)

	shutdownBarrier()
	require.NoError(t, closeFirstInstanceDB())
	require.NoError(t, closeSecondInstanceDB())

	stopTimeout := 30 * time.Second

	require.NotNil(t, postgresContinuityContainer)
	require.NoError(t, postgresContinuityContainer.Stop(ctx, &stopTimeout))
	require.NoError(t, postgresContinuityContainer.Start(ctx))
	require.NoError(t, wait.ForListeningPort("5432/tcp").
		WithStartupTimeout(cryptoutilSharedMagic.DBPostgresContainerStartupTimeout).
		WaitUntilReady(ctx, postgresContinuityContainer))

	restartedDSN, err := postgresContinuityContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	restartedDB, _ := openPostgresConnection(t, restartedDSN)
	require.NoError(t, applyPostgresContinuityMigrations(t, restartedDB))
	restartedBarrierService, shutdownRestartedBarrier := newPostgresContinuityBarrier(
		t,
		ctx,
		telemetryService,
		jwkGenService,
		unsealKeysService,
		restartedDB,
	)
	_, err = cryptoutilPKIServer.NewPKIComponentsWithAuthorityConfig(
		ctx,
		restartedDB,
		restartedBarrierService,
		authorityConfig,
	)
	require.NoError(t, err)
	restartedIssuerCertificate, restartedIssuerKey, restartedAuthority := loadPostgresContinuityAuthority(
		t,
		ctx,
		restartedDB,
		restartedBarrierService,
		authorityConfig.ID,
	)
	require.Equal(t, persistedAuthority.CertificateDER, restartedAuthority.CertificateDER)
	require.Equal(t, persistedAuthority.PrivateKeyEncrypted, restartedAuthority.PrivateKeyEncrypted)
	require.Equal(t, persistedAuthority.PrivateKeyEnvelopeVersion, restartedAuthority.PrivateKeyEnvelopeVersion)
	require.Equal(t, issuerCertificate.SubjectKeyId, restartedIssuerCertificate.SubjectKeyId)

	firstIssuerPublicKey, err := x509.MarshalPKIXPublicKey(issuerKey.Public())
	require.NoError(t, err)
	restartedIssuerPublicKey, err := x509.MarshalPKIXPublicKey(restartedIssuerKey.Public())
	require.NoError(t, err)
	require.Equal(t, firstIssuerPublicKey, restartedIssuerPublicKey)
	shutdownRestartedBarrier()

	restartedStore, err := cryptoutilCAStorage.NewGORMStore(restartedDB)
	require.NoError(t, err)
	revoked, err := restartedStore.GetRevokedByIssuerKeyID(ctx, issuerKeyID)
	require.NoError(t, err)
	require.Len(t, revoked, 1)
	require.Equal(t, issuedCertificate.SerialNumber.Text(16), revoked[0].SerialNumber)
	require.Equal(t, postgresContinuityAuthorityID, revoked[0].IssuerAuthorityID)
	require.Equal(t, requestID, revoked[0].RequestID)
	require.Equal(t, cryptoutilCAStorage.StatusRevoked, revoked[0].Status)
	require.NotNil(t, revoked[0].RevocationTime)
	require.NotNil(t, revoked[0].RevocationReason)
	require.Equal(t, cryptoutilCAStorage.ReasonKeyCompromise, *revoked[0].RevocationReason)

	restartedCRLService := newPostgresContinuityCRLService(t, issuerCertificate, issuerKey, restartedStore)
	restartedCRLDER, err := restartedCRLService.GenerateCRLWithRevokedCertificatesContext(
		ctx,
		postgreSQLContinuityRevocationEntries(t, revoked),
	)
	require.NoError(t, err)
	restartedCRL, err := x509.ParseRevocationList(restartedCRLDER)
	require.NoError(t, err)
	require.NoError(t, restartedCRL.CheckSignatureFrom(issuerCertificate))
	require.Equal(t, big.NewInt(allocationCount+2), restartedCRL.Number)
	require.Len(t, restartedCRL.RevokedCertificateEntries, 1)
	require.Equal(t, issuedCertificate.SerialNumber, restartedCRL.RevokedCertificateEntries[0].SerialNumber)
}

func createPostgresContinuityIssuedCertificate(
	t *testing.T,
	issuerCertificate *x509.Certificate,
	issuerKey crypto.Signer,
	requestID googleUuid.UUID,
) *x509.Certificate {
	t.Helper()

	now := time.Now().UTC()
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	leafTemplate := &x509.Certificate{
		SerialNumber: new(big.Int).SetBytes(requestID[:]),
		Subject:      pkix.Name{CommonName: "postgres-continuity.example.test"},
		NotBefore:    now.Add(-time.Minute),
		NotAfter:     now.Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	leafDER, err := x509.CreateCertificate(crand.Reader, leafTemplate, issuerCertificate, &leafKey.PublicKey, issuerKey)
	require.NoError(t, err)
	issuedCertificate, err := x509.ParseCertificate(leafDER)
	require.NoError(t, err)
	require.NoError(t, issuedCertificate.CheckSignatureFrom(issuerCertificate))

	return issuedCertificate
}

func applyPostgresContinuityMigrations(t *testing.T, db *gorm.DB) error {
	t.Helper()

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get SQL database for PKI continuity migrations: %w", err)
	}

	runner := cryptoutilAppsFrameworkServiceServerRepository.NewMigrationRunner(
		cryptoutilAppsFrameworkServiceServerRepository.NewMergedMigrationsFS(cryptoutilAppsSmKmsRepository.MigrationsFS),
		"migrations",
	)
	if err := runner.Apply(sqlDB, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypePostgreSQL); err != nil {
		return fmt.Errorf("apply PKI continuity migrations: %w", err)
	}

	return nil
}

func newPostgresContinuityBarrier(
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

func loadPostgresContinuityAuthority(
	t *testing.T,
	ctx context.Context,
	db *gorm.DB,
	barrierService *cryptoutilAppsFrameworkServiceServerBarrier.Service,
	authorityID string,
) (*x509.Certificate, crypto.Signer, *cryptoutilCAStorage.StoredAuthority) {
	t.Helper()

	authorityStore, err := cryptoutilCAStorage.NewGORMAuthorityStore(db)
	require.NoError(t, err)
	authority, err := authorityStore.Get(ctx, authorityID)
	require.NoError(t, err)

	issuerCertificate, err := x509.ParseCertificate(authority.CertificateDER)
	require.NoError(t, err)
	privateKeyDER, err := barrierService.DecryptContentWithContext(ctx, authority.PrivateKeyEncrypted)
	require.NoError(t, err)
	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyDER)
	clear(privateKeyDER)
	require.NoError(t, err)

	issuerKey, ok := privateKey.(crypto.Signer)
	require.True(t, ok)
	require.NoError(t, issuerCertificate.CheckSignatureFrom(issuerCertificate))

	return issuerCertificate, issuerKey, authority
}

func newPostgresContinuityCRLService(
	t *testing.T,
	issuerCertificate *x509.Certificate,
	issuerKey crypto.Signer,
	store *cryptoutilCAStorage.GORMStore,
) *cryptoutilPKIRevocation.CRLService {
	t.Helper()

	service, err := cryptoutilPKIRevocation.NewCRLService(&cryptoutilPKIRevocation.CRLConfig{
		Issuer:            issuerCertificate,
		PrivateKey:        issuerKey,
		Provider:          cryptoutilCACrypto.NewSoftwareProvider(),
		Validity:          time.Hour,
		AllocateCRLNumber: store.AllocateCRLNumber,
	})
	require.NoError(t, err)

	return service
}

func postgreSQLContinuityRevocationEntries(
	t *testing.T,
	certificates []*cryptoutilCAStorage.StoredCertificate,
) []cryptoutilPKIRevocation.RevokedCertificate {
	t.Helper()

	revocations := make([]cryptoutilPKIRevocation.RevokedCertificate, 0, len(certificates))
	for _, certificate := range certificates {
		require.NotNil(t, certificate.RevocationTime)
		require.NotNil(t, certificate.RevocationReason)
		serialNumber, ok := new(big.Int).SetString(certificate.SerialNumber, 16)
		require.True(t, ok)

		revocations = append(revocations, cryptoutilPKIRevocation.RevokedCertificate{
			SerialNumber:   serialNumber,
			RevocationTime: *certificate.RevocationTime,
			Reason:         cryptoutilPKIRevocation.RevocationReason(*certificate.RevocationReason),
		})
	}

	return revocations
}

func openPostgresConnection(t *testing.T, dsn string) (*gorm.DB, func() error) {
	t.Helper()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	var (
		closeOnce sync.Once
		closeErr  error
	)

	closeFn := func() error {
		closeOnce.Do(func() {
			closeErr = sqlDB.Close()
		})

		return closeErr
	}

	t.Cleanup(func() {
		require.NoError(t, closeFn())
	})

	return db, closeFn
}
