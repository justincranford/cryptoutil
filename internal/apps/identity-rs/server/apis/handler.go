// Copyright (c) 2025-2026 Justin Cranford.
// Package apis provides minimal server API wiring for identity-rs.
package apis

import http "net/http"

// RegisterMinimalHandler registers a minimal endpoint for package presence checks.
func RegisterMinimalHandler(mux *http.ServeMux) {
	if mux == nil {
		return
	}

	mux.HandleFunc("/api/v1/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}
