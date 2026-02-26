// Copyright (c) 2025 Justin Cranford
//
//

package telemetry

import (
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// NewTestTelemetrySettings creates a TelemetrySettings suitable for testing.
func NewTestTelemetrySettings(serviceName string) *TelemetrySettings {
return &TelemetrySettings{
LogLevel:        cryptoutilSharedMagic.DefaultLogLevelInfo,
VerboseMode:     false,
OTLPEnabled:     false,
OTLPConsole:     false,
OTLPService:     serviceName,
OTLPInstance:    "test-instance",
OTLPVersion:     "0.0.0-test",
OTLPEnvironment: "test",
OTLPHostname:    cryptoutilSharedMagic.DefaultOTLPHostnameDefault,
OTLPEndpoint:    cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
}
}
