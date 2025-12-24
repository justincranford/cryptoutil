# Minimum Versions & Consistency Requirements

**Version**: 1.0.0
**Last Updated**: 2025-12-24
**Referenced By**: `.github/instructions/02-04.versions.instructions.md`

## Version Consistency Principle

**CRITICAL: ALWAYS use the same version in every part of the project**

- Development environment
- CI/CD workflows
- Docker images
- GitHub Actions workflows
- Documentation examples

## Minimum Required Versions

### Core Languages

| Tool | Minimum Version | Release Date | Reference |
|------|----------------|--------------|-----------|
| **Go** | 1.25.5+ | 2025-12-02 | <https://go.dev/doc/devel/release#go1.25.5> |
| **Python** | 3.14+ | 2025-10-07 | <https://www.python.org/downloads/release/python-3140/> |
| **Node** | v24.11.1+ LTS | 2025-11-11 | <https://nodejs.org/en/blog/release/v24.11.1/> |
| **Java** | 21 LTS | 2023-09-19 | <https://jdk.java.net/21/> |

**Note**: Java required for Gatling load tests in `test/load/`

### Build Tools

| Tool | Minimum Version | Release Date | Reference |
|------|----------------|--------------|-----------|
| **golangci-lint** | v2.7.2+ | 2025-12-19 | <https://github.com/golangci/golangci-lint/releases/tag/v2.7.2> |
| **Maven** | 3.9+ | 2023-01-31 | <https://maven.apache.org/docs/history.html#3.9.0> |
| **pre-commit** | 2.20.0+ | 2022-07-10 | <https://github.com/pre-commit/pre-commit/releases/tag/v2.20.0> |

### Container Runtime

| Tool | Minimum Version | Release Date | Reference |
|------|----------------|--------------|-----------|
| **Docker** | 24+ | 2023-05-16 | <https://docs.docker.com/engine/release-notes/24.0/> |
| **Docker Compose** | v2+ | 2021-09-28 | <https://github.com/docker/compose/releases/tag/v2.0.0> |

## Version Update Policy

### Update Frequency

- **Security patches**: Apply immediately when available
- **Minor versions**: Update monthly (stability + new features)
- **Major versions**: Evaluate quarterly (breaking changes review)

### Verification Before Updates

**ALWAYS verify before suggesting version updates**:

1. Check official release notes for breaking changes
2. Review security advisories and CVE fixes
3. Validate compatibility with existing dependencies
4. Test in development environment before CI/CD updates
5. Update documentation to reflect new version requirements

### Consistency Enforcement

**Locations to update when changing versions**:

- `go.mod` (Go version)
- `pyproject.toml` (Python version)
- `package.json` (Node version)
- `.github/workflows/*.yml` (CI/CD workflow versions)
- `Dockerfile` (base image versions)
- `docker-compose.yml` (service image versions)
- `README.md` (developer setup instructions)
- `docs/DEV-SETUP.md` (detailed setup guide)

## Key Takeaways

1. **Version Consistency**: Same version across development, CI/CD, Docker, workflows
2. **Always Latest Stable**: Prefer latest stable versions for security and features
3. **Verify Before Update**: Check release notes, breaking changes, compatibility
4. **Update Everywhere**: go.mod, workflows, Dockerfiles, docs when changing versions
5. **Security First**: Apply security patches immediately, minor/major updates on schedule
