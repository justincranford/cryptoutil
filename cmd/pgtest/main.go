package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQL Connection Test Utility
//
// This utility helps test connections to the PostgreSQL database from the host machine.
// It attempts to connect using different connection string formats and reports the results.
//
// Usage:
//
//	go run cmd/pgtest/main.go
//
// If successful, it will show which connection string format worked.
// This is useful for troubleshooting connection issues between the host and PostgreSQL containers.
func main() {
	fmt.Println("Testing PostgreSQL connection...")

	// Connection parameters
	host := "localhost"
	port := 5432

	// Get project root directory
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	// Go up one directory if running from cmd/pgtest
	var rootDir string
	if filepath.Base(workingDir) == "pgtest" {
		rootDir = filepath.Dir(filepath.Dir(workingDir))
	} else {
		rootDir = workingDir
	}

	// Read credentials from secret files
	configDir := filepath.Join(rootDir, "config")
	user, err := os.ReadFile(filepath.Join(configDir, "postgres_user"))
	if err != nil {
		fmt.Printf("Error reading postgres_user: %v\n", err)
		os.Exit(1)
	}

	password, err := os.ReadFile(filepath.Join(configDir, "postgres_password"))
	if err != nil {
		fmt.Printf("Error reading postgres_password: %v\n", err)
		os.Exit(1)
	}

	dbname, err := os.ReadFile(filepath.Join(configDir, "postgres_db"))
	if err != nil {
		fmt.Printf("Error reading postgres_db: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Using credentials from secret files: user=%s, dbname=%s\n", string(user), string(dbname))

	// Try different connections
	connURLs := []string{
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", string(user), string(password), host, port, string(dbname)),
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, string(user), string(password), string(dbname)),
	}

	// Try to connect
	success := false
	for i, connURL := range connURLs {
		fmt.Printf("\n===== Connection String %d =====\n", i+1)
		fmt.Printf("Connecting with: %s\n", connURL)

		if tryConnect(connURL) {
			success = true
			fmt.Printf("✅ Connection %d successful!\n", i+1)
		} else {
			fmt.Printf("❌ Connection %d failed\n", i+1)
		}
	}

	if !success {
		fmt.Println("\n❌ All connection attempts failed")
		os.Exit(1)
	}

	fmt.Println("\n✅ At least one connection succeeded!")
}

// Helper function to try different connection methods
func tryConnect(connStr string) bool {
	// Try to connect with retries
	maxRetries := 3
	var db *sql.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		fmt.Printf("  Attempt %d/%d\n", i+1, maxRetries)

		db, err = sql.Open("postgres", connStr)
		if err != nil {
			fmt.Printf("  Error opening connection: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Set a timeout for the ping operation
		db.SetConnMaxLifetime(10 * time.Second)
		db.SetMaxIdleConns(1)
		db.SetMaxOpenConns(1)

		err = db.Ping()
		if err == nil {
			fmt.Println("  Successfully connected to the database!")

			// Try a simple query
			var result int
			err = db.QueryRow("SELECT 1").Scan(&result)
			if err != nil {
				fmt.Printf("  Query failed: %v\n", err)
				db.Close()
				return false
			}

			fmt.Printf("  Query result: %d\n", result)
			db.Close()
			return true
		}

		fmt.Printf("  Failed to ping database: %v\n", err)
		db.Close()
		time.Sleep(1 * time.Second)
	}

	return false
}
