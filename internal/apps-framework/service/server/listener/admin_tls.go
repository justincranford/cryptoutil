// Copyright (c) 2025-2026 Justin Cranford.
package listener

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// applyAdminMTLS applies admin TLS file overrides and explicit client-certificate policy.
// Server certificate overrides and client-certificate policy are independent concerns.
func applyAdminMTLS(
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	tlsMaterial *cryptoutilAppsFrameworkServiceConfig.TLSMaterial,
	osReadFileFn func(name string) ([]byte, error),
) error {
	// Override server cert from Cat 7 file paths when both cert and key are configured.
	if settings.AdminTLSCertFile != "" && settings.AdminTLSKeyFile != "" {
		certPEM, err := osReadFileFn(settings.AdminTLSCertFile)
		if err != nil {
			return fmt.Errorf("failed to read admin TLS cert file %q: %w", settings.AdminTLSCertFile, err)
		}

		keyPEM, err := osReadFileFn(settings.AdminTLSKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read admin TLS key file %q: %w", settings.AdminTLSKeyFile, err)
		}

		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return fmt.Errorf("failed to parse admin TLS cert+key pair: %w", err)
		}

		tlsMaterial.Config.Certificates = []tls.Certificate{cert}
	}

	clientAuth, err := cryptoutilSharedCryptoTls.ClientAuthTypeForPolicy(settings.AdminTLSClientPolicy)
	if err != nil {
		return fmt.Errorf("invalid admin TLS client policy %q: %w", settings.AdminTLSClientPolicy, err)
	}

	// Load client trust material from the Cat 6 CA truststore when configured.
	if settings.AdminTLSCAFile != "" {
		caPEM, err := osReadFileFn(settings.AdminTLSCAFile)
		if err != nil {
			return fmt.Errorf("failed to read admin TLS CA file %q: %w", settings.AdminTLSCAFile, err)
		}

		clientCAPool := x509.NewCertPool()

		for rest := caPEM; len(rest) > 0; {
			var block *pem.Block

			block, rest = pem.Decode(rest)
			if block == nil {
				break
			}

			if block.Type != cryptoutilSharedMagic.StringPEMTypeCertificate {
				continue
			}

			caCert, parseErr := x509.ParseCertificate(block.Bytes)
			if parseErr != nil {
				return fmt.Errorf("failed to parse CA certificate from %q: %w", settings.AdminTLSCAFile, parseErr)
			}

			clientCAPool.AddCert(caCert)
		}

		tlsMaterial.Config.ClientCAs = clientCAPool
	}

	if (clientAuth == tls.VerifyClientCertIfGiven || clientAuth == tls.RequireAndVerifyClientCert) && settings.AdminTLSCAFile == "" {
		return fmt.Errorf("admin TLS client policy %q requires a CA file", settings.AdminTLSClientPolicy)
	}

	tlsMaterial.Config.ClientAuth = clientAuth

	return nil
}
