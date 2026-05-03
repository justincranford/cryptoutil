//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.
//

// Package __SERVICE__ defines CLI usage text for the __PS_ID__ service entrypoint.
package __SERVICE__

// NOTE: This template is documentation/scaffold only and is NOT currently enforced by lint-fitness.

const (
	__USAGE_PREFIX__UsageMain = `__PS_ID__ [subcommand] [flags]`

	__USAGE_PREFIX__UsageServer = `__PS_ID__ server [flags]`
	__USAGE_PREFIX__UsageClient = `__PS_ID__ client [flags]`
	__USAGE_PREFIX__UsageInit   = `__PS_ID__ init [flags]`

	__USAGE_PREFIX__UsageHealth   = `__PS_ID__ health [flags]`
	__USAGE_PREFIX__UsageLivez    = `__PS_ID__ livez [flags]`
	__USAGE_PREFIX__UsageReadyz   = `__PS_ID__ readyz [flags]`
	__USAGE_PREFIX__UsageShutdown = `__PS_ID__ shutdown [flags]`
)
