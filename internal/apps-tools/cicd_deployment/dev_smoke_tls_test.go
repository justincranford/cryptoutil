// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	http "net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentPublicAPIVerifiesTLSIdentityAndClosesOwnedForward(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testDevVerifiedTLS, testDevMissingWorkload, testDevMissingTrust, testDevForwardFailure, testDevWrongHostname, testDevBadAuthentication} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			values, secrets, _ := publicationFixture(t)
			input := gen.BootstrapPreflightInput{RuntimeValues: values}
			pair, err := tls.X509KeyPair(secrets[0].Data[pki.SecretDataPublicCertificate], secrets[0].Data[pki.SecretDataPublicPrivateKey])
			require.NoError(t, err)

			if test == testDevMissingWorkload {
				input.RuntimeValues = nil
			}

			if test == testDevMissingTrust {
				secrets = nil
			}

			if test == testDevWrongHostname {
				input.RuntimeValues[0].Workloads[0].InstanceServices[0].Name = "not-covered-by-certificate"
			}

			config := execution.KubectlConfig{
				Kubectl:    filepath.Join(t.TempDir(), "explicit-test-kubectl.exe"),
				Kubeconfig: filepath.Join(t.TempDir(), "explicit-test-kubeconfig"), Context: "explicit-owned-context",
			}
			starts, closures := 0, 0
			start := func(_ context.Context, binary string, args []string) (func(), error) {
				starts++

				require.Equal(t, config.Kubectl, binary)
				require.Equal(t, []string{"--kubeconfig", config.Kubeconfig, "--context", config.Context}, args[:4])
				require.Contains(t, args, "--address=127.0.0.1")

				if test == testDevForwardFailure {
					return nil, errors.New("PRIVATE")
				}

				port := strings.Split(args[len(args)-1], ":")[0]
				listener, err := net.Listen("tcp", "127.0.0.1:"+port)
				require.NoError(t, err)

				server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.NotNil(t, r.TLS)
					require.Equal(t, uint16(tls.VersionTLS13), r.TLS.Version)

					if test == testDevBadAuthentication {
						w.WriteHeader(http.StatusForbidden)

						return
					}

					switch r.URL.Path {
					case cryptoutilSharedMagic.DefaultPublicServiceAuthClientRegisterPath:
						w.WriteHeader(http.StatusCreated)
						_, _ = io.WriteString(w, `{"client_id":"owned-test","client_secret":"synthetic-test-only","tenant_id":"owned-test-tenant"}`)
					case cryptoutilSharedMagic.DefaultPublicServiceAuthTokenPath:
						_, password, ok := r.BasicAuth()
						require.True(t, ok)

						if password != "synthetic-test-only" {
							w.WriteHeader(http.StatusUnauthorized)

							return
						}

						_, _ = io.WriteString(w, `{"token":"synthetic-test-token","token_type":"Bearer"}`)
					default:
						if r.Header.Get("Authorization") == "" {
							w.WriteHeader(http.StatusUnauthorized)
						}
					}
				}))
				require.NoError(t, server.Listener.Close())
				server.Listener = listener
				server.TLS = &tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{pair}}
				server.Config.ErrorLog = log.New(io.Discard, "", 0)
				server.StartTLS()

				return func() { closures++; server.Close() }, nil
			}

			timeout := time.Minute
			if test == testDevWrongHostname {
				timeout = time.Second
			}

			ctx, cancel := context.WithTimeout(t.Context(), timeout)
			defer cancel()

			err = verifyDevelopmentPublicAPIWithForward(ctx, input, secrets, config, start)
			if test == testDevVerifiedTLS {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "PRIVATE")
			}

			if test == testDevMissingWorkload || test == testDevMissingTrust {
				require.Zero(t, starts)
			} else {
				require.Equal(t, 1, starts)

				if test != testDevForwardFailure {
					require.Equal(t, 1, closures)
				}
			}
		})
	}
}
