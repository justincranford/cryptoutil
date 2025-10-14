---
description: "Instructions for Go dependency management"
applyTo: "**/*.go"
---
# Go Dependencies Instructions

## Dependency Updates

- Check for updates using `go list -u -m all | grep '\[.*\]$'`
- Update dependencies incrementally with `go get <package>@<version>`
- Run `go mod tidy` after updates to clean up
- Test after each update with `go test ./... --count=1 -timeout=20m`

## Version Synchronization

- **When direct and transient dependency versions mismatch**: Prefer to synchronize versions to the latest compatible version
- **Focus on ecosystem packages**: Only update transient dependencies that belong to the same ecosystem as your direct dependencies
- **Example**: If you have `go.opentelemetry.io/otel` as a direct dependency, synchronize related packages like `go.opentelemetry.io/contrib`, `go.opentelemetry.io/otel/exporters/*`
- **Ignore unrelated transients**: Standalone transient dependencies not related to your direct dependencies can remain at MVS-selected versions
- Update indirect dependencies that can be safely upgraded: `go get <indirect-package>@<latest-version>`
- Maintain semantic versioning compatibility within major versions
- Go's MVS (Minimal Version Selection) handles most conflicts, but manual synchronization improves consistency

## Best Practices

- Update direct dependencies first, then ecosystem-related indirect ones
- Keep related packages (e.g., OpenTelemetry ecosystem, GORM ecosystem) at consistent versions
- Use `go mod why <package>` to understand why indirect dependencies are needed
- Only synchronize versions for packages that share the same import path prefix as your direct dependencies
- Let Go's MVS handle unrelated transient dependencies automatically

## Security and Maintenance

- Regularly update dependencies for security patches
- Review changelog/release notes for breaking changes
- Prefer stable releases over pre-releases
- Update dependencies in small batches to isolate issues
