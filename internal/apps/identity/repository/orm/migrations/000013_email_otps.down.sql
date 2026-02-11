-- Copyright (c) 2025 Iwan van der Kleijn
-- SPDX-License-Identifier: MIT

-- Rollback email_otps table.

DROP INDEX IF EXISTS idx_email_otps_used;
DROP INDEX IF EXISTS idx_email_otps_expires_at;
DROP INDEX IF EXISTS idx_email_otps_user_id;
DROP TABLE IF EXISTS email_otps;
