package listener

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
)

func TestTLSRootCAPoolAccessors(t *testing.T) {
	t.Parallel()

	rootCAPool := x509.NewCertPool()
	tests := []struct {
		name    string
		get     func(*cryptoutilAppsFrameworkServiceConfig.TLSMaterial) *x509.CertPool
		wantNil bool
	}{
		{
			name: "admin accessor without TLS material",
			get: func(material *cryptoutilAppsFrameworkServiceConfig.TLSMaterial) *x509.CertPool {
				return (&AdminServer{tlsMaterial: material}).AdminTLSRootCAPool()
			},
			wantNil: true,
		},
		{
			name: "public accessor without TLS material",
			get: func(material *cryptoutilAppsFrameworkServiceConfig.TLSMaterial) *x509.CertPool {
				return (&PublicHTTPServer{tlsMaterial: material}).PublicTLSRootCAPool()
			},
			wantNil: true,
		},
		{
			name: "admin accessor returns configured root pool",
			get: func(material *cryptoutilAppsFrameworkServiceConfig.TLSMaterial) *x509.CertPool {
				return (&AdminServer{tlsMaterial: material}).AdminTLSRootCAPool()
			},
		},
		{
			name: "public accessor returns configured root pool",
			get: func(material *cryptoutilAppsFrameworkServiceConfig.TLSMaterial) *x509.CertPool {
				return (&PublicHTTPServer{tlsMaterial: material}).PublicTLSRootCAPool()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			material := &cryptoutilAppsFrameworkServiceConfig.TLSMaterial{RootCAPool: rootCAPool}
			if tc.wantNil {
				material = nil
			}

			got := tc.get(material)
			if tc.wantNil {
				require.Nil(t, got)

				return
			}

			require.Same(t, rootCAPool, got)
		})
	}
}
