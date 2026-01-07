// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"time"
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
		Name:             "authz",
		BindAddress:      "127.0.0.1",
		Port:             defaultAuthZPort,
		TLSEnabled:       false,
		TLSCertFile:      "",
		TLSKeyFile:       "",
		ReadTimeout:      defaultReadTimeoutSeconds * time.Second,
		WriteTimeout:     defaultWriteTimeoutSeconds * time.Second,
		IdleTimeout:      defaultIdleTimeoutSeconds * time.Second,
		AdminEnabled:     true,
		AdminBindAddress: "127.0.0.1",
		AdminPort:        defaultAuthZAdminPort,
	}
}

func defaultIDPConfig() *ServerConfig {
	return &ServerConfig{
		Name:             "idp",
		BindAddress:      "127.0.0.1",
		Port:             defaultIDPPort,
		TLSEnabled:       false,
		TLSCertFile:      "",
		TLSKeyFile:       "",
		ReadTimeout:      defaultReadTimeoutSeconds * time.Second,
		WriteTimeout:     defaultWriteTimeoutSeconds * time.Second,
		IdleTimeout:      defaultIdleTimeoutSeconds * time.Second,
		AdminEnabled:     true,
		AdminBindAddress: "127.0.0.1",
		AdminPort:        defaultIDPAdminPort,
	}
}

func defaultRSConfig() *ServerConfig {
	return &ServerConfig{
		Name:             "rs",
		BindAddress:      "127.0.0.1",
		Port:             defaultRSPort,
		TLSEnabled:       false,
		TLSCertFile:      "",
		TLSKeyFile:       "",
		ReadTimeout:      defaultReadTimeoutSeconds * time.Second,
		WriteTimeout:     defaultWriteTimeoutSeconds * time.Second,
		IdleTimeout:      defaultIdleTimeoutSeconds * time.Second,
		AdminEnabled:     false,
		AdminBindAddress: "127.0.0.1",
		AdminPort:        defaultRSAdminPort,
	}
}

func defaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:            "sqlite",
		DSN:             ":memory:",
		MaxOpenConns:    defaultMaxOpenConns,
		MaxIdleConns:    defaultMaxIdleConns,
		ConnMaxLifetime: defaultConnMaxLifetimeMin * time.Minute,
		ConnMaxIdleTime: defaultConnMaxIdleTimeMin * time.Minute,
		AutoMigrate:     true,
	}
}

func defaultTokenConfig() *TokenConfig {
	return &TokenConfig{
		AccessTokenLifetime:  defaultAccessTokenLifetimeSeconds * time.Second,
		RefreshTokenLifetime: defaultRefreshTokenLifetimeSeconds * time.Second,
		IDTokenLifetime:      defaultIDTokenLifetimeSeconds * time.Second,
		AccessTokenFormat:    tokenFormatJWS,
		RefreshTokenFormat:   tokenFormatUUID,
		IDTokenFormat:        tokenFormatJWS,
		Issuer:               "https://identity.example.com",
		SigningAlgorithm:     "RS256",
		SigningKeyID:         "",
		EncryptionEnabled:    false,
	}
}

func defaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		SessionLifetime: defaultSessionLifetimeSeconds * time.Second,
		IdleTimeout:     defaultIdleTimeoutSecondsSession * time.Second,
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
		PKCEChallengeMethod: "S256",
		RequireState:        true,
		RateLimitEnabled:    true,
		RateLimitRequests:   defaultRateLimitRequests,
		RateLimitWindow:     defaultRateLimitWindowSeconds * time.Second,
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
