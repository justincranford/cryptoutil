// Package magic provides commonly used magic numbers and values as named constants.
// This package centralizes magic values to avoid linter violations and improve code maintainability.
// All constants are grouped by category for better organization.
package magic

// Buffer sizes and memory allocations.
const (
	// BufferSize1KB - 1KB buffer size, common memory allocation.
	BufferSize1KB = 1024
	// BufferSize2KB - 2KB buffer size, RSA-2048 key size.
	BufferSize2KB = 2048
	// BufferSize4KB - 4KB buffer size, RSA-4096 key size.
	BufferSize4KB = 4096
)

// Timeouts and durations (in milliseconds unless otherwise noted).
const (
	// Timeout1SecondMs - 1 second in milliseconds, common timeout unit.
	Timeout1SecondMs = 1000
	// Timeout10SecondsMs - 10 seconds in milliseconds, rate limit maximum.
	Timeout10SecondsMs = 10000
	// Timeout1MinuteSeconds - 1 minute in seconds, common timeout.
	Timeout1MinuteSeconds = 60
	// Timeout10Seconds - 10 seconds timeout for system info operations.
	Timeout10Seconds = 10
)

// Percentages and tolerances (as decimal values).
const (
	// Tolerance5Percent - 5% tolerance, common percentage value.
	Tolerance5Percent = 0.05
	// Tolerance10Percent - 10% tolerance, common percentage value.
	Tolerance10Percent = 0.1
	// Tolerance50Percent - 50% tolerance, common percentage value.
	Tolerance50Percent = 0.5
)

// Percentage basis values.
const (
	// PercentageBasis10 - Percentage basis, common percentage value.
	PercentageBasis10 = 10
	// PercentageBasis100 - Percentage basis, common percentage value.
	PercentageBasis100 = 100
)

// Network ports.
const (
	// PortHTTPS - Standard HTTPS port.
	PortHTTPS = 443
	// PortDefaultBrowserAPI - Default browser/server API port.
	PortDefaultBrowserAPI = 8080
	// PortDefaultAdminAPI - Default admin API port.
	PortDefaultAdminAPI = 9090
)

// Common counts and limits.
const (
	// CountMinimumCLIArgs - Minimum CLI arguments, common small count.
	CountMinimumCLIArgs = 2
)

// Miscellaneous constants.
const (
	// AnswerToLifeUniverseEverything - Answer to life, the universe, and everything.
	AnswerToLifeUniverseEverything = 42
)
