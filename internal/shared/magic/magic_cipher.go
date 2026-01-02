// Copyright (c) 2025 Justin Cranford

package magic

import "time"

// Cipher-IM Service Magic Constants.
const (
	// CipherServicePort is the default public HTTPS server port for cipher-im.
	CipherServicePort = 8888

	// CipherAdminPort is the default private admin HTTPS server port for cipher-im.
	CipherAdminPort = 9090

	// CipherDefaultTimeout is the default timeout for HTTP client operations.
	CipherDefaultTimeout = 30 * time.Second

	// CipherJWEAlgorithm is the default JWE algorithm for message encryption.
	// Uses direct key agreement with AES-256-GCM (dir+A256GCM).
	CipherJWEAlgorithm = "dir+A256GCM"

	// CipherJWEEncryption is the default JWE encryption algorithm.
	CipherJWEEncryption = "A256GCM"

	// CipherPBKDF2Iterations is the OWASP 2023 recommended iteration count for PBKDF2.
	CipherPBKDF2Iterations = 600000
)

// User registration and authentication constraints.
const (
	// CipherMinUsernameLength is the minimum acceptable username length.
	CipherMinUsernameLength = 3

	// CipherMaxUsernameLength is the maximum acceptable username length.
	CipherMaxUsernameLength = 50

	// CipherMinPasswordLength is the minimum acceptable password length.
	CipherMinPasswordLength = 8
)

// JWT token configuration.
const (
	// CipherJWTIssuer is the issuer claim for JWT tokens.
	CipherJWTIssuer = "cipher-im"

	// CipherJWTExpiration is the default JWT token expiration time.
	CipherJWTExpiration = 24 * time.Hour
)

// Message validation constraints.
const (
	// CipherMessageMinLength is the minimum message length in characters.
	CipherMessageMinLength = 1

	// CipherMessageMaxLength is the maximum message length in characters.
	CipherMessageMaxLength = 10000

	// CipherRecipientsMinCount is the minimum recipients per message.
	CipherRecipientsMinCount = 1

	// CipherRecipientsMaxCount is the maximum recipients per message.
	CipherRecipientsMaxCount = 10
)

// Default realm password constraints.
const (
	// CipherDefaultPasswordMinLength is the default realm minimum password length.
	CipherDefaultPasswordMinLength = 12

	// CipherDefaultPasswordMinUniqueChars is the default realm minimum unique characters in password.
	CipherDefaultPasswordMinUniqueChars = 8

	// CipherDefaultPasswordMaxRepeatedChars is the default realm maximum consecutive repeated characters.
	CipherDefaultPasswordMaxRepeatedChars = 3
)

// Default realm session constraints (in seconds).
const (
	// CipherDefaultSessionTimeout is the default realm session timeout (1 hour).
	CipherDefaultSessionTimeout = 3600

	// CipherDefaultSessionAbsoluteMax is the default realm absolute maximum session duration (24 hours).
	CipherDefaultSessionAbsoluteMax = 86400
)

// Default realm rate limits (per minute).
const (
	// CipherDefaultLoginRateLimit is the default realm login attempts per minute.
	CipherDefaultLoginRateLimit = 5

	// CipherDefaultMessageRateLimit is the default realm messages sent per minute.
	CipherDefaultMessageRateLimit = 10
)

// Enterprise realm password constraints.
const (
	// CipherEnterprisePasswordMinLength is the enterprise realm minimum password length.
	CipherEnterprisePasswordMinLength = 16

	// CipherEnterprisePasswordMinUniqueChars is the enterprise realm minimum unique characters in password.
	CipherEnterprisePasswordMinUniqueChars = 12

	// CipherEnterprisePasswordMaxRepeatedChars is the enterprise realm maximum consecutive repeated characters.
	CipherEnterprisePasswordMaxRepeatedChars = 2
)

// Enterprise realm session constraints (in seconds).
const (
	// CipherEnterpriseSessionTimeout is the enterprise realm session timeout (30 minutes).
	CipherEnterpriseSessionTimeout = 1800

	// CipherEnterpriseSessionAbsoluteMax is the enterprise realm absolute maximum session duration (8 hours).
	CipherEnterpriseSessionAbsoluteMax = 28800
)

// Enterprise realm rate limits (per minute).
const (
	// CipherEnterpriseLoginRateLimit is the enterprise realm login attempts per minute.
	CipherEnterpriseLoginRateLimit = 3

	// CipherEnterpriseMessageRateLimit is the enterprise realm messages sent per minute.
	CipherEnterpriseMessageRateLimit = 5
)
