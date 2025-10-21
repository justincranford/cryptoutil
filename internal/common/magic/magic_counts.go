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
	// CountExpectedSysInfos - Expected number of system info items.
	CountExpectedSysInfos = 13
	// CountMaxSharedSecrets - Maximum number of shared secrets allowed.
	CountMaxSharedSecrets = 256
	// CountMinSharedSecretLength - Minimum shared secret length in bytes.
	CountMinSharedSecretLength = 32
	// CountMaxSharedSecretLength - Maximum shared secret length in bytes.
	CountMaxSharedSecretLength = 64
	// CountDerivedKeySizeBytes - Derived key size in bytes.
	CountDerivedKeySizeBytes = 32
	// CountDefaultPageSize - Default page size for pagination.
	CountDefaultPageSize = 25
	// CountMaxLogsBatchSize - Maximum batch size for logs.
	CountMaxLogsBatchSize = 1024
	// CountMaxMetricsBatchSize - Maximum batch size for metrics.
	CountMaxMetricsBatchSize = 2048
	// CountMaxTracesBatchSize - Maximum batch size for traces.
	CountMaxTracesBatchSize = 512
)

// UI display constants.
const (
	// UIConsoleSeparatorLength - Length of console separator lines.
	UIConsoleSeparatorLength = 50
)
