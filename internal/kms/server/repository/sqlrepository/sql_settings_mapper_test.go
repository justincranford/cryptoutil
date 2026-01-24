// Copyright (c) 2025 Justin Cranford

package sqlrepository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapDBTypeAndURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		devMode       bool
		databaseURL   string
		wantDBType    SupportedDBType
		wantURL       string
		wantError     bool
		errorContains string
	}{
		{
			name:        "dev mode uses SQLite in-memory with shared cache",
			devMode:     true,
			databaseURL: "ignored-url",
			wantDBType:  DBTypeSQLite,
			wantURL:     "file::memory:?cache=shared",
			wantError:   false,
		},
		{
			name:        "sqlite URL in-memory with shared cache",
			devMode:     false,
			databaseURL: "sqlite://file::memory:?cache=shared",
			wantDBType:  DBTypeSQLite,
			wantURL:     "file::memory:?cache=shared",
			wantError:   false,
		},
		{
			name:        "sqlite URL file-based database",
			devMode:     false,
			databaseURL: "sqlite:///tmp/test.db",
			wantDBType:  DBTypeSQLite,
			wantURL:     "/tmp/test.db",
			wantError:   false,
		},
		{
			name:        "postgres URL parsed correctly",
			devMode:     false,
			databaseURL: "postgres://user:pass@localhost/db",
			wantDBType:  DBTypePostgres,
			wantURL:     "postgres://user:pass@localhost/db",
			wantError:   false,
		},
		{
			name:        "unsupported database type returns error",
			devMode:     false,
			databaseURL: "mysql://user:pass@localhost/db",
			wantDBType:  SupportedDBType(""),
			wantURL:     "",
			wantError:   true,
		},
		{
			name:        "empty database URL returns error",
			devMode:     false,
			databaseURL: "",
			wantDBType:  SupportedDBType(""),
			wantURL:     "",
			wantError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dbType, url, err := mapDBTypeAndURL(testTelemetryService, tc.devMode, tc.databaseURL)

			if tc.wantError {
				require.Error(t, err)

				if tc.errorContains != "" {
					require.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.wantDBType, dbType)
			require.Equal(t, tc.wantURL, url)
		})
	}
}

func TestMapContainerMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         string
		wantMode      ContainerMode
		wantError     bool
		errorContains string
	}{
		{
			name:      "disabled mode parsed correctly",
			input:     string(ContainerModeDisabled),
			wantMode:  ContainerModeDisabled,
			wantError: false,
		},
		{
			name:      "preferred mode parsed correctly",
			input:     string(ContainerModePreferred),
			wantMode:  ContainerModePreferred,
			wantError: false,
		},
		{
			name:      "required mode parsed correctly",
			input:     string(ContainerModeRequired),
			wantMode:  ContainerModeRequired,
			wantError: false,
		},
		{
			name:      "invalid mode returns error",
			input:     "invalid-mode",
			wantMode:  ContainerMode(""),
			wantError: true,
		},
		{
			name:      "empty mode returns error",
			input:     "",
			wantMode:  ContainerMode(""),
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mode, err := mapContainerMode(testTelemetryService, tc.input)

			if tc.wantError {
				require.Error(t, err)

				if tc.errorContains != "" {
					require.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.wantMode, mode)
		})
	}
}
