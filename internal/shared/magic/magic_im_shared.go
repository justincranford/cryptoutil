// Copyright (c) 2025-2026 Justin Cranford.
//

package magic

import "time"

// Shared username/password constraints.
const (
	IMMinUsernameLength   = 3
	IMMaxUsernameLength   = 50
	IMMinPasswordLength   = 8
	IMMaxTenantNameLength = 100
)

// Shared JWT/session defaults.
const (
	IMJWTIssuer                 = "sm-kms"
	IMJWTExpiration             = 24 * time.Hour
	IMDefaultTimeout            = 30 * time.Second
	IMDefaultSessionTimeout     = 3600
	IMDefaultSessionAbsoluteMax = 86400
)

// Shared realm password policy defaults.
const (
	IMDefaultPasswordMinLength        = 12
	IMDefaultPasswordMinUniqueChars   = 8
	IMDefaultPasswordMaxRepeatedChars = 3
)

// Default and enterprise realm rate limits and constraints.
const (
	IMDefaultLoginRateLimit              = 5
	IMDefaultMessageRateLimit            = 10
	IMEnterprisePasswordMinLength        = 16
	IMEnterprisePasswordMinUniqueChars   = 12
	IMEnterprisePasswordMaxRepeatedChars = 2
	IMEnterpriseSessionTimeout           = 1800
	IMEnterpriseSessionAbsoluteMax       = 28800
	IMEnterpriseLoginRateLimit           = 3
	IMEnterpriseMessageRateLimit         = 5
)

// Shared message constraints.
const (
	IMMessageMinLength          = 1
	IMMessageMaxLength          = 10000
	IMRecipientsMinCount        = 1
	IMRecipientsMaxCount        = 10
	IME2EOtelCollectorContainer = "opentelemetry-collector-contrib"
)
