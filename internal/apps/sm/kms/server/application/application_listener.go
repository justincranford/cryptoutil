// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
"context"
"fmt"

cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
cryptoutilAppsFrameworkServiceServerApplication "cryptoutil/internal/apps/framework/service/server/application"
cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
)

// ServerApplicationListener holds TLS configurations for both public and private servers.
// Returned by StartServerListenerApplication after successful initialization.
type ServerApplicationListener struct {
PublicTLSServer  *cryptoutilSharedCryptoCertificate.Subject
PrivateTLSServer *cryptoutilSharedCryptoCertificate.Subject
ShutdownFunction func()
}

// StartServerListenerApplication initializes core infrastructure (including database connectivity),
// basic services, and in-memory TLS configurations for the public and private servers.
//
// Unlike ServerInit, TLS certificates are generated in memory without writing to disk,
// making this function safe to call from parallel tests.
//
// Returns an error if database connectivity fails (e.g., PostgreSQL not running).
func StartServerListenerApplication(settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*ServerApplicationListener, error) {
if settings == nil {
return nil, fmt.Errorf("settings cannot be nil")
}

ctx := context.Background()

// Initialize core infrastructure including database connectivity.
// Fails for unavailable databases (e.g., PostgreSQL not running in the test environment).
core, err := cryptoutilAppsFrameworkServiceServerApplication.StartCore(ctx, settings)
if err != nil {
return nil, fmt.Errorf("failed to start application core: %w", err)
}

// Initialize basic services (telemetry, unseal keys, JWK generation) for TLS cert generation.
serverApplicationBasic, err := StartServerApplicationBasic(ctx, settings)
if err != nil {
core.Shutdown()

return nil, fmt.Errorf("failed to start basic application services: %w", err)
}

// Generate TLS certificate subjects in memory (no disk I/O, safe for parallel tests).
publicSubject, privateSubject, err := generateTLSServerSubjectsInMemory(settings, serverApplicationBasic)
if err != nil {
serverApplicationBasic.Shutdown()()
core.Shutdown()

return nil, fmt.Errorf("failed to generate TLS server subjects: %w", err)
}

return &ServerApplicationListener{
PublicTLSServer:  publicSubject,
PrivateTLSServer: privateSubject,
ShutdownFunction: func() {
serverApplicationBasic.Shutdown()()
core.Shutdown()
},
}, nil
}
