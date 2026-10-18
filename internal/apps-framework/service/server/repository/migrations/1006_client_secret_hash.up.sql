-- The client repository stores only high-entropy client-secret hashes.
ALTER TABLE clients RENAME COLUMN client_secret TO client_secret_hash;
ALTER TABLE unverified_clients RENAME COLUMN client_secret TO client_secret_hash;
