-- Restore the legacy client-secret column name.
ALTER TABLE clients RENAME COLUMN client_secret_hash TO client_secret;
ALTER TABLE unverified_clients RENAME COLUMN client_secret_hash TO client_secret;
