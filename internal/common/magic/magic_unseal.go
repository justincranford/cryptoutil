package magic

import "time"

// Cryptographic algorithm and key constants.
// This file contains all crypto-related magic values used throughout the application.

const (
	// DefaultUnsealModeSysInfo - Sysinfo unseal mode.
	DefaultUnsealModeSysInfo = "sysinfo"
	// DefaultSysInfoAllTimeout - 10 seconds duration, rate limit maximum.
	DefaultSysInfoAllTimeout = 10 * time.Second //nolint:stylecheck // established API name
	// DefaultSysInfoCPUTimeout - Default system timeout duration.
	DefaultSysInfoCPUTimeout = 10 * time.Second //nolint:stylecheck // established API name
	// DefaultSysInfoMemoryTimeout - Default system memory timeout duration.
	DefaultSysInfoMemoryTimeout = 5 * time.Second //nolint:stylecheck // established API name
	// DefaultSysInfoHostIDTimeout - Default system host ID timeout duration.
	DefaultSysInfoHostTimeout = 5 * time.Second //nolint:stylecheck // established API namev

	// DefaultMaxUnsealFiles - Maximum number of files allowed.
	DefaultMaxUnsealFiles = 10
	// DefaultMaxBytesPerUnsealFile - Maximum bytes per file allowed.
	DefaultMaxBytesPerUnsealFile = 10 << 20 // 10MB

	// MaxUnsealSharedSecrets - Maximum number of shared secrets allowed.
	MaxUnsealSharedSecrets = 256
	// MinSharedSecretLength - Minimum shared secret length in bytes.
	MinSharedSecretLength = 32
	// MaxSharedSecretLength - Maximum shared secret length in bytes.
	MaxSharedSecretLength = 64

	// DerivedKeySizeBytes - Derived key size in bytes.
	DerivedKeySizeBytes = 32

	// RandomKeySizeBytes - Random bytes length for dev mode.
	RandomKeySizeBytes = 32
)
