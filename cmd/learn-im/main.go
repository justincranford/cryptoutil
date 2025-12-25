// Copyright (c) 2025 Justin Cranford
//
//

// Package main is the entrypoint for learn-im encrypted instant messaging service.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	"cryptoutil/internal/learn/server"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// Version information (injected during build).
var (
	version   = "dev"
	buildDate = "unknown"
	gitCommit = "unknown"
)

func main() {
	os.Exit(internalMain(os.Args))
}

// internalMain implements main logic with testable dependencies.
func internalMain(args []string) int {
	ctx := context.Background()

	if len(args) > 1 && args[1] == "version" {
		fmt.Printf("learn-im %s (built %s, commit %s)\n", version, buildDate, gitCommit)

		return 0
	}

	// Initialize SQLite in-memory database for demonstration.
	db, err := initDatabase(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to initialize database: %v\n", err)

		return 1
	}

	// Create learn-im server.
	cfg := &server.Config{
		PublicPort: int(cryptoutilMagic.DefaultPublicPortLearnIM),
		AdminPort:  cryptoutilMagic.DefaultPrivatePortLearnIM,
		DB:         db,
	}

	srv, err := server.New(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	// Start server with graceful shutdown.
	errChan := make(chan error, 1)

	go func() {
		fmt.Printf("🚀 Starting learn-im service...\n")
		fmt.Printf("   Public Server: https://127.0.0.1:%d\n", cryptoutilMagic.DefaultPublicPortLearnIM)
		fmt.Printf("   Admin Server:  https://127.0.0.1:%d\n", cryptoutilMagic.DefaultPrivatePortLearnIM)

		errChan <- srv.Start(ctx)
	}()

	// Wait for interrupt signal.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Server error: %v\n", err)

			return 1
		}
	case sig := <-sigChan:
		fmt.Printf("\n⏹️  Received signal %v, shutting down gracefully...\n", sig)
	}

	fmt.Println("✅ learn-im service stopped")

	return 0
}

// initDatabase initializes SQLite database with schema.
func initDatabase(ctx context.Context) (*gorm.DB, error) {
	// Use in-memory SQLite for demonstration.
	dsn := "file::memory:?cache=shared"

	// Open SQLite with modernc driver (CGO-free).
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	// Enable WAL mode for better concurrency.
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Set busy timeout (30 seconds).
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Create GORM db with SQLite dialector.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GORM database: %w", err)
	}

	// Configure connection pool for SQLite.
	sqlDB, err = db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0)

	// Run migrations.
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations creates database schema.
func runMigrations(db *gorm.DB) error {
	// Create users table.
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			public_key_jwk TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create messages table.
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			sender_id TEXT NOT NULL,
			encrypted_content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			FOREIGN KEY (sender_id) REFERENCES users(id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	// Create message_receivers table.
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS message_receivers (
			id TEXT PRIMARY KEY,
			message_id TEXT NOT NULL,
			receiver_id TEXT NOT NULL,
			read_at TIMESTAMP,
			deleted_at TIMESTAMP,
			FOREIGN KEY (message_id) REFERENCES messages(id),
			FOREIGN KEY (receiver_id) REFERENCES users(id),
			UNIQUE(message_id, receiver_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create message_receivers table: %w", err)
	}

	return nil
}
