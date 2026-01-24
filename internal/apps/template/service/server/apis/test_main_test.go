// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"context"
	"fmt"
	"os"
	"testing"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
)

var testGormDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Initialize shared test fixtures (TLS certificates).
	if err := cryptoutilAppsTemplateServiceServerTestutil.Initialize(); err != nil {
		panic("failed to initialize test fixtures: " + err.Error())
	}

	// Create SQLite in-memory database URL with unique identifier to prevent test pollution.
	databaseURL := fmt.Sprintf("file:%s?mode=memory&cache=shared", googleUuid.NewString())

	// Initialize GORM database with migrations.
	var err error

	testGormDB, err = cryptoutilAppsTemplateServiceServerRepository.InitSQLite(ctx, databaseURL, cryptoutilAppsTemplateServiceServerRepository.MigrationsFS)
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
