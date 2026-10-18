// Copyright (c) 2025-2026 Justin Cranford.

package server

import (
	"context"
	"fmt"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
)

func BenchmarkNewKMSServerFromConfig(b *testing.B) {
	ctx := context.Background()
	settingsTemplate := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest(cryptoutilSharedMagic.OTLPServiceSMKMS)
	settingsByIteration := make([]*cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings, b.N)

	for i := range b.N {
		databaseID, err := googleUuid.NewV7()
		if err != nil {
			b.Fatal(err)
		}

		settings := *settingsTemplate
		settings.DatabaseURL = fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, databaseID.String())
		settingsByIteration[i] = &settings
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		srv, err := NewKMSServerFromConfig(ctx, settingsByIteration[i])
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()

		err = srv.Shutdown(ctx)

		b.StartTimer()

		if err != nil {
			b.Fatal(err)
		}
	}
}
