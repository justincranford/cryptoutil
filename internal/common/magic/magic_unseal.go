package magic

// Cryptographic algorithm and key constants.
// This file contains all crypto-related magic values used throughout the application.

const (
	// StringUnsealModeSysinfo - Sysinfo unseal mode.
	StringUnsealModeSysinfo = "sysinfo"

	// CountMaxUnsealFiles - Maximum number of files allowed.
	CountMaxUnsealFiles = 10
	// CountMaxBytesPerUnsealFile - Maximum bytes per file allowed.
	CountMaxBytesPerUnsealFile = 10 << 20 // 10MB

	// CountMaxUnsealSharedSecrets - Maximum number of shared secrets allowed.
	CountMaxUnsealSharedSecrets = 256

	// CountMinSharedSecretLength - Minimum shared secret length in bytes.
	CountMinSharedSecretLength = 32
	// CountMaxSharedSecretLength - Maximum shared secret length in bytes.
	CountMaxSharedSecretLength = 64

	// DefaultDerivedKeySizeBytes - Derived key size in bytes.
	DefaultDerivedKeySizeBytes = 32

	// DefaultRandomKeySizeBytes - Random bytes length for dev mode.
	DefaultRandomKeySizeBytes = 32
)
