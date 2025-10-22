// Package magic provides commonly used magic numbers and values as named constants.
// This file contains count and limit constants.
package magic

// Common counts and limits.
const (
	// CountMinimumCLIArgs - Minimum CLI arguments, common small count.
	CountMinimumCLIArgs = 2
	// CountUIProgressInterval - Progress reporting interval for UI operations.
	CountUIProgressInterval = 10
	// CountMinActionMatchGroups - Minimum number of regex match groups for action parsing.
	CountMinActionMatchGroups = 3

	// CountDefaultPageSize - Default page size for pagination.
	CountDefaultPageSize = 25
	// CountMaxLogsBatchSize - Maximum batch size for logs.
	CountMaxLogsBatchSize = 1024
	// CountMaxMetricsBatchSize - Maximum batch size for metrics.
	CountMaxMetricsBatchSize = 2048
	// CountMaxTracesBatchSize - Maximum batch size for traces.
	CountMaxTracesBatchSize = 512
	// CountCertificateRandomizationRangeMinutes - Certificate validity randomization range in minutes.
	CountCertificateRandomizationRangeMinutes = 120
	// CountCORSMaxAge - Default CORS max age in seconds.
	CountCORSMaxAge uint16 = 3600
	// CountRequestBodyLimit - Default request body limit in bytes (2MB).
	CountRequestBodyLimit = 2 << 20
	// CountMaxFiles - Maximum number of files allowed.
	CountMaxFiles = 10
	// CountMaxBytesPerFile - Maximum bytes per file allowed.
	CountMaxBytesPerFile = 10 << 20 // 10MB
	// CountDevModeRandomBytesLength - Random bytes length for dev mode.
	CountDevModeRandomBytesLength = 32
)

// UI display constants.
const (
	// UIConsoleSeparatorLength - Length of console separator lines.
	UIConsoleSeparatorLength = 50
)
