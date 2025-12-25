// Copyright (c) 2025 Justin Cranford
//
//

// Package server implements the learn-im HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"cryptoutil/internal/learn/repository"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	templateServer "cryptoutil/internal/template/server"
)

// LearnIMServer represents the learn-im service application.
type LearnIMServer struct {
	app *templateServer.Application
	db  *gorm.DB

	// Repositories.
	userRepo    *repository.UserRepository
	messageRepo *repository.MessageRepository
}

// Config holds configuration for the learn-im server.
type Config struct {
	PublicPort int
	AdminPort  uint16
	DB         *gorm.DB
}

// New creates a new learn-im server using the template.
func New(ctx context.Context, cfg *Config) (*LearnIMServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if cfg.DB == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	// Initialize repositories.
	userRepo := repository.NewUserRepository(cfg.DB)
	messageRepo := repository.NewMessageRepository(cfg.DB)

	// Create public server with handlers.
	publicServer, err := NewPublicServer(ctx, cfg.PublicPort, userRepo, messageRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to create public server: %w", err)
	}

	// Create admin server.
	tlsCfg := &templateServer.TLSConfig{
		Mode:             templateServer.TLSModeAuto,
		AutoDNSNames:     []string{"localhost"},
		AutoIPAddresses:  []string{"127.0.0.1", "::1"},
		AutoValidityDays: cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	}

	adminServer, err := templateServer.NewAdminServer(ctx, cfg.AdminPort, tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin server: %w", err)
	}

	// Create application with both servers.
	app, err := templateServer.NewApplication(ctx, publicServer, adminServer)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return &LearnIMServer{
		app:         app,
		db:          cfg.DB,
		userRepo:    userRepo,
		messageRepo: messageRepo,
	}, nil
}

// Start starts both public and admin servers.
func (s *LearnIMServer) Start(ctx context.Context) error {
	//nolint:wrapcheck // Pass-through to template, wrapping not needed.
	return s.app.Start(ctx)
}

// Shutdown gracefully shuts down both servers.
func (s *LearnIMServer) Shutdown(ctx context.Context) error {
	//nolint:wrapcheck // Pass-through to template, wrapping not needed.
	return s.app.Shutdown(ctx)
}

// PublicPort returns the actual public server port.
func (s *LearnIMServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual admin server port.
func (s *LearnIMServer) AdminPort() (int, error) {
	//nolint:wrapcheck // Pass-through to template, wrapping not needed.
	return s.app.AdminPort()
}
