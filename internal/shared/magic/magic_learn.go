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
