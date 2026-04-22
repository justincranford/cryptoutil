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
	// OTLPTLSCertFile is the path to the client certificate (PEM) for mTLS to the OTLP endpoint.
	// When set together with OTLPTLSKeyFile and OTLPTLSCAFile, the OTLP exporter uses mTLS.
	OTLPTLSCertFile string
	// OTLPTLSKeyFile is the path to the client private key (PEM) for mTLS to the OTLP endpoint.
	OTLPTLSKeyFile string
	// OTLPTLSCAFile is the path to the CA certificate (PEM) used to verify the OTLP server cert.
	OTLPTLSCAFile string
}
