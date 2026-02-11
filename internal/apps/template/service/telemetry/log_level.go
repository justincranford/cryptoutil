// Copyright (c) 2025 Justin Cranford
//
//

package telemetry

import (
	"fmt"
	stdoutLogExporter "log/slog"
	"math"
	"strings"
)

// Log level values. We use slog.Level as the underlying type so the values
// are naturally compatible with the stdlib slog package while keeping
// descriptive names here for extended levels (TRACE, NOTICE, ALL, OFF).
const (
	// LevelAll represents the lowest possible level (enable everything).
	LevelAll stdoutLogExporter.Level = math.MinInt

	// OpenTelemetry / common level values that don't exist in slog.
	// TRACE = -8, DEBUG = -4, CONFIG between DEBUG and INFO.
	LevelTrace  stdoutLogExporter.Level = -8
	LevelDebug  stdoutLogExporter.Level = stdoutLogExporter.LevelDebug // -4
	LevelConfig stdoutLogExporter.Level = -2

	// Standard slog/OpenTelemetry levels.
	LevelInfo   stdoutLogExporter.Level = stdoutLogExporter.LevelInfo  // 0
	LevelNotice stdoutLogExporter.Level = 2                            // extra level
	LevelWarn   stdoutLogExporter.Level = stdoutLogExporter.LevelWarn  // 4
	LevelError  stdoutLogExporter.Level = stdoutLogExporter.LevelError // 8
	LevelFatal  stdoutLogExporter.Level = 12                           // extra level

	// LevelMax corresponds to OFF/disable logging (highest possible value).
	LevelMax stdoutLogExporter.Level = math.MaxInt
)

// ParseLogLevel converts a log level string to its corresponding slog Level value.
func ParseLogLevel(logLevelString string) (stdoutLogExporter.Level, error) {
	switch strings.ToUpper(logLevelString) {
	case "ALL": // Log4j ALL, lowest possible value
		return LevelAll, nil
	case "TRACE": // OpenTelemetry TRACE = -8, no constant in slog, but slog allows extra levels
		return LevelTrace, nil
	case "DEBUG": // OpenTelemetry DEBUG = -4
		return LevelDebug, nil
	case "CONFIG": // Java JUL CONFIG, between DEBUG and INFO
		return LevelConfig, nil
	case "INFO": // OpenTelemetry INFO = 0
		return LevelInfo, nil
	case "NOTICE": // OpenTelemetry NOTICE = 2, no constant in slog, but slog allows extra levels
		return LevelNotice, nil
	case "WARN": // OpenTelemetry WARN = 4
		return LevelWarn, nil
	case "ERROR": // OpenTelemetry ERROR = 8
		return LevelError, nil
	case "FATAL": // OpenTelemetry FATAL = 12
		return LevelFatal, nil
	case "OFF": // Log4j OFF, disables logging, highest possible value
		return LevelMax, nil
	default:
		return LevelMax, fmt.Errorf("invalid log level: %s", logLevelString)
	}
}
