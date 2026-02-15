-- Citus Distributed Database Initialization
-- Setup distributed tables and reference tables

-- Create extension if not exists
CREATE EXTENSION IF NOT EXISTS citus;

-- Example distributed table setup
-- Services will create their own distributed tables as needed
-- This file serves as a template

-- Tenant table (reference table - replicated to all workers)
-- CREATE TABLE tenants (
--     id UUID PRIMARY KEY,
--     name TEXT NOT NULL,
--     created_at TIMESTAMPTZ DEFAULT NOW()
-- );
-- SELECT create_reference_table('tenants');

-- User sessions table (distributed by tenant_id)
-- CREATE TABLE user_sessions (
--     id UUID PRIMARY KEY,
--     tenant_id UUID NOT NULL,
--     user_id UUID NOT NULL,
--     created_at TIMESTAMPTZ DEFAULT NOW(),
--     expires_at TIMESTAMPTZ NOT NULL
-- );
-- SELECT create_distributed_table('user_sessions', 'tenant_id');

-- Note: Individual services will create their distributed tables
-- using the Citus extension APIs in their migration scripts
