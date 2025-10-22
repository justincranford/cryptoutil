// Package magic provides commonly used magic numbers and values as named constants.
// This file contains count and limit constants.
package magic

import (
	"math"
	"time"
)

// Telemetry and monitoring constants.
const (
	// DefaultLogsTimeout - Default timeout for logs export.
	DefaultLogsTimeout = 500 * time.Millisecond
	// DefaultMetricsTimeout - Default timeout for metrics export.
	DefaultMetricsTimeout = 2000 * time.Millisecond
	// DefaultTracesTimeout - Default timeout for traces export.
	DefaultTracesTimeout = 1000 * time.Millisecond
	// DefaultForceFlushTimeout - Default timeout for force flush on shutdown.
	DefaultForceFlushTimeout = 3 * time.Second // 3s for force flush on shutdown

	// DefaultLogsBatchSize - Maximum batch size for logs.
	DefaultLogsBatchSize = 1024
	// DefaultMetricsBatchSize - Maximum batch size for metrics.
	DefaultMetricsBatchSize = 2048
	// DefaultTracesBatchSize - Maximum batch size for traces.
	DefaultTracesBatchSize = 512

	// DefaultLogLevelInfo - Default log level INFO.
	DefaultLogLevelInfo = "INFO"

	// DefaultBoolVerboseMode - Default verbose mode flag value.
	DefaultBoolVerboseMode = false
	// DefaultBoolOTLP - Default OTLP flag value.
	DefaultBoolOTLP = false
	// DefaultBoolOTLPConsole - Default OTLP console flag value.
	DefaultBoolOTLPConsole = false

	// LogLevelAllValue - Lowest possible level (enable everything).
	LogLevelAllValue = math.MinInt
	// LogLevelTrace - Trace log level.
	LogLevelTrace = -8
	// LogLevelConfig - Config log level.
	LogLevelConfig = -2
	// LogLevelWarn - Warning log level.
	LogLevelWarn = 4
	// LogLevelMax - Maximum log level.
	LogLevelMax = math.MaxInt

	// DefaultOTLPServiceDefault - Default OTLP service name.
	DefaultOTLPServiceDefault = "cryptoutil"
	// DefaultOTLPVersionDefault - Default OTLP version.
	DefaultOTLPVersionDefault = "0.0.1"
	// DefaultOTLPEnvironmentDefault - Default OTLP environment.
	DefaultOTLPEnvironmentDefault = "dev"
	// DefaultOTLPHostnameDefault - Default OTLP hostname.
	DefaultOTLPHostnameDefault = "localhost"
	// DefaultOTLPEndpointDefault - Default OTLP endpoint.
	DefaultOTLPEndpointDefault = "grpc://127.0.0.1:4317"
)
