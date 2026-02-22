// Copyright (c) 2025 Justin Cranford

package network

import (
	"context"
	"fmt"
	"io"
	http "net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// errorReadCloser is a test helper that returns configurable errors on Read and Close.
type errorReadCloser struct {
	readErr  error
	closeErr error
	data     []byte
	offset   int
}

func (e *errorReadCloser) Read(p []byte) (int, error) {
	if e.readErr != nil {
		return 0, e.readErr
	}

	if e.offset >= len(e.data) {
		return 0, io.EOF
	}

	n := copy(p, e.data[e.offset:])
	e.offset += n

	return n, nil
}

func (e *errorReadCloser) Close() error {
	return e.closeErr
}

func TestHTTPResponse_ReadAllInjectedError(t *testing.T) {
	// Cannot be parallel: modifies package-level injectable var.
	originalFn := networkReadAllFn
	originalRT := networkRoundTripperFn

	defer func() {
		networkReadAllFn = originalFn
		networkRoundTripperFn = originalRT
	}()

	networkRoundTripperFn = func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
			Body:       &errorReadCloser{data: []byte("body")},
		}, nil
	}

	networkReadAllFn = func(_ io.Reader) ([]byte, error) {
		return nil, fmt.Errorf("injected read error")
	}

	statusCode, _, _, err := HTTPResponse(context.Background(), http.MethodGet, "http://localhost/test", 0, true, nil, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read response body")
	require.Equal(t, http.StatusOK, statusCode)
}

func TestHTTPResponse_BodyCloseInjectedError(t *testing.T) {
	// Cannot be parallel: modifies package-level injectable var.
	originalRT := networkRoundTripperFn

	defer func() { networkRoundTripperFn = originalRT }()

	networkRoundTripperFn = func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
			Body:       &errorReadCloser{data: []byte("body"), closeErr: fmt.Errorf("injected close error")},
		}, nil
	}

	// The Close error only prints a warning — no error returned.
	statusCode, _, body, err := HTTPResponse(context.Background(), http.MethodGet, "http://localhost/test", 0, true, nil, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, []byte("body"), body)
}
