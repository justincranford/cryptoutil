---
description: "Instructions for VS Code Copilot customization"
applyTo: "**"
---
# VS Code Copilot Customization Instructions

## General Principles

- Keep instructions short and self-contained
- Each instruction should be a single, simple statement
- Don't reference external resources in instructions
- Store instructions in properly structured files for version control and team sharing

## Authorized Commands Reference

### Docker Commands
- `docker compose ps` - List container status (AUTHORIZED - no extra parameters)
- `docker compose logs <service>` - View service logs (AUTHORIZED)
- `docker compose exec <service> <command>` - Execute command in container (AUTHORIZED)
- `docker compose build <services>` - Build services (AUTHORIZED)
- `docker compose up -d <services>` - Start services in background (AUTHORIZED)
- `docker compose down -v` - Stop services and remove volumes (AUTHORIZED)
- `docker inspect <container>` - Inspect container (use without --format to avoid authorization)
- `docker ps` - List containers (use without --filter/--format to avoid authorization)

### Git Commands
- `git -C <path> status` - Show repository status
- `git -C <path> add <files>` - Stage files for commit
- `git -C <path> commit -m <message>` - Commit staged changes
- `git -C <path> log --oneline -<n>` - Show recent commit history
- `git -C <path> diff` - Show unstaged changes
- `git -C <path> checkout <branch>` - Switch branches

### Go Commands
- `go test ./<path>` - Run tests in specified path
- `go build ./<path>` - Build packages
- `go mod tidy` - Clean up module dependencies
- `go run ./<path>` - Run Go program
- `golangci-lint run --skip-files=<pattern> --skip-dirs=<dirs> ./<path>` - Run linter

### File/Directory Operations
- `pwd` - Show current working directory
- `ls -la <path>` - List directory contents with details
- `find <path> -name <pattern>` - Find files by name pattern
- `cat <file>` - Display file contents
- `head -n <lines> <file>` - Show first N lines of file
- `tail -n <lines> <file>` - Show last N lines of file

### Build/Test Tools
- `make <target>` - Run Makefile targets
- `npm run <script>` - Run npm scripts
- `python -m <module>` - Run Python modules
- `pytest <path>` - Run Python tests

### System Commands
- `which <command>` - Find command location
- `echo <text>` - Display text
- `date` - Show current date/time
- `sleep <seconds>` - Pause execution

## Commands Requiring Manual Authorization

These commands require manual authorization and should be avoided when possible:

### Directory Navigation
- `cd` commands - Change directory context
- `Set-Location` commands - Change directory context (PowerShell equivalent)

### Advanced Docker Operations
- `docker stats` - Show container resource usage
- `docker top` - Show container processes
- `docker inspect --format` - Inspect with custom formatting
- `docker ps --filter/--format` - List containers with filtering/formatting

### Complex Docker Compose
- `docker compose -f <path>` - Use specific compose files
