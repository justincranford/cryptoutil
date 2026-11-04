// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"encoding/base64"
	json "encoding/json"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testKubeNonCA         = "ca-nonca"
	testKubeExpiredCA     = "ca-expired"
	testKubeFutureCA      = "ca-future"
	testKubeDER           = "ca-der"
	testKubeExpiredClient = "client-expired"
	testKubeFutureClient  = "client-future"
	testKubeCAClient      = "client-ca"
	testKubeServerClient  = "client-server-only"
	testKubeKeyEncoding   = "client-key-encoding"
)

func TestDevelopmentKubeconfigRejectsInvalidCAAndClientPolicy(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{testKubeNonCA, testKubeExpiredCA, testKubeFutureCA, testKubeDER, testKubeExpiredClient, testKubeFutureClient, testKubeCAClient, testKubeServerClient, testKubeKeyEncoding} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()

			var config map[string]any
			require.NoError(t, json.Unmarshal(developmentKubeconfigIdentityFixture(t), &config))
			clusters, ok := config[developmentKubeClusters].([]any)
			require.True(t, ok)
			clusterEntry, ok := clusters[0].(map[string]any)
			require.True(t, ok)
			cluster, ok := clusterEntry[testDevelopmentClusterField].(map[string]any)
			require.True(t, ok)
			users, ok := config[developmentKubeUsers].([]any)
			require.True(t, ok)
			userEntry, ok := users[0].(map[string]any)
			require.True(t, ok)
			user, ok := userEntry[developmentKubeUser].(map[string]any)
			require.True(t, ok)

			key, err := ecdsa.GenerateKey(elliptic.P384(), crand.Reader)
			require.NoError(t, err)

			now := time.Now().UTC()
			template := &x509.Certificate{
				SerialNumber: big.NewInt(1), NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
				BasicConstraintsValid: true, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			ca := strings.HasPrefix(mode, "ca-")
			if ca || mode == testKubeCAClient {
				template.IsCA, template.KeyUsage = true, x509.KeyUsageCertSign
			}

			switch mode {
			case testKubeNonCA:
				template.IsCA = false
			case testKubeExpiredCA, testKubeExpiredClient:
				template.NotBefore, template.NotAfter = now.Add(-2*time.Hour), now.Add(-time.Hour)
			case testKubeFutureCA, testKubeFutureClient:
				template.NotBefore, template.NotAfter = now.Add(time.Hour), now.Add(2*time.Hour)
			case testKubeServerClient:
				template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
			}

			der, err := x509.CreateCertificate(crand.Reader, template, template, &key.PublicKey, key)
			require.NoError(t, err)

			if mode == testKubeDER {
				der = []byte("not a certificate")
			}

			certificate := pem.EncodeToMemory(&pem.Block{Type: testPKICertificatePEM, Bytes: der})
			if ca {
				cluster[developmentKubeCA] = base64.StdEncoding.EncodeToString(certificate)
			} else {
				encoded, err := x509.MarshalPKCS8PrivateKey(key)
				require.NoError(t, err)

				user[developmentKubeCertificate] = base64.StdEncoding.EncodeToString(certificate)

				user[developmentKubeKey] = base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: encoded}))
				if mode == testKubeKeyEncoding {
					user[developmentKubeKey] = "%%%invalid-client-key-encoding%%%"
				}
			}

			raw, err := json.Marshal(config)
			require.NoError(t, err)
			_, err = developmentKubeconfig(raw, gen.LocalK3sPlan{Run: gen.DevelopmentSmokeRun, APIHost: developmentLoopback, APIPort: 6445})
			require.Error(t, err)
		})
	}
}
