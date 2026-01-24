// Copyright (c) 2025 Justin Cranford
//
//

package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"
	cryptoutilJoseRepository "cryptoutil/internal/jose/repository"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

var (
	testCtx              context.Context
	testCtxCancel        context.CancelFunc
	testSQLDB            *sql.DB
	testDB               *gorm.DB
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	testBarrierService   *cryptoutilAppsTemplateServiceServerBarrier.Service
	testElasticRepo      cryptoutilJoseRepository.ElasticJWKRepository
	testMaterialRepo     cryptoutilJoseRepository.MaterialJWKRepository
	testElasticJWKSvc    *ElasticJWKService
)

func TestMain(m *testing.M) {
	testCtx, testCtxCancel = context.WithCancel(context.Background())
	defer testCtxCancel()

	// Open SQLite database.
	var err error

	testSQLDB, err = sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		panic("TestMain: failed to open SQLite: " + err.Error())
	}

	// Configure SQLite for concurrent access.
	if _, err := testSQLDB.ExecContext(testCtx, "PRAGMA journal_mode=WAL;"); err != nil {
		panic("TestMain: failed to enable WAL mode: " + err.Error())
	}

	if _, err := testSQLDB.ExecContext(testCtx, "PRAGMA busy_timeout = 30000;"); err != nil {
		panic("TestMain: failed to set busy timeout: " + err.Error())
	}

	testSQLDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetConnMaxLifetime(0)

	// Wrap with GORM.
	testDB, err = gorm.Open(sqlite.Dialector{Conn: testSQLDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("TestMain: failed to create GORM DB: " + err.Error())
	}

	// Create barrier tables.
	if err := createBarrierTables(testSQLDB); err != nil {
		panic("TestMain: failed to create barrier tables: " + err.Error())
	}

	// Auto-migrate domain models.
	if err := testDB.AutoMigrate(&cryptoutilJoseDomain.ElasticJWK{}, &cryptoutilJoseDomain.MaterialJWK{}); err != nil {
		panic("TestMain: failed to migrate domain models: " + err.Error())
	}

	// Initialize telemetry.
	telemetrySettings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	testTelemetryService, err = cryptoutilSharedTelemetry.NewTelemetryService(testCtx, telemetrySettings)
	if err != nil {
		panic("TestMain: failed to create telemetry: " + err.Error())
	}
	defer testTelemetryService.Shutdown()

	// Initialize JWK Generation Service.
	testJWKGenService, err = cryptoutilSharedCryptoJose.NewJWKGenService(testCtx, testTelemetryService, false)
	if err != nil {
		panic("TestMain: failed to create JWK service: " + err.Error())
	}
	defer testJWKGenService.Shutdown()

	// Initialize Barrier Service.
	_, testUnsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	if err != nil {
		panic("TestMain: failed to generate test unseal JWK: " + err.Error())
	}

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{testUnsealJWK})
	if err != nil {
		panic("TestMain: failed to create unseal keys service: " + err.Error())
	}
	defer unsealKeysService.Shutdown()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(testDB)
	if err != nil {
		panic("TestMain: failed to create barrier repository: " + err.Error())
	}
	defer barrierRepo.Shutdown()

	testBarrierService, err = cryptoutilAppsTemplateServiceServerBarrier.NewService(testCtx, testTelemetryService, testJWKGenService, barrierRepo, unsealKeysService)
	if err != nil {
		panic("TestMain: failed to create barrier service: " + err.Error())
	}
	defer testBarrierService.Shutdown()

	// Initialize repositories.
	testElasticRepo = cryptoutilJoseRepository.NewElasticJWKRepository(testDB)
	testMaterialRepo = cryptoutilJoseRepository.NewMaterialJWKRepository(testDB)

	// Initialize ElasticJWKService.
	testElasticJWKSvc = NewElasticJWKService(
		testElasticRepo,
		testMaterialRepo,
		testJWKGenService,
		testBarrierService,
	)

	// Run all tests.
	exitCode := m.Run()

	if closeErr := testSQLDB.Close(); closeErr != nil {
		panic("TestMain: failed to close test SQL DB: " + closeErr.Error())
	}

	os.Exit(exitCode)
}

// createBarrierTables creates the barrier encryption tables for testing.
func createBarrierTables(db *sql.DB) error {
	ctx := context.Background()

	schema := `
	CREATE TABLE IF NOT EXISTS barrier_root_keys (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		kek_uuid TEXT,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS barrier_intermediate_keys (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		kek_uuid TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS barrier_content_keys (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		kek_uuid TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	`

	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create barrier tables: %w", err)
	}

	return nil
}

// TestNewElasticJWKService tests service construction.
func TestNewElasticJWKService(t *testing.T) {
	require.NotNil(t, testElasticJWKSvc)
}

// TestCreateElasticJWK tests creating a new Elastic JWK.
func TestCreateElasticJWK(t *testing.T) {
	tests := []struct {
		name      string
		keyType   string
		algorithm string
		keyUse    string
		wantErr   bool
	}{
		{
			name:      "RSA signing key",
			keyType:   "RSA",
			algorithm: "RS256",
			keyUse:    "sig",
			wantErr:   false,
		},
		{
			name:      "EC signing key",
			keyType:   "EC",
			algorithm: "ES256",
			keyUse:    "sig",
			wantErr:   false,
		},
		{
			name:      "EdDSA signing key",
			keyType:   "OKP",
			algorithm: "EdDSA",
			keyUse:    "sig",
			wantErr:   false,
		},
		{
			name:      "HMAC signing key",
			keyType:   "oct",
			algorithm: "HS256",
			keyUse:    "sig",
			wantErr:   false,
		},
		{
			name:      "AES encryption key",
			keyType:   "oct",
			algorithm: "A256GCM",
			keyUse:    "enc",
			wantErr:   false,
		},
		{
			name:      "RSA-OAEP encryption key",
			keyType:   "RSA",
			algorithm: "RSA-OAEP",
			keyUse:    "enc",
			wantErr:   false,
		},
		{
			name:      "ECDH-ES encryption key",
			keyType:   "EC",
			algorithm: "ECDH-ES",
			keyUse:    "enc",
			wantErr:   false,
		},
		{
			name:      "Invalid algorithm",
			keyType:   "RSA",
			algorithm: "INVALID",
			keyUse:    "sig",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantID := googleUuid.New()
			realmID := googleUuid.New()

			req := &CreateElasticJWKRequest{
				TenantID: tenantID,
				RealmID:  realmID,
				KTY:      tt.keyType,
				ALG:      tt.algorithm,
				USE:      tt.keyUse,
			}

			resp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.ElasticJWK)
				require.NotNil(t, resp.MaterialJWK)
				require.NotNil(t, resp.PublicJWK)
				require.Equal(t, tenantID, resp.ElasticJWK.TenantID)
				require.Equal(t, realmID, resp.ElasticJWK.RealmID)
				require.Equal(t, tt.keyType, resp.ElasticJWK.KTY)
				require.Equal(t, tt.algorithm, resp.ElasticJWK.ALG)
				require.Equal(t, tt.keyUse, resp.ElasticJWK.USE)
			}
		})
	}
}

// TestGetElasticJWK tests retrieving an Elastic JWK by KID.
func TestGetElasticJWK(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an Elastic JWK first.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RS256",
		USE:      "sig",
	}

	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)
	require.NotNil(t, createResp)

	// Test getting the created JWK by KID.
	t.Run("Get existing JWK", func(t *testing.T) {
		retrieved, err := testElasticJWKSvc.GetElasticJWK(testCtx, tenantID, realmID, createResp.ElasticJWK.KID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, createResp.ElasticJWK.ID, retrieved.ID)
		require.Equal(t, tenantID, retrieved.TenantID)
		require.Equal(t, realmID, retrieved.RealmID)
	})

	// Test getting a non-existent JWK.
	t.Run("Get non-existent JWK", func(t *testing.T) {
		retrieved, err := testElasticJWKSvc.GetElasticJWK(testCtx, tenantID, realmID, "non-existent-kid")
		require.Error(t, err)
		require.Nil(t, retrieved)
	})

	// Test getting from wrong tenant.
	t.Run("Get from wrong tenant", func(t *testing.T) {
		wrongTenant := googleUuid.New()
		retrieved, err := testElasticJWKSvc.GetElasticJWK(testCtx, wrongTenant, realmID, createResp.ElasticJWK.KID)
		require.Error(t, err)
		require.Nil(t, retrieved)
	})
}

// TestListElasticJWKs tests listing Elastic JWKs.
func TestListElasticJWKs(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	realmID2 := googleUuid.New()

	// Create multiple Elastic JWKs.
	for i := 0; i < 3; i++ {
		req := &CreateElasticJWKRequest{
			TenantID: tenantID,
			RealmID:  realmID,
			KTY:      "RSA",
			ALG:      "RS256",
			USE:      "sig",
		}

		_, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
		require.NoError(t, err)
	}

	// Create one in a different realm.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID2,
		KTY:      "RSA",
		ALG:      "RS256",
		USE:      "sig",
	}
	_, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)

	// Test listing JWKs for specific realm.
	t.Run("List JWKs for realm", func(t *testing.T) {
		jwks, err := testElasticJWKSvc.ListElasticJWKs(testCtx, tenantID, realmID, 0, 100)
		require.NoError(t, err)
		require.Len(t, jwks, 3)
	})

	// Test listing JWKs for empty realm.
	t.Run("List JWKs for empty realm", func(t *testing.T) {
		emptyRealmID := googleUuid.New()
		jwks, err := testElasticJWKSvc.ListElasticJWKs(testCtx, tenantID, emptyRealmID, 0, 100)
		require.NoError(t, err)
		require.Empty(t, jwks)
	})
}

// TestGetActiveMaterialJWK tests getting the active Material JWK.
func TestGetActiveMaterialJWK(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an Elastic JWK.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RS256",
		USE:      "sig",
	}

	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)

	// Get the active material JWK.
	t.Run("Get active material JWK", func(t *testing.T) {
		materialJWK, privateJWK, publicJWK, err := testElasticJWKSvc.GetActiveMaterialJWK(testCtx, createResp.ElasticJWK.ID)
		require.NoError(t, err)
		require.NotNil(t, materialJWK)
		require.NotNil(t, privateJWK)
		require.NotNil(t, publicJWK)
		require.Equal(t, createResp.ElasticJWK.ID, materialJWK.ElasticJWKID)
		require.True(t, materialJWK.Active)
	})
}

// TestGetMaterialJWKByKID tests getting a Material JWK by KID.
func TestGetMaterialJWKByKID(t *testing.T) {
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create an Elastic JWK.
	req := &CreateElasticJWKRequest{
		TenantID: tenantID,
		RealmID:  realmID,
		KTY:      "RSA",
		ALG:      "RS256",
		USE:      "sig",
	}

	createResp, err := testElasticJWKSvc.CreateElasticJWK(testCtx, req)
	require.NoError(t, err)

	// Get the active material JWK to obtain its MaterialKID.
	activeMaterial, _, _, err := testElasticJWKSvc.GetActiveMaterialJWK(testCtx, createResp.ElasticJWK.ID)
	require.NoError(t, err)

	// Get material by MaterialKID.
	t.Run("Get material by KID", func(t *testing.T) {
		materialJWK, privateJWK, publicJWK, err := testElasticJWKSvc.GetMaterialJWKByKID(testCtx, createResp.ElasticJWK.ID, activeMaterial.MaterialKID)
		require.NoError(t, err)
		require.NotNil(t, materialJWK)
		require.NotNil(t, privateJWK)
		require.NotNil(t, publicJWK)
		require.Equal(t, activeMaterial.MaterialKID, materialJWK.MaterialKID)
	})

	// Get material by non-existent KID.
	t.Run("Get material by non-existent KID", func(t *testing.T) {
		materialJWK, privateJWK, publicJWK, err := testElasticJWKSvc.GetMaterialJWKByKID(testCtx, createResp.ElasticJWK.ID, "non-existent-kid")
		require.Error(t, err)
		require.Nil(t, materialJWK)
		require.Nil(t, privateJWK)
		require.Nil(t, publicJWK)
	})
}

// TestMapToJWASignatureAlgorithm tests the algorithm mapping function.
func TestMapToJWASignatureAlgorithm(t *testing.T) {
	tests := []struct {
		name    string
		alg     string
		wantErr bool
	}{
		{"RS256", "RS256", false},
		{"RS384", "RS384", false},
		{"RS512", "RS512", false},
		{"PS256", "PS256", false},
		{"PS384", "PS384", false},
		{"PS512", "PS512", false},
		{"ES256", "ES256", false},
		{"ES384", "ES384", false},
		{"ES512", "ES512", false},
		{"EdDSA", "EdDSA", false},
		{"HS256", "HS256", false},
		{"HS384", "HS384", false},
		{"HS512", "HS512", false},
		{"Invalid", "INVALID", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alg, err := mapToJWASignatureAlgorithm(tt.alg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, alg)
			}
		})
	}
}

// TestMapToJWAEncryptionAlgorithms tests the encryption algorithm mapping function.
func TestMapToJWAEncryptionAlgorithms(t *testing.T) {
	tests := []struct {
		name    string
		alg     string
		wantErr bool
	}{
		{"A128GCM", "A128GCM", false},
		{"A192GCM", "A192GCM", false},
		{"A256GCM", "A256GCM", false},
		{"A128CBC-HS256", "A128CBC-HS256", false},
		{"A192CBC-HS384", "A192CBC-HS384", false},
		{"A256CBC-HS512", "A256CBC-HS512", false},
		{"RSA-OAEP", "RSA-OAEP", false},
		{"RSA-OAEP-256", "RSA-OAEP-256", false},
		{"RSA-OAEP-384", "RSA-OAEP-384", false},
		{"RSA-OAEP-512", "RSA-OAEP-512", false},
		{"ECDH-ES", "ECDH-ES", false},
		{"ECDH-ES+A128KW", "ECDH-ES+A128KW", false},
		{"ECDH-ES+A192KW", "ECDH-ES+A192KW", false},
		{"ECDH-ES+A256KW", "ECDH-ES+A256KW", false},
		{"Invalid", "INVALID", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, keyAlg, err := mapToJWAEncryptionAlgorithms(tt.alg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, enc)
				require.NotNil(t, keyAlg)
			}
		})
	}
}
