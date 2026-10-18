// Copyright (c) 2025-2026 Justin Cranford.

package test_orch_integration

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	http "net/http"
	"sync"
	"testing"
	"time"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	testPublicBaseURL     = "https://127.0.0.1:14080"
	testAdminBaseURL      = "https://127.0.0.1:14090"
	testHealthEndpointURL = "https://localhost/health"
)

const (
	testCaseNilContext     = "nil context"
	testCaseStartError     = "start error"
	testCaseNilServer      = "nil server"
	testCaseReadinessError = "readiness error"
	testCaseSuccess        = "success"
	testReadinessFailure   = "injected readiness failure"
)

type fakeServiceServer struct {
	mu            sync.Mutex
	publicPort    int
	adminPort     int
	publicBaseURL string
	adminBaseURL  string
	publicRootCAs *x509.CertPool
	adminRootCAs  *x509.CertPool
	ready         bool
	startErr      error
	shutdownErr   error
	shutdownHit   bool
	setPorts      bool
}

func (f *fakeServiceServer) Start(context.Context) error {
	if f.startErr != nil {
		return f.startErr
	}

	if f.setPorts {
		f.mu.Lock()
		f.publicPort = 14080
		f.adminPort = 14090
		f.mu.Unlock()
	}

	return nil
}

func (f *fakeServiceServer) Shutdown(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.shutdownHit = true

	return f.shutdownErr
}

func (f *fakeServiceServer) DB() *gorm.DB                                           { return nil }
func (f *fakeServiceServer) App() *cryptoutilAppsFrameworkServiceServer.Application { return nil }

func (f *fakeServiceServer) PublicPort() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.publicPort
}

func (f *fakeServiceServer) AdminPort() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.adminPort
}

func (f *fakeServiceServer) SetReady(ready bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.ready = ready
}

func (f *fakeServiceServer) PublicBaseURL() string { return f.publicBaseURL }

func (f *fakeServiceServer) AdminBaseURL() string { return f.adminBaseURL }

func (f *fakeServiceServer) TLSRootCAPool() *x509.CertPool { return f.publicRootCAs }

func (f *fakeServiceServer) AdminTLSRootCAPool() *x509.CertPool                     { return f.adminRootCAs }
func (f *fakeServiceServer) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService      { return nil }
func (f *fakeServiceServer) Telemetry() *cryptoutilSharedTelemetry.TelemetryService { return nil }
func (f *fakeServiceServer) Barrier() *cryptoutilAppsFrameworkServiceServerBarrier.Service {
	return nil
}

func TestStartIntegrationServer_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		srv               cryptoutilAppsFrameworkServiceServer.ServiceServer
		ctx               context.Context
		readinessError    string
		shutdownError     error
		nilReadinessCheck bool
		wantError         string
	}{
		{name: testCaseNilContext, srv: &fakeServiceServer{setPorts: true}, wantError: integrationContextCannotBeNil},
		{name: testCaseNilServer, ctx: context.Background(), wantError: integrationServerCannotBeNil},
		{name: "nil readiness check", srv: &fakeServiceServer{setPorts: true}, ctx: context.Background(), nilReadinessCheck: true, wantError: "readiness check cannot be nil"},
		{name: testCaseStartError, srv: &fakeServiceServer{startErr: errors.New("injected start failure")}, ctx: context.Background(), wantError: integrationServerStartFailure},
		{name: testCaseReadinessError, srv: &fakeServiceServer{setPorts: true}, ctx: context.Background(), readinessError: testReadinessFailure, wantError: integrationReadinessCheckFailed},
		{name: "readiness and cleanup errors", srv: &fakeServiceServer{setPorts: true, shutdownErr: errors.New("injected shutdown failure")}, ctx: context.Background(), readinessError: testReadinessFailure, wantError: "cleanup failed"},
		{name: testCaseSuccess, srv: &fakeServiceServer{setPorts: true, publicBaseURL: testPublicBaseURL, adminBaseURL: testAdminBaseURL}, ctx: context.Background()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var readinessCheck readinessCheckFn
			if !tc.nilReadinessCheck {
				readinessCheck = func(context.Context, cryptoutilAppsFrameworkServiceServer.ServiceServer) error {
					if tc.readinessError != "" {
						return errors.New(tc.readinessError)
					}

					return nil
				}
			}

			is, err := startIntegrationServer(tc.ctx, t, tc.srv, nil, readinessCheck)
			if tc.wantError != "" {
				require.Nil(t, is)
				require.ErrorContains(t, err, tc.wantError)

				if tc.readinessError != "" {
					fake, ok := tc.srv.(*fakeServiceServer)
					require.True(t, ok)
					require.True(t, fake.shutdownHit)
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, is)
			require.Equal(t, testPublicBaseURL, is.PublicBaseURL())
			require.Equal(t, testAdminBaseURL, is.AdminBaseURL())
			require.Nil(t, is.DB())
			require.Equal(t, tc.srv, is.Server())
			require.NoError(t, is.Shutdown(context.Background()))
		})
	}
}

func TestStartIntegrationServerForTestMain_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		srv            cryptoutilAppsFrameworkServiceServer.ServiceServer
		ctx            context.Context
		readinessError string
		wantError      string
	}{
		{name: testCaseNilContext, srv: &fakeServiceServer{setPorts: true}, wantError: integrationContextCannotBeNil},
		{name: testCaseNilServer, ctx: context.Background(), wantError: integrationServerCannotBeNil},
		{name: testCaseStartError, srv: &fakeServiceServer{startErr: errors.New("injected start failure")}, ctx: context.Background(), wantError: integrationServerStartFailure},
		{name: "timeout waiting for ports", srv: &fakeServiceServer{}, ctx: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			return ctx
		}(), wantError: "timed out waiting for server ports"},
		{name: testCaseReadinessError, srv: &fakeServiceServer{setPorts: true}, ctx: context.Background(), readinessError: testReadinessFailure, wantError: integrationReadinessCheckFailed},
		{name: testCaseSuccess, srv: &fakeServiceServer{setPorts: true, publicBaseURL: testPublicBaseURL, adminBaseURL: testAdminBaseURL}, ctx: context.Background()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			readinessCheck := func(context.Context, cryptoutilAppsFrameworkServiceServer.ServiceServer) error {
				if tc.readinessError != "" {
					return errors.New(tc.readinessError)
				}

				return nil
			}

			is, err := startIntegrationServer(tc.ctx, nil, tc.srv, nil, readinessCheck)
			if tc.wantError != "" {
				require.Nil(t, is)
				require.ErrorContains(t, err, tc.wantError)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, is)
			require.NoError(t, is.Shutdown(context.Background()))
		})
	}
}

func TestStartIntegrationServer_ExportedWrapper(t *testing.T) {
	t.Parallel()

	_, err := StartIntegrationServer(
		context.Background(),
		t,
		&fakeServiceServer{startErr: errors.New("injected server start failure")},
		nil,
	)

	require.ErrorContains(t, err, integrationServerStartFailure)
}

func TestStartIntegrationServerForTestMain_ExportedWrapper(t *testing.T) {
	t.Parallel()

	_, err := StartIntegrationServerForTestMain(
		context.Background(),
		&fakeServiceServer{startErr: errors.New("injected server start failure")},
		nil,
	)

	require.ErrorContains(t, err, integrationServerStartFailure)
}

func TestWaitForServerReadiness_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	srv := &fakeServiceServer{
		publicRootCAs: x509.NewCertPool(),
		adminRootCAs:  x509.NewCertPool(),
		publicBaseURL: testPublicBaseURL,
		adminBaseURL:  testAdminBaseURL,
	}

	err := waitForServerReadiness(ctx, srv)

	require.ErrorIs(t, err, context.Canceled)
}

func TestWaitForServerReadinessWithProbe_InvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		srv     cryptoutilAppsFrameworkServiceServer.ServiceServer
		probe   readinessProbeFn
		wantErr string
	}{
		{name: "nil context", srv: &fakeServiceServer{}, probe: func(context.Context, *http.Client, string) (int, error) { return http.StatusOK, nil }, wantErr: integrationContextCannotBeNil},
		{name: "nil server", ctx: context.Background(), probe: func(context.Context, *http.Client, string) (int, error) { return http.StatusOK, nil }, wantErr: integrationServerCannotBeNil},
		{name: "nil probe", ctx: context.Background(), srv: &fakeServiceServer{}, wantErr: "readiness probe cannot be nil"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := waitForServerReadinessWithProbe(tc.ctx, tc.srv, tc.probe)

			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

type failingCloseBody struct{}

func (failingCloseBody) Read([]byte) (int, error) { return 0, io.EOF }

func (failingCloseBody) Close() error { return errors.New("injected response close failure") }

func TestRequestHealthEndpoint_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		prepare    func(*testing.T) (*http.Client, string)
		wantStatus int
		wantErr    string
	}{
		{
			name: "healthy HTTP response",
			prepare: func(*testing.T) (*http.Client, string) {
				client := &http.Client{
					Transport: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusNoContent,
							Header:     make(http.Header),
							Body:       http.NoBody,
							Request:    request,
						}, nil
					}),
				}

				return client, testHealthEndpointURL
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "invalid endpoint URL",
			prepare: func(*testing.T) (*http.Client, string) {
				return http.DefaultClient, "%"
			},
			wantErr: "create readiness request",
		},
		{
			name: "transport failure",
			prepare: func(*testing.T) (*http.Client, string) {
				client := &http.Client{
					Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
						return nil, errors.New("injected HTTP transport failure")
					}),
				}

				return client, testHealthEndpointURL
			},
			wantErr: "send readiness request",
		},
		{
			name: "response body close failure",
			prepare: func(*testing.T) (*http.Client, string) {
				client := &http.Client{
					Transport: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Header:     make(http.Header),
							Body:       failingCloseBody{},
							Request:    request,
						}, nil
					}),
				}

				return client, testHealthEndpointURL
			},
			wantErr: "close readiness response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, endpointURL := tc.prepare(t)
			statusCode, err := requestHealthEndpoint(context.Background(), client, endpointURL)

			if tc.wantErr != "" {
				require.Zero(t, statusCode)
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, statusCode)
		})
	}
}

type readinessProbeResult struct {
	statusCode int
	err        error
}

func TestWaitForServerReadinessWithProbe_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		publicRootCAs         *x509.CertPool
		adminRootCAs          *x509.CertPool
		publicBaseURL         string
		adminBaseURL          string
		probeResults          []readinessProbeResult
		ctxTimeout            time.Duration
		wantError             string
		wantProbeURLs         []string
		repeatLastProbeResult bool
	}{
		{
			name:          "both HTTPS listeners healthy",
			publicRootCAs: x509.NewCertPool(),
			adminRootCAs:  x509.NewCertPool(),
			publicBaseURL: testPublicBaseURL,
			adminBaseURL:  testAdminBaseURL,
			probeResults:  []readinessProbeResult{{statusCode: 200}, {statusCode: 200}},
			wantProbeURLs: []string{testPublicBaseURL + cryptoutilSharedMagic.DefaultServiceHealthEndpoint, testAdminBaseURL + cryptoutilSharedMagic.DefaultAdminLivezEndpoint},
		},
		{
			name:          "transient public error retries both listeners",
			publicRootCAs: x509.NewCertPool(),
			adminRootCAs:  x509.NewCertPool(),
			publicBaseURL: testPublicBaseURL,
			adminBaseURL:  testAdminBaseURL,
			probeResults:  []readinessProbeResult{{err: errors.New("injected startup connection error")}, {statusCode: 200}, {statusCode: 200}},
			wantProbeURLs: []string{testPublicBaseURL + cryptoutilSharedMagic.DefaultServiceHealthEndpoint, testPublicBaseURL + cryptoutilSharedMagic.DefaultServiceHealthEndpoint, testAdminBaseURL + cryptoutilSharedMagic.DefaultAdminLivezEndpoint},
		},
		{
			name:                  "non-success listener status fails after timeout",
			publicRootCAs:         x509.NewCertPool(),
			adminRootCAs:          x509.NewCertPool(),
			publicBaseURL:         testPublicBaseURL,
			adminBaseURL:          testAdminBaseURL,
			probeResults:          []readinessProbeResult{{statusCode: 503}},
			ctxTimeout:            cryptoutilSharedMagic.TestUnitShortTimeout,
			wantError:             "public health endpoint returned HTTP 503",
			wantProbeURLs:         []string{testPublicBaseURL + cryptoutilSharedMagic.DefaultServiceHealthEndpoint},
			repeatLastProbeResult: true,
		},
		{
			name:          "missing public trust pool is rejected",
			adminRootCAs:  x509.NewCertPool(),
			publicBaseURL: testPublicBaseURL,
			adminBaseURL:  testAdminBaseURL,
			wantError:     "public TLS root CA pool is nil",
		},
		{
			name:          "missing admin trust pool is rejected",
			publicRootCAs: x509.NewCertPool(),
			publicBaseURL: testPublicBaseURL,
			adminBaseURL:  testAdminBaseURL,
			wantError:     "admin TLS root CA pool is nil",
		},
		{
			name:          "missing public URL is rejected",
			publicRootCAs: x509.NewCertPool(),
			adminRootCAs:  x509.NewCertPool(),
			adminBaseURL:  testAdminBaseURL,
			wantError:     "public base URL is empty",
		},
		{
			name:          "missing admin URL is rejected",
			publicRootCAs: x509.NewCertPool(),
			adminRootCAs:  x509.NewCertPool(),
			publicBaseURL: testPublicBaseURL,
			wantError:     "admin base URL is empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fake := &fakeServiceServer{
				publicBaseURL: tc.publicBaseURL,
				adminBaseURL:  tc.adminBaseURL,
				publicRootCAs: tc.publicRootCAs,
				adminRootCAs:  tc.adminRootCAs,
			}
			ctx := context.Background()

			if tc.ctxTimeout > 0 {
				var cancel context.CancelFunc

				ctx, cancel = context.WithTimeout(ctx, tc.ctxTimeout)
				defer cancel()
			}

			var (
				probeIndex int
				probeURLs  []string
			)

			probe := func(_ context.Context, _ *http.Client, endpointURL string) (int, error) {
				probeURLs = append(probeURLs, endpointURL)

				if probeIndex >= len(tc.probeResults) {
					if tc.repeatLastProbeResult && len(tc.probeResults) > 0 {
						lastResult := tc.probeResults[len(tc.probeResults)-1]

						return lastResult.statusCode, lastResult.err
					}

					return 0, errors.New("unexpected extra readiness probe")
				}

				result := tc.probeResults[probeIndex]
				probeIndex++

				return result.statusCode, result.err
			}

			err := waitForServerReadinessWithProbe(ctx, fake, probe)
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantError)
			}

			if tc.repeatLastProbeResult {
				require.NotEmpty(t, probeURLs)

				for _, probedURL := range probeURLs {
					require.Equal(t, tc.wantProbeURLs[0], probedURL)
				}
			} else {
				require.Equal(t, tc.wantProbeURLs, probeURLs)
			}
		})
	}
}

func TestIntegrationServer_ShutdownAndAccessors(t *testing.T) {
	t.Parallel()

	t.Run("nil server shutdown is no-op", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, (&IntegrationServer{}).Shutdown(context.Background()))
	})

	t.Run("shutdown error propagated", func(t *testing.T) {
		t.Parallel()

		is := &IntegrationServer{srv: &fakeServiceServer{shutdownErr: fmt.Errorf("injected shutdown failure")}}
		err := is.Shutdown(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, "shutdown failed")
	})

	t.Run("cleanup error propagated", func(t *testing.T) {
		t.Parallel()

		is := &IntegrationServer{
			srv:       &fakeServiceServer{},
			cleanupFn: func() error { return fmt.Errorf("injected cleanup failure") },
		}
		err := is.Shutdown(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, "cleanup failed")
	})

	t.Run("shutdown error with testing tb logs and propagates", func(t *testing.T) {
		t.Parallel()
		is := &IntegrationServer{tb: t, srv: &fakeServiceServer{shutdownErr: fmt.Errorf("tb shutdown failure")}}
		err := is.Shutdown(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, "shutdown failed")
	})

	t.Run("cleanup error with testing tb logs and propagates", func(t *testing.T) {
		t.Parallel()
		is := &IntegrationServer{
			tb:        t,
			srv:       &fakeServiceServer{},
			cleanupFn: func() error { return fmt.Errorf("tb cleanup failure") },
		}
		err := is.Shutdown(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, "cleanup failed")
	})

	t.Run("empty urls when server nil", func(t *testing.T) {
		t.Parallel()

		is := &IntegrationServer{}
		require.Equal(t, "", is.PublicBaseURL())
		require.Equal(t, "", is.AdminBaseURL())
	})
}

func TestBuildBrokenFixtures_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		create        func(t *testing.T) (*IntegrationServer, error)
		wantBrokenDB  bool
		wantBrokenAPI bool
		wantErrPart   string
	}{
		{
			name: "broken db fixture",
			create: func(t *testing.T) (*IntegrationServer, error) {
				return BuildBrokenDBFixture(t, "db reason", &fakeServiceServer{})
			},
			wantBrokenDB: true,
			wantErrPart:  "db reason",
		},
		{
			name: "broken api fixture",
			create: func(t *testing.T) (*IntegrationServer, error) {
				return BuildBrokenAPIFixture(t, "api reason", &fakeServiceServer{}, nil)
			},
			wantBrokenAPI: true,
			wantErrPart:   "api reason",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			is, err := tc.create(t)
			require.NoError(t, err)
			require.NotNil(t, is)
			require.Equal(t, tc.wantBrokenDB, is.BrokenDB())
			require.Equal(t, tc.wantBrokenAPI, is.BrokenAPI())

			if tc.wantBrokenDB {
				require.ErrorContains(t, is.BrokenDBError(), tc.wantErrPart)
				require.Nil(t, is.BrokenAPIError())
			} else {
				require.ErrorContains(t, is.BrokenAPIError(), tc.wantErrPart)
				require.Nil(t, is.BrokenDBError())
			}
		})
	}
}

func TestBuildBrokenFixtures_CleanupClosurePaths(t *testing.T) {
	t.Parallel()

	t.Run("broken db cleanup closure logs cleanup error", func(t *testing.T) {
		t.Parallel()
		is, err := BuildBrokenDBFixture(t, "db cleanup path", &fakeServiceServer{})
		require.NoError(t, err)
		require.NotNil(t, is)
		is.cleanupFn = func() error { return fmt.Errorf("cleanup failure") }
	})

	t.Run("broken api cleanup closure logs cleanup error", func(t *testing.T) {
		t.Parallel()
		is, err := BuildBrokenAPIFixture(t, "api cleanup path", &fakeServiceServer{}, nil)
		require.NoError(t, err)
		require.NotNil(t, is)
		is.cleanupFn = func() error { return fmt.Errorf("cleanup failure") }
	})
}
