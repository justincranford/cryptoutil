package telemetry

import (
	"fmt"
	stdoutLogExporter "log/slog"
	"math"
	"strings"
)

func ParseLogLevel(logLevelString string) (stdoutLogExporter.Level, error) {
	switch strings.ToUpper(logLevelString) {
	case "ALL": // Log4j ALL, lowest possible value
		return math.MinInt, nil
	case "TRACE": // OpenTelemetry TRACE = -8, no constant in slog, but slog allows extra levels
		return -8, nil
	case "DEBUG": // OpenTelemetry DEBUG = -4
		return stdoutLogExporter.LevelDebug, nil
	case "CONFIG": // Java JUL CONFIG, between DEBUG and INFO
		return -2, nil
	case "INFO": // OpenTelemetry INFO = 0
		return stdoutLogExporter.LevelInfo, nil
	case "NOTICE": // OpenTelemetry NOTICE = 2, no constant in slog, but slog allows extra levels
		return 2, nil
	case "WARN": // OpenTelemetry WARN = 4
		return stdoutLogExporter.LevelWarn, nil
	case "ERROR": // OpenTelemetry ERROR = 8
		return stdoutLogExporter.LevelError, nil
	case "FATAL": // OpenTelemetry FATAL = 12
		return 12, nil
	case "OFF": // Log4j OFF, disables logging, highest possible value
		return 16, nil
	default:
		return math.MaxInt, fmt.Errorf("invalid log level: %s", logLevelString)
	}
}
