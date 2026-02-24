// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// RequireNewForTest creates a new identity config for testing.
func RequireNewForTest(_ string) *Config {
	return &Config{
		AuthZ: &ServerConfig{
			Name:             "authz",
			BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
			Port:             int(cryptoutilSharedMagic.DefaultPublicPortCryptoutil),
			TLSEnabled:       true,
			TLSCertFile:      "",
			TLSKeyFile:       "",
			ReadTimeout:      cryptoutilSharedMagic.TestTimeoutCryptoutilReady,
			WriteTimeout:     cryptoutilSharedMagic.TestTimeoutCryptoutilReady,
			IdleTimeout:      cryptoutilSharedMagic.TimeoutTestServerReady,
			AdminEnabled:     true,
			AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
			AdminPort:        int(cryptoutilSharedMagic.DefaultPrivatePortCryptoutil),
		},
		IDP: &ServerConfig{
			Name:             "idp",
			BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
			Port:             int(cryptoutilSharedMagic.DefaultPublicPortCryptoutilCompose1),
			TLSEnabled:       true,
			TLSCertFile:      "",
			TLSKeyFile:       "",
			ReadTimeout:      cryptoutilSharedMagic.TestTimeoutCryptoutilReady,
			WriteTimeout:     cryptoutilSharedMagic.TestTimeoutCryptoutilReady,
			IdleTimeout:      cryptoutilSharedMagic.TimeoutTestServerReady,
			AdminEnabled:     true,
			AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
			AdminPort:        int(cryptoutilSharedMagic.DefaultPrivatePortCryptoutil) + 1,
		},
		RS: &ServerConfig{
			Name:             "rs",
			BindAddress:      cryptoutilSharedMagic.IPv4Loopback,
			Port:             int(cryptoutilSharedMagic.DefaultPublicPortCryptoutilCompose2),
			TLSEnabled:       true,
			TLSCertFile:      "",
			TLSKeyFile:       "",
			ReadTimeout:      cryptoutilSharedMagic.TestTimeoutCryptoutilReady,
			WriteTimeout:     cryptoutilSharedMagic.TestTimeoutCryptoutilReady,
			IdleTimeout:      cryptoutilSharedMagic.TimeoutTestServerReady,
			AdminEnabled:     true,
			AdminBindAddress: cryptoutilSharedMagic.IPv4Loopback,
			AdminPort:        int(cryptoutilSharedMagic.DefaultPrivatePortCryptoutil) + 2,
		},
		Database: &DatabaseConfig{
			Type:         "sqlite",
			DSN:          ":memory:",
			MaxOpenConns: cryptoutilSharedMagic.SQLiteMaxOpenConnections,
		},
		Tokens: &TokenConfig{
			AccessTokenLifetime:  cryptoutilSharedMagic.TestTimeoutCryptoutilReady,
			RefreshTokenLifetime: cryptoutilSharedMagic.TestRefreshTokenLifetime,
			Issuer:               "test-issuer",
		},
		Sessions: &SessionConfig{
			CookieName:     "test_session",
			CookieSecure:   false,
			CookieHTTPOnly: true,
			CookieSameSite: "lax",
		},
		Security: &SecurityConfig{
			CORSAllowedOrigins: []string{"*"},
			RateLimitEnabled:   false,
			RateLimitRequests:  int(cryptoutilSharedMagic.DefaultPublicBrowserAPIIPRateLimit),
			RateLimitWindow:    time.Minute,
		},
		Observability: &ObservabilityConfig{
			LogLevel: "info",
		},
	}
}
