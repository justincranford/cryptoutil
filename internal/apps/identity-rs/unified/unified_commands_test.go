package rs

import (
	"context"
	"errors"
	"fmt"
	"io"
	http "net/http"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type testServiceSettings struct{}

type testServiceLifecycle struct {
	ready    bool
	startErr error
	started  bool
}

func (s *testServiceLifecycle) SetReady(ready bool) {
	s.ready = ready
}

func (s *testServiceLifecycle) Start(context.Context) error {
	s.started = true

	return s.startErr
}

func TestExecuteDispatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		parameters []string
		wantCall   string
		wantParams []string
		wantExit   int
	}{
		{name: "missing subcommand", wantExit: 1},
		{name: valueStart, parameters: []string{valueStart, testConfigPath}, wantCall: valueStart, wantParams: []string{testConfigPath}},
		{name: valueStop, parameters: []string{valueStop, testConfigPath}, wantCall: valueStop, wantParams: []string{testConfigPath}},
		{name: cryptoutilSharedMagic.StringStatus, parameters: []string{cryptoutilSharedMagic.StringStatus, testConfigPath}, wantCall: cryptoutilSharedMagic.StringStatus, wantParams: []string{testConfigPath}},
		{name: valueHealth, parameters: []string{valueHealth, testConfigPath}, wantCall: valueHealth, wantParams: []string{testConfigPath}},
		{name: "unknown subcommand", parameters: []string{cryptoutilSharedMagic.STRATEGY_UNKNOWN}, wantExit: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var (
				gotCall   string
				gotParams []string
				gotExit   int
			)

			record := func(name string) func([]string) {
				return func(parameters []string) {
					gotCall = name
					gotParams = parameters
				}
			}
			handlers := commandHandlers{
				start:  record(valueStart),
				stop:   record(valueStop),
				status: record(cryptoutilSharedMagic.StringStatus),
				health: record(valueHealth),
				exit:   func(code int) { gotExit = code },
			}

			execute(test.parameters, handlers)

			require.Equal(t, test.wantCall, gotCall)
			require.Equal(t, test.wantParams, gotParams)
			require.Equal(t, test.wantExit, gotExit)
		})
	}
}

func TestStartServiceLifecycle(t *testing.T) {
	t.Parallel()

	parseFailure := errors.New("parse failed")
	createFailure := errors.New("create failed")
	startFailure := errors.New("start failed")

	tests := []struct {
		name          string
		parseErr      error
		newServerErr  error
		startErr      error
		wantFatal     string
		wantNewServer bool
		wantReady     bool
		wantStarted   bool
	}{
		{name: "configuration parsing fails", parseErr: parseFailure, wantFatal: "Failed to load config from " + testConfigPath + ": parse failed"},
		{name: "server creation fails", newServerErr: createFailure, wantFatal: "Failed to create RS server: create failed", wantNewServer: true},
		{name: "server start fails", startErr: startFailure, wantFatal: "RS server error: start failed", wantNewServer: true, wantReady: true, wantStarted: true},
		{name: "server starts", wantNewServer: true, wantReady: true, wantStarted: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			settings := &testServiceSettings{}
			server := &testServiceLifecycle{startErr: test.startErr}

			var (
				gotArgs     []string
				gotExitHelp bool
				gotSettings *testServiceSettings
				gotFatal    string
			)

			parseConfig := func(args []string, exitIfHelp bool) (*testServiceSettings, error) {
				gotArgs = args
				gotExitHelp = exitIfHelp

				return settings, test.parseErr
			}
			newServer := func(_ context.Context, got *testServiceSettings) (serviceLifecycle, error) {
				gotSettings = got

				return server, test.newServerErr
			}
			fatalf := func(format string, args ...any) {
				gotFatal = fmt.Sprintf(format, args...)
			}

			startService([]string{"--config=" + testConfigPath}, parseConfig, newServer, fatalf)

			require.Equal(t, []string{valueStart, cryptoutilSharedMagic.IdentityCLIFlagConfig, testConfigPath}, gotArgs)
			require.False(t, gotExitHelp)
			require.Equal(t, test.wantFatal, gotFatal)
			require.Equal(t, test.wantNewServer, gotSettings != nil)

			if test.wantNewServer {
				require.Same(t, settings, gotSettings)
			}

			require.Equal(t, test.wantReady, server.ready)
			require.Equal(t, test.wantStarted, server.started)
		})
	}
}

type testRequestDoer struct {
	response *http.Response
	err      error
	calls    int
}

func (d *testRequestDoer) Do(request *http.Request) (*http.Response, error) {
	d.calls++

	return d.response, d.err
}

type testResponseBody struct {
	closeErr error
	closed   bool
}

func (b *testResponseBody) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (b *testResponseBody) Close() error {
	b.closed = true

	return b.closeErr
}

func TestSendServiceRequest(t *testing.T) {
	t.Parallel()

	requestFailure := errors.New("request failed")
	doFailure := errors.New("transport failed")
	closeFailure := errors.New("close failed")

	tests := []struct {
		name          string
		requestErr    error
		doErr         error
		statusCode    int
		closeErr      error
		wantExit      int
		wantDoCalls   int
		wantBodyClose bool
	}{
		{name: "successful response", statusCode: http.StatusOK, wantDoCalls: 1, wantBodyClose: true},
		{name: "request creation failure", requestErr: requestFailure, wantExit: 1},
		{name: "transport failure", doErr: doFailure, wantExit: 1, wantDoCalls: 1},
		{name: "unexpected response status", statusCode: http.StatusServiceUnavailable, wantExit: 1, wantDoCalls: 1, wantBodyClose: true},
		{name: "response body close failure", statusCode: http.StatusOK, closeErr: closeFailure, wantDoCalls: 1, wantBodyClose: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			body := &testResponseBody{closeErr: test.closeErr}
			doer := &testRequestDoer{
				response: &http.Response{StatusCode: test.statusCode, Body: body},
				err:      test.doErr,
			}

			var (
				gotRequest *http.Request
				exitCodes  []int
			)

			newRequest := func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
				if test.requestErr != nil {
					return nil, test.requestErr
				}

				request, err := http.NewRequestWithContext(ctx, method, url, body)
				gotRequest = request

				return request, err
			}
			request := serviceRequest{
				progressMessage:       "Testing request on port %d\n",
				method:                http.MethodGet,
				path:                  "/test",
				requestFailureMessage: "request creation failed",
				doFailureMessage:      "request execution failed",
				successMessage:        "request succeeded",
				statusFailureMessage:  "unexpected status: %d\n",
				expectedStatus:        http.StatusOK,
			}

			sendServiceRequest(
				[]string{testAdminPortFlag, testAdminPortValue},
				request,
				newRequest,
				doer,
				func(code int) { exitCodes = append(exitCodes, code) },
			)

			require.Equal(t, test.wantDoCalls, doer.calls)
			require.Equal(t, test.wantBodyClose, body.closed)

			if test.wantExit == 0 {
				require.Empty(t, exitCodes)
			} else {
				require.Equal(t, []int{test.wantExit}, exitCodes)
			}

			if test.requestErr == nil {
				require.NotNil(t, gotRequest)
				require.Equal(t, http.MethodGet, gotRequest.Method)
				require.Equal(t, cryptoutilSharedMagic.URLPrefixLocalhostHTTPS+testAdminPortValue+"/test", gotRequest.URL.String())
			}
		})
	}
}

type testRoundTripper func(*http.Request) (*http.Response, error)

func (f testRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

// Sequential: replaces http.DefaultTransport to exercise the real command wiring without network access.
func TestExecuteServiceHTTPWiring(t *testing.T) {
	originalTransport := http.DefaultTransport

	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	var requests []*http.Request

	http.DefaultTransport = testRoundTripper(func(request *http.Request) (*http.Response, error) {
		requests = append(requests, request)

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    request,
		}, nil
	})

	tests := []struct {
		name            string
		parameters      []string
		method          string
		path            string
		wantContentType bool
	}{
		{name: "stop command", parameters: []string{valueStop, testAdminPortFlag, testAdminPortValue}, method: http.MethodPost, path: "/admin/v1/shutdown", wantContentType: true},
		{name: "status command", parameters: []string{cryptoutilSharedMagic.StringStatus, testAdminPortFlag, testAdminPortValue}, method: http.MethodGet, path: "/admin/v1/readyz"},
		{name: "health command", parameters: []string{valueHealth, testAdminPortFlag, testAdminPortValue}, method: http.MethodGet, path: "/admin/v1/livez"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := len(requests)

			Execute(test.parameters)

			require.Len(t, requests, before+1)
			require.Equal(t, test.method, requests[before].Method)
			require.Equal(t, cryptoutilSharedMagic.URLPrefixLocalhostHTTPS+testAdminPortValue+test.path, requests[before].URL.String())

			if test.wantContentType {
				require.NotEmpty(t, requests[before].Header.Get("Content-Type"))
			} else {
				require.Empty(t, requests[before].Header.Get("Content-Type"))
			}
		})
	}
}
