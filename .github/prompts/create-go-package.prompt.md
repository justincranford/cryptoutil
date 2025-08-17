---
mode: agent
description: "Generate a new Go package with proper error handling, testing, and documentation"
tools: ["semantic_search", "file_search", "grep_search", "read_file", "create_file"]
---

# Create Go Package Template

Create a new Go package based on our project standards.

I need a new package under the specified directory with:
- Proper error handling using our custom error types
- Unit tests with at least 80% coverage
- Documentation that follows Go standards
- Implementation that follows our project layout standards

Package name: ${input:packageName}
Location: ${input:location:internal/pkg/}
Main functionality: ${input:functionality}
