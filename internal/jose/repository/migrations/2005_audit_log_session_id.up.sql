-- Copyright (c) 2025 Justin Cranford
--
--

-- Add session_id column to audit_log for linking audit entries to sessions.
ALTER TABLE tenant_audit_log ADD COLUMN session_id TEXT;
