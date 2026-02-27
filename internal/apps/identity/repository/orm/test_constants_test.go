// Copyright (c) 2025 Justin Cranford
//
//

package orm_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// testDSNInMemory is the SQLite DSN for in-memory isolated databases (parallel test safety).
const testDSNInMemory = cryptoutilSharedMagic.SQLiteMemoryPlaceholder
