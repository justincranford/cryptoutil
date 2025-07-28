# PostgreSQL Connection Troubleshooting

## Problem
When running cryptoutil on the host machine (via VSCode debugger), it failed to connect to the PostgreSQL container with "failed to ping database" errors.

## Root Causes
1. Newline characters in the Docker secret files caused encoding issues
2. Authentication issues between host and container
3. Container mode needed to be explicitly disabled

## Solution
1. Fixed secret files by ensuring no trailing newlines
2. Created a dedicated debugging configuration using:
   - Alternative connection string format: `host=localhost port=5432 user=USR password=PWD dbname=DB sslmode=disable`
   - Explicitly disabled container mode: `database-container: "disabled"`

## Updated Configuration Files
1. Created `postgresql-local-debug.yaml` for host debugging
2. Updated launch.json with a new "cryptoutil postgres-local" configuration
3. Fixed Docker secret files using PowerShell's Set-Content with -NoNewline flag

## Connection Testing
We created test utilities in the `cmd/pgtest` directory to validate the connection.
