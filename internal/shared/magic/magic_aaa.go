// Copyright (c) 2025-2026 Justin Cranford.
//
//

package magic

const (
	ERR_CODE_INVALID_REQUEST      = "INVALID_REQUEST"
	ERR_CODE_NOT_FOUND            = "NOT_FOUND"
	ERR_CODE_INTERNAL_ERROR       = "INTERNAL_ERROR"
	ERR_MSG_REQUEST_BODY_REQUIRED = "Request body is required"
)

const (
	STRATEGY_ROW_LEVEL      = "row-level"
	STRATEGY_SCHEMA_LEVEL   = "schema-level"
	STRATEGY_DATABASE_LEVEL = "database-level"
	STRATEGY_UNKNOWN        = "unknown"
	SCHEMA_PREFIX_TENANT    = "tenant_"
)

const (
	UNKNOWN         = "unknown"
	PBKDF2          = "pbkdf2"
	EMPTY           = "empty"
	PASSWORD123BANG = "Password123!"
)

const (
	CMD_SERVER = "server"
)

const (
	MSG_DEADCODE_REMOVED   = "removed in v2 (use unused linter)"
	MSG_INTERFACER_REMOVED = "removed in v2 (use revive)"
	SEVERITY_ERROR         = "ERROR"
)

const (
	DESC_TEST_ROLE   = "Test role"
	DESC_FIRST_ROLE  = "First role"
	DESC_SECOND_ROLE = "Second role"
	DESC_ADMIN_ROLE  = "Administrator role"
	DESC_TEST_TENANT = "Test tenant"
	NAME_TEST_TENANT = "Test Tenant" // Note: capitalized differently in your code
	ROLE_TYPE_DB     = "DB"
)

const (
	TEST_NAME_MULTILINE = "Multiline_content"
	TEST_NAME_BINARY    = "Binary_content"
	TEST_NEEDLE_BAR     = "bar"
	TEST_NEEDLE_WORLD   = "world"
	TEST_CONTENT_MULTI  = "Line 1\nLine 2\nLine 3\n"
)
