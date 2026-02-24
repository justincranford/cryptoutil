// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides the SPA relying party demo application.
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	http "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	// Parse command-line flags.
	port := flag.Int("port", cryptoutilSharedMagic.DefaultSPARPPort, "port for SPA RP server")
	bindAddress := flag.String("bind", "127.0.0.1", "bind address for SPA RP server")

	flag.Parse()

	// Create HTTP server with embedded static files.
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("failed to create static file system: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(staticFS)))
	// Handle OAuth callback by serving the main HTML page
	mux.HandleFunc("/callback", func(w http.ResponseWriter, _ *http.Request) {
		// Serve the index.html file for OAuth callback processing
		indexFile, err := staticFS.Open("index.html")
		if err != nil {
			http.Error(w, "file not found", http.StatusNotFound)

			return
		}

		defer func() {
			if closeErr := indexFile.Close(); closeErr != nil {
				log.Printf("Failed to close index file: %v", closeErr)
			}
		}()

		w.Header().Set("Content-Type", "text/html")

		if _, err := io.Copy(w, indexFile); err != nil {
			http.Error(w, "Failed to serve index.html", http.StatusInternalServerError)

			return
		}
	})

	addr := fmt.Sprintf("%s:%d", *bindAddress, *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  time.Duration(cryptoutilSharedMagic.FiberReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cryptoutilSharedMagic.FiberWriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cryptoutilSharedMagic.FiberIdleTimeoutSeconds) * time.Second,
	}

	// Start server in goroutine.
	go func() {
		log.Printf("SPA relying party server listening on %s", addr)
		log.Printf("Open http://%s in your browser", addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	log.Println("shutting down SPA RP server...")

	// Graceful shutdown with timeout.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cryptoutilSharedMagic.ShutdownTimeoutSeconds)*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("SPA RP server stopped gracefully")
}
