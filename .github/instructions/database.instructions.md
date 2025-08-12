---
description: "Instructions for database operations and ORM patterns"
applyTo: "**"
---
# Database and ORM Instructions

- Always use GORM ORM for database operations, never use sql.DB directly
- Use transaction-based operations with proper isolation levels (ReadCommitted for writes, ReadOnly for reads)
- Implement proper error mapping from GORM errors to application HTTP errors in `toAppErr` methods
- Use embedded SQL migrations with golang-migrate for schema changes
- Support both PostgreSQL (production) and SQLite (development/testing) backends
- Use proper connection pooling and timeout configurations
- Always apply database migrations on startup
- Use proper pagination with offset/limit patterns for large result sets
- Implement filter patterns using GORM's query builder methods
- Use UUIDv7 primary keys for all entities
- Ensure proper foreign key relationships and constraints
- Log database schema information in debug mode for troubleshooting
