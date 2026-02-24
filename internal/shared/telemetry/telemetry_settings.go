// Copyright (c) 2025 Justin Cranford
//
//

package telemetry

// TelemetrySettings contains the configuration needed by the telemetry service.
// This struct breaks the dependency on apps/template/service/config.
type TelemetrySettings struct {
LogLevel        string
VerboseMode     bool
OTLPEnabled     bool
OTLPConsole     bool
OTLPService     string
OTLPInstance    string
OTLPVersion     string
OTLPEnvironment string
OTLPHostname    string
OTLPEndpoint    string
}
