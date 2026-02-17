# Demo Brainstorm Documentation

## Purpose

This directory archives demo files that were removed from active deployments but preserved for future reference. These files serve as:

- **Historical Reference**: Examples of working demo configurations
- **Future Development**: Starting points for new demo implementations
- **Testing Patterns**: Proven patterns for demo environments
- **Educational Resources**: Learning materials for understanding service configurations

## Archived Files

### sm-kms Demo Compose File

**File**: `deployments/sm-kms/compose.demo.yml`
**Archived**: 2026-02-16
**Original Location**: `deployments/sm-kms/compose.demo.yml`

**Purpose**:
- Demonstrated SQLite-based single-instance deployment
- Showed minimal configuration for local development
- Provided working example of sm-kms service setup

**Archival Rationale**:
- No longer part of standard deployment patterns
- Conflicts were discovered with production compose structure
- Preserved for future demo development reference

**Key Features**:
- SQLite database configuration
- Single-instance deployment model
- Local development setup
- Environment variable configuration

## Creating Future Demos

When creating new demo files, consider:

1. **Isolation**: Demo files should not conflict with production deployments
2. **Documentation**: Clearly document demo purpose and setup steps
3. **Simplicity**: Keep demos minimal and focused on specific features
4. **Maintenance**: Regular testing to ensure demos remain functional

## Directory Structure

```
demo-brainstorm/
├── README.md                  # High-level overview
├── DEMO-BRAINSTORM.md        # This file - detailed documentation
└── deployments/              # Mirrors deployment structure
    └── sm-kms/
        └── compose.demo.yml  # Archived demo compose file
```

## See Also

- [ARCHITECTURE.md](/docs/ARCHITECTURE.md) - System architecture documentation
- [deployments/](/deployments/) - Production deployment configurations
- [configs/](/configs/) - Production configuration files

## Maintenance

This directory should be reviewed periodically to:
- Remove obsolete files
- Update documentation
- Validate that archived files still provide value
- Extract patterns for permanent documentation
