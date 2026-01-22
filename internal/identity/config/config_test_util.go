// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"time"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// RequireNewForTest creates a new identity config for testing.
func RequireNewForTest(_ string) *Config {
	return &Config{
		AuthZ: &ServerConfig{
			Name:             "authz",
			BindAddress:      cryptoutilMagic.IPv4Loopback,
			Port:             int(cryptoutilMagic.DefaultPublicPortCryptoutil),
			TLSEnabled:       true,
			TLSCertFile:      "",
			TLSKeyFile:       "",
			ReadTimeout:      cryptoutilMagic.TestTimeoutCryptoutilReady,
			WriteTimeout:     cryptoutilMagic.TestTimeoutCryptoutilReady,
			IdleTimeout:      cryptoutilMagic.TimeoutTestServerReady,
			AdminEnabled:     true,
			AdminBindAddress: cryptoutilMagic.IPv4Loopback,
			AdminPort:        int(cryptoutilMagic.DefaultPrivatePortCryptoutil),
		},
		IDP: &ServerConfig{
			Name:             "idp",
			BindAddress:      cryptoutilMagic.IPv4Loopback,
			Port:             int(cryptoutilMagic.DefaultPublicPortCryptoutilCompose1),
			TLSEnabled:       true,
			TLSCertFile:      "",
			TLSKeyFile:       "",
			ReadTimeout:      cryptoutilMagic.TestTimeoutCryptoutilReady,
			WriteTimeout:     cryptoutilMagic.TestTimeoutCryptoutilReady,
			IdleTimeout:      cryptoutilMagic.TimeoutTestServerReady,
			AdminEnabled:     true,
			AdminBindAddress: cryptoutilMagic.IPv4Loopback,
			AdminPort:        int(cryptoutilMagic.DefaultPrivatePortCryptoutil) + 1,
		},
		RS: &ServerConfig{
			Name:             "rs",
			BindAddress:      cryptoutilMagic.IPv4Loopback,
			Port:             int(cryptoutilMagic.DefaultPublicPortCryptoutilCompose2),
			TLSEnabled:       true,
			TLSCertFile:      "",
			TLSKeyFile:       "",
			ReadTimeout:      cryptoutilMagic.TestTimeoutCryptoutilReady,
			WriteTimeout:     cryptoutilMagic.TestTimeoutCryptoutilReady,
			IdleTimeout:      cryptoutilMagic.TimeoutTestServerReady,
			AdminEnabled:     true,
			AdminBindAddress: cryptoutilMagic.IPv4Loopback,
			AdminPort:        int(cryptoutilMagic.DefaultPrivatePortCryptoutil) + 2,
		},
		Database: &DatabaseConfig{
			Type:         "sqlite",
			DSN:          ":memory:",
			MaxOpenConns: cryptoutilMagic.SQLiteMaxOpenConnections,
		},
		Tokens: &TokenConfig{
			AccessTokenLifetime:  cryptoutilMagic.TestTimeoutCryptoutilReady,
			RefreshTokenLifetime: cryptoutilIdentityMagic.TestRefreshTokenLifetime,
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
			RateLimitRequests:  int(cryptoutilMagic.DefaultPublicBrowserAPIIPRateLimit),
			RateLimitWindow:    time.Minute,
		},
		Observability: &ObservabilityConfig{
			LogLevel: "info",
		},
	}
}
