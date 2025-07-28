package main

import (
	"database/sql"
	"fmt"
	"os"
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
	user := "USR"
	password := "PWD"
	dbname := "DB"

	// Try different connections
	connURLs := []string{
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname),
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname),
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
