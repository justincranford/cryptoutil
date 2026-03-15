// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides stub implementations for identity and integration demos
// pending Phase 8 domain reintegration.
package demo

import "context"

// runIdentityDemo is a stub pending Phase 8 identity reintegration.
func runIdentityDemo(_ context.Context, _ *Config) int {
return ExitFailure
}

// runIntegrationDemo is a stub pending Phase 8 identity+KMS integration reintegration.
func runIntegrationDemo(_ context.Context, _ *Config) int {
return ExitFailure
}
