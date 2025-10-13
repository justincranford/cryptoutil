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

## File Structure

- Use `.github/copilot-instructions.md` as the primary instruction file
- Create task-specific instruction files in `.github/instructions/` with proper YAML frontmatter
- Use the `applyTo` property to target specific file types
- Create reusable prompt files in `.github/prompts/` for common tasks
- Files are automatically discovered by VS Code - no manual registration required

## Best Practices

- Split instructions into multiple files organized by topic or task type
- Maintain consistency between related instruction files  
- Use proper `.instructions.md` extension and YAML frontmatter for autodiscovery
- Enable instruction files with `github.copilot.chat.codeGeneration.useInstructionFiles: true` setting
- Use the appropriate instruction type for each task:
  - Code generation: `.github/copilot-instructions.md` or task-specific `.instructions.md` files
  - Code review: Configure with `github.copilot.chat.reviewSelection.instructions` setting
  - Commit messages: Configure with `github.copilot.chat.commitMessageGeneration.instructions` setting
  - PR descriptions: Configure with `github.copilot.chat.pullRequestDescriptionGeneration.instructions` setting

## Terminal Command Guidelines

- **Avoid `cd` commands** in terminal operations - they are not authorized and break agentic iteration
- **Use full paths** when referencing files/directories outside current context
- **Work within current directory** when possible to maintain context
- **Prefer authorized commands** like `docker`, `pwd`, `git` for navigation and operations
- **Use command flags** like `-f` or `--file` to specify paths instead of changing directories

## Authorization Patterns

**Commands that work WITHOUT manual authorization:**
- Simple docker compose commands: `docker compose ps`, `docker compose logs <service>`, `docker compose exec <service> <command>`
- Basic docker commands without extra parameters: `docker ps`, `docker inspect <container>` (without --format)
- Git commands: `git status`, `git add`, `git commit`, `git log`, `git diff`
- File operations: `pwd`, `ls`, `cat`, `head`, `tail`
- Go commands: `go test`, `go build`, `go mod tidy`

**Commands that REQUIRE manual authorization (avoid these):**
- Docker commands with formatting/filtering: `docker inspect --format`, `docker ps --filter`, `docker ps --format`
- Complex docker compose with extra flags: `docker compose -f <path>` (when not needed)
- Directory navigation: `cd` commands
- Advanced docker operations: `docker stats`, `docker top`, etc.

**Pattern: Use simple commands first, add complexity only when necessary**

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
