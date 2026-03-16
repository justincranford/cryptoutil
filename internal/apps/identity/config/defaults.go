// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// DefaultConfig returns default configuration values.
func DefaultConfig() *Config {
	return &Config{
		AuthZ:         defaultAuthZConfig(),
		IDP:           defaultIDPConfig(),
		RS:            defaultRSConfig(),
		Database:      defaultDatabaseConfig(),
		Tokens:        defaultTokenConfig(),
		Sessions:      defaultSessionConfig(),
		Security:      defaultSecurityConfig(),
		Observability: defaultObservabilityConfig(),
	}
}

func defaultAuthZConfig() *ServerConfig {
	return &ServerConfig{
		Name:             cryptoutilSharedMagic.AuthzServiceName,
		BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
		Port:             cryptoutilSharedMagic.IdentityDefaultAuthZPort,
		TLSEnabled:       false,
		TLSCertFile:      "",
		TLSKeyFile:       "",
		ReadTimeout:      cryptoutilSharedMagic.IdentityDefaultReadTimeoutSeconds * time.Second,
		WriteTimeout:     cryptoutilSharedMagic.IdentityDefaultWriteTimeoutSeconds * time.Second,
		IdleTimeout:      cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds * time.Second,
		AdminEnabled:     true,
		AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
		AdminPort:        cryptoutilSharedMagic.IdentityDefaultAuthZAdminPort,
	}
}

func defaultIDPConfig() *ServerConfig {
	return &ServerConfig{
		Name:             cryptoutilSharedMagic.IDPServiceName,
		BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
		Port:             cryptoutilSharedMagic.IdentityDefaultIDPPort,
		TLSEnabled:       false,
		TLSCertFile:      "",
		TLSKeyFile:       "",
		ReadTimeout:      cryptoutilSharedMagic.IdentityDefaultReadTimeoutSeconds * time.Second,
		WriteTimeout:     cryptoutilSharedMagic.IdentityDefaultWriteTimeoutSeconds * time.Second,
		IdleTimeout:      cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds * time.Second,
		AdminEnabled:     true,
		AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
		AdminPort:        cryptoutilSharedMagic.IdentityDefaultIDPAdminPort,
	}
}

func defaultRSConfig() *ServerConfig {
	return &ServerConfig{
		Name:             "rs",
		BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
		Port:             cryptoutilSharedMagic.IdentityDefaultRSPort,
		TLSEnabled:       false,
		TLSCertFile:      "",
		TLSKeyFile:       "",
		ReadTimeout:      cryptoutilSharedMagic.IdentityDefaultReadTimeoutSeconds * time.Second,
		WriteTimeout:     cryptoutilSharedMagic.IdentityDefaultWriteTimeoutSeconds * time.Second,
		IdleTimeout:      cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds * time.Second,
		AdminEnabled:     false,
		AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
		AdminPort:        cryptoutilSharedMagic.IdentityDefaultRSAdminPort,
	}
}

func defaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:            "sqlite",
		DSN:             cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
		MaxOpenConns:    cryptoutilSharedMagic.IdentityDefaultMaxOpenConns,
		MaxIdleConns:    cryptoutilSharedMagic.IdentityDefaultMaxIdleConns,
		ConnMaxLifetime: cryptoutilSharedMagic.IdentityDefaultConnMaxLifetimeMin * time.Minute,
		ConnMaxIdleTime: cryptoutilSharedMagic.IdentityDefaultConnMaxIdleTimeMin * time.Minute,
		AutoMigrate:     true,
	}
}

func defaultTokenConfig() *TokenConfig {
	return &TokenConfig{
		AccessTokenLifetime:  cryptoutilSharedMagic.IdentityDefaultAccessTokenLifetimeSeconds * time.Second,
		RefreshTokenLifetime: cryptoutilSharedMagic.IdentityDefaultRefreshTokenLifetimeSeconds * time.Second,
		IDTokenLifetime:      cryptoutilSharedMagic.IdentityDefaultIDTokenLifetimeSeconds * time.Second,
		AccessTokenFormat:    cryptoutilSharedMagic.IdentityTokenFormatJWS,
		RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
		IDTokenFormat:        cryptoutilSharedMagic.IdentityTokenFormatJWS,
		Issuer:               "https://identity.example.com",
		SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		SigningKeyID:         "",
		EncryptionEnabled:    false,
	}
}

func defaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		SessionLifetime: cryptoutilSharedMagic.IdentityDefaultSessionLifetimeSeconds * time.Second,
		IdleTimeout:     cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSecondsSession * time.Second,
		CookieName:      "identity_session",
		CookieDomain:    "",
		CookiePath:      "/",
		CookieSecure:    true,
		CookieHTTPOnly:  true,
		CookieSameSite:  "Lax",
	}
}

func defaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RequirePKCE:         true,
		PKCEChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
		RequireState:        true,
		RateLimitEnabled:    true,
		RateLimitRequests:   cryptoutilSharedMagic.IdentityDefaultRateLimitRequests,
		RateLimitWindow:     cryptoutilSharedMagic.IdentityDefaultRateLimitWindowSeconds * time.Second,
		CORSEnabled:         true,
		CORSAllowedOrigins:  []string{"https://localhost:3000"},
		CSRFEnabled:         true,
	}
}

func defaultObservabilityConfig() *ObservabilityConfig {
	return &ObservabilityConfig{
		LogLevel:       "info",
		LogFormat:      "json",
		MetricsEnabled: true,
		MetricsPath:    "/metrics",
		TracingEnabled: true,
		TracingBackend: "otlp",
	}
}
