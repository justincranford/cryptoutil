// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"fmt"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceServerApplication "cryptoutil/internal/apps/framework/service/server/application"
)

// ServerApplicationListener holds TLS configurations for both public and private servers.
// Returned by StartServerListenerApplication after successful initialization.
// Deprecated: Use cryptoutilAppsFrameworkServiceServerApplication.TLSListener directly.
type ServerApplicationListener = cryptoutilAppsFrameworkServiceServerApplication.TLSListener

// StartServerListenerApplication initializes core infrastructure (including database connectivity),
// basic services, and in-memory TLS configurations for the public and private servers.
//
// Unlike ServerInit, TLS certificates are generated in memory without writing to disk,
// making this function safe to call from parallel tests.
//
// Returns an error if database connectivity fails (e.g., PostgreSQL not running).
// Deprecated: Use cryptoutilAppsFrameworkServiceServerApplication.StartTLSListener directly.
func StartServerListenerApplication(settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*ServerApplicationListener, error) {
	result, err := cryptoutilAppsFrameworkServiceServerApplication.StartTLSListener(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start TLS listener: %w", err)
	}

	return result, nil
}
