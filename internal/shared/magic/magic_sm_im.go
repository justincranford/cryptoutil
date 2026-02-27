// Copyright (c) 2025 Justin Cranford

package magic

import (
	"time"
)

// SM-IM Service Magic Constants.
const (
	// OTLPServiceSMIM is the OTLP service name for sm-im telemetry.
	OTLPServiceSMIM = "sm-im"

	// IMServiceID is the canonical service identifier for sm-im.
	IMServiceID = OTLPServiceSMIM

	// IMProductName is the product name component of the SM product (same as SMProductName).
	IMProductName = SMProductName

	// IMServiceName is the service name component of the sm-im service.
	IMServiceName = "im"

	// IMServicePort is the default public HTTPS server port for sm-im.
	IMServicePort = 8700

	// IMAdminPort is the default private admin HTTPS server port for sm-im.
	IMAdminPort = 9090

	// IMDefaultTimeout is the default timeout for HTTP client operations.
	IMDefaultTimeout = 30 * time.Second

	// IMJWEAlgorithm is the default JWE algorithm for message encryption.
	// Uses direct key agreement with AES-256-GCM (dir+A256GCM).
	IMJWEAlgorithm = "dir+A256GCM"

	// IMJWEEncryption is the default JWE encryption algorithm.
	IMJWEEncryption = "A256GCM"

	// IMPBKDF2Iterations is the OWASP 2023 recommended iteration count for PBKDF2.
	IMPBKDF2Iterations = 600000
)

// User registration and authentication constraints.
const (
	// IMMinUsernameLength is the minimum acceptable username length.
	IMMinUsernameLength = 3

	// IMMaxUsernameLength is the maximum acceptable username length.
	IMMaxUsernameLength = 50

	// IMMinPasswordLength is the minimum acceptable password length.
	IMMinPasswordLength = 8

	// IMMaxTenantNameLength is the maximum acceptable tenant name length.
	IMMaxTenantNameLength = 100
)

// JWT token configuration.
const (
	// IMJWTIssuer is the issuer claim for JWT tokens.
	IMJWTIssuer = "sm-im"

	// IMJWTExpiration is the default JWT token expiration time.
	IMJWTExpiration = 24 * time.Hour
)

// Message validation constraints.
const (
	// IMMessageMinLength is the minimum message length in characters.
	IMMessageMinLength = 1

	// IMMessageMaxLength is the maximum message length in characters.
	IMMessageMaxLength = 10000

	// IMRecipientsMinCount is the minimum recipients per message.
	IMRecipientsMinCount = 1

	// IMRecipientsMaxCount is the maximum recipients per message.
	IMRecipientsMaxCount = 10
)

// Default realm password constraints.
const (
	// IMDefaultPasswordMinLength is the default realm minimum password length.
	IMDefaultPasswordMinLength = 12

	// IMDefaultPasswordMinUniqueChars is the default realm minimum unique characters in password.
	IMDefaultPasswordMinUniqueChars = 8

	// IMDefaultPasswordMaxRepeatedChars is the default realm maximum consecutive repeated characters.
	IMDefaultPasswordMaxRepeatedChars = 3
)

// Default realm session constraints (in seconds).
const (
	// IMDefaultSessionTimeout is the default realm session timeout (1 hour).
	IMDefaultSessionTimeout = 3600

	// IMDefaultSessionAbsoluteMax is the default realm absolute maximum session duration (24 hours).
	IMDefaultSessionAbsoluteMax = 86400
)

// Default realm rate limits (per minute).
const (
	// IMDefaultLoginRateLimit is the default realm login attempts per minute.
	IMDefaultLoginRateLimit = 5

	// IMDefaultMessageRateLimit is the default realm messages sent per minute.
	IMDefaultMessageRateLimit = 10
)

// Enterprise realm password constraints.
const (
	// IMEnterprisePasswordMinLength is the enterprise realm minimum password length.
	IMEnterprisePasswordMinLength = 16

	// IMEnterprisePasswordMinUniqueChars is the enterprise realm minimum unique characters in password.
	IMEnterprisePasswordMinUniqueChars = 12

	// IMEnterprisePasswordMaxRepeatedChars is the enterprise realm maximum consecutive repeated characters.
	IMEnterprisePasswordMaxRepeatedChars = 2
)

// Enterprise realm session constraints (in seconds).
const (
	// IMEnterpriseSessionTimeout is the enterprise realm session timeout (30 minutes).
	IMEnterpriseSessionTimeout = 1800

	// IMEnterpriseSessionAbsoluteMax is the enterprise realm absolute maximum session duration (8 hours).
	IMEnterpriseSessionAbsoluteMax = 28800
)

// Enterprise realm rate limits (per minute).
const (
	// IMEnterpriseLoginRateLimit is the enterprise realm login attempts per minute.
	IMEnterpriseLoginRateLimit = 3

	// IMEnterpriseMessageRateLimit is the enterprise realm messages sent per minute.
	IMEnterpriseMessageRateLimit = 5
)

// E2E Test Configuration.
const (
	// IME2EComposeFile is the path to the sm-im docker compose file (relative from e2e test directory).
	// Path: internal/apps/sm/im/e2e → ../../../../../deployments/sm-im/compose.yml
	// Levels: e2e→im(1)→sm(2)→apps(3)→internal(4)→cryptoutil(5), then deployments/sm-im.
	IME2EComposeFile = "../../../../../deployments/sm-im/compose.yml"

	// IME2ESQLiteContainer is the SQLite instance service name in compose.
	IME2ESQLiteContainer = "sm-im-app-sqlite-1"

	// IME2EPostgreSQL1Container is the PostgreSQL instance 1 service name in compose.
	IME2EPostgreSQL1Container = "sm-im-app-postgres-1"

	// IME2EPostgreSQL2Container is the PostgreSQL instance 2 service name in compose.
	IME2EPostgreSQL2Container = "sm-im-app-postgres-2"

	// IME2EOtelCollectorContainer is the OpenTelemetry Collector service name in compose.
	IME2EOtelCollectorContainer = "opentelemetry-collector-contrib"

	// IME2EGrafanaContainer is the Grafana LGTM container name.
	IME2EGrafanaContainer = "sm-im-grafana"

	// IME2EHealthTimeout is the timeout for health checks during E2E tests.
	// Must account for cascade dependencies: sqlite (30s) → pg-1 (30s) → pg-2 (30s) = 90s worst case.
	// Increased to 180s to handle slower CI/CD environments and Windows systems.
	IME2EHealthTimeout = 180 * time.Second

	// IME2EHealthPollInterval is the interval between health check attempts.
	IME2EHealthPollInterval = 2 * time.Second

	// IME2ESQLitePublicPort is the SQLite instance public HTTPS port.
	IME2ESQLitePublicPort = 8700

	// IME2EPostgreSQL1PublicPort is the PostgreSQL instance 1 public HTTPS port.
	IME2EPostgreSQL1PublicPort = 8701

	// IME2EPostgreSQL2PublicPort is the PostgreSQL instance 2 public HTTPS port.
	IME2EPostgreSQL2PublicPort = 8702

	// IME2EGrafanaPort is the Grafana UI port.
	IME2EGrafanaPort = 3000

	// IME2EOtelCollectorGRPCPort is the OTLP gRPC port.
	IME2EOtelCollectorGRPCPort = 4317

	// IME2EOtelCollectorHTTPPort is the OTLP HTTP port.
	IME2EOtelCollectorHTTPPort = 4318

	// IME2EHealthEndpoint is the public health check endpoint.
	// Uses /service/api/v1/health for headless client health checks (per 02-03.https-ports.instructions.md).
	IME2EHealthEndpoint = "/service/api/v1/health"
)

// SM-IM API path constants.
const (
	// IMAPIV1AuthRegister is the API v1 user registration path for sm-im.
	IMAPV1AuthRegister = "/api/v1/auth/register"

	// IMAPIV1Messages is the API v1 messages content path for sm-im.
	IMAPV1Messages = "/api/v1/messages"
)
