// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"archive/tar"
	"bytes"
	"context"
	sha256 "crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	json "encoding/json"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentSmokeFlagsNeverLaunchInvalidInput(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{testDevelopmentHelp},
		{"--admission-certified"},
		{testDevelopmentTimeout},
		{"--image=local", testDevelopmentTimeout, "--repository-root=relative"},
	} {
		t.Run("flags="+stringJoinDevelopment(args), func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			code := developmentSmokeMain(args, &stdout, &stderr)
			if args[0] == testDevelopmentHelp {
				require.Zero(t, code)
			} else {
				require.Equal(t, 2, code)
			}
		})
	}
}

func stringJoinDevelopment(args []string) string {
	raw, _ := json.Marshal(args)

	return string(raw)
}

func TestDevelopmentImageTransportRequiresActualDigestBlob(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testDevelopmentActualIndex, testDevelopmentForgedIndex, testDevelopmentAbsentIndex, testDevelopmentInvalidReference} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			index := []byte(`{"schemaVersion":2,"manifests":[]}`)
			digest := sha256.Sum256(index)
			reference := "owned/suite@sha256:" + hex.EncodeToString(digest[:])

			var archive bytes.Buffer

			writer := tar.NewWriter(&archive)

			name := "blobs/sha256/" + hex.EncodeToString(digest[:])
			if test == testDevelopmentAbsentIndex {
				name = "not-the-inspected-index"
			}

			if test == testDevelopmentForgedIndex {
				index = []byte("different")
			}

			require.NoError(t, writer.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeReg, Size: int64(len(index)), Mode: 0o600}))
			_, err := writer.Write(index)
			require.NoError(t, err)
			require.NoError(t, writer.Close())

			if test == testDevelopmentInvalidReference {
				reference = "suite:latest"
			}

			err = verifyDevelopmentImageArchive(archive.Bytes(), reference)
			if test == testDevelopmentActualIndex {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDevelopmentKubeconfigRequiresOwnedTLSContext(t *testing.T) {
	t.Parallel()

	for _, test := range []string{
		testDevelopmentOwnedContext, testDevelopmentInvalidYAML, testDevelopmentWrongContext,
		testDevelopmentAbsentClusters, testDevelopmentWrongServer, testDevelopmentAbsentTrust,
	} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			_, material, _ := publicationFixture(t)
			config := map[string]any{
				developmentKubeCurrent: developmentDefaultContext,
				developmentKubeClusters: []any{map[string]any{developmentName: developmentDefaultContext, testDevelopmentClusterField: map[string]any{
					developmentServerField: developmentDefaultServer, developmentKubeCA: base64.StdEncoding.EncodeToString(material[len(material)-1].Data[testPKICAKey]),
				}}},
				developmentKubeContexts: []any{map[string]any{developmentName: developmentDefaultContext, developmentKubeContext: map[string]any{
					testDevelopmentClusterField: developmentDefaultContext, developmentKubeUser: developmentDefaultContext,
				}}},
				developmentKubeUsers: []any{map[string]any{developmentName: developmentDefaultContext, developmentKubeUser: map[string]any{
					developmentKubeCertificate: base64.StdEncoding.EncodeToString(material[1].Data[testPKITLSCertKey]),
					developmentKubeKey:         base64.StdEncoding.EncodeToString(material[1].Data[testPKITLSPrivateKey]),
				}}},
			}

			switch test {
			case testDevelopmentWrongContext:
				config[developmentKubeCurrent] = "shared"
			case testDevelopmentAbsentClusters:
				delete(config, developmentKubeClusters)
			case testDevelopmentWrongServer, testDevelopmentAbsentTrust:
				clusters, ok := config[developmentKubeClusters].([]any)
				require.True(t, ok)
				entry, ok := clusters[0].(map[string]any)
				require.True(t, ok)
				details, ok := entry[testDevelopmentClusterField].(map[string]any)
				require.True(t, ok)

				if test == testDevelopmentWrongServer {
					details[cryptoutilSharedMagic.CMD_SERVER] = "https://shared:6443"
				} else {
					delete(details, developmentKubeCA)
				}
			}

			raw, err := json.Marshal(config)
			require.NoError(t, err)

			if test == testDevelopmentInvalidYAML {
				raw = []byte("{")
			}

			out, err := developmentKubeconfig(raw, gen.LocalK3sPlan{Run: gen.DevelopmentSmokeRun, APIHost: developmentLoopback, APIPort: 6445})
			if test == testDevelopmentOwnedContext {
				require.NoError(t, err)
				require.Contains(t, string(out), gen.DevelopmentSmokeRun)
				require.Contains(t, string(out), "https://127.0.0.1:6445")
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDevelopmentAuthenticationExercisesRealOrderedClientFlow(t *testing.T) {
	t.Parallel()

	for _, test := range []string{
		testDevelopmentAuthComplete, testDevelopmentAuthRegistration, testDevelopmentAuthCredentials,
		testDevelopmentAuthPassword, testDevelopmentAuthToken, testDevelopmentAuthDenial,
		testDevelopmentAuthJSONRegistration, testDevelopmentAuthJSONToken, testDevelopmentAuthTokenResponse,
	} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				switch r.URL.Path {
				case "/service/api/v1/auth/register-client":
					if test == testDevelopmentAuthRegistration {
						w.WriteHeader(http.StatusInternalServerError)

						return
					}

					w.WriteHeader(http.StatusCreated)

					if test == testDevelopmentAuthJSONRegistration {
						_, _ = io.WriteString(w, "{")

						return
					}

					if test == testDevelopmentAuthCredentials {
						_, _ = io.WriteString(w, `{}`)
					} else {
						_, _ = io.WriteString(w, `{"client_id":"owned-test","client_secret":"test-only","tenant_id":"owned"}`)
					}
				case "/service/api/v1/auth/token":
					user, password, ok := r.BasicAuth()
					require.True(t, ok)
					require.Equal(t, "owned-test", user)

					if password != "test-only" && test != testDevelopmentAuthPassword {
						w.WriteHeader(http.StatusUnauthorized)

						return
					}

					if test == testDevelopmentAuthJSONToken {
						_, _ = io.WriteString(w, "{")

						return
					}

					if test == testDevelopmentAuthTokenResponse {
						w.WriteHeader(http.StatusInternalServerError)

						return
					}

					if test == testDevelopmentAuthToken {
						_, _ = io.WriteString(w, `{}`)
					} else {
						_, _ = io.WriteString(w, `{"token":"test-service-token","token_type":"Bearer"}`)
					}
				case cryptoutilSharedMagic.DefaultServiceHealthEndpoint:
					if r.Header.Get("Authorization") == "" {
						w.WriteHeader(http.StatusUnauthorized)

						return
					}

					require.Equal(t, "Bearer test-service-token", r.Header.Get("Authorization"))

					if test == testDevelopmentAuthDenial {
						w.WriteHeader(http.StatusUnauthorized)
					}
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			transport := &http.Transport{DisableKeepAlives: true}
			defer transport.CloseIdleConnections()

			err := verifyDevelopmentAuthentication(ctx, &http.Client{Transport: transport}, server.URL)
			if test == testDevelopmentAuthComplete {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "test-only")
			}
		})
	}
}
