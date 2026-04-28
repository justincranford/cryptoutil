// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

// Package apis provides minimal server API wiring for identity-rp.
package apis

import http "net/http"

// RegisterMinimalHandler registers a minimal health-like endpoint for package presence checks.
func RegisterMinimalHandler(mux *http.ServeMux) {
	if mux == nil {
		return
	}

	mux.HandleFunc("/api/v1/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}
