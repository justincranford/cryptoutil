// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

// Package testutil provides test utilities for HTTP server testing.
package testutil

import (
	"crypto/tls"
	http "net/http"
	"net/http/httptest"
	"time"
)

// NewInsecureTLSClient creates an HTTP client that accepts self-signed certificates.
// Used for testing HTTPS endpoints with self-signed certificates.
func NewInsecureTLSClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // G402: Tests use self-signed certs
			},
		},
		Timeout: timeout,
	}
}

// NewMockServerOK creates a test server that returns 200 OK responses.
// Used for testing health check success cases.
func NewMockServerOK() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	return httptest.NewTLSServer(handler)
}

// NewMockServerError creates a test server that returns 503 unavailable responses.
// Used for testing health check error handling.
func NewMockServerError() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("Service Unavailable"))
	})

	return httptest.NewTLSServer(handler)
}

// NewMockServerSlow creates a test server with configurable delay.
// Used for testing timeout handling.
func NewMockServerSlow(delay time.Duration) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"delayed"}`))
	})

	return httptest.NewTLSServer(handler)
}

// NewMockServerCustom creates a test server with custom path-based responses.
// Used for testing multiple endpoints with different responses.
func NewMockServerCustom() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/admin/api/v1/health":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("All systems operational"))
		case "/admin/api/v1/livez":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Process is alive and running"))
		case "/admin/api/v1/readyz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Service is ready to accept requests"))
		case "/admin/api/v1/shutdown":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Graceful shutdown initiated"))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not found"))
		}
	})

	return httptest.NewTLSServer(handler)
}
