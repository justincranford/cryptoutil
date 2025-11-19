// Copyright (c) 2025 Justin Cranford
//
//

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

	// DefaultVerboseMode - Default verbose mode flag value.
	DefaultVerboseMode = false
	// DefaultOTLPEnabled - Default OTLP enabled flag value.
	DefaultOTLPEnabled = false
	// DefaultOTLPConsole - Default OTLP console flag value.
	DefaultOTLPConsole = false

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
	// DefaultOTLPEndpointDefault - Default OTLP endpoint. GRPC preferred over HTTP for performance.
	DefaultOTLPEndpointDefault = "grpc://127.0.0.1:4317"

	// DefaultSidecarHealthCheckMaxRetries - Maximum number of retries for sidecar health check.
	DefaultSidecarHealthCheckMaxRetries = 5
	// DefaultSidecarHealthCheckRetryDelay - Delay between sidecar health check retries.
	DefaultSidecarHealthCheckRetryDelay = 2 * time.Second
)
