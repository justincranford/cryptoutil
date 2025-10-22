// Package magic provides commonly used magic numbers and values as named constants.
// This file contains count and limit constants.
package magic

import "math"

// Telemetry and monitoring constants.
const (
	// CountMaxLogsBatchSize - Maximum batch size for logs.
	CountMaxLogsBatchSize = 1024
	// CountMaxMetricsBatchSize - Maximum batch size for metrics.
	CountMaxMetricsBatchSize = 2048
	// CountMaxTracesBatchSize - Maximum batch size for traces.
	CountMaxTracesBatchSize = 512

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
)
