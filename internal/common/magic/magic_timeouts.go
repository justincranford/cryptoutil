// Package magic provides commonly used magic numbers and values as named constants.
// This file contains timeout and duration constants.
package magic

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
	// Timeout5Seconds - 5 seconds timeout for memory and host ID operations.
	Timeout5Seconds = 5
	// Timeout100Milliseconds - 100 milliseconds timeout for brief backoff operations.
	Timeout100Milliseconds = 100
)
