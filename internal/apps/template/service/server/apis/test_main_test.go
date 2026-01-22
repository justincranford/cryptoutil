package apis

import (
	"context"
	"fmt"
	"os"
	"testing"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilTemplateServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTemplateServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
)

var testGormDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Initialize shared test fixtures (TLS certificates).
	if err := cryptoutilTemplateServerTestutil.Initialize(); err != nil {
		panic("failed to initialize test fixtures: " + err.Error())
	}

	// Create SQLite in-memory database URL with unique identifier to prevent test pollution.
	databaseURL := fmt.Sprintf("file:%s?mode=memory&cache=shared", googleUuid.NewString())

	// Initialize GORM database with migrations.
	var err error

	testGormDB, err = cryptoutilTemplateServerRepository.InitSQLite(ctx, databaseURL, cryptoutilTemplateServerRepository.MigrationsFS)
	if err != nil {
		panic("failed to create test database: " + err.Error())
	}

	// Run all tests.
	exitCode := m.Run()

	// Cleanup: Close database connection.
	if testGormDB != nil {
		sqlDB, err := testGormDB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}

	os.Exit(exitCode)
}
