// Copyright (c) 2025 Justin Cranford

package magic

import "time"

// Learn-IM Service Magic Constants.
const (
	// LearnServicePort is the default public HTTPS server port for learn-im.
	LearnServicePort = 8888

	// LearnAdminPort is the default private admin HTTPS server port for learn-im.
	LearnAdminPort = 9090

	// LearnDefaultTimeout is the default timeout for HTTP client operations.
	LearnDefaultTimeout = 30 * time.Second

	// LearnJWEAlgorithm is the default JWE algorithm for message encryption.
	// Uses direct key agreement with AES-256-GCM (dir+A256GCM).
	LearnJWEAlgorithm = "dir+A256GCM"

	// LearnJWEEncryption is the default JWE encryption algorithm.
	LearnJWEEncryption = "A256GCM"

	// LearnPBKDF2Iterations is the OWASP 2023 recommended iteration count for PBKDF2.
	LearnPBKDF2Iterations = 600000
)

// User registration and authentication constraints.
const (
	// LearnMinUsernameLength is the minimum acceptable username length.
	LearnMinUsernameLength = 3

	// LearnMaxUsernameLength is the maximum acceptable username length.
	LearnMaxUsernameLength = 50

	// LearnMinPasswordLength is the minimum acceptable password length.
	LearnMinPasswordLength = 8
)

// JWT token configuration.
const (
	// LearnJWTIssuer is the issuer claim for JWT tokens.
	LearnJWTIssuer = "learn-im"

	// LearnJWTExpiration is the default JWT token expiration time.
	LearnJWTExpiration = 24 * time.Hour
)

// Message validation constraints.
const (
	// LearnMessageMinLength is the minimum message length in characters.
	LearnMessageMinLength = 1

	// LearnMessageMaxLength is the maximum message length in characters.
	LearnMessageMaxLength = 10000

	// LearnRecipientsMinCount is the minimum recipients per message.
	LearnRecipientsMinCount = 1

	// LearnRecipientsMaxCount is the maximum recipients per message.
	LearnRecipientsMaxCount = 10
)

// Default realm password constraints.
const (
	// LearnDefaultPasswordMinLength is the default realm minimum password length.
	LearnDefaultPasswordMinLength = 12

	// LearnDefaultPasswordMinUniqueChars is the default realm minimum unique characters in password.
	LearnDefaultPasswordMinUniqueChars = 8

	// LearnDefaultPasswordMaxRepeatedChars is the default realm maximum consecutive repeated characters.
	LearnDefaultPasswordMaxRepeatedChars = 3
)

// Default realm session constraints (in seconds).
const (
	// LearnDefaultSessionTimeout is the default realm session timeout (1 hour).
	LearnDefaultSessionTimeout = 3600

	// LearnDefaultSessionAbsoluteMax is the default realm absolute maximum session duration (24 hours).
	LearnDefaultSessionAbsoluteMax = 86400
)

// Default realm rate limits (per minute).
const (
	// LearnDefaultLoginRateLimit is the default realm login attempts per minute.
	LearnDefaultLoginRateLimit = 5

	// LearnDefaultMessageRateLimit is the default realm messages sent per minute.
	LearnDefaultMessageRateLimit = 10
)

// Enterprise realm password constraints.
const (
	// LearnEnterprisePasswordMinLength is the enterprise realm minimum password length.
	LearnEnterprisePasswordMinLength = 16

	// LearnEnterprisePasswordMinUniqueChars is the enterprise realm minimum unique characters in password.
	LearnEnterprisePasswordMinUniqueChars = 12

	// LearnEnterprisePasswordMaxRepeatedChars is the enterprise realm maximum consecutive repeated characters.
	LearnEnterprisePasswordMaxRepeatedChars = 2
)

// Enterprise realm session constraints (in seconds).
const (
	// LearnEnterpriseSessionTimeout is the enterprise realm session timeout (30 minutes).
	LearnEnterpriseSessionTimeout = 1800

	// LearnEnterpriseSessionAbsoluteMax is the enterprise realm absolute maximum session duration (8 hours).
	LearnEnterpriseSessionAbsoluteMax = 28800
)

// Enterprise realm rate limits (per minute).
const (
	// LearnEnterpriseLoginRateLimit is the enterprise realm login attempts per minute.
	LearnEnterpriseLoginRateLimit = 3

	// LearnEnterpriseMessageRateLimit is the enterprise realm messages sent per minute.
	LearnEnterpriseMessageRateLimit = 5
)
